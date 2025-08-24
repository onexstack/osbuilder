// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package cmd create a root cobra command and add subcommands to it.
package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/ptr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	genericcmd "k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kubectl/pkg/util/term"

	"github.com/onexstack/osbuilder/internal/osbuilder/cmd/color"
	"github.com/onexstack/osbuilder/internal/osbuilder/cmd/completion"
	"github.com/onexstack/osbuilder/internal/osbuilder/cmd/create"

	//"github.com/onexstack/osbuilder/internal/osbuilder/cmd/jwt"
	"github.com/onexstack/osbuilder/internal/osbuilder/cmd/cmd"
	"github.com/onexstack/osbuilder/internal/osbuilder/cmd/options"
	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"

	//"github.com/onexstack/osbuilder/internal/osbuilder/cmd/validate"
	"github.com/onexstack/osbuilder/internal/osbuilder/cmd/version"
	clioptions "github.com/onexstack/osbuilder/internal/osbuilder/util/options"
)

const (
	osbuilderCmdHeaders = "OSCTL_COMMAND_HEADERS"
	defaultConfigName   = "osbuilder.yaml"
	defaultHomeDir      = ".onexstack"
)

type OSCtlOptions struct {
	PluginHandler genericcmd.PluginHandler
	Arguments     []string
	// ConfigFlags   *genericiooptions.ConfigFlags

	genericiooptions.IOStreams
}

// NewDefaultOSCtlCommand creates the `osbuilder` command with default arguments.
func NewDefaultOSCtlCommand() *cobra.Command {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	return NewDefaultOSCtlCommandWithArgs(OSCtlOptions{
		PluginHandler: genericcmd.NewDefaultPluginHandler([]string{"osbuilder"}),
		Arguments:     os.Args,
		IOStreams:     ioStreams,
	})
}

// NewDefaultOSCtlCommandWithArgs creates the `osbuilder` command with arguments.
func NewDefaultOSCtlCommandWithArgs(o OSCtlOptions) *cobra.Command {
	cmd := NewOSCtlCommand(o)

	if o.PluginHandler == nil {
		return cmd
	}

	if len(o.Arguments) > 1 {
		cmdPathPieces := o.Arguments[1:]

		// only look for suitable extension executables if
		// the specified command does not already exist
		if foundCmd, foundArgs, err := cmd.Find(cmdPathPieces); err != nil {
			// Also check the commands that will be added by Cobra.
			// These commands are only added once rootCmd.Execute() is called, so we
			// need to check them explicitly here.
			var cmdName string // first "non-flag" arguments
			for _, arg := range cmdPathPieces {
				if !strings.HasPrefix(arg, "-") {
					cmdName = arg
					break
				}
			}

			switch cmdName {
			case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
				// Don't search for a plugin
			default:
				if err := genericcmd.HandlePluginCommand(o.PluginHandler, cmdPathPieces, 1); err != nil {
					fmt.Fprintf(o.IOStreams.ErrOut, "Error: %v\n", err)
					os.Exit(1)
				}
			}
		} else if err == nil {
			if !cmdutil.CmdPluginAsSubcommand.IsDisabled() {
				// Command exists(e.g. kubectl create), but it is not certain that
				// subcommand also exists (e.g. kubectl create networkpolicy)
				// we also have to eliminate kubectl create -f
				if genericcmd.IsSubcommandPluginAllowed(foundCmd.Name()) && len(foundArgs) >= 1 && !strings.HasPrefix(foundArgs[0], "-") {
					subcommand := foundArgs[0]
					builtinSubcmdExist := false
					for _, subcmd := range foundCmd.Commands() {
						if subcmd.Name() == subcommand {
							builtinSubcmdExist = true
							break
						}
					}

					if !builtinSubcmdExist {
						if err := genericcmd.HandlePluginCommand(o.PluginHandler, cmdPathPieces, len(cmdPathPieces)-len(foundArgs)+1); err != nil {
							fmt.Fprintf(o.IOStreams.ErrOut, "Error: %v\n", err)
							os.Exit(1)
						}
					}
				}
			}
		}
	}

	return cmd
}

// NewOSCtlCommand creates the `osbuilder` command and its nested children.
func NewOSCtlCommand(o OSCtlOptions) *cobra.Command {
	warningHandler := rest.NewWarningWriter(o.IOStreams.ErrOut, rest.WarningWriterOptions{Deduplicate: true, Color: term.AllowsColorOutput(o.IOStreams.ErrOut)})
	warningsAsErrors := false
	opts := clioptions.NewOptions()
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "osbuilder",
		Short: i18n.T("osbuilder is a command-line tool for the onex technology stack scaffold"),
		Long: templates.LongDesc(`
		osbuilder is a command-line tool for the onex technology stack scaffold..

		Find more information at:
			https://github.com/onexstack/osbuilder/blob/master/docs/guide/en-US/cmd/osbuilder/osbuilder.md`),
		Run: runHelp,
		// Hook before and after Run initialize and write profiles to disk,
		// respectively.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			rest.SetDefaultWarningHandler(warningHandler)

			if cmd.Name() == cobra.ShellCompRequestCmd {
				// This is the __complete or __completeNoDesc command which
				// indicates shell completion has been requested.
				plugin.SetupPluginCompletion(cmd, args)
			}

			opts.Complete()

			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			if err := flushProfiling(); err != nil {
				return err
			}
			if warningsAsErrors {
				count := warningHandler.WarningCount()
				switch count {
				case 0:
					// no warnings
				case 1:
					return fmt.Errorf("%d warning received", count)
				default:
					return fmt.Errorf("%d warnings received", count)
				}
			}
			return nil
		},
	}
	// From this point and forward we get warnings on flags that contain "_" separators
	// when adding them with hyphen instead of the original name.
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	flags := cmds.PersistentFlags()

	addProfilingFlags(flags)

	flags.BoolVar(&warningsAsErrors, "warnings-as-errors", warningsAsErrors, "Treat warnings received from the server as errors and exit with a non-zero exit code")

	opts.AddFlags(flags)
	// Updates hooks to add onexctl command headers: SIG CLI KEP 859.
	addCmdHeaderHooks(cmds, opts)

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	_ = viper.BindPFlags(cmds.PersistentFlags())
	cobra.OnInitialize(core.OnInitialize(ptr.To(viper.GetString(clioptions.FlagConfig)), "OSCTL", searchDirs(), defaultConfigName))
	cmds.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	f := cmdutil.NewFactory(opts)

	groups := templates.CommandGroups{
		{
			Message: "Basic Commands (Beginner):",
			Commands: []*cobra.Command{
				color.NewCmdColor(f, o.IOStreams),
				cmd.NewCmdCmd(f, o.IOStreams),
			},
		},
		{
			Message: "Project Commands:",
			Commands: []*cobra.Command{
				create.NewCmdCreate(f, o.IOStreams),
			},
		},
		{
			Message:  "Troubleshooting and Debugging Commands:",
			Commands: []*cobra.Command{},
		},
		{
			Message: "Settings Commands:",
			Commands: []*cobra.Command{
				completion.NewCmdCompletion(o.IOStreams.Out, ""),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}

	// Hide the "alpha" subcommand if there are no alpha commands in this build.
	alpha := NewCmdAlpha(f, o.IOStreams)
	if !alpha.HasSubCommands() {
		filters = append(filters, alpha.Name())
	}

	// Add plugin command group to the list of command groups.
	// The commands are only injected for the scope of showing help and completion, they are not
	// invoked directly.
	pluginCommandGroup := plugin.GetPluginCommandGroup(cmds)
	groups = append(groups, pluginCommandGroup)

	templates.ActsAsRootCommand(cmds, filters, groups...)

	cmds.AddCommand(alpha)
	// cmds.AddCommand(cmdconfig.NewCmdConfig(f, clientcmd.NewDefaultPathOptions(), o.IOStreams))
	cmds.AddCommand(plugin.NewCmdPlugin(o.IOStreams))
	cmds.AddCommand(version.NewCmdVersion(f, o.IOStreams))
	cmds.AddCommand(options.NewCmdOptions(o.IOStreams.Out))

	// Stop warning about normalization of flags. That makes it possible to
	// add the klog flags later.
	cmds.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)
	return cmds
}

// addCmdHeaderHooks performs updates on two hooks:
//  1. Modifies the passed "cmds" persistent pre-run function to parse command headers.
//     These headers will be subsequently added as X-headers to every
//     REST call.
//  2. Adds CommandHeaderRoundTripper as a wrapper around the standard
//     RoundTripper. CommandHeaderRoundTripper adds X-Headers then delegates
//     to standard RoundTripper.
//
// For beta, these hooks are updated unless the ONEX_COMMAND_HEADERS environment variable
// is set, and the value of the env var is false (or zero).
// See SIG CLI KEP 859 for more information:
//
//	https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/859-kubectl-headers
func addCmdHeaderHooks(cmds *cobra.Command, _ *clioptions.Options) {
	// If the feature gate env var is set to "false", then do no add kubectl command headers.
	if value, exists := os.LookupEnv(osbuilderCmdHeaders); exists {
		if value == "false" || value == "0" {
			klog.V(5).Infoln("onexctl command headers turned off")
			return
		}
	}
	klog.V(5).Infoln("onexctl command headers turned on")
	crt := &genericclioptions.CommandHeaderRoundTripper{}
	existingPreRunE := cmds.PersistentPreRunE
	// Add command parsing to the existing persistent pre-run function.
	cmds.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		crt.ParseCommandHeaders(cmd, args)
		return existingPreRunE(cmd, args)
	}
}

func runHelp(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
}

// searchDirs 返回默认的配置文件搜索目录.
func searchDirs() []string {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	// 如果获取用户主目录失败，则打印错误信息并退出程序
	cobra.CheckErr(err)
	return []string{filepath.Join(homeDir, defaultHomeDir), "."}
}
