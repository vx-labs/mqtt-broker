package topics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode_Upsert(t *testing.T) {
	n := NewNode("_root", "_all")
	child := n.Upsert("_default", NewTopic([]byte("devices/phones/nexus")))
	assert.Equal(t, "nexus", n.children["devices"].children["phones"].children["nexus"].Id)
	child.Id = "a"
	assert.Equal(t, "a", n.children["devices"].children["phones"].children["nexus"].Id)
}
func TestNode_Apply(t *testing.T) {
	n := NewNode("_root", "_all")
	n.Upsert("_default", NewTopic([]byte("devices/sensors/1")))
	n.Upsert("_default", NewTopic([]byte("devices/sensors/2")))
	n.Upsert("_default", NewTopic([]byte("devices/sensors/3")))

	n.Upsert("_default", NewTopic([]byte("planes/1/light/level")))
	n.Upsert("_default", NewTopic([]byte("planes/2/light/level")))
	n.Upsert("hidden", NewTopic([]byte("planes/4/light/level")))

	t.Run("simple", func(t *testing.T) {
		topics := []string{}
		n.Apply("_default", NewTopic([]byte("devices/sensors/1")), func(node *Node) bool {
			topics = append(topics, node.Id)
			return false
		})
		assert.Equal(t, 1, len(topics))
	})
	t.Run("slw", func(t *testing.T) {
		topics := []string{}
		n.Apply("_default", NewTopic([]byte("devices/+/1")), func(node *Node) bool {
			topics = append(topics, node.Id)
			return false
		})
		assert.Equal(t, 1, len(topics))
	})
	t.Run("slw multiple matches", func(t *testing.T) {
		topics := []string{}
		n.Apply("_default", NewTopic([]byte("devices/+/+")), func(node *Node) bool {
			topics = append(topics, node.Id)
			return false
		})
		assert.Equal(t, 3, len(topics))
	})
	t.Run("mlw", func(t *testing.T) {
		topics := []string{}
		n.Apply("_default", NewTopic([]byte("devices/#")), func(node *Node) bool {
			topics = append(topics, node.Id)
			return false
		})
		assert.Equal(t, 4, len(topics))
		assert.Equal(t, 4, len(topics))
	})
	t.Run("complex slw matches", func(t *testing.T) {
		topics := []string{}
		n.Apply("_default", NewTopic([]byte("planes/+/light/level")), func(node *Node) bool {
			topics = append(topics, node.Id)
			return false
		})
		assert.Equal(t, 2, len(topics))
	})

}

func BenchmarkNode(b *testing.B) {
	n := NewNode("_root", "_all")

	b.Run("insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for i := 0; i < 100; i++ {
				n.Upsert("_default", NewTopic([]byte(fmt.Sprintf("devices/%d/a", i))))
			}
		}
	})
	b.Run("apply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			n.Apply("_default", NewTopic([]byte("devices/+/a")), func(node *Node) bool {
				return false
			})
		}
	})
}
