package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/pkg/generate"
)

var GenerateDockerLabelsCmd = &cobra.Command{
	Use:   "docker-labels [source-config-uri]",
	Short: "Generate Docker labels from a Kubeit configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		generateSetOptions.SourceConfigURI = args[0]

		labelArgs, generateErrs, loadFileErrs := generate.DockerLabels(
			&generateSetOptions,
		)

		errorMap := make(map[string][]string) // Map to store errors per file
		if len(loadFileErrs) != 0 {
			for file, errList := range loadFileErrs {
				for _, err := range errList {
					errorMap[file] = append(errorMap[file], fmt.Sprintf("- %v", err))
				}
			}
		}

		for _, err := range generateErrs {
			errorMap["Generate Errors"] = append(
				errorMap["Generate Errors"],
				fmt.Sprintf("- %v", err),
			)
		}

		// If there are errors, format them nicely
		if len(errorMap) > 0 {
			var formattedErrors []string
			for file, errList := range errorMap {
				formattedErrors = append(formattedErrors,
					fmt.Sprintf("%s:\n  %s", file, strings.Join(errList, "\n  ")))
			}
			finalErr := fmt.Errorf("\n%s", strings.Join(formattedErrors, "\n"))
			cmd.SetContext(context.WithValue(cmd.Context(), cmdErrorKey, finalErr))
			logger.Errorf("Error generating docker labels: %v", finalErr)

		}

		fmt.Print(labelArgs)
	},
}
