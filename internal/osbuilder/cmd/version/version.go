// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package version print the client and server version information.
package version

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/onexstack/onexstack/pkg/version"
	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

const currentModule = "github.com/onexstack/osbuilder"

// Version is a struct for version information.
type Version struct {
	ClientVersion *version.Info `json:"clientVersion,omitempty" yaml:"clientVersion,omitempty"`
}

var versionExample = templates.Examples(`
		# Print the client and server versions for the current context
		onexctl version`)

// Options is a struct to support version command.
type Options struct {
	Short  bool
	Output string

	genericiooptions.IOStreams
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericiooptions.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

// NewCmdVersion returns a cobra command for fetching versions.
func NewCmdVersion(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the client version information",
		Long:    "Print the client version information for the current context",
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")
	cmd.Flags().BoolVar(&o.Short, "short", o.Short, "If true, print just the version number.")

	return cmd
}

// Complete completes all the required options.
func (o *Options) Complete(f cmdutil.Factory, cmd *cobra.Command) error {
	return nil
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
		return errors.New(`--output must be 'yaml' or 'json'`)
	}

	return nil
}

// Run executes version command.
func (o *Options) Run() error {
	var (
		serverErr   error
		versionInfo Version
	)

	clientVersion := version.Get()
	if clientVersion.GitVersion == "" {
		clientVersion = version.GetFromDebugInfo(currentModule)
	}

	switch o.Output {
	case "":
		if o.Short {
			fmt.Fprintf(o.Out, "%s\n", clientVersion.String())
		} else {
			fmt.Fprintf(o.Out, "Client Version: %s\n", clientVersion.ToJSON())
		}
	case "yaml":
		marshaled, err := yaml.Marshal(&versionInfo)
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	case "json":
		marshaled, err := json.MarshalIndent(&versionInfo, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	default:
		// There is a bug in the program if we hit this case.
		// However, we follow a policy of never panicking.
		return fmt.Errorf("VersionOptions were not validated: --output=%q should have been rejected", o.Output)
	}

	return serverErr
}
