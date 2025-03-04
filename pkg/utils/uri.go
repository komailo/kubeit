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
	if exists, err := CheckDockerImageExists(ref.String()); exists && err == nil {
		logger.Debugf("URI %s matches Docker image pattern and exists", uri)
		return true, ref.String()
	}

	logger.Debugf("URI %s matches Docker image pattern but does not exist - so guessing its not a Docker image ref", uri)
	return false, ""
}

func SourceConfigUriParser(uri string) (string, string, error) {
	var scheme, path string
	var ok bool
	// check to see if a scheme:// is present in the uri
	schemeRegex := regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*):\/\/(.*)`)
	matches := schemeRegex.FindStringSubmatch(uri)
	if len(matches) > 1 {
		scheme = matches[1]
		path = matches[2]

		knownSchemes := map[string]bool{
			"docker": true,
			"file":   true,
		}
		if !knownSchemes[scheme] {
			return "", "", fmt.Errorf("unknown scheme %s", scheme)
		}

		if scheme == "file" {
			absPath, err := filepath.Abs(path)
			if err != nil {
				logger.Warnf("unable to get absolute path for file %s", path)
				return scheme, "", nil
			}
			return scheme, absPath, nil
		} else if scheme == "docker" {
			ref, err := reference.ParseNormalizedNamed(path)
			if err != nil {
				logger.Debugf("unable to parse and normalize Docker path: %s %v", path, err)
				return scheme, "", nil
			}
			return scheme, ref.String(), nil
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
