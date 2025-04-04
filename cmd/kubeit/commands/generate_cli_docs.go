package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/komailo/kubeit/pkg/generate"
)

// genDocsCmd represents the command to generate CLI documentation
var generateCliDocsCmd = &cobra.Command{
	Use:   "cli-docs",
	Short: "Generate CLI documentation",
	Long:  `Generates markdown documentation for the CLI commands.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		err := generate.CliDocs(RootCmd, &generateSetOptions)
		if err != nil {
			return fmt.Errorf("failed to generate CLI docs: %w", err)
		}

		return nil
	},
}
