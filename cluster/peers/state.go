package peers

import (
	"io"
	"log"

	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/cluster/pb"
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
func (m memDBStore) dumpSubscriptions() *pb.PeerMetadataList {
	sessionList := pb.PeerMetadataList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(peerTable, "id")
		if (err != nil && err != ErrPeerNotFound) || iterator == nil {
			return err
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(Peer)
			sessionList.Metadatas = append(sessionList.Metadatas, &sess.Metadata)
		}
	})
	return &sessionList
}

func (m *memDBStore) runGC() error {
	return m.write(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(peerTable, "id")
		if err != nil || iterator == nil {
			return err
		}
		return crdt.GCEntries(crdt.ExpireAfter8Hours(), func() (crdt.Entry, error) {
			payload := iterator.Next()
			if payload == nil {
				return nil, io.EOF
			}
			sess := payload.(Peer)
			return &sess, nil
		}, func(id string) error {
			return tx.Delete(peerTable, Peer{
				Metadata: pb.Metadata{ID: id},
			})
		},
		)
	})
}

func (m *memDBStore) newRemotePeer(remote pb.Metadata) Peer {
	return Peer{
		Metadata: remote,
	}
}
func (m *memDBStore) insertPBRemoteSubscription(sub Peer, tx *memdb.Txn) error {
	return tx.Insert(peerTable, sub)
}
func (m *memDBStore) Merge(inc []byte, _ bool) error {
	set := &pb.PeerMetadataList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	changedPeers := SubscriptionSet{}
	err = m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Metadatas {
			localData, err := tx.First(peerTable, "id", remote.ID)
			if err != nil || localData == nil {
				peer := m.newRemotePeer(*remote)
				err := m.insertPBRemoteSubscription(peer, tx)
				if err != nil {
					return err
				}
				changedPeers = append(changedPeers, peer)
				continue
			}
			local, ok := localData.(Peer)
			if !ok {
				log.Printf("WARN: invalid data found in store")
				continue
			}
			if crdt.IsEntryOutdated(&local, remote) {
				peer := m.newRemotePeer(*remote)
				err := m.insertPBRemoteSubscription(peer, tx)
				if err != nil {
					return err
				}
				changedPeers = append(changedPeers, peer)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, peer := range changedPeers {
		m.emitPeerEvent(peer)
	}
	return nil
}
