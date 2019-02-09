package broker

import (
	"github.com/sirupsen/logrus"
	"github.com/vx-labs/mqtt-broker/peers"
	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/subscriptions"
)

func (b *Broker) setupLogs() {
	logger := logrus.New()
	subscriptionLogger := logger.WithField("emitter", "subscription-store")
	b.Subscriptions.On(subscriptions.SubscriptionCreated, func(s *subscriptions.Subscription) {
		subscriptionLogger.WithField("session_id", s.SessionID).
			WithField("peer", s.Peer).
			WithField("mutation", subscriptions.SubscriptionCreated).
			WithField("pattern", string(s.Pattern)).
			Printf("session subscribed")
	})
	b.Subscriptions.On(subscriptions.SubscriptionDeleted, func(s *subscriptions.Subscription) {
		subscriptionLogger.WithField("session_id", s.SessionID).
			WithField("peer", s.Peer).
			WithField("mutation", subscriptions.SubscriptionDeleted).
			WithField("pattern", string(s.Pattern)).
			Printf("session unsubscribed")
	})

	sessionLogger := logger.WithField("emitter", "session-store")
	b.Sessions.On(sessions.SessionCreated, func(s sessions.Session) {
		sessionLogger.WithField("session_id", s.ID).
			WithField("peer", s.Peer).
			WithField("mutation", sessions.SessionCreated).
			WithField("client_id", string(s.ClientID)).
			Printf("session created")
	})
	b.Sessions.On(sessions.SessionDeleted, func(s sessions.Session) {
		sessionLogger.WithField("session_id", s.ID).
			WithField("peer", s.Peer).
			WithField("mutation", sessions.SessionDeleted).
			WithField("client_id", string(s.ClientID)).
			WithField("closure_reason", s.ClosureReason).
			Printf("session closed")
	})
	peerLogger := logger.WithField("emitter", "peer-store")
	b.Peers.On(peers.PeerDeleted, func(s *peers.Peer) {
		peerLogger.WithField("peer_id", s.ID).
			WithField("mesh_id", s.MeshID).
			WithField("mutation", peers.PeerCreated).
			WithField("hostname", s.Hostname).
			Printf("peer lost")
	})

}
