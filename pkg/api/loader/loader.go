package loader

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/docker/docker/client"

	"github.com/go-playground/validator/v10"

	"github.com/scorebet/reflow/common"
	"github.com/scorebet/reflow/internal/logger"
	"github.com/scorebet/reflow/pkg/api"
	v1 "github.com/scorebet/reflow/pkg/api/v1"
	"github.com/scorebet/reflow/pkg/utils"
)

// Loader handles the decoding of YAML/JSON documents into API objects
type Loader struct {
	// registry maps Kind and Version to the corresponding type
	registry         map[string]map[string]registeredType
	SourceMeta       api.SourceMeta
	HelmApplications []*v1.HelmApplication
	NamedValues      []*v1.NamedValues
	Services         []*v1.Service
	ServiceApps      []*v1.ServiceApp
	KindsCount       map[string]int
	ResourceCount    int
}

type registeredType struct {
	New      func() api.Object
	AppendFn func(obj api.Object)
	GetAll   func() []api.Object
	GetKind  func() string
}

// NewDecoder creates a new decoder with registered types
func NewLoader() *Loader {
	l := &Loader{
		registry:   make(map[string]map[string]registeredType),
		KindsCount: make(map[string]int),
	}

	register(
		l,
		"HelmApplication",
		common.APIVersionV1Alpha1,
		func() *v1.HelmApplication { return &v1.HelmApplication{} },
		&l.HelmApplications,
	)
	register(
		l,
		"NamedValues",
		common.APIVersionV1Alpha1,
		func() *v1.NamedValues { return &v1.NamedValues{} },
		&l.NamedValues,
	)

	register(
		l,
		"Service",
		common.APIVersionV1Alpha1,
		func() *v1.Service { return &v1.Service{} },
		&l.Services,
	)

	register(
		l,
		"ServiceApplication",
		common.APIVersionV1Alpha1,
		func() *v1.ServiceApp { return &v1.ServiceApp{} },
		&l.ServiceApps,
	)

	return l
}

func register[T api.Object](l *Loader, kind, version string, constructor func() T, slicePtr *[]T) {
	if _, ok := l.registry[kind]; !ok {
		l.registry[kind] = make(map[string]registeredType)
	}

	l.registry[kind][version] = registeredType{
		New: func() api.Object {
			return constructor()
		},
		AppendFn: func(obj api.Object) {
			*slicePtr = append(*slicePtr, obj.(T))
		},
		GetAll: func() []api.Object {
			result := make([]api.Object, len(*slicePtr))
			for i, v := range *slicePtr {
				result[i] = v
			}
			return result
		},
		GetKind: func() string {
			return kind
		},
	}
}

// TypeMeta is used to determine the type of resource
type TypeMeta struct {
	Kind       string `json:"kind"       yaml:"kind"`
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
}

func (l *Loader) unmarshal(data []byte, sourceMeta api.SourceMeta) error {
	var meta TypeMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("failed to decode type metadata: %w", err)
	}

	versionMap, ok := l.registry[meta.Kind]
	if !ok {
		return fmt.Errorf("unknown kind: %s", meta.Kind)
	}

	rt, ok := versionMap[meta.APIVersion]
	if !ok {
		return fmt.Errorf("unknown version %s for kind %s", meta.APIVersion, meta.Kind)
	}

	obj := rt.New()
	if err := yaml.UnmarshalStrict(data, obj); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", meta.Kind, err)
	}

	obj.SetSourceMeta(sourceMeta)

	rt.AppendFn(obj)

	l.ResourceCount++
	l.KindsCount[meta.Kind]++

	return nil
}

// UnmarshalMulti decodes a YAML file into the appropriate API object
func (l *Loader) unmarshalMulti(data []byte, sourceMeta api.SourceMeta) []error {
	var errors []error

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)

	for {
		var rawMessage json.RawMessage
		if err := decoder.Decode(&rawMessage); err != nil {
			break // End of input
		}

		unmarshalErr := l.unmarshal(rawMessage, sourceMeta)
		if unmarshalErr != nil {
			errors = append(errors, unmarshalErr)
			continue
		}
	}

	return errors
}

func (l *Loader) fromDir() map[string][]error {
	dirPath := l.SourceMeta.Source

	errs := make(map[string][]error)

	walkErr := filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			errs[filePath] = append(errs[filePath], fmt.Errorf("error accessing file: %w", err))
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") &&
				filepath.Dir(filePath) == filepath.Clean(dirPath) {
				logger.Debugf(
					"Skiping root directory to load %s resources from: %s",
					common.AppName,
					filePath,
				)

				return filepath.SkipDir
			}

			logger.Debugf(
				"Found directory to walk to %s resources from: %s",
				common.AppName,
				filePath,
			)

			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			logger.Debugf(
				"Skiping hidden file: %s",
				filePath,
			)

			return nil
		}

		if !strings.HasSuffix(info.Name(), ".yaml") &&
			!strings.HasSuffix(info.Name(), ".yml") &&
			!strings.HasSuffix(info.Name(), ".json") {
			logger.Debugf(
				"Skiping file as not a valid extension: %s",
				filePath,
			)

			return nil
		}

		logger.Infof("Loading file: %s", filePath)

		data, err := os.ReadFile(filePath)
		if err != nil {
			errs[filePath] = append(errs[filePath], fmt.Errorf("failed to read file: %w", err))
			return nil
		}

		sourceMeta := api.SourceMeta{
			Scheme:    l.SourceMeta.Scheme,
			SourceURI: l.SourceMeta.SourceURI,
			Source:    filePath,
		}

		unmarhsalErr := l.unmarshalMulti(data, sourceMeta)
		if unmarhsalErr != nil {
			errs[filePath] = append(errs[filePath], unmarhsalErr...)
		}

		return nil
	})
	if walkErr != nil {
		errs[dirPath] = append(
			errs[dirPath],
			fmt.Errorf("failed to walk directory: %w", walkErr),
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

	if exists, err := utils.CheckDockerImageExists(dockerClientInstance, imageRef, true); !exists ||
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

	labelKey := common.ServiceDomain + "/resources"

	base64Resource, ok := imageInspect.Config.Labels[labelKey]
	if !ok {
		errs[imageRef] = append(
			errs[imageRef],
			fmt.Errorf("no %s resources found in image: %s", common.AppName, imageRef),
		)

		return errs
	}

	decodedResources, err := base64.StdEncoding.DecodeString(base64Resource)
	if err != nil {
		errs[imageRef] = append(
			errs[imageRef],
			fmt.Errorf("failed to decode base64 resources: %w", err),
		)

		return errs
	}

	logger.Debugf("Decoded resources:\n%s", decodedResources)

	sourceMeta := api.SourceMeta{
		Scheme:    l.SourceMeta.Scheme,
		SourceURI: l.SourceMeta.SourceURI,
		Source:    l.SourceMeta.SourceURI,
	}

	unmarhsalErr := l.unmarshalMulti(decodedResources, sourceMeta)
	if unmarhsalErr != nil {
		errs[imageRef] = append(errs[imageRef], unmarhsalErr...)
	}

	return errs
}

func (l *Loader) checkResourceUniqueness() map[string][]error {
	errors := make(map[string][]error)

	seen := make(map[string]string)

	for _, versions := range l.registry {
		for _, kind := range versions {
			for _, resource := range kind.GetAll() {
				name := resource.GetObjectMeta().Name
				kindName := kind.GetKind()
				uniqueKey := fmt.Sprintf("%T-%s", kindName, name)

				if _, ok := seen[uniqueKey]; ok {
					errors[resource.GetSourceMeta().Source] = append(
						errors[resource.GetSourceMeta().Source],
						fmt.Errorf(
							"Resource of kind %s with name %s is not unique. Already seen in: %s",
							kindName,
							name,
							seen[uniqueKey],
						),
					)
				}

				seen[uniqueKey] = resource.GetSourceMeta().Source
			}
		}
	}

	return errors
}

func (l *Loader) LogResources() {
	resourceCount := l.ResourceCount
	if resourceCount != 0 {
		for kind, count := range l.KindsCount {
			logger.Infof("%s: %d", kind, count)
		}

		logger.Infof("Found %d %s resources", resourceCount, common.AppName)
	}
}

func (l *Loader) Marshal() (strings.Builder, []error) {
	var errs []error

	var resourcesYaml strings.Builder

	for _, versions := range l.registry {
		for _, rt := range versions {
			for _, resource := range rt.GetAll() {
				yamlStr, err := marshalResourceToYAML(resource)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				resourcesYaml.WriteString("---\n")
				resourcesYaml.WriteString(yamlStr)
			}
		}
	}

	return resourcesYaml, errs
}

func (l *Loader) FromSourceURI(sourceConfigURI string) map[string][]error {
	logger.Infof("Loading %s resources from %s", common.AppName, sourceConfigURI)

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
	default: // this should never happen as SourceConfigURIParser would error out
	}

	if len(errs) == 0 {
		validateErrs := l.Validate()
		if len(validateErrs) != 0 {
			return validateErrs
		}
	}

	return errs
}

func (l *Loader) Validate() map[string][]error {
	validationErrors := make(map[string][]error) // Initialize the map

	// Uniqueness check
	uniquenessErrors := l.checkResourceUniqueness()
	if len(uniquenessErrors) != 0 {
		for key, errs := range uniquenessErrors {
			validationErrors[key] = append(validationErrors[key], errs...)
		}
	}

	// Validation check for each resource
	for _, versions := range l.registry {
		for _, kind := range versions {
			for _, resource := range kind.GetAll() {
				if serviceApp, ok := resource.(*v1.ServiceApp); ok {
					jsonBytes, err := json.MarshalIndent(serviceApp, "", "  ")
					if err != nil {
						logger.Debugf("Failed to marshal ServiceApp: %v\n", err)
					} else {
						logger.Debugf("ServiceApp:\n%s\n", string(jsonBytes))
					}
				}

				kindName := kind.GetKind()

				validate := validator.New(validator.WithRequiredStructEnabled())

				// struct validation
				err := validate.Struct(resource)
				if err != nil {
					// this check is only needed when your code could produce
					// an invalid value for validation such as interface with nil
					// value most including myself do not usually have code like this.
					var invalidValidationError *validator.InvalidValidationError
					if errors.As(err, &invalidValidationError) {
						validationErrors[resource.GetSourceMeta().Source] = append(
							validationErrors[resource.GetSourceMeta().Source],
							fmt.Errorf(
								"invalid validation error %w", err,
							),
						)
					}

					var validateErrs validator.ValidationErrors
					if errors.As(err, &validateErrs) {
						for _, e := range validateErrs {
							validationErrors[resource.GetSourceMeta().Source] = append(
								validationErrors[resource.GetSourceMeta().Source],
								fmt.Errorf(
									"validation failed on kind %s for field '%s': %s",
									kindName,
									e.Namespace(),
									e.Tag(),
								),
							)
						}
					}
				}

				// API resource implemented validation
				if err := resource.Validate(); err != nil {
					name := resource.GetObjectMeta().Name
					validationErrors[resource.GetSourceMeta().Source] = append(
						validationErrors[resource.GetSourceMeta().Source],
						fmt.Errorf(
							"Resource of kind %s with name %s has errors: %w",
							kindName,
							name,
							err,
						),
					)
				}
			}
		}
	}

	if len(validationErrors) != 0 {
		return validationErrors
	}

	return nil
}

func MarshalToYaml[T any](resource any) ([]byte, error) {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("error marshalling while marshalling resource to yaml: %w", err)
	}

	yamlBytes, err := k8syaml.JSONToYAML(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf(
			"error doing json to yaml while marshalling resource to yaml: %w",
			err,
		)
	}

	return yamlBytes, nil
}

func marshalResourceToYAML(resource api.Object) (string, error) {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return "", fmt.Errorf("error marshalling while marshalling resource to yaml: %w", err)
	}

	yamlBytes, err := k8syaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", fmt.Errorf(
			"error doing json to yaml while marshalling resource to yaml: %w",
			err,
		)
	}

	return string(yamlBytes), nil
}

func FindResourcesByName[T api.Object](resources []T, names []string) []T {
	var matched []T

	for _, res := range resources {
		if utils.Contains(names, res.GetObjectMeta().Name) {
			matched = append(matched, res)
		}
	}

	return matched
}
