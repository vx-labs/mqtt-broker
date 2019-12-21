package cobra

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/events"
)

const streamTemplate = `  • {{ .ID | bold | green }}
    {{ "Shards" | faint }}:
{{ range $id := .ShardIDs }}      • {{ $id }}
{{end}}`
const streamStatisticsTemplate = `  • {{ .ID | bold | green }}
    {{ "Shards" | faint }}:
{{- range .ShardStatistics }}
    • {{ .ShardID }}
        {{ "Stored bytes" |faint}}: {{ .StoredBytes }}
        {{ "Stored record count" |faint}}: {{ .StoredRecordCount }}
        {{ "Current offset" |faint}}: {{ .CurrentOffset }}{{end}}
`

func Events(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "events",
	}
	c.AddCommand(Stream(ctx, config))
	return c
}

func Stream(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
			config.BindPFlag("shard-id", c.Flags().Lookup("shard-id"))
			config.BindPFlag("from-offset", c.Flags().Lookup("from-offset"))
			config.BindPFlag("poll", c.Flags().Lookup("poll"))
			config.BindPFlag("from-now", c.Flags().Lookup("from-now"))
			config.BindPFlag("batch-size", c.Flags().Lookup("batch-size"))
		},
		Use: "stream",
		Run: func(cmd *cobra.Command, _ []string) {
			client := getStreamClient(config)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			size := config.GetInt("batch-size")
			streamConfig, err := client.GetStream(ctx, config.GetString("stream-id"))
			if err != nil {
				logrus.Error(err)
				return
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			wg := sync.WaitGroup{}
			for _, shard := range streamConfig.ShardIDs {
				wg.Add(1)
				go func(shard string) {
					offset := uint64(0)
					if config.GetBool("from-now") {
						offset = uint64(time.Now().UnixNano())
					}
					for {
						next, messages, err := client.GetMessages(ctx, config.GetString("stream-id"), shard, offset, size)
						if err != nil {
							logrus.Errorf("failed to consume stream: %v", err)
							return
						}
						for _, message := range messages {
							events, err := events.Decode(message.Payload)
							if err != nil {
								logrus.Errorf("invalid event payload: %v", err)
								continue
							}
							for _, event := range events {
								encoder.Encode(event)
							}
						}
						offset = next
						<-ticker.C
					}
				}(shard)
			}
			wg.Wait()
		},
	}
	c.Flags().StringP("stream-id", "i", "events", "Stream unique ID")
	c.Flags().Bool("from-now", false, "Stream from now")
	c.Flags().Int("batch-size", 10, "maximum batch size")
	return c
}
