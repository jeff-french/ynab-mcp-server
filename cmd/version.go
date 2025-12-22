package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set via ldflags during build
var Version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ynab-mcp-server version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
