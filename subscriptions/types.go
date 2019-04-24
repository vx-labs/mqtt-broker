package subscriptions

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/subscriptions/ --go_out=plugins=grpc:. types.proto

type SubscriptionFilter func(Subscription) bool
type SubscriptionSet []Subscription

func (set SubscriptionSet) Filter(filters ...SubscriptionFilter) SubscriptionSet {
	copy := make(SubscriptionSet, 0, cap(set))
	for _, sub := range set {
		matched := true
		for _, f := range filters {
			if !f(sub) {
				matched = false
			}
		}
		if matched {
			copy = append(copy, sub)
		}
	}
	return copy
}

func (set SubscriptionSet) Apply(f func(Subscription)) {
	for _, sub := range set {
		f(sub)
	}
}
func (set SubscriptionSet) ApplyE(f func(Subscription) error) (err error) {
	for _, sub := range set {
		err = f(sub)
		if err != nil {
			break
		}
	}
	return err
}

func (set SubscriptionSet) ApplyIdx(f func(idx int, s Subscription)) {
	for idx, session := range set {
		f(idx, session)
	}
}
