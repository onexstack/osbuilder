package semver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/enescakir/emoji"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/config"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/context"
	semverutil "github.com/onexstack/osbuilder/internal/osbuilder/semver/semver"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/bump"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gitcheck"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gitcommit"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gpgimport"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/after"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/afterbump"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/before"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/beforebump"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/nextcommit"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/nextsemver"
)

// BumpOptions defines configuration for the "bump" subcommand.
type BumpOptions struct {
	// Command specific options
	Prerelease string // prerelease suffix

	// Semver configuration (inherited from parent semver command)
	SemverOptions *SemverOptions

	genericiooptions.IOStreams
}

var (
	bumpLongDesc = templates.LongDesc(`Calculates the next semantic version based on conventional commits since the
    last release (or identifiable tag) and bumps (or patches) a configurable set of
    files with said version.
    
    JSON Path or Regex Pattern matching is supported when scanning files for an 
    existing semantic version. OSBuilder automatically handles the staging and 
    pushing of modified files to the git remote, but this behavior can be disabled, 
    to manage this action manually.

    Configuring a bump requires an OSBuilder configuration file to exist within the
    root of your project. The tool supports various file formats and patterns for
    version bumping including JSON, YAML, XML, and plain text files.`)

	bumpExample = templates.Examples(`# Bump (patch) all configured files with the next calculated semantic version
        osbuilder semver bump

        # Append a prerelease suffix to the next calculated semantic version
        osbuilder semver bump --prerelease beta.1

        # Bump files but do not stage or push any changes back to the git remote
        osbuilder semver --no-stage bump

        # Perform a dry run to see what would be changed
        osbuilder semver --dry-run bump

        # Bump files with debug output
        osbuilder semver --debug bump --prerelease alpha.1`)
)

// Pipeline for bumping files
var bumpFilesPipeline = []task.Runner{
	gitcheck.Task{},
	before.Task{},
	gpgimport.Task{},
	nextsemver.Task{},
	nextcommit.Task{},
	beforebump.Task{},
	bump.Task{},
	afterbump.Task{},
	gitcommit.Task{},
	after.Task{},
}

// NewBumpCmd creates the "bump" subcommand.
func NewBumpCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams, semverOptions *SemverOptions) *cobra.Command {
	opts := &BumpOptions{
		IOStreams:     ioStreams,
		SemverOptions: semverOptions,
	}

	cmd := &cobra.Command{
		Use:                   "bump",
		Short:                 "Bump the semantic version within files",
		Long:                  bumpLongDesc,
		Example:               bumpExample,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}

	// Bump-specific flags (only flags unique to this subcommand)
	cmd.Flags().StringVar(&opts.Prerelease, "prerelease", "", "Append a prerelease suffix to next calculated semantic version")

	// Note: semver flags are inherited from parent semver command via persistent flags
	// No need to redefine --dry-run, --debug, --silent, etc.

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *BumpOptions) Complete() error {
	// Validate that semver options is provided
	if o.SemverOptions == nil {
		return fmt.Errorf("semver options is required")
	}

	// Additional completion logic can be added here
	return nil
}

// Validate ensures provided inputs are valid.
func (o *BumpOptions) Validate() error {
	// Check if we're in a git repository
	if _, err := os.Stat(filepath.Join(o.SemverOptions.RootDir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories): %s", o.SemverOptions.RootDir)
	}

	// Validate prerelease format if provided
	if o.Prerelease != "" {
		if _, _, err := semverutil.ParsePrerelease(o.Prerelease); err != nil {
			return fmt.Errorf("invalid prerelease format: %w", err)
		}
	}

	// Validate conflicting options (semver + local)
	if o.SemverOptions.Silent && o.SemverOptions.Debug {
		return fmt.Errorf("cannot use --silent and --debug flags together")
	}

	return nil
}

// Run performs the bump operation.
func (o *BumpOptions) Run() error {
	return o.executeBumpPipeline()
}

// executeBumpPipeline runs the main bump pipeline
func (o *BumpOptions) executeBumpPipeline() error {
	ctx, err := o.setupBumpContext()
	if err != nil {
		return fmt.Errorf("failed to setup context: %w", err)
	}

	if err := task.Execute(ctx, bumpFilesPipeline); err != nil {
		return fmt.Errorf("failed to execute bump pipeline: %w", err)
	}

	o.PrintGettingStarted()
	return nil
}

// setupBumpContext creates and configures the context for bump operations
func (o *BumpOptions) setupBumpContext() (*context.Context, error) {
	cfg, err := o.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.New(cfg, o.Out)

	// Set values from semver options
	ctx.Debug = o.SemverOptions.Debug
	ctx.DryRun = o.SemverOptions.DryRun
	ctx.NoPush = o.SemverOptions.NoPush
	ctx.NoStage = o.SemverOptions.NoStage
	ctx.Out = o.Out

	// Handle prerelease suffix if one is provided
	if o.Prerelease != "" {
		var err error
		if ctx.Prerelease, ctx.Metadata, err = semverutil.ParsePrerelease(o.Prerelease); err != nil {
			return nil, fmt.Errorf("failed to parse prerelease: %w", err)
		}
	}

	// Set values from semver options
	ctx.IgnoreExistingPrerelease = o.SemverOptions.IgnoreExistingPrerelease
	ctx.FilterOnPrerelease = o.SemverOptions.FilterOnPrerelease

	// Handle git config. Command line flag takes precedence
	ctx.IgnoreDetached = o.SemverOptions.IgnoreDetached
	if !ctx.IgnoreDetached && ctx.Config.Git != nil {
		ctx.IgnoreDetached = ctx.Config.Git.IgnoreDetached
	}

	ctx.IgnoreShallow = o.SemverOptions.IgnoreShallow
	if !ctx.IgnoreShallow && ctx.Config.Git != nil {
		ctx.IgnoreShallow = ctx.Config.Git.IgnoreShallow
	}

	return ctx, nil
}

// loadConfig loads configuration from the specified directory
func (o *BumpOptions) loadConfig() (config.Uplift, error) {
	if o.SemverOptions.ConfigDir == "" {
		// Return default config if no config directory specified
		return config.Uplift{}, nil
	}

	// TODO: Implement actual configuration loading logic
	// This should load configuration from o.SemverOptions.ConfigDir
	// For example:
	// return config.LoadFromDirectory(o.SemverOptions.ConfigDir)

	// For now, return a default config
	return config.Uplift{}, nil
}

// PrintGettingStarted prints formatted success information.
func (o *BumpOptions) PrintGettingStarted() {
	// Don't print anything if silent mode is enabled
	if o.SemverOptions.Silent {
		return
	}

	if o.SemverOptions.DryRun {
		fmt.Fprintf(o.Out, "%s Bump operation completed (dry-run mode)\n", emoji.CheckMarkButton)
	} else {
		fmt.Fprintf(o.Out, "%s Successfully bumped version in files\n", emoji.CheckMarkButton)
	}

	if o.SemverOptions.Debug {
		fmt.Fprintf(o.Out, "%s Debug: Working directory: %s\n", emoji.Information, o.SemverOptions.RootDir)
		fmt.Fprintf(o.Out, "%s Debug: Config directory: %s\n", emoji.Information, o.SemverOptions.ConfigDir)
		fmt.Fprintf(o.Out, "%s Debug: No stage: %t\n", emoji.Information, o.SemverOptions.NoStage)
		fmt.Fprintf(o.Out, "%s Debug: Silent mode: %t\n", emoji.Information, o.SemverOptions.Silent)
		if o.Prerelease != "" {
			fmt.Fprintf(o.Out, "%s Debug: Prerelease: %s\n", emoji.Information, o.Prerelease)
		}
	}
}

// ShouldSuppressOutput returns true if output should be suppressed based on options
func (o *BumpOptions) ShouldSuppressOutput() bool {
	return o.SemverOptions.Silent
}

// ShouldSkipStaging returns true if git staging should be skipped
func (o *BumpOptions) ShouldSkipStaging() bool {
	return o.SemverOptions.NoStage || o.SemverOptions.DryRun
}

// ShouldSkipCommit returns true if git commit should be skipped
func (o *BumpOptions) ShouldSkipCommit() bool {
	return o.SemverOptions.DryRun
}

// GetPrerelease returns the prerelease suffix if set
func (o *BumpOptions) GetPrerelease() string {
	return o.Prerelease
}
