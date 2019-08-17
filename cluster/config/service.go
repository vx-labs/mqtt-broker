package config

import "github.com/vx-labs/mqtt-broker/cluster/pb"

type Config struct {
	ID            string
	AdvertiseAddr string
	AdvertisePort int
	BindPort      int
	OnNodeJoin    func(id string, meta pb.NodeMeta)
	OnNodeLeave   func(id string, meta pb.NodeMeta)
}
