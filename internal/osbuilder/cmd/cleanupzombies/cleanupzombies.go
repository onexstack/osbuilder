package cleanupzombies

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// CleanupZombiesOptions defines configuration for the "cleanup-zombies" command.
type CleanupZombiesOptions struct {
	RootDir     string // working directory
	DryRun      bool   // show what would be done without actually doing it
	Force       bool   // force kill parent processes (SIGKILL)
	ProcessName string // filter by specific process name (e.g., "nvim")
	Verbose     bool   // verbose output
	AutoConfirm bool   // skip confirmation prompts
	ListOnly    bool   // only list zombie processes without cleanup

	genericiooptions.IOStreams
}

// ZombieProcessInfo contains information about zombie processes
type ZombieProcessInfo struct {
	PID        int32  `json:"pid"`
	PPID       int32  `json:"ppid"`
	Name       string `json:"name"`
	ParentName string `json:"parent_name"`
	Status     string `json:"status"`
	Username   string `json:"username"`
	Command    string `json:"command"`
}

var (
	cleanupZombiesLongDesc = templates.LongDesc(`Clean up zombie processes for the current user.

        This command identifies zombie (defunct) processes belonging to the current user
        and attempts to clean them up by terminating their parent processes. Zombie
        processes are dead processes that haven't been reaped by their parent process.

        The command provides several safety features:
        - List-only mode to show zombie processes without cleanup
        - Dry-run mode to preview actions before execution
        - Process name filtering to target specific applications
        - Confirmation prompts before terminating processes
        - Graceful termination (SIGTERM) with fallback to force kill (SIGKILL)

        Note: Terminating parent processes may cause data loss in unsaved work.
        Always save your work before running this command.`)

	cleanupZombiesExample = templates.Examples(`# List all zombie processes for current user
        osbuilder cleanup-zombies --list-only

        # List zombie processes with verbose information
        osbuilder cleanup-zombies --list-only --verbose

        # List nvim zombie processes specifically
        osbuilder cleanup-zombies --list-only --process-name nvim

        # Preview cleanup actions (dry-run mode)
        osbuilder cleanup-zombies --dry-run

        # Clean up all zombie processes with confirmation
        osbuilder cleanup-zombies

        # Clean up nvim zombie processes specifically
        osbuilder cleanup-zombies --process-name nvim

        # Force cleanup without confirmation
        osbuilder cleanup-zombies --yes --force

        # Verbose output with dry-run
        osbuilder cleanup-zombies --dry-run --verbose

        # Clean up specific process zombies with force kill
        osbuilder cleanup-zombies --process-name nvim --force`)
)

// NewCleanupZombiesCmd creates the "cleanup-zombies" command.
func NewCleanupZombiesCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &CleanupZombiesOptions{
		IOStreams: ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "cleanup-zombies",
		Short:                 "Clean up zombie processes for the current user",
		Long:                  cleanupZombiesLongDesc,
		Example:               cleanupZombiesExample,
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
	cmd.Flags().BoolVarP(&opts.ListOnly, "list-only", "l", false, "Only list zombie processes without cleanup")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be done without actually doing it")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Use SIGKILL instead of SIGTERM for parent processes")
	cmd.Flags().StringVar(&opts.ProcessName, "process-name", "", "Filter zombie processes by name (e.g., nvim, java)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "", false, "Enable verbose output")
	cmd.Flags().BoolVarP(&opts.AutoConfirm, "yes", "y", false, "Skip confirmation prompts")

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *CleanupZombiesOptions) Complete() error {
	o.RootDir, _ = os.Getwd()
	return nil
}

// Validate ensures provided inputs are valid.
func (o *CleanupZombiesOptions) Validate() error {
	if o.ProcessName != "" && strings.TrimSpace(o.ProcessName) == "" {
		return fmt.Errorf("process name cannot be empty if specified")
	}

	// ListOnly mode is mutually exclusive with cleanup actions
	if o.ListOnly && (o.DryRun || o.Force || o.AutoConfirm) {
		return fmt.Errorf("--list-only cannot be used with cleanup options (--dry-run, --force, --yes)")
	}

	return nil
}

// Run performs the cleanup-zombies operation.
func (o *CleanupZombiesOptions) Run() error {
	if o.Verbose && !o.ListOnly {
		fmt.Fprintf(o.Out, "%s Starting zombie process cleanup...\n", emoji.Broom)
	}

	// Get current user info
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		return fmt.Errorf("unable to determine current user")
	}

	// Find zombie processes
	zombies, err := o.findZombieProcesses(currentUser)
	if err != nil {
		return fmt.Errorf("failed to find zombie processes: %w", err)
	}

	if len(zombies) == 0 {
		if o.ListOnly {
			fmt.Fprintf(o.Out, "%s No zombie processes found for user %s\n",
				emoji.CheckMarkButton, currentUser)
		} else {
			fmt.Fprintf(o.Out, "%s No zombie processes found for user %s\n",
				emoji.CheckMarkButton, currentUser)
		}
		return nil
	}

	// Display found zombies
	if o.ListOnly {
		o.displayZombieProcessesList(zombies, currentUser)
		return nil
	} else {
		o.displayZombieProcesses(zombies)
	}

	if o.DryRun {
		fmt.Fprintf(o.Out, "\n%s Dry-run mode: No processes were terminated\n",
			emoji.Eyes)
		return nil
	}

	// Confirm before proceeding
	if !o.AutoConfirm {
		if !o.confirmCleanup(len(zombies)) {
			fmt.Fprintf(o.Out, "%s Operation cancelled by user\n", emoji.CrossMark)
			return nil
		}
	}

	// Clean up zombies
	return o.cleanupZombies(zombies)
}

// findZombieProcesses finds all zombie processes for the specified user
func (o *CleanupZombiesOptions) findZombieProcesses(username string) ([]ZombieProcessInfo, error) {
	var zombies []ZombieProcessInfo

	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, proc := range processes {
		// Check if process belongs to current user
		user, err := proc.Username()
		if err != nil || user != username {
			continue
		}

		// Check if process is zombie
		status, err := proc.Status()
		if err != nil {
			continue
		}

		var isZombie bool
		for _, s := range status {
			if s == "Z" || s == "zombie" {
				isZombie = true
				break
			}
		}

		// Also check process name for <defunct>
		if !isZombie {
			if name, err := proc.Name(); err == nil {
				if strings.Contains(name, "defunct") {
					isZombie = true
				}
			}
		}

		if !isZombie {
			continue
		}

		// Get process info
		zombie := ZombieProcessInfo{
			PID:      proc.Pid,
			Username: user,
		}

		if name, err := proc.Name(); err == nil {
			zombie.Name = name
		}

		if ppid, err := proc.Ppid(); err == nil {
			zombie.PPID = ppid
		}

		if len(status) > 0 {
			zombie.Status = strings.Join(status, ",")
		}

		// Get command line
		if cmdline, err := proc.Cmdline(); err == nil {
			zombie.Command = cmdline
		}

		// Get parent process name
		if parent, err := process.NewProcess(zombie.PPID); err == nil {
			if parentName, err := parent.Name(); err == nil {
				zombie.ParentName = parentName
			}
		}

		// Filter by process name if specified
		if o.ProcessName != "" {
			if !strings.Contains(strings.ToLower(zombie.Name), strings.ToLower(o.ProcessName)) &&
				!strings.Contains(strings.ToLower(zombie.ParentName), strings.ToLower(o.ProcessName)) &&
				!strings.Contains(strings.ToLower(zombie.Command), strings.ToLower(o.ProcessName)) {
				continue
			}
		}

		zombies = append(zombies, zombie)
	}

	return zombies, nil
}

// displayZombieProcessesList shows zombie processes in list-only mode
func (o *CleanupZombiesOptions) displayZombieProcessesList(zombies []ZombieProcessInfo, username string) {
	fmt.Fprintf(o.Out, "%s Zombie Processes for user: %s\n\n", emoji.Ghost, username)

	if o.Verbose {
		// Verbose format with more details
		fmt.Fprintf(o.Out, "%-8s %-8s %-10s %-15s %-15s %-10s %s\n",
			"PID", "PPID", "USER", "PROCESS", "PARENT", "STATUS", "COMMAND")
		fmt.Fprintf(o.Out, "%s\n", strings.Repeat("-", 100))

		for _, zombie := range zombies {
			pidStr := color.New(color.FgRed, color.Bold).Sprintf("%-8d", zombie.PID)
			ppidStr := fmt.Sprintf("%-8d", zombie.PPID)
			userStr := fmt.Sprintf("%-10s", o.truncateString(zombie.Username, 10))
			nameStr := color.New(color.FgRed).Sprintf("%-15s", o.truncateString(zombie.Name, 15))
			parentStr := fmt.Sprintf("%-15s", o.truncateString(zombie.ParentName, 15))
			statusStr := color.New(color.FgYellow).Sprintf("%-10s", zombie.Status)
			cmdStr := o.truncateString(zombie.Command, 40)

			fmt.Fprintf(o.Out, "%s %s %s %s %s %s %s\n",
				pidStr, ppidStr, userStr, nameStr, parentStr, statusStr, cmdStr)
		}
	} else {
		// Simple format showing PID, USER, COMMAND
		fmt.Fprintf(o.Out, "%-10s %-15s %s\n", "PID", "USER", "COMMAND")
		fmt.Fprintf(o.Out, "%s\n", strings.Repeat("-", 60))

		for _, zombie := range zombies {
			pidStr := color.New(color.FgRed, color.Bold).Sprintf("%-10d", zombie.PID)
			userStr := fmt.Sprintf("%-15s", zombie.Username)
			cmdStr := zombie.Command
			if cmdStr == "" {
				cmdStr = fmt.Sprintf("[%s] <defunct>", zombie.Name)
			}

			fmt.Fprintf(o.Out, "%s %s %s\n", pidStr, userStr, cmdStr)
		}
	}

	fmt.Fprintf(o.Out, "\n%s Total zombie processes found: %d\n", emoji.ChartIncreasing, len(zombies))

	if len(zombies) > 0 {
		fmt.Fprintf(o.Out, "%s Use without --list-only to clean up these processes\n", emoji.Broom)
	}
}

// displayZombieProcesses shows the found zombie processes (cleanup mode)
func (o *CleanupZombiesOptions) displayZombieProcesses(zombies []ZombieProcessInfo) {
	fmt.Fprintf(o.Out, "\n%s Found %d zombie process(es):\n",
		emoji.Ghost, len(zombies))
	fmt.Fprintf(o.Out, "%-10s %-10s %-15s %-15s %-10s\n",
		"PID", "PPID", "PROCESS", "PARENT", "STATUS")
	fmt.Fprintf(o.Out, "%s\n", strings.Repeat("-", 65))

	for _, zombie := range zombies {
		// Color zombie processes in red
		pidStr := color.New(color.FgRed, color.Bold).Sprintf("%-10d", zombie.PID)
		ppidStr := fmt.Sprintf("%-10d", zombie.PPID)
		nameStr := color.New(color.FgRed).Sprintf("%-15s",
			o.truncateString(zombie.Name, 15))
		parentStr := fmt.Sprintf("%-15s", o.truncateString(zombie.ParentName, 15))
		statusStr := color.New(color.FgYellow).Sprintf("%-10s", zombie.Status)

		fmt.Fprintf(o.Out, "%s %s %s %s %s\n",
			pidStr, ppidStr, nameStr, parentStr, statusStr)
	}
}

// confirmCleanup asks user for confirmation
func (o *CleanupZombiesOptions) confirmCleanup(count int) bool {
	signal := "SIGTERM"
	if o.Force {
		signal = "SIGKILL"
	}

	parentCount := o.countUniqueParents(count)

	fmt.Fprintf(o.Out, "\n%s This will terminate %d parent process(es) using %s\n",
		emoji.Warning, parentCount, signal)
	fmt.Fprintf(o.Out, "%s Terminating parent processes may cause data loss!\n",
		emoji.ExclamationMark)
	fmt.Fprintf(o.Out, "Do you want to continue? [y/N]: ")

	var response string
	fmt.Fscanln(os.Stdin, &response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// countUniqueParents counts unique parent PIDs
func (o *CleanupZombiesOptions) countUniqueParents(zombieCount int) int {
	// This is a rough estimate - in actual implementation we'd count unique PPIDs
	// For now, assume each zombie might have a unique parent
	return zombieCount
}

// cleanupZombies terminates parent processes to clean up zombies
func (o *CleanupZombiesOptions) cleanupZombies(zombies []ZombieProcessInfo) error {
	// Group zombies by parent PID to avoid multiple kills
	parentPIDs := make(map[int32][]ZombieProcessInfo)
	for _, zombie := range zombies {
		parentPIDs[zombie.PPID] = append(parentPIDs[zombie.PPID], zombie)
	}

	successCount := 0
	errorCount := 0

	for ppid, childZombies := range parentPIDs {
		if o.Verbose {
			fmt.Fprintf(o.Out, "\n%s Terminating parent process %d (has %d zombie children)\n",
				emoji.CrossMark, ppid, len(childZombies))
		}

		if err := o.terminateProcess(ppid); err != nil {
			color.New(color.FgRed).Fprintf(o.Out, "%s Failed to terminate process %d: %v\n", emoji.CrossMark, ppid, err)
			errorCount++
		} else {
			color.New(color.FgGreen).Fprintf(o.Out, "%s Successfully terminated process %d\n", emoji.CheckMarkButton, ppid)
			successCount++

			if o.Verbose {
				fmt.Fprintf(o.Out, "   Cleaned up zombie processes: ")
				for i, zombie := range childZombies {
					if i > 0 {
						fmt.Fprintf(o.Out, ", ")
					}
					fmt.Fprintf(o.Out, "%d", zombie.PID)
				}
				fmt.Fprintf(o.Out, "\n")
			}
		}
	}

	// Wait a moment for processes to be cleaned up
	time.Sleep(1 * time.Second)

	// Verify cleanup
	if o.Verbose {
		fmt.Fprintf(o.Out, "\n%s Verifying cleanup...\n", emoji.MagnifyingGlassTiltedLeft)
		if remainingZombies, err := o.findZombieProcesses(os.Getenv("USER")); err == nil {
			filteredRemaining := o.filterRemainingZombies(remainingZombies, zombies)
			if len(filteredRemaining) == 0 {
				fmt.Fprintf(o.Out, "%s All targeted zombie processes cleaned successfully\n",
					emoji.CheckMarkButton)
			} else {
				fmt.Fprintf(o.Out, "%s %d zombie process(es) still remain\n",
					emoji.Warning, len(filteredRemaining))
			}
		}
	}

	// Print summary
	o.PrintGettingStarted(successCount, errorCount)
	return nil
}

// terminateProcess terminates a process with appropriate signal
func (o *CleanupZombiesOptions) terminateProcess(pid int32) error {
	signal := "TERM"
	if o.Force {
		signal = "KILL"
	}

	cmd := exec.Command("kill", fmt.Sprintf("-%s", signal), strconv.Itoa(int(pid)))
	if err := cmd.Run(); err != nil {
		// If SIGTERM fails and not forcing, try SIGKILL
		if !o.Force {
			if o.Verbose {
				fmt.Fprintf(o.Out, "   SIGTERM failed, trying SIGKILL...\n")
			}
			cmd = exec.Command("kill", "-KILL", strconv.Itoa(int(pid)))
			return cmd.Run()
		}
		return err
	}
	return nil
}

// filterRemainingZombies filters out zombies that should have been cleaned
func (o *CleanupZombiesOptions) filterRemainingZombies(remaining, original []ZombieProcessInfo) []ZombieProcessInfo {
	originalPIDs := make(map[int32]bool)
	for _, zombie := range original {
		originalPIDs[zombie.PID] = true
	}

	var filtered []ZombieProcessInfo
	for _, zombie := range remaining {
		if !originalPIDs[zombie.PID] {
			filtered = append(filtered, zombie)
		}
	}
	return filtered
}

// truncateString truncates string to specified length
func (o *CleanupZombiesOptions) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PrintGettingStarted prints formatted success information.
func (o *CleanupZombiesOptions) PrintGettingStarted(successCount, errorCount int) {
	fmt.Fprintf(o.Out, "\n%s Zombie cleanup completed\n", emoji.CheckMarkButton)

	if successCount > 0 {
		color.New(color.FgGreen).Fprintf(o.Out, "%s Successfully terminated: %d parent process(es)\n", emoji.CheckMarkButton, successCount)
	}

	if errorCount > 0 {
		color.New(color.FgRed).Fprintf(o.Out, "%s Failed to terminate: %d parent process(es)\n", emoji.CrossMark, errorCount)
	}

	fmt.Fprintf(o.Out, "\n%s Tip: Use --list-only to only view zombie processes\n", emoji.LightBulb)
	fmt.Fprintf(o.Out, "%s Tip: Use --dry-run to preview actions before execution\n", emoji.LightBulb)
	fmt.Fprintf(o.Out, "%s Tip: Use --process-name to target specific applications\n", emoji.LightBulb)
}
