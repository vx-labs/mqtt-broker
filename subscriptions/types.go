package subscriptions

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/subscriptions/ --go_out=plugins=grpc:. types.proto

type SubscriptionList []*Subscription
type SubscriptionFilter func(*Subscription) bool

func (set SubscriptionList) Filter(filters ...SubscriptionFilter) SubscriptionList {
	copy := make(SubscriptionList, 0, cap(set))
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

func (set SubscriptionList) Apply(f func(*Subscription)) {
	for _, sub := range set {
		f(sub)
	}
}
func (set SubscriptionList) ApplyE(f func(*Subscription) error) (err error) {
	for _, sub := range set {
		err = f(sub)
		if err != nil {
			break
		}
	}
	return err
}
