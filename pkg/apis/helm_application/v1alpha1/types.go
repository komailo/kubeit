package v1alpha1

import (
	"encoding/json"
	"errors"
	"fmt"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/pkg/utils"
)

const (
	GroupVersion = common.APIVersionV1Alpha1
	Kind         = "HelmApplication"
)

var ValidValueTypes = []string{
	"env",
	"mapping",
	"raw",
}

type HelmApplication struct {
	k8smetav1.TypeMeta `         json:",inline"`
	Metadata           Metadata `json:"metadata,omitempty"`
	Spec               Spec     `json:"spec"`
}

type Metadata struct {
	Name string `json:"name" validate:"required"`
}

type Spec struct {
	Chart  Chart        `json:"chart"            validate:"required"`
	Values []ValueEntry `json:"values,omitempty"`
	// RawValues             any                   `json:"rawValues,omitempty"`
	// GenerateValueMappings `json:"generateValueMappings,omitempty"`
}

type Chart struct {
	Repository  string `json:"repository,omitempty"`
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Version     string `json:"version"`
	ReleaseName string `json:"releaseName"`
	Namespace   string `json:"namespace,omitempty"`
}

type ValueEntry struct {
	Type   string          `json:"type"             validate:"required"`
	Data   json.RawMessage `json:"data,omitempty"` // Handle different structures
	Source string          `json:"source,omitempty"`
}

type GenerateValueMappings map[string]string

// Method to get the API metadata
func (c *HelmApplication) GetAPIMetadata() k8smetav1.TypeMeta {
	return c.TypeMeta
}

// Custom validation function for HelmApplication
func (c *HelmApplication) Validate() error {
	if c.Spec.Chart.URL == "" && (c.Spec.Chart.Repository == "" || c.Spec.Chart.Name == "") {
		return errors.New(
			"either spec.chart.url must be provided or both spec.chart.repository and spec.chart.name must be provided",
		)
	}

	typesWithSource := []string{"env"}

	if c.Spec.Values != nil {
		for _, value := range c.Spec.Values {
			if !utils.Contains(ValidValueTypes, value.Type) {
				return fmt.Errorf("spec.values[*].type must be one of %s", ValidValueTypes)
			}

			if utils.Contains(typesWithSource, value.Type) {
				if value.Source == "" {
					return fmt.Errorf(
						"spec.values[*].source must be provided when type: %s",
						value.Type,
					)
				}
			} else if value.Data == nil {
				return fmt.Errorf("spec.values[*].data must be provided when type: %s", value.Type)
			}
		}
	}
	return nil
}

// Custom unmarshal function for ValueEntry
func (v *ValueEntry) UnmarshalJSON(data []byte) error {
	// Define an alias to avoid infinite recursion
	type Alias ValueEntry
	aux := &struct {
		Data json.RawMessage `json:"data,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(v),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	v.Data = aux.Data // Preserve raw data for later decoding

	// Validate Data type based on Type
	switch v.Type {
	case "mapping":
		var mappings map[string]string
		if err := json.Unmarshal(v.Data, &mappings); err != nil {
			return fmt.Errorf("invalid mappings format: %w", err)
		}
	case "raw":
		var raw map[string]any
		if err := json.Unmarshal(v.Data, &raw); err != nil {
			return fmt.Errorf("invalid raw format: %w", err)
		}
	case "env":
		if v.Source == "" {
			return errors.New("source field is required when type is 'env'")
		}
	default:
		return fmt.Errorf("unknown type: %s", v.Type)
	}

	return nil
}
