package v1

import (
	"errors"

	"github.com/komailo/kubeit/pkg/api"
)

// HelmApplication represents a Helm-based application deployment
type HelmApplication struct {
	api.Resource `                    json:",inline"`
	Spec         HelmApplicationSpec `json:"spec"`
}

// HelmApplicationSpec defines the desired state of HelmApplication
type HelmApplicationSpec struct {
	Chart  ChartSpec    `json:"chart"            validate:"required"`
	Values []ValueEntry `json:"values,omitempty"`
}

type ChartSpec struct {
	Repository  string `json:"repository,omitempty"`
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Version     string `json:"version"`
	ReleaseName string `json:"releaseName"`
	Namespace   string `json:"namespace,omitempty"`
}

// Custom validation function for HelmApplication
func (c HelmApplication) Validate() error {
	if c.Spec.Chart.URL == "" && (c.Spec.Chart.Repository == "" || c.Spec.Chart.Name == "") {
		return errors.New(
			"either spec.chart.url must be provided or both spec.chart.repository and spec.chart.name must be provided",
		)
	}

	return nil
}

// GetMetadata returns the metadata of the NamedValues
func (c HelmApplication) GetMetadata() api.ObjectMeta {
	return c.Metadata
}
