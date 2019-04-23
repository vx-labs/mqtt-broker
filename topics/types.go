package topics

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/topics/ --go_out=plugins=grpc:. types.proto

type Store interface {
	Create(message *Metadata) error
	ByTopicPattern(tenant string, pattern []byte) (RetainedMessageMetadataList, error)
	All() (RetainedMessageMetadataList, error)
}

type retainedMessageFilter func(*Metadata) bool

func (set RetainedMessageMetadataList) Filter(filters ...retainedMessageFilter) RetainedMessageMetadataList {
	copy := make([]*Metadata, 0, len(set.Metadatas))
	for _, message := range set.Metadatas {
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
	return RetainedMessageMetadataList{
		Metadatas: copy,
	}
}

func (set RetainedMessageMetadataList) Apply(f func(s *Metadata)) {
	for _, message := range set.Metadatas {
		f(message)
	}
}

func (set RetainedMessageMetadataList) ApplyE(f func(s *Metadata) error) error {
	for _, message := range set.Metadatas {
		if err := f(message); err != nil {
			return err
		}
	}
	return nil
}

func HasID(id string) retainedMessageFilter {
	return func(s *Metadata) bool {
		return s.ID == id
	}
}
func MatchTopicPattern(pattern []byte) retainedMessageFilter {
	return func(s *Metadata) bool {
		t := NewTopic(s.Topic)
		return t.Match(pattern)
	}
}
func HasIDIn(set []string) retainedMessageFilter {
	wantedIDs := make(map[string]struct{}, len(set))
	for _, id := range set {
		wantedIDs[id] = struct{}{}
	}
	return func(s *Metadata) bool {
		_, ok := wantedIDs[s.ID]
		return ok
	}
}
func HasTenant(tenant string) retainedMessageFilter {
	return func(s *Metadata) bool {
		return s.Tenant == tenant
	}
}
func (set RetainedMessageMetadataList) ApplyIdx(f func(idx int, s *Metadata)) {
	for idx, session := range set.Metadatas {
		f(idx, session)
	}
}

func (s *Metadata) IsAdded() bool {
	return s.LastAdded > 0 && s.LastAdded > s.LastDeleted
}
func (s *Metadata) IsRemoved() bool {
	return s.LastDeleted > 0 && s.LastAdded < s.LastDeleted
}

func IsOutdated(s *Metadata, remote *Metadata) (outdated bool) {
	if s.LastAdded < remote.LastAdded {
		outdated = true
	}
	if s.LastDeleted < remote.LastDeleted {
		outdated = true
	}
	return outdated
}
