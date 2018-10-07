package sessions

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/sessions/ --go_out=plugins=grpc:. types.proto

type SessionList []*Session
type sessionFilter func(*Session) bool

func (set SessionList) Filter(filters ...sessionFilter) SessionList {
	copy := make([]*Session, 0, len(set))
	for _, session := range set {
		accepted := true
		for _, f := range filters {
			if !f(session) {
				accepted = false
				break
			}
		}
		if accepted {
			copy = append(copy, session)
		}
	}
	return copy
}
func (set SessionList) Apply(f func(s *Session)) {
	for _, session := range set {
		f(session)
	}
}

func (set SessionList) ApplyE(f func(s *Session) error) error {
	for _, session := range set {
		if err := f(session); err != nil {
			return err
		}
	}
	return nil
}
