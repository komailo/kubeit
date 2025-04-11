package loader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/scorebet/reflow/pkg/utils"
)

func TestLoader_FromSourceURI(t *testing.T) {
	testCases := []struct {
		name           string
		sourceURI      string
		sourceScheme   string
		expectedError  bool
		expectedErrors map[string][]error
		objectLengths  map[string]int
	}{
		{
			name:          "Valid resources",
			sourceURI:     "testdata/valid",
			expectedError: false,
			sourceScheme:  "file",
			objectLengths: map[string]int{
				"HelmApplication": 1,
				"NamedValues":     3,
			},
		},
		{
			name:          "Invalid resources",
			sourceURI:     "testdata/invalid",
			sourceScheme:  "file",
			expectedError: true,
			expectedErrors: map[string][]error{
				"testdata/invalid/invalid_kind.yaml": {
					errors.New("unknown kind: UnknownKind"),
				},
				"testdata/invalid/invalid_version.yaml": {
					errors.New(
						"unknown version reflow.scorebet.github.io/v2alpha1 for kind HelmApplication",
					),
				},
				"testdata/invalid/invalid_helm_application.yaml": {
					errors.New(
						"failed to unmarshal HelmApplication: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal bool into Go struct field HelmApplicationSpec.spec.chart of type v1.ChartSpec",
					),
				},
				"testdata/invalid/invalid_meta_type.yaml": {
					errors.New(
						"failed to decode type metadata: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal array into Go struct field TypeMeta.kind of type string",
					),
				},
			},
		},
		{
			name:          "Edge cases",
			sourceURI:     "testdata/edge_cases",
			sourceScheme:  "file",
			expectedError: true,
		},
		{
			name:          "Unknown scheme in sourceURI",
			sourceURI:     "notknown://does/not/matter",
			expectedError: true,
			expectedErrors: map[string][]error{
				"SourceConfigURIParser": {
					errors.New("unknown scheme notknown"),
				},
			},
		},
		{
			name:          "Invalid file path with a null character",
			sourceURI:     fmt.Sprintf("file://foo%s", []byte{0x00}),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loader := NewLoader()

			// Load resources from the directory
			errors := loader.FromSourceURI(tc.sourceURI)

			// Check the lengths of the loaded objects
			if !tc.expectedError {
				assert.Len(
					t,
					loader.HelmApplications,
					tc.objectLengths["HelmApplication"],
					"Expected HelmApplication length mismatch",
				)
				assert.Len(
					t,
					loader.NamedValues,
					tc.objectLengths["NamedValues"],
					"Expected NamedValues length mismatch",
				)
			}

			var kindsCount int

			// check KindsCound
			for kind, expectedCount := range tc.objectLengths {
				// check loader.KindsCount[kind] exists
				if _, ok := loader.KindsCount[kind]; !ok {
					t.Fatalf("Expected KindsCount for %s to exist, but it does not", kind)
				}

				assert.Equal(
					t,
					expectedCount,
					loader.KindsCount[kind],
					"Expected KindsCount mismatch for %s",
					kind,
				)

				kindsCount += expectedCount
			}

			// check loader.ResourceCount
			if tc.objectLengths != nil {
				assert.Equal(t, kindsCount, loader.ResourceCount, "Expected ResourceCount mismatch")
			}

			// check loader.SourceMeta
			if tc.sourceScheme != "" {
				assert.Equal(
					t,
					tc.sourceScheme,
					loader.SourceMeta.Scheme,
					"Expected SourceMeta scheme mismatch",
				)
			}

			if tc.sourceScheme == "file" {
				expectFileSource, err := filepath.Abs(tc.sourceURI)
				require.NoError(t, err, "Expected no error when getting absolute path")
				assert.Equal(
					t,
					expectFileSource,
					loader.SourceMeta.Source,
					"Expected SourceMeta source uri mismatch",
				)
			}

			// check errors
			if tc.expectedError {
				assert.NotEmpty(t, errors, "Expected errors but got none")
			} else {
				assert.Empty(t, errors, "Expected no errors but got some")
			}

			if tc.expectedErrors != nil {
				nonFileErrors := []string{
					"SourceConfigURIParser",
				}
				// Convert keys in expectedErrors to absolute paths
				updatedExpectedErrors := make(
					map[string][]string,
				) // Change to map of strings for comparison

				for k, v := range tc.expectedErrors {
					if utils.Contains(nonFileErrors, k) {
						// Convert errors to strings
						errorMessages := make([]string, len(v))
						for i, err := range v {
							errorMessages[i] = err.Error()
						}

						updatedExpectedErrors[k] = errorMessages

						continue
					}

					absPath, err := filepath.Abs(k)
					require.NoError(t, err, "Expected no error when getting absolute path")

					// Convert errors to strings
					errorMessages := make([]string, len(v))
					for i, err := range v {
						errorMessages[i] = err.Error()
					}

					updatedExpectedErrors[absPath] = errorMessages
				}

				// Convert actual errors to strings for comparison
				actualErrors := make(map[string][]string)

				for k, v := range errors {
					errorMessages := make([]string, len(v))
					for i, err := range v {
						errorMessages[i] = err.Error()
					}

					actualErrors[k] = errorMessages
				}

				assert.Equal(t, updatedExpectedErrors, actualErrors, "Expected errors mismatch")
			}
		})
	}
}

func TestLoader_Marshal(t *testing.T) {
	testCases := []struct {
		name             string
		sourceURI        string
		validMarshalFile string
	}{
		{
			name:             "Valid resources",
			sourceURI:        "testdata/valid",
			validMarshalFile: "testdata/valid_marshalled.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loader := NewLoader()

			errors := loader.FromSourceURI(tc.sourceURI)
			assert.Empty(t, errors, "Expected no errors but got some")

			marshalledDataStr, marshalErrs := loader.Marshal()
			assert.Empty(t, marshalErrs, "Expected no errors but got some")

			var marshalledData []json.RawMessage

			decoder := yaml.NewYAMLOrJSONDecoder(
				bytes.NewReader([]byte(marshalledDataStr.String())),
				4096,
			)

			for {
				var rawMessage json.RawMessage
				if err := decoder.Decode(&rawMessage); err != nil {
					break // End of input
				}

				marshalledData = append(marshalledData, rawMessage)
			}

			expsectedDataStr, err := os.ReadFile(tc.validMarshalFile)
			require.NoError(t, err, "Expected no error when reading expected file")

			var expectedData []json.RawMessage

			decoder = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(expsectedDataStr), 4096)

			for {
				var rawMessage json.RawMessage
				if err := decoder.Decode(&rawMessage); err != nil {
					break // End of input
				}

				expectedData = append(expectedData, rawMessage)
			}

			// Compare the normalized YAML strings
			assert.Equal(t, expectedData, marshalledData, "Expected marshaled data mismatch")
		})
	}
}

func TestLoader_FindResourcesByName(t *testing.T) {
	loader := NewLoader()

	errors := loader.FromSourceURI("testdata/valid")
	assert.Empty(t, errors, "Expected no errors but got some")

	resources := FindResourcesByName(loader.NamedValues, []string{"staging", "canary"})
	assert.Len(t, resources, 2, "Expected 1 resource but got %d", len(resources))
}
