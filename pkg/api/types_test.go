package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResource_GetTypeMeta(t *testing.T) {
	typeMeta := k8smetav1.TypeMeta{
		Kind:       "TestKind",
		APIVersion: "v1",
	}
	resource := BaseObject{
		TypeMeta: typeMeta,
	}

	result := resource.GetTypeMeta()

	assert.Equal(t, typeMeta, result, "GetTypeMeta should return the correct TypeMeta")
}

func TestResource_GetResourceMeta(t *testing.T) {
	objectMeta := ObjectMeta{
		Name: "test-object",
	}
	resource := BaseObject{
		Metadata: objectMeta,
	}

	result := resource.GetObjectMeta()

	assert.Equal(t, objectMeta, result, "GetObjectMeta should return the correct ObjectMeta")
}

func TestResource_GetSourceMeta(t *testing.T) {
	sourceMeta := SourceMeta{
		SourceURI: "http://example.com",
		Source:    "example-source",
		Scheme:    "https",
	}
	resource := BaseObject{
		SourceMeta: sourceMeta,
	}

	result := resource.GetSourceMeta()

	assert.Equal(t, sourceMeta, result, "GetSourceMeta should return the correct SourceMeta")
}

func TestResourceValidate(t *testing.T) {
	sourceMeta := SourceMeta{
		SourceURI: "http://example.com",
		Source:    "example-source",
		Scheme:    "https",
	}
	resource := BaseObject{
		SourceMeta: sourceMeta,
	}

	err := resource.Validate()

	assert.NoError(
		t,
		err,
		"Validate should return no error as its not implemented by base Resource",
	)
}

func TestResource_ImplementsObjectInterface(t *testing.T) {
	resource := &BaseObject{}

	var obj Object = resource // This will fail to compile if Resource does not implement Object

	assert.NotNil(t, obj, "Resource should implement the Object interface")
}
