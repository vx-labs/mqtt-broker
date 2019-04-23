package topics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vx-labs/mqtt-broker/broker/cluster"
)

func TestMemDB(t *testing.T) {
	db, err := NewMemDBStore(cluster.MockedMesh())
	assert.Nil(t, err)
	assert.Nil(t, db.Create(&Metadata{
		ID:      "a1",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/a/temperature"),
	}))
	assert.Nil(t, db.Create(&Metadata{
		ID:      "a2",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/b/temperature"),
	}))
	assert.Nil(t, db.Create(&Metadata{
		ID:      "b1",
		Payload: []byte("bla"),
		Tenant:  "tenant2",
		Topic:   []byte("devices/c/temperature"),
	}))
	m, err := db.ByTenant("tenant1")
	assert.Nil(t, err)
	if !assert.Equal(t, 2, len(m.Metadatas)) {
		t.Fail()
		return
	}
	assert.Equal(t, "a1", m.Metadatas[0].ID)
	assert.Equal(t, "a2", m.Metadatas[1].ID)
	set, err := db.ByTopicPattern("tenant1", []byte("devices/+/+"))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(set.Metadatas))
	set, err = db.ByTopicPattern("tenant1", []byte("devices/#"))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(set.Metadatas))
}

func TestMemDBDump(t *testing.T) {
	db, err := NewMemDBStore(cluster.MockedMesh())
	assert.Nil(t, err)
	assert.Nil(t, db.Create(&Metadata{
		ID:      "a1",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/a/temperature"),
	}))
	assert.Nil(t, db.Create(&Metadata{
		ID:      "a2",
		Payload: []byte("bla"),
		Tenant:  "tenant1",
		Topic:   []byte("devices/b/temperature"),
	}))
	assert.Nil(t, db.Create(&Metadata{
		ID:      "b1",
		Payload: []byte("bla"),
		Tenant:  "tenant2",
		Topic:   []byte("devices/c/temperature"),
	}))
}
