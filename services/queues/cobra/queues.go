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

const queueTemplate = `  • {{ . | bold | green }}
`
const streamStatisticsTemplate = `  • {{ .ID | bold | green }}
    {{ "Message count" |faint}}: {{ .MessageCount }}
    {{ "Inflight count" |faint}}: {{ .InflightCount }}
`

func Queues(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "queues",
		Aliases: []string{"queue"},
	}
	c.AddCommand(ListQueues(ctx, config, adapter))
	c.AddCommand(ReadQueueStatistics(ctx, config, adapter))
	c.AddCommand(DeleteQueue(ctx, config, adapter))
	return c
}

func ListQueues(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(queueTemplate)
			if err != nil {
				panic(err)
			}
			streams, err := client.ListQueues(ctx)
			if err != nil {
				logrus.Errorf("failed to list streams: %v", err)
				return
			}
			for _, stream := range streams {
				err = tpl.Execute(cmd.OutOrStdout(), stream)
				if err != nil {
					logrus.Errorf("failed to display stream %q: %v", stream, err)
				}
			}
		},
	}
	return c
}
func ReadQueueStatistics(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "statistics",
		Aliases: []string{"stats"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(streamStatisticsTemplate)
			if err != nil {
				panic(err)
			}
			if ok, _ := cmd.Flags().GetBool("all"); ok {
				queues, err := client.ListQueues(ctx)
				if err != nil {
					logrus.Errorf("failed to list queues: %v", err)
					return
				}
				argv = queues
			}
			for _, id := range argv {
				statistics, err := client.GetQueueStatistics(ctx, id)
				if err != nil {
					logrus.Errorf("failed to read queue statistics %q: %v", id, err)
					continue
				}
				err = tpl.Execute(cmd.OutOrStdout(), statistics)
				if err != nil {
					logrus.Errorf("failed to display stream %q: %v", id, err)
					continue
				}
			}
		},
	}
	c.Flags().BoolP("all", "a", false, "list statistics for all queues")
	return c
}
func DeleteQueue(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			for _, id := range argv {
				err := client.Delete(ctx, id)
				if err != nil {
					logrus.Errorf("failed to delete queue %q: %v", id, err)
					continue
				}
			}
		},
	}
	return c
}
