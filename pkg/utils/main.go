package utils

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/komailo/kubeit/internal/logger"
)

// Given the sourceConfigUri, parse it and return a valid url.URL
func ParseSourceConfigURI(sourceConfigUri string) (url.URL, error) {
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
	if !Contains(validSchemes, parsedURL.Scheme) {
		return url.URL{}, fmt.Errorf(
			"source scheme %s is invalid in '%s'. Valid sources: %s",
			parsedURL.Scheme, sourceConfigUri, strings.Join(validSchemes, ", "),
		)
	}

	return *parsedURL, nil
}

// contains checks if a string is in a slice of strings
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func UriIsFile(uri string) (bool, string) {
	_, err := os.Stat(uri)
	if err == nil {
		logger.Debugf("File %s found so URI does look like a file", uri)
		absPath, err := filepath.Abs(uri)
		if err != nil {
			logger.Warnf("Failed to get absolute path for file %s", uri)
			return false, ""
		}
		return true, fmt.Sprintf("file://%s", absPath)
	}
	logger.Debugf("File %s not found so URI does not look like a file", uri)
	return false, ""
}
