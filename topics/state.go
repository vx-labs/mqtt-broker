package topics

import (
	"io"
	"log"

	proto "github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/crdt"
)

func (m memDBStore) dumpRetainedMessages() *RetainedMessageMetadataList {
	RetainedMessageList := RetainedMessageMetadataList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("messages", "id")
		if err != nil || iterator == nil {
			return nil
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(RetainedMessage)
			RetainedMessageList.Metadatas = append(RetainedMessageList.Metadatas, &sess.Metadata)
		}
	})
	return &RetainedMessageList
}

func (m memDBStore) MarshalBinary() []byte {
	set := m.dumpRetainedMessages()
	payload, err := proto.Marshal(set)
	if err != nil {
		log.Printf("ERR: failed to marshal state: %v", err)
		return nil
	}
	return payload
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
			sess := payload.(RetainedMessage)
			return &sess, nil
		}, func(id string) error {
			return tx.Delete(table, RetainedMessage{
				Metadata: Metadata{ID: id},
			})
		},
		)
	})
}
func (m *memDBStore) insertPBRemoteSubscription(remote Metadata, tx *memdb.Txn) error {
	sub := RetainedMessage{
		Metadata: remote,
	}
	m.topicIndex.Index(sub)
	return tx.Insert(table, sub)
}
func (m *memDBStore) Merge(inc []byte, _ bool) error {
	set := &RetainedMessageMetadataList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	return m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Metadatas {
			localData, err := tx.First(table, "id", remote.ID)
			if err != nil || localData == nil {
				err := m.insertPBRemoteSubscription(*remote, tx)
				if err != nil {
					return err
				}
				continue
			}
			local, ok := localData.(RetainedMessage)
			if !ok {
				log.Printf("WARN: invalid data found in store")
				continue
			}
			if crdt.IsEntryOutdated(&local, remote) {
				err := m.insertPBRemoteSubscription(*remote, tx)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
