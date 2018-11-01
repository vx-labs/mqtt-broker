package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/spf13/viper"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/vx-labs/mqtt-broker/broker/rpc/client"
	"github.com/vx-labs/mqtt-broker/sessions"

	"google.golang.org/grpc"
)

type APIWrapper struct {
	api *client.Client
}

func (a *APIWrapper) API() *client.Client {
	return a.api
}
func logInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// Logic before invoking the invoker
	start := time.Now()
	// Calls the invoker to execute RPC
	err := invoker(ctx, method, req, reply, cc, opts...)
	// Logic after invoking the invoker
	log.Printf("DEBUG: Invoked RPC method=%s; Duration=%s; Error=%v", method,
		time.Since(start), err)
	return err
}
func main() {
	helper := &APIWrapper{}
	var conn *grpc.ClientConn
	var err error
	ctx := context.Background()
	root := &cobra.Command{
		Use: "mqttctl",
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if conn != nil {
				conn.Close()
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			endpoint := viper.GetString("endpoint")
			conn, err = grpc.Dial(endpoint,
				grpc.WithInsecure())
			if err != nil {
				log.Fatalf("FATAL: failed to dial %s: %v", endpoint, err)
			}
			helper.api = client.New(conn)
		},
	}
	root.PersistentFlags().StringP("endpoint", "e", "localhost:9090", "Broker GRPC endpoint")
	viper.BindPFlag("endpoint", root.PersistentFlags().Lookup("endpoint"))
	root.AddCommand(Sessions(ctx, helper))
	root.Execute()
}

var SessionTemplate = `• {{ .ID | green | bold }}
  {{ "Tenant:"     | faint }} {{ .Tenant }}
  {{ "Peer:"     | faint }} {{ .Peer }}`

func Sessions(ctx context.Context, helper *APIWrapper) *cobra.Command {
	c := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"}}
	c.AddCommand(SessionsList(ctx, helper))
	return c
}

func SessionsList(ctx context.Context, helper *APIWrapper) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			set, err := helper.API().ListSessions(ctx)
			if err != nil {
				log.Printf("ERR: failed to list sessions: %v", err)
				return
			}
			tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(fmt.Sprintf("%s\n", SessionTemplate))
			if err != nil {
				log.Printf("ERR: failed to parse session template: %v", err)
				return
			}
			set.Apply(func(s *sessions.Session) {
				tpl.Execute(os.Stdout, s)
			})
		},
	}
	return c
}
