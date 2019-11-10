package pb

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopic_Chop(t *testing.T) {
	n := Topic{
		payload: []byte("devices/phones/nexus"),
	}
	topic, token, ok := n.Chop()
	assert.Equal(t, "devices", token)
	assert.Equal(t, []byte("phones/nexus"), topic.payload)
	assert.True(t, ok)

}
func TestTopic_Count(t *testing.T) {
	n := Topic{
		payload: []byte("devices/phones/nexus"),
	}
	assert.Equal(t, 3, n.Length())

}
func TestTopic_Next(t *testing.T) {
	path := []byte("devices/phones/nexus")
	n := NewTopic(path)
	assert.Equal(t, []byte{}, n.Head())
	assert.Equal(t, path, n.Tail())
	assert.Nil(t, n.Next())
	assert.Equal(t, "devices", string(n.Cur()))
	assert.Equal(t, "", string(n.Head()))
	assert.Equal(t, "phones/nexus", string(n.Tail()))
	assert.Nil(t, n.Next())
	assert.Equal(t, "phones", string(n.Cur()))
	assert.Equal(t, "devices/", string(n.Head()))
	assert.Equal(t, "nexus", string(n.Tail()))

	assert.Nil(t, n.Next())
	assert.Equal(t, "nexus", string(n.Cur()))
	assert.Equal(t, "devices/phones/", string(n.Head()))
	assert.Equal(t, "", string(n.Tail()))
	assert.Equal(t, io.EOF, n.Next())
}

func TestTopic_Match(t *testing.T) {
	n := NewTopic([]byte("devices/phones/nexus"))
	assert.True(t, n.Match([]byte("devices/phones/nexus")))
	assert.False(t, n.Match([]byte("devices/phones/nexus/jambon")))
	assert.True(t, n.Match([]byte("devices/+/nexus")))
	assert.True(t, n.Match([]byte("devices/#")))
	assert.False(t, n.Match([]byte("devices/+/nexus/#")))
	assert.False(t, n.Match([]byte("device/+/nexus")))
}
func BenchmarkTopic_Match(b *testing.B) {
	n := NewTopic([]byte("devices/phones/nexus/a/b/f/a/d/de/a/adzada/f"))
	for i := 0; i < b.N; i++ {
		n.Match([]byte("devices/phones/nexus/a/b/f/a/d/de/a/adzada/f"))
	}
}
