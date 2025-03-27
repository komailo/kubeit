package v1alpha1

import "time"

type ObjectMeta struct {
	Name string `json:"name" validate:"required"`
}

type SourceMeta struct {
	SourceURI string `json:"sourceURI" validate:"required"`
	Source    string `json:"source"    validate:"required"`
	Scheme    string `json:"scheme"    validate:"required"`
}

type LoaderMeta struct {
	LoadedOn time.Time `json:"loadTime" validate:"required"`
}
