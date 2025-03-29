package api

import (
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Object is an interface that all API types must implement
type Object interface {
	GetObjectMeta() ResourceMeta
	GetTypeMeta() k8smetav1.TypeMeta
	GetSourceMeta() SourceMeta
	Validate() error
}

// ResourceMeta contains metadata about the object
type ResourceMeta struct {
	Name string `json:"name"`
}

// Resource is the base type that all resource types should embed
type Resource struct {
	k8smetav1.TypeMeta `json:",inline"`
	Metadata           ResourceMeta `json:"metadata"`
	SourceMeta         SourceMeta
}

// SourceMeta contains metadata about the source of the object
type SourceMeta struct {
	SourceURI string
	Source    string
	Scheme    string
}

// GetTypeMeta implements the Object interface
func (r Resource) GetTypeMeta() k8smetav1.TypeMeta {
	return r.TypeMeta
}

// GetObjectMeta implements the Object interface
func (r Resource) GetObjectMeta() ResourceMeta {
	return r.Metadata
}

func (r Resource) GetSourceMeta() SourceMeta {
	return r.SourceMeta
}

func (r Resource) Validate() error {
	return nil
}
