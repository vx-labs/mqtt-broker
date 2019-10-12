package store

import (
	"errors"

	"github.com/boltdb/bolt"
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)

var (
	// ErrKeyNotFound is an error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
	// ErrQueueNotFound is an error indicating a given queue does not exist
	ErrQueueNotFound = errors.New("queue not found")
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
	path string
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
		conn: handle,
		path: options.Path,
	}
	return store, nil
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
	return tx.Commit()
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
	err = bucket.Put(uint64ToBytes(index), payload)
	if err != nil {
		return err
	}
	return tx.Commit()
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
	cursor := bucket.Cursor()
	offset := from
	firstLoop := func() ([]byte, []byte) {
		if from == 0 {
			return cursor.First()
		}
		cursor.Seek(uint64ToBytes(from))
		return cursor.Next()
	}

	for itemKey, itemValue := firstLoop(); itemKey != nil && idx < len(buff); itemKey, itemValue = cursor.Next() {
		buff[idx] = itemValue
		offset = bytesToUint64(itemKey)
		idx++
	}
	return offset, idx, nil
}
