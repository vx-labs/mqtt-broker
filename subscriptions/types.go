package subscriptions

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/subscriptions/ --go_out=plugins=grpc:. types.proto

type SubscriptionFilter func(*Subscription) bool

func (set SubscriptionList) Filter(filters ...SubscriptionFilter) SubscriptionList {
	copy := make([]*Subscription, 0, cap(set.Subscriptions))
	for _, sub := range set.Subscriptions {
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
	return SubscriptionList{
		Subscriptions: copy,
	}
}

func (set SubscriptionList) Apply(f func(*Subscription)) {
	for _, sub := range set.Subscriptions {
		f(sub)
	}
}
func (set SubscriptionList) ApplyE(f func(*Subscription) error) (err error) {
	for _, sub := range set.Subscriptions {
		err = f(sub)
		if err != nil {
			break
		}
	}
	return err
}

func (set SubscriptionList) ApplyIdx(f func(idx int, s *Subscription)) {
	for idx, session := range set.Subscriptions {
		f(idx, session)
	}
}

func (s *Subscription) IsAdded() bool {
	return s.LastAdded > 0 && s.LastAdded > s.LastDeleted
}
func (s *Subscription) IsRemoved() bool {
	return s.LastDeleted > 0 && s.LastAdded < s.LastDeleted
}

func IsOutdated(s *Subscription, remote *Subscription) (outdated bool) {
	if s.LastAdded < remote.LastAdded {
		outdated = true
	}
	if s.LastDeleted < remote.LastDeleted {
		outdated = true
	}
	return outdated
}
