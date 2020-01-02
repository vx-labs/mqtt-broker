package cobra

import (
	"context"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
)

const sessionTemplate = `  • {{ .ID | bold | green }}
    {{ "Tenant" | faint }}: {{ .Tenant }}
    {{ "Created" | faint }}: {{ .Created }}
    {{ "Client ID" | faint }}: {{ .ClientID }}
    {{ "Transport" | faint }}: {{ .Transport }}
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
			tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(sessionTemplate)
			if err != nil {
				panic(err)
			}
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
