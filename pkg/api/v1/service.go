package v1

import (
	"fmt"

	"github.com/scorebet/reflow/pkg/api"
	"github.com/scorebet/reflow/pkg/utils"
)

var validOrgs = []string{
	"bet",
	"data-eng",
	"infra",
	"media",
}

var validEnvironments = []string{
	"audit1",
	"demo",
	"internal-services",
	"production",
	"ps",
	"staging",
	"uat",
}

var prodEnvironments = []string{
	"production",
}

type Service struct {
	api.BaseObject
	Spec ServiceSpec `json:"spec"`
}

type ServiceSpec struct {
	Environments []string `json:"environments" validate:"required"`

	Org string `json:"organization" validate:"required"`

	Regulated bool `json:"regulated,omitempty"`

	SourceRepo string `json:"sourceRepository" validate:"required"`
}

func (c Service) Validate() error {
	if !utils.Contains(validOrgs, c.Spec.Org) {
		return fmt.Errorf(
			"invalid organization provided: %s. Valid organizations: %s",
			c.Spec.Org,
			validOrgs,
		)
	}

	// Check if the environments are valid
	for _, env := range c.Spec.Environments {
		if !utils.Contains(validEnvironments, env) {
			return fmt.Errorf(
				"invalid environment provided: %s. Valid environments: %s",
				env,
				validEnvironments,
			)
		}
	}

	return nil
}

func (c Service) IsProd(env string) bool {
	return utils.Contains(prodEnvironments, env)
}
