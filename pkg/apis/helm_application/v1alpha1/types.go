package v1alpha1

import (
	"errors"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/komailo/kubeit/common"
	helmvaluesv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_values/v1alpha1"
	metav1alpha1 "github.com/komailo/kubeit/pkg/apis/meta/v1alpha1"
)

const (
	GroupVersion = common.APIVersionV1Alpha1
	Kind         = "HelmApplication"
)

type HelmApplication struct {
	k8smetav1.TypeMeta `json:",inline"`
	Metadata           metav1alpha1.ObjectMeta `json:"metadata"`
	Spec               Spec                    `json:"spec"`
}

type Spec struct {
	Chart                   Chart                           `json:"chart"            validate:"required"`
	Values                  []helmvaluesv1alpha1.ValueEntry `json:"values,omitempty"`
	helmvaluesv1alpha1.Spec `json:",inline"`
}

type Chart struct {
	Repository  string `json:"repository,omitempty"`
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Version     string `json:"version"`
	ReleaseName string `json:"releaseName"`
	Namespace   string `json:"namespace,omitempty"`
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

	err := c.Spec.Validate()
	if err != nil {
		return err
	}

	return nil
}

// Method to get the metadata
func (c *HelmApplication) GetMetadata() metav1alpha1.ObjectMeta {
	return c.Metadata
}
