package identity

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNomadService(t *testing.T) {
	os.Setenv("NOMAD_IP_test", "192.0.2.1")
	os.Setenv("NOMAD_HOST_PORT_test", "10001")
	os.Setenv("NOMAD_PORT_test", "10002")

	s, err := NomadService("test")
	assert.Nil(t, err)
	assert.Equal(t, "0455d566d1faaa9bb72b9e783cdba539324b8fd8", s.ID())
	assert.Equal(t, "192.0.2.1", s.Public().Host())
	assert.Equal(t, 10001, s.Public().Port())
}
