package store

import (
	"errors"

	"github.com/vx-labs/mqtt-broker/services/auth/pb"
)

var entities = []*pb.Entity{
	&pb.Entity{ID: "demo", Tenant: "demo", AuthProvider: "static-token", Permissions: []string{"*"}},
	&pb.Entity{ID: "vx:psk", Tenant: "vx:psk", AuthProvider: "static-token", Permissions: []string{"*"}},
}

var (
	ErrEntityNotFound = errors.New("entity not found")
)

type Static struct{}

func (_ *Static) ListTenants() []string {
	return []string{
		"_default",
		"demo",
	}
}
func (_ *Static) ListEntities() []*pb.Entity {
	return entities
}
func (_ *Static) ListEntitiesByTenant(tenant string) []*pb.Entity {
	out := []*pb.Entity{}
	for _, e := range entities {
		if e.Tenant == tenant {
			out = append(out, e)
		}
	}
	return out
}
func (_ *Static) GetEntitiesByID(tenant, id string) (*pb.Entity, error) {
	for _, e := range entities {
		if e.Tenant == tenant && e.ID == id {
			return e, nil
		}
	}
	return nil, ErrEntityNotFound
}
