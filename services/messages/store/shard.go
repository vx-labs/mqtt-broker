package store

import (
	"github.com/boltdb/bolt"
	"github.com/vx-labs/mqtt-broker/services/messages/pb"
)

func (b *BoltStore) put(backend *bolt.Bucket, index uint64, payload []byte) error {
	err := backend.Put(uint64ToBytes(index), payload)
	if err != nil {
		return err
	}
	return backend.SetSequence(index)
}
func (b *BoltStore) get(backend *bolt.Bucket, offset uint64) []byte {
	return backend.Get(uint64ToBytes(offset))
}
func (b *BoltStore) getOffset(backend *bolt.Bucket) uint64 {
	return backend.Sequence()
}
func (b *BoltStore) getRange(backend *bolt.Bucket, offset uint64, buff []*pb.StoredMessage) (int, uint64, error) {
	idx := 0
	cursor := backend.Cursor()
	count := len(buff)
	for itemKey, itemValue := cursor.Seek(uint64ToBytes(offset)); itemKey != nil && idx < count; itemKey, itemValue = cursor.Next() {
		offset = bytesToUint64(itemKey)
		buff[idx] = &pb.StoredMessage{Offset: offset, Payload: itemValue}
		idx++
	}
	if idx == 0 {
		return idx, offset, nil
	}
	return idx, offset + 1, nil
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

func (b *BoltStore) statistics(backend *bolt.Bucket) *pb.ShardStatistics {
	stats := backend.Stats()
	return &pb.ShardStatistics{
		StoredRecordCount: int64(stats.KeyN),
		StoredBytes:       int64(stats.LeafInuse + stats.BranchInuse),
		CurrentOffset:     backend.Sequence(),
	}
}
