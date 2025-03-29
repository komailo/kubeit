package v1

import (
	"github.com/komailo/kubeit/pkg/api"
)

type NamedValues struct {
	api.Resource `                json:",inline"`
	Spec         NamedValuesSpec `json:"spec"`
}

type NamedValuesSpec struct {
	Values []ValueEntry `json:"values" validate:"required"`
}
