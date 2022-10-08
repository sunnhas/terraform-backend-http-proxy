package cmd

import (
	"log"
	"terraform-backend-http-proxy/pid"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops any running terraform-backend-http-proxy",
	Long: `Stops any running terraform-backend-http-proxy
by looking at the current pid file.`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := pid.RemoveFile(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
