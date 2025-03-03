package v1alpha1

import (
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GroupVersion = "kubeit.komailo.github.io/v1alpha1"
const Kind = "Application"

type Application struct {
	Metadata Metadata `json:"metadata" yaml:"metadata"`
	Spec     any      `json:"spec" yaml:"spec"`
	k8smetav1.TypeMeta
}

type Metadata struct {
	Name string `json:"name" yaml:"name" validate:"required"`
}

// Method to get the API metadata
func (h *Application) GetAPIMetadata() k8smetav1.TypeMeta {
	return h.TypeMeta
}

// Method to set the API metadata
func (h *Application) SetAPIMetadata(meta k8smetav1.TypeMeta) {
	h.TypeMeta = meta
}

// Custom validation function for Application
func (c *Application) Validate() error {
	return nil
}
