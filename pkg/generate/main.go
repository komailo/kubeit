package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/pkg/apis"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func GenerateManifests(generateSetOptions *GenerateOptions, sourceConfigUri string) ([]error, map[string][]error) {

	logger.Infof("Generating manifests from %s", sourceConfigUri)
	var kubeitFileResources, loaderErrs, fileLoadErrs = apis.Loader(sourceConfigUri)

	if loaderErrs != nil {
		return []error{loaderErrs}, fileLoadErrs
	}

	resourceCount := len(kubeitFileResources)
	if resourceCount == 0 {
		return []error{fmt.Errorf("no Kubeit resources found when traversing: %s", sourceConfigUri)}, nil
	} else {
		kindCounts := apis.CountResources(kubeitFileResources)
		for kind, count := range kindCounts {
			logger.Infof("%s: %d", kind, count)
		}

		logger.Infof("Found %d Kubeit resources", resourceCount)
	}

	for _, kubeitFileResource := range kubeitFileResources {
		logger.Debugf("Found resource Kind: %s, API Version: %s in file: %s", kubeitFileResource.APIMetadata.Kind, kubeitFileResource.APIMetadata.APIVersion, kubeitFileResource.FileName)
	}

	generateErrs := generateHelmTemplates(kubeitFileResources, generateSetOptions)
	if generateErrs != nil {
		return generateErrs, nil
	}
	return nil, nil

}

func GenerateCliDocs(rootCmd *cobra.Command, generateSetOptions *GenerateOptions) error {
	docsDir := filepath.Join(generateSetOptions.OutputDir, "cli", common.KubeitCLIName)
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	err := doc.GenMarkdownTree(rootCmd, docsDir)
	if err != nil {
		return fmt.Errorf("failed to generate CLI docs: %w", err)
	}

	fmt.Printf("CLI documentation generated in %s\n", docsDir)
	return nil
}

func GenerateSchemas() {

}
