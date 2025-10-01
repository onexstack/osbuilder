package app

import (
	"fmt"
	"os"
	"path/filepath"
	"context"

	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	genericapiserver "k8s.io/apiserver/pkg/server"
	logsv1 "k8s.io/component-base/logs/api/v1"

	"{{.D.ModuleName}}/cmd/{{.Web.BinaryName}}/app/options"
)

const (
	// defaultHomeDir defines the default directory to store the configuration for the {{.Web.BinaryName}} service.
	defaultHomeDir = ".{{.D.ProjectName}}"

	// defaultConfigName specifies the default configuration file name for the {{.Web.BinaryName}} service.
	defaultConfigName = "{{.Web.BinaryName}}.yaml"
)

// Path to the configuration file
var configFile string

// NewWebServerCommand creates a *cobra.Command object used to start the application.
func NewWebServerCommand() *cobra.Command {
	// Create default application command-line options
	opts := options.NewServerOptions()

	cmd := &cobra.Command{
		// Specify the name of the command, which will appear in the help information
		Use: "{{.Web.BinaryName}}",
		// A short description of the command
		Short: "TODO: Update the short description of the binary file.",
		// A detailed description of the command
		Long: `TODO: Update the detailed description of the binary file.`,
		// Do not print help information when the command encounters an error.
		// Setting this to true ensures that errors are immediately visible.
		SilenceUsage: true,
		// Specify the Run function to execute when cmd.Execute() is called
		RunE: func(cmd *cobra.Command, args []string) error {
			// If the --version flag is passed, print version information and exit
			version.PrintAndExitIfRequested()

			// Unmarshal the configuration from viper into opts
			if err := viper.Unmarshal(opts); err != nil {
				return fmt.Errorf("failed to unmarshal configuration: %w", err)
			}

			if err := logsv1.ValidateAndApply(opts.LogOptions.Native(), nil); err != nil {
				return err
			}

			// Validate command-line options
			if err := opts.Validate(); err != nil {
				return fmt.Errorf("invalid options: %w", err)
			}

			ctx := genericapiserver.SetupSignalContext()

			return run(ctx, opts)
		},
		// Set argument validation for the command. No command-line arguments are required.
		// For example: ./{{.Web.BinaryName}} param1 param2
		Args: cobra.NoArgs,
	}

	// Initialize configuration function, called when each command runs
	cobra.OnInitialize(core.OnInitialize(&configFile, "{{.Web.EnvironmentPrefix}}", searchDirs(), defaultConfigName))

	// cobra supports persistent flags, which apply to the assigned command and all its subcommands.
	// It is recommended to use configuration files for application configuration to make it easier to manage configuration items.
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", filePath(), "Path to the {{.Web.BinaryName}} configuration file.")

	// Add server options as flags
	opts.AddFlags(cmd.PersistentFlags())

	// Add the --version flag
	version.AddFlags(cmd.PersistentFlags())

	return cmd
}

// run contains the main logic for initializing and running the server.
func run(ctx context.Context, opts *options.ServerOptions) error {
	// Retrieve application configuration
	// Separating command-line options and application configuration allows more flexible handling of these two types of configurations.
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and start the server
	server, err := cfg.NewServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Run the server
	return server.Run(ctx)
}

// searchDirs returns the default directories to search for the configuration file.
func searchDirs() []string {
	// Get the user's home directory.
	homeDir, err := os.UserHomeDir()
	// If unable to get the user's home directory, print an error message and exit the program.
	cobra.CheckErr(err)
	return []string{filepath.Join(homeDir, defaultHomeDir), "."}
}

// filePath retrieves the full path to the default configuration file.
func filePath() string {
	home, err := os.UserHomeDir()
	// If the user's home directory cannot be retrieved, log an error and return an empty path.
	cobra.CheckErr(err)
	return filepath.Join(home, defaultHomeDir, defaultConfigName)
}
