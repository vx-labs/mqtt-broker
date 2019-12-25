package store

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"github.com/vx-labs/mqtt-broker/services/kv/pb"
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
		return &pb.KVMetadata{
			Key:     key,
			Version: 0,
		}, nil
	}
	md := &pb.KVMetadata{}
	err := proto.Unmarshal(data, md)
	if err != nil {
		return nil, err
	}
	return md, nil
}
func saveMetadata(bucket *bolt.Bucket, key []byte, md *pb.KVMetadata) error {
	data, err := proto.Marshal(md)
	if err != nil {
		return err
	}
	return bucket.Put(key, data)
}
func (b *BoltStore) Put(key []byte, value []byte, deadline uint64, version uint64) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(kvBucket)
	if bucket == nil {
		return ErrBucketNotFound
	}
	metadatasBucket := tx.Bucket(metadatasBucketName)
	if err != nil {
		return ErrMDBucketNotFound
	}

	var deadlineBucket *bolt.Bucket

	var md *pb.KVMetadata
	md, err = loadMetadata(metadatasBucket, key)
	if err != nil {
		return err
	}

	if md.Version != version {
		return ErrIndexOutdated
	}

	if deadline > 0 || md.Deadline > 0 {
		deadlineBucket = tx.Bucket(deadlinesBucketName)
		if deadlineBucket == nil {
			return ErrTTLBucketNotFound
		}
	}

	if md.Deadline > 0 {
		deadlineBucket.Delete(uint64ToBytes(md.Deadline))
	}

	md.Version++
	md.Deadline = deadline

	err = bucket.Put(key, value)
	if err != nil {
		return err
	}

	mdPayload, err := proto.Marshal(md)
	if err != nil {
		return err
	}
	err = metadatasBucket.Put(key, mdPayload)
	if err != nil {
		return err
	}

	if deadline > 0 {
		err = deadlineBucket.Put(uint64ToBytes(deadline), mdPayload)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}
func (b *BoltStore) Delete(key []byte, version uint64) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = b.deleteKey(tx, key, version)
	if err == nil {
		return tx.Commit()
	}
	return err
}
func (b *BoltStore) deleteKey(tx *bolt.Tx, key []byte, version uint64) error {
	bucket := tx.Bucket(kvBucket)
	if bucket == nil {
		return ErrBucketNotFound
	}
	metadatasBucket := tx.Bucket(metadatasBucketName)
	if metadatasBucket == nil {
		return ErrMDBucketNotFound
	}

	md, err := loadMetadata(metadatasBucket, key)
	if err != nil {
		return err
	}
	if version != md.Version {
		return ErrIndexOutdated
	}
	md.Version++
	err = saveMetadata(metadatasBucket, key, md)
	if err != nil {
		return err
	}
	deadlinesBucket := tx.Bucket(deadlinesBucketName)
	if deadlinesBucket == nil {
		return ErrTTLBucketNotFound
	}
	if md.Deadline > 0 {
		deadlinesBucket.Delete(uint64ToBytes(md.Deadline))
	}

	return bucket.Delete(key)
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
func (b *BoltStore) GetMetadata(key []byte) (*pb.KVMetadata, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(metadatasBucketName)
	if bucket == nil {
		return nil, ErrBucketNotFound
	}
	return loadMetadata(bucket, key)
}
func (b *BoltStore) GetWithMetadata(key []byte) ([]byte, *pb.KVMetadata, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	mdBucket := tx.Bucket(metadatasBucketName)
	if mdBucket == nil {
		return nil, nil, ErrMDBucketNotFound
	}
	bucket := tx.Bucket(kvBucket)
	if bucket == nil {
		return nil, nil, ErrBucketNotFound
	}
	md, err := loadMetadata(mdBucket, key)
	if err != nil {
		return nil, nil, ErrMDNotFound
	}
	if md == nil {
		md = &pb.KVMetadata{Key: key, Version: 0}
	}
	value := bucket.Get(key)
	return value, md, nil
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

func (b *BoltStore) DeleteBatch(keys []*pb.KVMetadata) error {
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for idx := range keys {
		elt := keys[idx]
		err := b.deleteKey(tx, elt.Key, elt.Version)
		if err != nil {
			if err == ErrIndexOutdated {
				continue
			}
			return err
		}
	}
	return tx.Commit()
}
func (b *BoltStore) ListExpiredKeys(from uint64, until uint64) ([]*pb.KVMetadata, error) {
	nowKey := uint64ToBytes(until)
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	deadlineBucket := tx.Bucket(deadlinesBucketName)
	if deadlineBucket == nil {
		return nil, ErrTTLBucketNotFound
	}
	mdBucket := tx.Bucket(metadatasBucketName)
	if mdBucket == nil {
		return nil, ErrMDBucketNotFound
	}
	out := []*pb.KVMetadata{}
	cursor := deadlineBucket.Cursor()

	for itemKey, itemValue := cursor.Seek(uint64ToBytes(from)); itemKey != nil && bytes.Compare(itemKey, nowKey) < 0; itemKey, itemValue = cursor.Next() {
		md := &pb.KVMetadata{}
		err := proto.Unmarshal(itemValue, md)
		if err == nil {
			out = append(out, md)
		}
	}
	return out, err
}

func (b *BoltStore) ListKeys(prefix []byte) ([]*pb.KVMetadata, error) {
	tx, err := b.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	mdBucket := tx.Bucket(metadatasBucketName)
	if mdBucket == nil {
		return nil, ErrMDBucketNotFound
	}
	out := []*pb.KVMetadata{}
	cursor := mdBucket.Cursor()
	if len(prefix) > 0 {
		for itemKey, itemValue := cursor.Seek(prefix); itemKey != nil && bytes.HasPrefix(prefix, itemKey); itemKey, itemValue = cursor.Next() {
			md := &pb.KVMetadata{}
			err := proto.Unmarshal(itemValue, md)
			if err == nil {
				out = append(out, md)
			}
		}
	} else {
		for itemKey, itemValue := cursor.First(); itemKey != nil; itemKey, itemValue = cursor.Next() {
			md := &pb.KVMetadata{}
			err := proto.Unmarshal(itemValue, md)
			if err == nil {
				out = append(out, md)
			}
		}
	}
	return out, err
}
