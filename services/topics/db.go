package topics

import (
	"errors"

	"github.com/vx-labs/mqtt-broker/crdt"
	"github.com/vx-labs/mqtt-broker/services/topics/pb"

	memdb "github.com/hashicorp/go-memdb"
)

func (m *memDBStore) do(write bool, f func(*memdb.Txn) error) error {
	tx := m.db.Txn(write)
	defer tx.Abort()
	return f(tx)
}
func (m *memDBStore) read(f func(*memdb.Txn) error) error {
	return m.do(false, f)
}
func (m *memDBStore) write(f func(*memdb.Txn) error) error {
	return m.do(true, f)
}

func (m *memDBStore) first(tx *memdb.Txn, index string, value ...interface{}) (pb.RetainedMessage, error) {
	var ok bool
	var res pb.RetainedMessage
	data, err := tx.First("messages", index, value...)
	if err != nil {
		return res, err
	}
	res, ok = data.(pb.RetainedMessage)
	if !ok {
		return res, errors.New("invalid type fetched")
	}
	return res, nil
}
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (pb.RetainedMessageSet, error) {
	var set pb.RetainedMessageSet
	iterator, err := tx.Get("messages", index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(pb.RetainedMessage)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		if crdt.IsEntryAdded(&res) {
			set = append(set, res)
		}
	}
}
