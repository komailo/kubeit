package commands

import (
	"fmt"

	"github.com/komailo/kubeit/pkg/generate"
	"github.com/spf13/cobra"
)

var GenerateManifestCmd = &cobra.Command{
	Use:   "manifest [source-config-uri]",
	Short: "Generate Kubernetes manifests from a Kubeit configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceConfigUri := args[0]

		errs := generate.GenerateManifests(&generateSetOptions, sourceConfigUri)
		if len(errs) > 0 {
			return fmt.Errorf("Failed to generate manifests: %v", errs)
		}
		return nil
	},
}
