package apis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metav1alpha1 "github.com/komailo/kubeit/pkg/apis/meta/v1alpha1"
)

type MockKubeitResource struct {
	Name       string
	Kind       string
	APIVersion string
}

func (m MockKubeitResource) Validate() error {
	return nil
}

func (m MockKubeitResource) GetAPIMetadata() k8smetav1.TypeMeta {
	return k8smetav1.TypeMeta{Kind: m.Kind, APIVersion: m.APIVersion}
}

func (m MockKubeitResource) GetMetadata() metav1alpha1.ObjectMeta {
	return metav1alpha1.ObjectMeta{Name: m.Name}
}

func TestFilterKubeitFileResources(t *testing.T) {
	resources := KubeitFileResources{
		{
			FileName: "file1.yaml",
			Resource: MockKubeitResource{Name: "resource1", Kind: "Kind1", APIVersion: "v1"},
		},
		{
			FileName: "file2.yaml",
			Resource: MockKubeitResource{Name: "resource2", Kind: "Kind1", APIVersion: "v1"},
		},
		{
			FileName: "file3.yaml",
			Resource: MockKubeitResource{Name: "resource3", Kind: "Kind2", APIVersion: "v2"},
		},
		{
			FileName: "file3.yaml",
			Resource: MockKubeitResource{Name: "resource4", Kind: "Kind2", APIVersion: "v2"},
		},
	}

	// Case 1: Filter with a matching name
	filtered := FilterKubeitFileResources[MockKubeitResource](
		resources,
		"Kind1",
		"v1",
		[]string{"resource1", "resource2"},
	)
	assert.Len(t, filtered, 2)
	assert.Equal(t, "resource1", filtered[0].Name)
	assert.Equal(t, "resource2", filtered[1].Name)

	// Case 2: Filter with a non-matching name
	filtered = FilterKubeitFileResources[MockKubeitResource](
		resources,
		"Kind1",
		"v1",
		[]string{"dne"},
	)
	assert.Empty(t, filtered)

	// Case 3: No filtering (nil metadataNames)
	filtered = FilterKubeitFileResources[MockKubeitResource](resources, "Kind2", "v2", nil)
	assert.Len(t, filtered, 2)
	assert.Equal(t, "resource3", filtered[0].Name)
	assert.Equal(t, "resource4", filtered[1].Name)
}
