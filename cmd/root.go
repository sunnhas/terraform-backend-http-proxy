package cmd

import (
	"log"
	"os"
	"terraform-backend-http-proxy/pid"
	"terraform-backend-http-proxy/server"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terraform-backend-http-proxy",
	Short: "A http backend integrating to a Git repository",
	Long: `A http backend which integrates to any Git repository
and stores any terraform state within that repository.

Send the backend to the background with 'terraform-backend-http-proxy &'.
This will make sure to start the service and still output content in the
terminal that you are working within.`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := pid.CreateFile(); err != nil {
			log.Fatal(err)
		}

		server.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terraform-backend-http-proxy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
