package addlicense

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/enescakir/emoji"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

const tmplApache = `Copyright {{.Year}} {{.Holder}}

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`

const tmplBSD = `Copyright (c) {{.Year}} {{.Holder}} All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.`

const tmplMIT = `Copyright (c) {{.Year}} {{.Holder}}

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.`

const tmplMPL = `This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.`

type copyrightData struct {
	Year   string
	Holder string
}

type file struct {
	path string
	mode os.FileMode
}

// AddlicenseOptions defines configuration for the "addlicense" command.
type AddlicenseOptions struct {
	RootDir     string   // working directory
	Holder      string   // copyright holder
	License     string   // license type: apache, bsd, mit, mpl
	LicenseFile string   // custom license file path
	Year        string   // copyright year(s)
	Verbose     bool     // verbose mode
	CheckOnly   bool     // check only mode
	SkipDirs    []string // regexps of directories to skip
	SkipFiles   []string // regexps of files to skip
	Patterns    []string // file/directory patterns to process

	licenseTemplate  map[string]*template.Template
	skipDirPatterns  []*regexp.Regexp
	skipFilePatterns []*regexp.Regexp

	genericiooptions.IOStreams
}

var (
	addlicenseLongDesc = templates.LongDesc(`The addlicense command ensures source code files have copyright license headers
		by scanning directory patterns recursively.

		It modifies all source files in place and avoids adding a license header
		to any file that already has one.

		The pattern argument can be provided multiple times, and may also refer
		to single files.`)

	addlicenseExample = templates.Examples(`# Add Apache license headers to all files in current directory
		osbuilder addlicense .

		# Add MIT license with custom holder
		osbuilder addlicense --license mit --holder "John Doe" ./src

		# Check if files have license headers without modifying
		osbuilder addlicense --check ./src

		# Skip certain directories and files
		osbuilder addlicense --skip-dirs "vendor,node_modules" --skip-files ".*_test.go" ./src

		# Use custom license file
		osbuilder addlicense --licensef ./custom-license.txt ./src`)
)

// NewAddlicenseCmd creates the "addlicense" command.
func NewAddlicenseCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &AddlicenseOptions{
		IOStreams:       ioStreams,
		Holder:          "Google LLC",
		License:         "apache",
		Year:            fmt.Sprint(time.Now().Year()),
		licenseTemplate: make(map[string]*template.Template),
	}

	// Initialize license templates
	opts.initLicenseTemplates()

	cmd := &cobra.Command{
		Use:                   "addlicense [flags] pattern [pattern ...]",
		Short:                 "Add license headers to source code files",
		Long:                  addlicenseLongDesc,
		Example:               addlicenseExample,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Patterns = args
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.Holder, "holder", "c", opts.Holder, "Copyright holder")
	cmd.Flags().StringVarP(&opts.License, "license", "l", opts.License, "License type: apache, bsd, mit, mpl")
	cmd.Flags().StringVarP(&opts.LicenseFile, "licensef", "f", "", "Custom license file (no default)")
	cmd.Flags().StringVarP(&opts.Year, "year", "y", opts.Year, "Copyright year(s)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "", false, "Verbose mode: print the name of the files that are modified")
	cmd.Flags().BoolVar(&opts.CheckOnly, "check", false, "Check only mode: verify presence of license headers and exit with non-zero code if missing")
	cmd.Flags().StringSliceVar(&opts.SkipDirs, "skip-dirs", nil, "Regexps of directories to skip")
	cmd.Flags().StringSliceVar(&opts.SkipFiles, "skip-files", nil, "Regexps of files to skip")

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *AddlicenseOptions) Complete() error {
	var err error
	o.RootDir, err = os.Getwd()
	if err != nil {
		return err
	}

	// Compile skip directory patterns
	if len(o.SkipDirs) > 0 {
		o.skipDirPatterns, err = o.getPatterns(o.SkipDirs)
		if err != nil {
			return err
		}
	}

	// Compile skip file patterns
	if len(o.SkipFiles) > 0 {
		o.skipFilePatterns, err = o.getPatterns(o.SkipFiles)
		if err != nil {
			return err
		}
	}

	return nil
}

// Validate ensures provided inputs are valid.
func (o *AddlicenseOptions) Validate() error {
	if o.LicenseFile == "" {
		if _, exists := o.licenseTemplate[o.License]; !exists {
			return fmt.Errorf("unknown license: %s", o.License)
		}
	}

	return nil
}

// Run performs the addlicense operation.
func (o *AddlicenseOptions) Run() error {
	data := &copyrightData{
		Year:   o.Year,
		Holder: o.Holder,
	}

	var tmpl *template.Template

	if o.LicenseFile != "" {
		d, err := ioutil.ReadFile(o.LicenseFile)
		if err != nil {
			return fmt.Errorf("license file: %v", err)
		}
		tmpl, err = template.New("").Parse(string(d))
		if err != nil {
			return fmt.Errorf("license file: %v", err)
		}
	} else {
		tmpl = o.licenseTemplate[o.License]
	}

	// Process at most 1000 files in parallel
	ch := make(chan *file, 1000)
	done := make(chan struct{})
	hasErrors := false

	go func() {
		var wg errgroup.Group
		for f := range ch {
			f := f // capture loop variable
			wg.Go(func() error {
				if o.CheckOnly {
					return o.checkFile(f.path, tmpl, data)
				} else {
					return o.processFile(f.path, f.mode, tmpl, data)
				}
			})
		}
		err := wg.Wait()
		close(done)
		if err != nil {
			hasErrors = true
		}
	}()

	for _, pattern := range o.Patterns {
		o.walk(ch, pattern)
	}
	close(ch)
	<-done

	if hasErrors {
		return errors.New("some files had errors")
	}

	if !o.CheckOnly {
		fmt.Fprintf(o.Out, "%s Successfully added license headers\n", emoji.CheckMarkButton)
	} else {
		fmt.Fprintf(o.Out, "%s License header check completed\n", emoji.CheckMarkButton)
	}

	return nil
}

func (o *AddlicenseOptions) initLicenseTemplates() {
	o.licenseTemplate["apache"] = template.Must(template.New("").Parse(tmplApache))
	o.licenseTemplate["mit"] = template.Must(template.New("").Parse(tmplMIT))
	o.licenseTemplate["bsd"] = template.Must(template.New("").Parse(tmplBSD))
	o.licenseTemplate["mpl"] = template.Must(template.New("").Parse(tmplMPL))
}

func (o *AddlicenseOptions) getPatterns(patterns []string) ([]*regexp.Regexp, error) {
	patternsRe := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		patternRe, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("can't compile regexp %q: %w", p, err)
		}
		patternsRe = append(patternsRe, patternRe)
	}
	return patternsRe, nil
}

func (o *AddlicenseOptions) walk(ch chan<- *file, start string) {
	_ = filepath.Walk(start, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(o.ErrOut, "%s error: %v\n", path, err)
			return nil
		}
		if fi.IsDir() {
			for _, pattern := range o.skipDirPatterns {
				if pattern.MatchString(fi.Name()) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		for _, pattern := range o.skipFilePatterns {
			if pattern.MatchString(fi.Name()) {
				return nil
			}
		}

		ch <- &file{path, fi.Mode()}
		return nil
	})
}

func (o *AddlicenseOptions) checkFile(path string, tmpl *template.Template, data *copyrightData) error {
	// Check if file extension is known
	lic, err := o.licenseHeader(path, tmpl, data)
	if err != nil {
		fmt.Fprintf(o.ErrOut, "%s: %v\n", path, err)
		return err
	}
	if lic == nil { // Unknown file extension
		return nil
	}

	// Check if file has a license
	isMissingLicenseHeader, err := o.fileHasLicense(path)
	if err != nil {
		fmt.Fprintf(o.ErrOut, "%s: %v\n", path, err)
		return err
	}
	if isMissingLicenseHeader {
		fmt.Fprintf(o.Out, "%s\n", path)
		return errors.New("missing license header")
	}
	return nil
}

func (o *AddlicenseOptions) processFile(path string, fmode os.FileMode, tmpl *template.Template, data *copyrightData) error {
	modified, err := o.addLicense(path, fmode, tmpl, data)
	if err != nil {
		fmt.Fprintf(o.ErrOut, "%s: %v\n", path, err)
		return err
	}
	if o.Verbose && modified {
		fmt.Fprintf(o.Out, "%s added license\n", path)
	}
	return nil
}

func (o *AddlicenseOptions) addLicense(path string, fmode os.FileMode, tmpl *template.Template, data *copyrightData) (bool, error) {
	lic, err := o.licenseHeader(path, tmpl, data)
	if err != nil || lic == nil {
		return false, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}
	if o.hasLicense(b) {
		return false, nil
	}

	line := o.hashBang(b)
	if len(line) > 0 {
		b = b[len(line):]
		if line[len(line)-1] != '\n' {
			line = append(line, '\n')
		}
		line = append(line, '\n')
		lic = append(line, lic...)
	}
	b = append(lic, b...)

	return true, ioutil.WriteFile(path, b, fmode)
}

func (o *AddlicenseOptions) fileHasLicense(path string) (bool, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	if o.hasLicense(b) {
		return false, nil
	}

	return true, nil
}

func (o *AddlicenseOptions) licenseHeader(path string, tmpl *template.Template, data *copyrightData) ([]byte, error) {
	var lic []byte
	var err error
	switch o.fileExtension(path) {
	default:
		return nil, nil
	case ".c", ".h":
		lic, err = o.prefix(tmpl, data, "/*", " * ", " */")
	case ".js", ".mjs", ".cjs", ".jsx", ".tsx", ".css", ".tf", ".ts":
		lic, err = o.prefix(tmpl, data, "/**", " * ", " */")
	case ".cc", ".cpp", ".cs", ".go", ".hh", ".hpp", ".java", ".m", ".mm", ".proto", ".rs", ".scala", ".swift", ".dart", ".groovy", ".kt", ".kts":
		lic, err = o.prefix(tmpl, data, "", "// ", "")
	case ".py", ".sh", ".yaml", ".yml", ".dockerfile", "dockerfile", ".rb", "gemfile":
		lic, err = o.prefix(tmpl, data, "", "# ", "")
	case ".el", ".lisp":
		lic, err = o.prefix(tmpl, data, "", ";; ", "")
	case ".erl":
		lic, err = o.prefix(tmpl, data, "", "% ", "")
	case ".hs", ".sql":
		lic, err = o.prefix(tmpl, data, "", "-- ", "")
	case ".html", ".xml", ".vue":
		lic, err = o.prefix(tmpl, data, "<!--", " ", "-->")
	case ".php":
		lic, err = o.prefix(tmpl, data, "", "// ", "")
	case ".ml", ".mli", ".mll", ".mly":
		lic, err = o.prefix(tmpl, data, "(**", "   ", "*)")
	}

	return lic, err
}

func (o *AddlicenseOptions) fileExtension(name string) string {
	if v := filepath.Ext(name); v != "" {
		return strings.ToLower(v)
	}
	return strings.ToLower(filepath.Base(name))
}

var head = []string{
	"#!",                       // shell script
	"<?xml",                    // XML declaration
	"<!doctype",                // HTML doctype
	"# encoding:",              // Ruby encoding
	"# frozen_string_literal:", // Ruby interpreter instruction
	"<?php",                    // PHP opening tag
}

func (o *AddlicenseOptions) hashBang(b []byte) []byte {
	line := make([]byte, 0, len(b))
	for _, c := range b {
		line = append(line, c)
		if c == '\n' {
			break
		}
	}
	first := strings.ToLower(string(line))
	for _, h := range head {
		if strings.HasPrefix(first, h) {
			return line
		}
	}
	return nil
}

func (o *AddlicenseOptions) hasLicense(b []byte) bool {
	n := 1000
	if len(b) < 1000 {
		n = len(b)
	}

	return bytes.Contains(bytes.ToLower(b[:n]), []byte("copyright")) ||
		bytes.Contains(bytes.ToLower(b[:n]), []byte("mozilla public"))
}

func (o *AddlicenseOptions) prefix(t *template.Template, d *copyrightData, top, mid, bot string) ([]byte, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, d); err != nil {
		return nil, fmt.Errorf("render template failed: %w", err)
	}
	var out bytes.Buffer
	if top != "" {
		fmt.Fprintln(&out, top)
	}
	s := bufio.NewScanner(&buf)
	for s.Scan() {
		fmt.Fprintln(&out, strings.TrimRightFunc(mid+s.Text(), unicode.IsSpace))
	}
	if bot != "" {
		fmt.Fprintln(&out, bot)
	}
	fmt.Fprintln(&out)

	return out.Bytes(), nil
}
