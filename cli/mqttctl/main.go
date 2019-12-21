package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	eventsCommand "github.com/vx-labs/mqtt-broker/events/cobra"
	kvCommand "github.com/vx-labs/mqtt-broker/kv/cobra"
	messagesCommand "github.com/vx-labs/mqtt-broker/messages/cobra"
)

func main() {
	rootCmd := &cobra.Command{}
	config := viper.New()
	rootCmd.PersistentFlags().StringP("host", "", "", "remote GRPC endpoint")
	config.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	ctx := context.Background()
	messagesCommand.Register(ctx, rootCmd, config)
	kvCommand.Register(ctx, rootCmd, config)
	eventsCommand.Register(ctx, rootCmd, config)
	rootCmd.Execute()
}
