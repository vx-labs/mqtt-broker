package cluster

type mockedChannel struct{}

func (m *mockedChannel) Broadcast([]byte) {}

type mockedMesh struct{}

func (m *mockedMesh) AddState(key string, state State) (Channel, error) {
	return &mockedChannel{}, nil
}
func (m *mockedMesh) Join(hosts []string) error {
	return nil
}
func (m *mockedMesh) MemberRPCAddress(string) (string, error) {
	return "", ErrNodeNotFound
}
func (m *mockedMesh) ID() string {
	return "id"
}

func MockedMesh() Mesh {
	return &mockedMesh{}
}
