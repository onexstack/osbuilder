package create

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	nirvanaproject "github.com/caicloud/nirvana/utils/project"
	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"

	"github.com/onexstack/osbuilder/internal/osbuilder/types"
)

//
// Configuration I/O
//

// LoadProjectFromFile opens, reads, and decodes a YAML project file.
//
// It enables strict decoding by default: unknown YAML fields will cause an error.
func LoadProjectFromFile(filename string) (*types.Project, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open project config %q: %w", filename, err)
	}
	defer func() { _ = f.Close() }()

	return DecodeProjectYAML(f, true)
}

// LoadProjectFromBase64 decodes base64 string and loads project configuration
func LoadProjectFromBase64(configBase64 string) (*types.Project, error) {
	// Decode base64 string
	yamlData, err := base64.StdEncoding.DecodeString(configBase64)
	if err != nil {
		return nil, fmt.Errorf("decode base64 config: %w", err)
	}

	// Create a string reader from decoded data
	reader := strings.NewReader(string(yamlData))

	// Use DecodeProjectYAML to parse the YAML
	return DecodeProjectYAML(reader, true)
}

// DecodeProjectYAML parses YAML from r. When strict is true, unknown fields are rejected.
func DecodeProjectYAML(r io.Reader, strict bool) (*types.Project, error) {
	dec := yaml.NewDecoder(r)
	if strict {
		dec.KnownFields(true)
	}
	var proj types.Project
	if err := dec.Decode(&proj); err != nil {
		return nil, fmt.Errorf("decode project yaml: %w", err)
	}
	return &proj, nil
}

// SaveProjectToFile writes the project as YAML to filename, creating parent dirs as needed.
func SaveProjectToFile(filename string, proj *types.Project) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create project config %q: %w", filename, err)
	}
	defer func() { _ = f.Close() }()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(proj); err != nil {
		_ = enc.Close()
		return fmt.Errorf("encode project yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("close yaml encoder: %w", err)
	}
	return nil
}

// PrintClosingTips prints a short closing section (thanks, docs link, etc.).
func PrintClosingTips(projectName string) {
	learnURL := color.MagentaString("https://t.zsxq.com/5T0qC")
	fmt.Printf("\n%s Thanks for using osbuilder!\n", emoji.Handshake)
	fmt.Printf("%s Visit %s to learn how to develop %s project.\n", emoji.BackhandIndexPointingRight, learnURL, projectName)
}

// MustModulePath returns the Go module path for rootDir.
// It never panics; the module path is returned with any lookup error ignored.
func MustModulePath(modulePath string, rootDir string) string {
	if modulePath != "" {
		return modulePath
	}

	modulePath, _ = nirvanaproject.PackageForPath(rootDir)
	return modulePath
}
