package subscriptions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	sub_1 = &Subscription{ID: "1", Pattern: []byte("devices/+/degrees")}
	sub_2 = &Subscription{ID: "2"}
	sub_3 = &Subscription{ID: "3"}
	sub_4 = &Subscription{ID: "4"}
)

func TestNode(t *testing.T) {
	tenant := "_default"
	t.Run("add subscription", func(t *testing.T) {
		a := NewNode(tenant, []byte("test"))
		a = a.AddSubscription(tenant, sub_1)
		require.Equal(t, 1, len(a.data))
		a = a.AddSubscription(tenant, sub_2)
		a = a.AddSubscription(tenant, sub_3)
		require.Equal(t, 3, len(a.data))
	})
	t.Run("del subscription", func(t *testing.T) {
		a := NewNode(tenant, []byte("test"))
		a.data = []*Subscription{sub_1, sub_2, sub_3}
		a = a.DelSubscription("2")
		require.Equal(t, 2, len(a.data))
	})
	t.Run("add subscription", func(t *testing.T) {
		a := NewNode(tenant, []byte("test"))
		a.inode.AddNode(NewNode(tenant, []byte("b")))
		require.Equal(t, []byte("b"), a.inode.nodes[0].pattern)
	})
	t.Run("insert", func(t *testing.T) {
		root := NewINode()
		root.Insert(Topic([]byte("devices/a/degrees")), tenant, sub_1)
		require.Equal(t, 1, len(root.nodes))
		require.Equal(t, []byte("devices"), root.nodes[0].pattern)
		require.Equal(t, 1, len(root.nodes[0].inode.nodes))
		require.Equal(t, []byte("a"), root.nodes[0].inode.nodes[0].pattern)
		require.Equal(t, 1, len(root.nodes[0].inode.nodes[0].inode.nodes))
		require.Equal(t, []byte("degrees"), root.nodes[0].inode.nodes[0].inode.nodes[0].pattern)
		require.Equal(t, 1, len(root.nodes[0].inode.nodes[0].inode.nodes[0].data))
		root.Insert(Topic([]byte("devices/a/degrees")), tenant, sub_2)
		require.Equal(t, 2, len(root.nodes[0].inode.nodes[0].inode.nodes[0].data))

	})

	t.Run("select", func(t *testing.T) {
		root := NewINode()
		topic := Topic([]byte("devices/a/degrees"))
		a := NewNode(tenant, []byte("devices"))
		b := NewNode(tenant, []byte("+"))
		c := NewNode(tenant, []byte("a"))
		a.inode.AddNode(b)
		a.inode.AddNode(c)

		d := NewNode(tenant, []byte("degrees"))
		e := NewNode(tenant, []byte("degrees"))

		d = d.AddSubscription(tenant, sub_1)
		e = e.AddSubscription(tenant, sub_2)

		b.inode.AddNode(d)
		c.inode.AddNode(e)
		root.AddNode(a)

		/*
			a (devices) -> b (+)  -> d (degrees)
			 	        `->  c (a)  -> e (degrees)
		*/
		set := root.Select(tenant, nil, topic)
		require.Equal(t, 2, len(set))
		require.NoError(t, root.Remove(tenant, sub_1.ID, Topic(sub_1.Pattern)))
		set = root.Select(tenant, nil, topic)
		require.Equal(t, 1, len(set))
	})

}

func BenchmarkMatchPattern(bench *testing.B) {
	token := []byte("device")
	pattern := []byte("device")
	for i := 0; i < bench.N; i++ {
		matchPattern(token, pattern)
	}
}
func BenchmarkNode(b *testing.B) {
	tenant := "_default"

	b.Run("select", func(bench *testing.B) {
		root := NewINode()
		topic := Topic([]byte("devices/a/temperature/degrees"))
		a := NewNode(tenant, []byte("devices"))
		b := NewNode(tenant, []byte("+"))
		c := NewNode(tenant, []byte("a"))
		a.inode.AddNode(b)
		a.inode.AddNode(c)

		d := NewNode(tenant, []byte("temperature"))
		e := NewNode(tenant, []byte("temperature"))

		f := NewNode(tenant, []byte("degrees"))
		g := NewNode(tenant, []byte("+"))

		f = f.AddSubscription(tenant, sub_1)
		g = g.AddSubscription(tenant, sub_2)

		b.inode.AddNode(d)
		c.inode.AddNode(e)
		d.inode.AddNode(f)
		e.inode.AddNode(g)
		root.AddNode(a)

		/*
			a (devices) ->  b (+)  -> d (temperature) -> f (degrees)
			 	          `-> c (a)  -> e (temperature) -> g (+)
		*/
		for i := 0; i < bench.N; i++ {
			root.Select(tenant, nil, topic)
		}
	})
}
