package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// genDocsCmd represents the command to generate CLI documentation
var generateCliDocsCmd = &cobra.Command{
	Use:   "generate-cli-docs",
	Short: "Generate CLI documentation",
	Long:  `Generates markdown documentation for the CLI commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir := "./docs/kubeit/cli"
		if err := os.MkdirAll(docsDir, 0755); err != nil {
			return fmt.Errorf("failed to create docs directory: %w", err)
		}

		err := doc.GenMarkdownTree(RootCmd, docsDir)
		if err != nil {
			return fmt.Errorf("failed to generate CLI docs: %w", err)
		}

		fmt.Printf("CLI documentation generated in %s\n", docsDir)
		return nil
	},
}
