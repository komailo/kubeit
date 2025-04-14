package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	argopappv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/scorebet/reflow/common"
	"github.com/scorebet/reflow/internal/logger"
	"github.com/scorebet/reflow/pkg/api/loader"

	v1 "github.com/scorebet/reflow/pkg/api/v1"
)

var orgToClusterProject = map[string]string{
	"bet":      "scorebet",
	"data-eng": "data-engineering",
	"infra":    "scorebet",
	"media":    "scoremedia",
}

func ArgoAppSet(
	argoAppSet v1.ServiceApp,
	service v1.Service,
	generateSetOptions *Options,
) error {
	for _, env := range service.Spec.Environments {
		dirSelector := "non-prod"
		if service.IsProd(env) {
			dirSelector = "prod"
		}

		outputDirPath := filepath.Join(generateSetOptions.OutputDir,
			service.Spec.Org,
			"appset",
			dirSelector,
			service.Metadata.Name)
		appSetFilePath := filepath.Join(outputDirPath, env+".yaml")
		// kustomizationFilePath := filepath.Join(outputDirPath, "kustomization.yaml")
		appSet := createEnvArgoAppSet(env, service, argoAppSet)

		yamlBytes, err := loader.MarshalToYaml[argopappv1alpha1.ApplicationSet](appSet)
		if err != nil {
			return fmt.Errorf("failed to marshal ApplicationSet to YAML: %w", err)
		}

		err = os.MkdirAll(outputDirPath, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		err = os.WriteFile(appSetFilePath, yamlBytes, 0o644)
		if err != nil {
			return fmt.Errorf("failed to write ApplicationSet to file: %w", err)
		}
	}

	return nil
}

func createEnvArgoAppSet(
	env string,
	service v1.Service,
	argoAppSet v1.ServiceApp,
) argopappv1alpha1.ApplicationSet {
	syncFactor := int64(2)

	return argopappv1alpha1.ApplicationSet{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "ApplicationSet",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s", service.Metadata.Name, argoAppSet.Metadata.Name, env),
			Labels: map[string]string{
				common.ServiceDomain + "/env":              env,
				common.ServiceDomain + "/generated-by":     common.AppName,
				common.ServiceDomain + "/service-app-name": argoAppSet.Metadata.Name,
				common.ServiceDomain + "/service-name":     service.Metadata.Name,
				common.ServiceDomain + "/source-repo":      service.Spec.SourceRepo,
			},
		},
		Spec: argopappv1alpha1.ApplicationSetSpec{
			Generators: []argopappv1alpha1.ApplicationSetGenerator{
				getClusterGenerator(
					service.Spec.Org,
					env,
					argoAppSet,
					service,
				),
			},
			Template: argopappv1alpha1.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: argopappv1alpha1.ApplicationSetTemplateMeta{
					Name: fmt.Sprintf(
						"%s-%s-%s",
						service.Metadata.Name,
						argoAppSet.Metadata.Name,
						env,
					),
					Labels: map[string]string{
						common.ServiceDomain + "/env":              env,
						common.ServiceDomain + "/generated-by":     common.AppName,
						common.ServiceDomain + "/service-app-name": argoAppSet.Metadata.Name,
						common.ServiceDomain + "/service-name":     service.Metadata.Name,
						common.ServiceDomain + "/source-repo":      service.Spec.SourceRepo,
					},
					Annotations: map[string]string{
						"notifications.argoproj.io/subscribe.cd-visibility-trigger.cd-visibility-webhook": "",
						"dd_env":     env,
						"dd_service": service.Metadata.Name,
					},
					Finalizers: []string{
						"resources-finalizer.argocd.argoproj.io",
					},
				},
				Spec: argopappv1alpha1.ApplicationSpec{
					Project: service.Metadata.Name,
					SyncPolicy: &argopappv1alpha1.SyncPolicy{
						SyncOptions: []string{
							"CreateNamespace=false",
							"PruneLast=true",
							"PrunePropagationPolicy=foreground",
							"Validate=false",
						},
						Retry: &argopappv1alpha1.RetryStrategy{
							Limit: 5,
							Backoff: &argopappv1alpha1.Backoff{
								Duration:    "5s",
								MaxDuration: "3m",
								Factor:      &syncFactor,
							},
						},
					},
					Source: &argopappv1alpha1.ApplicationSource{
						RepoURL:        argoAppSet.Spec.ValuesRepository,
						TargetRevision: "main",
						Path:           "values",
						Plugin: &argopappv1alpha1.ApplicationSourcePlugin{
							Name: "helm-resolver-v3",
							Env: argopappv1alpha1.Env{
								&argopappv1alpha1.EnvEntry{
									Name:  "REPO",
									Value: "https://thescore.jfrog.io/artifactory/thescore-helm",
								},
								&argopappv1alpha1.EnvEntry{
									Name:  "REPO_NAME",
									Value: "thescore-helm",
								},
								&argopappv1alpha1.EnvEntry{
									Name: "RELEASE_NAME",
									Value: fmt.Sprintf(
										"%s-%s",
										service.Metadata.Name,
										argoAppSet.Spec.ClusterRole,
									),
								},
								&argopappv1alpha1.EnvEntry{
									Name:  "CHART_NAME",
									Value: argoAppSet.Spec.ChartName,
								},
								&argopappv1alpha1.EnvEntry{
									Name:  "CHART_VERSION",
									Value: argoAppSet.Spec.ChartName,
								},
								&argopappv1alpha1.EnvEntry{
									Name:  "VALUES_FILE",
									Value: strings.Join(argoAppSet.Spec.ValuesFiles, " "),
								},
							},
						},
					},
					Destination: argopappv1alpha1.ApplicationDestination{
						Server:    "{{ name }}",
						Namespace: service.Metadata.Name,
					},
				},
			},
		},
	}
}

func getClusterGenerator(
	org string, env string, argoAppSet v1.ServiceApp, service v1.Service,
) argopappv1alpha1.ApplicationSetGenerator {
	matchLabels := make(map[string]string)

	envType := "non-prod" // default to non-prod
	if service.IsProd(env) {
		envType = "prod"
	}

	matchLabels["env"] = envType
	matchLabels["role"] = argoAppSet.Spec.ClusterRole

	clusterProject, ok := orgToClusterProject[org]
	if !ok {
		logger.Fatalf("no cluster project found for org %s", org)
	}

	matchLabels["project"] = clusterProject

	return argopappv1alpha1.ApplicationSetGenerator{
		Clusters: &argopappv1alpha1.ClusterGenerator{
			Selector: k8smetav1.LabelSelector{
				MatchLabels: matchLabels,
			},
		},
	}
}
