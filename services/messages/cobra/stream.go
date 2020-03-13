package cobra

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/format"
)

const streamTemplate = `  • {{ .ID | bold | green }}
    {{ "Shards" | faint }}:
{{ range $id := .ShardIDs }}      • {{ $id | cyan }}
{{end}}`
const streamStatisticsTemplate = `  • {{ .ID | bold | green }}
    {{ "Shards" | faint }}:
{{- range .ShardStatistics }}
    • {{ .ShardID | cyan }}
        {{ "Stored bytes" |faint}}: {{ .StoredBytes | humanBytes }}
        {{ "Stored record count" |faint}}: {{ .StoredRecordCount }}
        {{ "Current offset" |faint}}: {{ .CurrentOffset }}{{end}}
`

func Stream(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "stream",
	}
	c.AddCommand(CreateStream(ctx, config, adapter))
	c.AddCommand(ReadStream(ctx, config, adapter))
	c.AddCommand(ReadStreamStatistics(ctx, config, adapter))
	c.AddCommand(ListStreams(ctx, config, adapter))
	c.AddCommand(PutMessageInStream(ctx, config, adapter))
	c.AddCommand(ConsumeStream(ctx, config, adapter))
	c.AddCommand(Benchmark(ctx, config, adapter))
	return c
}

func ConsumeStream(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
			config.BindPFlag("shard-id", c.Flags().Lookup("shard-id"))
			config.BindPFlag("from-offset", c.Flags().Lookup("from-offset"))
			config.BindPFlag("poll", c.Flags().Lookup("poll"))
			config.BindPFlag("from-now", c.Flags().Lookup("from-now"))
			config.BindPFlag("offset-only", c.Flags().Lookup("offset-only"))
			config.BindPFlag("batch-size", c.Flags().Lookup("batch-size"))
		},
		Use: "consume",
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(adapter)
			offset := config.GetUint64("from-offset")
			size := config.GetInt("batch-size")
			if config.GetBool("from-now") {
				offset = uint64(time.Now().UnixNano())
			}
			for {
				next, messages, err := client.GetMessages(ctx, config.GetString("stream-id"), config.GetString("shard-id"), offset, size)
				if err != nil {
					logrus.Errorf("failed to consume stream: %v", err)
					return
				}
				logrus.Infof("from offset: %d", offset)
				for _, message := range messages {
					if config.GetBool("offset-only") {
						fmt.Printf("%d\n", message.Offset)
					} else {
						fmt.Printf("%d\n\t%s\n", message.Offset, string(message.Payload))
					}
				}
				fmt.Printf("\n")
				if !config.GetBool("poll") {
					logrus.Infof("next offset: %d", next)
					return
				}
				offset = next
			}
		},
	}
	c.Flags().StringP("stream-id", "i", "", "Stream unique ID")
	c.MarkFlagRequired("stream-id")
	c.Flags().StringP("shard-id", "s", "", "Stream shard id")
	c.MarkFlagRequired("shard-id")
	c.Flags().Uint64P("from-offset", "o", 0, "Stream from offset")
	c.Flags().Bool("from-now", false, "Stream from now")
	c.Flags().Bool("offset-only", false, "Only display message offsets")
	c.Flags().BoolP("poll", "", false, "Continuously polls the stream for new messages")
	c.Flags().Int("batch-size", 10, "maximum batch size")
	return c
}
func PutMessageInStream(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "put-message",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
			config.BindPFlag("shard-key", c.Flags().Lookup("shard-key"))
			config.BindPFlag("message", c.Flags().Lookup("message"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(adapter)
			err := client.Put(ctx, config.GetString("stream-id"), config.GetString("shard-key"), []byte(config.GetString("message")))
			if err != nil {
				logrus.Errorf("failed to put message in stream: %v", err)
			}
			logrus.Info("message put")
		},
	}
	c.Flags().StringP("stream-id", "i", "", "Stream unique ID")
	c.MarkFlagRequired("stream-id")
	c.Flags().StringP("shard-key", "s", "", "Message shard key")
	c.MarkFlagRequired("shard-key")
	c.Flags().StringP("message", "m", "", "Message's payload")
	c.MarkFlagRequired("message")
	return c
}
func CreateStream(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "create",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("shard-count", c.Flags().Lookup("shard-count"))
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(adapter)
			err := client.CreateStream(ctx, config.GetString("stream-id"), config.GetInt("shard-count"))
			if err != nil {
				logrus.Errorf("failed to create stream: %v", err)
			}
			logrus.Info("stream created")
		},
	}
	c.Flags().IntP("shard-count", "c", 1, "Number of shard to create in the stream")
	c.Flags().StringP("stream-id", "i", "", "Stream unique ID")
	c.MarkFlagRequired("stream-id")
	return c
}
func ReadStream(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "read",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(streamTemplate)
			for _, id := range argv {
				stream, err := client.GetStream(ctx, id)
				if err != nil {
					logrus.Errorf("failed to read stream %q: %v", id, err)
					continue
				}
				err = tpl.Execute(cmd.OutOrStdout(), stream)
				if err != nil {
					logrus.Errorf("failed to display stream %q: %v", id, err)
					continue
				}
			}
		},
	}
	return c
}
func ReadStreamStatistics(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "statistics",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(streamStatisticsTemplate)
			for _, id := range argv {
				statistics, err := client.GetStreamStatistics(ctx, id)
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
func ListStreams(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)
			tpl := format.ParseTemplate(streamTemplate)
			streams, err := client.ListStreams(ctx)
			if err != nil {
				logrus.Errorf("failed to list streams: %v", err)
				return
			}
			for _, stream := range streams {
				err = tpl.Execute(cmd.OutOrStdout(), stream)
				if err != nil {
					logrus.Errorf("failed to display stream %q: %v", stream.ID, err)
				}
			}
		},
	}
	return c
}
func Benchmark(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "benchmark",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
		},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(adapter)

			count := 0
			total := 1000
			start := time.Now()
			for count < total {
				key := []byte{byte(count)}
				err := client.Put(ctx, config.GetString("stream-id"), "benchmark", []byte("benchmark"))
				if err != nil {
					logrus.Errorf("failed to write record %q: %v", string(key), err)
					return
				}
				count++
				if count%(total/3) == 0 {
					logrus.Infof("written %d key in %s", count, time.Since(start).String())
				}
			}
			timer := time.Since(start)
			logrus.Infof("written %d key in %s", count, timer.String())
		},
	}
	c.Flags().StringP("stream-id", "i", "", "Stream unique ID to write in")
	c.MarkFlagRequired("stream-id")
	return c
}
