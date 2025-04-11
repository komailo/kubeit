package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scorebet/reflow/common"
	"github.com/scorebet/reflow/internal/logger"
	"github.com/scorebet/reflow/pkg/generate"
)

var GenerateManifestCmd = &cobra.Command{
	Use:   "manifest [source-config-uri]",
	Short: fmt.Sprintf("Generate Kubernetes manifests from a %s configuration", common.AppName),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		generateSetOptions.SourceConfigURI = args[0]

		generateErrs, loadFileErrs := generate.Manifests(&generateSetOptions)

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

func init() {
	GenerateManifestCmd.PersistentFlags().StringArrayVarP(
		&generateSetOptions.NamedValues,
		"named-values",
		"e",
		nil,
		"Name of the NamedValues to use while generating manifests, multiple names can be provided by using args multiple times and they will be used in the order they are provided",
	)
}
