package topics

import (
	"errors"

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

func (m *memDBStore) first(tx *memdb.Txn, index string, value ...interface{}) (*Metadata, error) {
	var ok bool
	var res *Metadata
	data, err := tx.First("messages", index, value...)
	if err != nil {
		return res, err
	}
	res, ok = data.(*Metadata)
	if !ok {
		return res, errors.New("invalid type fetched")
	}
	return res, nil
}
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (RetainedMessageMetadataList, error) {
	var set RetainedMessageMetadataList
	iterator, err := tx.Get("messages", index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(*Metadata)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		if res.IsAdded() {
			set.Metadatas = append(set.Metadatas, res)
		}
	}
}
