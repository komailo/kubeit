package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/komailo/kubeit/internal/logger"
	helmappv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_application/v1alpha1"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

func GenerateManifestFromHelm(HelmApplication helmappv1alpha1.HelmApplication, generateSetOptions *GenerateOptions) error {
	name := HelmApplication.Spec.Chart.Name
	releaseName := HelmApplication.Spec.Chart.ReleaseName
	namespace := HelmApplication.Spec.Chart.Namespace
	version := HelmApplication.Spec.Chart.Version
	repository := HelmApplication.Spec.Chart.Repository
	url := HelmApplication.Spec.Chart.URL
	kubeVersion := generateSetOptions.KubeVersion
	rawValues := HelmApplication.Spec.RawValues

	// Initialize Helm environment
	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), logger.Infof); err != nil {
		logger.Fatalf("Failed to initialize Helm action configuration: %v", err)
	}

	chartPath, error := pullHelmChart(
		settings,
		actionConfig,
		repository,
		name,
		url,
		version,
		generateSetOptions,
	)

	if error != nil {
		return error
	}
	chart, err := loader.Load(chartPath)
	if err != nil {
		logger.Fatalf("Failed to load Helm chart: %v", err)
	}

	// Prepare values and render templates
	values := chartutil.CoalesceTables(chart.Values, rawValues.(map[string]any))
	installClient := action.NewInstall(actionConfig)
	installClient.DryRun = true
	installClient.ReleaseName = releaseName
	installClient.Namespace = namespace
	installClient.ClientOnly = true

	if kubeVersion != "" {
		parsedKubeVersion, err := chartutil.ParseKubeVersion(kubeVersion)
		if err != nil {
			return fmt.Errorf("invalid kube version '%s': %s", kubeVersion, err)
		}
		installClient.KubeVersion = parsedKubeVersion
	}
	release, err := installClient.Run(chart, values)
	if err != nil {
		logger.Fatalf("Failed to render templates: %v", err)
	}

	// fail if manifest file is empty
	if release.Manifest == "" {
		return fmt.Errorf("No manifest file generated")
	}
	// TODO: Add common labels and annotations to the manifest
	//processedManifest, err := addCommonLabelsAndAnnotationsToK8sObject(release.Manifest)
	processedManifest := release.Manifest
	if err != nil {
		logger.Fatalf("Failed to process manifest: %v", err)
	}

	// Define the file path where you want to write the manifest
	manifestFilePath := filepath.Join(generateSetOptions.OutputDir, "manifest.yaml")

	// Write the manifest content to the file
	err = os.WriteFile(manifestFilePath, []byte(processedManifest), os.ModePerm)
	if err != nil {
		logger.Fatalf("Failed to write manifest to file: %v", err)
	}

	return nil
}

// pull helm charts and place them in the work directory
func pullHelmChart(
	settings *cli.EnvSettings,
	actionConfig *action.Configuration,
	repository, name, url, version string,
	generateSetOptions *GenerateOptions,
) (string, error) {
	// Pull OCI chart
	pullClient := action.NewPullWithOpts(action.WithConfig(actionConfig))
	registryClient, err := registry.NewClient()
	if err != nil {
		logger.Fatalf("Failed to create registry client: %v", err)
	}
	pullClient.SetRegistryClient(registryClient)
	pullClient.Settings = settings

	destinationDir := filepath.Join(generateSetOptions.WorkDir, "charts")

	// Delete the destiantion directory if it already exists
	if _, err := os.Stat(destinationDir); err == nil {
		if err := os.RemoveAll(destinationDir); err != nil {
			return "", fmt.Errorf("Failed to remove destination directory: %v", err)
		}
	}
	// Create the destination directory
	if err := os.MkdirAll(destinationDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("Failed to create destination directory: %v", err)
	}

	var chartRef string

	if name == "" && repository == "" && url == "" {
		return "", fmt.Errorf("either chart name and repository or url must be provided")
	} else if name == "" && repository != "" {
		return "", fmt.Errorf("chart name must be provided when using a repository")
	} else if name != "" && repository == "" {
		return "", fmt.Errorf("repository must be provided when using chart name")
	} else if url != "" {
		logger.Infof("Pulling chart from %s", url)
		chartRef = url
	} else {
		logger.Infof("Pulling chart %s from %s", name, repository)
		pullClient.RepoURL = repository
		chartRef = name
	}

	pullClient.Version = version
	pullClient.DestDir = destinationDir
	out, err := pullClient.Run(chartRef)
	if err != nil {
		return "", fmt.Errorf("Failed to pull chart: %v", err)
	}
	if out != "" {
		logger.Infof("helm pull run output %s", out)
	}

	// Check if the chart was downloaded to the destination directory
	files, err := os.ReadDir(destinationDir)
	if err != nil {
		return "", fmt.Errorf("Failed to read destination directory: %v", err)
	}

	var chartFileName string
	// we also want to fail if there are multiple chart files in the destination directory
	if len(files) > 1 {
		return "", fmt.Errorf("Multiple chart files found in destination directory")
	}
	chartFileName = files[0].Name()

	if chartFileName == "" {
		return "", fmt.Errorf("No chart file found in destination directory")
	}

	logger.Infof("Pulled Helm chart file: %s", chartFileName)

	return filepath.Join(destinationDir, chartFileName), nil
}

// func addCommonLabelsAndAnnotationsToK8sObject(manifest string) (string, error) {
// 	var processedDocuments []string

// 	// Create a YAML decoder
// 	decoder := yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(manifest)))

// 	for {
// 		var rawObj runtime.RawExtension
// 		if err := decoder.Decode(&rawObj); err != nil {
// 			if err.Error() == "EOF" {
// 				break
// 			}
// 			logger.Printf("Skipping invalid Kubernetes object: %v", err)
// 			continue
// 		}
// 		if rawObj.Raw == nil {
// 			continue
// 		}

// 		// Decode the raw object into a Kubernetes object
// 		obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(rawObj.Raw, nil, nil)
// 		if err != nil {
// 			logger.Printf("Skipping invalid Kubernetes object: %v\n%s", err, rawObj)
// 			processedDocuments = append(processedDocuments, string(rawObj.Raw))
// 			continue
// 		}

// 		// Add labels to valid Kubernetes objects
// 		accessor, err := meta.Accessor(obj)
// 		if err != nil {
// 			logger.Printf("Skipping non-Kubernetes object: %v", err)
// 			continue
// 		}
// 		commonLabels, commonAnnotations := generateCommonK8sLabelsAndAnnotationsToK8sObject()

// 		// Add the labels
// 		labelsMap := accessor.GetLabels()
// 		if labelsMap == nil {
// 			labelsMap = make(map[string]string)
// 		}
// 		for key, value := range commonLabels.GenerateLabels() {
// 			labelsMap[key] = value
// 		}
// 		accessor.SetLabels(labelsMap)

// 		// Add the annotations
// 		annotationsMap := accessor.GetAnnotations()
// 		if annotationsMap == nil {
// 			annotationsMap = make(map[string]string)
// 		}
// 		for key, value := range commonAnnotations.GenerateAnnotations() {
// 			annotationsMap[key] = value
// 		}
// 		accessor.SetAnnotations(annotationsMap)

// 		// Serialize the modified object back to YAML
// 		serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
// 		var buf bytes.Buffer
// 		if err := serializer.Encode(obj, &buf); err != nil {
// 			return "", fmt.Errorf("failed to encode object: %v", err)
// 		}

// 		processedDocuments = append(processedDocuments, buf.String())
// 	}

// 	// Join the processed documents back into a single manifest
// 	return strings.Join(processedDocuments, "---\n"), nil
// }

// func generateCommonK8sLabelsAndAnnotationsToK8sObject() (CommonK8sLabels, CommonK8sAnnotations) {
// 	var labels CommonK8sLabels

// 	annotations := CommonK8sAnnotations{
// 		AppName:     "kubeit",
// 		AppType:     "v0.1.0",
// 		GeneratedBy: "kubeit",
// 	}

// 	return labels, annotations
// }
