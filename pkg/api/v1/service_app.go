package v1

import (
	"errors"
	"fmt"

	"github.com/scorebet/reflow/pkg/api"
	"github.com/scorebet/reflow/pkg/utils"
)

type ServiceApp struct {
	api.BaseObject
	Spec ServiceAppSpec `json:"spec"`
}

type ServiceAppSpec struct {
	ChartName        string   `json:"chartName,omitempty"`
	ChartVersion     string   `json:"chartVersion,omitempty"`
	ClusterRole      string   `json:"clusterRole,omitempty"  validate:"required"`
	Jurisdiction     string   `json:"jurisdiction,omitempty"`
	ServiceName      string   `json:"serviceName"            validate:"required"`
	ValuesFiles      []string `json:"valuesFiles,omitempty"`
	ValuesRepository string   `json:"valuesRepository"       validate:"required"`
}

var validClusterRoles = []string{
	"internal-services",
	"core",
	"edge",
	"data-engineering",
}

func (c ServiceApp) Validate() error {
	// Check if the cluster role is valid
	if !utils.Contains(validClusterRoles, c.Spec.ClusterRole) {
		return fmt.Errorf(
			"invalid cluster role provided: %s. Valid cluster roles: %s",
			c.Spec.ClusterRole,
			validClusterRoles,
		)
	}

	// require jurisdiction if cluster role is "edge"
	if c.Spec.ClusterRole == "edge" && c.Spec.Jurisdiction == "" {
		return errors.New("jurisdiction is required when cluster role is 'edge'")
	}

	return nil
}
