package store

import (
	"fmt"

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
	// conn is the underlying handle to the db.
	conn *bolt.DB

	// The path to the Bolt database file
	path     string
	eventBus *events.Bus
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
		path:     options.Path,
		eventBus: events.NewEventBus(),
	}
	return store, nil
}

func (b *BoltStore) On(queue string, event string, f func(payload interface{})) (cancel func()) {
	return b.eventBus.Subscribe(fmt.Sprintf("%s/%s", queue, event), func(e events.Event) {
		f(e.Entry)
	})
}

func (b *BoltStore) Exists(id string) bool {
	bucketName := []byte(id)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return false
	}
	defer tx.Rollback()
	return tx.Bucket(bucketName) != nil
}
func (b *BoltStore) DeleteQueue(id string) error {
	bucketName := []byte(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.DeleteBucket(bucketName); err != nil {
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
	bucketName := []byte(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
		return err
	}
	return tx.Commit()
}
func (b *BoltStore) Put(id string, index uint64, payload []byte) error {
	bucketName := []byte(id)
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
func (b *BoltStore) All() []string {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	out := []string{}
	err = tx.ForEach(func(bucketName []byte, _ *bolt.Bucket) error {
		out = append(out, string(bucketName))
		return nil
	})
	if err != nil {
		return nil
	}
	return out
}
func (b *BoltStore) GetRange(id string, from uint64, buff [][]byte) (uint64, int, error) {
	idx := 0
	bucketName := []byte(id)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return 0, idx, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return 0, idx, ErrQueueNotFound
	}
	return b.getRange(bucket, from, buff)
}

func (b *BoltStore) GetRangeBatch(input []BatchInput) ([]BatchOutput, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	out := make([]BatchOutput, len(input))
	for idx := range input {
		out[idx] = BatchOutput{ID: input[idx].ID}
		bucket := tx.Bucket([]byte(input[idx].ID))
		if bucket == nil {
			out[idx].Err = ErrQueueNotFound
			continue
		}
		out[idx].Offset, out[idx].Count, out[idx].Err = b.getRange(bucket, input[idx].Offset, input[idx].Data)
	}
	return out, nil
}
