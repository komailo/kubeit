package apis

import (
	"fmt"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metav1alpha1 "github.com/komailo/kubeit/pkg/apis/meta/v1alpha1"
)

// the base type that all Kubeit resources must implement
type KubeitResource interface {
	Validate() error
	GetAPIMetadata() k8smetav1.TypeMeta
	GetMetadata() metav1alpha1.ObjectMeta
}

// KubeitFileResource is a struct that holds a kubeit resources
// loaded from the file system
type KubeitFileResource struct {
	FileName    string
	Resource    KubeitResource
	APIMetadata k8smetav1.TypeMeta
}

type LoaderMeta struct {
	Source string
	Scheme string
}

type KubeitFileResources []KubeitFileResource

type KubeitResources []KubeitResource

// uniqueResources checks to see if KubeitResources does not have the same resource
// more than once. This is based on the resource's Kind and metadata.Name.
//
// Parameters:
//   - resources: A slice of KubeitResource structs to be checked.
//
// Returns:
//   - []error: A list of errors encountered while checking for unique resources.
func (resources KubeitResources) CheckResourceUniqueness() []error {
	var errs []error

	seen := make(map[string]bool)

	for _, resource := range resources {
		kind := resource.GetAPIMetadata().Kind
		name := resource.GetMetadata().Name
		uniqueKey := fmt.Sprintf("%s-%s", kind, name)

		if _, ok := seen[uniqueKey]; ok {
			errs = append(errs, fmt.Errorf("Resource %s with name %s is not unique", kind, name))
		}

		seen[uniqueKey] = true
	}

	return errs
}

// uniqueResources checks to see if KubeitResources does not have the same resource
// more than once. This is based on the resource's Kind and metadata.Name.
//
// Parameters:
//   - fileResources: A slice of KubeitFileResource structs to be checked.
//
// Returns:
//   - []error: A list of errors encountered while checking for unique resources.
func (fileResources KubeitFileResources) CheckResourceUniqueness() map[string][]error {
	errors := make(map[string][]error)

	seen := make(map[string]string)

	for _, fileResource := range fileResources {
		kind := fileResource.Resource.GetAPIMetadata().Kind
		name := fileResource.Resource.GetMetadata().Name
		uniqueKey := fmt.Sprintf("%s-%s", kind, name)

		if _, ok := seen[uniqueKey]; ok {
			errors[fileResource.FileName] = append(
				errors[fileResource.FileName],
				fmt.Errorf(
					"Resource %s with name %s is not unique. Already seen in: %s",
					kind,
					name,
					seen[uniqueKey],
				),
			)
		}

		seen[uniqueKey] = fileResource.FileName
	}

	return errors
}
