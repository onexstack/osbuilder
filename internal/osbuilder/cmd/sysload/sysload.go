package sysload

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	//"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	//"syscall"
	"time"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// SysloadOptions defines configuration for the "sysload" command.
type SysloadOptions struct {
	RootDir             string  // working directory
	Output              string  // output format: table, json
	TopN                int     // top N processes to show
	Watch               bool    // continuous monitoring
	Interval            int     // refresh interval in seconds
	NoColor             bool    // disable colored output
	SaveToFile          string  // save output to file
	CommandWidth        int     // width for command column display
	ProcessCPUThreshold float64 // CPU% threshold for process highlighting
	ProcessMemThreshold float64 // MEM% threshold for process highlighting
	Wide                bool    // show extended information including I/O utilization

	genericiooptions.IOStreams
}

// SystemMetrics contains system performance metrics
type SystemMetrics struct {
	Timestamp   time.Time     `json:"timestamp"`
	CPU         CPUMetrics    `json:"cpu"`
	Memory      MemoryMetrics `json:"memory"`
	Disk        DiskMetrics   `json:"disk"`
	LoadAverage LoadMetrics   `json:"load_average"`
	TopCPU      []ProcessInfo `json:"top_cpu_processes"`
	TopMemory   []ProcessInfo `json:"top_memory_processes"`
	TopIO       []ProcessInfo `json:"top_io_processes,omitempty"`
	TopIOWait   []ProcessInfo `json:"top_iowait_processes,omitempty"`
}

// CPUMetrics contains CPU-related metrics
type CPUMetrics struct {
	Usage         float64   `json:"usage_percent"`
	User          float64   `json:"user_percent"`
	System        float64   `json:"system_percent"`
	Idle          float64   `json:"idle_percent"`
	IOWait        float64   `json:"iowait_percent"`
	IRQ           float64   `json:"irq_percent"`
	SoftIRQ       float64   `json:"soft_irq_percent"`
	Steal         float64   `json:"steal_percent"`
	Guest         float64   `json:"guest_percent"`
	GuestNice     float64   `json:"guest_nice_percent"`
	LoadAvg1      float64   `json:"load_avg_1min"`
	LoadAvg5      float64   `json:"load_avg_5min"`
	LoadAvg15     float64   `json:"load_avg_15min"`
	PhysicalCores int       `json:"physical_cores"`
	LogicalCores  int       `json:"logical_cores"`
	Architecture  string    `json:"architecture"`
	CPUInfo       []CPUInfo `json:"cpu_info,omitempty"`
	PerCoreUsage  []float64 `json:"per_core_usage,omitempty"`
	PerCoreIOWait []float64 `json:"per_core_iowait,omitempty"`
	IsAbnormal    bool      `json:"is_abnormal"`
	IsIOWaitHigh  bool      `json:"is_iowait_high"`
}

// CPUInfo contains detailed CPU information
type CPUInfo struct {
	ModelName string   `json:"model_name"`
	Family    string   `json:"family"`
	Model     string   `json:"model"`
	Stepping  int32    `json:"stepping"`
	MHz       float64  `json:"mhz"`
	CacheSize int32    `json:"cache_size"`
	Vendor    string   `json:"vendor"`
	Flags     []string `json:"flags,omitempty"`
}

// MemoryMetrics contains memory-related metrics
type MemoryMetrics struct {
	Total        uint64  `json:"total_bytes"`
	Used         uint64  `json:"used_bytes"`
	Available    uint64  `json:"available_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	AvailableGB  float64 `json:"available_gb"`
	IsAbnormal   bool    `json:"is_abnormal"`
}

// DiskMetrics contains disk I/O metrics
type DiskMetrics struct {
	Usage        float64           `json:"usage_percent"`
	TotalGB      float64           `json:"total_gb"`
	UsedGB       float64           `json:"used_gb"`
	FreeGB       float64           `json:"free_gb"`
	ReadMBPS     float64           `json:"read_mbps,omitempty"`
	WriteMBPS    float64           `json:"write_mbps,omitempty"`
	IOUtil       float64           `json:"io_util_percent,omitempty"`
	AvgQueueSize float64           `json:"avg_queue_size,omitempty"`
	AwaitTime    float64           `json:"await_time_ms,omitempty"`
	DiskStats    []DiskStatMetrics `json:"disk_stats,omitempty"`
	IsAbnormal   bool              `json:"is_abnormal"`
}

// DiskStatMetrics contains per-device disk statistics
type DiskStatMetrics struct {
	Device       string `json:"device"`
	ReadCount    uint64 `json:"read_count"`
	WriteCount   uint64 `json:"write_count"`
	ReadBytes    uint64 `json:"read_bytes"`
	WriteBytes   uint64 `json:"write_bytes"`
	ReadTime     uint64 `json:"read_time_ms"`
	WriteTime    uint64 `json:"write_time_ms"`
	IOTime       uint64 `json:"io_time_ms"`
	WeightedIO   uint64 `json:"weighted_io_time_ms"`
	IopsInFlight uint64 `json:"iops_in_flight"`
}

// LoadMetrics contains system load metrics
type LoadMetrics struct {
	Load1      float64 `json:"load1"`
	Load5      float64 `json:"load5"`
	Load15     float64 `json:"load15"`
	IsAbnormal bool    `json:"is_abnormal"`
}

// ProcessInfo contains process information
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryMB      float64 `json:"memory_mb"`
	MemoryPercent float64 `json:"memory_percent"`
	Command       string  `json:"command"`
	IOReadMB      float64 `json:"io_read_mb,omitempty"`
	IOWriteMB     float64 `json:"io_write_mb,omitempty"`
	IOReadRate    float64 `json:"io_read_rate_mbps,omitempty"`
	IOWriteRate   float64 `json:"io_write_rate_mbps,omitempty"`
	Status        string  `json:"status,omitempty"`
	IOWaitTime    float64 `json:"iowait_time_ms,omitempty"`
	BlockedReason string  `json:"blocked_reason,omitempty"`
}

var (
	sysloadLongDesc = templates.LongDesc(`Display comprehensive system load information to assist with performance troubleshooting.  

    This command provides detailed system performance metrics including CPU usage, memory consumption,   
    disk I/O statistics, IOWait analysis, and system load averages. It also shows the top processes   
    consuming CPU, memory, and I/O resources to help identify performance bottlenecks.  

    Abnormal metrics are highlighted in red to quickly draw attention to potential issues:  

    - CPU usage > 80%  

    - IOWait > 20%  

    - Memory usage > 85%  

    - Disk usage > 85%  

    - Load average > CPU cores * 2  

    - Process CPU% > threshold (default: 50%)  

    - Process MEM% > threshold (default: 20%)  

    The command supports both one-time snapshots and continuous monitoring modes, with multiple   
    output formats for integration with monitoring and alerting systems.`)

	sysloadExample = templates.Examples(`# Show current system load with top 5 processes  
        osbuilder sysload  

        # Show top 10 processes in each category  
        osbuilder sysload --top 10  

        # Show extended information including I/O utilization
        osbuilder sysload --wide

        # Set custom thresholds for process highlighting  
        osbuilder sysload --process-cpu-threshold 30 --process-mem-threshold 5  

        # Set command display width to 120 characters  
        osbuilder sysload --command-width 120  

        # Output in JSON format with extended information
        osbuilder sysload --output json --wide

        # Continuous monitoring with 10-second intervals and wide output
        osbuilder sysload --watch --interval 10 --wide

        # Save extended output to file  
        osbuilder sysload --save-to-file system-load.json --output json --wide

        # Disable colored output (for scripts)  
        osbuilder sysload --no-color`)
)

// Thresholds for abnormal conditions
const (
	CPUThreshold               = 80.0 // CPU usage > 80%
	IOWaitThreshold            = 20.0 // IOWait > 20%
	MemoryThreshold            = 85.0 // Memory usage > 85%
	DiskThreshold              = 85.0 // Disk usage > 85%
	LoadThreshold              = 2.0  // Load > cores * 2
	IOUtilThreshold            = 80.0 // Disk I/O utilization > 80%
	DefaultProcessCPUThreshold = 50.0 // Process CPU% > 50%
	DefaultProcessMemThreshold = 20.0 // Process MEM% > 20%
)

// NewSysloadCmd creates the "sysload" command.
func NewSysloadCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &SysloadOptions{
		IOStreams:           ioStreams,
		Output:              "table",
		TopN:                5,
		Interval:            5,
		CommandWidth:        80, // Default command width
		ProcessCPUThreshold: DefaultProcessCPUThreshold,
		ProcessMemThreshold: DefaultProcessMemThreshold,
		Wide:                false, // Default to compact output
	}

	cmd := &cobra.Command{
		Use:                   "sysload",
		Short:                 "Display system load information for performance troubleshooting",
		Long:                  sysloadLongDesc,
		Example:               sysloadExample,
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

	// Add flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", opts.Output, "Output format: table, json")
	cmd.Flags().IntVarP(&opts.TopN, "top", "n", opts.TopN, "Show top N processes (default: 5)")
	cmd.Flags().BoolVarP(&opts.Watch, "watch", "w", false, "Continuously monitor system load")
	cmd.Flags().IntVar(&opts.Interval, "interval", opts.Interval, "Refresh interval in seconds for watch mode")
	cmd.Flags().BoolVar(&opts.NoColor, "no-color", false, "Disable colored output")
	cmd.Flags().StringVar(&opts.SaveToFile, "save-to-file", "", "Save output to specified file")
	cmd.Flags().IntVar(&opts.CommandWidth, "command-width", opts.CommandWidth, "Maximum width for command column display (default: 80)")
	cmd.Flags().Float64Var(&opts.ProcessCPUThreshold, "process-cpu-threshold", opts.ProcessCPUThreshold, "CPU percentage threshold for process highlighting (default: 50)")
	cmd.Flags().Float64Var(&opts.ProcessMemThreshold, "process-mem-threshold", opts.ProcessMemThreshold, "Memory percentage threshold for process highlighting (default: 10)")
	cmd.Flags().BoolVar(&opts.Wide, "wide", false, "Show extended information including I/O utilization and detailed metrics")

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *SysloadOptions) Complete() error {
	o.RootDir, _ = os.Getwd()

	// Normalize output format
	o.Output = strings.ToLower(o.Output)

	// Disable color in non-interactive environments or when requested
	if o.NoColor || o.SaveToFile != "" {
		color.NoColor = true
	}

	return nil
}

// Validate ensures provided inputs are valid.
func (o *SysloadOptions) Validate() error {
	// Validate output format
	if o.Output != "table" && o.Output != "json" {
		return fmt.Errorf("invalid output format %q, supported formats: table, json", o.Output)
	}

	// Validate top N
	if o.TopN < 1 || o.TopN > 50 {
		return fmt.Errorf("top N must be between 1 and 50, got %d", o.TopN)
	}

	// Validate interval
	if o.Interval < 1 || o.Interval > 300 {
		return fmt.Errorf("interval must be between 1 and 300 seconds, got %d", o.Interval)
	}

	// Validate command width
	if o.CommandWidth < 20 || o.CommandWidth > 500 {
		return fmt.Errorf("command width must be between 20 and 500 characters, got %d", o.CommandWidth)
	}

	// Validate process thresholds
	if o.ProcessCPUThreshold < 0 || o.ProcessCPUThreshold > 100 {
		return fmt.Errorf("process CPU threshold must be between 0 and 100, got %.1f", o.ProcessCPUThreshold)
	}

	if o.ProcessMemThreshold < 0 || o.ProcessMemThreshold > 100 {
		return fmt.Errorf("process memory threshold must be between 0 and 100, got %.1f", o.ProcessMemThreshold)
	}

	// Validate save file path
	if o.SaveToFile != "" {
		dir := filepath.Dir(o.SaveToFile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cannot create directory for save file: %w", err)
		}
	}

	return nil
}

// Run performs the sysload operation.
func (o *SysloadOptions) Run() error {
	if o.Watch {
		return o.watchMode()
	}
	return o.snapshotMode()
}

// snapshotMode displays current system load once
func (o *SysloadOptions) snapshotMode() error {
	metrics, err := o.collectSystemMetrics()
	if err != nil {
		return fmt.Errorf("failed to collect system metrics: %w", err)
	}

	return o.displayMetrics(metrics)
}

// watchMode continuously monitors system load
func (o *SysloadOptions) watchMode() error {
	if !o.NoColor {
		fmt.Fprintf(o.Out, "%s Starting continuous monitoring (interval: %ds)\n",
			emoji.Eyes, o.Interval)
		fmt.Fprintf(o.Out, "Press Ctrl+C to stop...\n\n")
	}

	ticker := time.NewTicker(time.Duration(o.Interval) * time.Second)
	defer ticker.Stop()

	// Initial display
	if err := o.snapshotMode(); err != nil {
		return err
	}

	for range ticker.C {
		// Clear screen for table output
		if o.Output == "table" && o.SaveToFile == "" {
			fmt.Fprintf(o.Out, "\033[2J\033[H")
		}

		if err := o.snapshotMode(); err != nil {
			return err
		}
	}

	return nil
}

// collectSystemMetrics gathers all system performance metrics
func (o *SysloadOptions) collectSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// Collect detailed CPU metrics with IOWait
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU times: %w", err)
	}

	// Get per-core CPU times for IOWait analysis (only in wide mode)
	var perCoreCPUTimes []cpu.TimesStat
	if o.Wide {
		perCoreCPUTimes, err = cpu.Times(true)
		if err != nil {
			perCoreCPUTimes = []cpu.TimesStat{} // fallback
		}
	}

	var totalCPU cpu.TimesStat
	if len(cpuTimes) > 0 {
		totalCPU = cpuTimes[0]
	}

	// Calculate CPU percentages
	total := totalCPU.User + totalCPU.System + totalCPU.Nice + totalCPU.Idle +
		totalCPU.Iowait + totalCPU.Irq + totalCPU.Softirq + totalCPU.Steal +
		totalCPU.Guest + totalCPU.GuestNice

	var perCoreIOWait []float64
	if o.Wide {
		for _, coreCPU := range perCoreCPUTimes {
			coreTotal := coreCPU.User + coreCPU.System + coreCPU.Nice + coreCPU.Idle +
				coreCPU.Iowait + coreCPU.Irq + coreCPU.Softirq + coreCPU.Steal +
				coreCPU.Guest + coreCPU.GuestNice
			if coreTotal > 0 {
				perCoreIOWait = append(perCoreIOWait, (coreCPU.Iowait/coreTotal)*100)
			} else {
				perCoreIOWait = append(perCoreIOWait, 0)
			}
		}
	}

	// Get per-core CPU usage (only in wide mode)
	var perCorePercent []float64
	if o.Wide {
		perCorePercent, err = cpu.Percent(time.Second, true)
		if err != nil {
			perCorePercent = []float64{} // fallback to empty slice
		}
	}

	// Get CPU counts
	physicalCores, err := cpu.Counts(false) // physical cores
	if err != nil {
		physicalCores = 1 // fallback
	}

	logicalCores, err := cpu.Counts(true) // logical cores (including hyperthreading)
	if err != nil {
		logicalCores = physicalCores // fallback to physical cores
	}

	// Get detailed CPU info (only in wide mode)
	var cpuInfoList []CPUInfo
	if o.Wide {
		cpuInfos, err := cpu.Info()
		if err == nil && len(cpuInfos) > 0 {
			// Usually we only need info from the first CPU as they're typically identical
			info := cpuInfos[0]
			cpuInfoList = append(cpuInfoList, CPUInfo{
				ModelName: info.ModelName,
				Family:    info.Family,
				Model:     info.Model,
				Stepping:  info.Stepping,
				MHz:       info.Mhz,
				CacheSize: info.CacheSize,
				Vendor:    info.VendorID,
				Flags:     info.Flags,
			})
		}
	}

	// Get architecture info
	architecture := runtime.GOARCH
	if o.Wide && len(cpuInfoList) > 0 && cpuInfoList[0].Vendor != "" {
		// Enhance architecture info with vendor
		switch cpuInfoList[0].Vendor {
		case "GenuineIntel":
			architecture = fmt.Sprintf("%s (Intel)", architecture)
		case "AuthenticAMD":
			architecture = fmt.Sprintf("%s (AMD)", architecture)
		default:
			architecture = fmt.Sprintf("%s (%s)", architecture, cpuInfoList[0].Vendor)
		}
	}

	loadAvg, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("failed to get load average: %w", err)
	}

	var iowaitPercent float64
	if total > 0 {
		iowaitPercent = (totalCPU.Iowait / total) * 100
	}

	metrics.CPU = CPUMetrics{
		Usage:         100 - ((totalCPU.Idle / total) * 100),
		User:          (totalCPU.User / total) * 100,
		System:        (totalCPU.System / total) * 100,
		Idle:          (totalCPU.Idle / total) * 100,
		IOWait:        iowaitPercent,
		IRQ:           (totalCPU.Irq / total) * 100,
		SoftIRQ:       (totalCPU.Softirq / total) * 100,
		Steal:         (totalCPU.Steal / total) * 100,
		Guest:         (totalCPU.Guest / total) * 100,
		GuestNice:     (totalCPU.GuestNice / total) * 100,
		LoadAvg1:      loadAvg.Load1,
		LoadAvg5:      loadAvg.Load5,
		LoadAvg15:     loadAvg.Load15,
		PhysicalCores: physicalCores,
		LogicalCores:  logicalCores,
		Architecture:  architecture,
		CPUInfo:       cpuInfoList,
		PerCoreUsage:  perCorePercent,
		PerCoreIOWait: perCoreIOWait,
		IsAbnormal:    (100 - ((totalCPU.Idle / total) * 100)) > CPUThreshold,
		IsIOWaitHigh:  iowaitPercent > IOWaitThreshold,
	}

	// Collect memory metrics
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	metrics.Memory = MemoryMetrics{
		Total:        memInfo.Total,
		Used:         memInfo.Used,
		Available:    memInfo.Available,
		UsagePercent: memInfo.UsedPercent,
		TotalGB:      float64(memInfo.Total) / 1024 / 1024 / 1024,
		UsedGB:       float64(memInfo.Used) / 1024 / 1024 / 1024,
		AvailableGB:  float64(memInfo.Available) / 1024 / 1024 / 1024,
		IsAbnormal:   memInfo.UsedPercent > MemoryThreshold,
	}

	// Collect enhanced disk metrics
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}

	// Get disk I/O statistics (only in wide mode)
	var diskStats []DiskStatMetrics
	var totalIOUtil float64
	if o.Wide {
		diskIOStats, err := disk.IOCounters()
		if err == nil {
			for device, stats := range diskIOStats {
				diskStat := DiskStatMetrics{
					Device:       device,
					ReadCount:    stats.ReadCount,
					WriteCount:   stats.WriteCount,
					ReadBytes:    stats.ReadBytes,
					WriteBytes:   stats.WriteBytes,
					ReadTime:     stats.ReadTime,
					WriteTime:    stats.WriteTime,
					IOTime:       stats.IoTime,
					WeightedIO:   stats.WeightedIO,
					IopsInFlight: stats.IopsInProgress,
				}
				diskStats = append(diskStats, diskStat)

				// Calculate I/O utilization (simplified)
				if stats.IoTime > 0 {
					totalIOUtil += float64(stats.IoTime) / 1000.0 // Convert to percentage
				}
			}
			// Average I/O utilization across devices
			if len(diskStats) > 0 {
				totalIOUtil = totalIOUtil / float64(len(diskStats))
				if totalIOUtil > 100 {
					totalIOUtil = 100
				}
			}
		}
	}

	diskMetrics := DiskMetrics{
		Usage:      diskInfo.UsedPercent,
		TotalGB:    float64(diskInfo.Total) / 1024 / 1024 / 1024,
		UsedGB:     float64(diskInfo.Used) / 1024 / 1024 / 1024,
		FreeGB:     float64(diskInfo.Free) / 1024 / 1024 / 1024,
		IsAbnormal: diskInfo.UsedPercent > DiskThreshold,
	}

	// Only include extended disk metrics in wide mode
	if o.Wide {
		diskMetrics.IOUtil = totalIOUtil
		diskMetrics.DiskStats = diskStats
	}

	metrics.Disk = diskMetrics

	// Collect load metrics
	metrics.LoadAverage = LoadMetrics{
		Load1:      loadAvg.Load1,
		Load5:      loadAvg.Load5,
		Load15:     loadAvg.Load15,
		IsAbnormal: loadAvg.Load1 > float64(logicalCores)*LoadThreshold,
	}

	// Collect top processes
	topCPU, err := o.getTopProcesses("cpu", o.TopN)
	if err != nil {
		return nil, fmt.Errorf("failed to get top CPU processes: %w", err)
	}
	metrics.TopCPU = topCPU

	topMemory, err := o.getTopProcesses("memory", o.TopN)
	if err != nil {
		return nil, fmt.Errorf("failed to get top memory processes: %w", err)
	}
	metrics.TopMemory = topMemory

	// Only collect I/O and IOWait processes in wide mode
	if o.Wide {
		topIO, err := o.getTopProcesses("io", o.TopN)
		if err != nil {
			return nil, fmt.Errorf("failed to get top I/O processes: %w", err)
		}
		metrics.TopIO = topIO

		// Collect top IOWait processes
		topIOWait, err := o.getTopProcesses("iowait", o.TopN)
		if err != nil {
			return nil, fmt.Errorf("failed to get top IOWait processes: %w", err)
		}
		metrics.TopIOWait = topIOWait
	}

	return metrics, nil
}

// getTopProcesses returns top N processes sorted by specified metric
func (o *SysloadOptions) getTopProcesses(sortBy string, n int) ([]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var procInfos []ProcessInfo

	for _, proc := range processes {
		info := ProcessInfo{PID: proc.Pid}

		// Get process name
		if name, err := proc.Name(); err == nil {
			info.Name = name
		}

		// Get username
		if username, err := proc.Username(); err == nil {
			info.Username = username
		}

		// Get process status (only in wide mode for I/O and IOWait sorting)
		if o.Wide && (sortBy == "io" || sortBy == "iowait") {
			if statuses, err := proc.Status(); err == nil {
				if len(statuses) > 0 {
					info.Status = strings.Join(statuses, ",")
					// Check if process is in uninterruptible sleep (D state) which often indicates I/O wait
					for _, status := range statuses {
						if status == "D" || status == "disk sleep" || status == "uninterruptible" {
							info.BlockedReason = "I/O Wait"
							break
						}
					}
				}
			}
		}

		// Get command line with configurable width
		if cmdline, err := proc.Cmdline(); err == nil {
			info.Command = cmdline
			// Truncate long command lines based on configured width
			if len(info.Command) > o.CommandWidth {
				info.Command = info.Command[:o.CommandWidth-3] + "..."
			}
		}

		// Get CPU percent
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			info.CPUPercent = cpuPercent
		}

		// Get memory info
		if memInfo, err := proc.MemoryInfo(); err == nil {
			info.MemoryMB = float64(memInfo.RSS) / 1024 / 1024
		}

		if memPercent, err := proc.MemoryPercent(); err == nil {
			info.MemoryPercent = float64(memPercent)
		}

		// Get I/O info (only in wide mode for I/O-related sorting)
		if o.Wide && (sortBy == "io" || sortBy == "iowait") {
			if ioCounters, err := proc.IOCounters(); err == nil {
				info.IOReadMB = float64(ioCounters.ReadBytes) / 1024 / 1024
				info.IOWriteMB = float64(ioCounters.WriteBytes) / 1024 / 1024

				// Estimate I/O wait based on I/O activity and process state
				totalIO := info.IOReadMB + info.IOWriteMB
				if info.Status == "D" && totalIO > 0 {
					// This is a rough estimation - in reality, you'd need more sophisticated timing
					info.IOWaitTime = totalIO * 10 // Rough approximation
				}
			}
		}

		procInfos = append(procInfos, info)
	}

	// Sort processes based on the specified metric
	sort.Slice(procInfos, func(i, j int) bool {
		switch sortBy {
		case "cpu":
			return procInfos[i].CPUPercent > procInfos[j].CPUPercent
		case "memory":
			return procInfos[i].MemoryMB > procInfos[j].MemoryMB
		case "io":
			return (procInfos[i].IOReadMB + procInfos[i].IOWriteMB) >
				(procInfos[j].IOReadMB + procInfos[j].IOWriteMB)
		case "iowait":
			// Sort by processes in D state first, then by I/O activity
			if procInfos[i].Status == "D" && procInfos[j].Status != "D" {
				return true
			}
			if procInfos[i].Status != "D" && procInfos[j].Status == "D" {
				return false
			}
			return (procInfos[i].IOReadMB + procInfos[i].IOWriteMB) >
				(procInfos[j].IOReadMB + procInfos[j].IOWriteMB)
		default:
			return procInfos[i].CPUPercent > procInfos[j].CPUPercent
		}
	})

	// Return top N processes
	if len(procInfos) > n {
		procInfos = procInfos[:n]
	}

	return procInfos, nil
}

// displayMetrics displays the collected metrics in the specified format
func (o *SysloadOptions) displayMetrics(metrics *SystemMetrics) error {
	var output string
	var err error

	switch o.Output {
	case "json":
		output, err = o.formatJSON(metrics)
	case "table":
		output, err = o.formatTable(metrics)
	default:
		return fmt.Errorf("unsupported output format: %s", o.Output)
	}

	if err != nil {
		return err
	}

	// Write to file if specified
	if o.SaveToFile != "" {
		if err := os.WriteFile(o.SaveToFile, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
		fmt.Fprintf(o.Out, "%s Output saved to: %s\n", emoji.FloppyDisk, o.SaveToFile)
		return nil
	}

	// Write to stdout
	fmt.Fprint(o.Out, output)
	return nil
}

// formatJSON formats metrics as JSON
func (o *SysloadOptions) formatJSON(metrics *SystemMetrics) (string, error) {
	// Create a custom struct based on wide mode
	type OutputMetrics struct {
		Timestamp   time.Time     `json:"timestamp"`
		CPU         interface{}   `json:"cpu"`
		Memory      MemoryMetrics `json:"memory"`
		Disk        interface{}   `json:"disk"`
		LoadAverage LoadMetrics   `json:"load_average"`
		TopCPU      []ProcessInfo `json:"top_cpu_processes"`
		TopMemory   []ProcessInfo `json:"top_memory_processes"`
		TopIO       []ProcessInfo `json:"top_io_processes,omitempty"`
		TopIOWait   []ProcessInfo `json:"top_iowait_processes,omitempty"`
	}

	output := OutputMetrics{
		Timestamp:   metrics.Timestamp,
		Memory:      metrics.Memory,
		LoadAverage: metrics.LoadAverage,
		TopCPU:      metrics.TopCPU,
		TopMemory:   metrics.TopMemory,
	}

	// Customize CPU output based on wide mode
	if o.Wide {
		output.CPU = metrics.CPU
	} else {
		// Simplified CPU metrics for non-wide mode
		output.CPU = struct {
			Usage         float64 `json:"usage_percent"`
			User          float64 `json:"user_percent"`
			System        float64 `json:"system_percent"`
			Idle          float64 `json:"idle_percent"`
			IOWait        float64 `json:"iowait_percent"`
			LoadAvg1      float64 `json:"load_avg_1min"`
			LoadAvg5      float64 `json:"load_avg_5min"`
			LoadAvg15     float64 `json:"load_avg_15min"`
			PhysicalCores int     `json:"physical_cores"`
			LogicalCores  int     `json:"logical_cores"`
			Architecture  string  `json:"architecture"`
			IsAbnormal    bool    `json:"is_abnormal"`
			IsIOWaitHigh  bool    `json:"is_iowait_high"`
		}{
			Usage:         metrics.CPU.Usage,
			User:          metrics.CPU.User,
			System:        metrics.CPU.System,
			Idle:          metrics.CPU.Idle,
			IOWait:        metrics.CPU.IOWait,
			LoadAvg1:      metrics.CPU.LoadAvg1,
			LoadAvg5:      metrics.CPU.LoadAvg5,
			LoadAvg15:     metrics.CPU.LoadAvg15,
			PhysicalCores: metrics.CPU.PhysicalCores,
			LogicalCores:  metrics.CPU.LogicalCores,
			Architecture:  metrics.CPU.Architecture,
			IsAbnormal:    metrics.CPU.IsAbnormal,
			IsIOWaitHigh:  metrics.CPU.IsIOWaitHigh,
		}
	}

	// Customize Disk output based on wide mode
	if o.Wide {
		output.Disk = metrics.Disk
	} else {
		// Simplified disk metrics for non-wide mode
		output.Disk = struct {
			Usage      float64 `json:"usage_percent"`
			TotalGB    float64 `json:"total_gb"`
			UsedGB     float64 `json:"used_gb"`
			FreeGB     float64 `json:"free_gb"`
			IsAbnormal bool    `json:"is_abnormal"`
		}{
			Usage:      metrics.Disk.Usage,
			TotalGB:    metrics.Disk.TotalGB,
			UsedGB:     metrics.Disk.UsedGB,
			FreeGB:     metrics.Disk.FreeGB,
			IsAbnormal: metrics.Disk.IsAbnormal,
		}
	}

	// Include I/O processes only in wide mode
	if o.Wide {
		output.TopIO = metrics.TopIO
		output.TopIOWait = metrics.TopIOWait
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// formatTable formats metrics as a table
func (o *SysloadOptions) formatTable(metrics *SystemMetrics) (string, error) {
	var sb strings.Builder

	// Header with timestamp
	sb.WriteString(fmt.Sprintf("%s System Load Report - %s\n\n",
		emoji.ChartIncreasing, metrics.Timestamp.Format("2006-01-02 15:04:05")))

	// System Overview
	sb.WriteString(fmt.Sprintf("%s System Overview\n", emoji.DesktopComputer))
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	// CPU metrics with enhanced information
	cpuColor := color.New(color.FgGreen)
	if metrics.CPU.IsAbnormal {
		cpuColor = color.New(color.FgRed, color.Bold)
	}

	sb.WriteString(fmt.Sprintf("Overall Usage: %s\n",
		cpuColor.Sprintf("%.1f%%", metrics.CPU.Usage)))

	// IOWait information
	iowaitColor := color.New(color.FgGreen)
	if metrics.CPU.IsIOWaitHigh {
		iowaitColor = color.New(color.FgRed, color.Bold)
	} else if metrics.CPU.IOWait > IOWaitThreshold*0.5 {
		iowaitColor = color.New(color.FgYellow)
	}
	sb.WriteString(fmt.Sprintf("IOWait:        %s\n",
		iowaitColor.Sprintf("%.1f%%", metrics.CPU.IOWait)))

	// Detailed CPU breakdown
	sb.WriteString(fmt.Sprintf("User:          %.1f%%, System: %.1f%%, Idle: %.1f%%\n",
		metrics.CPU.User, metrics.CPU.System, metrics.CPU.Idle))

	sb.WriteString(fmt.Sprintf("Cores:         %d physical, %d logical\n",
		metrics.CPU.PhysicalCores, metrics.CPU.LogicalCores))

	sb.WriteString(fmt.Sprintf("Architecture:  %s\n", metrics.CPU.Architecture))

	// Show CPU model if available (only in wide mode)
	if o.Wide && len(metrics.CPU.CPUInfo) > 0 {
		cpuInfo := metrics.CPU.CPUInfo[0]
		if cpuInfo.ModelName != "" {
			sb.WriteString(fmt.Sprintf("CPU Model:     %s\n", cpuInfo.ModelName))
		}
		if cpuInfo.MHz > 0 {
			sb.WriteString(fmt.Sprintf("CPU Speed:     %.0f MHz\n", cpuInfo.MHz))
		}
	}

	// Show per-core IOWait if available and IOWait is high (only in wide mode)
	if o.Wide && len(metrics.CPU.PerCoreIOWait) > 0 && len(metrics.CPU.PerCoreIOWait) <= 16 && metrics.CPU.IOWait > 5 {
		sb.WriteString("Per-Core IOWait: ")
		for i, iowait := range metrics.CPU.PerCoreIOWait {
			if i > 0 {
				sb.WriteString(", ")
			}
			coreColor := color.New(color.FgGreen)
			if iowait > IOWaitThreshold {
				coreColor = color.New(color.FgRed, color.Bold)
			} else if iowait > IOWaitThreshold*0.5 {
				coreColor = color.New(color.FgYellow)
			}
			sb.WriteString(coreColor.Sprintf("%.1f%%", iowait))
		}
		sb.WriteString("\n")
	}

	// Show per-core usage if available and not too many cores (only in wide mode)
	if o.Wide && len(metrics.CPU.PerCoreUsage) > 0 && len(metrics.CPU.PerCoreUsage) <= 16 {
		sb.WriteString("Per-Core Usage:  ")
		for i, usage := range metrics.CPU.PerCoreUsage {
			if i > 0 {
				sb.WriteString(", ")
			}
			coreColor := color.New(color.FgGreen)
			if usage > CPUThreshold {
				coreColor = color.New(color.FgRed, color.Bold)
			} else if usage > CPUThreshold*0.7 {
				coreColor = color.New(color.FgYellow)
			}
			sb.WriteString(coreColor.Sprintf("%.1f%%", usage))
		}
		sb.WriteString("\n")
	}

	// Load average
	loadColor := color.New(color.FgGreen)
	if metrics.LoadAverage.IsAbnormal {
		loadColor = color.New(color.FgRed, color.Bold)
	}
	sb.WriteString(fmt.Sprintf("Load Avg:      %s\n",
		loadColor.Sprintf("%.2f, %.2f, %.2f",
			metrics.LoadAverage.Load1,
			metrics.LoadAverage.Load5,
			metrics.LoadAverage.Load15)))

	// Memory metrics
	memColor := color.New(color.FgGreen)
	if metrics.Memory.IsAbnormal {
		memColor = color.New(color.FgRed, color.Bold)
	}
	sb.WriteString(fmt.Sprintf("Memory:        %s / %.1fG (%s)\n",
		memColor.Sprintf("%.1fG", metrics.Memory.UsedGB),
		metrics.Memory.TotalGB,
		memColor.Sprintf("%.1f%%", metrics.Memory.UsagePercent)))

	// Disk metrics
	diskColor := color.New(color.FgGreen)
	if metrics.Disk.IsAbnormal {
		diskColor = color.New(color.FgRed, color.Bold)
	}
	sb.WriteString(fmt.Sprintf("Disk Usage:    %s / %.1fG (%s)\n",
		diskColor.Sprintf("%.1fG", metrics.Disk.UsedGB),
		metrics.Disk.TotalGB,
		diskColor.Sprintf("%.1f%%", metrics.Disk.Usage)))

	// I/O Utilization (only show in wide mode)
	if o.Wide && metrics.Disk.IOUtil > 0 {
		ioUtilColor := color.New(color.FgGreen)
		if metrics.Disk.IOUtil > IOUtilThreshold {
			ioUtilColor = color.New(color.FgRed, color.Bold)
		} else if metrics.Disk.IOUtil > IOUtilThreshold*0.7 {
			ioUtilColor = color.New(color.FgYellow)
		}
		sb.WriteString(fmt.Sprintf("I/O Util:      %s\n",
			ioUtilColor.Sprintf("%.1f%%", metrics.Disk.IOUtil)))
	}

	// Display mode info
	if o.Wide {
		sb.WriteString(fmt.Sprintf("Mode:          Wide (showing extended metrics)\n"))
	}

	// Display process thresholds info
	sb.WriteString(fmt.Sprintf("Process Thresholds: CPU > %.1f%%, MEM > %.1f%% (highlighted in red)\n",
		o.ProcessCPUThreshold, o.ProcessMemThreshold))

	sb.WriteString("\n")

	// Top CPU processes
	sb.WriteString(o.formatProcessTable("Top CPU Processes", metrics.TopCPU, "cpu"))
	sb.WriteString("\n")

	// Top Memory processes
	sb.WriteString(o.formatProcessTable("Top Memory Processes", metrics.TopMemory, "memory"))
	sb.WriteString("\n")

	// Top I/O processes (only in wide mode)
	if o.Wide && len(metrics.TopIO) > 0 {
		sb.WriteString(o.formatProcessTable("Top I/O Processes", metrics.TopIO, "io"))
		sb.WriteString("\n")
	}

	// Top IOWait processes (only in wide mode and if IOWait is significant)
	if o.Wide && metrics.CPU.IOWait > 5 && len(metrics.TopIOWait) > 0 {
		sb.WriteString(o.formatProcessTable("Top IOWait Processes", metrics.TopIOWait, "iowait"))
	}

	return sb.String(), nil
}

// formatProcessTable formats process information as a table
func (o *SysloadOptions) formatProcessTable(title string, processes []ProcessInfo, sortBy string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s\n", emoji.Trophy, title))
	sb.WriteString(strings.Repeat("-", len(title)+4) + "\n")

	if len(processes) == 0 {
		sb.WriteString("No processes found\n")
		return sb.String()
	}

	// Create table
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	// Set headers based on sort type
	switch sortBy {
	case "cpu":
		table.SetHeader([]string{"PID", "USER", "CPU%", "MEM(MB)", "COMMAND"})
	case "memory":
		table.SetHeader([]string{"PID", "USER", "MEM(MB)", "MEM%", "COMMAND"})
	case "io":
		table.SetHeader([]string{"PID", "USER", "READ(MB)", "WRITE(MB)", "COMMAND"})
	case "iowait":
		table.SetHeader([]string{"PID", "USER", "STATE", "READ(MB)", "WRITE(MB)", "COMMAND"})
	}

	// Add rows with conditional coloring
	for _, proc := range processes {
		var row []string
		switch sortBy {
		case "cpu":
			// Color CPU% if above threshold
			cpuStr := fmt.Sprintf("%.1f", proc.CPUPercent)
			if proc.CPUPercent > o.ProcessCPUThreshold {
				cpuStr = color.New(color.FgRed, color.Bold).Sprint(cpuStr)
			}

			// Color MEM(MB) if above threshold (convert MB to percentage for comparison)
			memStr := fmt.Sprintf("%.1f", proc.MemoryMB)
			if proc.MemoryPercent > o.ProcessMemThreshold {
				memStr = color.New(color.FgRed, color.Bold).Sprint(memStr)
			}

			row = []string{
				strconv.Itoa(int(proc.PID)),
				o.truncateString(proc.Username, 10),
				cpuStr,
				memStr,
				o.truncateString(proc.Command, o.CommandWidth),
			}
		case "memory":
			// Color MEM(MB) if above threshold
			memMBStr := fmt.Sprintf("%.1f", proc.MemoryMB)
			memPercentStr := fmt.Sprintf("%.1f", proc.MemoryPercent)
			if proc.MemoryPercent > o.ProcessMemThreshold {
				memMBStr = color.New(color.FgRed, color.Bold).Sprint(memMBStr)
				memPercentStr = color.New(color.FgRed, color.Bold).Sprint(memPercentStr)
			}

			row = []string{
				strconv.Itoa(int(proc.PID)),
				o.truncateString(proc.Username, 10),
				memMBStr,
				memPercentStr,
				o.truncateString(proc.Command, o.CommandWidth),
			}
		case "io":
			row = []string{
				strconv.Itoa(int(proc.PID)),
				o.truncateString(proc.Username, 10),
				fmt.Sprintf("%.1f", proc.IOReadMB),
				fmt.Sprintf("%.1f", proc.IOWriteMB),
				o.truncateString(proc.Command, o.CommandWidth),
			}
		case "iowait":
			stateColor := ""
			if proc.Status == "D" {
				stateColor = color.New(color.FgRed, color.Bold).Sprint(proc.Status)
			} else {
				stateColor = proc.Status
			}
			row = []string{
				strconv.Itoa(int(proc.PID)),
				o.truncateString(proc.Username, 10),
				stateColor,
				fmt.Sprintf("%.1f", proc.IOReadMB),
				fmt.Sprintf("%.1f", proc.IOWriteMB),
				o.truncateString(proc.Command, o.CommandWidth),
			}
		}
		table.Append(row)
	}

	table.SetBorder(false)
	table.SetRowSeparator("")
	table.SetColumnSeparator("  ")
	table.SetHeaderLine(false)
	table.Render()

	sb.WriteString(tableString.String())
	return sb.String()
}

// truncateString truncates string to specified length with method receiver to access options
func (o *SysloadOptions) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PrintGettingStarted prints formatted success information.
func (o *SysloadOptions) PrintGettingStarted() {
	if o.SaveToFile != "" {
		return // Don't print if saving to file
	}

	fmt.Fprintf(o.Out, "\n%s System load information collected successfully\n",
		emoji.CheckMarkButton)

	if o.Watch {
		fmt.Fprintf(o.Out, "%s Monitoring will continue every %ds\n",
			emoji.Eyes, o.Interval)
	}

	// Show configuration info
	fmt.Fprintf(o.Out, "%s Configuration: Top %d processes, Command width: %d chars\n",
		emoji.Gear, o.TopN, o.CommandWidth)

	fmt.Fprintf(o.Out, "%s Process Thresholds: CPU > %.1f%%, MEM > %.1f%%\n",
		emoji.Warning, o.ProcessCPUThreshold, o.ProcessMemThreshold)

	if o.Wide {
		fmt.Fprintf(o.Out, "%s Wide mode enabled - showing extended metrics including I/O utilization\n",
			emoji.Memo)
	}
}
