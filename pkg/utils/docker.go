package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// DockerClientInterface defines the methods required for a Docker client
type DockerClientInterface interface {
	ImageInspect(ctx context.Context, imageRef string) (image.InspectResponse, error)
}

// RealDockerClient wraps the actual Docker client
type RealDockerClient struct {
	client *client.Client
}

// NewRealDockerClient initializes a new RealDockerClient
func NewRealDockerClient() (*RealDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &RealDockerClient{client: cli}, nil
}

// ImageInspect implements DockerClientInterface
func (r *RealDockerClient) ImageInspect(
	ctx context.Context,
	imageRef string,
) (image.InspectResponse, error) {
	return r.client.ImageInspect(ctx, imageRef)
}

// CheckDockerImageExists checks if a Docker image exists
func CheckDockerImageExists(dockerClient DockerClientInterface, imageRef string) (bool, error) {
	if dockerClient == nil {
		return false, errors.New("docker client is not initialized")
	}

	_, err := dockerClient.ImageInspect(context.Background(), imageRef)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect Docker image: %w", err)
	}
	return true, nil
}
