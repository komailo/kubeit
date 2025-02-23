package apis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/jsonschema"

	appv1alpha1 "github.com/komailo/kubeit/pkg/apis/application/v1alpha1"
)

// schemaMap contains all API versions and their corresponding struct
var schemaMap = map[string]interface{}{
	appv1alpha1.GroupVersion + ".Application": appv1alpha1.Application{},
}

// GenerateSchemas generates JSON schemas for all registered API types
func GenerateSchemas(outputDir string) error {
	for name, obj := range schemaMap {
		schema := jsonschema.Reflect(obj)
		schemaJSON, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to generate schema for %s: %w", name, err)
		}

		// Extract API Group and Version
		parts := strings.Split(name, "/") // Example: kubeit.komailo.github.io/v1alpha1.Application
		if len(parts) < 2 {
			return fmt.Errorf("invalid GroupVersion format: %s", name)
		}

		group := parts[0]          // kubeit.komailo.github.io
		versionAndKind := parts[1] // v1alpha1.Application
		versionParts := strings.Split(versionAndKind, ".")
		version := versionParts[0] // v1alpha1
		kind := versionParts[1]    // Application

		// Define schema output directory
		dirPath := filepath.Join(outputDir, group, version)

		// Ensure directory exists
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create schema directory: %w", err)
		}

		// Define schema file path (Fix: No need for v1alpha1 in filename)
		fileName := filepath.Join(dirPath, kind+".json")

		// Write schema file
		err = os.WriteFile(fileName, schemaJSON, 0644)
		if err != nil {
			return fmt.Errorf("failed to write schema file %s: %w", fileName, err)
		}

		fmt.Println("Generated schema:", fileName)
	}
	return nil
}
