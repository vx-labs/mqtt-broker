package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/vx-labs/mqtt-broker/queues/pb"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func benchmark() *cobra.Command {
	c := &cobra.Command{
		Use: "benchmark",
	}
	c.AddCommand(benchmarkProduce())
	c.AddCommand(benchmarkConsume())
	c.PersistentFlags().StringP("queue", "q", "test", "remote GRPC endpoint")
	viper.BindPFlag("queue", c.PersistentFlags().Lookup("queue"))
	return c
}
func benchmarkProduce() *cobra.Command {
	c := &cobra.Command{
		Use: "produce",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(cmd)
			queueID := viper.GetString("queue")
			sigc := make(chan os.Signal)
			max := viper.GetInt("count")
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				signal.Notify(sigc,
					syscall.SIGINT,
					syscall.SIGTERM,
					syscall.SIGQUIT)
				<-sigc
				cancel()
			}()
			started := time.Now()
			count := 0
			for max == 0 || max > count {
				payload := packet.Publish{Header: &packet.Header{}, Topic: []byte("test"), Payload: []byte(time.Now().String())}
				err := client.PutMessage(ctx, queueID, &payload)
				if err != nil {
					log.Printf("received %v", err)
					break
				}
				if count%10 == 0 {
					log.Printf("published %d messages in %s", count, time.Since(started).String())
				}
				count++
			}
			log.Printf("SUMMARY: published %d messages in %s", count, time.Since(started).String())
		},
	}
	c.Flags().IntP("count", "c", 0, "Number of messages to produce. Set to 0 to never stop.")
	viper.BindPFlag("count", c.Flags().Lookup("count"))
	return c
}
func benchmarkConsume() *cobra.Command {
	return &cobra.Command{
		Use: "consume",
		Run: func(cmd *cobra.Command, argv []string) {
			client := getClient(cmd)
			queueID := viper.GetString("queue")
			sigc := make(chan os.Signal)
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				signal.Notify(sigc,
					syscall.SIGINT,
					syscall.SIGTERM,
					syscall.SIGQUIT)
				<-sigc
				cancel()
			}()
			err := client.StreamMessages(ctx, queueID, func(ack uint64, p *packet.Publish) error {
				log.Printf("received %q", p.Payload)
				return nil
			})
			log.Printf("received %v", err)
		},
	}
}

func getClient(cmd *cobra.Command) *pb.Client {
	host := viper.GetString("host")
	conn, err := grpc.Dial(host,
		grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(300*time.Millisecond))
	if err != nil {
		log.Fatalf("failed to connect %s: %v", host, err)
	}
	return pb.NewClient(conn)
}

func main() {
	rootCmd := &cobra.Command{}
	rootCmd.PersistentFlags().StringP("host", "", "", "remote GRPC endpoint")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	rootCmd.AddCommand(benchmark())
	rootCmd.Execute()
}
