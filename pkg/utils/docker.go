package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	"github.com/scorebet/reflow/internal/logger"
)

// DockerClientInterface defines the methods required for a Docker client
type DockerClientInterface interface {
	ImageInspect(ctx context.Context, imageRef string) (image.InspectResponse, error)
	ImagePull(ctx context.Context, imageRef string) error
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
	inspectResponse, err := r.client.ImageInspect(ctx, imageRef)
	if err != nil {
		return inspectResponse, fmt.Errorf("failed to inspect image: %w", err)
	}

	return inspectResponse, nil
}

func (r *RealDockerClient) ImagePull(
	ctx context.Context,
	imageRef string,
) error {
	pullOptions := image.PullOptions{}

	readCloser, err := r.client.ImagePull(ctx, imageRef, pullOptions)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	defer readCloser.Close()

	// Process the JSON stream from the ReadCloser
	decoder := json.NewDecoder(readCloser)

	for {
		var progress map[string]interface{}
		if err := decoder.Decode(&progress); err == io.EOF {
			break // End of stream
		} else if err != nil {
			return fmt.Errorf("failed to decode progress: %w", err)
		}

		// Extract progress information
		status := progress["status"]
		id := progress["id"]
		progressDetail := progress["progress"]

		// Format the output to mimic `docker pull`
		switch {
		case id != nil && progressDetail != nil:
			// Display progress with ID and progress details
			fmt.Printf("\r\033[K%s: %s %s", id, status, progressDetail)
		case id != nil:
			// Display status with ID only
			fmt.Printf("\r\033[K%s: %s", id, status)
		default:
			// Display status only
			fmt.Printf("\r\033[K%s", status)
		}
	}

	// Empty out the line after the pull is complete
	// This is to ensure the last line is cleared
	// and doesn't leave any artifacts on the terminal
	fmt.Printf("\r\033[K")

	return nil
}

// CheckDockerImageExists checks if a Docker image exists
func CheckDockerImageExists(
	dockerClient DockerClientInterface,
	imageRef string,
	pullImage bool,
) (bool, error) {
	if dockerClient == nil {
		return false, errors.New("docker client is not initialized")
	}

	_, err := dockerClient.ImageInspect(context.Background(), imageRef)
	if err != nil {
		if client.IsErrNotFound(err) {
			logger.Debugf("Docker image %s not found when inspecting", imageRef)

			if pullImage {
				logger.Infof("Pulling Docker image %s", imageRef)

				if err := dockerClient.ImagePull(context.Background(), imageRef); err != nil {
					return false, fmt.Errorf("failed to pull Docker image: %w", err)
				}

				logger.Infof("Docker image %s pulled successfully", imageRef)

				return true, nil
			}

			return false, nil
		}

		return false, fmt.Errorf("failed to inspect Docker image: %w", err)
	}

	return true, nil
}
