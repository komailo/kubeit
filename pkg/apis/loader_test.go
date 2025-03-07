package apis

import (
	"os"
	"path/filepath"
	"testing"

	appv1alpha1 "github.com/komailo/kubeit/pkg/apis/application/v1alpha1"
	helmappv1alpha1 "github.com/komailo/kubeit/pkg/apis/helm_application/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLoadKubeitResource(t *testing.T) {
	tests := []struct {
		name             string
		yamlData         string
		expectedType     interface{}
		expectedData     interface{}
		expectedMetadata k8smetav1.TypeMeta
		expectError      bool
		errorContains    string
		disabled         bool
	}{
		{
			name: "Valid kubeit resource",
			yamlData: `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
metadata:
  name: my-app
`,
			expectedType: &appv1alpha1.Application{},
			expectError:  false,
			expectedData: &appv1alpha1.Application{
				Metadata: appv1alpha1.Metadata{
					Name: "my-app",
				},
				TypeMeta: k8smetav1.TypeMeta{
					APIVersion: "kubeit.komailo.github.io/v1alpha1",
					Kind:       "Application",
				},
			},
		},
		{
			name: "Invalid resource data",
			yamlData: `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
metadata:
  foo: invalid
`,
			errorContains: "validation error: Key: 'Application.Metadata.Name",
			expectError:   true,
		},
		{
			name: "Invalid yaml",
			yamlData: `
:lmsn&&&&&&
`,
			expectError:   true,
			errorContains: "failed to unmarshal file",
		},
		{
			name: "Unsupported kind",
			yamlData: `
apiVersion: does not really matter
kind: UnknownKind
`,
			expectError:   true,
			errorContains: "unknown resource kind",
		},
		{
			name: "Unsupported apiVersion",
			yamlData: `
apiVersion: invalid.group/v1
kind: Application
`,
			expectError:   true,
			errorContains: "unsupported apiVersion",
		},
		{
			name: "empty api metadata value",
			yamlData: `
apiVersion: ""
kind: ""
`,
			expectError:   true,
			errorContains: "missing apiVersion or kind in resource",
		},
	}

	for _, tc := range tests {
		if tc.disabled {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			resource, err := loadKubeitResource([]byte(tc.yamlData))

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tc.expectedType, resource)
				assert.Equal(t, tc.expectedData, resource)
			}
		})
	}
}

func TestLoadKubeitResourcesFromDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "kubeit-test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name                   string
		fileName               string
		content                string
		expectsError           bool
		disabled               bool
		expectedResourcesCount int
		expectedErrorsCount    int
	}{
		{
			name:     "Valid YAML with single resource",
			fileName: "valid1.yaml",
			content: `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  project: default
  source:
    repoURL: https://github.com/my/repo.git
    path: charts/my-app
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-namespace
`,
			expectsError: false,
		},
		{
			name:     "Valid YAML with multiple resources",
			fileName: "multiple-resources.yaml",
			content: `
---
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  project: default
  source:
    repoURL: https://github.com/my/repo.git
    path: charts/my-app
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-namespace
---
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: HelmApplication
metadata:
  name: my-app
spec:
  project: default
  source:
    repoURL: https://github.com/my/repo.git
    path: charts/my-app
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-namespace
`,
			expectsError: false,
		},
		{
			name:     "Invalid YAML",
			fileName: "invalid.yaml",
			content: `
		dhdfhdf
		`,
			expectsError: true,
		},
		{
			name:     "Invalid resource in YAML",
			fileName: "wrong.yaml",
			content: `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: af
`,
			expectsError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tc.fileName)
			require.NoError(t, os.WriteFile(filePath, []byte(tc.content), 0644))
		})
	}

	resources, errors := loadKubeitResourcesFromDir(tempDir)

	var validCount, errorCount int
	for _, tc := range testCases {
		if tc.expectsError {
			errorCount++
		} else {
			validCount++
		}
	}

	assert.Len(t, resources, validCount)
	assert.Len(t, errors, errorCount)
}

func TestLoadKubeitResourcesFromDirNew(t *testing.T) {
	// Define test data as a map (mimicking files)
	testFiles := map[string]string{
		"valid1.yaml": `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  project: default
  source:
    repoURL: https://github.com/my/repo.git
    path: charts/my-app
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-namespace
`,
		"subdir/multiple-resources.yaml": `
---
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  project: default
  source:
    repoURL: https://github.com/my/repo.git
    path: charts/my-app
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-namespace
---
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: HelmApplication
metadata:
  name: my-app
spec:
  chart: my-chart
  version: 1.2.3
  releaseName: my-release
  namespace: my-namespace
`,
		"invalid.yaml": `
invalidyaml
		`,
		"wrong.yaml": `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: UnknownKind
`,
	}

	// Expected number of resources and errors
	expectedResources := []KubeitFileResource{
		{
			FileName: "valid1.yaml",
			Resource: &appv1alpha1.Application{
				TypeMeta: k8smetav1.TypeMeta{
					APIVersion: "kubeit.komailo.github.io/v1alpha1",
					Kind:       "Application",
				},
				Metadata: appv1alpha1.Metadata{Name: "my-app"},
			},
		},
		{
			FileName: "subdir/multiple-resources.yaml",
			Resource: &appv1alpha1.Application{
				TypeMeta: k8smetav1.TypeMeta{
					APIVersion: "kubeit.komailo.github.io/v1alpha1",
					Kind:       "Application",
				},
				Metadata: appv1alpha1.Metadata{Name: "my-app"},
			},
		},
		{
			FileName: "subdir/multiple-resources.yaml",
			Resource: &helmappv1alpha1.HelmApplication{
				TypeMeta: k8smetav1.TypeMeta{
					APIVersion: "kubeit.komailo.github.io/v1alpha1",
					Kind:       "HelmApplication",
				},
				Metadata: helmappv1alpha1.Metadata{Name: "my-app"},
				Spec: helmappv1alpha1.Spec{
					Chart: helmappv1alpha1.Chart{
						Name:        "my-chart",
						Version:     "1.2.3",
						ReleaseName: "my-release",
						Namespace:   "my-namespace",
					},
				},
			},
		},
	}

	expectedErrors := 2 // From "invalid.yaml" and "wrong.yaml"

	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("/tmp", "kubeit-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create files based on testFiles map
	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)

		// Ensure directory exists
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))

		// Write the test YAML content to the file
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	// Run function under test
	resources, errors := loadKubeitResourcesFromDir(tempDir)

	// Validate expected resources count
	assert.Len(t, resources, len(expectedResources))

	// Validate expected errors count
	assert.Len(t, errors, expectedErrors)

	// Compare actual resources with expected
	for i, expected := range expectedResources {
		assert.Equal(t, expected.FileName, resources[i].FileName)
		assert.IsType(t, expected.Resource, resources[i].Resource)
	}
}
