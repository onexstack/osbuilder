package types

import (
	"path/filepath"
	"strings"
)

// GeneratedData holds derived metadata and common paths used during code generation.
type GeneratedData struct {
	// WorkDir is the project root directory (absolute or relative).
	WorkDir string
	// APIVersion is the API version (e.g., "v1").
	APIVersion string
	// ModuleName is the Go module path for the project (e.g., "github.com/acme/project").
	ModuleName string
	// APIAlias is a short alias for the API version (defaults to APIVersion if empty).
	APIAlias string
	// Boilerplate holds license/header boilerplate text inserted into generated files.
	Boilerplate string

	// ProjectName is the human-friendly project name; defaults to the basename of WorkDir.
	ProjectName string
	// RegistryPrefix is the image registry prefix used for builds and releases.
	RegistryPrefix string
	// EnvironmentPrefix is a normalized prefix (e.g., "PROJECT") for environment variables.
	EnvironmentPrefix string
}

// Complete fills sensible defaults if fields are empty and returns the receiver.
//
// Defaults applied:
// - APIAlias ← APIVersion (if empty)
// - ProjectName ← base(WorkDir) (if empty)
// - EnvironmentPrefix ← upper(ProjectName) with '-' replaced by '_' (if empty)
func (d *GeneratedData) Complete() *GeneratedData {
	if d == nil {
		return d
	}
	if strings.TrimSpace(d.APIAlias) == "" {
		d.APIAlias = d.APIVersion
	}
	if strings.TrimSpace(d.ProjectName) == "" && strings.TrimSpace(d.WorkDir) != "" {
		d.ProjectName = filepath.Base(d.WorkDir)
	}
	if strings.TrimSpace(d.EnvironmentPrefix) == "" && strings.TrimSpace(d.ProjectName) != "" {
		p := strings.ToUpper(d.ProjectName)
		p = strings.ReplaceAll(p, "-", "_")
		d.EnvironmentPrefix = p
	}
	return d
}
