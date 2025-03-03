package apis

import (
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// the base type that all Kubeit resources must implement
type KubeitResource interface {
	GetAPIMetadata() k8smetav1.TypeMeta
	SetAPIMetadata(meta k8smetav1.TypeMeta)
}

// KubeitFileResource is a struct that holds a kubeit resources
// loaded from the file system
type KubeitFileResource struct {
	FileName    string
	Resource    KubeitResource
	APIMetadata k8smetav1.TypeMeta
}
