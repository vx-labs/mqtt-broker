package store

import (
	"io"
	"os"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/vx-labs/mqtt-broker/services/messages/pb"
)

var configKey = []byte("configuration")

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
		conn:    handle,
		options: options,
	}
	return store, nil
}

func getStreamName(id string) []byte {
	return []byte(id)
}

func (b *BoltStore) Exists(id string) bool {
	bucketName := getStreamName(id)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return false
	}
	defer tx.Rollback()
	return tx.Bucket(bucketName) != nil
}
func (b *BoltStore) DeleteStream(id string) error {
	bucketName := getStreamName(id)

	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.DeleteBucket(bucketName); err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func (b *BoltStore) CreateStream(config *pb.StreamConfig) error {
	id := config.ID
	shardIDs := config.ShardIDs
	bucketName := getStreamName(id)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stream, err := tx.CreateBucket(bucketName)
	if err != nil {
		return errors.Wrap(err, "failed to create stream")
	}
	for _, id := range shardIDs {
		_, err := stream.CreateBucket([]byte(id))
		if err != nil {
			return errors.Wrap(err, "failed to create shard")
		}
	}
	configPayload, err := proto.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "failed to encode stream config")
	}
	err = stream.Put(configKey, configPayload)
	if err != nil {
		return errors.Wrap(err, "failed to save stream config")
	}
	return tx.Commit()
}
func (b *BoltStore) Put(stream string, shardID string, index uint64, payload []byte) error {
	bucketName := getStreamName(stream)
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return ErrStreamNotFound
	}
	shard := bucket.Bucket([]byte(shardID))
	if shard == nil {
		return ErrShardNotFound
	}
	err = b.put(shard, index, payload)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err

}
func (b *BoltStore) GetStream(id string) *pb.StreamConfig {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	bucket := tx.Bucket(getStreamName(id))
	if bucket == nil {
		return nil
	}

	return b.getConfig(bucket)
}

func (b *BoltStore) getConfig(bucket *bolt.Bucket) *pb.StreamConfig {
	configPayload := bucket.Get(configKey)
	if configPayload == nil {
		return nil
	}
	config := &pb.StreamConfig{}
	err := proto.Unmarshal(configPayload, config)
	if err != nil {
		return nil
	}
	return config
}
func (b *BoltStore) ListStreams() []*pb.StreamConfig {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	out := []*pb.StreamConfig{}
	err = tx.ForEach(func(k []byte, bucket *bolt.Bucket) error {
		config := b.getConfig(bucket)
		if config != nil {
			out = append(out, config)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return out
}

func (b *BoltStore) GetStreamStatistics(id string) (*pb.StreamStatistics, error) {
	bucketName := getStreamName(id)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return nil, ErrStreamNotFound
	}
	config := b.getConfig(bucket)
	statistics := []*pb.ShardStatistics{}
	for _, shardID := range config.ShardIDs {
		shard := bucket.Bucket([]byte(shardID))
		if shard == nil {
			return nil, ErrShardNotFound
		}
		stats := b.statistics(shard)
		stats.StreamID = id
		stats.ShardID = shardID
		statistics = append(statistics, stats)
	}
	return &pb.StreamStatistics{
		ID:              id,
		ShardStatistics: statistics,
	}, nil
}

func (b *BoltStore) GetRange(id string, shardID string, from uint64, buff []*pb.StoredMessage) (int, uint64, error) {
	bucketName := getStreamName(id)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return 0, 0, ErrStreamNotFound
	}
	shard := bucket.Bucket([]byte(shardID))
	if shard == nil {
		return 0, 0, ErrShardNotFound
	}
	return b.getRange(shard, from, buff)
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
