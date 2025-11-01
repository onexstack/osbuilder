package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ImageConfig holds container image build configuration and Dockerfile generation options.
type ImageConfig struct {
	// RegistryPrefix is the image registry prefix where images are published.
	// (e.g., "docker.io/nginx", "gcr.io/distroless").
	RegistryPrefix string `yaml:"registryPrefix"`

	// DockerfileMode selects how Dockerfiles are generated for the project.
	// Supported values:
	// - "none": Do not generate a Dockerfile. You must provide your own at
	//   build/docker/<component_name>/Dockerfile.
	// - "runtime-only": Generate a runtime-only Dockerfile (expects an external build artifact).
	//   Useful for local debugging or when CI produces binaries/assets separately.
	// - "multi-stage": Generate a multi-stage Dockerfile (builder + runtime).
	// - "combined": Generate both Dockerfile variants:
	//   - Multi-stage: saved as "Dockerfile"
	//   - Runtime-only: saved as "Dockerfile.runtime-only"
	DockerfileMode string `yaml:"dockerfileMode"`

	// Distroless controls the base image used for the runtime stage.
	// - true: Use "gcr.io/distroless/base-debian12:nonroot" (recommended for production).
	// - false: Use "debian:bookworm" (convenient for testing/debugging).
	Distroless bool `yaml:"distroless"`
}

// Project is the top-level configuration for a generated project.
type Project struct {
	// Scaffold identifies the scaffold preset used.
	Scaffold string `yaml:"scaffold"`
	// Version tracks the schema or scaffold version.
	Version string `yaml:"version"`
	// Metadata provides project-level details (author, deployment, registry, etc.).
	Metadata *Metadata `yaml:"metadata"`

	// Use lowerCamelCase in YAML and pointer slices + omitempty for sparsity.
	WebServers []*WebServer      `yaml:"webServers,omitempty"`
	Jobs       []*Job            `yaml:"jobs,omitempty"`    // previously: JobServer
	CLIApps    []*CLIApplication `yaml:"cliApps,omitempty"` // previously: CLIApp

	// Common derived data used during generation (not serialized).
	D *GeneratedData `yaml:"-"`
}

// Metadata holds general project information and build/deploy preferences.
type Metadata struct {
	// ModulePath is the Go module import path as declared in the project's go.mod file.
	ModulePath string `yaml:"modulePath"`
	// Short is the short description shown in the 'help' output.
	ShortDescription string `yaml:"shortDescription"`
	// Long is the long message shown in the 'help <this-command>' output.
	LongMessage string `yaml:"longMessage"`
	// Image holds image build configuration.
	Image ImageConfig `yaml:"image"`
	// DeploymentMethod selects how to deploy (e.g., "kubernetes", "systemd").
	// Note: name kept for backward compatibility; often referred to as "deploymentMode".
	DeploymentMethod string `yaml:"deploymentMethod"`
	// MakefileMode selects the Makefile generation mode used by the project
	// (e.g., "none", "structed", "unstructured").
	MakefileMode string `yaml:"makefileMode"`
	// Author is the maintainer name.
	Author string `yaml:"author"`
	// Email is the maintainer email.
	Email string `yaml:"email"`
}

// FindWebServer returns the first web server whose BinaryName matches.
// Deprecated: use WebServerByBinary to get an ok flag and avoid nil-struct ambiguity.
func (p *Project) FindWebServer(binaryName string) *WebServer {
	ws, _ := p.WebServerByBinary(binaryName)
	return ws
}

// WebServerByBinary returns a web server and a boolean indicating if it was found.
func (p *Project) WebServerByBinary(binaryName string) (*WebServer, bool) {
	for _, ws := range p.WebServers {
		if ws != nil && ws.BinaryName == binaryName {
			return ws, true
		}
	}
	return nil, false
}

// Join builds a path under the project root (WorkDir).
// If WorkDir is empty, it joins the elements as-is.
func (p *Project) Join(elements ...string) string {
	base := ""
	if p != nil && p.D != nil {
		base = p.D.WorkDir
	}
	parts := elements
	if base != "" {
		parts = append([]string{base}, elements...)
	}
	return filepath.Join(parts...)
}

// Root returns the project root directory (WorkDir). Empty if not set.
func (p *Project) Root() string {
	if p == nil || p.D == nil {
		return ""
	}
	return p.D.WorkDir
}

// InternalPkg returns the path to the shared internal package directory.
func (p *Project) InternalPkg() string {
	return filepath.Join("internal", "pkg")
}

// Configs returns the path of configs.
func (p *Project) Configs() string {
	return "configs"
}

// Scripts returns the path to the scripts directory.
func (p *Project) Scripts() string {
	return filepath.Join("scripts")
}

// Save writes the project to a YAML file at filename.
// It creates parent directories as needed and uses a 2-space indentation.
func (p *Project) Save(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Simple, reliable write with proper close semantics.
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %q: %w", filename, err)
	}
	defer func() { _ = f.Close() }()

	// Header to prepend. It will be turned into YAML comments (`# ...`).
	header := `Copyright 2024 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file. The original repo for
this file is https://github.com/onexstack/onex.

Code generated by osbuilder.
Only allow adding new components.
Modifying existing components or their configurations IS NOT ALLOWED.
Delete PROJECT or rename PROJECT IS NOT ALLOWED.`

	// Write header as YAML comments to keep the file a valid YAML document.
	// We prefix each line with "# ".
	commented := "# " + strings.ReplaceAll(header, "\n", "\n# ")
	if _, err := fmt.Fprintln(f, commented); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(p); err != nil {
		_ = enc.Close()
		return fmt.Errorf("encode yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}
	return nil
}
