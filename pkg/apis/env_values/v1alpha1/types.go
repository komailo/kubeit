package v1alpha1

import (
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/komailo/kubeit/common"
	helmvaluesv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_values/v1alpha1"
	metav1alpha1 "github.com/komailo/kubeit/pkg/apis/meta/v1alpha1"
)

const (
	GroupVersion = common.APIVersionV1Alpha1
	Kind         = "EnvValues"
)

type Values struct {
	k8smetav1.TypeMeta `json:",inline"`
	Metadata           metav1alpha1.ObjectMeta `json:"metadata"`
	Spec               Spec                    `json:"spec"     validate:"required"`
}

type Spec struct {
	helmvaluesv1alpha1.Spec `json:",inline"`
}

// Method to get the API metadata
func (c *Values) GetAPIMetadata() k8smetav1.TypeMeta {
	return c.TypeMeta
}

// Custom validation function for HelmEnvValues
func (c *Values) Validate() error {
	return nil
}

// Method to get the metadata
func (c *Values) GetMetadata() metav1alpha1.ObjectMeta {
	return c.Metadata
}
