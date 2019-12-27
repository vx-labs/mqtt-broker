package cobra

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Register(ctx context.Context, cmd *cobra.Command, config *viper.Viper) {
	cmd.AddCommand(Auth(ctx, config))
}
