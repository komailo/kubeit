package commands

import (
	"fmt"

	"github.com/komailo/kubeit/common"
	"github.com/spf13/cobra"
)

// Version of KubeIt (replace with actual versioning logic)

// VersionCmd provides the version of the tool
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Kubeit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Kubeit version: %s\n", common.Version)
	},
}

func init() {
	// Ensure this command gets registered in the root command
}
