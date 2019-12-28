package mesh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainsAll(t *testing.T) {
	require.True(t, containsAll([]string{"a", "b"}, []string{"a", "b"}))
	require.True(t, containsAll([]string{"b", "a"}, []string{"a", "b"}))
	require.True(t, containsAll([]string{"a", "b"}, []string{"b", "a"}))
	require.True(t, containsAll([]string{}, []string{}))
	require.True(t, containsAll([]string{}, nil))
	require.True(t, containsAll(nil, []string{}))
	require.True(t, containsAll(nil, nil))
	require.True(t, containsAll([]string{}, []string{"b", "a"}))
	require.True(t, containsAll(nil, []string{"b", "a"}))
}
