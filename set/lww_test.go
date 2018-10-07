package set

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLWW(t *testing.T) {
	lww := NewLWW()
	require.NoError(t, lww.Set("test"))
	require.True(t, lww.IsSet("test"))
	require.False(t, lww.IsRemoved("test"))
	count := 0
	require.Nil(t, lww.Iterate(func(_ string, e Entry) error {
		count++
		return nil
	}))
	require.Equal(t, 1, count)

	lww2 := NewLWW()
	lww2.Set("test2")
	lww2.Remove("test")
	delta := lww.Merge(lww2)
	require.True(t, delta.IsSet("test2"))
	require.True(t, delta.IsRemoved("test"))
	require.True(t, lww.IsSet("test2"))
	require.True(t, lww.IsRemoved("test"))
	buf := lww.Serialize()
	lww3, err := DecodeLWW(buf)
	require.NoError(t, err)
	require.True(t, lww3.IsSet("test2"))
	require.True(t, lww3.IsRemoved("test"))
}

func BenchmarkLWW(b *testing.B) {
	b.Run("set", func(b *testing.B) {
		lww := NewLWW()
		for i := 0; i < b.N; i++ {
			lww.Set("test")
		}
	})
	b.Run("isSet", func(b *testing.B) {
		lww := NewLWW()
		lww.Set("test")
		for i := 0; i < b.N; i++ {
			lww.IsSet("test")
		}
	})
}
