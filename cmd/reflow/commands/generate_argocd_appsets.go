package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scorebet/reflow/internal/logger"
	"github.com/scorebet/reflow/pkg/generate"
)

var GenerateArgoCDAppSetsCmd = &cobra.Command{
	Use:   "argocd-appsets",
	Short: "Generate ArgoCD Application Sets",
	Long:  `Generates the ArgoCD Application Sets from the Reflow resources of kind ArgoAppSet`,
	Run: func(cmd *cobra.Command, args []string) {
		generateSetOptions.SourceConfigURI = args[0]

		generateErrs, loadFileErrs := generate.ArgoAppSets(&generateSetOptions)
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
			logger.Errorf("Error generating manifests: %v", finalErr)

		}
	},
}
