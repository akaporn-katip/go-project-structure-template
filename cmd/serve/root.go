package serve

import (
	"github.com/akaporn-katip/go-project-structure-template/daemon"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	var serveCmd = &cobra.Command{
		Use: "serve",
		Run: func(cmd *cobra.Command, args []string) {
			configVal, _ := cmd.Flags().GetString("config")
			daemon.Serve(configVal)
		},
	}

	return serveCmd
}
