package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress(t *testing.T) {
	a := &address{
		host: "192.0.2.1",
		port: 4555,
	}
	assert.Equal(t, 4555, a.Port())
	assert.Equal(t, "192.0.2.1", a.Host())
	assert.Equal(t, "192.0.2.1:4555", a.String())
}
func BenchmarkAddressString(b *testing.B) {
	a := &address{
		host: "192.0.2.1",
		port: 4555,
	}
	for i := 0; i < b.N; i++ {
		_ = a.String()
	}
}
