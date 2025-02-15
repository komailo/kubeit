package commands

import (
	"fmt"

	"github.com/komailo/kubeit/cmd/kubeit/commands/internal"
	"github.com/spf13/cobra"
)

// Options specific to the generate command
var generateOpts internal.GenerateOptions

var GenerateCmd = &cobra.Command{
	Use:   "generate [config]",
	Short: "Generate Kubernetes manifests from a Kubeit configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := args[0]

		fmt.Printf("Generating manifests from %s into %s\n", configFile, generateOpts.OutputDir)
		return nil
	},
}

func init() {
	// Bind the output-dir flag to generateOpts
	GenerateCmd.Flags().StringVarP(
		&generateOpts.OutputDir,
		"output-dir",
		"o",
		"./build",
		"Output directory where manifests are stored",
	)
}
