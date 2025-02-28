package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/pkg/apis"
	helmappv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_application/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func GenerateManifests(generateSetOptions *GenerateOptions, sourceConfigUri string) []error {
	parsedSource, err := parseSourceConfigURI(sourceConfigUri)
	if err != nil {
		return []error{err}
	}

	logger.Infof("Generating manifests from %s", sourceConfigUri)
	var kubeitFileResources []apis.KubeitFileResource

	if parsedSource.Scheme == "file" {
		var errs map[string][]error
		kubeitFileResources, errs = apis.LoadKubeitResourcesFromDir(parsedSource.Path)
		for fileName, err := range errs {
			logger.Errorf("Error loading Kubeit resource from file: %s", fileName)
			for _, e := range err {
				logger.Errorf("    %s", e)
			}
		}
	}

	resourceCount := len(kubeitFileResources)
	if resourceCount == 0 {
		logger.Warn("No Kubeit resources found")
		return nil
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

	errs := generateHelmTemplates(kubeitFileResources, generateSetOptions)
	if errs != nil {
		return errs
	}
	return nil

}

func generateHelmTemplates(kubeitFileResources []apis.KubeitFileResource, generateSetOptions *GenerateOptions) []error {
	var errs []error
	for _, kubeitFileResource := range kubeitFileResources {
		if kubeitFileResource.APIMetadata.Kind != helmappv1alpha1.Kind {
			continue
		}

		if kubeitFileResource.APIMetadata.APIVersion != helmappv1alpha1.GroupVersion {
			continue
		}

		if resource, ok := kubeitFileResource.Resource.(*helmappv1alpha1.HelmApplication); ok {
			err := GenerateManifestFromHelm(*resource, generateSetOptions)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
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
