package apis

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/docker/docker/client"
	"github.com/go-playground/validator/v10"
	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/internal/logger"
	appv1alpha1 "github.com/komailo/kubeit/pkg/apis/application/v1alpha1"
	helmappv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_application/v1alpha1"
	"github.com/komailo/kubeit/pkg/utils"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gopkg.in/yaml.v3"
)

// TypeRegistry holds mappings of Kind -> APIVersion -> Struct Type
// This registry is used to dynamically load the correct struct type based on the
// API metadata (apiVersion and kind) extracted from the YAML data.
var TypeRegistry = map[string]map[string]reflect.Type{
	appv1alpha1.Kind: {
		appv1alpha1.GroupVersion: reflect.TypeOf(&appv1alpha1.Application{}),
	},
	helmappv1alpha1.Kind: {
		helmappv1alpha1.GroupVersion: reflect.TypeOf(&helmappv1alpha1.HelmApplication{}),
	},
}

// loadKubeitResource dynamically loads the correct struct based on the provided single
// YAML document.
// It extracts the API metadata (apiVersion and kind) to determine the appropriate
// struct type,
// unmarshals the data into the struct, validates it, and ensures it implements the
// KubeitResource interface.
//
// This function is used by LoadKubeitResources to process individual YAML documents.
// It is safe to use LoadKubeitResources even if you have a single YAML document.
//
// Parameters:
//   - data: A byte slice containing the YAML data to be processed.
//
// Returns:
//   - KubeitResource: The loaded and validated KubeitResource struct.
//   - error: An error encountered during the process, or nil if no error occurred.
//
// The function performs the following steps:
//  1. Extracts the API metadata (apiVersion and kind) from the YAML data.
//  2. Looks up the appropriate struct type based on the extracted metadata.
//  3. Creates a new instance of the struct and unmarshals the YAML data into it.
//  4. Validates the unmarshaled struct using the go-playground/validator library.
//  5. Ensures the struct implements the KubeitResource interface.
//  6. Sets the TypeMeta field of the struct using the extracted metadata.
func loadKubeitResource(data []byte) (KubeitResource, error) {
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

// loadKubeitResources loads Kubeit resources from a byte slice containing YAML data.
// It supports multi-document YAML files and processes each document to extract Kubeit
// resources.
//
// Parameters:
//   - data: A byte slice containing the YAML data to be processed.
//
// Returns:
//   - []KubeitResource: A slice of KubeitResource structs extracted from the YAML data.
//   - []error: A slice of errors encountered while processing the YAML data.
//
// The function performs the following steps:
//  1. Decodes the YAML data into individual documents.
//  2. Marshals each document back into YAML for processing.
//  3. Loads each document as a Kubeit resource using the LoadKubeitResource function.
//  4. Collects and returns any errors encountered during the process.
func loadKubeitResources(data []byte) ([]KubeitResource, []error) {
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
		resource, err := loadKubeitResource(yamlData)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to load resource: %w", err))
			continue
		}

		resources = append(resources, resource)
	}

	return resources, errors
}

// loadKubeitResourcesFromDir loads all Kubeit resources from a specified directory,
// supporting multi-document YAML files.
// It recursively traverses the directory tree, reading and processing YAML files to
// extract Kubeit resources.
//
// Parameters:
//   - dir: The root directory to start loading Kubeit resources from.
//
// Returns:
//   - []KubeitFileResource:	A slice of KubeitFileResource structs, each containing
//     the full file path, resource, and API metadata.
//   - map[string][]error: 		A map where the keys are file paths and the values are
//     slices of errors encountered while processing those files.
//
// The function performs the following steps:
//  1. Recursively walks through the directory tree starting from the specified root
//     directory.
//  2. Skips directories at the root level that start with a dot (e.g., .generated).
//  3. Reads and processes each YAML file to extract Kubeit resources.
//  4. Collects and returns any errors encountered during the process.
//
// Partially loaded resources are returned in case of errors. Always check the errors
// map to ensure all resources were loaded successfully.
func loadKubeitResourcesFromDir(dir string) ([]KubeitFileResource, map[string][]error) {
	var resources []KubeitFileResource
	errors := make(map[string][]error)

	absDirPath, err := filepath.Abs(dir)
	if err != nil {
		logger.Warnf("Failed to get absolute path for file %s", dir)
		errors[dir] = append(errors[dir], fmt.Errorf("failed to walk directory: %w", err))
		return nil, errors
	}

	err = filepath.Walk(absDirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			errors[filePath] = append(errors[filePath], fmt.Errorf("error accessing file: %w", err))
			return nil
		}

		// Skip directories at the root level that start with a dot
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && filepath.Dir(filePath) == filepath.Clean(dir) {
				logger.Debugf("Skiping root directory to load Kubeit resources from: %s", filePath)
				return filepath.SkipDir
			} else {
				logger.Debugf("Found directory to walk to Kubeit resources from: %s", filePath)
				return nil
			}
		}

		logger.Infof("Loading file: %s", filePath)

		data, err := os.ReadFile(filePath)
		if err != nil {
			errors[filePath] = append(errors[filePath], fmt.Errorf("failed to read file: %w", err))
			return nil
		}

		fileResources, fileErrors := loadKubeitResources(data)
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
		errors[absDirPath] = append(errors[absDirPath], fmt.Errorf("failed to walk directory: %w", err))
	}

	return resources, errors
}

// loadKubeitResourcesFromDockerImage loads Kubeit resources from a Docker image
// reference.
// It inspects the Docker image to extract labels and processes the labels to extract
// Kubeit resources.
//
// Parameters:
//   - imageRef: A string representing the Docker image reference.
//
// Returns:
//   - []KubeitFileResource: A slice of KubeitFileResource structs extracted from the
//     Docker image labels.
//   - map[string][]error: A map where the keys are image references and the values are
//     slices of errors encountered while processing the image.
func loadKubeitResourcesFromDockerImage(imageRef string) ([]KubeitFileResource, map[string][]error) {
	var resources []KubeitFileResource
	errors := make(map[string][]error)

	if exists, err := utils.CheckDockerImageExists(imageRef); !exists || err != nil {
		errors[imageRef] = append(errors[imageRef], fmt.Errorf("failed to find image: %w", err))
		return nil, errors
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		errors[imageRef] = append(errors[imageRef], fmt.Errorf("failed to create Docker client: %w", err))
		return nil, errors
	}

	imageInspect, _, err := cli.ImageInspectWithRaw(context.Background(), imageRef)
	if err != nil {
		errors[imageRef] = append(errors[imageRef], fmt.Errorf("failed to inspect Docker image: %w", err))
		return nil, errors
	}

	// check if kubeit.komail.io/resources is present in the labels
	labelKey := fmt.Sprintf(common.KubeitDomain + "/resources")
	base64Resource, ok := imageInspect.Config.Labels[labelKey]
	if !ok {
		errors[imageRef] = append(errors[imageRef], fmt.Errorf("no Kubeit resources found in image: %s", imageRef))
		return nil, errors
	}

	// decode the base64 encoded resources
	decodedResources, err := base64.StdEncoding.DecodeString(base64Resource)
	if err != nil {
		errors[imageRef] = append(errors[imageRef], fmt.Errorf("failed to decode base64 resources: %w", err))
		return nil, errors
	}

	logger.Debugf("Decoded resources:\n%s", decodedResources)

	fileResources, fileErrors := loadKubeitResources(decodedResources)
	if len(fileErrors) > 0 {
		errors[imageRef] = append(errors[imageRef], fileErrors...)
	}

	for _, resource := range fileResources {
		resources = append(resources, KubeitFileResource{
			FileName:    imageRef,
			Resource:    resource,
			APIMetadata: resource.GetAPIMetadata(),
		})
	}

	return resources, errors
}

// CountResources counts the number of Kubeit resources by their kind.
// It takes a slice of KubeitFileResource structs and returns a map where the keys are
// the kinds of resources and the values are the counts of each kind.
//
// Parameters:
//   - resources: A slice of KubeitFileResource structs to be counted.
//
// Returns:
//   - map[string]int: 	A map where the keys are resource kinds and the values are the
//     counts of each kind.
func CountResources(resources []KubeitFileResource) map[string]int {
	counts := make(map[string]int)
	for _, resource := range resources {
		counts[resource.APIMetadata.Kind]++
	}
	return counts
}

// LogResources logs the details of Kubeit resources.
// It logs the total number of resources found and the count of each kind of resource.
// Additionally, it logs detailed information about each resource in debug mode.
//
// Parameters:
//   - kubeitFileResources: A slice of KubeitFileResource structs to be logged.
//
// The function performs the following steps:
//  1. Counts the number of resources by their kind using the CountResources function.
//  2. Logs the count of each kind of resource.
//  3. Logs the total number of resources found.
//  4. Logs detailed information about each resource in debug mode, including the kind,
//     API version, and file name.
func LogResources(kubeitFileResources []KubeitFileResource) {
	resourceCount := len(kubeitFileResources)
	if resourceCount != 0 {
		kindCounts := CountResources(kubeitFileResources)
		for kind, count := range kindCounts {
			logger.Infof("%s: %d", kind, count)
		}

		logger.Infof("Found %d Kubeit resources", resourceCount)
	}

	for _, kubeitFileResource := range kubeitFileResources {
		logger.Debugf("Found resource Kind: %s, API Version: %s in file: %s", kubeitFileResource.APIMetadata.Kind, kubeitFileResource.APIMetadata.APIVersion, kubeitFileResource.FileName)
	}
}

// Loader loads Kubeit resources from a specified source configuration URI.
// It supports loading resources from local file directories.
//
// Parameters:
//   - sourceConfigUri: A string representing the source configuration URI. See
//     supported Schemes bellow.
//
// Returns:
//   - []KubeitFileResource: A slice of KubeitFileResource structs, each containing the
//     full file path, resource, and API metadata.
//   - error: Errors encountered during the process, or nil if no errors occurred.
//   - map[string][]error: A map where the keys are file paths and the values are slices
//     of errors encountered while processing those files. If errors are found, the
//     error in the second return will not be nil.
//
// Supported Schemes:
//   - file: Loads Kubeit resources from a local file directory.
//
// The function performs the following steps:
//  1. Parses the source configuration URI.
//  2. Depending on the source scheme it will load Kubeit resources from a local file
//     directory or a Docker image or any other supported scheme.
//  3. Returns the loaded resources, any errors encountered, and a map of
//     file-specific errors.
func Loader(sourceConfigUri string) ([]KubeitFileResource, error, map[string][]error) {
	sourceScheme, source, err := utils.SourceConfigUriParser(sourceConfigUri)
	if err != nil {
		return nil, err, nil
	}

	logger.Infof("Loading Kubeit resources from %s", sourceConfigUri)
	var kubeitFileResources []KubeitFileResource
	var loadErrs map[string][]error

	if sourceScheme == "file" {
		kubeitFileResources, loadErrs = loadKubeitResourcesFromDir(source)
	} else if sourceScheme == "docker" {
		kubeitFileResources, loadErrs = loadKubeitResourcesFromDockerImage(source)
	} else {
		return nil, fmt.Errorf("unsupported source config URI scheme: %s", sourceScheme), nil
	}

	if len(loadErrs) != 0 {
		errMsg := fmt.Sprintf("%d files have errors while loading Kubeit resources", len(loadErrs))
		logger.Error(errMsg)
		return nil, fmt.Errorf("%v", errMsg), loadErrs
	}

	resourceCount := len(kubeitFileResources)
	if resourceCount == 0 {
		return nil, fmt.Errorf("no Kubeit resources found when traversing: %s", sourceConfigUri), nil
	} else {
		kindCounts := CountResources(kubeitFileResources)
		for kind, count := range kindCounts {
			logger.Infof("%s: %d", kind, count)
		}

		logger.Infof("Found %d Kubeit resources", resourceCount)
	}

	for _, kubeitFileResource := range kubeitFileResources {
		logger.Debugf("Found resource Kind: %s, API Version: %s in file: %s", kubeitFileResource.APIMetadata.Kind, kubeitFileResource.APIMetadata.APIVersion, kubeitFileResource.FileName)
	}

	return kubeitFileResources, nil, nil
}
