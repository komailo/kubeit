package utils

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseDockerImage(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectRepo string
		expectTag  string
		wantError  bool
	}{
		{"Valid image with tag", "alpine:latest", "docker.io/library/alpine", "latest", false},
		{"Valid image without tag", "ubuntu", "docker.io/library/ubuntu", "", false},
		{
			"Valid image with tag and full repo",
			"mytestrepo.io:5000/test:hello",
			"mytestrepo.io:5000/test",
			"hello",
			false,
		},
		{"Invalid image", "!!invalid!!", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, tag, err := ParseDockerImage(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf(
					"ParseDockerImage(%q) error = %v, wantError %v",
					tt.input,
					err,
					tt.wantError,
				)
			}

			if repo != tt.expectRepo {
				t.Errorf("ParseDockerImage(%q) repo = %q, want %q", tt.input, repo, tt.expectRepo)
			}

			if tag != tt.expectTag {
				t.Errorf("ParseDockerImage(%q) tag = %q, want %q", tt.input, tag, tt.expectTag)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "Item exists in slice",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "banana",
			expected: true,
		},
		{
			name:     "Item does not exist in slice",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "grape",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			item:     "banana",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}
