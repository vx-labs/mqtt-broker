package cobra

import (
	"context"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const queueTemplate = `  • {{ . | bold | green }}
`
const streamStatisticsTemplate = `  • {{ .ID | bold | green }}
    {{ "Message count" |faint}}: {{ .MessageCount }}
    {{ "Inflight count" |faint}}: {{ .InflightCount }}
`

func Queues(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:     "queues",
		Aliases: []string{"queue"},
	}
	c.AddCommand(ListQueues(ctx, config))
	c.AddCommand(ReadQueueStatistics(ctx, config))
	return c
}

func ListQueues(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
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
func ReadQueueStatistics(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:     "statistics",
		Aliases: []string{"stats"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(streamStatisticsTemplate)
			if err != nil {
				panic(err)
			}
			for _, id := range argv {
				statistics, err := client.GetQueueStatistics(ctx, id)
				if err != nil {
					logrus.Errorf("failed to read stream %q: %v", id, err)
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
	return c
}
