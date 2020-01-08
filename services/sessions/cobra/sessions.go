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
    {{ "Keepalive interval" | faint }}: {{ .KeepaliveInterval }}s
    {{ "Remote address" | faint }}: {{ .RemoteAddress }}
`

func Sessions(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "sessions",
	}
	c.AddCommand(List(ctx, config, adapter))
	c.AddCommand(Read(ctx, config, adapter))
	return c
}

func List(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(sessionTemplate)
			sessions, err := client.All(ctx)
			if err != nil {
				logrus.Errorf("failed to list sessions: %v", err)
				return
			}
			for _, session := range sessions {
				err = tpl.Execute(cmd.OutOrStdout(), session)
				if err != nil {
					logrus.Errorf("failed to display session %q: %v", session.ID, err)
				}
			}
		},
	}
	return c
}
func Read(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "read",
		Aliases: []string{"get"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(sessionTemplate)
			for _, arg := range argv {
				session, err := client.ByID(ctx, arg)
				if err != nil {
					logrus.Errorf("failed to list sessions: %v", err)
					return
				}
				err = tpl.Execute(cmd.OutOrStdout(), session)
				if err != nil {
					logrus.Errorf("failed to display session %q: %v", session.ID, err)
				}
			}
		},
	}
	return c
}
