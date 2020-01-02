package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/cli"
	"github.com/vx-labs/mqtt-broker/services/api"
	"github.com/vx-labs/mqtt-broker/services/auth"
	"github.com/vx-labs/mqtt-broker/services/broker"
	"github.com/vx-labs/mqtt-broker/services/endpoints"
	"github.com/vx-labs/mqtt-broker/services/kv"
	"github.com/vx-labs/mqtt-broker/services/listener"
	"github.com/vx-labs/mqtt-broker/services/messages"
	"github.com/vx-labs/mqtt-broker/services/queues"
	"github.com/vx-labs/mqtt-broker/services/router"
	"github.com/vx-labs/mqtt-broker/services/sessions"
	"github.com/vx-labs/mqtt-broker/services/subscriptions"
	"github.com/vx-labs/mqtt-broker/services/topics"
	"go.uber.org/zap"
)

type service interface {
	Name() string
	Register(cmd *cobra.Command, config *viper.Viper) error
	Run(id string, config *viper.Viper, logger *zap.Logger, mesh discovery.DiscoveryAdapter) cli.Service
}

func Services() []service {
	return []service{
		&sessions.Service{},
		&subscriptions.Service{},
		&kv.Service{},
		&messages.Service{},
		&topics.Service{},
		&router.Service{},
		&listener.Service{},
		&queues.Service{},
		&api.Service{},
		&broker.Service{},
		&auth.Service{},
		&endpoints.Service{},
	}
}
