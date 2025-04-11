package v1

import (
	"github.com/scorebet/reflow/pkg/api"
)

type NamedValues struct {
	api.BaseObject `                json:",inline"`
	Spec           NamedValuesSpec `json:"spec"`
}

type NamedValuesSpec struct {
	Values []ValueEntry `json:"values" validate:"required"`
}
