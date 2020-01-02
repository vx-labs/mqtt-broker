package cobra

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/services/auth/pb"
)

func Auth(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "auth",
	}
	c.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		config.BindPFlag("discovery-url", c.PersistentFlags().Lookup("discovery-url"))
	}
	c.PersistentFlags().StringP("discovery-url", "d", "http://localhost:8081", "discovery api URL")
	c.AddCommand(createToken(ctx, config, adapter))
	return c
}

func createToken(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "create-token",
		Aliases: []string{"ct"},
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("username", c.Flags().Lookup("username"))
			config.BindPFlag("password", c.Flags().Lookup("password"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(adapter)

			resp, err := client.CreateToken(ctx, pb.ProtocolContext{
				Username: config.GetString("username"),
				Password: config.GetString("password"),
			}, pb.TransportContext{})

			if err != nil {
				logrus.Errorf("authentication failed: %v", err)
			} else {
				logrus.Infof("authentication succeeded")
				fmt.Println(resp.JWT)
			}
		},
	}
	c.Flags().StringP("username", "u", "", "Username")
	c.Flags().StringP("password", "p", "", "Password")
	c.MarkFlagRequired("username")
	c.MarkFlagRequired("password")
	return c
}
