package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sessions "github.com/vx-labs/mqtt-broker/sessions/pb"
	"google.golang.org/grpc"
)

func main() {
	rootCmd := &cobra.Command{
		Run: func(cmd *cobra.Command, _ []string) {
			host := viper.GetString("host")
			conn, err := grpc.Dial(host,
				grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(300*time.Millisecond))
			if err != nil {
				panic(fmt.Sprintf("failed to connect %s: %v", host, err))
			}
			ctx := context.Background()
			client := sessions.NewClient(conn)
			for i := 0; i < 5000; i++ {
				id := uuid.New().String()
				err := client.Create(ctx, sessions.SessionCreateInput{
					ID:                id,
					ClientID:          id,
					KeepaliveInterval: 10,
					Peer:              id,
					Tenant:            "_default",
				})
				if err != nil {
					log.Printf("ERR: failed to create session: %v", err)
				}
				_, err = client.ByID(ctx, id)
				if err != nil {
					log.Printf("WARN: failed to read session after creation: %v", err)
				}
			}
		},
	}
	rootCmd.Flags().StringP("host", "", "", "remote GRPC endpoint")
	viper.BindPFlag("host", rootCmd.Flags().Lookup("host"))
	rootCmd.Execute()
}
