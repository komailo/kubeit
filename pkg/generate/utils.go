package generate

import (
	"fmt"
	"net/url"
	"strings"
)

// Given the sourceConfigUri, parse it and return a valid url.URL
func parseSourceConfigURI(sourceConfigUri string) (url.URL, error) {
	parsedURL, err := url.Parse(sourceConfigUri)
	if err != nil || parsedURL.Scheme == "" {
		return url.URL{}, fmt.Errorf("source config uri '%s' is invalid", sourceConfigUri)
	}

	if parsedURL.Path == "" {
		return url.URL{}, fmt.Errorf(
			"Incomplete source config uri '%s'. Expected format: <scheme>://<path>",
			sourceConfigUri,
		)
	}

	validSchemes := []string{"file"}
	if !contains(validSchemes, parsedURL.Scheme) {
		return url.URL{}, fmt.Errorf(
			"source scheme %s is invalid in '%s'. Valid sources: %s",
			parsedURL.Scheme, sourceConfigUri, strings.Join(validSchemes, ", "),
		)
	}

	return *parsedURL, nil
}

// contains checks if a string is in a slice of strings
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
