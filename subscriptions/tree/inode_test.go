package tree

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vx-labs/mqtt-broker/subscriptions/topic"
)

func TestINode(t *testing.T) {
	i := NewINode()
	testTopic := topic.Topic([]byte("devices/a/degrees"))
	i.Insert(topic.Topic(sub_1.Pattern), "default", sub_1)
	i.Insert(topic.Topic(sub_1.Pattern), "default", sub_1)
	i.Insert(topic.Topic(sub_0.Pattern), "default", sub_0)
	results := i.Select("default", nil, testTopic)
	require.Equal(t, 2, len(results))
	require.Equal(t, sub_1.ID, results[0].ID)
	require.Equal(t, sub_0.ID, results[1].ID)
}
