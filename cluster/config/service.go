package config

type Config struct {
	ID            string
	AdvertiseAddr string
	AdvertisePort int
	BindPort      int
	JoinList      []string
	OnNodeJoin    func(id string, meta []byte)
	OnNodeLeave   func(id string, meta []byte)
	OnNodeUpdate  func(id string, meta []byte)
}
