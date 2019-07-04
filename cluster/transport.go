package cluster

type meshTransport struct {
	id    string
	peers PeerStore
	mesh  Mesh
}
