package identity

type nomadIdentity struct {
	id                string
	name              string
	tag               string
	advertisedAddress string
	advertisedPort    int
	bindAddress       string
	bindPort          int
}

func (c *nomadIdentity) Name() string {
	return c.name
}
func (c *nomadIdentity) ID() string {
	return c.id
}
func (c *nomadIdentity) Tag() string {
	return c.tag
}
func (c *nomadIdentity) AdvertisedAddress() string {
	return c.advertisedAddress
}
func (c *nomadIdentity) AdvertisedPort() int {
	if c.bindPort == 0 {
		panic("invalid advertised port: 0")
	}
	return c.advertisedPort
}
func (c *nomadIdentity) BindPort() int {
	if c.bindPort == 0 {
		panic("invalid bind port: 0")
	}
	return c.bindPort
}
func (c *nomadIdentity) BindAddress() string {
	return c.bindAddress
}
