package store

import "github.com/boltdb/bolt"

func (b *BoltStore) put(backend *bolt.Bucket, index uint64, payload []byte) error {
	if index < backend.Sequence() {
		return ErrIndexOutdated
	}
	err := backend.Put(uint64ToBytes(index), payload)
	return err
}
func (b *BoltStore) getRange(backend *bolt.Bucket, offset uint64, buff [][]byte) (uint64, int, error) {
	idx := 0
	cursor := backend.Cursor()

	firstLoop := func() ([]byte, []byte) {
		if offset == 0 {
			return cursor.First()
		}
		cursor.Seek(uint64ToBytes(offset))
		return cursor.Next()
	}

	for itemKey, itemValue := firstLoop(); itemKey != nil && idx < len(buff); itemKey, itemValue = cursor.Next() {
		buff[idx] = itemValue
		offset = bytesToUint64(itemKey)
		idx++
	}
	return offset, idx, nil
}