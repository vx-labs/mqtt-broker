package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {
	i := identity{
		private: &address{
			host: "127.0.0.1",
			port: 10001,
		},
		public: &address{
			host: "192.0.2.1",
			port: 10000,
		},
	}
	assert.Equal(t, "127.0.0.1", i.Private().Host())
	assert.Equal(t, "192.0.2.1", i.Public().Host())
	j := i.WithID("1")
	assert.Equal(t, "1", j.ID())

}
