package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/distribution/reference"

	"github.com/komailo/kubeit/internal/logger"
)

func uriIsFile(uri string) (bool, string) {
	_, err := os.Stat(uri)
	if err == nil {
		logger.Debugf("File %s found so URI does look like a file", uri)

		absPath, err := filepath.Abs(uri)
		if err != nil {
			logger.Warnf("Failed to get absolute path for file %s", uri)
			return false, ""
		}

		return true, absPath
	}

	logger.Debugf("File %s not found so URI does not look like a file", uri)

	return false, ""
}

// UriIsDockerImgRef checks if the given URI is a Docker image reference
func uriIsDockerImgRef(uri string) (bool, string) {
	// Parse the Docker image reference
	ref, err := reference.ParseNormalizedNamed(uri)
	if err != nil {
		logger.Debugf("URI %s does not match Docker image pattern: %v", uri, err)
		return false, ""
	}

	// Check if the image exists
	dockerClientInstance, err := NewRealDockerClient()
	if err != nil {
		logger.Errorf("failed to create Docker client: %v", err)
		return false, ""
	}

	if exists, err := CheckDockerImageExists(dockerClientInstance, ref.String()); exists &&
		err == nil {
		logger.Debugf("URI %s matches Docker image pattern and exists", uri)
		return true, ref.String()
	}

	logger.Debugf(
		"URI %s matches Docker image pattern but does not exist - so guessing its not a Docker image ref",
		uri,
	)

	return false, ""
}

func SourceConfigURIParser(uri string) (string, string, error) {
	var scheme, path string

	var ok bool

	schemeRegex := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*):\/\/(.*)`)

	matches := schemeRegex.FindStringSubmatch(uri)
	if len(matches) == 3 {
		scheme = matches[1]
		path = matches[2]

		switch scheme {
		case "file":
			absPath, err := filepath.Abs(path)
			if err != nil {
				return scheme, path, fmt.Errorf("unable to get absolute path for file %s", path)
			}
			fmt.Printf("File %s found so URI does look like a file", absPath)

			return scheme, absPath, nil
		case "docker":
			ref, err := reference.ParseNormalizedNamed(path)
			if err != nil {
				return scheme, path, fmt.Errorf(
					"unable to parse and normalize Docker path: %s %w",
					path,
					err,
				)
			}

			return scheme, ref.String(), nil
		default:
			return scheme, path, fmt.Errorf("unknown scheme %s", scheme)
		}
	}

	if ok, path = uriIsFile(uri); ok {
		logger.Debugf("URI %s is a valid file, converting to file scheme", uri)

		scheme = "file"
	} else if ok, path = uriIsDockerImgRef(uri); ok {
		logger.Debugf("URI %s is a Docker Image Ref, converting to docker scheme", uri)

		scheme = "docker"
	} else {
		return "", "", fmt.Errorf("URI %s is not guessable", uri)
	}

	return scheme, path, nil
}
