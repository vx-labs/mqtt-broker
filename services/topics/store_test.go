package topics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vx-labs/mqtt-broker/services/topics/pb"
)

func TestMemDB(t *testing.T) {
	db := NewMemDBStore()
	assert.Nil(t, db.Create(pb.RetainedMessage{
		ID:      "a1",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/a/temperature"),
	}))
	assert.Nil(t, db.Create(pb.RetainedMessage{
		ID:      "a2",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/b/temperature"),
	}))
	assert.Nil(t, db.Create(pb.RetainedMessage{
		ID:      "b1",
		Payload: []byte("bla"),
		Tenant:  "tenant2",
		Topic:   []byte("devices/c/temperature"),
	}))
	m, err := db.ByTenant("tenant1")
	assert.Nil(t, err)
	if !assert.Equal(t, 2, len(m)) {
		t.Fail()
		return
	}
	assert.Equal(t, "a1", m[0].ID)
	assert.Equal(t, "a2", m[1].ID)
	set, err := db.ByTopicPattern("tenant1", []byte("devices/+/+"))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(set))
	set, err = db.ByTopicPattern("tenant1", []byte("devices/#"))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(set))
}

func TestMemDBDump(t *testing.T) {
	db := NewMemDBStore()
	assert.Nil(t, db.Create(pb.RetainedMessage{
		ID:      "a1",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/a/temperature"),
	}))
	assert.Nil(t, db.Create(pb.RetainedMessage{
		ID:      "a2",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/b/temperature"),
	}))
	assert.Nil(t, db.Create(pb.RetainedMessage{
		ID:      "b1",
		Payload: []byte("bla"),
		Tenant:  "tenant2",
		Topic:   []byte("devices/c/temperature"),
	}))
}
