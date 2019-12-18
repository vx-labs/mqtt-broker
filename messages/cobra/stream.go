package cobra

import (
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const streamTemplate = `  • {{ .ID | bold | green }}
    {{ "Shards" | faint }}:
{{ range $id := .ShardIDs }}      • {{ $id }}
{{end}}`

func Stream(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "stream",
	}
	c.AddCommand(CreateStream(ctx, config))
	c.AddCommand(ReadStream(ctx, config))
	c.AddCommand(ListStreams(ctx, config))
	c.AddCommand(PutMessageInStream(ctx, config))
	c.AddCommand(ConsumeStream(ctx, config))
	return c
}

func ConsumeStream(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
			config.BindPFlag("shard-id", c.Flags().Lookup("shard-id"))
			config.BindPFlag("from-offset", c.Flags().Lookup("from-offset"))
			config.BindPFlag("poll", c.Flags().Lookup("poll"))
			config.BindPFlag("from-now", c.Flags().Lookup("from-now"))
			config.BindPFlag("batch-size", c.Flags().Lookup("batch-size"))
		},
		Use: "consume",
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(config)
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
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
				for _, message := range messages {
					fmt.Printf("%d - %s\n", message.Offset, string(message.Payload))
				}
				if !config.GetBool("poll") {
					logrus.Infof("next offset: %d", next)
					return
				}
				<-ticker.C
			}
		},
	}
	c.Flags().StringP("stream-id", "i", "", "Stream unique ID")
	c.MarkFlagRequired("stream-id")
	c.Flags().StringP("shard-id", "s", "", "Stream shard id")
	c.MarkFlagRequired("shard-id")
	c.Flags().Uint64P("from-offset", "o", 0, "Stream from offset")
	c.Flags().Bool("from-now", false, "Stream from now")
	c.Flags().BoolP("poll", "", false, "Continuously polls the stream for new messages")
	c.Flags().Int("batch-size", 10, "maximum batch size")
	return c
}
func PutMessageInStream(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "put-message",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
			config.BindPFlag("shard-key", c.Flags().Lookup("shard-key"))
			config.BindPFlag("message", c.Flags().Lookup("message"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(config)
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
func CreateStream(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "create",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("shard-count", c.Flags().Lookup("shard-count"))
			config.BindPFlag("stream-id", c.Flags().Lookup("stream-id"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(config)
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
func ReadStream(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "read",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(streamTemplate)
			if err != nil {
				panic(err)
			}
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
func ListStreams(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(streamTemplate)
			if err != nil {
				panic(err)
			}
			for _, id := range argv {
				streams, err := client.ListStreams(ctx)
				if err != nil {
					logrus.Errorf("failed to list streams: %v", err)
					continue
				}
				for _, stream := range streams {
					err = tpl.Execute(cmd.OutOrStdout(), stream)
					if err != nil {
						logrus.Errorf("failed to display stream %q: %v", id, err)
						continue
					}
				}
			}
		},
	}
	return c
}
