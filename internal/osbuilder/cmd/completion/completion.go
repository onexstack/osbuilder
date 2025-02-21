// Package completion output shell completion code for the specified shell (bash or zsh).

//nolint:errcheck
package completion

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

const defaultBoilerPlate = `
# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for 
# this file is https://github.com/onexstack/onex.
`

var (
	completionLong = templates.LongDesc(`
		Output shell completion code for the specified shell (bash or zsh).
		The shell code must be evaluated to provide interactive
		completion of osbuilder commands.  This can be done by sourcing it from
		the .bash_profile.

		Detailed instructions on how to do this are available here:
		http://github.com/onexstack/onex/docs/installation/osbuilder.md#enabling-shell-autocompletion

		Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2`)

	completionExample = templates.Examples(`
		# Installing bash completion on macOS using homebrew
		## If running Bash 3.2 included with macOS
		    brew install bash-completion
		## or, if running Bash 4.1+
		    brew install bash-completion@2
		## If osbuilder is installed via homebrew, this should start working immediately.
		## If you've installed via other means, you may need add the completion to your completion directory
		    osbuilder completion bash > $(brew --prefix)/etc/bash_completion.d/osbuilder


		# Installing bash completion on Linux
		## If bash-completion is not installed on Linux, please install the 'bash-completion' package
		## via your distribution's package manager.
		## Load the osbuilder completion code for bash into the current shell
		    source <(osbuilder completion bash)
		## Write bash completion code to a file and source if from .bash_profile
		    osbuilder completion bash > ~/.onex/osbuilder.completion.bash.inc
		    printf "
		      # OneX shell completion
		      source '$HOME/.onex/osbuilder.completion.bash.inc'
		      " >> $HOME/.bash_profile
		    source $HOME/.bash_profile

		# Load the osbuilder completion code for zsh[1] into the current shell
		    source <(osbuilder completion zsh)
		# Set the osbuilder completion code for zsh[1] to autoload on startup
		    osbuilder completion zsh > "${fpath[1]}/_osbuilder"

		# Load the osbuilder completion code for fish[2] into the current shell
		    osbuilder completion fish | source
		# To load completions for each session, execute once: 
		    osbuilder completion fish > ~/.config/fish/completions/osbuilder.fish

		# Load the osbuilder completion code for powershell into the current shell
		    osbuilder completion powershell | Out-String | Invoke-Expression
		# Set osbuilder completion code for powershell to run on startup
		## Save completion code to a script and execute in the profile
		    osbuilder completion powershell > $HOME\.onex\completion.ps1
		    Add-Content $PROFILE "$HOME\.onex\completion.ps1"
		## Execute completion code in the profile
		    Add-Content $PROFILE "if (Get-Command osbuilder -ErrorAction SilentlyContinue) {
		        osbuilder completion powershell | Out-String | Invoke-Expression
		    }"
		## Add completion code directly to the $PROFILE script
		    osbuilder completion powershell >> $PROFILE`)
)

var completionShells = map[string]func(out io.Writer, boilerPlate string, cmd *cobra.Command) error{
	"bash":       runCompletionBash,
	"zsh":        runCompletionZsh,
	"fish":       runCompletionFish,
	"powershell": runCompletionPwsh,
}

// NewCmdCompletion creates the `completion` command.
func NewCmdCompletion(out io.Writer, boilerPlate string) *cobra.Command {
	shells := []string{}
	for s := range completionShells {
		shells = append(shells, s)
	}

	cmd := &cobra.Command{
		Use:                   "completion SHELL",
		DisableFlagsInUseLine: true,
		Short:                 "Output shell completion code for the specified shell (bash, zsh, fish, or powershell)",
		Long:                  completionLong,
		Example:               completionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(RunCompletion(out, boilerPlate, cmd, args))
		},
		ValidArgs: shells,
	}

	return cmd
}

// RunCompletion checks given arguments and executes command.
func RunCompletion(out io.Writer, boilerPlate string, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, "Shell not specified.")
	}
	if len(args) > 1 {
		return cmdutil.UsageErrorf(cmd, "Too many arguments. Expected only the shell type.")
	}
	run, found := completionShells[args[0]]
	if !found {
		return cmdutil.UsageErrorf(cmd, "Unsupported shell type %q.", args[0])
	}

	return run(out, boilerPlate, cmd.Parent())
}

func runCompletionBash(out io.Writer, boilerPlate string, kubectl *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}

	return kubectl.GenBashCompletionV2(out, true)
}

func runCompletionZsh(out io.Writer, boilerPlate string, kubectl *cobra.Command) error {
	zshHead := fmt.Sprintf("#compdef %[1]s\ncompdef _%[1]s %[1]s\n", kubectl.Name())
	out.Write([]byte(zshHead))

	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}

	return kubectl.GenZshCompletion(out)
}

func runCompletionFish(out io.Writer, boilerPlate string, kubectl *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}

	return kubectl.GenFishCompletion(out, true)
}

func runCompletionPwsh(out io.Writer, boilerPlate string, kubectl *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}

	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}

	return kubectl.GenPowerShellCompletionWithDesc(out)
}
