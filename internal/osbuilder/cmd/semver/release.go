package semver

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/changelog"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/fetchtag"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gitcheck"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gitcommit"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gittag"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/gpgimport"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/after"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/afterbump"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/afterchangelog"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/aftertag"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/before"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/beforebump"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/beforechangelog"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/hook/beforetag"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/nextcommit"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/nextsemver"
	"github.com/onexstack/osbuilder/internal/osbuilder/semver/task/scm"
)

// ReleaseOptions defines configuration for the "release" subcommand.
type ReleaseOptions struct {
	// Release-specific options
	FetchTags      bool     // fetch all tags from remote repository
	Check          bool     // check if a release will be triggered
	Prerelease     string   // prerelease suffix
	SkipChangelog  bool     // skip changelog creation or amendment
	SkipBumps      bool     // skip bumping of any files
	NoPrefix       bool     // strip 'v' prefix from semantic version
	Exclude        []string // regexes for excluding conventional commits from changelog
	Include        []string // regexes to cherry-pick conventional commits for changelog
	Sort           string   // sort order of commits within changelog entry
	Multiline      bool     // include multiline commit messages in changelog
	SkipPrerelease bool     // skip changelog entry for prerelease
	TrimHeader     bool     // strip lines preceding conventional commit type

	// New release asset and publishing options
	Assets   []string // paths to release asset files to upload
	Draft    bool     // create a draft release instead of publishing immediately
	Template string   // path to custom release notes template file

	// semver configuration (inherited from parent semver command)
	SemverOptions *SemverOptions

	genericiooptions.IOStreams
}

var (
	releaseLongDesc = templates.LongDesc(`Release the next semantic version of your git repository. A release consists of
    a three-stage process. First, all configured files will be bumped (patched) using
    the next semantic version. Second, a changelog entry containing all commits for
    the latest semantic release will be created. Finally, OSBuilder will tag the
    repository. 
    
    OSBuilder automatically handles the staging and pushing of modified files and the 
    tagging of the repository with two separate git pushes. But this behavior can be 
    disabled to manage these actions manually.

    Parts of this release process can be disabled if needed. You can skip file bumping,
    changelog generation, or specific steps in the release pipeline to customize the
    process according to your needs.

    Release assets can be specified using the --assets flag to upload files alongside
    the release. The --draft flag allows creating draft releases for review before
    publishing. Custom release notes templates can be provided using the --template flag.`)

	releaseExample = templates.Examples(`# Release the next semantic version
        osbuilder semver release

        # Release the next semantic version without bumping any files
        osbuilder semver release --skip-bumps

        # Release the next semantic version without generating a changelog
        osbuilder semver release --skip-changelog

        # Append a prerelease suffix to the next calculated semantic version
        osbuilder semver release --prerelease beta.1

        # Ensure any "v" prefix is stripped from the next calculated semantic version
        # to explicitly adhere to the SemVer specification
        osbuilder semver release --no-prefix

        # Check if a release will be triggered without actually performing it
        osbuilder semver release --check

        # Fetch all tags from remote before releasing
        osbuilder semver release --fetch-all

        # Release with custom changelog options
        osbuilder semver release --exclude "^ci,^chore" --multiline

        # Create a draft release with assets
        osbuilder semver release --draft --assets dist/binary-linux-amd64 --assets dist/binary-windows.exe

        # Release with custom release notes template
        osbuilder semver release --template .github/release-template.md

        # Release with multiple assets and custom template
        osbuilder semver release --assets "dist/*.tar.gz" --assets "dist/*.zip" --template release-notes.tmpl

        # Dry run to see what would happen
        osbuilder semver --dry-run release`)
)

var (
	// Pipeline for full release process
	releaseFullPipeline = []task.Runner{
		gitcheck.Task{},
		before.Task{},
		gpgimport.Task{},
		scm.Task{},
		fetchtag.Task{},
		nextsemver.Task{},
		nextcommit.Task{},
		beforebump.Task{},
		bump.Task{},
		afterbump.Task{},
		beforechangelog.Task{},
		changelog.Task{},
		afterchangelog.Task{},
		gitcommit.Task{},
		beforetag.Task{},
		gittag.Task{},
		aftertag.Task{},
		after.Task{},
	}

	// Pipeline for release check (minimal pipeline)
	releaseCheckPipeline = []task.Runner{
		nextsemver.Task{},
	}
)

// NewReleaseCmd creates the "release" subcommand.
func NewReleaseCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams, semverOptions *SemverOptions) *cobra.Command {
	opts := &ReleaseOptions{
		IOStreams:     ioStreams,
		SemverOptions: semverOptions,
		Sort:          "",         // Default sort order
		Assets:        []string{}, // Default to no assets
		Draft:         false,      // Default to immediate release
		Template:      "",         // Default to no custom template
	}

	cmd := &cobra.Command{
		Use:                   "release",
		Short:                 "Release the next semantic version of a repository",
		Long:                  releaseLongDesc,
		Example:               releaseExample,
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

	// Release-specific flags (only flags unique to this subcommand)
	cmd.Flags().BoolVar(&opts.FetchTags, "fetch-all", false, "Fetch all tags from the remote repository")
	cmd.Flags().BoolVar(&opts.Check, "check", false, "Check if a release will be triggered")
	cmd.Flags().StringVar(&opts.Prerelease, "prerelease", "", "Append a prerelease suffix to next calculated semantic version")
	cmd.Flags().BoolVar(&opts.SkipChangelog, "skip-changelog", false, "Skip the creation or amendment of a changelog")
	cmd.Flags().BoolVar(&opts.SkipBumps, "skip-bumps", false, "Skip the bumping of any files")
	cmd.Flags().BoolVar(&opts.NoPrefix, "no-prefix", false, "Strip the default 'v' prefix from the next calculated semantic version")
	cmd.Flags().StringSliceVar(&opts.Exclude, "exclude", []string{}, "A list of regexes for excluding conventional commits from the changelog")
	cmd.Flags().StringSliceVar(&opts.Include, "include", []string{}, "A list of regexes to cherry-pick conventional commits for the changelog")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "The sort order of commits within each changelog entry (asc/desc)")
	cmd.Flags().BoolVar(&opts.Multiline, "multiline", false, "Include multiline commit messages within changelog (skips truncation)")
	cmd.Flags().BoolVar(&opts.SkipPrerelease, "skip-changelog-prerelease", false, "Skip the creation of a changelog entry for a prerelease")
	cmd.Flags().BoolVar(&opts.TrimHeader, "trim-header", false, "Strip any lines preceding the conventional commit type in the commit message")

	// New asset and publishing flags
	cmd.Flags().StringSliceVar(&opts.Assets, "assets", []string{}, "Paths or glob patterns for release asset files to upload")
	cmd.Flags().BoolVar(&opts.Draft, "draft", false, "Create a draft release instead of publishing immediately")
	cmd.Flags().StringVar(&opts.Template, "template", "", "Path to custom release notes template file")

	// Note: semver flags are inherited from parent semver command via persistent flags

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *ReleaseOptions) Complete() error {
	// Validate that semver options is provided
	if o.SemverOptions == nil {
		return fmt.Errorf("semver options is required")
	}

	// Validate sort order
	if o.Sort != "" && o.Sort != "asc" && o.Sort != "desc" {
		return fmt.Errorf("invalid sort order %q, valid values are: asc, desc", o.Sort)
	}

	// Expand asset file paths and validate they exist
	if len(o.Assets) > 0 {
		expandedAssets, err := o.expandAssetPaths()
		if err != nil {
			return fmt.Errorf("failed to expand asset paths: %w", err)
		}
		o.Assets = expandedAssets
	}

	return nil
}

// expandAssetPaths expands glob patterns and validates asset files exist
func (o *ReleaseOptions) expandAssetPaths() ([]string, error) {
	var expandedPaths []string
	baseDir := o.SemverOptions.RootDir

	for _, assetPath := range o.Assets {
		// Handle relative paths
		fullPath := assetPath
		if !filepath.IsAbs(assetPath) {
			fullPath = filepath.Join(baseDir, assetPath)
		}

		// Expand glob patterns
		matches, err := filepath.Glob(fullPath)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", assetPath, err)
		}

		if len(matches) == 0 {
			// If no glob matches, check if it's a regular file
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("asset file not found: %s", assetPath)
			}
			expandedPaths = append(expandedPaths, fullPath)
		} else {
			// Add all glob matches
			for _, match := range matches {
				// Verify each match is a regular file (not a directory)
				if info, err := os.Stat(match); err != nil {
					return nil, fmt.Errorf("cannot access asset file %s: %w", match, err)
				} else if info.IsDir() {
					return nil, fmt.Errorf("asset path %s is a directory, expected a file", match)
				}
				expandedPaths = append(expandedPaths, match)
			}
		}
	}

	return expandedPaths, nil
}

// Validate ensures provided inputs are valid.
func (o *ReleaseOptions) Validate() error {
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

	// Validate include/exclude patterns (basic validation)
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

	// Validate template file exists if provided
	if o.Template != "" {
		templatePath := o.Template
		if !filepath.IsAbs(templatePath) {
			templatePath = filepath.Join(o.SemverOptions.RootDir, templatePath)
		}

		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("template file not found: %s", o.Template)
		} else if err != nil {
			return fmt.Errorf("cannot access template file %s: %w", o.Template, err)
		}
	}

	// Validate asset files (already expanded in Complete())
	for _, assetPath := range o.Assets {
		if info, err := os.Stat(assetPath); os.IsNotExist(err) {
			return fmt.Errorf("asset file not found: %s", assetPath)
		} else if err != nil {
			return fmt.Errorf("cannot access asset file %s: %w", assetPath, err)
		} else if info.IsDir() {
			return fmt.Errorf("asset path %s is a directory, expected a file", assetPath)
		}
	}

	// Validate logical constraints
	if o.Check && (len(o.Assets) > 0 || o.Draft || o.Template != "") {
		return fmt.Errorf("--check mode cannot be used with --assets, --draft, or --template flags")
	}

	return nil
}

// Run performs the release operation.
func (o *ReleaseOptions) Run() error {
	if o.Check {
		return o.executeReleaseCheck()
	}
	return o.executeReleaseProcess()
}

// executeReleaseProcess runs the full release pipeline
func (o *ReleaseOptions) executeReleaseProcess() error {
	ctx, err := o.setupReleaseContext()
	if err != nil {
		return fmt.Errorf("failed to setup context: %w", err)
	}

	if err := task.Execute(ctx, releaseFullPipeline); err != nil {
		return fmt.Errorf("failed to execute release pipeline: %w", err)
	}

	o.PrintGettingStarted()
	return nil
}

// executeReleaseCheck runs a minimal check to see if a release would be triggered
func (o *ReleaseOptions) executeReleaseCheck() error {
	ctx, err := o.setupReleaseContext()
	if err != nil {
		return fmt.Errorf("failed to setup context: %w", err)
	}

	if err := task.Execute(ctx, releaseCheckPipeline); err != nil {
		return fmt.Errorf("failed to execute release check: %w", err)
	}

	if ctx.NoVersionChanged {
		return errors.New("no release detected")
	}

	o.PrintReleaseCheck()
	return nil
}

// setupReleaseContext creates and configures the context for release operations
func (o *ReleaseOptions) setupReleaseContext() (*context.Context, error) {
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

	// Set release-specific options
	ctx.FetchTags = o.FetchTags
	ctx.SkipChangelog = o.SkipChangelog
	ctx.SkipBumps = o.SkipBumps
	ctx.NoPrefix = o.NoPrefix

	// Set new release asset and publishing options
	ctx.Assets = o.Assets
	ctx.Draft = o.Draft
	ctx.Template = o.Template

	// Enable pre-tagging support for generating a changelog
	ctx.Changelog.PreTag = true

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

	// Merge config and command line arguments for changelog
	ctx.Changelog.Multiline = o.Multiline
	if !ctx.Changelog.Multiline && ctx.Config.Changelog != nil {
		ctx.Changelog.Multiline = ctx.Config.Changelog.Multiline
	}

	ctx.Changelog.Include = o.Include
	ctx.Changelog.Exclude = o.Exclude
	if ctx.Config.Changelog != nil {
		ctx.Changelog.Include = append(ctx.Changelog.Include, ctx.Config.Changelog.Include...)
		ctx.Changelog.Exclude = append(ctx.Changelog.Exclude, ctx.Config.Changelog.Exclude...)
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

	// By default ensure the ci(osbuilder): commits are excluded
	ctx.Changelog.Exclude = append(ctx.Changelog.Exclude, "ci(osbuilder):")

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
func (o *ReleaseOptions) loadConfig() (config.Uplift, error) {
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
func (o *ReleaseOptions) PrintGettingStarted() {
	// Don't print anything if silent mode is enabled
	if o.SemverOptions.Silent {
		return
	}

	if o.SemverOptions.DryRun {
		fmt.Fprintf(o.Out, "%s Release operation completed (dry-run mode)\n", emoji.CheckMarkButton)
	} else if o.Draft {
		fmt.Fprintf(o.Out, "%s Successfully created draft release\n", emoji.Memo)
	} else {
		fmt.Fprintf(o.Out, "%s Successfully released next semantic version\n", emoji.CheckMarkButton)
	}

	// Print asset information
	if len(o.Assets) > 0 {
		fmt.Fprintf(o.Out, "%s Release assets (%d):\n", emoji.Package, len(o.Assets))
		for _, asset := range o.Assets {
			// Show relative path for cleaner output
			relPath, err := filepath.Rel(o.SemverOptions.RootDir, asset)
			if err != nil {
				relPath = asset
			}

			// Get file size for display
			if info, err := os.Stat(asset); err == nil {
				size := formatFileSize(info.Size())
				fmt.Fprintf(o.Out, "  • %s (%s)\n", relPath, size)
			} else {
				fmt.Fprintf(o.Out, "  • %s\n", relPath)
			}
		}
	}

	// Print template information
	if o.Template != "" {
		relPath, err := filepath.Rel(o.SemverOptions.RootDir, o.Template)
		if err != nil {
			relPath = o.Template
		}
		fmt.Fprintf(o.Out, "%s Using release notes template: %s\n", emoji.Memo, relPath)
	}

	if o.SemverOptions.Debug {
		fmt.Fprintf(o.Out, "%s Debug: Working directory: %s\n", emoji.Information, o.SemverOptions.RootDir)
		fmt.Fprintf(o.Out, "%s Debug: Config directory: %s\n", emoji.Information, o.SemverOptions.ConfigDir)
		fmt.Fprintf(o.Out, "%s Debug: Fetch tags: %t\n", emoji.Information, o.FetchTags)
		fmt.Fprintf(o.Out, "%s Debug: Skip changelog: %t\n", emoji.Information, o.SkipChangelog)
		fmt.Fprintf(o.Out, "%s Debug: Skip bumps: %t\n", emoji.Information, o.SkipBumps)
		fmt.Fprintf(o.Out, "%s Debug: No prefix: %t\n", emoji.Information, o.NoPrefix)
		fmt.Fprintf(o.Out, "%s Debug: Draft release: %t\n", emoji.Information, o.Draft)
		fmt.Fprintf(o.Out, "%s Debug: Assets count: %d\n", emoji.Information, len(o.Assets))
		if o.Prerelease != "" {
			fmt.Fprintf(o.Out, "%s Debug: Prerelease: %s\n", emoji.Information, o.Prerelease)
		}
		if o.Template != "" {
			fmt.Fprintf(o.Out, "%s Debug: Template: %s\n", emoji.Information, o.Template)
		}
		if len(o.Include) > 0 {
			fmt.Fprintf(o.Out, "%s Debug: Include patterns: %v\n", emoji.Information, o.Include)
		}
		if len(o.Exclude) > 0 {
			fmt.Fprintf(o.Out, "%s Debug: Exclude patterns: %v\n", emoji.Information, o.Exclude)
		}
	}
}

// PrintReleaseCheck prints release check results
func (o *ReleaseOptions) PrintReleaseCheck() {
	if o.SemverOptions.Silent {
		return
	}

	fmt.Fprintf(o.Out, "%s Release will be triggered\n", emoji.CheckMarkButton)
}

// formatFileSize formats file size in human readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ShouldSuppressOutput returns true if output should be suppressed based on options
func (o *ReleaseOptions) ShouldSuppressOutput() bool {
	return o.SemverOptions.Silent
}

// ShouldSkipStaging returns true if git staging should be skipped
func (o *ReleaseOptions) ShouldSkipStaging() bool {
	return o.SemverOptions.NoStage || o.SemverOptions.DryRun
}

// ShouldSkipCommit returns true if git commit should be skipped
func (o *ReleaseOptions) ShouldSkipCommit() bool {
	return o.SemverOptions.DryRun
}

// IsCheckMode returns true if only checking if release would be triggered
func (o *ReleaseOptions) IsCheckMode() bool {
	return o.Check
}

// GetPrerelease returns the prerelease suffix if set
func (o *ReleaseOptions) GetPrerelease() string {
	return o.Prerelease
}

// GetIncludePatterns returns the include patterns
func (o *ReleaseOptions) GetIncludePatterns() []string {
	return o.Include
}

// GetExcludePatterns returns the exclude patterns
func (o *ReleaseOptions) GetExcludePatterns() []string {
	return o.Exclude
}

// GetAssets returns the list of asset file paths
func (o *ReleaseOptions) GetAssets() []string {
	return o.Assets
}

// IsDraft returns true if this should be a draft release
func (o *ReleaseOptions) IsDraft() bool {
	return o.Draft
}

// GetTemplate returns the release notes template path
func (o *ReleaseOptions) GetTemplate() string {
	return o.Template
}

// HasAssets returns true if assets are specified
func (o *ReleaseOptions) HasAssets() bool {
	return len(o.Assets) > 0
}

// HasTemplate returns true if a custom template is specified
func (o *ReleaseOptions) HasTemplate() bool {
	return o.Template != ""
}
