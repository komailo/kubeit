package apis

import (
	"testing"

	appv1alpha1 "github.com/komailo/kubeit/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestLoadKubeitResource(t *testing.T) {
	tests := []struct {
		name          string
		yamlData      string
		expectedType  interface{}
		expectError   bool
		errorContains string
		disabled      bool
	}{
		{
			name: "Valid Kind of Application",
			yamlData: `
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
			expectedType: &appv1alpha1.Application{},
			expectError:  false,
		},
		{
			name: "Invalid kind data with extra properties",
			yamlData: `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
metadata:
  foo: invalid
`,
			errorContains: "failed to parse resource strictly",
			expectError:   true,
		},
		{
			name: "Invalid kind data with missing required properties",
			yamlData: `
apiVersion: kubeit.komailo.github.io/v1alpha1
kind: Application
`,
			errorContains: "validation error",
			expectError:   true,
		},
		{
			name: "Completely invalid yaml",
			yamlData: `
:lmsn&&&&&&
`,
			expectError:   true,
			errorContains: "failed to parse YAML: yaml: unmarshal errors",
		},
		{
			name: "Invalid kind",
			yamlData: `
apiVersion: does not really matter
kind: UnknownKind
`,
			expectError:   true,
			errorContains: "unknown resource kind",
		},
		{
			name: "Invalid apiVersion",
			yamlData: `
apiVersion: invalid.group/v1
kind: Application
`,
			expectError:   true,
			errorContains: "unsupported apiVersion",
		},
		{
			name: "Invalid API metadata",
			yamlData: `
apiVersion: [1]
`,
			expectError:   true,
			errorContains: "failed to extract api metadata",
		},
		{
			name: "empty api metadata value",
			yamlData: `
apiVersion: ""
kind: ""
`,
			expectError:   true,
			errorContains: "validation error: Key: 'APIMetadata.APIVersion' Error:Field validation for 'APIVersion' failed on the 'required' tag\nKey: 'APIMetadata.Kind' Error:Field validation for 'Kind' failed on the 'required' tag",
		},
		{
			name: "Unmarshal failure on valid API version and kind",
			yamlData: `
		apiVersion: kubeit.komailo.github.io/v1alpha1
		kind: Application
		`,
			expectError:   true,
			errorContains: "failed to parse YAML",
		},
	}

	for _, tc := range tests {
		if tc.disabled {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			result, err := LoadKubeitResource([]byte(tc.yamlData))

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tc.expectedType, result)
			}
		})
	}
}
