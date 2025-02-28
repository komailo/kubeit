package commands

import (
	"fmt"

	"github.com/komailo/kubeit/pkg/apis"

	"github.com/spf13/cobra"
)

// generateSchemaCmd is a CLI command for generating JSON schemas
var generateSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generates JSON schemas for all API versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir := "schemas"
		fmt.Println("Generating schemas...")

		if err := apis.GenerateSchemas(outputDir); err != nil {
			return fmt.Errorf("failed to generate schemas: %w", err)
		}

		fmt.Println("Schemas successfully generated in", outputDir)
		return nil
	},
}
