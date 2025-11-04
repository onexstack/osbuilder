package semver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enescakir/emoji"
	git "github.com/purpleclay/gitz"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/config"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/context"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/changelog"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gitcheck"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gitcommit"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/after"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/afterchangelog"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/before"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/beforechangelog"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/nextcommit"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/scm"
)

// ChangelogOptions defines configuration for the "changelog" subcommand.
type ChangelogOptions struct {
	// Changelog-specific options
	DiffOnly       bool     // output the changelog diff only
	Exclude        []string // regexes for excluding conventional commits
	Include        []string // regexes to cherry-pick conventional commits
	All            bool     // generate changelog from entire history
	Sort           string   // sort order of commits within changelog entry
	Multiline      bool     // include multiline commit messages
	SkipPrerelease bool     // skip changelog entry for prerelease
	TrimHeader     bool     // strip lines preceding conventional commit type

	// semver configuration (inherited from parent semver command)
	SemverOptions *SemverOptions

	genericiooptions.IOStreams
}

var (
	changelogLongDesc = templates.LongDesc(`Scans the git log for the latest semantic release and generates a changelog
    entry. If this is a first release, all commits between the last release (or
    identifiable tag) and the repository trunk will be written to the changelog.
    
    Any subsequent entry within the changelog will only contain commits between
    the latest set of tags. Basic customization is supported. Optionally commits
    can be explicitly included or excluded from the entry and sorted in ascending
    or descending order. OSBuilder automatically handles the staging and pushing of
    changes to the CHANGELOG.md file to the git remote, but this behavior can be
    disabled, to manage this action manually.

    OSBuilder bases its changelog format on the Keep a Changelog specification:
    https://keepachangelog.com/en/1.0.0/`)

	changelogExample = templates.Examples(`# Generate the next changelog entry for the latest semantic release
        osbuilder semver changelog

        # Generate a changelog for the entire history of the repository
        osbuilder semver changelog --all

        # Generate the next changelog entry and write it to stdout
        osbuilder semver changelog --diff-only

        # Generate the next changelog entry by excluding any conventional commits
        # with the ci, chore or test prefixes
        osbuilder semver changelog --exclude "^ci,^chore,^test"

        # Generate the next changelog entry with commits that only include the
        # following scope
        osbuilder semver changelog --include "^.*$scope$"

        # Generate the next changelog entry but do not stage or push any changes
        # back to the git remote
        osbuilder semver --no-stage changelog

        # Generate a changelog with multiline commit messages
        osbuilder semver changelog --multiline

        # Generate a changelog trimming any lines preceding the conventional commit type
        osbuilder semver changelog --trim-header

        # Generate a changelog with prerelease tags being skipped
        osbuilder semver changelog --skip-prerelease

        # Sort commits in ascending order
        osbuilder semver changelog --sort asc`)
)

var (
	// Pipeline for full changelog generation
	changelogFullPipeline = []task.Runner{
		gitcheck.Task{},
		before.Task{},
		scm.Task{},
		nextcommit.Task{},
		beforechangelog.Task{},
		changelog.Task{},
		afterchangelog.Task{},
		gitcommit.Task{},
		after.Task{},
	}

	// Pipeline for diff-only changelog generation
	changelogDiffPipeline = []task.Runner{
		gitcheck.Task{},
		before.Task{},
		scm.Task{},
		changelog.Task{},
		after.Task{},
	}
)

// NewChangelogCmd creates the "changelog" subcommand.
func NewChangelogCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams, semverOptions *SemverOptions) *cobra.Command {
	opts := &ChangelogOptions{
		IOStreams:     ioStreams,
		SemverOptions: semverOptions,
		Sort:          "", // Default sort order
	}

	cmd := &cobra.Command{
		Use:                   "changelog",
		Short:                 "Create or update a changelog with the latest semantic release",
		Long:                  changelogLongDesc,
		Example:               changelogExample,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Always lowercase sort
			opts.Sort = strings.ToLower(opts.Sort)
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}

	// Changelog-specific flags (only flags unique to this subcommand)
	cmd.Flags().BoolVar(&opts.DiffOnly, "diff-only", false, "Output the changelog diff only")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Generate a changelog from the entire history of this repository")
	cmd.Flags().StringSliceVar(&opts.Exclude, "exclude", []string{}, "A list of regexes for excluding conventional commits from the changelog")
	cmd.Flags().StringSliceVar(&opts.Include, "include", []string{}, "A list of regexes to cherry-pick conventional commits for the changelog")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "The sort order of commits within each changelog entry (asc/desc)")
	cmd.Flags().BoolVar(&opts.Multiline, "multiline", false, "Include multiline commit messages within changelog (skips truncation)")
	cmd.Flags().BoolVar(&opts.SkipPrerelease, "skip-prerelease", false, "Skip the creation of a changelog entry for a prerelease")
	cmd.Flags().BoolVar(&opts.TrimHeader, "trim-header", false, "Strip any lines preceding the conventional commit type in the commit message")

	// Note: semver flags are inherited from parent semver command via persistent flags
	// No need to redefine --dry-run, --debug, --silent, etc.

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *ChangelogOptions) Complete() error {
	// Validate that semver options is provided
	if o.SemverOptions == nil {
		return fmt.Errorf("semver options is required")
	}

	// Validate sort order
	if o.Sort != "" && o.Sort != "asc" && o.Sort != "desc" {
		return fmt.Errorf("invalid sort order %q, valid values are: asc, desc", o.Sort)
	}

	return nil
}

// Validate ensures provided inputs are valid.
func (o *ChangelogOptions) Validate() error {
	// Check if we're in a git repository
	if _, err := os.Stat(filepath.Join(o.SemverOptions.RootDir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (or any of the parent directories): %s", o.SemverOptions.RootDir)
	}

	// Validate conflicting options
	if o.SemverOptions.Silent && o.SemverOptions.Debug {
		return fmt.Errorf("cannot use --silent and --debug flags together")
	}

	// Validate include/exclude patterns (basic regex validation)
	for _, pattern := range o.Include {
		if strings.TrimSpace(pattern) == "" {
			return fmt.Errorf("include pattern cannot be empty")
		}
	}

	for _, pattern := range o.Exclude {
		if strings.TrimSpace(pattern) == "" {
			return fmt.Errorf("exclude pattern cannot be empty")
		}
	}

	return nil
}

// Run performs the changelog operation.
func (o *ChangelogOptions) Run() error {
	if o.DiffOnly {
		return o.executeChangelogDiffPipeline()
	}
	return o.executeChangelogFullPipeline()
}

// executeChangelogFullPipeline runs the full changelog generation pipeline
func (o *ChangelogOptions) executeChangelogFullPipeline() error {
	ctx, err := o.setupChangelogContext()
	if err != nil {
		return fmt.Errorf("failed to setup context: %w", err)
	}

	if err := task.Execute(ctx, changelogFullPipeline); err != nil {
		return fmt.Errorf("failed to execute changelog pipeline: %w", err)
	}

	o.PrintGettingStarted()
	return nil
}

// executeChangelogDiffPipeline runs the condensed changelog diff pipeline
func (o *ChangelogOptions) executeChangelogDiffPipeline() error {
	ctx, err := o.setupChangelogContext()
	if err != nil {
		return fmt.Errorf("failed to setup context: %w", err)
	}

	if err := task.Execute(ctx, changelogDiffPipeline); err != nil {
		return fmt.Errorf("failed to execute changelog diff pipeline: %w", err)
	}

	o.PrintGettingStarted()
	return nil
}

// setupChangelogContext creates and configures the context for changelog operations
func (o *ChangelogOptions) setupChangelogContext() (*context.Context, error) {
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

	// Set changelog-specific options
	ctx.Changelog.DiffOnly = o.DiffOnly
	ctx.Changelog.All = o.All
	ctx.Changelog.Multiline = o.Multiline
	if !ctx.Changelog.Multiline && ctx.Config.Changelog != nil {
		ctx.Changelog.Multiline = ctx.Config.Changelog.Multiline
	}

	ctx.Changelog.SkipPrerelease = o.SkipPrerelease
	if !ctx.Changelog.SkipPrerelease && ctx.Config.Changelog != nil {
		ctx.Changelog.SkipPrerelease = ctx.Config.Changelog.SkipPrerelease
	}

	ctx.Changelog.TrimHeader = o.TrimHeader
	if !ctx.Changelog.TrimHeader && ctx.Config.Changelog != nil {
		ctx.Changelog.TrimHeader = ctx.Config.Changelog.TrimHeader
	}

	// Sort order provided as a command-line flag takes precedence
	ctx.Changelog.Sort = o.Sort
	if ctx.Changelog.Sort == "" && cfg.Changelog != nil {
		ctx.Changelog.Sort = strings.ToLower(cfg.Changelog.Sort)
	}

	// Merge config and command line arguments together
	ctx.Changelog.Include = o.Include
	ctx.Changelog.Exclude = o.Exclude
	if ctx.Config.Changelog != nil {
		ctx.Changelog.Include = append(ctx.Changelog.Include, ctx.Config.Changelog.Include...)
		ctx.Changelog.Exclude = append(ctx.Changelog.Exclude, ctx.Config.Changelog.Exclude...)
	}

	// By default ensure the ci(osbuilder): commits are excluded
	ctx.Changelog.Exclude = append(ctx.Changelog.Exclude, `ci$osbuilder$`)

	// Handle version tags if not generating full history
	if !ctx.Changelog.All {
		if err := o.setupVersionTags(ctx); err != nil {
			return nil, fmt.Errorf("failed to setup version tags: %w", err)
		}
	}

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

// setupVersionTags retrieves and sets up version tags for changelog generation
func (o *ChangelogOptions) setupVersionTags(ctx *context.Context) error {
	// Attempt to retrieve the latest 2 tags for generating a changelog entry
	tags, err := ctx.GitClient.Tags(git.WithShellGlob("*.*.*"),
		git.WithSortBy(git.CreatorDateDesc, git.VersionDesc))
	if err != nil {
		return fmt.Errorf("failed to retrieve git tags: %w", err)
	}

	if len(tags) == 1 {
		ctx.NextVersion.Raw = tags[0]
	} else if len(tags) > 1 {
		ctx.NextVersion.Raw = tags[0]
		ctx.CurrentVersion.Raw = tags[1]
	}

	return nil
}

// loadConfig loads configuration from the specified directory
func (o *ChangelogOptions) loadConfig() (config.Uplift, error) {
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
func (o *ChangelogOptions) PrintGettingStarted() {
	// Don't print anything if silent mode is enabled
	if o.SemverOptions.Silent {
		return
	}

	if o.DiffOnly {
		fmt.Fprintf(o.Out, "%s Changelog diff generated successfully\n", emoji.CheckMarkButton)
	} else if o.SemverOptions.DryRun {
		fmt.Fprintf(o.Out, "%s Changelog operation completed (dry-run mode)\n", emoji.CheckMarkButton)
	} else {
		fmt.Fprintf(o.Out, "%s Successfully generated changelog entry\n", emoji.CheckMarkButton)
	}

	if o.SemverOptions.Debug {
		fmt.Fprintf(o.Out, "%s Debug: Working directory: %s\n", emoji.Information, o.SemverOptions.RootDir)
		fmt.Fprintf(o.Out, "%s Debug: Config directory: %s\n", emoji.Information, o.SemverOptions.ConfigDir)
		fmt.Fprintf(o.Out, "%s Debug: Diff only: %t\n", emoji.Information, o.DiffOnly)
		fmt.Fprintf(o.Out, "%s Debug: Generate all: %t\n", emoji.Information, o.All)
		fmt.Fprintf(o.Out, "%s Debug: Multiline: %t\n", emoji.Information, o.Multiline)
		fmt.Fprintf(o.Out, "%s Debug: Skip prerelease: %t\n", emoji.Information, o.SkipPrerelease)
		fmt.Fprintf(o.Out, "%s Debug: Trim header: %t\n", emoji.Information, o.TrimHeader)
		if o.Sort != "" {
			fmt.Fprintf(o.Out, "%s Debug: Sort order: %s\n", emoji.Information, o.Sort)
		}
		if len(o.Include) > 0 {
			fmt.Fprintf(o.Out, "%s Debug: Include patterns: %v\n", emoji.Information, o.Include)
		}
		if len(o.Exclude) > 0 {
			fmt.Fprintf(o.Out, "%s Debug: Exclude patterns: %v\n", emoji.Information, o.Exclude)
		}
	}
}

// ShouldSuppressOutput returns true if output should be suppressed based on options
func (o *ChangelogOptions) ShouldSuppressOutput() bool {
	return o.SemverOptions.Silent
}

// ShouldSkipStaging returns true if git staging should be skipped
func (o *ChangelogOptions) ShouldSkipStaging() bool {
	return o.SemverOptions.NoStage || o.SemverOptions.DryRun
}

// ShouldSkipCommit returns true if git commit should be skipped
func (o *ChangelogOptions) ShouldSkipCommit() bool {
	return o.SemverOptions.DryRun
}

// IsDiffOnly returns true if only diff output is requested
func (o *ChangelogOptions) IsDiffOnly() bool {
	return o.DiffOnly
}

// IsGenerateAll returns true if full history changelog is requested
func (o *ChangelogOptions) IsGenerateAll() bool {
	return o.All
}

// GetIncludePatterns returns the include patterns
func (o *ChangelogOptions) GetIncludePatterns() []string {
	return o.Include
}

// GetExcludePatterns returns the exclude patterns
func (o *ChangelogOptions) GetExcludePatterns() []string {
	return o.Exclude
}

// GetSortOrder returns the sort order
func (o *ChangelogOptions) GetSortOrder() string {
	return o.Sort
}
