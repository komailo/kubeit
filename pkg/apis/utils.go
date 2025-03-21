package apis

import (
	"reflect"

	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/pkg/utils"
)

func FilterKubeitFileResources[T any](
	fileResources KubeitFileResources,
	kind string,
	apiVersion string,
	metadataNames []string,
) []T {
	var extractedResources []T

	for _, fileResource := range fileResources {
		rkind, rapiVersion := fileResource.Resource.GetAPIMetadata().Kind, fileResource.Resource.GetAPIMetadata().APIVersion
		if rkind != kind {
			logger.Debugf("filter is skipping resource of kind: %s", rkind)
			continue
		}

		if rapiVersion != apiVersion {
			logger.Debugf("filter is skipping resource of apiVersion: %s", rapiVersion)
			continue
		}

		if metadataNames != nil {
			metadataName := fileResource.Resource.GetMetadata().Name
			if !utils.Contains(metadataNames, metadataName) {
				logger.Debugf("filter is skipping resource of metadata.name: %s", metadataName)
				continue
			}
		}

		logger.Debugf("Processing resource of type: %s", reflect.TypeOf(fileResource.Resource))

		// Correctly assert the type assuming T is a pointer type
		if resource, ok := fileResource.Resource.(T); ok {
			extractedResources = append(extractedResources, resource)
		}
	}

	if len(extractedResources) == 0 {
		logger.Debugf("No resources found for type %T", new(T))
	}

	return extractedResources
}
