package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

// Version holds the version of the built binary - it's injected from build process via -ldflags="-X 'cmd.Version={version}'"
var Version = "dev"

// Commit holds the commit sha of the built binary - it's injected from build process via -ldflags="-X 'cmd.Commit={commit-sha}'"
var Commit = "none"

// Date holds the date for when the binary was built - it's injected from build process via -ldflags="-X 'cmd.Date={date}'"
var Date = time.Now().Format(time.UnixDate)

// versionCmd will print the version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`Version: %s
Commit: %s
Date: %s`, Version, Commit, Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
