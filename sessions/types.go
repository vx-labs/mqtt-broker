package sessions

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/sessions/ --go_out=plugins=grpc:. types.proto

type sessionFilter func(*Session) bool

func (set SessionList) Filter(filters ...sessionFilter) SessionList {
	copy := make([]*Session, 0, len(set.Sessions))
	for _, session := range set.Sessions {
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
	return SessionList{
		Sessions: copy,
	}
}
func (set SessionList) Apply(f func(s *Session)) {
	for _, session := range set.Sessions {
		f(session)
	}
}
func (set SessionList) ApplyIdx(f func(idx int, s *Session)) {
	for idx, session := range set.Sessions {
		f(idx, session)
	}
}

func (set SessionList) ApplyE(f func(s *Session) error) error {
	for _, session := range set.Sessions {
		if err := f(session); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) IsAdded() bool {
	return s.LastUpdated > 0 && s.LastUpdated > s.LastDeleted
}
func (s *Session) IsRemoved() bool {
	return s.LastDeleted > 0 && s.LastUpdated < s.LastDeleted
}

func IsOutdated(s *Session, remote *Session) (outdated bool) {
	if s.LastUpdated < remote.LastUpdated {
		outdated = true
	}
	if s.LastDeleted < remote.LastDeleted {
		outdated = true
	}
	return outdated
}
