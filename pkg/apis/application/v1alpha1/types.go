package v1alpha1

import (
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metav1alpha1 "github.com/komailo/kubeit/pkg/apis/meta/v1alpha1"
)

const (
	GroupVersion = "kubeit.komailo.github.io/v1alpha1"
	Kind         = "Application"
)

type Application struct {
	k8smetav1.TypeMeta `json:",inline"`
	Metadata           metav1alpha1.ObjectMeta `json:"metadata"`
	Spec               any                     `json:"spec"`
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

// Method to get the metadata
func (c *Application) GetMetadata() metav1alpha1.ObjectMeta {
	return c.Metadata
}
