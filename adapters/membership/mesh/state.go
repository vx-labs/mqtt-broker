package mesh

func (l *GossipMembershipAdapter) NotifyMsg(b []byte) {
}
func (l *GossipMembershipAdapter) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}
func (l *GossipMembershipAdapter) LocalState(join bool) []byte {
	return nil
}

func (m *GossipMembershipAdapter) MergeRemoteState(buf []byte, join bool) {}
