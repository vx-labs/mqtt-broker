package subscriptions

import (
	"io"
	"log"

	proto "github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/crdt"
	"github.com/vx-labs/mqtt-broker/subscriptions/pb"
)

func (m memDBStore) dumpSubscriptions() *pb.SubscriptionMetadataList {
	RetainedMessageList := pb.SubscriptionMetadataList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(table, "id")
		if err != nil || iterator == nil {
			return nil
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*pb.Subscription)
			RetainedMessageList.Subscriptions = append(RetainedMessageList.Subscriptions, sess)
		}
	})
	return &RetainedMessageList
}

func (m memDBStore) MarshalBinary() []byte {
	set := m.dumpSubscriptions()
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
			sess := payload.(*pb.Subscription)
			return sess, nil
		}, func(id string) error {
			return tx.Delete(table, &pb.Subscription{ID: id})
		})
	})
}
func (m *memDBStore) insertPBRemoteSubscription(remote *pb.Subscription, tx *memdb.Txn) error {
	return tx.Insert(table, remote)
}
func (m *memDBStore) Merge(inc []byte, _ bool) error {
	set := &pb.SubscriptionMetadataList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	return m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Subscriptions {
			localData, err := tx.First(table, "id", remote.ID)
			if err != nil || localData == nil {
				err := m.insertPBRemoteSubscription(remote, tx)
				if err != nil {
					return err
				}
				continue
			}
			local, ok := localData.(*pb.Subscription)
			if !ok {
				log.Printf("WARN: invalid data found in store")
				continue
			}
			if crdt.IsEntryOutdated(local, remote) {
				err := m.insertPBRemoteSubscription(remote, tx)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
