package main

import (
	"fmt"

	"github.com/spf13/viper"

	_ "net/http/pprof"
	"os"

	"github.com/vx-labs/mqtt-broker/cli"

	"github.com/spf13/cobra"
)

func main() {
	config := viper.New()
	root := &cobra.Command{
		Use: "broker",
	}
	dev := &cobra.Command{
		Use: "dev",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("certificate-file", c.Flags().Lookup("certificate-file"))
			config.BindPFlag("private-key-file", c.Flags().Lookup("private-key-file"))
		},
		Short: "Start all services",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cli.Bootstrap(cmd, config)
			for _, service := range Services() {
				if config.GetBool(fmt.Sprintf("start-%s", service.Name())) {
					ctx.AddService(cmd, config, service.Name(), service.Run)
				}
			}
			ctx.Run(config)
		},
	}
	dev.Flags().StringP("certificate-file", "c", "./run_config/cert.pem", "Write certificate to this file")
	dev.Flags().StringP("private-key-file", "k", "./run_config/privkey.pem", "Write private key to this file")
	for _, service := range Services() {
		cli.AddServiceFlags(dev, config, service.Name())
		err := service.Register(dev, config)
		if err != nil {
			panic(err)
		}
	}

	cli.AddClusterFlags(dev, config)

	services := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services"},
	}
	serviceList := Services()
	for idx := range serviceList {
		service := serviceList[idx]
		serviceConfig := viper.New()
		serviceCommand := &cobra.Command{
			Use: service.Name(),
			PreRun: func(cmd *cobra.Command, _ []string) {
				if os.Getenv("TLS_PRIVATE_KEY") == "" {
					os.Setenv("TLS_PRIVATE_KEY", serviceConfig.GetString("grpc-tls-private-key"))
				}
				if os.Getenv("TLS_CERTIFICATE") == "" {
					os.Setenv("TLS_CERTIFICATE", serviceConfig.GetString("grpc-tls-certificate"))
				}
				if os.Getenv("TLS_CA_CERTIFICATE") == "" {
					os.Setenv("TLS_CA_CERTIFICATE", serviceConfig.GetString("grpc-tls-certificate-authority"))
				}
			},
			Run: func(cmd *cobra.Command, _ []string) {
				ctx := cli.Bootstrap(cmd, serviceConfig)
				ctx.AddService(cmd, serviceConfig, service.Name(), service.Run)
				ctx.Run(serviceConfig)
			},
		}
		err := service.Register(serviceCommand, serviceConfig)
		if err != nil {
			panic(err)
		}
		cli.AddClusterFlags(serviceCommand, serviceConfig)
		cli.AddServiceFlags(serviceCommand, serviceConfig, service.Name())
		services.AddCommand(serviceCommand)
	}

	root.AddCommand(dev)
	root.AddCommand(services)
	root.Execute()
}
