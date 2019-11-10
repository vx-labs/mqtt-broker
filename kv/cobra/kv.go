package cobra

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/kv/pb"
)

func KV(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "kv",
	}
	c.AddCommand(Get(ctx, config))
	c.AddCommand(GetMetadata(ctx, config))
	c.AddCommand(Set(ctx, config))
	c.AddCommand(Delete(ctx, config))
	return c
}

func Get(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "get",
		PreRun: func(c *cobra.Command, _ []string) {
		},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			for _, key := range argv {
				value, err := client.Get(ctx, []byte(key))
				if err != nil {
					logrus.Errorf("failed to get %q: %v", key, err)
				} else {
					logrus.Infof("%s: %v", key, string(value))
				}
			}
		},
	}
	return c
}
func GetMetadata(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:     "get-metadata",
		Aliases: []string{"md"},
		PreRun: func(c *cobra.Command, _ []string) {
		},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			for _, key := range argv {
				value, err := client.GetMetadata(ctx, []byte(key))
				if err != nil {
					logrus.Errorf("failed to get %q: %v", key, err)
				} else {
					logrus.Infof("%s: version=%v, deadline=%d", key, value.Version, value.Deadline)
				}
			}
		},
	}
	return c
}
func Delete(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "delete",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("version", c.Flags().Lookup("version"))
		},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			for _, key := range argv {
				var err error
				if cmd.Flag("version").Changed {
					err = client.DeleteWithVersion(ctx, []byte(key), config.GetUint64("version"))
				} else {
					err = client.Delete(ctx, []byte(key))
				}
				if err != nil {
					logrus.Errorf("failed to delete %q: %v", key, err)
				} else {
					logrus.Infof("%s: deleted", key)
				}
			}
		},
	}
	c.Flags().Uint64("version", 0, "Key version to delete")
	return c
}
func Set(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "set",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("key", c.Flags().Lookup("key"))
			config.BindPFlag("value", c.Flags().Lookup("value"))
			config.BindPFlag("ttl", c.Flags().Lookup("ttl"))
			config.BindPFlag("version", c.Flags().Lookup("version"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			client := getClient(config)
			key := config.GetString("key")
			ttl := config.GetDuration("ttl")
			version := config.GetUint64("version")
			var err error
			if cmd.Flag("version").Changed {
				err = client.SetWithVersion(ctx, []byte(key), []byte(config.GetString("value")), version, pb.WithTimeToLive(ttl))
			} else {
				err = client.Set(ctx, []byte(key), []byte(config.GetString("value")), pb.WithTimeToLive(ttl))
			}

			if err != nil {
				logrus.Errorf("failed to set %s: %v", key, err)
			} else {
				logrus.Infof("%s: set", key)
			}
		},
	}
	c.Flags().StringP("key", "k", "", "Key's name")
	c.Flags().StringP("value", "v", "", "Key's value")
	c.Flags().Duration("ttl", 0, "Key's time to live")
	c.Flags().Uint64("version", 0, "Key version to update")
	c.MarkFlagRequired("key")
	c.MarkFlagRequired("value")
	return c
}
