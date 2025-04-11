package generate

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/scorebet/reflow/common"
	"github.com/scorebet/reflow/internal/logger"
	"github.com/scorebet/reflow/internal/version"
	"github.com/scorebet/reflow/pkg/api/loader"
)

func Manifests(generateSetOptions *Options) ([]error, map[string][]error) {
	sourceConfigURI := generateSetOptions.SourceConfigURI
	logger.Infof("Generating manifests from %s", sourceConfigURI)

	loaderInt := loader.NewLoader()
	loaderErr := loaderInt.FromSourceURI(sourceConfigURI)

	if len(loaderErr) != 0 {
		return nil, loaderErr
	}

	if loaderInt.ResourceCount == 0 {
		return []error{
			fmt.Errorf("no %s resources found when traversing: %s", common.AppName, sourceConfigURI),
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
	generateSetOptions *Options,
) (string, []error, map[string][]error) {
	sourceConfigURI := generateSetOptions.SourceConfigURI
	logger.Infof("Generating Docker Labels from %s", sourceConfigURI)

	loaderInt := loader.NewLoader()
	loaderErr := loaderInt.FromSourceURI(sourceConfigURI)

	if len(loaderErr) != 0 {
		return "", nil, loaderErr
	}

	if loaderInt.ResourceCount == 0 {
		return "", []error{
			fmt.Errorf("no %s resources found when traversing: %s", common.AppName, sourceConfigURI),
		}, nil
	}

	loaderInt.LogResources()

	marshalString, marshalErr := loaderInt.Marshal()
	if marshalErr != nil {
		return "", marshalErr, nil
	}

	// Base64-encode the YAML string
	encodedResources := base64.StdEncoding.EncodeToString([]byte(marshalString.String()))

	// Generate the docker build command with multiple labels
	labels := []string{
		fmt.Sprintf("%s/version=%s", common.ServiceDomain, version.GetBuildInfo().Version),
		fmt.Sprintf("%s/resources=%s", common.ServiceDomain, encodedResources),
	}

	var labelArgs strings.Builder
	for _, label := range labels {
		labelArgs.WriteString(fmt.Sprintf("--label %s ", label))
	}

	return labelArgs.String(), nil, nil
}

func CliDocs(rootCmd *cobra.Command, generateSetOptions *Options) error {
	docsDir := filepath.Join(generateSetOptions.OutputDir, "cli", common.MainCLIName)
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
