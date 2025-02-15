package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version of KubeIt (replace with actual versioning logic)
var Version = "0.1.0"

// VersionCmd provides the version of the tool
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Kubeit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Kubeit version: %s\n", Version)
	},
}

func init() {
	// Ensure this command gets registered in the root command
}
