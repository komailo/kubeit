package v1alpha1

import (
	"encoding/json"
	"fmt"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/komailo/kubeit/common"
	metav1alpha1 "github.com/komailo/kubeit/pkg/apis/meta/v1alpha1"
	"github.com/komailo/kubeit/pkg/utils"
)

const (
	GroupVersion = common.APIVersionV1Alpha1
	Kind         = "HelmValues"
)

var ValidValueTypes = []string{
	"env",
	"mapping",
	"raw",
}

var typesWithNoData = []string{"env"}

type HelmValues struct {
	k8smetav1.TypeMeta `json:",inline"`
	Metadata           metav1alpha1.ObjectMeta `json:"metadata"`
	Spec               Spec                    `json:"spec"`
}

type Spec struct {
	Values []ValueEntry `json:"values,omitempty"`
}

type ValueEntry struct {
	Type string          `json:"type"           validate:"required"`
	Data json.RawMessage `json:"data,omitempty"` // Handle different structures
}

type GenerateValueMappings map[string]string

// Method to get the API metadata
func (c *HelmValues) GetAPIMetadata() k8smetav1.TypeMeta {
	return c.TypeMeta
}

// Custom validation function for HelmValues
func (c *HelmValues) Validate() error {
	return nil
}

func (c *Spec) Validate() error {
	if c.Values != nil {
		for _, value := range c.Values {
			if !utils.Contains(ValidValueTypes, value.Type) {
				return fmt.Errorf("spec.values[*].type must be one of %s", ValidValueTypes)
			}

			if utils.Contains(typesWithNoData, value.Type) {
				if value.Data != nil {
					return fmt.Errorf(
						"spec.values[*].data must NOT be provided when type: %s",
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

	default:
		return fmt.Errorf("unknown type: %s", v.Type)
	}

	return nil
}

// Method to get the metadata
func (c *HelmValues) GetMetadata() metav1alpha1.ObjectMeta {
	return c.Metadata
}
