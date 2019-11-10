package store

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/kv/pb"
)

var kvBucket = []byte("kvstore")
var deadlinesBucketName = []byte("kvstore.deadlines")
var metadatasBucketName = []byte("kvstore.metadata")

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

	return store, store.initStore()
}

func (b *BoltStore) initStore() error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.CreateBucketIfNotExists(kvBucket)
	if err != nil {
		return err
	}
	_, err = tx.CreateBucketIfNotExists(deadlinesBucketName)
	if err != nil {
		return err
	}
	_, err = tx.CreateBucketIfNotExists(metadatasBucketName)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func loadMetadata(bucket *bolt.Bucket, key []byte) (*pb.KVMetadata, error) {
	data := bucket.Get(key)
	if data == nil {
		return nil, ErrKeyNotFound
	}
	md := &pb.KVMetadata{}
	err := proto.Unmarshal(data, md)
	if err != nil {
		return nil, err
	}
	return md, nil
}
func (b *BoltStore) Put(key []byte, value []byte, deadline uint64) error {
	bucketName := kvBucket
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return ErrBucketNotFound
	}
	err = bucket.Put(key, value)
	if err != nil {
		return err
	}
	if deadline > 0 {
		deadlineBucket := tx.Bucket(deadlinesBucketName)
		if deadlineBucket == nil {
			return ErrTTLBucketNotFound
		}
		err = deadlineBucket.Put(uint64ToBytes(deadline), key)
		if err != nil {
			return err
		}
	}
	metadatasBucket := tx.Bucket(metadatasBucketName)
	if err != nil {
		return err
	}
	md := &pb.KVMetadata{
		Deadline: deadline,
		Version:  1,
	}
	mdPayload, err := proto.Marshal(md)
	if err != nil {
		return err
	}
	metadatasBucket.Put(key, mdPayload)
	err = tx.Commit()
	return err
}
func (b *BoltStore) Delete(key []byte) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = b.deleteKey(tx, key)
	if err == nil {
		return tx.Commit()
	}
	return err
}
func (b *BoltStore) deleteKey(tx *bolt.Tx, key []byte) error {
	bucketName := kvBucket
	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return ErrBucketNotFound
	}
	err := bucket.Delete(key)
	if err != nil {
		return err
	}
	metadatasBucket := tx.Bucket(metadatasBucketName)
	if err != nil {
		return err
	}
	md, err := loadMetadata(metadatasBucket, key)
	if err == nil {
		if md.Deadline > 0 {
			deadlineBucket := tx.Bucket(deadlinesBucketName)
			if deadlineBucket == nil {
				return ErrTTLBucketNotFound
			}
			err = deadlineBucket.Delete(uint64ToBytes(md.Deadline))
			if err != nil {
				return err
			}
		}
		err := metadatasBucket.Delete(key)
		if err != nil {
			return err
		}
	}
	return err
}
func (b *BoltStore) Get(key []byte) ([]byte, error) {
	bucketName := kvBucket
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(bucketName)
	if bucket == nil {
		return nil, ErrBucketNotFound
	}
	value := bucket.Get(key)
	return value, nil
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

func (b *BoltStore) DeleteBatch(keys [][]byte) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, key := range keys {
		err := b.deleteKey(tx, key)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (b *BoltStore) ListExpiredKeys(now uint64) ([][]byte, error) {
	nowKey := uint64ToBytes(now)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	deadlineBucket := tx.Bucket(deadlinesBucketName)
	if deadlineBucket == nil {
		return nil, ErrTTLBucketNotFound
	}
	out := [][]byte{}
	cursor := deadlineBucket.Cursor()
	for itemKey, itemValue := cursor.First(); itemKey != nil && bytes.Compare(itemKey, nowKey) < 0; itemKey, itemValue = cursor.Next() {
		out = append(out, itemValue)
	}
	return out, err
}
