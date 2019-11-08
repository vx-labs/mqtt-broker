package store

import (
	"github.com/boltdb/bolt"
)

func (b *BoltStore) delete(backend *bolt.Bucket, index uint64) error {
	return backend.Delete(uint64ToBytes(index))
}
func (b *BoltStore) put(backend *bolt.Bucket, index uint64, payload []byte) error {
	if index < backend.Sequence() {
		return ErrIndexOutdated
	}
	err := backend.Put(uint64ToBytes(index), payload)
	return err
}
func (b *BoltStore) get(backend *bolt.Bucket, offset uint64) []byte {
	return backend.Get(uint64ToBytes(offset))
}
func (b *BoltStore) getRange(backend *bolt.Bucket, offset uint64, buff []StoredMessage) (int, uint64, error) {
	idx := 0
	cursor := backend.Cursor()
	count := len(buff)
	firstLoop := func() ([]byte, []byte) {
		if offset == 0 {
			return cursor.First()
		}
		cursor.Seek(uint64ToBytes(offset))
		return cursor.Next()
	}

	for itemKey, itemValue := firstLoop(); itemKey != nil && idx < count; itemKey, itemValue = cursor.Next() {
		offset = bytesToUint64(itemKey)
		buff[idx] = StoredMessage{Offset: offset, Payload: itemValue}
		idx++
	}
	return idx, offset, nil
}
func (b *BoltStore) walk(backend *bolt.Bucket, f func(key []byte, payload []byte) error) error {
	cursor := backend.Cursor()
	for itemKey, itemValue := cursor.First(); itemKey != nil; itemKey, itemValue = cursor.Next() {
		err := f(itemKey, itemValue)
		if err != nil {
			return err
		}
	}
	return nil
}
