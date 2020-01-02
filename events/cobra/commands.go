package cobra

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vx-labs/mqtt-broker/adapters/discovery"
)

func Register(ctx context.Context, cmd *cobra.Command, config *viper.Viper, adapter discovery.DiscoveryAdapter) {
	cmd.AddCommand(Events(ctx, config, adapter))
}
