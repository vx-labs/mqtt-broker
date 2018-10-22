package topics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetainedMessageList(t *testing.T) {
	local := &RetainedMessageList{
		RetainedMessages: []*RetainedMessage{
			&RetainedMessage{
				Id:          "1",
				LastUpdated: 1,
			},
			&RetainedMessage{
				Id:          "2",
				Topic:       []byte("1"),
				LastUpdated: 1,
			},
		},
	}
	remote := &RetainedMessageList{
		RetainedMessages: []*RetainedMessage{
			&RetainedMessage{
				Id:          "2",
				Topic:       []byte("2"),
				LastUpdated: 2,
			},
		},
	}
	data := local.Merge(remote)
	delta := data.(*RetainedMessageList)
	require.Equal(t, 1, len(delta.RetainedMessages))
	require.Equal(t, []byte("2"), delta.RetainedMessages[0].Topic)
	require.Equal(t, []byte("2"), local.RetainedMessages[1].Topic)
}
