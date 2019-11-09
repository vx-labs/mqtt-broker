package store

import (
	"io"
	"os"
	"sync"

	"github.com/boltdb/bolt"
)

var kvBucket = []byte("kvstore")

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
	bucketName := kvBucket
	tx, err := b.conn.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (b *BoltStore) Put(key []byte, value []byte) error {
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
	err = tx.Commit()
	return err
}
func (b *BoltStore) Delete(key []byte) error {
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
	err = bucket.Delete(key)
	if err != nil {
		return err
	}
	err = tx.Commit()
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
