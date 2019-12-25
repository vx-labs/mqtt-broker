package sessions

import (
	"io"
	"log"

	proto "github.com/golang/protobuf/proto"
	memdb "github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/crdt"
	"github.com/vx-labs/mqtt-broker/services/sessions/pb"
)

var table = "sessions"

func (m memDBStore) dumpSessions() *pb.SessionMetadataList {
	RetainedMessageList := pb.SessionMetadataList{}
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
			sess := payload.(*pb.Session)
			RetainedMessageList.Sessions = append(RetainedMessageList.Sessions, sess)
		}
	})
	return &RetainedMessageList
}

func (m memDBStore) MarshalBinary() []byte {
	set := m.dumpSessions()
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
			sess := payload.(*pb.Session)
			return sess, nil
		}, func(id string) error {
			return tx.Delete(table, &pb.Session{ID: id})
		})
	})
}
func (m *memDBStore) insertPBRemoteSession(remote *pb.Session, tx *memdb.Txn) error {
	return tx.Insert(table, remote)
}
func (m *memDBStore) Merge(inc []byte, _ bool) error {
	set := &pb.SessionMetadataList{}
	err := proto.Unmarshal(inc, set)
	if err != nil {
		return err
	}
	return m.write(func(tx *memdb.Txn) error {
		for _, remote := range set.Sessions {
			localData, err := tx.First(table, "id", remote.ID)
			if err != nil || localData == nil {
				err := m.insertPBRemoteSession(remote, tx)
				if err != nil {
					return err
				}
				continue
			}
			local, ok := localData.(*pb.Session)
			if !ok {
				log.Printf("WARN: invalid data found in store")
				continue
			}
			if crdt.IsEntryOutdated(local, remote) {
				err := m.insertPBRemoteSession(remote, tx)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
