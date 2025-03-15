package utils

import (
	"context"
	"errors"
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
	return args.Get(0).(image.InspectResponse), args.Error(1)
}

func TestCheckDockerImageExists_ImageExists(t *testing.T) {
	mockClient := new(MockDockerClient)
	imageRef := "existing-image"
	mockClient.On("ImageInspect", mock.Anything, imageRef).Return(image.InspectResponse{}, nil)

	exists, err := CheckDockerImageExists(mockClient, imageRef)

	require.Error(t, err)
	assert.True(t, exists)
	mockClient.AssertExpectations(t)
}

func TestCheckDockerImageExists_ImageNotFound(t *testing.T) {
	mockClient := new(MockDockerClient)
	imageRef := "nonexistent-image"
	mockClient.On("ImageInspect", mock.Anything, imageRef).
		Return(image.InspectResponse{}, errdefs.ErrNotFound)

	exists, err := CheckDockerImageExists(mockClient, imageRef)

	require.Error(t, err)
	assert.False(t, exists)
	mockClient.AssertExpectations(t)
}

func TestCheckDockerImageExists_ClientError(t *testing.T) {
	mockClient := new(MockDockerClient)
	imageRef := "error-image"
	mockClient.On("ImageInspect", mock.Anything, imageRef).
		Return(image.InspectResponse{}, errors.New("unexpected error"))

	exists, err := CheckDockerImageExists(mockClient, imageRef)

	require.Error(t, err)
	assert.False(t, exists)
	mockClient.AssertExpectations(t)
}

func TestCheckDockerImageExists_DockerClientNotInit(t *testing.T) {
	exists, err := CheckDockerImageExists(nil, "")

	require.Error(t, err)
	assert.False(t, exists)
}
