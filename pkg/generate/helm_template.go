package generate

import (
	"github.com/komailo/kubeit/pkg/apis"
	helmappv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_application/v1alpha1"
)

func generateHelmTemplates(kubeitFileResources []apis.KubeitFileResource, loaderMeta apis.LoaderMeta, generateSetOptions *GenerateOptions) []error {
	var errs []error
	for _, kubeitFileResource := range kubeitFileResources {
		if kubeitFileResource.APIMetadata.Kind != helmappv1alpha1.Kind {
			continue
		}

		if kubeitFileResource.APIMetadata.APIVersion != helmappv1alpha1.GroupVersion {
			continue
		}

		if resource, ok := kubeitFileResource.Resource.(*helmappv1alpha1.HelmApplication); ok {
			err := GenerateManifestFromHelm(*resource, &loaderMeta, generateSetOptions)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}
