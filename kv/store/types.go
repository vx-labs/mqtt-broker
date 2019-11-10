package store

import "errors"

var (
	// ErrKeyNotFound is an error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
	// ErrBucketNotFound is an error indicating the data store bucket does not exist
	ErrBucketNotFound = errors.New("bucket not found")
	// ErrTTLBucketNotFound is an error indicating the TTL index bucket does not exist
	ErrTTLBucketNotFound = errors.New("ttl bucket not found")
	// ErrMDBucketNotFound is an error indicating the metadata bucket does not exist
	ErrMDBucketNotFound = errors.New("metadata bucket not found")
	// ErrMDNotFound is an error indicating metadata for the given key does not exist
	ErrMDNotFound = errors.New("metadata not found")
	// ErrIndexOutdated is an error indicating that the supplied index is outdated
	ErrIndexOutdated = errors.New("index outdated")
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)
