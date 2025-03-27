package v1

import (
	"github.com/komailo/kubeit/pkg/api"
)

type NamedValues struct {
	api.Resource `                json:",inline"`
	Spec         NamedValuesSpec `json:"spec"`
}

type NamedValuesSpec struct {
	Values []ValueEntry `json:"values" validate:"required"`
}

// Custom validation function for HelmEnvValues
func (c *NamedValues) Validate() error {
	return nil
}

// GetMetadata returns the metadata of the NamedValues
func (c *NamedValues) GetMetadata() api.ObjectMeta {
	return c.Metadata
}
