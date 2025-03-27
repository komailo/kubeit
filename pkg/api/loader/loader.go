package loader

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/docker/docker/client"

	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/pkg/api"
	v1 "github.com/komailo/kubeit/pkg/api/v1"
	"github.com/komailo/kubeit/pkg/utils"
)

// Loader handles the decoding of YAML/JSON documents into API objects
type Loader struct {
	// registry maps Kind and Version to the corresponding type
	registry         map[string]map[string]api.Object
	SourceMeta       api.SourceMeta
	HelmApplications []*v1.HelmApplication
	NamedValues      []*v1.NamedValues
	KindsCount       map[string]int
	ResourceCount    int
}

// NewDecoder creates a new decoder with registered types
func NewLoader() *Loader {
	l := &Loader{
		registry:   make(map[string]map[string]api.Object),
		KindsCount: make(map[string]int),
	}

	// Register v1 types
	l.RegisterType("HelmApplication", "kubeit.komailo.github.io/v1alpha1", &v1.HelmApplication{})
	l.RegisterType("NamedValues", "kubeit.komailo.github.io/v1alpha1", &v1.NamedValues{})

	return l
}

// RegisterType registers a new type with the decoder
func (l *Loader) RegisterType(kind, version string, obj api.Object) {
	if _, ok := l.registry[kind]; !ok {
		l.registry[kind] = make(map[string]api.Object)
	}

	l.registry[kind][version] = obj
}

// TypeMeta is used to determine the type of resource
type TypeMeta struct {
	Kind       string `json:"kind"       yaml:"kind"`
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
}

func (l *Loader) Unmarshal(data []byte) error {
	// First decode just the TypeMeta to determine the type
	var meta TypeMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("failed to decode type metadata: %w", err)
	}

	// Look up the registered type
	versionMap, ok := l.registry[meta.Kind]
	if !ok {
		return fmt.Errorf("unknown kind: %s", meta.Kind)
	}

	prototypeObj, ok := versionMap[meta.APIVersion]
	if !ok {
		return fmt.Errorf("unknown version %s for kind %s", meta.APIVersion, meta.Kind)
	}

	// Unmarshal into the registered type
	if err := yaml.Unmarshal(data, prototypeObj); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", meta.Kind, err)
	}

	// Set source metadata and add to appropriate collection
	switch typedObj := prototypeObj.(type) {
	case *v1.HelmApplication:
		typedObj.SourceMeta = l.SourceMeta
		l.HelmApplications = append(l.HelmApplications, typedObj)
	case *v1.NamedValues:
		typedObj.SourceMeta = l.SourceMeta
		l.NamedValues = append(l.NamedValues, typedObj)
	default:
		return fmt.Errorf("unsupported type for collection: %T", prototypeObj)
	}

	l.ResourceCount++
	l.KindsCount[meta.Kind]++

	return nil
}

// UnmarshalMulti decodes a YAML file into the appropriate API object
func (l *Loader) UnmarshalMulti(data []byte) []error {
	var errors []error

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)

	for {
		var rawMessage json.RawMessage
		if err := decoder.Decode(&rawMessage); err != nil {
			break // End of input
		}

		unmarshalErr := l.Unmarshal(rawMessage)
		if unmarshalErr != nil {
			errors = append(errors, unmarshalErr)
			continue
		}
	}

	return errors
}

func (l *Loader) fromDir() map[string][]error {
	dirPath := l.SourceMeta.Source
	scheme := l.SourceMeta.Scheme

	if scheme != "file" {
		logger.Fatalf("fromDir called with non-file source: %s", l.SourceMeta.Scheme)
	}

	errs := make(map[string][]error)

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errs[dirPath] = append(
			errs[dirPath],
			fmt.Errorf("failed to get absolute path for file: %w", err),
		)

		return errs
	}

	walkErr := filepath.Walk(absDirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			errs[filePath] = append(errs[filePath], fmt.Errorf("error accessing file: %w", err))
			return nil
		}

		// Skip directories at the root level that start with a dot
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") &&
				filepath.Dir(filePath) == filepath.Clean(dirPath) {
				logger.Debugf("Skiping root directory to load Kubeit resources from: %s", filePath)
				return filepath.SkipDir
			}

			logger.Debugf("Found directory to walk to Kubeit resources from: %s", filePath)

			return nil
		}

		logger.Infof("Loading file: %s", filePath)

		data, err := os.ReadFile(filePath)
		if err != nil {
			errs[filePath] = append(errs[filePath], fmt.Errorf("failed to read file: %w", err))
			return nil
		}

		unmarhsalErr := l.UnmarshalMulti(data)
		if unmarhsalErr != nil {
			errs[filePath] = append(errs[filePath], unmarhsalErr...)
		}

		return nil
	})
	if walkErr != nil {
		errs[absDirPath] = append(
			errs[absDirPath],
			fmt.Errorf("failed to walk directory: %w", err),
		)
	}

	return errs
}

func (l *Loader) fromDockerImage() map[string][]error {
	imageRef := l.SourceMeta.Source

	errs := make(map[string][]error)

	dockerClientInstance, err := utils.NewRealDockerClient()
	if err != nil {
		errs[imageRef] = append(
			errs[imageRef],
			fmt.Errorf("failed to create Docker client: %w", err),
		)

		return errs
	}

	if exists, err := utils.CheckDockerImageExists(dockerClientInstance, imageRef); !exists ||
		err != nil {
		errs[imageRef] = append(errs[imageRef], fmt.Errorf("failed to find image: %w", err))
		return errs
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		errs[imageRef] = append(
			errs[imageRef],
			fmt.Errorf("failed to create Docker client: %w", err),
		)

		return errs
	}

	imageInspect, err := cli.ImageInspect(context.Background(), imageRef)
	if err != nil {
		errs[imageRef] = append(
			errs[imageRef],
			fmt.Errorf("failed to inspect Docker image: %w", err),
		)

		return errs
	}

	// check if kubeit.komail.io/resources is present in the labels
	labelKey := common.KubeitDomain + "/resources"

	base64Resource, ok := imageInspect.Config.Labels[labelKey]
	if !ok {
		errs[imageRef] = append(
			errs[imageRef],
			fmt.Errorf("no Kubeit resources found in image: %s", imageRef),
		)

		return errs
	}

	// decode the base64 encoded resources
	decodedResources, err := base64.StdEncoding.DecodeString(base64Resource)
	if err != nil {
		errs[imageRef] = append(
			errs[imageRef],
			fmt.Errorf("failed to decode base64 resources: %w", err),
		)

		return errs
	}

	logger.Debugf("Decoded resources:\n%s", decodedResources)

	unmarhsalErr := l.UnmarshalMulti(decodedResources)
	if unmarhsalErr != nil {
		errs[imageRef] = append(errs[imageRef], unmarhsalErr...)
	}

	return errs
}

func (l *Loader) FromSourceURI(sourceConfigURI string) map[string][]error {
	logger.Infof("Loading Kubeit resources from %s", sourceConfigURI)

	errs := make(map[string][]error)

	sourceScheme, source, err := utils.SourceConfigURIParser(sourceConfigURI)
	if err != nil {
		errs["SourceConfigURIParser"] = append(errs["SourceConfigURIParser"], err)
		return errs
	}

	l.SourceMeta = api.SourceMeta{
		Scheme:    sourceScheme,
		SourceURI: sourceConfigURI,
		Source:    source,
	}

	switch sourceScheme {
	case "file":
		errs = l.fromDir()
	case "docker":
		errs = l.fromDockerImage()
	default:
		errs["SourceConfigURIParser"] = append(errs["SourceConfigURIParser"], fmt.Errorf(
			"unsupported source config URI scheme: %s",
			sourceScheme,
		))

		return errs
	}

	// uniquenessErrs := kubeitFileResources.CheckResourceUniqueness()

	// // merge uniqueness errors with load errors
	// for file, errs := range uniquenessErrs {
	// 	loadErrs[file] = append(loadErrs[file], errs...)
	// }

	// uniquenessErr := kubeitFileResources.CheckResourceUniqueness()

	// if len(uniquenessErr) != 0 {
	// 	return nil, loaderMeta, nil, fmt.Errorf(
	// 		"%d resources are not unique",
	// 		len(uniquenessErr),
	// 	)
	// }

	// resourceCount := len(kubeitFileResources)
	// if resourceCount == 0 {
	// 	return nil, loaderMeta, nil, fmt.Errorf(
	// 		"no Kubeit resources found when traversing: %s",
	// 		sourceConfigURI,
	// 	)
	// }

	return errs
}

func (l *Loader) LogResources() {
	resourceCount := l.ResourceCount
	if resourceCount != 0 {
		for kind, count := range l.KindsCount {
			logger.Infof("%s: %d", kind, count)
		}

		logger.Infof("Found %d Kubeit resources", resourceCount)
	}
}

func (l *Loader) Marshal() (strings.Builder, []error) {
	var errs []error

	var resourcesYaml strings.Builder

	allResources := []api.Object{}
	for _, r := range l.HelmApplications {
		allResources = append(allResources, r)
	}

	for _, r := range l.NamedValues {
		allResources = append(allResources, r)
	}

	for _, resource := range allResources {
		jsonString, err := json.Marshal(resource)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to marshal resource: %w", err))
			continue
		}

		yamlString, err := k8syaml.JSONToYAML(jsonString)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to convert resource to yaml: %w", err))
			continue
		}

		resourcesYaml.WriteString("---\n")
		resourcesYaml.WriteString(string(yamlString))
	}

	return resourcesYaml, errs
}

// FindResource finds resources by their names
func FindResourcesByName[T api.Object](resources []T, names []string) []T {
	var matched []T

	for _, res := range resources {
		if utils.Contains(names, res.GetMetadata().Name) {
			matched = append(matched, res)
		}
	}

	return matched
}
