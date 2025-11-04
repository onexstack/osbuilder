package semver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// CheckOptions defines configuration for the "check" subcommand.
type CheckOptions struct {
	// Command specific options
	ConfigFile string // specific config file to check
	Verbose    bool   // verbose output

	// semver configuration (inherited from parent semver command)
	SemverOptions *SemverOptions

	genericiooptions.IOStreams
}

var (
	checkLongDesc = templates.LongDesc(`Check if an OSBuilder configuration file is valid.

    This command validates the syntax and content of OSBuilder configuration files.
    It checks for proper YAML/JSON structure, validates configuration values,
    and ensures all required fields are present. The command supports various
    configuration file formats including .osbuilder.yml, .osbuilder.yaml,
    osbuilder.yml, and osbuilder.yaml.

    The validation includes:
    - File syntax validation (YAML/JSON)
    - Schema validation against OSBuilder configuration spec  
    - Required field validation
    - Value range and format validation
    - Cross-field dependency validation`)

	checkExample = templates.Examples(`# Check default configuration file in current directory
        osbuilder semver check

        # Check specific configuration file
        osbuilder semver check --config-file .osbuilder.yml

        # Check with verbose output to see all validation details
        osbuilder semver check --verbose

        # Check configuration file in specific directory
        osbuilder semver --config-dir ./config check

        # Check with debug information
        osbuilder semver --debug check --verbose`)
)

// NewCheckCmd creates the "check" subcommand.
func NewCheckCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams, semverOptions *SemverOptions) *cobra.Command {
	opts := &CheckOptions{
		IOStreams:     ioStreams,
		SemverOptions: semverOptions,
	}

	cmd := &cobra.Command{
		Use:                   "check",
		Short:                 "Check if a configuration file is valid",
		Long:                  checkLongDesc,
		Example:               checkExample,
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

	// Check-specific flags (only flags unique to this subcommand)
	cmd.Flags().StringVar(&opts.ConfigFile, "config-file", "", "Specific configuration file to check (overrides auto-detection)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output with detailed validation information")

	// Note: semver flags are inherited from parent semver command via persistent flags
	// No need to redefine --debug, --config-dir, etc.

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *CheckOptions) Complete() error {
	// Validate that semver options is provided
	if o.SemverOptions == nil {
		return fmt.Errorf("semver options is required")
	}

	// If no specific config file is provided, try to find one automatically
	if o.ConfigFile == "" {
		configFile := o.findConfigFile()
		if configFile != "" {
			o.ConfigFile = configFile
		}
	}

	// Convert relative path to absolute if config file is specified
	if o.ConfigFile != "" && !filepath.IsAbs(o.ConfigFile) {
		o.ConfigFile = filepath.Join(o.SemverOptions.RootDir, o.ConfigFile)
	}

	return nil
}

// Validate ensures provided inputs are valid.
func (o *CheckOptions) Validate() error {
	// Check if config file exists (if specified)
	if o.ConfigFile != "" {
		if _, err := os.Stat(o.ConfigFile); os.IsNotExist(err) {
			return fmt.Errorf("configuration file not found: %s", o.ConfigFile)
		}
	}

	// If no config file found at all, check if we're in the right directory
	if o.ConfigFile == "" && !o.hasAnyConfigFile() {
		return fmt.Errorf("no OSBuilder configuration file found in %s or %s",
			o.SemverOptions.RootDir, o.SemverOptions.ConfigDir)
	}

	return nil
}

// Run performs the check operation.
func (o *CheckOptions) Run() error {
	// Print what we're checking
	if !o.SemverOptions.Silent {
		fmt.Fprintf(o.Out, "Checking configuration file: %s\n", o.ConfigFile)
		if o.Verbose {
			fmt.Fprintf(o.Out, "Using configuration directory: %s\n", o.SemverOptions.ConfigDir)
		}
	}

	cfg, err := LoadFromFile(o.ConfigFile)
	if err != nil {
		return err
	}
	return cfg.Validate()
}

// findConfigFile searches for a configuration file and returns its path
func (o *CheckOptions) findConfigFile() string {
	searchDirs := []string{o.SemverOptions.RootDir}
	if o.SemverOptions.ConfigDir != "" && o.SemverOptions.ConfigDir != o.SemverOptions.RootDir {
		searchDirs = append(searchDirs, o.SemverOptions.ConfigDir)
	}

	for _, dir := range searchDirs {
		for _, fileName := range configFileNames {
			configPath := filepath.Join(dir, fileName)
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}
		}
	}

	return ""
}

// hasAnyConfigFile checks if any configuration file exists
func (o *CheckOptions) hasAnyConfigFile() bool {
	return o.findConfigFile() != ""
}
