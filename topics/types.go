package topics

import "github.com/vx-labs/mqtt-broker/topics/pb"

type Store interface {
	Create(message *pb.RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (pb.RetainedMessageMetadataList, error)
	All() (pb.RetainedMessageMetadataList, error)
}
