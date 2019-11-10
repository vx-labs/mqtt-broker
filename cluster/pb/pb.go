package pb

import "net/url"

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/cluster/pb/ --go_out=plugins=grpc:. pb.proto

func GetTagValue(needle string, stack []*ServiceTag) string {
	for _, tag := range stack {
		if needle == tag.Key {
			return tag.Value
		}
	}
	return ""
}

func ParseFilter(values url.Values) []*ServiceTag {
	out := make([]*ServiceTag, len(values))
	idx := 0
	for k, v := range values {
		out[idx] = &ServiceTag{Key: k, Value: v[0]}
		idx++
	}
	return out
}

func MatchFilter(needles []*ServiceTag, slice []*ServiceTag) bool {
	if needles == nil || len(needles) == 0 {
		return true
	}
	for _, needle := range needles {
		if GetTagValue(needle.Key, slice) != needle.Value {
			return false
		}
	}
	return true
}
