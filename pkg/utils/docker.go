package utils

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// CheckDockerImageExists checks if a Docker image exists
func CheckDockerImageExists(imageRef string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, fmt.Errorf("failed to create Docker client: %w", err)
	}

	_, err = cli.ImageInspect(context.Background(), imageRef)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect Docker image: %w", err)
	}

	return true, nil
}
