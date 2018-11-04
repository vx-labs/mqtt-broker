package topics

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/topics/ --go_out=plugins=grpc:. types.proto

type Store interface {
	Create(message *RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (RetainedMessageList, error)
	All() (RetainedMessageList, error)
}

type retainedMessageFilter func(*RetainedMessage) bool

func (set RetainedMessageList) Filter(filters ...retainedMessageFilter) RetainedMessageList {
	copy := make([]*RetainedMessage, 0, len(set.RetainedMessages))
	for _, message := range set.RetainedMessages {
		accepted := true
		for _, f := range filters {
			if !f(message) {
				accepted = false
				break
			}
		}
		if accepted {
			copy = append(copy, message)
		}
	}
	return RetainedMessageList{
		RetainedMessages: copy,
	}
}

func (set RetainedMessageList) Apply(f func(s *RetainedMessage)) {
	for _, message := range set.RetainedMessages {
		f(message)
	}
}

func (set RetainedMessageList) ApplyE(f func(s *RetainedMessage) error) error {
	for _, message := range set.RetainedMessages {
		if err := f(message); err != nil {
			return err
		}
	}
	return nil
}

func HasID(id string) retainedMessageFilter {
	return func(s *RetainedMessage) bool {
		return s.ID == id
	}
}
func MatchTopicPattern(pattern []byte) retainedMessageFilter {
	return func(s *RetainedMessage) bool {
		t := NewTopic(s.Topic)
		return t.Match(pattern)
	}
}
func HasIDIn(set []string) retainedMessageFilter {
	wantedIDs := make(map[string]struct{}, len(set))
	for _, id := range set {
		wantedIDs[id] = struct{}{}
	}
	return func(s *RetainedMessage) bool {
		_, ok := wantedIDs[s.ID]
		return ok
	}
}
func HasTenant(tenant string) retainedMessageFilter {
	return func(s *RetainedMessage) bool {
		return s.Tenant == tenant
	}
}
func (set RetainedMessageList) ApplyIdx(f func(idx int, s *RetainedMessage)) {
	for idx, session := range set.RetainedMessages {
		f(idx, session)
	}
}

func (s *RetainedMessage) IsAdded() bool {
	return s.LastAdded > 0 && s.LastAdded > s.LastDeleted
}
func (s *RetainedMessage) IsRemoved() bool {
	return s.LastDeleted > 0 && s.LastAdded < s.LastDeleted
}

func IsOutdated(s *RetainedMessage, remote *RetainedMessage) (outdated bool) {
	if s.LastAdded < remote.LastAdded {
		outdated = true
	}
	if s.LastDeleted < remote.LastDeleted {
		outdated = true
	}
	return outdated
}
