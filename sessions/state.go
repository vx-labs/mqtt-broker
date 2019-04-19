package sessions

import (
	"io"
	"log"

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
func (m memDBStore) dumpSessions() *SessionList {
	sessionList := SessionList{}
	m.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("sessions", "id")
		if (err != nil && err != ErrSessionNotFound) || iterator == nil {
			return err
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(Session)
			sessionList.Sessions = append(sessionList.Sessions, &sess)
		}
	})
	return &sessionList
}

func (m *memDBStore) runGC() error {
	return m.write(func(tx *memdb.Txn) error {
		iterator, err := tx.Get("sessions", "id")
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
			return tx.Delete("sessions", Session{ID: id})
		},
		)
	})
}

func (m *memDBStore) Merge(inc []byte) error {
	set := &SessionList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	return m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Sessions {
			localData, err := tx.First("sessions", "id", remote.ID)
			if err != nil || localData == nil {
				err := tx.Insert("sessions", *remote)
				if err != nil {
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
				err := tx.Insert("sessions", *remote)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
