package topics

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/topics/ --go_out=plugins=grpc:. types.proto

type Store interface {
	Create(message *Metadata) error
	ByTopicPattern(tenant string, pattern []byte) (RetainedMessageMetadataList, error)
	All() (RetainedMessageMetadataList, error)
}

type retainedMessageFilter func(RetainedMessage) bool
type RetainedMessageSet []RetainedMessage

func (set RetainedMessageSet) Filter(filters ...retainedMessageFilter) RetainedMessageSet {
	copy := make(RetainedMessageSet, 0, len(set))
	for _, message := range set {
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
	return copy
}

func (set RetainedMessageSet) Apply(f func(s RetainedMessage)) {
	for _, message := range set {
		f(message)
	}
}

func (set RetainedMessageSet) ApplyE(f func(s RetainedMessage) error) error {
	for _, message := range set {
		if err := f(message); err != nil {
			return err
		}
	}
	return nil
}

func HasID(id string) retainedMessageFilter {
	return func(s RetainedMessage) bool {
		return s.ID == id
	}
}
func MatchTopicPattern(pattern []byte) retainedMessageFilter {
	return func(s RetainedMessage) bool {
		t := NewTopic(s.Topic)
		return t.Match(pattern)
	}
}
func HasIDIn(set []string) retainedMessageFilter {
	wantedIDs := make(map[string]struct{}, len(set))
	for _, id := range set {
		wantedIDs[id] = struct{}{}
	}
	return func(s RetainedMessage) bool {
		_, ok := wantedIDs[s.ID]
		return ok
	}
}
func HasTenant(tenant string) retainedMessageFilter {
	return func(s RetainedMessage) bool {
		return s.Tenant == tenant
	}
}
func (set RetainedMessageSet) ApplyIdx(f func(idx int, s RetainedMessage)) {
	for idx, session := range set {
		f(idx, session)
	}
}
