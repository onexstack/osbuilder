package types

import (
	"fmt"
	"strings"
)

// CLIApplication represents a CLI app component in project config.
// AppType corresponds to YAML field "type" (e.g., "cli").
type CLIApplication struct {
	// AppType is the category of the CLI app as defined in the config YAML key "type".
	// This keeps the YAML key as "type" for backward compatibility.
	AppType string `yaml:"type"`
}

// Validate ensures the CLIApplication is well-formed.
func (c CLIApplication) Validate() error {
	if strings.TrimSpace(c.AppType) == "" {
		return fmt.Errorf("cli application: field 'type' must not be empty")
	}
	return nil
}

// IsZero reports whether the struct has no meaningful data.
func (c CLIApplication) IsZero() bool {
	return strings.TrimSpace(c.AppType) == ""
}

// NewCLIApplication creates a new CLIApplication instance.
func NewCLIApplication(appType string) CLIApplication {
	return CLIApplication{AppType: strings.TrimSpace(appType)}
}
