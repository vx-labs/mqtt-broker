package store

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vx-labs/mqtt-broker/queues/pb"

	"github.com/boltdb/bolt"
	"github.com/vx-labs/mqtt-broker/events"
)

type Options struct {
	// Path is the file path to the BoltDB to use
	Path string

	// BoltOptions contains any specific BoltDB options you might
	// want to specify [e.g. open timeout]
	BoltOptions *bolt.Options

	// NoSync causes the database to skip fsync calls after each
	// write to the log. This is unsafe, so it should be used
	// with caution.
	NoSync bool
}

type BoltStore struct {
	conn        *bolt.DB
	options     Options
	eventBus    *events.Bus
	restoreLock sync.Mutex
}

// Close is used to gracefully close the DB connection.
func (b *BoltStore) Close() error {
	return b.conn.Close()
}

type BatchInput struct {
	ID     string
	Offset uint64
	Data   [][]byte
}

type BatchOutput struct {
	ID     string
	Offset uint64
	Count  int
	Err    error
}

func New(options Options) (*BoltStore, error) {
	// Try to connect
	handle, err := bolt.Open(options.Path, dbFileMode, options.BoltOptions)
	if err != nil {
		return nil, err
	}
	handle.NoSync = options.NoSync

	// Create the new store
	store := &BoltStore{
		conn:     handle,
		options:  options,
		eventBus: events.NewEventBus(),
	}
	return store, nil
}

func getBucketName(id string) []byte {
	return []byte(id)
}
func getInflightBucketName(id string) []byte {
	return []byte(fmt.Sprintf("%s.inflight", id))
}

func (b *BoltStore) On(queue string, event string, f func(payload interface{})) (cancel func()) {
	return b.eventBus.Subscribe(fmt.Sprintf("%s/%s", queue, event), func(e events.Event) {
		f(e.Entry)
	})
}

func (b *BoltStore) Exists(id string) bool {
	bucketName := getBucketName(id)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return false
	}
	defer tx.Rollback()
	return tx.Bucket(bucketName) != nil
}
func (b *BoltStore) DeleteQueue(id string) error {
	bucketName := getBucketName(id)
	inflightBucketName := getInflightBucketName(id)

	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.DeleteBucket(bucketName); err != nil {
		return err
	}
	if err := tx.DeleteBucket(inflightBucketName); err != nil {
		return err
	}
	err = tx.Commit()
	if err == nil {
		b.eventBus.Emit(events.Event{
			Key: fmt.Sprintf("%s/queue_deleted", id),
		})
	}
	return err
}
func (b *BoltStore) CreateQueue(id string) error {
	bucketName := getBucketName(id)
	inflightBucketName := getInflightBucketName(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
		return err
	}
	if _, err := tx.CreateBucketIfNotExists(inflightBucketName); err != nil {
		return err
	}
	return tx.Commit()
}
func (b *BoltStore) GetStatistics(id string) (*pb.QueueStatistics, error) {
	bucketName := getBucketName(id)
	inflightBucketName := getInflightBucketName(id)

	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return &pb.QueueStatistics{
		ID:            id,
		MessageCount:  int64(tx.Bucket(bucketName).Stats().KeyN),
		InflightCount: int64(tx.Bucket(inflightBucketName).Stats().KeyN),
	}, nil
}
func (b *BoltStore) ListQueues() ([]string, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	out := []string{}
	tx.ForEach(func(bucketName []byte, bucket *bolt.Bucket) error {
		if !bytes.HasSuffix(bucketName, []byte(".inflight")) {
			out = append(out, string(bucketName))
		}
		return nil
	})
	return out, nil
}
func (b *BoltStore) Delete(id string, index uint64) error {
	bucketName := getBucketName(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return ErrQueueNotFound
	}
	if err := b.delete(bucket, index); err != nil {
		return err
	}
	return tx.Commit()
}
func (b *BoltStore) Put(id string, index uint64, payload []byte) error {
	bucketName := getBucketName(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return ErrQueueNotFound
	}
	err = b.put(bucket, index, payload)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err == nil {
		b.eventBus.Emit(events.Event{
			Key: fmt.Sprintf("%s/message_put", id),
		})
	}
	return err
}
func (b *BoltStore) PutBatch(batches []*pb.QueueStateTransitionMessagePut) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, batch := range batches {
		bucket := tx.Bucket(getBucketName(batch.QueueID))
		if bucket == nil {
			return ErrQueueNotFound
		}
		err = b.put(bucket, batch.Offset, batch.Payload)
		if err != nil {
			return err
		}
		return err
	}
	err = tx.Commit()
	if err == nil {
		for _, batch := range batches {
			b.eventBus.Emit(events.Event{
				Key: fmt.Sprintf("%s/message_put", batch.QueueID),
			})
		}
	}
	return err
}
func (b *BoltStore) AckInflight(id string, index uint64) error {
	inflightBucketName := getInflightBucketName(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(inflightBucketName)
	if bucket == nil {
		return ErrQueueNotFound
	}
	if err := b.delete(bucket, index); err != nil {
		return err
	}
	return tx.Commit()
}
func (b *BoltStore) SetInflight(id string, index uint64, deadline uint64) error {
	bucketName := getBucketName(id)
	inflightBucketName := getInflightBucketName(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return ErrQueueNotFound
	}
	inflightBucket := tx.Bucket(inflightBucketName)
	if bucket == nil {
		return ErrQueueNotFound
	}

	value := b.get(bucket, index)
	err = inflightBucket.Put(uint64ToBytes(deadline), value)
	if err != nil {
		return err
	}
	err = bucket.Delete(uint64ToBytes(index))
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}
func (b *BoltStore) TickExpiredMessages(messages []*pb.ExpiredInflights) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	for _, message := range messages {
		bucket := tx.Bucket([]byte(message.QueueID))
		inflightBucket := tx.Bucket(getInflightBucketName(message.QueueID))
		for _, offset := range message.Offsets {
			key := uint64ToBytes(offset)
			inflightMessage := inflightBucket.Get(key)
			if inflightMessage != nil {
				err := bucket.Put(key, inflightMessage)
				if err != nil {
					return err
				}
			}
		}
	}
	err = tx.Commit()
	if err == nil {
		for _, message := range messages {
			b.eventBus.Emit(events.Event{
				Key: fmt.Sprintf("%s/message_put", message.QueueID),
			})
		}
	}
	return err
}
func (b *BoltStore) GetExpiredInflights(currentTime time.Time) []*pb.ExpiredInflights {
	now := currentTime.UnixNano()
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	expiredInflights := []*pb.ExpiredInflights{}
	tx.ForEach(func(bucketName []byte, bucket *bolt.Bucket) error {
		if bytes.HasSuffix(bucketName, []byte(".inflight")) {
			queueBucketName := bytes.TrimSuffix(bucketName, []byte(".inflight"))
			queueBucket := tx.Bucket(queueBucketName)
			expiredInflights = append(expiredInflights, &pb.ExpiredInflights{
				QueueID: string(queueBucketName),
				Offsets: b.getExpiredInflight(tx, queueBucket, bucket, uint64(now)),
			})
			return nil
		}
		return nil
	})
	return expiredInflights
}
func (b *BoltStore) getExpiredInflight(tx *bolt.Tx, queueBucket *bolt.Bucket, bucket *bolt.Bucket, now uint64) []uint64 {
	out := make([]uint64, 0)
	b.walk(bucket, func(key []byte, payload []byte) error {
		deadline := bytesToUint64(key)
		if deadline < now {
			out = append(out, bytesToUint64(key))
			return nil
		}
		return io.EOF
	})
	return out
}
func (b *BoltStore) All() []string {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	out := []string{}
	err = tx.ForEach(func(bucketName []byte, _ *bolt.Bucket) error {
		if !bytes.HasSuffix(bucketName, []byte(".inflight")) {
			out = append(out, string(bucketName))
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return out
}

type StoredMessage struct {
	Offset  uint64
	Payload []byte
}

func (b *BoltStore) GetRange(id string, from uint64, buff []StoredMessage) (int, uint64, error) {
	bucketName := getBucketName(id)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return 0, 0, ErrQueueNotFound
	}
	return b.getRange(bucket, from, buff)
}

func (b *BoltStore) WriteTo(out io.Writer) error {
	return b.conn.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(out)
		return err
	})
}
func (b *BoltStore) Restore(r io.Reader) error {
	b.restoreLock.Lock()
	defer b.restoreLock.Unlock()
	err := b.conn.Close()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(b.options.Path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)

	handle, err := bolt.Open(b.options.Path, dbFileMode, b.options.BoltOptions)
	if err != nil {
		return err
	}
	handle.NoSync = b.options.NoSync
	b.conn = handle
	return err
}
