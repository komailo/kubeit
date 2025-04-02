package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceConfigURIParser(t *testing.T) {
	// Create a temporary file to test file URIs
	tempFile, err := os.CreateTemp(t.TempDir(), "testfile")
	require.NoError(t, err, "Expected no error when creating a temporary file")
	defer os.Remove(tempFile.Name()) // Clean up the file after the test

	// Get the absolute path of the temporary file
	absPath, err := filepath.Abs(tempFile.Name())
	require.NoError(t, err, "Expected no error when getting absolute path")

	testCases := []struct {
		name           string
		uri            string
		expectedScheme string
		expectedPath   string
		expectedError  string
	}{
		{
			name:           "Valid file URI",
			uri:            "file://" + absPath,
			expectedScheme: "file",
			expectedPath:   absPath,
			expectedError:  "",
		},
		{
			name:           "Valid file without scheme",
			uri:            absPath,
			expectedScheme: "file",
			expectedPath:   absPath,
			expectedError:  "",
		},
		{
			name:           "Valid Docker image URI",
			uri:            "docker://nginx:latest",
			expectedScheme: "docker",
			expectedPath:   "docker.io/library/nginx:latest",
			expectedError:  "",
		},
		{
			name:           "Valid Docker image without scheme",
			uri:            "nginx:latest",
			expectedScheme: "docker",
			expectedPath:   "docker.io/library/nginx:latest",
			expectedError:  "",
		},
		{
			name:          "Unknown scheme",
			uri:           "unknown://example",
			expectedError: "unknown scheme unknown",
		},
		{
			name:          "Invalid URI",
			uri:           "invalid-uri",
			expectedError: "URI invalid-uri is not guessable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme, path, err := SourceConfigURIParser(tc.uri)

			if tc.expectedScheme != "" {
				assert.Equal(t, tc.expectedScheme, scheme, "Expected scheme mismatch")
			}

			if tc.expectedPath != "" {
				assert.Equal(t, tc.expectedPath, path, "Expected path mismatch")
			}

			// Check the error
			if tc.expectedError == "" {
				assert.NoError(t, err, "Expected no error but got one")
			} else {
				require.Error(t, err, "Expected an error but got none")
				assert.Contains(t, err.Error(), tc.expectedError, "Expected error message mismatch")
			}
		})
	}
}
