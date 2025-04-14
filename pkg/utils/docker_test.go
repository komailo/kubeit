package utils

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/image"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDockerClient is a mock implementation of DockerClientInterface
type MockDockerClient struct {
	mock.Mock
}

// ImageInspect mocks the ImageInspect method
func (m *MockDockerClient) ImageInspect(
	ctx context.Context,
	imageRef string,
) (image.InspectResponse, error) {
	args := m.Called(ctx, imageRef)

	if args.Error(1) != nil {
		return args.Get(0).(image.InspectResponse), fmt.Errorf(
			"mock error while inspecting image: %w",
			args.Error(1),
		)
	}

	return args.Get(0).(image.InspectResponse), nil
}

// ImagePull mocks the ImagePull method
func (m *MockDockerClient) ImagePull(
	ctx context.Context,
	imageRef string,
) error {
	args := m.Called(ctx, imageRef)

	if args.Error(1) != nil {
		return fmt.Errorf(
			"mock error while inspecting image: %w",
			args.Error(1),
		)
	}

	return nil
}

func TestCheckDockerImageExists_ImageExists(t *testing.T) {
	mockClient := new(MockDockerClient)
	imageRef := "existing-image"
	mockClient.On("ImageInspect", mock.Anything, imageRef).Return(image.InspectResponse{}, nil)

	exists, err := CheckDockerImageExists(mockClient, imageRef, false)

	require.NoError(t, err)
	assert.True(t, exists)
	mockClient.AssertExpectations(t)
}

func TestCheckDockerImageExists_ImageNotFound(t *testing.T) {
	mockClient := new(MockDockerClient)
	imageRef := "nonexistent-image"
	mockClient.On("ImageInspect", mock.Anything, imageRef).
		Return(image.InspectResponse{}, errdefs.ErrNotFound)

	exists, err := CheckDockerImageExists(mockClient, imageRef, false)

	require.NoError(t, err)
	assert.False(t, exists)
	mockClient.AssertExpectations(t)
}

func TestCheckDockerImageExists_ClientError(t *testing.T) {
	mockClient := new(MockDockerClient)
	imageRef := "error-image"
	mockClient.On("ImageInspect", mock.Anything, imageRef).
		Return(image.InspectResponse{}, errors.New("unexpected error"))

	exists, err := CheckDockerImageExists(mockClient, imageRef, false)

	require.Error(t, err)
	assert.False(t, exists)
	mockClient.AssertExpectations(t)
}

func TestCheckDockerImageExists_DockerClientNotInit(t *testing.T) {
	exists, err := CheckDockerImageExists(nil, "", false)

	require.Error(t, err)
	assert.False(t, exists)
}

func TestRealDockerClient_ImagePull(t *testing.T) {
	// Create a real Docker client
	client, err := NewRealDockerClient()
	require.NoError(t, err, "Expected no error when creating Docker client")

	// Pull a valid image
	err = client.ImagePull(t.Context(), "nginx:latest")
	require.NoError(t, err, "Expected no error for pulling a valid image")

	// Pull an invalid image
	err = client.ImagePull(t.Context(), "nonexistent-image:latest")
	require.Error(t, err, "Expected an error for pulling a nonexistent image")
	assert.Contains(
		t,
		err.Error(),
		"failed to pull image",
		"Expected error message to contain 'failed to pull image'",
	)
}
