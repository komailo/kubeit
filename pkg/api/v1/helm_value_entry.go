package v1

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/pkg/api"
	"github.com/komailo/kubeit/pkg/utils"
)

const (
	GroupVersion = common.APIVersionV1Alpha1
	Kind         = "HelmValues"
)

var ValidValueTypes = []string{
	"named",
	"mapping",
	"raw",
}

var typesWithNoData = []string{"named"}

type HelmValues struct {
	api.BaseObject `               json:",inline"`
	Spec           HelmValuesSpec `json:"spec"`
}

type HelmValuesSpec struct {
	Values []ValueEntry `json:"values,omitempty"`
}

type ValueEntry struct {
	Type string          `json:"type"           validate:"required"`
	Data json.RawMessage `json:"data,omitempty"` // Handle different structures
}

type GenerateValueMappings map[string]string

// Custom validation function for HelmValues
func (c HelmValues) Validate() error {
	return nil
}

func (c HelmValuesSpec) Validate() error {
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
	case "named":
		if v.Data != nil {
			return errors.New("named values must not have data")
		}

	default:
		return fmt.Errorf("unknown Helm values type: %s", v.Type)
	}

	return nil
}
