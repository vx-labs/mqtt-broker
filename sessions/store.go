package sessions

import (
	"github.com/vx-labs/mqtt-broker/sessions/pb"
	"go.uber.org/zap"

	memdb "github.com/hashicorp/go-memdb"
)

const (
	memdbTable = "sessions"
)
const (
	SessionCreated string = "session_created"
	SessionDeleted string = "session_deleted"
)

type SessionStore interface {
	ByID(id string) (*pb.Session, error)
	ByClientID(id string) (*pb.SessionMetadataList, error)
	ByPeer(peer string) (*pb.SessionMetadataList, error)
	All() (*pb.SessionMetadataList, error)
	Exists(id string) bool
	Delete(id string) error
	Create(sess *pb.Session) error
	Update(id string, op func(pb.Session) *pb.Session) error
}

type memDBStore struct {
	db     *memdb.MemDB
	logger *zap.Logger
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
	}
	return s
}
func (s *memDBStore) all() *pb.SessionMetadataList {
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
			sessionList.Sessions = append(sessionList.Sessions, sess)
		}
	})
}
func (s *memDBStore) All() (*pb.SessionMetadataList, error) {
	return s.all(), nil
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
			sessionList.Sessions = append(sessionList.Sessions, sess)
		}
	})
}

func (s *memDBStore) Create(sess *pb.Session) error {
	return s.insert(sess)
}
func (s *memDBStore) Update(id string, op func(pb.Session) *pb.Session) error {
	return s.write(func(tx *memdb.Txn) error {
		session, err := s.first(tx, "id", id)
		if err != nil {
			return err
		}
		return tx.Insert(memdbTable, op(*session))
	})
}
func (s *memDBStore) insert(sess *pb.Session) error {
	err := s.write(func(tx *memdb.Txn) error {
		return tx.Insert(memdbTable, sess)
	})
	return err
}
func (s *memDBStore) Delete(id string) error {
	err := s.write(func(tx *memdb.Txn) error {
		sess, err := s.first(tx, "id", id)
		if err != nil {
			return nil
		}
		return tx.Delete(memdbTable, sess)
	})
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
