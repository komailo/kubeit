package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
	helmCliValues "helm.sh/helm/v3/pkg/cli/values"

	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/internal/version"
	"github.com/komailo/kubeit/pkg/apis"
	helmvaluesv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_values/v1alpha1"
	namedvaluesv1alpha1 "github.com/komailo/kubeit/pkg/apis/named_values/v1alpha1"

	"github.com/komailo/kubeit/pkg/utils"
)

func generateHelmValues(
	values []helmvaluesv1alpha1.ValueEntry,
	namedValues []*namedvaluesv1alpha1.Values,
	loaderMeta *apis.LoaderMeta,
	generateSetOptions *Options,
) (helmCliValues.Options, error) {
	valuesFile, err := os.CreateTemp(generateSetOptions.WorkDir, "helm-values-*.yaml")
	if err != nil {
		return helmCliValues.Options{}, fmt.Errorf(
			"failed to create temporary file for helm values: %w",
			err,
		)
	}

	helmCliValuesOptions := helmCliValues.Options{
		ValueFiles:    []string{},
		StringValues:  []string{},
		Values:        []string{},
		FileValues:    []string{},
		JSONValues:    []string{},
		LiteralValues: []string{},
	}

	var jsonValues []json.RawMessage

	var processValues func(values []helmvaluesv1alpha1.ValueEntry) error
	processValues = func(values []helmvaluesv1alpha1.ValueEntry) error {
		for _, value := range values {
			switch value.Type {
			case "named":
				for _, namedValue := range namedValues {
					logger.Infof("Processing named values: %s", namedValue.Metadata.Name)

					if err := processValues(namedValue.Spec.Values); err != nil {
						return err
					}
				}
			case "raw":
				jsonValues = append(jsonValues, value.Data)
			case "mapping":
				mappingValues, err := generateValueMappings(value.Data, loaderMeta)
				if err != nil {
					return fmt.Errorf(
						"failed to generate value mappings: %w",
						err,
					)
				}

				helmCliValuesOptions.Values = append(helmCliValuesOptions.Values, mappingValues...)
			default:
				return fmt.Errorf("unsupported value type: %s", value.Type)
			}
		}

		return nil
	}

	if err := processValues(values); err != nil {
		return helmCliValuesOptions, err
	}

	// Create a YAML encoder
	yamlEncoder := yaml.NewEncoder(valuesFile)
	defer yamlEncoder.Close()

	// Convert and write each JSON object as a separate YAML document
	for _, jsonValue := range jsonValues {
		var yamlValue any

		// Unmarshal JSON into a generic interface
		if err := json.Unmarshal(jsonValue, &yamlValue); err != nil {
			return helmCliValuesOptions, fmt.Errorf("failed to unmarshal JSON value: %w", err)
		}

		// Encode YAML as a separate document
		if err := yamlEncoder.Encode(yamlValue); err != nil {
			return helmCliValuesOptions, fmt.Errorf("error encoding yaml: %w", err)
		}
	}

	if len(helmCliValuesOptions.Values) > 0 {
		logger.Infof(
			"Generated Helm set values from %d entries\n%s",
			len(helmCliValuesOptions.Values),
			strings.Join(helmCliValuesOptions.Values, "\n"),
		)
	}

	if len(jsonValues) > 0 {
		helmCliValuesOptions.ValueFiles = append(helmCliValuesOptions.ValueFiles, valuesFile.Name())

		_, err := valuesFile.Seek(0, 0)
		if err != nil {
			return helmCliValuesOptions, fmt.Errorf(
				"failed to seek to the beginning of the generated Helm values file: %w",
				err,
			)
		}

		valuesFileRead, err := os.ReadFile(valuesFile.Name())
		if err != nil {
			return helmCliValuesOptions, fmt.Errorf(
				"failed to read generated Helm values file: %w",
				err,
			)
		}

		logger.Infof("Generated Helm values from %d entries\n%s", len(jsonValues), valuesFileRead)
	}

	return helmCliValuesOptions, nil
}

// The function will substitute $VAR or ${VAR} with the actual value
func generateValueMappings(
	data json.RawMessage,
	loaderMeta *apis.LoaderMeta,
) ([]string, error) {
	var mappings stringMap
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mappings data: %w", err)
	}

	var setValues []string

	varPattern := regexp.MustCompile(`\$\{([^}]+)\}|\$([a-zA-Z_][a-zA-Z0-9_]*)`)

	var dockerRepo, dockerTag string

	var err error

	if loaderMeta.Scheme == "docker" {
		dockerRepo, dockerTag, err = utils.ParseDockerImage(loaderMeta.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Docker image: %w", err)
		}
	} else {
		dockerRepo = "$dockerImageRepository!!NOT_GENERATED_FROM_DOCKER_IMAGE!!"
		dockerTag = "$dockerImageTag!!NOT_GENERATED_FROM_DOCKER_IMAGE!!"
	}

	for key, value := range mappings {
		newValue := varPattern.ReplaceAllStringFunc(value, func(match string) string {
			// Extract variable name
			varName := strings.TrimPrefix(strings.Trim(match, "${}"), "$")
			switch varName {
			case "dockerImageRepository":
				return dockerRepo
			case "dockerImageTag":
				return dockerTag
			case "kubeitVersion":
				return version.GetBuildInfo().Version
			}
			// If not found, return the original match unchanged
			return match
		})

		setValues = append(setValues, fmt.Sprintf("%s=%s", key, newValue))
	}

	return setValues, nil
}
