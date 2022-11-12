package cli

import (
	"github.com/replicate/cog/pkg/api"
	"github.com/replicate/cog/pkg/util/console"
	"github.com/spf13/cobra"
)

func newAPICommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "api",
		Short: "Configure your project for use with Cog",
		RunE: func(cmd *cobra.Command, args []string) error {
			return apiCommand(args)
		},
		Args: cobra.MaximumNArgs(0),
	}

	return cmd
}

func apiCommand(args []string) error {
	console.Infof("\nListening for replicate API...\n")

	return api.Serve()
}
