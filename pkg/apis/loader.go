package apis

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/komailo/kubeit/internal/logger"
	appv1alpha1 "github.com/komailo/kubeit/pkg/apis/application/v1alpha1"
	helmappv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_application/v1alpha1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gopkg.in/yaml.v3"
)

// TypeRegistry holds mappings of Kind -> APIVersion -> Struct Type
var TypeRegistry = map[string]map[string]reflect.Type{
	appv1alpha1.Kind: {
		appv1alpha1.GroupVersion: reflect.TypeOf(&appv1alpha1.Application{}),
	},
	helmappv1alpha1.Kind: {
		helmappv1alpha1.GroupVersion: reflect.TypeOf(&helmappv1alpha1.HelmApplication{}),
	},
}

// LoadKubeitResource dynamically loads the correct struct
func LoadKubeitResource(data []byte) (KubeitResource, error) {
	var metaOnly struct {
		APIVersion string `json:"apiVersion" yaml:"apiVersion"`
		Kind       string `json:"kind" yaml:"kind"`
	}

	// Extract API metadata first
	if err := yaml.Unmarshal(data, &metaOnly); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	if metaOnly.APIVersion == "" || metaOnly.Kind == "" {
		return nil, fmt.Errorf("missing apiVersion or kind in resource")
	}

	// Lookup the resource type
	kindRegistry, kindExists := TypeRegistry[metaOnly.Kind]
	if !kindExists {
		return nil, fmt.Errorf("unknown resource kind: %s", metaOnly.Kind)
	}
	resourceType, versionExists := kindRegistry[metaOnly.APIVersion]
	if !versionExists {
		return nil, fmt.Errorf("unsupported apiVersion: %s", metaOnly.APIVersion)
	}

	// Create a new instance of the resource
	resourceInstance := reflect.New(resourceType.Elem()).Interface()

	// Unmarshal data into the correct struct
	if err := yaml.Unmarshal(data, resourceInstance); err != nil {
		return nil, fmt.Errorf("failed to parse resource: %w", err)
	}

	// Validate the full struct
	validate := validator.New()
	if err := validate.Struct(resourceInstance); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Ensure resourceInstance implements KubeitResource
	res, ok := resourceInstance.(KubeitResource)
	if !ok {
		return nil, fmt.Errorf("failed to assert resource as KubeitResource, got: %T", resourceInstance)
	}

	// Set TypeMeta using the new interface method
	res.SetAPIMetadata(k8smetav1.TypeMeta{
		APIVersion: metaOnly.APIVersion,
		Kind:       metaOnly.Kind,
	})

	return res, nil
}

func LoadKubeitResources(data []byte) ([]KubeitResource, []error) {
	var resources []KubeitResource
	var errors []error

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var rawDoc map[string]interface{}
		if err := decoder.Decode(&rawDoc); err != nil {
			if err == io.EOF {
				break // End of YAML documents
			}
			errors = append(errors, fmt.Errorf("failed to decode YAML document: %w", err))
			break
		}

		// Marshal the individual document back into YAML for processing
		yamlData, _ := yaml.Marshal(rawDoc)
		resource, err := LoadKubeitResource(yamlData)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to load resource: %w", err))
			continue
		}

		resources = append(resources, resource)
	}

	return resources, errors
}

// LoadKubeitResourcesFromDir loads all resources from a directory, supporting multi-document YAML files
func LoadKubeitResourcesFromDir(dir string) ([]KubeitFileResource, map[string][]error) {
	var resources []KubeitFileResource
	errors := make(map[string][]error)

	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			errors[filePath] = append(errors[filePath], fmt.Errorf("error accessing file: %w", err))
			return nil
		}

		// Skip directories at the root level that start with a dot
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && filepath.Dir(filePath) == filepath.Clean(dir) {
			logger.Debugf("Skiping directory to load Kubeit resources from: %s", filePath)
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		log.Printf("Loading file: %s", filePath)

		data, err := os.ReadFile(filePath)
		if err != nil {
			errors[filePath] = append(errors[filePath], fmt.Errorf("failed to read file: %w", err))
			return nil
		}

		fileResources, fileErrors := LoadKubeitResources(data)
		if len(fileErrors) > 0 {
			errors[filePath] = append(errors[filePath], fileErrors...)
		}

		for _, resource := range fileResources {
			resources = append(resources, KubeitFileResource{
				FileName:    filePath,
				Resource:    resource,
				APIMetadata: resource.GetAPIMetadata(),
			})
		}

		return nil
	})

	if err != nil {
		errors[dir] = append(errors[dir], fmt.Errorf("failed to walk directory: %w", err))
	}

	return resources, errors
}

// count the resources by kind
func CountResources(resources []KubeitFileResource) map[string]int {
	counts := make(map[string]int)
	for _, resource := range resources {
		counts[resource.APIMetadata.Kind]++
	}
	return counts
}
