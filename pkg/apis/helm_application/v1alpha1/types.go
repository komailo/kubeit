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
	Metadata           Metadata `json:"metadata,omitempty"`
	Spec               Spec     `json:"spec"`
}

type Metadata struct {
	Name string `json:"name" validate:"required"`
}

type Spec struct {
	Chart                 Chart `json:"chart" validate:"required"`
	RawValues             any   `json:"rawValues,omitempty"`
	GenerateValueMappings any   `json:"generateValueMappings,omitempty"`
}

type Chart struct {
	Repository  string `json:"repository,omitempty"`
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Version     string `json:"version"`
	ReleaseName string `json:"releaseName"`
	Namespace   string `json:"namespace,omitempty"`
}

// Method to get the API metadata
func (h *HelmApplication) GetAPIMetadata() k8smetav1.TypeMeta {
	return h.TypeMeta
}

// Custom validation function for HelmApplication
func (c *HelmApplication) Validate() error {
	if c.Spec.Chart.URL == "" && (c.Spec.Chart.Repository == "" || c.Spec.Chart.Name == "") {
		return errors.New("either spec.chart.url must be provided or both spec.chart.repository and spec.chart.name must be provided")
	}
	return nil
}
