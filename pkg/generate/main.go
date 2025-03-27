package generate

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"sigs.k8s.io/yaml"

	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/internal/version"
	"github.com/komailo/kubeit/pkg/api/loader"
	"github.com/komailo/kubeit/pkg/apis"
)

func Manifests(generateSetOptions *Options) ([]error, map[string][]error) {
	sourceConfigURI := generateSetOptions.SourceConfigURI
	logger.Infof("Generating manifests from %s", sourceConfigURI)

	loaderInt := loader.NewLoader()
	loaderErr := loaderInt.FromSourceURI(sourceConfigURI)

	if loaderErr != nil {
		return nil, loaderErr
	}

	if loaderInt.ResourceCount == 0 {
		return []error{
			fmt.Errorf("no Kubeit resources found when traversing: %s", sourceConfigURI),
		}, nil
	}

	loaderInt.LogResources()

	generateErrs := ManifestsFromHelm(loaderInt, generateSetOptions)
	if generateErrs != nil {
		return generateErrs, nil
	}

	return nil, nil
}

func DockerLabels(
	_ *Options,
	sourceConfigURI string,
) ([]error, map[string][]error) {
	logger.Infof("Generating Docker Labels from %s", sourceConfigURI)

	kubeitFileResources, _, fileLoadErrs, loaderErrs := apis.Loader(sourceConfigURI)

	if loaderErrs != nil {
		return []error{loaderErrs}, fileLoadErrs
	}

	resourceCount := len(kubeitFileResources)
	if resourceCount == 0 {
		return []error{
			fmt.Errorf("no Kubeit resources found when traversing: %s", sourceConfigURI),
		}, nil
	}

	apis.LogResources(kubeitFileResources)

	var kubeitResourcesYaml strings.Builder

	for _, kubeitFileResource := range kubeitFileResources {
		// each Resource in kubeitFileResource is a type, we want to combine them all
		// to create a single yaml file string but with multiple YAML docs
		jsonString, err := json.Marshal(kubeitFileResource.Resource)
		if err != nil {
			return []error{err}, nil
		}

		yamlString, err := yaml.JSONToYAML(jsonString)
		if err != nil {
			return []error{err}, nil
		}

		kubeitResourcesYaml.WriteString("---\n")
		kubeitResourcesYaml.WriteString(string(yamlString))
	}

	// Base64-encode the YAML string
	encodedResources := base64.StdEncoding.EncodeToString([]byte(kubeitResourcesYaml.String()))

	// Generate the docker build command with multiple labels
	labels := []string{
		fmt.Sprintf("%s/version=%s", common.KubeitDomain, version.GetBuildInfo().Version),
		fmt.Sprintf("%s/resources=%s", common.KubeitDomain, encodedResources),
	}

	var labelArgs strings.Builder
	for _, label := range labels {
		labelArgs.WriteString(fmt.Sprintf("--label %s ", label))
	}

	fmt.Printf("%s", labelArgs.String())

	return nil, nil
}

func CliDocs(rootCmd *cobra.Command, generateSetOptions *Options) error {
	docsDir := filepath.Join(generateSetOptions.OutputDir, "cli", common.KubeitCLIName)
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	err := doc.GenMarkdownTree(rootCmd, docsDir)
	if err != nil {
		return fmt.Errorf("failed to generate CLI docs: %w", err)
	}

	fmt.Printf("CLI documentation generated in %s\n", docsDir)

	return nil
}

func Schemas() {
}
