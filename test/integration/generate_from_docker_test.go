package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/komailo/kubeit/pkg/generate"
)

func buildDockerImage(dockerContext string, imageName string, labelArgs string) error {
	// Split the labelArgs into separate arguments
	labelArgsSplit := []string{}
	if labelArgs != "" {
		labelArgsSplit = strings.Fields(labelArgs) // Split on whitespace
	}

	// Construct the full command arguments
	cmdArgs := append([]string{"build", "-t", imageName}, labelArgsSplit...)
	cmdArgs = append(cmdArgs, dockerContext)

	// Construct the Docker build command
	cmd := exec.Command("docker", cmdArgs...)

	// Set the command's output to the current process's stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	return nil
}

func TestGenerateFromDockerImage(t *testing.T) {
	generateSetOptions := generate.Options{
		SourceConfigURI: "file://./testdata/kubeit-valid",
	}
	labelArgs, generateErrs, loadFileErrs := generate.DockerLabels(
		&generateSetOptions,
	)

	if len(loadFileErrs) != 0 {
		t.Errorf("Load file errors: %v", loadFileErrs)
	}

	if len(generateErrs) != 0 {
		t.Errorf("Generate errors: %v", generateErrs)
	}

	dockerContext, _ := filepath.Abs("./testdata/") // Path to the Dockerfile directory
	imageName := "my-image:latest"

	if err := buildDockerImage(dockerContext, imageName, labelArgs); err != nil {
		fmt.Printf("docker build error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Docker image built successfully!")

	generateSetOptions = generate.Options{
		SourceConfigURI: "docker://my-image:latest",
		KubeVersion:     "1.25.0",
	}
	generateErrs, loadFileErrs = generate.Manifests(
		&generateSetOptions,
	)

	if len(loadFileErrs) != 0 {
		t.Errorf("Load file errors: %v", loadFileErrs)
	}

	if len(generateErrs) != 0 {
		t.Errorf("Generate errors: %v", generateErrs)
	}
}
