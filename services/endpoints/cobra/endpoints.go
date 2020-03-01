package cobra

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
	"github.com/vx-labs/mqtt-broker/format"
)

func Endpoints(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "endpoints",
	}
	c.AddCommand(getEndpointsByService(ctx, config, adapter))
	return c
}

const endpointTemplate = `  • {{ .ID  | bold | green }}
    {{ "Name" | faint }}: {{ .Name }}
    {{ "Tag" | faint }}: {{ .Tag }}
    {{ "NetworkAddress" | faint }}: {{ .NetworkAddress }}
    {{ "Peer" | faint }}: {{ .Peer }}
    {{ "Health" | faint }}: {{ .Health }}
`

func getEndpointsByService(ctx context.Context, config *viper.Viper, adapter discovery.DiscoveryAdapter) *cobra.Command {
	c := &cobra.Command{
		Use: "by-service",
		PreRun: func(c *cobra.Command, _ []string) {
			config.BindPFlag("service", c.Flags().Lookup("service"))
			config.BindPFlag("tag", c.Flags().Lookup("tag"))
		},
		Run: func(cmd *cobra.Command, _ []string) {
			tpl := format.ParseTemplate(endpointTemplate)
			resp, err := adapter.EndpointsByService(config.GetString("service"), config.GetString("tag"))

			if err != nil {
				logrus.Errorf("endpoints resolution failed: %v", err)
			} else {
				for _, elt := range resp {
					tpl.Execute(cmd.OutOrStdout(), elt)
				}
			}
		},
	}
	c.Flags().StringP("service", "s", "", "Service name")
	c.Flags().StringP("tag", "t", "", "Tag")
	c.MarkFlagRequired("service")
	c.MarkFlagRequired("tag")
	return c
}
