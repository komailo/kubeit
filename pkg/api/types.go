package api

import (
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Object is an interface that all API types must implement. Some are implemeted by BaseObjects
type Object interface {
	GetObjectMeta() ObjectMeta
	GetTypeMeta() k8smetav1.TypeMeta
	GetSourceMeta() SourceMeta
	Validate() error
}

// ObjectMeta contains metadata about the object
type ObjectMeta struct {
	Name string `json:"name"`
}

// BaseObject is the base type that all resource types must embed
type BaseObject struct {
	k8smetav1.TypeMeta `json:",inline"`
	Metadata           ObjectMeta `json:"metadata"`
	SourceMeta         SourceMeta
	Spec               any `json:"spec,omitempty"`
}

// SourceMeta contains metadata about the source of the object
type SourceMeta struct {
	SourceURI string
	Source    string
	Scheme    string
}

// GetTypeMeta implements the Object interface
func (r BaseObject) GetTypeMeta() k8smetav1.TypeMeta {
	return r.TypeMeta
}

// GetObjectMeta implements the Object interface
func (r BaseObject) GetObjectMeta() ObjectMeta {
	return r.Metadata
}

func (r BaseObject) GetSourceMeta() SourceMeta {
	return r.SourceMeta
}

func (r BaseObject) Validate() error {
	return nil
}
