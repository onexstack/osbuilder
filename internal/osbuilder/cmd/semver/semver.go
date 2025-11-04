package semver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log/handlers/discard"
	"github.com/enescakir/emoji"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// SemverOptions defines configuration for the "semver" root command.
type SemverOptions struct {
	RootDir string // working directory

	DryRun                   bool
	Debug                    bool
	NoPush                   bool
	NoStage                  bool
	Silent                   bool
	IgnoreExistingPrerelease bool
	FilterOnPrerelease       bool
	IgnoreDetached           bool
	IgnoreShallow            bool
	ConfigDir                string

	genericiooptions.IOStreams
}

var (
	semverLongDesc = templates.LongDesc(`Semantic versioning the easy way.

    A comprehensive tool for managing semantic versions in Git repositories.
    This tool helps you automate semantic versioning by analyzing conventional commits,
    generating tags, updating changelogs, and managing releases. It follows the
    Semantic Versioning 2.0.0 specification and Conventional Commits 1.0.0.

    The tool provides several subcommands to handle different aspects of semantic versioning:

    - tag: Create git tags with calculated semantic versions

    - bump: Update version numbers in files

    - release: Manage semantic releases

    - changelog: Generate and update changelogs

    - check: Validate configuration files`)

	semverExample = templates.Examples(`# Tag the repository with the next calculated semantic version
        osbuilder semver tag

        # Bump version in files
        osbuilder semver bump

        # Create or update changelog
        osbuilder semver changelog

        # Check configuration
        osbuilder semver check

        # Use semver options
        osbuilder semver tag --dry-run --debug
        osbuilder semver bump --silent --no-push`)
)

// NewSemverCmd creates the "semver" root command.
func NewSemverCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &SemverOptions{
		IOStreams: ioStreams,
		ConfigDir: ".", // default to current directory
	}

	cmd := &cobra.Command{
		Use:                   "semver",
		Short:                 "Semantic versioning the easy way",
		Long:                  semverLongDesc,
		Example:               semverExample,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: false,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize logging
			log.SetHandler(cli.Default)

			// Handle debug mode
			if opts.Debug {
				log.SetLevel(log.DebugLevel)
			}

			// Handle silent mode
			if opts.Silent {
				// Switch logging handler to discard all logging
				log.SetHandler(discard.Default)
				// Also redirect IOStreams.Out to discard for non-essential output
				// but preserve ErrOut for critical errors
			}

			// Validate conflicting options
			if opts.Debug && opts.Silent {
				fmt.Fprintf(opts.ErrOut, "Warning: --debug and --silent flags conflict, debug output will be suppressed\n")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Allow visiting child flags without requiring explicit subcommand.
	cmd.TraverseChildren = true

	// Persistent flags (available to all subcommands)
	pf := cmd.PersistentFlags()
	pf.StringVar(&opts.ConfigDir, "config-dir", opts.ConfigDir, "a custom path to a directory containing configuration")
	pf.BoolVar(&opts.DryRun, "dry-run", false, "run without making any changes")
	pf.BoolVar(&opts.Debug, "debug", false, "show me everything that happens")
	pf.BoolVar(&opts.NoPush, "no-push", false, "no changes will be pushed to the git remote")
	pf.BoolVar(&opts.NoStage, "no-stage", false, "no changes will be git staged")
	pf.BoolVar(&opts.Silent, "silent", false, "silence all logging")
	pf.BoolVar(&opts.IgnoreDetached, "ignore-detached", false, "ignore reported git detached HEAD error")
	pf.BoolVar(&opts.IgnoreShallow, "ignore-shallow", false, "ignore reported git shallow clone error")
	pf.BoolVar(&opts.IgnoreExistingPrerelease, "ignore-existing-prerelease", false, "ignore any existing prerelease when calculating next semantic version")
	pf.BoolVar(&opts.FilterOnPrerelease, "filter-on-prerelease", false, "filter tags that only match the provided prerelease when calculating next semantic version")

	// Add subcommands - pass semver options to each subcommand
	cmd.AddCommand(
		NewTagCmd(factory, ioStreams, opts),
		NewBumpCmd(factory, ioStreams, opts),
		NewReleaseCmd(factory, ioStreams, opts),
		NewChangelogCmd(factory, ioStreams, opts),
		NewCheckCmd(factory, ioStreams, opts),
		// version.NewVersionCmd(factory, ioStreams),
		// completion.NewCompletionCmd(factory, ioStreams),
		// manpage.NewManPageCmd(factory, ioStreams),
	)

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *SemverOptions) Complete() error {
	var err error
	o.RootDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Resolve config directory path
	if !filepath.IsAbs(o.ConfigDir) {
		o.ConfigDir = filepath.Join(o.RootDir, o.ConfigDir)
	}

	return nil
}

// Validate ensures provided inputs are valid.
func (o *SemverOptions) Validate() error {
	// Validate config directory exists or can be created
	if err := os.MkdirAll(o.ConfigDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", o.ConfigDir, err)
	}

	// Validate conflicting options
	if o.Debug && o.Silent {
		return fmt.Errorf("cannot use --debug and --silent flags together")
	}

	return nil
}

// Run performs the semver root operation.
func (o *SemverOptions) Run() error {
	// Complete and validate options
	if err := o.Complete(); err != nil {
		return err
	}

	if err := o.Validate(); err != nil {
		return err
	}

	o.PrintGettingStarted()
	return nil
}

// PrintGettingStarted prints formatted success information.
func (o *SemverOptions) PrintGettingStarted() {
	if o.Silent {
		return
	}

	fmt.Fprintf(o.Out, "%s Semantic versioning tool initialized\n", emoji.CheckMarkButton)
	fmt.Fprintf(o.Out, "Working directory: %s\n", o.RootDir)
	fmt.Fprintf(o.Out, "Config directory: %s\n", o.ConfigDir)

	if o.Debug {
		fmt.Fprintf(o.Out, "%s Debug mode enabled\n", emoji.Information)
	}

	if o.DryRun {
		fmt.Fprintf(o.Out, "%s Dry-run mode enabled\n", emoji.Information)
	}

	fmt.Fprintf(o.Out, "\nAvailable commands:\n")
	fmt.Fprintf(o.Out, "  tag        - Tag repository with semantic version\n")
	fmt.Fprintf(o.Out, "  bump       - Bump version in files\n")
	fmt.Fprintf(o.Out, "  release    - Manage semantic releases\n")
	fmt.Fprintf(o.Out, "  changelog  - Generate/update changelogs\n")
	fmt.Fprintf(o.Out, "  check      - Validate configuration\n")
	fmt.Fprintf(o.Out, "  version    - Show version information\n")
	fmt.Fprintf(o.Out, "\nUse 'semver <command> --help' for more information about a command.\n")
}
