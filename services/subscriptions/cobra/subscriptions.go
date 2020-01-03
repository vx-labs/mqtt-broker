package cobra

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/format"
)

const subscriptionsTemplate = `  • {{ .ID | shorten| bold | green }}
    {{ "Session" | faint }}: {{ .SessionID | shorten }}
    {{ "Pattern" | faint }}: {{ .Pattern | bufferToString }}
    {{ "QoS" |faint }}: {{ .Qos }}
`

func Subscriptions(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "subscriptions",
	}
	c.AddCommand(ByTopic(ctx, config, adapter))
	c.AddCommand(List(ctx, config, adapter))
	return c
}

func ByTopic(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "by-topic",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(subscriptionsTemplate)
			tenant, _ := cmd.Flags().GetString("tenant")
			for _, pattern := range argv {
				subscriptions, err := client.ByTopic(ctx, tenant, []byte(pattern))
				if err != nil {
					logrus.Errorf("failed to list subscriptions for topic %s: %v", pattern, err)
					return
				}
				for _, subscription := range subscriptions {
					err = tpl.Execute(cmd.OutOrStdout(), subscription)
					if err != nil {
						logrus.Errorf("failed to display subscription %q: %v", subscription.ID, err)
					}
				}
			}
		},
	}
	c.Flags().StringP("tenant", "t", "_default", "search for subscriptions in the given tenant")
	return c
}
func List(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(subscriptionsTemplate)
			subscriptions, err := client.All(ctx)
			if err != nil {
				logrus.Errorf("failed to list subscriptions: %v", err)
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
