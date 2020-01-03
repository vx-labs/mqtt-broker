package cobra

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/format"
)

const sessionTemplate = `  • {{ .ID | shorten | bold | green }}
    {{ "Tenant" | faint }}: {{ .Tenant }}
    {{ "Created" | faint }}: {{ .Created }}
    {{ "Client ID" | faint }}: {{ .ClientID }}
    {{ "Transport" | faint }}: {{ .Transport }}
    {{ "Connected" | faint }}: {{ .Created | timeToDuration }}
    {{ "Last keepalive" | faint }}: {{ .LastKeepAlive | timeToDuration }}
    {{ "Remote address" | faint }}: {{ .RemoteAddress }}
`

func Sessions(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "sessions",
	}
	c.AddCommand(List(ctx, config, adapter))
	return c
}

func List(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(sessionTemplate)
			subscriptions, err := client.All(ctx)
			if err != nil {
				logrus.Errorf("failed to list sessions: %v", err)
				return
			}
			for _, subscription := range subscriptions {
				err = tpl.Execute(cmd.OutOrStdout(), subscription)
				if err != nil {
					logrus.Errorf("failed to display subscription %q: %v", subscription.ID, err)
				}
			}
		},
	}
	return c
}
