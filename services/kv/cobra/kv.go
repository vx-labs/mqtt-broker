package cobra

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/services/kv/pb"
)

func KV(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "kv",
	}
	c.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		config.BindPFlag("discovery-url", c.PersistentFlags().Lookup("discovery-url"))
	}
	c.PersistentFlags().StringP("discovery-url", "d", "http://localhost:8081", "discovery api URL")
	c.AddCommand(Get(ctx, config))
	c.AddCommand(List(ctx, config))
	c.AddCommand(GetMetadata(ctx, config))
	c.AddCommand(GetWithMetadata(ctx, config))
	c.AddCommand(Set(ctx, config))
	c.AddCommand(Delete(ctx, config))
	c.AddCommand(Benchmark(ctx, config))
	c.AddCommand(Pressure(ctx, config))
	return c
}

func Get(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "get",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			for _, key := range argv {
				value, err := client.Get(ctx, []byte(key))
				if err != nil {
					logrus.Errorf("failed to get %q: %v", key, err)
				} else {
					fmt.Println(string(value))
				}
			}
		},
	}
	return c
}
func Benchmark(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "benchmark",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)

			count := 0
			total := 1000
			start := time.Now()
			for count < total {
				key := []byte{byte(count)}
				err := client.Set(ctx, key, []byte("benchmark"))
				if err != nil {
					logrus.Errorf("failed to write key %q: %v", string(key), err)
					return
				}
				count++
				if count%(total/3) == 0 {
					logrus.Infof("written %d key in %s", count, time.Since(start).String())
				}
			}
			timer := time.Since(start)
			logrus.Infof("written %d key in %s", count, timer.String())

			count = 0
			start = time.Now()
			for count < total {
				key := []byte{byte(count)}
				_, err := client.Get(ctx, key)
				if err != nil {
					logrus.Errorf("failed to read key %q: %v", string(key), err)
					break
				}
				count++
				if count%(total/3) == 0 {
					logrus.Infof("read %d key in %s", count, time.Since(start).String())
				}
			}
			timer = time.Since(start)
			logrus.Infof("read %d key in %s", count, timer.String())
		},
	}
	return c
}
func Pressure(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:   "pressure",
		Short: "read and write indefinitely on the kv store",
		Run: func(cmd *cobra.Command, argv []string) {
			retries := 5
			for retries > 0 {
				retries--
				client := getClient(config)
			mainLoop:
				for {
					count := 0
					total := 5
					for count < total {
						key := []byte{byte(count)}
						err := client.Set(ctx, key, []byte("benchmark"))
						if err != nil {
							logrus.Errorf("failed to write key %q: %v", string(key), err)
							<-time.After(5 * time.Second)
							client = getClient(config)
							break mainLoop
						}
						count++
					}

					count = 0
					for count < total {
						key := []byte{byte(count)}
						_, err := client.Get(ctx, key)
						if err != nil {
							logrus.Errorf("failed to read key %q: %v", string(key), err)
							<-time.After(5 * time.Second)
							client = getClient(config)
							break mainLoop
						}
						count++
					}
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
func GetWithMetadata(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:     "get-with-metadata",
		Aliases: []string{"gmd"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			for _, key := range argv {
				value, md, err := client.GetWithMetadata(ctx, []byte(key))
				if err != nil {
					logrus.Errorf("failed to get %q: %v", key, err)
				} else {
					fmt.Println(string(value))
					logrus.Infof("%s: version=%v, deadline=%d", key, md.Version, md.Deadline)
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
func List(ctx context.Context, config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(config)
			mds, err := client.List(ctx)
			if err != nil {
				logrus.Errorf("failed to list keys: %v", err)
				return
			}
			for _, md := range mds {
				logrus.Infof("%s: version=%v, deadline=%d", md.Key, md.Version, md.Deadline)
			}
		},
	}
	return c
}
