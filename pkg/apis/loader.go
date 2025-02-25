package apis

import (
	"bytes"
	"fmt"

	"github.com/go-playground/validator/v10"
	appv1alpha1 "github.com/komailo/kubeit/pkg/apis/application/v1alpha1"
	"gopkg.in/yaml.v3"
)

// APIMetadata extracts apiVersion and kind before parsing the full struct
type APIMetadata struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion" validate:"required"`
	Kind       string `json:"kind" yaml:"kind" validate:"required"`
}

// LoadKubeitResource dynamically loads the correct struct
func LoadKubeitResource(data []byte) (interface{}, error) {
	// Step 1: Check if the data is valid YAML
	var fullYaml map[string]interface{}
	err := yaml.Unmarshal(data, &fullYaml)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Step 2: Extract API metadata
	var apiMetadata APIMetadata
	err = yaml.Unmarshal(data, &apiMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to extract api metadata: %w", err)
	}

	// Step 3: Validate required fields in APIMetadata
	validate := validator.New()
	if err := validate.Struct(apiMetadata); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Step 4: Clean up the API metadata keys
	delete(fullYaml, "apiVersion")
	delete(fullYaml, "kind")
	apiData, _ := yaml.Marshal(fullYaml)

	// Step 5: Process valid API versions and Kind
	var resource interface{}
	switch apiMetadata.Kind {
	case appv1alpha1.Kind:
		switch apiMetadata.APIVersion {
		case appv1alpha1.GroupVersion:
			resource = &appv1alpha1.Application{}
		default:
			return nil, fmt.Errorf("unsupported apiVersion: %s", apiMetadata.APIVersion)
		}
	default:
		return nil, fmt.Errorf("unknown resource kind: %s", apiMetadata.Kind)
	}

	// Step 6: Strictly unmarshal the remaining data into the correct struct
	decoder := yaml.NewDecoder(bytes.NewReader(apiData))
	decoder.KnownFields(true)
	if err := decoder.Decode(resource); err != nil {
		return nil, fmt.Errorf("failed to parse resource strictly: %w", err)
	}

	// Step 7: Validate required fields in the full struct
	if err := validate.Struct(resource); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return resource, nil
}
