package sessions

import (
	"log"
	"time"

	"github.com/vx-labs/mqtt-broker/crdt"

	"github.com/golang/protobuf/proto"
	ap "github.com/vx-labs/mqtt-broker/adapters/ap/pb"
	"github.com/vx-labs/mqtt-broker/services/sessions/pb"
	"go.uber.org/zap"

	memdb "github.com/hashicorp/go-memdb"
)

const (
	memdbTable = "sessions"
)

type Channel interface {
	Broadcast([]byte)
}

type SessionStore interface {
	ap.APState
	ByID(id string) (*pb.Session, error)
	ByClientID(id string) (*pb.SessionMetadataList, error)
	ByPeer(peer string) (*pb.SessionMetadataList, error)
	All(*pb.SessionFilterInput) (*pb.SessionMetadataList, error)
	Exists(id string) bool
	Delete(id string) error
	Create(sess *pb.Session) error
	Update(id string, op func(pb.Session) *pb.Session) error
}

type memDBStore struct {
	db     *memdb.MemDB
	logger *zap.Logger
	events chan []byte
}

func NewSessionStore(logger *zap.Logger) SessionStore {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			memdbTable: {
				Name: memdbTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name: "id",
						Indexer: &memdb.StringFieldIndex{
							Field: "ID",
						},
						Unique:       true,
						AllowMissing: false,
					},
					"tenant": {
						Name: "tenant",
						Indexer: &memdb.StringFieldIndex{
							Field: "Tenant",
						},
						Unique:       false,
						AllowMissing: false,
					},
					"client_id": {
						Name: "client_id",
						Indexer: &memdb.StringFieldIndex{
							Field: "ClientID",
						},
						Unique:       false,
						AllowMissing: false,
					},
					"peer": &memdb.IndexSchema{
						Name:         "peer",
						AllowMissing: false,
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "Peer"},
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	s := &memDBStore{
		db:     db,
		logger: logger,
		events: make(chan []byte),
	}
	go func() {
		for range time.Tick(1 * time.Hour) {
			err := s.runGC()
			if err != nil {
				log.Printf("WARN: failed to GC sessions: %v", err)
			}
		}
	}()
	return s
}
func (s *memDBStore) Events() chan []byte {
	return s.events
}
func (s *memDBStore) all(filter *pb.SessionFilterInput) *pb.SessionMetadataList {
	sessionList := &pb.SessionMetadataList{Sessions: make([]*pb.Session, 0)}
	s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(memdbTable, "id")
		if err != nil || iterator == nil {
			return pb.ErrSessionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*pb.Session)
			if crdt.IsEntryRemoved(sess) {
				continue
			}
			if filter != nil {
				filterMatched := false
				if filter.ID != nil {
					for _, wanted := range filter.ID {
						if sess.ID == wanted {
							filterMatched = true
							break
						}
					}
					if !filterMatched {
						continue
					}
				}
			}
			sessionList.Sessions = append(sessionList.Sessions, sess)
		}
	})
	return sessionList
}

func (s *memDBStore) Exists(id string) bool {
	_, err := s.ByID(id)
	return err == nil
}
func (s *memDBStore) ByID(id string) (*pb.Session, error) {
	var session *pb.Session
	return session, s.read(func(tx *memdb.Txn) error {
		sess, err := s.first(tx, "id", id)
		if err != nil {
			return err
		}
		session = sess
		if crdt.IsEntryRemoved(sess) {
			return pb.ErrSessionNotFound
		}
		return nil
	})
}
func (s *memDBStore) ByClientID(id string) (*pb.SessionMetadataList, error) {
	sessionList := &pb.SessionMetadataList{Sessions: make([]*pb.Session, 0)}
	return sessionList, s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(memdbTable, "client_id", id)
		if err != nil || iterator == nil {
			return pb.ErrSessionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*pb.Session)
			if crdt.IsEntryRemoved(sess) {
				continue
			}
			sessionList.Sessions = append(sessionList.Sessions, sess)
		}
	})
}
func (s *memDBStore) All(f *pb.SessionFilterInput) (*pb.SessionMetadataList, error) {
	return s.all(f), nil
}

func (s *memDBStore) ByPeer(peer string) (*pb.SessionMetadataList, error) {
	sessionList := &pb.SessionMetadataList{Sessions: make([]*pb.Session, 0)}
	return sessionList, s.read(func(tx *memdb.Txn) error {
		iterator, err := tx.Get(memdbTable, "peer", peer)
		if err != nil || iterator == nil {
			return pb.ErrSessionNotFound
		}
		for {
			payload := iterator.Next()
			if payload == nil {
				return nil
			}
			sess := payload.(*pb.Session)
			if crdt.IsEntryRemoved(sess) {
				continue
			}
			sessionList.Sessions = append(sessionList.Sessions, sess)
		}
	})
}

func (s *memDBStore) notify(b []byte) {
	select {
	case s.events <- b:
	default:
	}
}
func (s *memDBStore) Create(sess *pb.Session) error {
	sess.LastAdded = time.Now().UnixNano()
	err := s.write(func(tx *memdb.Txn) error {
		return tx.Insert(memdbTable, sess)
	})
	if err == nil {
		buf, err := proto.Marshal(&pb.SessionMetadataList{
			Sessions: []*pb.Session{
				sess,
			},
		})
		if err != nil {
			return err
		}
		s.notify(buf)
	}
	return err
}
func (s *memDBStore) Update(id string, op func(pb.Session) *pb.Session) error {
	var session *pb.Session
	err := s.write(func(tx *memdb.Txn) error {
		var err error
		session, err = s.first(tx, "id", id)
		if err != nil {
			return err
		}
		session.LastAdded = time.Now().UnixNano()
		return tx.Insert(memdbTable, op(*session))
	})
	if err == nil {
		buf, err := proto.Marshal(&pb.SessionMetadataList{
			Sessions: []*pb.Session{
				session,
			},
		})
		if err != nil {
			return err
		}
		s.notify(buf)
	}
	return err
}

func (s *memDBStore) Delete(id string) error {
	var session *pb.Session
	err := s.write(func(tx *memdb.Txn) error {
		var err error
		session, err = s.first(tx, "id", id)
		if err != nil {
			return err
		}
		session.LastDeleted = time.Now().UnixNano()
		return tx.Insert(memdbTable, session)
	})
	if err == nil {
		buf, err := proto.Marshal(&pb.SessionMetadataList{
			Sessions: []*pb.Session{
				session,
			},
		})
		if err != nil {
			return err
		}
		s.notify(buf)
	}
	return err
}

func (s *memDBStore) read(statement func(tx *memdb.Txn) error) error {
	tx := s.db.Txn(false)
	return s.run(tx, statement)
}
func (s *memDBStore) write(statement func(tx *memdb.Txn) error) error {
	tx := s.db.Txn(true)
	return s.run(tx, statement)
}
func (s *memDBStore) run(tx *memdb.Txn, statement func(tx *memdb.Txn) error) error {
	defer tx.Abort()
	err := statement(tx)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (s *memDBStore) first(tx *memdb.Txn, idx, id string) (*pb.Session, error) {
	data, err := tx.First(memdbTable, idx, id)
	if err != nil || data == nil {
		return nil, pb.ErrSessionNotFound
	}
	sess := data.(*pb.Session)
	return sess, nil
}
