package sessions

import (
	"io"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/crdt"
)

func (m memDBStore) MarshalBinary() []byte {
	set := m.dumpSessions()
	payload, err := proto.Marshal(set)
	if err != nil {
		log.Printf("ERR: failed to marshal state: %v", err)
		return nil
	}
	return payload
}
func (m memDBStore) dumpSessions() *SessionMetadataList {
	sessionList := SessionMetadataList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(memdbTable, "id")
		if (err != nil && err != ErrSessionNotFound) || iterator == nil {
			return err
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(Session)
			sessionList.Metadatas = append(sessionList.Metadatas, &sess.Metadata)
		}
	})
	return &sessionList
}

func (m *memDBStore) runGC() error {
	return m.write(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(memdbTable, "id")
		if err != nil || iterator == nil {
			return err
		}
		return crdt.GCEntries(crdt.ExpireAfter8Hours(), func() (crdt.Entry, error) {
			payload := iterator.Next()
			if payload == nil {
				return nil, io.EOF
			}
			sess := payload.(Session)
			return &sess, nil
		}, func(id string) error {
			return tx.Delete(memdbTable, Session{
				Metadata: Metadata{ID: id},
			})
		},
		)
	})
}
func (m *memDBStore) insertPBRemoteSession(remote Metadata, tx *memdb.Txn) error {
	return tx.Insert(memdbTable, Session{
		Transport: m.remoteTransportProvider(remote.Peer, remote.ID),
		remote:    true,
		Metadata:  remote,
	})
}
func (m *memDBStore) Merge(inc []byte) error {
	now := time.Now()
	set := &SessionMetadataList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	log.Printf("DEBUG: starting remote session state merge")
	return m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Metadatas {
			localData, err := tx.First(memdbTable, "id", remote.ID)
			if err != nil || localData == nil {
				log.Printf("DEBUG: session merge: session %s does not exist localy, adding it", remote.ID)
				err := m.insertPBRemoteSession(*remote, tx)
				if err != nil {
					log.Printf("DEBUG: session merge: failed to add session %s: %v", remote.ID, err)
					return err
				}
				continue
			}
			local, ok := localData.(Session)
			if !ok {
				log.Printf("WARN: invalid data found in store")
				continue
			}
			if crdt.IsEntryOutdated(&local, remote) {
				log.Printf("DEBUG: session merge: session %s is outdated localy, replacing it", remote.ID)
				err := m.insertPBRemoteSession(*remote, tx)
				if err != nil {
					return err
				}
				if local.Transport != nil && !local.remote {
					log.Printf("DEBUG: session merge: closing local session %s", remote.ID)
					local.Transport.Close()
				}
			}
		}
		log.Printf("DEBUG: session merge done (%s elapsed)", time.Now().Sub(now).String())
		return nil
	})
}
