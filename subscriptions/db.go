package subscriptions

import (
	"errors"

	"github.com/vx-labs/mqtt-broker/subscriptions/pb"

	memdb "github.com/hashicorp/go-memdb"
)

func (m *memDBStore) do(write bool, f func(*memdb.Txn) error) error {
	tx := m.db.Txn(write)
	defer tx.Abort()
	err := f(tx)
	if write && err == nil {
		tx.Commit()
	}
	return err
}
func (m *memDBStore) read(f func(*memdb.Txn) error) error {
	return m.do(false, f)
}
func (m *memDBStore) write(f func(*memdb.Txn) error) error {
	return m.do(true, f)
}

func (m *memDBStore) first(tx *memdb.Txn, index string, value ...interface{}) (*pb.Metadata, error) {
	var ok bool
	var res *pb.Metadata
	data, err := tx.First(table, index, value...)
	if err != nil {
		return res, err
	}
	res, ok = data.(*pb.Metadata)
	if !ok {
		return res, errors.New("invalid type fetched")
	}
	return res, nil
}
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (*pb.SubscriptionMetadataList, error) {
	set := &pb.SubscriptionMetadataList{Metadatas: make([]*pb.Metadata, 0)}
	iterator, err := tx.Get(table, index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(*pb.Metadata)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		set.Metadatas = append(set.Metadatas, res)
	}
}
