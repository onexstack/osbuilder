package upgrade

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

const (
	// GitHub API URL for latest release
	githubAPIURL = "https://api.github.com/repos/onexstack/osbuilder/releases/latest"
	// Go module path
	modulePath = "github.com/onexstack/osbuilder/cmd/osbuilder"
	// Current version (this should be set during build)
	currentVersion = "v1.0.0" // TODO: Replace with actual version from build
)

// UpgradeOptions defines configuration for the "upgrade" command.
type UpgradeOptions struct {
	// Upgrade specific options
	CheckOnly   bool   // only check for updates, don't install
	Version     string // specific version to upgrade to
	Force       bool   // force upgrade even if already latest
	Source      string // upgrade source (official, mirror, etc.)
	AutoConfirm bool   // skip confirmation prompts
	Timeout     int    // timeout for download operations (seconds)

	// Internal state
	currentBinary string // path to current osbuilder binary

	genericiooptions.IOStreams
}

// VersionInfo represents version information
type VersionInfo struct {
	Current string
	Latest  string
	IsNewer bool
}

var (
	upgradeLongDesc = templates.LongDesc(`
        Upgrade osbuilder to the latest version or a specific version.
        
        This command will check for the latest version of osbuilder from GitHub
        releases and upgrade the tool using 'go install'. It supports various
        upgrade modes including checking for updates, upgrading to specific
        versions, and automatic confirmations.
        
        The upgrade process:
        1. Check current version
        2. Fetch latest version from GitHub
        3. Compare versions
        4. Download and install new version
        5. Verify installation`)

	upgradeExample = templates.Examples(`
        # Check for updates without installing
        osbuilder upgrade --check-only
        
        # Upgrade to the latest version
        osbuilder upgrade
        
        # Upgrade to a specific version
        osbuilder upgrade --version v1.2.3
        
        # Force upgrade even if already on latest version
        osbuilder upgrade --force
        
        # Upgrade with auto-confirmation (no prompts)
        osbuilder upgrade --yes
        
        # Check upgrade from different source
        osbuilder upgrade --source mirror`)
)

// NewUpgradeCmd creates the "upgrade" command.
func NewUpgradeCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &UpgradeOptions{
		IOStreams: ioStreams,
		Source:    "official",
		Timeout:   60,
	}

	cmd := &cobra.Command{
		Use:                   "upgrade [flags]",
		Short:                 "Upgrade osbuilder to the latest version",
		Long:                  upgradeLongDesc,
		Example:               upgradeExample,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run())
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&opts.CheckOnly, "check-only", false, "Only check for updates, don't install")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Upgrade to specific version (e.g., v1.2.3)")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force upgrade even if already on latest version")
	cmd.Flags().StringVar(&opts.Source, "source", opts.Source, "Upgrade source (official, mirror)")
	cmd.Flags().BoolVarP(&opts.AutoConfirm, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", opts.Timeout, "Timeout for download operations (seconds)")

	return cmd
}

// Complete sets default values and discovers current installation.
func (o *UpgradeOptions) Complete() error {
	// Find current osbuilder binary
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}
	o.currentBinary = executablePath

	return nil
}

// Validate ensures provided inputs are valid.
func (o *UpgradeOptions) Validate() error {
	// Validate version format if specified
	if o.Version != "" {
		if !isValidVersionFormat(o.Version) {
			return fmt.Errorf("invalid version format: %s (expected format: v1.2.3)", o.Version)
		}
	}

	// Validate timeout
	if o.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got: %d", o.Timeout)
	}

	// Validate source
	validSources := []string{"official", "mirror"}
	if !contains(validSources, o.Source) {
		return fmt.Errorf("invalid source: %s (valid sources: %s)", o.Source, strings.Join(validSources, ", "))
	}

	return nil
}

// Run performs the upgrade operation.
func (o *UpgradeOptions) Run() error {
	fmt.Printf("%s Checking osbuilder version...\n", emoji.MagnifyingGlassTiltedLeft)

	// Get version information
	versionInfo, err := o.getVersionInfo()
	if err != nil {
		return fmt.Errorf("failed to get version information: %w", err)
	}

	// Display current version
	fmt.Printf("%s Current version: %s\n", emoji.Information, versionInfo.Current)

	// Get target version
	targetVersion := versionInfo.Latest
	if o.Version != "" {
		targetVersion = o.Version
		fmt.Printf("%s Target version: %s\n", emoji.DirectHit, targetVersion)
	} else {
		fmt.Printf("%s Latest version: %s\n", emoji.Star, versionInfo.Latest)
	}

	// Check if upgrade is needed
	if !o.Force && !o.needsUpgrade(versionInfo.Current, targetVersion) {
		fmt.Printf("%s Already on the latest version (%s)\n", emoji.CheckMarkButton, versionInfo.Current)
		if o.CheckOnly {
			return nil
		}
		if !o.AutoConfirm {
			fmt.Printf("%s Use --force to reinstall the same version\n", emoji.Information)
		}
		return nil
	}

	// If check-only mode, just report and exit
	if o.CheckOnly {
		if versionInfo.IsNewer || o.Version != "" {
			fmt.Printf("%s Update available: %s -> %s\n", emoji.UpArrow, versionInfo.Current, targetVersion)
		}
		return nil
	}

	// Confirm upgrade unless auto-confirm is enabled
	if !o.AutoConfirm {
		if !o.confirmUpgrade(versionInfo.Current, targetVersion) {
			fmt.Printf("%s Upgrade cancelled by user\n", emoji.CrossMark)
			return nil
		}
	}

	// Perform the upgrade
	return o.performUpgrade(targetVersion)
}

// getVersionInfo retrieves current and latest version information
func (o *UpgradeOptions) getVersionInfo() (*VersionInfo, error) {
	// Get current version
	currentVer := o.getCurrentVersion()

	// Get latest version from GitHub
	latestVer, err := o.getLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest version: %w", err)
	}

	// Compare versions
	isNewer := o.isNewerVersion(latestVer, currentVer)

	return &VersionInfo{
		Current: currentVer,
		Latest:  latestVer,
		IsNewer: isNewer,
	}, nil
}

// getCurrentVersion gets the current version of osbuilder
func (o *UpgradeOptions) getCurrentVersion() string {
	// Try to get version from osbuilder version command
	cmd := exec.Command(o.currentBinary, "version", "--short")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		if version != "" {
			return version
		}
	}

	// Fallback to build-time version
	return currentVersion
}

// getLatestVersion fetches the latest version from GitHub API
func (o *UpgradeOptions) getLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: time.Duration(o.Timeout) * time.Second,
	}

	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract tag_name from JSON response (simple regex approach)
	re := regexp.MustCompile(`"tag_name":\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract version from GitHub API response")
	}

	return matches[1], nil
}

// performUpgrade executes the upgrade process
func (o *UpgradeOptions) performUpgrade(targetVersion string) error {
	fmt.Printf("%s Starting upgrade to %s...\n", emoji.Rocket, targetVersion)

	// Construct go install command
	moduleWithVersion := modulePath
	if targetVersion != "" && targetVersion != "latest" {
		moduleWithVersion = fmt.Sprintf("%s@%s", modulePath, targetVersion)
	} else {
		moduleWithVersion = fmt.Sprintf("%s@latest", modulePath)
	}

	// Execute go install
	fmt.Printf("%s Installing %s...\n", emoji.Gear, moduleWithVersion)

	cmd := exec.Command("go", "install", moduleWithVersion)
	cmd.Env = os.Environ()

	// Capture output for better error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("upgrade failed: %w\nOutput: %s", err, string(output))
	}

	// Verify the installation
	if err := o.verifyInstallation(targetVersion); err != nil {
		return fmt.Errorf("upgrade verification failed: %w", err)
	}

	o.printUpgradeSuccess(targetVersion)
	return nil
}

// verifyInstallation checks if the upgrade was successful
func (o *UpgradeOptions) verifyInstallation(expectedVersion string) error {
	// Find osbuilder in PATH
	osbuilderPath, err := exec.LookPath("osbuilder")
	if err != nil {
		return fmt.Errorf("osbuilder not found in PATH after upgrade: %w", err)
	}

	// Check version
	cmd := exec.Command(osbuilderPath, "version", "--short")
	output, err := cmd.Output()
	if err != nil {
		// Version command might not exist, just check if binary exists and is executable
		return nil
	}

	installedVersion := strings.TrimSpace(string(output))
	if expectedVersion != "latest" && installedVersion != expectedVersion {
		return fmt.Errorf("version mismatch: expected %s, got %s", expectedVersion, installedVersion)
	}

	return nil
}

// confirmUpgrade prompts user for confirmation
func (o *UpgradeOptions) confirmUpgrade(currentVer, targetVer string) bool {
	fmt.Printf("%s Do you want to upgrade from %s to %s? [y/N]: ",
		emoji.QuestionMark, currentVer, targetVer)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// needsUpgrade determines if an upgrade is needed
func (o *UpgradeOptions) needsUpgrade(current, target string) bool {
	if o.Version != "" {
		// Specific version requested
		return current != target
	}

	// Check if target is newer than current
	return o.isNewerVersion(target, current)
}

// isNewerVersion compares two version strings (simple string comparison for now)
func (o *UpgradeOptions) isNewerVersion(v1, v2 string) bool {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// For now, use simple string comparison
	// TODO: Implement proper semantic version comparison
	return v1 > v2
}

// printUpgradeSuccess prints formatted success information
func (o *UpgradeOptions) printUpgradeSuccess(version string) {
	fmt.Printf("\n%s Successfully upgraded osbuilder to %s!\n", emoji.PartyingFace, version)
	fmt.Printf("%s Installation location: %s\n", emoji.Information, o.getInstallLocation())
	fmt.Printf("%s Run 'osbuilder version' to verify the installation\n", emoji.Information)

	// Print additional getting started info
	fmt.Printf("\n%s Getting Started:\n", emoji.Rocket)
	fmt.Printf("  • Run 'osbuilder --help' to see available commands\n")
	fmt.Printf("  • Check 'osbuilder version' to confirm the upgrade\n")
	fmt.Printf("  • Visit https://github.com/onexstack/osbuilder for documentation\n")
}

// getInstallLocation returns the likely installation location
func (o *UpgradeOptions) getInstallLocation() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		homeDir, _ := os.UserHomeDir()
		gopath = filepath.Join(homeDir, "go")
	}

	binaryName := "osbuilder"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	return filepath.Join(gopath, "bin", binaryName)
}

// Helper functions

// isValidVersionFormat checks if version follows semantic versioning
func isValidVersionFormat(version string) bool {
	// Simple regex for semantic versioning (v1.2.3 format)
	re := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9\.-]+)?(\+[a-zA-Z0-9\.-]+)?$`)
	return re.MatchString(version)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
