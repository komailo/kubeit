package utils

import (
	"fmt"

	"github.com/distribution/reference"
)

// contains checks if a string is in a slice of strings
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

// ParseDockerImage parses a Docker image string and returns the repository and tag
// values.
func ParseDockerImage(image string) (string, string, error) {
	// Parse the Docker image reference
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse Docker image reference: %w", err)
	}

	// Get the repository name
	repo := ref.Name()

	// Get the tag (if any)
	var tag string
	if tagged, ok := ref.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	return repo, tag, nil
}
