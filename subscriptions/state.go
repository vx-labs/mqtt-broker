package subscriptions

import (
	"io"
	"log"

	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/crdt"
)

func (m memDBStore) MarshalBinary() []byte {
	set := m.dumpSubscriptions()
	payload, err := proto.Marshal(set)
	if err != nil {
		log.Printf("ERR: failed to marshal state: %v", err)
		return nil
	}
	return payload
}
func (m memDBStore) dumpSubscriptions() *SubscriptionMetadataList {
	sessionList := SubscriptionMetadataList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(table, "id")
		if (err != nil && err != ErrSubscriptionNotFound) || iterator == nil {
			return err
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(Subscription)
			sessionList.Metadatas = append(sessionList.Metadatas, &sess.Metadata)
		}
	})
	return &sessionList
}

func (m *memDBStore) runGC() error {
	return m.write(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(table, "id")
		if err != nil || iterator == nil {
			return err
		}
		return crdt.GCEntries(crdt.ExpireAfter8Hours(), func() (crdt.Entry, error) {
			payload := iterator.Next()
			if payload == nil {
				return nil, io.EOF
			}
			sess := payload.(Subscription)
			return &sess, nil
		}, func(id string) error {
			return tx.Delete(table, Subscription{
				Metadata: Metadata{ID: id},
			})
		},
		)
	})
}
func insertPBRemoteSession(remote Metadata, tx *memdb.Txn) error {
	return tx.Insert(table, Subscription{
		Close: func() error {
			log.Printf("WARN: tried to close a remote session")
			return nil
		},
		Metadata: remote,
	})
}
func (m *memDBStore) Merge(inc []byte) error {
	set := &SubscriptionMetadataList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	return m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Metadatas {
			localData, err := tx.First(table, "id", remote.ID)
			if err != nil || localData == nil {
				err := insertPBRemoteSession(*remote, tx)
				if err != nil {
					return err
				}
				continue
			}
			local, ok := localData.(Subscription)
			if !ok {
				log.Printf("WARN: invalid data found in store")
				continue
			}
			if crdt.IsEntryOutdated(&local, remote) {
				err := insertPBRemoteSession(*remote, tx)
				if err != nil {
					return err
				}
				local.Close()
			}
		}
		return nil
	})
}
