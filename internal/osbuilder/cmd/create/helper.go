package create

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	nirvanaproject "github.com/caicloud/nirvana/utils/project"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"

	"github.com/onexstack/osbuilder/internal/osbuilder/file"
	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
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

//
// Template-based file generation
//

// Generate renders templates to files using the provided FileManager.
//
// pairs maps destination relative paths to template paths under the embedded template FS.
// funcs contains template functions, and tplData is the data context for templates.
func Generate(fm *file.FileManager, pairs map[string]string, funcs template.FuncMap, tplData *types.TemplateData) error {
	fs := helper.NewFileSystem("/")

	for relPath, tplPath := range pairs {
		dstPath := tplData.Project.Join(relPath)

		// Parse template
		tmpl, err := template.New(filepath.Base(tplPath)).Funcs(funcs).Parse(fs.Content(tplPath))
		if err != nil {
			return fmt.Errorf("parse template %q for %q: %w", tplPath, dstPath, err)
		}

		// Execute template
		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, tplData); err != nil {
			return fmt.Errorf("execute template for %q (tpl: %s): %w", dstPath, color.RedString("%s", tplPath), err)
		}

		// Optional Go formatting
		var out []byte
		if isGoFile(dstPath) {
			out, err = format.Source(buf.Bytes())
			if err != nil {
				// Print the unformatted content to aid debugging
				fmt.Printf(color.RedString(buf.String()))
				return fmt.Errorf("format go source %q: %w", dstPath, err)
			}
		} else {
			out = buf.Bytes()
		}

		// Write output
		if err = fm.WriteFile(dstPath, out); err != nil {
			return fmt.Errorf("write file %q: %w", dstPath, err)
		}
	}

	return nil
}

func isGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}

//
// Console helpers
//

// PrintClosingTips prints a short closing section (thanks, docs link, etc.).
func PrintClosingTips(projectName string) {
	learnURL := color.MagentaString("https://t.zsxq.com/5T0qC")
	fmt.Println("\n🤝 Thanks for using osbuilder.")
	fmt.Printf("👉 Visit %s to learn how to develop %s project.\n", learnURL, projectName)
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
