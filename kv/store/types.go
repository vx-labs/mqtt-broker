package store

import "errors"

var (
	// ErrKeyNotFound is an error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
	// ErrBucketNotFound is an error indicating a given stream does not exist
	ErrBucketNotFound = errors.New("bucket not found")
	// ErrIndexOutdated is an error indicating that the supplied index is outdated
	ErrIndexOutdated = errors.New("index outdated")
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)
