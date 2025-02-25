package v1alpha1

import (
	"errors"

	"github.com/komailo/kubeit/common"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GroupVersion = common.APIVersionV1Alpha1
const Kind = "HelmApplication"

type HelmApplication struct {
	k8smetav1.TypeMeta `json:",inline"`
	Metadata           Metadata `json:"metadata" yaml:"metadata"`
	Spec               Spec     `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Name string `json:"name" yaml:"name" validate:"required"`
}

type Spec struct {
	Chart     Chart `json:"chart" yaml:"chart" validate:"required"`
	RawValues any   `json:"rawValues" yaml:"rawValues"`
}

type Chart struct {
	Repository  string `json:"repository" yaml:"repository"`
	Name        string `json:"name" yaml:"name"`
	URL         string `json:"url" yaml:"url"`
	Version     string `json:"version" yaml:"version" validate:"required"`
	ReleaseName string `json:"releaseName" yaml:"releaseName" validate:"required"`
	Namespace   string `json:"namespace" yaml:"namespace"`
}

// Method to get the API metadata
func (h *HelmApplication) GetAPIMetadata() k8smetav1.TypeMeta {
	return h.TypeMeta
}

// Method to set the API metadata
func (h *HelmApplication) SetAPIMetadata(meta k8smetav1.TypeMeta) {
	h.TypeMeta = meta
}

// Custom validation function for HelmApplication
func (c *HelmApplication) Validate() error {
	if c.Spec.Chart.URL == "" && (c.Spec.Chart.Repository == "" || c.Spec.Chart.Name == "") {
		return errors.New("either spec.chart.url must be provided or both spec.chart.repository and spec.chart.name must be provided")
	}
	return nil
}
