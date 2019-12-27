package pb

type Assertion func(p *ProtocolContext, c *TransportContext) bool

func AlwaysAllowTransport() Assertion {
	return func(p *ProtocolContext, c *TransportContext) bool {
		return true
	}
}
func TransportMustBeEncrypted() Assertion {
	return func(p *ProtocolContext, c *TransportContext) bool {
		return c.Encrypted
	}
}
