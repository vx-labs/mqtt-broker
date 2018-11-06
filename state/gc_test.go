package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGC(t *testing.T) {
	back := MockedBackend{
		"1": &MockedEntry{
			ID:          "1",
			LastDeleted: 10,
		},
		"2": &MockedEntry{
			ID:          "2",
			LastAdded:   15,
			LastDeleted: 11,
		},
		"3": &MockedEntry{
			ID:        "3",
			LastAdded: 1,
		},
	}
	store := Store{
		backend: back,
	}
	store.gC(12)
	assert.Nil(t, back["1"])
	assert.NotNil(t, back["2"])
	assert.NotNil(t, back["3"])
}
