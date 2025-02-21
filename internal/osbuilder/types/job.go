package types

import (
	"fmt"
	"strings"
)

// Job represents a background job component in the project configuration.
// JobType is mapped to the YAML key "type" for backward compatibility.
type Job struct {
	// JobType defines the category/implementation of the job (YAML: "type").
	// Example values depend on your schema (e.g., "cron", "worker").
	JobType string `yaml:"type"`
}

// Type returns the job type (preferred accessor).
func (j *Job) Type() string {
	return j.JobType
}

// Validate ensures the job configuration has the required fields set.
func (j *Job) Validate() error {
	if strings.TrimSpace(j.JobType) == "" {
		return fmt.Errorf("job: field 'type' must not be empty")
	}
	return nil
}
