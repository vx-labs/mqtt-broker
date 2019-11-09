package store

import "errors"

var (
	// ErrKeyNotFound is an error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
	// ErrStreamNotFound is an error indicating a given stream does not exist
	ErrStreamNotFound = errors.New("stream not found")
	// ErrShardNotFound is an error indicating a given shard does not exist
	ErrShardNotFound = errors.New("shard not found")
	// ErrIndexOutdated is an error indicating that the supplied index is outdated
	ErrIndexOutdated = errors.New("index outdated")
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)

type Store interface {
	DeleteQueue(id string) error
	CreateQueue(id string) error
	Put(id string, index uint64, payload []byte) error
	GetRange(id string, from uint64, buff [][]byte) (uint64, int, error)
	All() []string
}
