package cli

import (
	"fmt"

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

	cmd.Flags().IntVarP(&listenPort, "port", "p", 5555, "port to serve api on the host")
	cmd.Flags().StringVarP(&listenHost, "listen", "l", "127.0.0.1", "ip to listen for api")

	return cmd
}

var listenPort int
var listenHost string

func apiCommand(args []string) error {

	listen := fmt.Sprintf("%s:%d", listenHost, listenPort)

	console.Infof("Starting replicate API...")
	console.Infof(" -- listening on http://%s ", listen)

	return api.Serve(listen)
}
