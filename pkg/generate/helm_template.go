package generate

import (
	"github.com/komailo/kubeit/pkg/apis"
	helmappv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_application/v1alpha1"
)

func generateHelmTemplates(
	kubeitFileResources []apis.KubeitFileResource,
	loaderMeta apis.LoaderMeta,
	generateSetOptions *Options,
) []error {
	var errs []error
	// var helmEnvValues []*helmenvvaluesv1alpha1.Values

	// for _, kubeitFileResource := range kubeitFileResources {
	// 	if kubeitFileResource.APIMetadata.Kind != helmenvvaluesv1alpha1.Kind {
	// 		continue
	// 	}

	// 	if kubeitFileResource.APIMetadata.APIVersion != helmenvvaluesv1alpha1.GroupVersion {
	// 		continue
	// 	}

	// 	if values, ok := kubeitFileResource.Resource.(*helmenvvaluesv1alpha1.Values); ok {
	// 		helmEnvValues = append(helmEnvValues, values)
	// 	}
	// }

	for _, kubeitFileResource := range kubeitFileResources {
		if kubeitFileResource.APIMetadata.Kind != helmappv1alpha1.Kind {
			continue
		}

		if kubeitFileResource.APIMetadata.APIVersion != helmappv1alpha1.GroupVersion {
			continue
		}

		if resource, ok := kubeitFileResource.Resource.(*helmappv1alpha1.HelmApplication); ok {
			err := ManifestFromHelm(*resource, &loaderMeta, generateSetOptions)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}
