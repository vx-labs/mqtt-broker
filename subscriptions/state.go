package subscriptions

import (
	"context"
	"io"
	"log"

	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/crdt"
	"github.com/vx-labs/mqtt-protocol/packet"
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
func (m *memDBStore) insertPBRemoteSubscription(sub Subscription, tx *memdb.Txn) error {
	return tx.Insert(table, sub)
}

func (m *memDBStore) newRemoteSubscription(remote Metadata) Subscription {
	return Subscription{
		Sender: func(ctx context.Context, p packet.Publish) error {
			return m.sender(remote.Peer, remote.SessionID, p)
		},
		Metadata: remote,
	}
}
func (m *memDBStore) Merge(inc []byte, _ bool) error {
	set := &SubscriptionMetadataList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	changeset := SubscriptionSet{}
	err = m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Metadatas {
			localData, err := tx.First(table, "id", remote.ID)
			if err != nil || localData == nil {
				sub := m.newRemoteSubscription(*remote)
				err := m.insertPBRemoteSubscription(sub, tx)
				if err != nil {
					return err
				}
				changeset = append(changeset, sub)
				continue
			}
			local, ok := localData.(Subscription)
			if !ok {
				log.Printf("WARN: invalid data found in subscription store")
				continue
			}
			if crdt.IsEntryOutdated(&local, remote) {
				sub := m.newRemoteSubscription(*remote)
				err := m.insertPBRemoteSubscription(sub, tx)
				if err != nil {
					return err
				}
				changeset = append(changeset, sub)
			}
		}
		return nil
	})
	for _, sub := range changeset {
		m.emitSubscriptionEvent(sub)
	}
	return nil
}
