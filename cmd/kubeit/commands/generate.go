package commands

import (
	"fmt"
	"os"

	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/pkg/generate"
	"github.com/spf13/cobra"
)

// Options specific to the generate command
var generateSetOptions generate.GenerateOptions

// GenerateCmd is the base sub command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate artifacts",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Create the work directory
		workDir := generateSetOptions.WorkDir
		outputDir := generateSetOptions.OutputDir

		// if the work directory exists, error out
		// TODO: add option to overwrite the work directory
		if _, err := os.Stat(workDir); err == nil {
			logger.Fatalf("Work directory already exists: %s", workDir)
		}

		// Create the work directory
		if err := os.MkdirAll(workDir, os.ModePerm); err != nil {
			logger.Fatalf("Failed to create work directory: %v", err)
		}
		logger.Debugf("Work directory created at: %s", workDir)

		// Create the output directory
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			logger.Fatalf("Failed to create output directory: %v", err)
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Delete the work directory
		workDir := generateSetOptions.WorkDir
		if err := os.RemoveAll(workDir); err != nil {
			return fmt.Errorf("Failed to delete work directory: %v", err)
		} else {
			logger.Debugf("Work directory deleted: %s", workDir)
		}
		return nil
	},
}

func init() {
	// Register subcommands
	GenerateCmd.AddCommand(GenerateManifestCmd)
	GenerateCmd.AddCommand(generateCliDocsCmd)
	GenerateCmd.AddCommand(generateSchemaCmd)

	// Bind the output-dir flag to generateOpts
	GenerateCmd.PersistentFlags().StringVarP(
		&generateSetOptions.OutputDir,
		"output-dir",
		"o",
		"./.kubeit/.generated",
		"Output directory where the generated artifacts will be stored.",
	)

	GenerateCmd.PersistentFlags().StringVarP(
		&generateSetOptions.WorkDir,
		"work-dir",
		"w",
		"./.kubeit/.workdir",
		"Working directory where temporary artifacts and results will be stored.",
	)
}
