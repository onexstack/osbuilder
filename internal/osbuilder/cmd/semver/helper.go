package semver

import (
	"fmt"
	"os"

	"github.com/onexstack/osbuilder/internal/osbuilder/semver/config"
)

// configFileNames defines the supported configuration file names in order of precedence
var configFileNames = []string{".osbuilder.yml", ".osbuilder.yaml", "osbuilder.yml", "osbuilder.yaml"}

const (
	currentWorkingDir = "."
)

// LoadFromFile loads configuration from a specific file path
func LoadFromFile(filePath string) (config.Uplift, error) {
	if filePath == "" {
		return config.Uplift{}, fmt.Errorf("configuration file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return config.Uplift{}, fmt.Errorf("configuration file not found: %s", filePath)
	}

	// Load the configuration
	cfg, err := config.Load(filePath)
	if err != nil {
		return config.Uplift{}, fmt.Errorf("failed to load configuration from %s: %w", filePath, err)
	}

	return cfg, nil
}
