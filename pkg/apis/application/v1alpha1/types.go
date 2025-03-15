package v1alpha1

import (
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupVersion = "kubeit.komailo.github.io/v1alpha1"
	Kind         = "Application"
)

type Application struct {
	Metadata Metadata `json:"metadata"`
	Spec     any      `json:"spec"`
	k8smetav1.TypeMeta
}

type Metadata struct {
	Name string `json:"name" validate:"required"`
}

// Method to get the API metadata
func (c *Application) GetAPIMetadata() k8smetav1.TypeMeta {
	return c.TypeMeta
}

// Custom validation function for Application
func (c *Application) Validate() error {
	return nil
}
