package pb

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/services/topics/pb/ --go_out=plugins=grpc:. types.proto

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
func (set RetainedMessageSet) ApplyIdx(f func(idx int, s RetainedMessage)) {
	for idx, session := range set {
		f(idx, session)
	}
}
