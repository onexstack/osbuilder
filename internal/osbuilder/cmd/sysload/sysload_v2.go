package sysload

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// SysloadV2Options defines configuration for the "sysload" command.
type SysloadV2Options struct {
	RootDir     string // working directory
	Output      string // output format: table, json
	TopN        int    // top N processes to show
	Watch       bool   // continuous monitoring
	Interval    int    // refresh interval in seconds
	NoColor     bool   // disable colored output
	SaveToFile  string // save output to file
	Detailed    bool   // show detailed diagnostics
	SkipNetwork bool   // skip network checks

	genericiooptions.IOStreams
}

// SystemMetricsV2 contains enhanced system performance metrics
type SystemMetricsV2 struct {
	Timestamp       time.Time        `json:"timestamp"`
	Hostname        string           `json:"hostname"`
	Uptime          time.Duration    `json:"uptime"`
	KernelVersion   string           `json:"kernel_version"`
	OSInfo          OSInfo           `json:"os_info"`
	CPU             CPUMetricsV2     `json:"cpu"`
	Memory          MemoryMetricsV2  `json:"memory"`
	Disk            DiskMetricsV2    `json:"disk"`
	Network         NetworkMetrics   `json:"network"`
	LoadAverage     LoadMetricsV2    `json:"load_average"`
	ProcessStats    ProcessStats     `json:"process_stats"`
	SystemHealth    SystemHealth     `json:"system_health"`
	Performance     PerformanceStats `json:"performance"`
	TopCPU          []ProcessInfoV2  `json:"top_cpu_processes"`
	TopMemory       []ProcessInfoV2  `json:"top_memory_processes"`
	TopIO           []ProcessInfoV2  `json:"top_io_processes"`
	Alerts          []Alert          `json:"alerts"`
	Recommendations []string         `json:"recommendations"`
}

// OSInfo contains operating system information
type OSInfo struct {
	Platform        string    `json:"platform"`
	PlatformFamily  string    `json:"platform_family"`
	PlatformVersion string    `json:"platform_version"`
	Architecture    string    `json:"architecture"`
	BootTime        time.Time `json:"boot_time"`
}

// Enhanced CPUMetricsV2 with detailed breakdown
type CPUMetricsV2 struct {
	Usage           float64   `json:"usage_percent"`
	UserPercent     float64   `json:"user_percent"`
	SystemPercent   float64   `json:"system_percent"`
	IdlePercent     float64   `json:"idle_percent"`
	IOWaitPercent   float64   `json:"iowait_percent"`
	StealPercent    float64   `json:"steal_percent"`
	NicePercent     float64   `json:"nice_percent"`
	IrqPercent      float64   `json:"irq_percent"`
	SoftIrqPercent  float64   `json:"softirq_percent"`
	LoadAvg1        float64   `json:"load_avg_1min"`
	LoadAvg5        float64   `json:"load_avg_5min"`
	LoadAvg15       float64   `json:"load_avg_15min"`
	Cores           int       `json:"cores"`
	LogicalCores    int       `json:"logical_cores"`
	ContextSwitches uint64    `json:"context_switches"`
	Interrupts      uint64    `json:"interrupts"`
	CPUFreq         float64   `json:"cpu_freq_mhz"`
	Temperature     []float64 `json:"temperatures"`
	IsAbnormal      bool      `json:"is_abnormal"`
	Issues          []string  `json:"issues"`
}

// Enhanced MemoryMetricsV2 with detailed breakdown
type MemoryMetricsV2 struct {
	Total        uint64  `json:"total_bytes"`
	Used         uint64  `json:"used_bytes"`
	Available    uint64  `json:"available_bytes"`
	Free         uint64  `json:"free_bytes"`
	Cached       uint64  `json:"cached_bytes"`
	Buffers      uint64  `json:"buffers_bytes"`
	Shared       uint64  `json:"shared_bytes"`
	Active       uint64  `json:"active_bytes"`
	Inactive     uint64  `json:"inactive_bytes"`
	Dirty        uint64  `json:"dirty_bytes"`
	Writeback    uint64  `json:"writeback_bytes"`
	Slab         uint64  `json:"slab_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	AvailableGB  float64 `json:"available_gb"`
	CachedGB     float64 `json:"cached_gb"`
	BuffersGB    float64 `json:"buffers_gb"`

	// Swap information
	SwapTotal   uint64  `json:"swap_total_bytes"`
	SwapUsed    uint64  `json:"swap_used_bytes"`
	SwapFree    uint64  `json:"swap_free_bytes"`
	SwapPercent float64 `json:"swap_percent"`
	SwapTotalGB float64 `json:"swap_total_gb"`
	SwapUsedGB  float64 `json:"swap_used_gb"`

	// Memory statistics
	PageFaults      uint64 `json:"page_faults"`
	MajorPageFaults uint64 `json:"major_page_faults"`
	PageIns         uint64 `json:"page_ins"`
	PageOuts        uint64 `json:"page_outs"`
	SwapIns         uint64 `json:"swap_ins"`
	SwapOuts        uint64 `json:"swap_outs"`

	IsAbnormal bool     `json:"is_abnormal"`
	Issues     []string `json:"issues"`
}

// Enhanced DiskMetricsV2 with detailed I/O statistics
type DiskMetricsV2 struct {
	Partitions      []PartitionInfo `json:"partitions"`
	IOStats         []DiskIOStats   `json:"io_stats"`
	TotalReadMB     float64         `json:"total_read_mb"`
	TotalWriteMB    float64         `json:"total_write_mb"`
	TotalReadIOPS   uint64          `json:"total_read_iops"`
	TotalWriteIOPS  uint64          `json:"total_write_iops"`
	AvgReadLatency  float64         `json:"avg_read_latency_ms"`
	AvgWriteLatency float64         `json:"avg_write_latency_ms"`
	QueueDepth      float64         `json:"avg_queue_depth"`
	IsAbnormal      bool            `json:"is_abnormal"`
	Issues          []string        `json:"issues"`
}

// PartitionInfo contains disk partition details
type PartitionInfo struct {
	Device        string   `json:"device"`
	Mountpoint    string   `json:"mountpoint"`
	Filesystem    string   `json:"filesystem"`
	Options       string   `json:"options"`
	UsagePercent  float64  `json:"usage_percent"`
	TotalGB       float64  `json:"total_gb"`
	UsedGB        float64  `json:"used_gb"`
	FreeGB        float64  `json:"free_gb"`
	InodesTotal   uint64   `json:"inodes_total"`
	InodesUsed    uint64   `json:"inodes_used"`
	InodesFree    uint64   `json:"inodes_free"`
	InodesPercent float64  `json:"inodes_percent"`
	ReadOnly      bool     `json:"read_only"`
	Issues        []string `json:"issues"`
}

// DiskIOStats contains detailed disk I/O statistics
type DiskIOStats struct {
	Device         string   `json:"device"`
	ReadCount      uint64   `json:"read_count"`
	WriteCount     uint64   `json:"write_count"`
	ReadBytes      uint64   `json:"read_bytes"`
	WriteBytes     uint64   `json:"write_bytes"`
	ReadTime       uint64   `json:"read_time_ms"`
	WriteTime      uint64   `json:"write_time_ms"`
	IOTime         uint64   `json:"io_time_ms"`
	WeightedIOTime uint64   `json:"weighted_io_time_ms"`
	ReadMerged     uint64   `json:"read_merged"`
	WriteMerged    uint64   `json:"write_merged"`
	DiscardCount   uint64   `json:"discard_count"`
	DiscardBytes   uint64   `json:"discard_bytes"`
	DiscardTime    uint64   `json:"discard_time_ms"`
	ReadLatencyMS  float64  `json:"read_latency_ms"`
	WriteLatencyMS float64  `json:"write_latency_ms"`
	Utilization    float64  `json:"utilization_percent"`
	QueueDepth     float64  `json:"queue_depth"`
	Issues         []string `json:"issues"`
}

// NetworkMetrics contains detailed network statistics
type NetworkMetrics struct {
	Interfaces          []NetworkInterface `json:"interfaces"`
	Connections         []ConnectionInfo   `json:"connections"`
	ListenPorts         []PortInfo         `json:"listen_ports"`
	TotalRxBytes        uint64             `json:"total_rx_bytes"`
	TotalTxBytes        uint64             `json:"total_tx_bytes"`
	TotalRxPackets      uint64             `json:"total_rx_packets"`
	TotalTxPackets      uint64             `json:"total_tx_packets"`
	TotalRxErrors       uint64             `json:"total_rx_errors"`
	TotalTxErrors       uint64             `json:"total_tx_errors"`
	TotalRxDropped      uint64             `json:"total_rx_dropped"`
	TotalTxDropped      uint64             `json:"total_tx_dropped"`
	TotalRxMBPS         float64            `json:"total_rx_mbps"`
	TotalTxMBPS         float64            `json:"total_tx_mbps"`
	ActiveConnections   int                `json:"active_connections"`
	TimeWaitConnections int                `json:"time_wait_connections"`
	TCPRetransmits      uint64             `json:"tcp_retransmits"`
	TCPResets           uint64             `json:"tcp_resets"`
	IsAbnormal          bool               `json:"is_abnormal"`
	Issues              []string           `json:"issues"`
}

// NetworkInterface contains detailed interface statistics
type NetworkInterface struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"hardware_addr"`
	MTU          int      `json:"mtu"`
	Flags        []string `json:"flags"`
	Addrs        []string `json:"addresses"`
	BytesSent    uint64   `json:"bytes_sent"`
	BytesRecv    uint64   `json:"bytes_recv"`
	PacketsSent  uint64   `json:"packets_sent"`
	PacketsRecv  uint64   `json:"packets_recv"`
	ErrIn        uint64   `json:"err_in"`
	ErrOut       uint64   `json:"err_out"`
	DropIn       uint64   `json:"drop_in"`
	DropOut      uint64   `json:"drop_out"`
	FifoIn       uint64   `json:"fifo_in"`
	FifoOut      uint64   `json:"fifo_out"`
	RxMBPS       float64  `json:"rx_mbps"`
	TxMBPS       float64  `json:"tx_mbps"`
	ErrorRate    float64  `json:"error_rate_percent"`
	DropRate     float64  `json:"drop_rate_percent"`
	IsUp         bool     `json:"is_up"`
	Issues       []string `json:"issues"`
}

// ConnectionInfo contains network connection details
type ConnectionInfo struct {
	Type       string `json:"type"`
	LocalAddr  string `json:"local_addr"`
	LocalPort  uint32 `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort uint32 `json:"remote_port"`
	Status     string `json:"status"`
	PID        int32  `json:"pid"`
	Process    string `json:"process"`
	Family     uint32 `json:"family"`
}

// PortInfo contains listening port information
type PortInfo struct {
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
	Process  string `json:"process"`
	PID      int32  `json:"pid"`
	Address  string `json:"address"`
}

// LoadMetricsV2 contains enhanced system load information
type LoadMetricsV2 struct {
	Load1         float64  `json:"load1"`
	Load5         float64  `json:"load5"`
	Load15        float64  `json:"load15"`
	RunningProcs  uint64   `json:"running_processes"`
	TotalProcs    uint64   `json:"total_processes"`
	BlockedProcs  uint64   `json:"blocked_processes"`
	ZombieProcs   uint64   `json:"zombie_processes"`
	LoadPerCore1  float64  `json:"load_per_core_1min"`
	LoadPerCore5  float64  `json:"load_per_core_5min"`
	LoadPerCore15 float64  `json:"load_per_core_15min"`
	CPUSaturation float64  `json:"cpu_saturation_percent"`
	IsAbnormal    bool     `json:"is_abnormal"`
	Issues        []string `json:"issues"`
}

// ProcessStats contains detailed process statistics
type ProcessStats struct {
	Total            int      `json:"total"`
	Running          int      `json:"running"`
	Sleeping         int      `json:"sleeping"`
	Stopped          int      `json:"stopped"`
	Zombie           int      `json:"zombie"`
	Idle             int      `json:"idle"`
	Wait             int      `json:"wait"`
	Locked           int      `json:"locked"`
	HighCPUProcesses int      `json:"high_cpu_processes"`
	HighMemProcesses int      `json:"high_memory_processes"`
	HighIOProcesses  int      `json:"high_io_processes"`
	ThreadsTotal     int      `json:"threads_total"`
	FDsTotal         int      `json:"file_descriptors_total"`
	FDsLimit         int      `json:"file_descriptors_limit"`
	FDsUsagePercent  float64  `json:"file_descriptors_usage_percent"`
	Issues           []string `json:"issues"`
}

// SystemHealth contains overall system health indicators
type SystemHealth struct {
	Status              string           `json:"status"` // healthy, warning, critical
	Score               int              `json:"score"`  // 0-100
	Issues              []string         `json:"issues"`
	Warnings            []string         `json:"warnings"`
	CriticalAlerts      []string         `json:"critical_alerts"`
	SystemLimits        SystemLimits     `json:"system_limits"`
	ResourcePressure    ResourcePressure `json:"resource_pressure"`
	StabilityIndicators StabilityMetrics `json:"stability_indicators"`
}

// SystemLimits contains system resource limits
type SystemLimits struct {
	MaxFiles            uint64  `json:"max_files"`
	MaxProcesses        uint64  `json:"max_processes"`
	MaxMemoryGB         float64 `json:"max_memory_gb"`
	MaxFileSize         uint64  `json:"max_file_size"`
	MaxOpenFiles        uint64  `json:"max_open_files"`
	MaxThreads          uint64  `json:"max_threads"`
	MaxSockets          uint64  `json:"max_sockets"`
	CurrentFiles        uint64  `json:"current_files"`
	CurrentProcesses    uint64  `json:"current_processes"`
	CurrentThreads      uint64  `json:"current_threads"`
	FileUsagePercent    float64 `json:"file_usage_percent"`
	ProcessUsagePercent float64 `json:"process_usage_percent"`
	ThreadUsagePercent  float64 `json:"thread_usage_percent"`
}

// ResourcePressure indicates system resource pressure
type ResourcePressure struct {
	CPUPressure     PressureInfo `json:"cpu_pressure"`
	MemoryPressure  PressureInfo `json:"memory_pressure"`
	IOPressure      PressureInfo `json:"io_pressure"`
	OverallPressure string       `json:"overall_pressure"` // low, medium, high, critical
}

// PressureInfo contains PSI (Pressure Stall Information) metrics
type PressureInfo struct {
	Some10s  float64 `json:"some_10s"`
	Some60s  float64 `json:"some_60s"`
	Some300s float64 `json:"some_300s"`
	Full10s  float64 `json:"full_10s"`
	Full60s  float64 `json:"full_60s"`
	Full300s float64 `json:"full_300s"`
	Level    string  `json:"level"` // low, medium, high, critical
}

// StabilityMetrics contains system stability indicators
type StabilityMetrics struct {
	UptimeDays         float64 `json:"uptime_days"`
	KernelErrors       uint64  `json:"kernel_errors"`
	OOMKills           uint64  `json:"oom_kills"`
	MemoryLeaks        int     `json:"suspected_memory_leaks"`
	DeadlockDetections int     `json:"deadlock_detections"`
	CorruptedFiles     int     `json:"corrupted_files"`
	HardwareErrors     int     `json:"hardware_errors"`
	RecentCrashes      int     `json:"recent_crashes"`
	StabilityScore     int     `json:"stability_score"` // 0-100
}

// PerformanceStats contains performance-related statistics
type PerformanceStats struct {
	ResponseTime     ResponseTimeStats `json:"response_time"`
	Throughput       ThroughputStats   `json:"throughput"`
	Bottlenecks      []Bottleneck      `json:"bottlenecks"`
	PerformanceScore int               `json:"performance_score"` // 0-100
	OptimizationTips []string          `json:"optimization_tips"`
}

// ResponseTimeStats contains response time metrics
type ResponseTimeStats struct {
	AvgCPUScheduling float64 `json:"avg_cpu_scheduling_ms"`
	AvgDiskIO        float64 `json:"avg_disk_io_ms"`
	AvgNetworkIO     float64 `json:"avg_network_io_ms"`
	AvgMemoryAccess  float64 `json:"avg_memory_access_ms"`
	P50ResponseTime  float64 `json:"p50_response_time_ms"`
	P95ResponseTime  float64 `json:"p95_response_time_ms"`
	P99ResponseTime  float64 `json:"p99_response_time_ms"`
}

// ThroughputStats contains throughput metrics
type ThroughputStats struct {
	CPUInstructionsPerSec uint64  `json:"cpu_instructions_per_sec"`
	DiskIOPS              uint64  `json:"disk_iops"`
	NetworkPPS            uint64  `json:"network_packets_per_sec"`
	MemoryBandwidthMBPS   float64 `json:"memory_bandwidth_mbps"`
	SystemCallsPerSec     uint64  `json:"system_calls_per_sec"`
}

// Bottleneck represents a system bottleneck
type Bottleneck struct {
	Component   string  `json:"component"` // cpu, memory, disk, network
	Severity    string  `json:"severity"`  // low, medium, high, critical
	Impact      float64 `json:"impact_percent"`
	Description string  `json:"description"`
	Cause       string  `json:"cause"`
	Solution    string  `json:"solution"`
}

// Enhanced ProcessInfoV2 with more details
type ProcessInfoV2 struct {
	PID             int32     `json:"pid"`
	PPID            int32     `json:"ppid"`
	Name            string    `json:"name"`
	Username        string    `json:"username"`
	Status          string    `json:"status"`
	CreateTime      time.Time `json:"create_time"`
	CPUPercent      float64   `json:"cpu_percent"`
	CPUTime         float64   `json:"cpu_time_seconds"`
	MemoryMB        float64   `json:"memory_mb"`
	VirtualMemoryMB float64   `json:"virtual_memory_mb"`
	MemoryPercent   float64   `json:"memory_percent"`
	Command         string    `json:"command"`
	Cmdline         string    `json:"cmdline"`
	Cwd             string    `json:"cwd"`
	Exe             string    `json:"exe"`
	IOReadMB        float64   `json:"io_read_mb"`
	IOWriteMB       float64   `json:"io_write_mb"`
	IOReadCount     uint64    `json:"io_read_count"`
	IOWriteCount    uint64    `json:"io_write_count"`
	NetConnections  int32     `json:"net_connections"`
	OpenFiles       int32     `json:"open_files"`
	NumThreads      int32     `json:"num_threads"`
	NumCtxSwitches  uint64    `json:"num_ctx_switches"`
	Priority        int32     `json:"priority"`
	Nice            int32     `json:"nice"`
	IsRunning       bool      `json:"is_running"`
	Issues          []string  `json:"issues"`
}

// Alert represents a system alert
type Alert struct {
	Level        string    `json:"level"`     // info, warning, critical
	Component    string    `json:"component"` // cpu, memory, disk, network, process
	Metric       string    `json:"metric"`
	CurrentValue float64   `json:"current_value"`
	Threshold    float64   `json:"threshold"`
	Message      string    `json:"message"`
	Impact       string    `json:"impact"`
	Action       string    `json:"recommended_action"`
	Timestamp    time.Time `json:"timestamp"`
}

var (
	sysloadLongDescV2 = templates.LongDesc(`Display comprehensive system load information for advanced performance troubleshooting.

    This enhanced command provides detailed system performance metrics, health checks, and diagnostic
    information to help identify and resolve system issues including:

    System Metrics:
    - CPU usage breakdown (user/system/iowait/idle), context switches, interrupts
    - Memory usage, swap, cache/buffer statistics, page faults
    - Disk I/O performance, latency, IOPS, queue depth per device
    - Network interface statistics, connection analysis, error rates
    - Process states, file descriptors, system limits

    Health Checks:
    - Resource utilization with customizable thresholds
    - System pressure indicators (CPU/Memory/IO pressure)
    - Process analysis (zombie, blocked, high-resource processes)
    - System stability metrics and error detection
    - Performance bottleneck identification
    - Actionable recommendations for optimization

    Abnormal conditions are highlighted with detailed diagnostics and solutions.`)

	sysloadExampleV2 = templates.Examples(`# Show comprehensive system analysis
        osbuilder sysload --detailed

        # Monitor with custom process count
        osbuilder sysload --top 15 --watch --interval 3

        # Skip network analysis for faster execution
        osbuilder sysload --skip-network

        # Export detailed diagnostics to JSON
        osbuilder sysload --detailed --output json --save-to-file diagnostics.json

        # Continuous monitoring with alerts
        osbuilder sysload --watch --detailed > system-monitor.log`)
)

// Enhanced thresholds for abnormal conditions
const (
	CPUThresholdV2      = 80.0   // CPU usage > 80%
	IOWaitThresholdV2   = 20.0   // IOWait > 20%
	MemoryThresholdV2   = 85.0   // Memory usage > 85%
	SwapThreshold       = 10.0   // Swap usage > 10%
	DiskThresholdV2     = 90.0   // Disk usage > 90%
	InodeThreshold      = 90.0   // Inode usage > 90%
	LoadThresholdV2     = 2.0    // Load > cores * 2
	NetworkErrorRate    = 1.0    // Network error rate > 1%
	FDThreshold         = 80.0   // File descriptor usage > 80%
	ContextSwitchRate   = 100000 // Context switches > 100k/sec
	InterruptRate       = 50000  // Interrupts > 50k/sec
	ZombieProcessLimit  = 10     // Zombie processes > 10
	HighCPUProcessLimit = 5      // Processes using > 80% CPU
	HighMemProcessLimit = 3      // Processes using > 20% memory
)

// NewSysloadV2Cmd creates the enhanced "sysload" command.
func NewSysloadV2Cmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &SysloadV2Options{
		IOStreams: ioStreams,
		Output:    "table",
		TopN:      5,
		Interval:  5,
	}

	cmd := &cobra.Command{
		Use:                   "sysloadv2",
		Short:                 "Advanced system diagnostics and performance troubleshooting",
		Long:                  sysloadLongDescV2,
		Example:               sysloadExampleV2,
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
	cmd.Flags().BoolVarP(&opts.Detailed, "detailed", "d", false, "Show detailed diagnostics and health checks")
	cmd.Flags().BoolVar(&opts.SkipNetwork, "skip-network", false, "Skip network analysis for faster execution")

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *SysloadV2Options) Complete() error {
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
func (o *SysloadV2Options) Validate() error {
	// Validate output format
	if o.Output != "table" && o.Output != "json" {
		return fmt.Errorf("invalid output format %q, supported formats: table, json", o.Output)
	}

	// Validate top N
	if o.TopN < 1 || o.TopN > 100 {
		return fmt.Errorf("top N must be between 1 and 100, got %d", o.TopN)
	}

	// Validate interval
	if o.Interval < 1 || o.Interval > 3600 {
		return fmt.Errorf("interval must be between 1 and 3600 seconds, got %d", o.Interval)
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
func (o *SysloadV2Options) Run() error {
	if o.Watch {
		return o.watchMode()
	}
	return o.snapshotMode()
}

// snapshotMode displays current system load once
func (o *SysloadV2Options) snapshotMode() error {
	metrics, err := o.collectSystemMetricsV2()
	if err != nil {
		return fmt.Errorf("failed to collect system metrics: %w", err)
	}

	return o.displayMetrics(metrics)
}

// watchMode continuously monitors system load
func (o *SysloadV2Options) watchMode() error {
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

// collectSystemMetricsV2 gathers comprehensive system performance metrics
func (o *SysloadV2Options) collectSystemMetricsV2() (*SystemMetricsV2, error) {
	metrics := &SystemMetricsV2{
		Timestamp: time.Now(),
	}

	// Collect host information
	if err := o.collectHostInfo(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect host info: %w", err)
	}

	// Collect CPU metrics
	if err := o.collectCPUMetricsV2(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect CPU metrics: %w", err)
	}

	// Collect memory metrics
	if err := o.collectMemoryMetricsV2(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	// Collect disk metrics
	if err := o.collectDiskMetricsV2(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect disk metrics: %w", err)
	}

	// Collect network metrics (unless skipped)
	if !o.SkipNetwork {
		if err := o.collectNetworkMetrics(metrics); err != nil {
			// Don't fail on network errors, just log them
			fmt.Fprintf(os.Stderr, "Warning: failed to collect network metrics: %v\n", err)
		}
	}

	// Collect load metrics
	if err := o.collectLoadMetricsV2(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect load metrics: %w", err)
	}

	// Collect process statistics
	if err := o.collectProcessStats(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect process stats: %w", err)
	}

	// Collect top processes
	if err := o.collectTopProcesses(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect top processes: %w", err)
	}

	// Analyze system health and performance
	if o.Detailed {
		o.analyzeSystemHealth(metrics)
		o.analyzePerformance(metrics)
		o.generateRecommendations(metrics)
	}

	// Generate alerts
	o.generateAlerts(metrics)

	return metrics, nil
}

// collectHostInfo collects host and OS information
func (o *SysloadV2Options) collectHostInfo(metrics *SystemMetricsV2) error {
	hostInfo, err := host.Info()
	if err != nil {
		return err
	}

	metrics.Hostname = hostInfo.Hostname
	metrics.Uptime = time.Duration(hostInfo.Uptime) * time.Second
	metrics.KernelVersion = hostInfo.KernelVersion

	metrics.OSInfo = OSInfo{
		Platform:        hostInfo.Platform,
		PlatformFamily:  hostInfo.PlatformFamily,
		PlatformVersion: hostInfo.PlatformVersion,
		Architecture:    hostInfo.KernelArch,
		BootTime:        time.Unix(int64(hostInfo.BootTime), 0),
	}

	return nil
}

// collectCPUMetricsV2 collects detailed CPU performance metrics
func (o *SysloadV2Options) collectCPUMetricsV2(metrics *SystemMetricsV2) error {
	// Get CPU percentage with breakdown
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return err
	}

	if len(cpuTimes) == 0 {
		return fmt.Errorf("no CPU time data available")
	}

	cpuTime := cpuTimes[0]
	totalTime := cpuTime.User + cpuTime.System + cpuTime.Idle + cpuTime.Nice +
		cpuTime.Iowait + cpuTime.Irq + cpuTime.Softirq + cpuTime.Steal

	metrics.CPU = CPUMetricsV2{
		UserPercent:    (cpuTime.User / totalTime) * 100,
		SystemPercent:  (cpuTime.System / totalTime) * 100,
		IdlePercent:    (cpuTime.Idle / totalTime) * 100,
		IOWaitPercent:  (cpuTime.Iowait / totalTime) * 100,
		StealPercent:   (cpuTime.Steal / totalTime) * 100,
		NicePercent:    (cpuTime.Nice / totalTime) * 100,
		IrqPercent:     (cpuTime.Irq / totalTime) * 100,
		SoftIrqPercent: (cpuTime.Softirq / totalTime) * 100,
	}

	// Calculate overall usage
	metrics.CPU.Usage = 100 - metrics.CPU.IdlePercent

	// Get CPU counts
	physicalCores, err := cpu.Counts(false)
	if err != nil {
		physicalCores = 1
	}
	logicalCores, err := cpu.Counts(true)
	if err != nil {
		logicalCores = physicalCores
	}

	metrics.CPU.Cores = physicalCores
	metrics.CPU.LogicalCores = logicalCores

	// Get load average
	loadAvg, err := load.Avg()
	if err != nil {
		return err
	}

	metrics.CPU.LoadAvg1 = loadAvg.Load1
	metrics.CPU.LoadAvg5 = loadAvg.Load5
	metrics.CPU.LoadAvg15 = loadAvg.Load15

	// Get CPU frequency
	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		metrics.CPU.CPUFreq = cpuInfo[0].Mhz
	}

	// Get CPU temperatures (if available)
	metrics.CPU.Temperature = o.getCPUTemperatures()

	// Check for abnormal conditions
	metrics.CPU.IsAbnormal = o.checkCPUAbnormalities(&metrics.CPU)

	return nil
}

// collectMemoryMetricsV2 collects detailed memory statistics
func (o *SysloadV2Options) collectMemoryMetricsV2(metrics *SystemMetricsV2) error {
	// Get virtual memory statistics
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	metrics.Memory = MemoryMetricsV2{
		Total:        vmStat.Total,
		Used:         vmStat.Used,
		Available:    vmStat.Available,
		Free:         vmStat.Free,
		Cached:       vmStat.Cached,
		Buffers:      vmStat.Buffers,
		Shared:       vmStat.Shared,
		Active:       vmStat.Active,
		Inactive:     vmStat.Inactive,
		UsagePercent: vmStat.UsedPercent,
		TotalGB:      float64(vmStat.Total) / 1024 / 1024 / 1024,
		UsedGB:       float64(vmStat.Used) / 1024 / 1024 / 1024,
		AvailableGB:  float64(vmStat.Available) / 1024 / 1024 / 1024,
		CachedGB:     float64(vmStat.Cached) / 1024 / 1024 / 1024,
		BuffersGB:    float64(vmStat.Buffers) / 1024 / 1024 / 1024,
	}

	// Get swap statistics
	swapStat, err := mem.SwapMemory()
	if err == nil {
		metrics.Memory.SwapTotal = swapStat.Total
		metrics.Memory.SwapUsed = swapStat.Used
		metrics.Memory.SwapFree = swapStat.Free
		metrics.Memory.SwapPercent = swapStat.UsedPercent
		metrics.Memory.SwapTotalGB = float64(swapStat.Total) / 1024 / 1024 / 1024
		metrics.Memory.SwapUsedGB = float64(swapStat.Used) / 1024 / 1024 / 1024
	}

	// Get additional memory statistics from /proc/meminfo
	o.collectExtendedMemoryStats(&metrics.Memory)

	// Check for abnormal conditions
	metrics.Memory.IsAbnormal = o.checkMemoryAbnormalities(&metrics.Memory)

	return nil
}

// collectDiskMetricsV2 collects detailed disk I/O statistics
func (o *SysloadV2Options) collectDiskMetricsV2(metrics *SystemMetricsV2) error {
	metrics.Disk = DiskMetricsV2{}

	// Get all partitions
	partitions, err := disk.Partitions(true)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		partInfo := o.collectPartitionInfo(partition)
		metrics.Disk.Partitions = append(metrics.Disk.Partitions, partInfo)
	}

	// Get disk I/O statistics
	ioStats, err := disk.IOCounters()
	if err != nil {
		return err
	}

	var totalReadBytes, totalWriteBytes uint64
	var totalReadCount, totalWriteCount uint64

	for device, stat := range ioStats {
		diskStat := DiskIOStats{
			Device:         device,
			ReadCount:      stat.ReadCount,
			WriteCount:     stat.WriteCount,
			ReadBytes:      stat.ReadBytes,
			WriteBytes:     stat.WriteBytes,
			ReadTime:       stat.ReadTime,
			WriteTime:      stat.WriteTime,
			IOTime:         stat.IoTime,
			WeightedIOTime: stat.WeightedIO,
			ReadMerged:     stat.MergedReadCount,
			WriteMerged:    stat.MergedWriteCount,
		}

		// Calculate latencies and utilization
		if stat.ReadCount > 0 {
			diskStat.ReadLatencyMS = float64(stat.ReadTime) / float64(stat.ReadCount)
		}
		if stat.WriteCount > 0 {
			diskStat.WriteLatencyMS = float64(stat.WriteTime) / float64(stat.WriteCount)
		}

		// Calculate utilization as percentage of time spent doing I/O
		diskStat.Utilization = float64(stat.IoTime) / 10000 // Convert to percentage
		if diskStat.Utilization > 100 {
			diskStat.Utilization = 100
		}

		// Check for disk issues
		diskStat.Issues = o.checkDiskIssues(&diskStat)

		metrics.Disk.IOStats = append(metrics.Disk.IOStats, diskStat)

		// Accumulate totals
		totalReadBytes += stat.ReadBytes
		totalWriteBytes += stat.WriteBytes
		totalReadCount += stat.ReadCount
		totalWriteCount += stat.WriteCount
	}

	metrics.Disk.TotalReadMB = float64(totalReadBytes) / 1024 / 1024
	metrics.Disk.TotalWriteMB = float64(totalWriteBytes) / 1024 / 1024
	metrics.Disk.TotalReadIOPS = totalReadCount
	metrics.Disk.TotalWriteIOPS = totalWriteCount

	// Check for overall disk abnormalities
	metrics.Disk.IsAbnormal = o.checkDiskAbnormalities(&metrics.Disk)

	return nil
}

// collectNetworkMetrics collects detailed network statistics
func (o *SysloadV2Options) collectNetworkMetrics(metrics *SystemMetricsV2) error {
	metrics.Network = NetworkMetrics{}

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	// Get network I/O statistics
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	// Create a map for quick lookup
	ioMap := make(map[string]net.IOCountersStat)
	for _, counter := range ioCounters {
		ioMap[counter.Name] = counter
	}

	var totalRxBytes, totalTxBytes uint64
	var totalRxPackets, totalTxPackets uint64
	var totalRxErrors, totalTxErrors uint64
	var totalRxDropped, totalTxDropped uint64

	for _, iface := range interfaces {
		netIface := NetworkInterface{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr,
			MTU:          iface.MTU,
			Flags:        o.parseNetworkFlags(iface.Flags),
		}

		// Get interface addresses
		for _, addr := range iface.Addrs {
			netIface.Addrs = append(netIface.Addrs, addr.String())
		}

		// Get I/O statistics for this interface
		if counter, exists := ioMap[iface.Name]; exists {
			netIface.BytesSent = counter.BytesSent
			netIface.BytesRecv = counter.BytesRecv
			netIface.PacketsSent = counter.PacketsSent
			netIface.PacketsRecv = counter.PacketsRecv
			netIface.ErrIn = counter.Errin
			netIface.ErrOut = counter.Errout
			netIface.DropIn = counter.Dropin
			netIface.DropOut = counter.Dropout
			netIface.FifoIn = counter.Fifoin
			netIface.FifoOut = counter.Fifoout

			// Calculate rates and error percentages
			totalPackets := counter.PacketsRecv + counter.PacketsSent
			if totalPackets > 0 {
				netIface.ErrorRate = float64(counter.Errin+counter.Errout) / float64(totalPackets) * 100
				netIface.DropRate = float64(counter.Dropin+counter.Dropout) / float64(totalPackets) * 100
			}

			// Accumulate totals
			totalRxBytes += counter.BytesRecv
			totalTxBytes += counter.BytesSent
			totalRxPackets += counter.PacketsRecv
			totalTxPackets += counter.PacketsSent
			totalRxErrors += counter.Errin
			totalTxErrors += counter.Errout
			totalRxDropped += counter.Dropin
			totalTxDropped += counter.Dropout
		}

		// Check if interface is up
		netIface.IsUp = o.isInterfaceUp(iface.Flags)

		// Check for interface issues
		netIface.Issues = o.checkNetworkInterfaceIssues(&netIface)

		metrics.Network.Interfaces = append(metrics.Network.Interfaces, netIface)
	}

	// Set totals
	metrics.Network.TotalRxBytes = totalRxBytes
	metrics.Network.TotalTxBytes = totalTxBytes
	metrics.Network.TotalRxPackets = totalRxPackets
	metrics.Network.TotalTxPackets = totalTxPackets
	metrics.Network.TotalRxErrors = totalRxErrors
	metrics.Network.TotalTxErrors = totalTxErrors
	metrics.Network.TotalRxDropped = totalRxDropped
	metrics.Network.TotalTxDropped = totalTxDropped

	// Get network connections
	o.collectNetworkConnections(&metrics.Network)

	// Get listening ports
	o.collectListeningPorts(&metrics.Network)

	// Check for network abnormalities
	metrics.Network.IsAbnormal = o.checkNetworkAbnormalities(&metrics.Network)

	return nil
}

// collectLoadMetricsV2 collects enhanced load average metrics
func (o *SysloadV2Options) collectLoadMetricsV2(metrics *SystemMetricsV2) error {
	loadAvg, err := load.Avg()
	if err != nil {
		return err
	}

	// Get misc stats for process counts
	miscStat, err := load.Misc()
	if err == nil {
		metrics.LoadAverage = LoadMetricsV2{
			Load1:        loadAvg.Load1,
			Load5:        loadAvg.Load5,
			Load15:       loadAvg.Load15,
			RunningProcs: uint64(miscStat.ProcsRunning),
			TotalProcs:   uint64(miscStat.ProcsTotal),
			BlockedProcs: uint64(miscStat.ProcsBlocked),
		}
	} else {
		metrics.LoadAverage = LoadMetricsV2{
			Load1:  loadAvg.Load1,
			Load5:  loadAvg.Load5,
			Load15: loadAvg.Load15,
		}
	}

	// Calculate load per core
	cores := float64(metrics.CPU.Cores)
	if cores > 0 {
		metrics.LoadAverage.LoadPerCore1 = loadAvg.Load1 / cores
		metrics.LoadAverage.LoadPerCore5 = loadAvg.Load5 / cores
		metrics.LoadAverage.LoadPerCore15 = loadAvg.Load15 / cores
	}

	// Calculate CPU saturation
	metrics.LoadAverage.CPUSaturation = (loadAvg.Load1 / cores) * 100
	if metrics.LoadAverage.CPUSaturation > 100 {
		metrics.LoadAverage.CPUSaturation = 100
	}

	// Count zombie processes
	if processes, err := process.Processes(); err == nil {
		var zombies uint64
		for _, proc := range processes {
			if status, err := proc.Status(); err == nil {
				// status is []string, check each status string
				for _, s := range status {
					if strings.Contains(s, "Z") || strings.Contains(s, "zombie") {
						zombies++
						break // Found zombie status, no need to check other status strings for this process
					}
				}
			}
		}
		metrics.LoadAverage.ZombieProcs = zombies
	}

	// Check for load abnormalities
	metrics.LoadAverage.IsAbnormal = o.checkLoadAbnormalities(&metrics.LoadAverage, metrics.CPU.Cores)

	return nil
}

// collectProcessStats collects detailed process statistics
func (o *SysloadV2Options) collectProcessStats(metrics *SystemMetricsV2) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	metrics.ProcessStats = ProcessStats{}
	statusCounts := make(map[string]int)
	var totalThreads, totalFDs int
	var highCPU, highMem, highIO int

	for _, proc := range processes {
		metrics.ProcessStats.Total++

		// Get process status
		if statusSlice, err := proc.Status(); err == nil {
			// Process each status in the slice
			for _, status := range statusSlice {
				statusCounts[status]++

				// Check status and increment appropriate counters
				switch strings.ToUpper(status) {
				case "R", "RUNNING":
					metrics.ProcessStats.Running++
				case "S", "SLEEPING":
					metrics.ProcessStats.Sleeping++
				case "T", "STOPPED":
					metrics.ProcessStats.Stopped++
				case "Z", "ZOMBIE":
					metrics.ProcessStats.Zombie++
				case "I", "IDLE":
					metrics.ProcessStats.Idle++
				case "D", "DISK-SLEEP":
					metrics.ProcessStats.Wait++
				}
			}
		}

		// Count threads
		if numThreads, err := proc.NumThreads(); err == nil {
			totalThreads += int(numThreads)
		}

		// Count file descriptors
		if openFiles, err := proc.OpenFiles(); err == nil {
			totalFDs += len(openFiles)
		}

		// Check for high resource usage
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			if cpuPercent > CPUThresholdV2 {
				highCPU++
			}
		}

		if memPercent, err := proc.MemoryPercent(); err == nil {
			if float64(memPercent) > 20.0 { // Consider >20% as high memory usage
				highMem++
			}
		}

		// Check for high I/O (simplified check)
		if ioCounters, err := proc.IOCounters(); err == nil {
			totalIO := ioCounters.ReadBytes + ioCounters.WriteBytes
			if totalIO > 100*1024*1024 { // >100MB total I/O
				highIO++
			}
		}
	}

	metrics.ProcessStats.ThreadsTotal = totalThreads
	metrics.ProcessStats.FDsTotal = totalFDs
	metrics.ProcessStats.HighCPUProcesses = highCPU
	metrics.ProcessStats.HighMemProcesses = highMem
	metrics.ProcessStats.HighIOProcesses = highIO

	// Get system limits
	o.collectSystemLimits(&metrics.ProcessStats)

	// Check for process-related issues
	metrics.ProcessStats.Issues = o.checkProcessIssues(&metrics.ProcessStats)

	return nil
}

// collectTopProcesses collects top processes by different metrics
func (o *SysloadV2Options) collectTopProcesses(metrics *SystemMetricsV2) error {
	var err error

	metrics.TopCPU, err = o.getTopProcesses("cpu", o.TopN)
	if err != nil {
		return fmt.Errorf("failed to get top CPU processes: %w", err)
	}

	metrics.TopMemory, err = o.getTopProcesses("memory", o.TopN)
	if err != nil {
		return fmt.Errorf("failed to get top memory processes: %w", err)
	}

	metrics.TopIO, err = o.getTopProcesses("io", o.TopN)
	if err != nil {
		return fmt.Errorf("failed to get top I/O processes: %w", err)
	}

	return nil
}

// getTopProcesses returns enhanced top N processes sorted by specified metric
func (o *SysloadV2Options) getTopProcesses(sortBy string, n int) ([]ProcessInfoV2, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var procInfos []ProcessInfoV2

	for _, proc := range processes {
		info := ProcessInfoV2{
			PID: proc.Pid,
		}

		// Get basic process information
		if name, err := proc.Name(); err == nil {
			info.Name = name
		}

		if ppid, err := proc.Ppid(); err == nil {
			info.PPID = ppid
		}

		if username, err := proc.Username(); err == nil {
			info.Username = username
		}

		// Handle status array - convert []string to string
		if statusSlice, err := proc.Status(); err == nil {
			if len(statusSlice) > 0 {
				// Use the first status as primary status
				info.Status = statusSlice[0]

				// Or if you want to show all statuses, join them
				// info.Status = strings.Join(statusSlice, ",")

				// Check for zombie status specifically
				for _, s := range statusSlice {
					if strings.ToUpper(s) == "Z" || strings.ToLower(s) == "zombie" {
						info.Status = "Z" // Mark as zombie for display
						break
					}
				}
			}
		}

		if createTime, err := proc.CreateTime(); err == nil {
			info.CreateTime = time.Unix(createTime/1000, 0)
		}

		if cwd, err := proc.Cwd(); err == nil {
			info.Cwd = cwd
		}

		if exe, err := proc.Exe(); err == nil {
			info.Exe = exe
		}

		// Get command line
		if cmdline, err := proc.Cmdline(); err == nil {
			info.Cmdline = cmdline
			info.Command = cmdline
			// Truncate long command lines for display
			if len(info.Command) > 80 {
				info.Command = info.Command[:77] + "..."
			}
		}

		// Get CPU information
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			info.CPUPercent = cpuPercent
		}

		if times, err := proc.Times(); err == nil {
			info.CPUTime = times.User + times.System
		}

		// Get memory information
		if memInfo, err := proc.MemoryInfo(); err == nil {
			info.MemoryMB = float64(memInfo.RSS) / 1024 / 1024
			info.VirtualMemoryMB = float64(memInfo.VMS) / 1024 / 1024
		}

		if memPercent, err := proc.MemoryPercent(); err == nil {
			info.MemoryPercent = float64(memPercent)
		}

		// Get I/O information
		if ioCounters, err := proc.IOCounters(); err == nil {
			info.IOReadMB = float64(ioCounters.ReadBytes) / 1024 / 1024
			info.IOWriteMB = float64(ioCounters.WriteBytes) / 1024 / 1024
			info.IOReadCount = ioCounters.ReadCount
			info.IOWriteCount = ioCounters.WriteCount
		}

		// Get additional process details
		if connections, err := proc.Connections(); err == nil {
			info.NetConnections = int32(len(connections))
		}

		if openFiles, err := proc.OpenFiles(); err == nil {
			info.OpenFiles = int32(len(openFiles))
		}

		if numThreads, err := proc.NumThreads(); err == nil {
			info.NumThreads = numThreads
		}

		if ctxSwitches, err := proc.NumCtxSwitches(); err == nil {
			info.NumCtxSwitches = uint64(ctxSwitches.Voluntary) + uint64(ctxSwitches.Involuntary)
		}

		if nice, err := proc.Nice(); err == nil {
			info.Nice = nice
		}

		if isRunning, err := proc.IsRunning(); err == nil {
			info.IsRunning = isRunning
		}

		// Check for process-specific issues
		info.Issues = o.checkProcessSpecificIssues(&info)

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

// collectPartitionInfo collects detailed information about a disk partition
func (o *SysloadV2Options) collectPartitionInfo(partition disk.PartitionStat) PartitionInfo {
	partInfo := PartitionInfo{
		Device:     partition.Device,
		Mountpoint: partition.Mountpoint,
		Filesystem: partition.Fstype,
		Options:    strings.Join(partition.Opts, ","),
	}

	// Get usage statistics
	if usage, err := disk.Usage(partition.Mountpoint); err == nil {
		partInfo.UsagePercent = usage.UsedPercent
		partInfo.TotalGB = float64(usage.Total) / 1024 / 1024 / 1024
		partInfo.UsedGB = float64(usage.Used) / 1024 / 1024 / 1024
		partInfo.FreeGB = float64(usage.Free) / 1024 / 1024 / 1024
		partInfo.InodesTotal = usage.InodesTotal
		partInfo.InodesUsed = usage.InodesUsed
		partInfo.InodesFree = usage.InodesFree
		partInfo.InodesPercent = usage.InodesUsedPercent
	}

	// Check if partition is read-only
	partInfo.ReadOnly = strings.Contains(partInfo.Options, "ro")

	// Check for partition issues
	partInfo.Issues = o.checkPartitionIssues(&partInfo)

	return partInfo
}

// collectExtendedMemoryStats collects additional memory statistics from /proc/meminfo
func (o *SysloadV2Options) collectExtendedMemoryStats(memMetrics *MemoryMetricsV2) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		// Convert kB to bytes
		valueBytes := value * 1024

		switch key {
		case "Dirty":
			memMetrics.Dirty = valueBytes
		case "Writeback":
			memMetrics.Writeback = valueBytes
		case "Slab":
			memMetrics.Slab = valueBytes
		}
	}

	// Get page fault statistics
	o.collectPageFaultStats(memMetrics)
}

// collectPageFaultStats collects page fault statistics from /proc/vmstat
func (o *SysloadV2Options) collectPageFaultStats(memMetrics *MemoryMetricsV2) {
	file, err := os.Open("/proc/vmstat")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := fields[0]
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "pgfault":
			memMetrics.PageFaults = value
		case "pgmajfault":
			memMetrics.MajorPageFaults = value
		case "pswpin":
			memMetrics.SwapIns = value
		case "pswpout":
			memMetrics.SwapOuts = value
		case "pgpgin":
			memMetrics.PageIns = value
		case "pgpgout":
			memMetrics.PageOuts = value
		}
	}
}

// collectNetworkConnections collects active network connections
func (o *SysloadV2Options) collectNetworkConnections(netMetrics *NetworkMetrics) {
	connections, err := net.Connections("all")
	if err != nil {
		return
	}

	var activeConns, timeWaitConns int

	for _, conn := range connections {
		connInfo := ConnectionInfo{
			Type:       o.connectionTypeToString(conn.Type),
			LocalAddr:  conn.Laddr.IP,
			LocalPort:  conn.Laddr.Port,
			RemoteAddr: conn.Raddr.IP,
			RemotePort: conn.Raddr.Port,
			Status:     conn.Status,
			PID:        conn.Pid,
			Family:     conn.Family,
		}

		// Get process name for the connection
		if conn.Pid > 0 {
			if proc, err := process.NewProcess(conn.Pid); err == nil {
				if name, err := proc.Name(); err == nil {
					connInfo.Process = name
				}
			}
		}

		netMetrics.Connections = append(netMetrics.Connections, connInfo)

		// Count connection states
		switch conn.Status {
		case "ESTABLISHED", "CONNECTED":
			activeConns++
		case "TIME_WAIT":
			timeWaitConns++
		}
	}

	netMetrics.ActiveConnections = activeConns
	netMetrics.TimeWaitConnections = timeWaitConns
}

// collectListeningPorts collects information about listening ports
func (o *SysloadV2Options) collectListeningPorts(netMetrics *NetworkMetrics) {
	connections, err := net.Connections("all")
	if err != nil {
		return
	}

	portMap := make(map[string]PortInfo)

	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			key := fmt.Sprintf("%s:%d", conn.Type, conn.Laddr.Port)
			if _, exists := portMap[key]; !exists {
				portInfo := PortInfo{
					Port:     conn.Laddr.Port,
					Protocol: o.connectionTypeToString(conn.Type),
					Address:  conn.Laddr.IP,
					PID:      conn.Pid,
				}

				// Get process name
				if conn.Pid > 0 {
					if proc, err := process.NewProcess(conn.Pid); err == nil {
						if name, err := proc.Name(); err == nil {
							portInfo.Process = name
						}
					}
				}

				portMap[key] = portInfo
			}
		}
	}

	// Convert map to slice
	for _, portInfo := range portMap {
		netMetrics.ListenPorts = append(netMetrics.ListenPorts, portInfo)
	}

	// Sort by port number
	sort.Slice(netMetrics.ListenPorts, func(i, j int) bool {
		return netMetrics.ListenPorts[i].Port < netMetrics.ListenPorts[j].Port
	})
}

// collectSystemLimits collects system resource limits
func (o *SysloadV2Options) collectSystemLimits(procStats *ProcessStats) {
	// Get file descriptor limits
	if data, err := ioutil.ReadFile("/proc/sys/fs/file-nr"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			if used, err := strconv.Atoi(fields[0]); err == nil {
				procStats.FDsTotal = used
			}
			if max, err := strconv.Atoi(fields[2]); err == nil {
				procStats.FDsLimit = max
			}
		}
	}

	// Calculate usage percentage
	if procStats.FDsLimit > 0 {
		procStats.FDsUsagePercent = float64(procStats.FDsTotal) / float64(procStats.FDsLimit) * 100
	}
}

// getCPUTemperatures gets CPU temperature readings (Linux-specific)
func (o *SysloadV2Options) getCPUTemperatures() []float64 {
	var temperatures []float64

	// Try to read from thermal zones
	thermalPath := "/sys/class/thermal"
	if entries, err := ioutil.ReadDir(thermalPath); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "thermal_zone") {
				tempFile := filepath.Join(thermalPath, entry.Name(), "temp")
				if data, err := ioutil.ReadFile(tempFile); err == nil {
					if temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
						// Convert from millicelsius to celsius
						temperatures = append(temperatures, temp/1000)
					}
				}
			}
		}
	}

	return temperatures
}

// parseNetworkFlags parses network interface flags
func (o *SysloadV2Options) parseNetworkFlags(flags []string) []string {
	var parsedFlags []string
	for _, flag := range flags {
		parsedFlags = append(parsedFlags, flag)
	}
	return parsedFlags
}

// isInterfaceUp checks if a network interface is up
func (o *SysloadV2Options) isInterfaceUp(flags []string) bool {
	for _, flag := range flags {
		if flag == "up" {
			return true
		}
	}
	return false
}

// checkCPUAbnormalities checks for CPU-related issues
func (o *SysloadV2Options) checkCPUAbnormalities(cpuMetrics *CPUMetricsV2) bool {
	var issues []string
	abnormal := false

	// High CPU usage
	if cpuMetrics.Usage > CPUThresholdV2 {
		issues = append(issues, fmt.Sprintf("High CPU usage: %.1f%% (threshold: %.1f%%)",
			cpuMetrics.Usage, CPUThresholdV2))
		abnormal = true
	}

	// High I/O wait
	if cpuMetrics.IOWaitPercent > IOWaitThresholdV2 {
		issues = append(issues, fmt.Sprintf("High I/O wait: %.1f%% (threshold: %.1f%%)",
			cpuMetrics.IOWaitPercent, IOWaitThresholdV2))
		abnormal = true
	}

	// High load average
	loadThreshold := float64(cpuMetrics.Cores) * LoadThresholdV2
	if cpuMetrics.LoadAvg1 > loadThreshold {
		issues = append(issues, fmt.Sprintf("High load average: %.2f (threshold: %.2f)",
			cpuMetrics.LoadAvg1, loadThreshold))
		abnormal = true
	}

	// High context switches
	if cpuMetrics.ContextSwitches > ContextSwitchRate {
		issues = append(issues, fmt.Sprintf("High context switches: %d/sec (threshold: %d/sec)",
			cpuMetrics.ContextSwitches, ContextSwitchRate))
		abnormal = true
	}

	// High interrupts
	if cpuMetrics.Interrupts > InterruptRate {
		issues = append(issues, fmt.Sprintf("High interrupt rate: %d/sec (threshold: %d/sec)",
			cpuMetrics.Interrupts, InterruptRate))
		abnormal = true
	}

	// High CPU temperature
	for i, temp := range cpuMetrics.Temperature {
		if temp > 80.0 {
			issues = append(issues, fmt.Sprintf("High CPU temperature on core %d: %.1f°C", i, temp))
			abnormal = true
		}
	}

	// High steal time (virtualized environments)
	if cpuMetrics.StealPercent > 10.0 {
		issues = append(issues, fmt.Sprintf("High CPU steal time: %.1f%% (possible resource contention)",
			cpuMetrics.StealPercent))
		abnormal = true
	}

	cpuMetrics.Issues = issues
	return abnormal
}

// checkMemoryAbnormalities checks for memory-related issues
func (o *SysloadV2Options) checkMemoryAbnormalities(memMetrics *MemoryMetricsV2) bool {
	var issues []string
	abnormal := false

	// High memory usage
	if memMetrics.UsagePercent > MemoryThresholdV2 {
		issues = append(issues, fmt.Sprintf("High memory usage: %.1f%% (threshold: %.1f%%)",
			memMetrics.UsagePercent, MemoryThresholdV2))
		abnormal = true
	}

	// Swap usage
	if memMetrics.SwapPercent > SwapThreshold {
		issues = append(issues, fmt.Sprintf("Swap in use: %.1f%% (%.2fGB)",
			memMetrics.SwapPercent, memMetrics.SwapUsedGB))
		abnormal = true
	}

	// Low available memory
	availablePercent := (float64(memMetrics.Available) / float64(memMetrics.Total)) * 100
	if availablePercent < 10.0 {
		issues = append(issues, fmt.Sprintf("Low available memory: %.1f%% (%.2fGB)",
			availablePercent, memMetrics.AvailableGB))
		abnormal = true
	}

	// High dirty pages
	if memMetrics.Dirty > 0 {
		dirtyPercent := (float64(memMetrics.Dirty) / float64(memMetrics.Total)) * 100
		if dirtyPercent > 5.0 {
			issues = append(issues, fmt.Sprintf("High dirty pages: %.1f%% (%.2fGB)",
				dirtyPercent, float64(memMetrics.Dirty)/1024/1024/1024))
			abnormal = true
		}
	}

	// High page fault rate
	if memMetrics.MajorPageFaults > 1000 {
		issues = append(issues, "High major page fault rate (possible memory pressure)")
		abnormal = true
	}

	// Active swap I/O
	if memMetrics.SwapIns > 0 || memMetrics.SwapOuts > 0 {
		issues = append(issues, fmt.Sprintf("Active swapping detected (in: %d, out: %d)",
			memMetrics.SwapIns, memMetrics.SwapOuts))
		abnormal = true
	}

	memMetrics.Issues = issues
	return abnormal
}

// checkDiskAbnormalities checks for disk-related issues
func (o *SysloadV2Options) checkDiskAbnormalities(diskMetrics *DiskMetricsV2) bool {
	abnormal := false
	var issues []string

	// Check each partition
	for _, partition := range diskMetrics.Partitions {
		if partition.UsagePercent > DiskThresholdV2 {
			issues = append(issues, fmt.Sprintf("High disk usage on %s: %.1f%%",
				partition.Mountpoint, partition.UsagePercent))
			abnormal = true
		}

		if partition.InodesPercent > InodeThreshold {
			issues = append(issues, fmt.Sprintf("High inode usage on %s: %.1f%%",
				partition.Mountpoint, partition.InodesPercent))
			abnormal = true
		}
	}

	// Check I/O statistics
	for _, ioStat := range diskMetrics.IOStats {
		if ioStat.Utilization > 90.0 {
			issues = append(issues, fmt.Sprintf("High I/O utilization on %s: %.1f%%",
				ioStat.Device, ioStat.Utilization))
			abnormal = true
		}

		if ioStat.ReadLatencyMS > 100.0 {
			issues = append(issues, fmt.Sprintf("High read latency on %s: %.1fms",
				ioStat.Device, ioStat.ReadLatencyMS))
			abnormal = true
		}

		if ioStat.WriteLatencyMS > 100.0 {
			issues = append(issues, fmt.Sprintf("High write latency on %s: %.1fms",
				ioStat.Device, ioStat.WriteLatencyMS))
			abnormal = true
		}
	}

	diskMetrics.Issues = issues
	return abnormal
}

// checkNetworkAbnormalities checks for network-related issues
func (o *SysloadV2Options) checkNetworkAbnormalities(netMetrics *NetworkMetrics) bool {
	abnormal := false
	var issues []string

	// Check interface error rates
	for _, iface := range netMetrics.Interfaces {
		if iface.ErrorRate > NetworkErrorRate {
			issues = append(issues, fmt.Sprintf("High error rate on %s: %.2f%%",
				iface.Name, iface.ErrorRate))
			abnormal = true
		}

		if iface.DropRate > NetworkErrorRate {
			issues = append(issues, fmt.Sprintf("High drop rate on %s: %.2f%%",
				iface.Name, iface.DropRate))
			abnormal = true
		}
	}

	// Check for too many TIME_WAIT connections
	if netMetrics.TimeWaitConnections > 1000 {
		issues = append(issues, fmt.Sprintf("High TIME_WAIT connections: %d",
			netMetrics.TimeWaitConnections))
		abnormal = true
	}

	// Check for too many active connections
	if netMetrics.ActiveConnections > 10000 {
		issues = append(issues, fmt.Sprintf("High active connections: %d",
			netMetrics.ActiveConnections))
		abnormal = true
	}

	netMetrics.Issues = issues
	return abnormal
}

// checkLoadAbnormalities checks for load-related issues
func (o *SysloadV2Options) checkLoadAbnormalities(loadMetrics *LoadMetricsV2, cores int) bool {
	var issues []string
	abnormal := false

	threshold := float64(cores) * LoadThresholdV2

	if loadMetrics.Load1 > threshold {
		issues = append(issues, fmt.Sprintf("High 1-min load: %.2f (threshold: %.2f)",
			loadMetrics.Load1, threshold))
		abnormal = true
	}

	if loadMetrics.Load5 > threshold {
		issues = append(issues, fmt.Sprintf("High 5-min load: %.2f (threshold: %.2f)",
			loadMetrics.Load5, threshold))
		abnormal = true
	}

	if loadMetrics.ZombieProcs > ZombieProcessLimit {
		issues = append(issues, fmt.Sprintf("Too many zombie processes: %d",
			loadMetrics.ZombieProcs))
		abnormal = true
	}

	if loadMetrics.BlockedProcs > 10 {
		issues = append(issues, fmt.Sprintf("Many blocked processes: %d",
			loadMetrics.BlockedProcs))
		abnormal = true
	}

	loadMetrics.Issues = issues
	return abnormal
}

// checkProcessIssues checks for process-related issues
func (o *SysloadV2Options) checkProcessIssues(procStats *ProcessStats) []string {
	var issues []string

	if procStats.Zombie > ZombieProcessLimit {
		issues = append(issues, fmt.Sprintf("Too many zombie processes: %d", procStats.Zombie))
	}

	if procStats.HighCPUProcesses > HighCPUProcessLimit {
		issues = append(issues, fmt.Sprintf("Many high-CPU processes: %d", procStats.HighCPUProcesses))
	}

	if procStats.HighMemProcesses > HighMemProcessLimit {
		issues = append(issues, fmt.Sprintf("Many high-memory processes: %d", procStats.HighMemProcesses))
	}

	if procStats.FDsUsagePercent > FDThreshold {
		issues = append(issues, fmt.Sprintf("High file descriptor usage: %.1f%%", procStats.FDsUsagePercent))
	}

	return issues
}

// checkDiskIssues checks for disk-specific issues
func (o *SysloadV2Options) checkDiskIssues(diskStat *DiskIOStats) []string {
	var issues []string

	if diskStat.Utilization > 90.0 {
		issues = append(issues, fmt.Sprintf("High utilization: %.1f%%", diskStat.Utilization))
	}

	if diskStat.ReadLatencyMS > 100.0 {
		issues = append(issues, fmt.Sprintf("High read latency: %.1fms", diskStat.ReadLatencyMS))
	}

	if diskStat.WriteLatencyMS > 100.0 {
		issues = append(issues, fmt.Sprintf("High write latency: %.1fms", diskStat.WriteLatencyMS))
	}

	if diskStat.QueueDepth > 32.0 {
		issues = append(issues, fmt.Sprintf("High queue depth: %.1f", diskStat.QueueDepth))
	}

	return issues
}

// checkPartitionIssues checks for partition-specific issues
func (o *SysloadV2Options) checkPartitionIssues(partInfo *PartitionInfo) []string {
	var issues []string

	if partInfo.UsagePercent > DiskThresholdV2 {
		issues = append(issues, fmt.Sprintf("High space usage: %.1f%%", partInfo.UsagePercent))
	}

	if partInfo.InodesPercent > InodeThreshold {
		issues = append(issues, fmt.Sprintf("High inode usage: %.1f%%", partInfo.InodesPercent))
	}

	if partInfo.ReadOnly {
		issues = append(issues, "Mounted read-only")
	}

	if partInfo.FreeGB < 1.0 && partInfo.TotalGB > 10.0 {
		issues = append(issues, fmt.Sprintf("Low free space: %.2fGB", partInfo.FreeGB))
	}

	return issues
}

// checkNetworkInterfaceIssues checks for network interface issues
func (o *SysloadV2Options) checkNetworkInterfaceIssues(netIface *NetworkInterface) []string {
	var issues []string

	if !netIface.IsUp && netIface.Name != "lo" {
		issues = append(issues, "Interface is down")
	}

	if netIface.ErrorRate > NetworkErrorRate {
		issues = append(issues, fmt.Sprintf("High error rate: %.2f%%", netIface.ErrorRate))
	}

	if netIface.DropRate > NetworkErrorRate {
		issues = append(issues, fmt.Sprintf("High drop rate: %.2f%%", netIface.DropRate))
	}

	if netIface.ErrIn > 1000 || netIface.ErrOut > 1000 {
		issues = append(issues, fmt.Sprintf("Errors detected (in: %d, out: %d)", netIface.ErrIn, netIface.ErrOut))
	}

	return issues
}

// checkProcessSpecificIssues checks for process-specific issues
func (o *SysloadV2Options) checkProcessSpecificIssues(procInfo *ProcessInfoV2) []string {
	var issues []string

	if procInfo.CPUPercent > CPUThresholdV2 {
		issues = append(issues, fmt.Sprintf("High CPU usage: %.1f%%", procInfo.CPUPercent))
	}

	if procInfo.MemoryPercent > 20.0 {
		issues = append(issues, fmt.Sprintf("High memory usage: %.1f%%", procInfo.MemoryPercent))
	}

	if procInfo.Status == "Z" {
		issues = append(issues, "Zombie process")
	}

	if procInfo.Status == "D" {
		issues = append(issues, "Uninterruptible sleep (I/O wait)")
	}

	if procInfo.NumThreads > 1000 {
		issues = append(issues, fmt.Sprintf("Many threads: %d", procInfo.NumThreads))
	}

	if procInfo.OpenFiles > 1000 {
		issues = append(issues, fmt.Sprintf("Many open files: %d", procInfo.OpenFiles))
	}

	if !procInfo.IsRunning {
		issues = append(issues, "Process not running")
	}

	return issues
}

// analyzeSystemHealth performs comprehensive system health analysis
func (o *SysloadV2Options) analyzeSystemHealth(metrics *SystemMetricsV2) {
	health := SystemHealth{
		Score: 100,
	}

	// Analyze CPU health
	cpuScore := o.analyzeCPUHealth(&metrics.CPU, &health)

	// Analyze memory health
	memoryScore := o.analyzeMemoryHealth(&metrics.Memory, &health)

	// Analyze disk health
	diskScore := o.analyzeDiskHealth(&metrics.Disk, &health)

	// Analyze network health (if not skipped)
	networkScore := 100
	if !o.SkipNetwork {
		networkScore = o.analyzeNetworkHealth(&metrics.Network, &health)
	}

	// Analyze load health
	loadScore := o.analyzeLoadHealth(&metrics.LoadAverage, &health)

	// Analyze process health
	processScore := o.analyzeProcessHealth(&metrics.ProcessStats, &health)

	// Calculate overall score (weighted average)
	weights := map[string]float64{
		"cpu":     0.25,
		"memory":  0.25,
		"disk":    0.20,
		"network": 0.10,
		"load":    0.10,
		"process": 0.10,
	}

	health.Score = int(float64(cpuScore)*weights["cpu"] +
		float64(memoryScore)*weights["memory"] +
		float64(diskScore)*weights["disk"] +
		float64(networkScore)*weights["network"] +
		float64(loadScore)*weights["load"] +
		float64(processScore)*weights["process"])

	// Determine overall status
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 60:
		health.Status = "warning"
	default:
		health.Status = "critical"
	}

	// Collect system limits
	o.collectSystemLimitsDetailed(&health.SystemLimits)

	// Analyze resource pressure
	o.analyzeResourcePressure(&health.ResourcePressure, metrics)

	// Analyze stability indicators
	o.analyzeStabilityIndicators(&health.StabilityIndicators, metrics)

	metrics.SystemHealth = health
}

// analyzeCPUHealth analyzes CPU health and returns a score
func (o *SysloadV2Options) analyzeCPUHealth(cpu *CPUMetricsV2, health *SystemHealth) int {
	score := 100

	if cpu.Usage > 90 {
		health.CriticalAlerts = append(health.CriticalAlerts, "Critical CPU usage")
		score -= 30
	} else if cpu.Usage > CPUThresholdV2 {
		health.Warnings = append(health.Warnings, "High CPU usage")
		score -= 15
	}

	if cpu.IOWaitPercent > 30 {
		health.CriticalAlerts = append(health.CriticalAlerts, "Critical I/O wait time")
		score -= 25
	} else if cpu.IOWaitPercent > IOWaitThresholdV2 {
		health.Warnings = append(health.Warnings, "High I/O wait time")
		score -= 10
	}

	loadThreshold := float64(cpu.Cores) * LoadThresholdV2
	if cpu.LoadAvg1 > loadThreshold*1.5 {
		health.CriticalAlerts = append(health.CriticalAlerts, "Critical load average")
		score -= 20
	} else if cpu.LoadAvg1 > loadThreshold {
		health.Warnings = append(health.Warnings, "High load average")
		score -= 10
	}

	if cpu.ContextSwitches > ContextSwitchRate*2 {
		health.Issues = append(health.Issues, "Very high context switch rate")
		score -= 10
	} else if cpu.ContextSwitches > ContextSwitchRate {
		health.Issues = append(health.Issues, "High context switch rate")
		score -= 5
	}

	return max(score, 0)
}

// analyzeMemoryHealth analyzes memory health and returns a score
func (o *SysloadV2Options) analyzeMemoryHealth(memory *MemoryMetricsV2, health *SystemHealth) int {
	score := 100

	if memory.UsagePercent > 95 {
		health.CriticalAlerts = append(health.CriticalAlerts, "Critical memory usage")
		score -= 30
	} else if memory.UsagePercent > MemoryThresholdV2 {
		health.Warnings = append(health.Warnings, "High memory usage")
		score -= 15
	}

	if memory.SwapPercent > 50 {
		health.CriticalAlerts = append(health.CriticalAlerts, "Heavy swap usage")
		score -= 25
	} else if memory.SwapPercent > SwapThreshold {
		health.Warnings = append(health.Warnings, "Swap in use")
		score -= 10
	}

	availablePercent := (float64(memory.Available) / float64(memory.Total)) * 100
	if availablePercent < 5 {
		health.CriticalAlerts = append(health.CriticalAlerts, "Very low available memory")
		score -= 20
	} else if availablePercent < 10 {
		health.Warnings = append(health.Warnings, "Low available memory")
		score -= 10
	}

	if memory.SwapIns > 0 || memory.SwapOuts > 0 {
		health.Issues = append(health.Issues, "Active swapping detected")
		score -= 5
	}

	return max(score, 0)
}

// analyzeDiskHealth analyzes disk health and returns a score
func (o *SysloadV2Options) analyzeDiskHealth(disk *DiskMetricsV2, health *SystemHealth) int {
	score := 100

	for _, partition := range disk.Partitions {
		if partition.UsagePercent > 95 {
			health.CriticalAlerts = append(health.CriticalAlerts,
				fmt.Sprintf("Critical disk space on %s", partition.Mountpoint))
			score -= 20
		} else if partition.UsagePercent > DiskThresholdV2 {
			health.Warnings = append(health.Warnings,
				fmt.Sprintf("High disk usage on %s", partition.Mountpoint))
			score -= 10
		}

		if partition.InodesPercent > 95 {
			health.CriticalAlerts = append(health.CriticalAlerts,
				fmt.Sprintf("Critical inode usage on %s", partition.Mountpoint))
			score -= 15
		} else if partition.InodesPercent > InodeThreshold {
			health.Warnings = append(health.Warnings,
				fmt.Sprintf("High inode usage on %s", partition.Mountpoint))
			score -= 5
		}
	}

	for _, ioStat := range disk.IOStats {
		if ioStat.Utilization > 95 {
			health.Issues = append(health.Issues,
				fmt.Sprintf("Very high I/O utilization on %s", ioStat.Device))
			score -= 10
		}

		if ioStat.ReadLatencyMS > 200 || ioStat.WriteLatencyMS > 200 {
			health.Issues = append(health.Issues,
				fmt.Sprintf("High I/O latency on %s", ioStat.Device))
			score -= 10
		}
	}

	return max(score, 0)
}

// analyzeNetworkHealth analyzes network health and returns a score
func (o *SysloadV2Options) analyzeNetworkHealth(network *NetworkMetrics, health *SystemHealth) int {
	score := 100

	for _, iface := range network.Interfaces {
		if iface.ErrorRate > 5.0 {
			health.Warnings = append(health.Warnings,
				fmt.Sprintf("High error rate on %s", iface.Name))
			score -= 10
		}

		if iface.DropRate > 5.0 {
			health.Warnings = append(health.Warnings,
				fmt.Sprintf("High drop rate on %s", iface.Name))
			score -= 10
		}
	}

	if network.TimeWaitConnections > 5000 {
		health.Issues = append(health.Issues, "Many TIME_WAIT connections")
		score -= 5
	}

	if network.ActiveConnections > 50000 {
		health.Issues = append(health.Issues, "Very high active connections")
		score -= 10
	}

	return max(score, 0)
}

// analyzeLoadHealth analyzes load health and returns a score
func (o *SysloadV2Options) analyzeLoadHealth(load *LoadMetricsV2, health *SystemHealth) int {
	score := 100

	if load.ZombieProcs > ZombieProcessLimit*2 {
		health.Warnings = append(health.Warnings, "Many zombie processes")
		score -= 15
	} else if load.ZombieProcs > ZombieProcessLimit {
		health.Issues = append(health.Issues, "Some zombie processes detected")
		score -= 5
	}

	if load.BlockedProcs > 50 {
		health.Warnings = append(health.Warnings, "Many blocked processes")
		score -= 10
	} else if load.BlockedProcs > 10 {
		health.Issues = append(health.Issues, "Some processes blocked")
		score -= 5
	}

	return max(score, 0)
}

// analyzeProcessHealth analyzes process health and returns a score
func (o *SysloadV2Options) analyzeProcessHealth(process *ProcessStats, health *SystemHealth) int {
	score := 100

	if process.FDsUsagePercent > 90 {
		health.CriticalAlerts = append(health.CriticalAlerts, "Critical file descriptor usage")
		score -= 20
	} else if process.FDsUsagePercent > FDThreshold {
		health.Warnings = append(health.Warnings, "High file descriptor usage")
		score -= 10
	}

	if process.HighCPUProcesses > HighCPUProcessLimit*2 {
		health.Issues = append(health.Issues, "Many high-CPU processes")
		score -= 10
	}

	if process.HighMemProcesses > HighMemProcessLimit*2 {
		health.Issues = append(health.Issues, "Many high-memory processes")
		score -= 10
	}

	return max(score, 0)
}

// analyzePerformance performs performance analysis
func (o *SysloadV2Options) analyzePerformance(metrics *SystemMetricsV2) {
	perf := PerformanceStats{
		PerformanceScore: 100,
	}

	// Analyze response times
	o.analyzeResponseTimes(&perf.ResponseTime, metrics)

	// Analyze throughput
	o.analyzeThroughput(&perf.Throughput, metrics)

	// Identify bottlenecks
	perf.Bottlenecks = o.identifyBottlenecks(metrics)

	// Calculate performance score based on bottlenecks
	for _, bottleneck := range perf.Bottlenecks {
		switch bottleneck.Severity {
		case "critical":
			perf.PerformanceScore -= int(bottleneck.Impact * 0.8)
		case "high":
			perf.PerformanceScore -= int(bottleneck.Impact * 0.6)
		case "medium":
			perf.PerformanceScore -= int(bottleneck.Impact * 0.4)
		case "low":
			perf.PerformanceScore -= int(bottleneck.Impact * 0.2)
		}
	}

	if perf.PerformanceScore < 0 {
		perf.PerformanceScore = 0
	}

	// Generate optimization tips
	perf.OptimizationTips = o.generateOptimizationTips(metrics)

	metrics.Performance = perf
}

// analyzeResponseTimes analyzes system response times
func (o *SysloadV2Options) analyzeResponseTimes(rt *ResponseTimeStats, metrics *SystemMetricsV2) {
	// Estimate CPU scheduling latency
	if metrics.LoadAverage.Load1 > float64(metrics.CPU.Cores) {
		rt.AvgCPUScheduling = (metrics.LoadAverage.Load1 / float64(metrics.CPU.Cores)) * 10
	} else {
		rt.AvgCPUScheduling = 1.0
	}

	// Estimate disk I/O latency from disk stats
	var totalLatency float64
	var deviceCount int
	for _, disk := range metrics.Disk.IOStats {
		if disk.ReadLatencyMS > 0 || disk.WriteLatencyMS > 0 {
			totalLatency += (disk.ReadLatencyMS + disk.WriteLatencyMS) / 2
			deviceCount++
		}
	}
	if deviceCount > 0 {
		rt.AvgDiskIO = totalLatency / float64(deviceCount)
	}

	// Simple estimates for other metrics
	rt.AvgMemoryAccess = 0.1 // Typical DRAM access time
	rt.AvgNetworkIO = 1.0    // Typical network latency

	// Percentile estimates based on load
	loadFactor := metrics.LoadAverage.Load1 / float64(metrics.CPU.Cores)
	rt.P50ResponseTime = rt.AvgCPUScheduling + rt.AvgDiskIO
	rt.P95ResponseTime = rt.P50ResponseTime * (1 + loadFactor)
	rt.P99ResponseTime = rt.P95ResponseTime * (1 + loadFactor*0.5)
}

// analyzeThroughput analyzes system throughput
func (o *SysloadV2Options) analyzeThroughput(tp *ThroughputStats, metrics *SystemMetricsV2) {
	// Estimate CPU instructions per second
	cpuUsage := metrics.CPU.Usage / 100
	estimatedFreq := metrics.CPU.CPUFreq * 1000000 // Convert MHz to Hz
	tp.CPUInstructionsPerSec = uint64(estimatedFreq * cpuUsage * float64(metrics.CPU.Cores))

	// Sum disk IOPS
	for _, disk := range metrics.Disk.IOStats {
		tp.DiskIOPS += disk.ReadCount + disk.WriteCount
	}

	// Sum network packets per second
	for _, iface := range metrics.Network.Interfaces {
		tp.NetworkPPS += iface.PacketsRecv + iface.PacketsSent
	}

	// Estimate memory bandwidth (simplified)
	memUsage := metrics.Memory.UsagePercent / 100
	tp.MemoryBandwidthMBPS = memUsage * 1000 // Rough estimate

	// Estimate system calls per second based on context switches
	tp.SystemCallsPerSec = metrics.CPU.ContextSwitches / 10 // Rough ratio
}

// identifyBottlenecks identifies system bottlenecks
func (o *SysloadV2Options) identifyBottlenecks(metrics *SystemMetricsV2) []Bottleneck {
	var bottlenecks []Bottleneck

	// CPU bottlenecks
	if metrics.CPU.Usage > 90 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Component:   "cpu",
			Severity:    "critical",
			Impact:      metrics.CPU.Usage,
			Description: "CPU utilization is critically high",
			Cause:       "High computational load or inefficient processes",
			Solution:    "Optimize CPU-intensive processes, consider scaling up/out",
		})
	} else if metrics.CPU.Usage > CPUThresholdV2 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Component:   "cpu",
			Severity:    "high",
			Impact:      metrics.CPU.Usage,
			Description: "CPU utilization is high",
			Cause:       "Increased computational workload",
			Solution:    "Monitor processes, optimize algorithms",
		})
	}

	// I/O wait bottlenecks
	if metrics.CPU.IOWaitPercent > 30 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Component:   "disk",
			Severity:    "critical",
			Impact:      metrics.CPU.IOWaitPercent,
			Description: "High I/O wait time indicating disk bottleneck",
			Cause:       "Slow storage, high I/O load, or disk failure",
			Solution:    "Check disk health, optimize I/O, consider SSD upgrade",
		})
	}

	// Memory bottlenecks
	if metrics.Memory.UsagePercent > 95 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Component:   "memory",
			Severity:    "critical",
			Impact:      metrics.Memory.UsagePercent,
			Description: "Memory usage is critically high",
			Cause:       "Memory leak, large datasets, or insufficient RAM",
			Solution:    "Identify memory-hungry processes, add RAM, optimize memory usage",
		})
	}

	// Swap bottlenecks
	if metrics.Memory.SwapPercent > 50 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Component:   "memory",
			Severity:    "high",
			Impact:      metrics.Memory.SwapPercent,
			Description: "Heavy swap usage degrading performance",
			Cause:       "Insufficient physical memory",
			Solution:    "Add more RAM, optimize memory usage, tune swappiness",
		})
	}

	// Network bottlenecks
	for _, iface := range metrics.Network.Interfaces {
		if iface.ErrorRate > 5.0 {
			bottlenecks = append(bottlenecks, Bottleneck{
				Component:   "network",
				Severity:    "medium",
				Impact:      iface.ErrorRate,
				Description: fmt.Sprintf("High error rate on interface %s", iface.Name),
				Cause:       "Network congestion, faulty hardware, or driver issues",
				Solution:    "Check network hardware, update drivers, investigate traffic",
			})
		}
	}

	return bottlenecks
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// generateRecommendations generates actionable recommendations
func (o *SysloadV2Options) generateRecommendations(metrics *SystemMetricsV2) {
	var recommendations []string

	// CPU recommendations
	if metrics.CPU.Usage > CPUThresholdV2 {
		if metrics.CPU.IOWaitPercent > IOWaitThresholdV2 {
			recommendations = append(recommendations, "High I/O wait detected - consider upgrading storage or optimizing disk access patterns")
		} else {
			recommendations = append(recommendations, "High CPU usage - identify and optimize CPU-intensive processes")
		}

		if metrics.LoadAverage.Load1 > float64(metrics.CPU.Cores)*2 {
			recommendations = append(recommendations, "Load average is high - consider adding more CPU cores or distributing workload")
		}
	}

	if metrics.CPU.ContextSwitches > ContextSwitchRate {
		recommendations = append(recommendations, "High context switch rate - review multithreading and process scheduling")
	}

	// Memory recommendations
	if metrics.Memory.UsagePercent > MemoryThresholdV2 {
		recommendations = append(recommendations, "High memory usage - identify memory-intensive processes and consider adding RAM")

		if metrics.Memory.SwapPercent > SwapThreshold {
			recommendations = append(recommendations, "Active swapping detected - add more physical memory to avoid performance degradation")
		}
	}

	if metrics.Memory.SwapPercent > 0 && metrics.Memory.SwapPercent < SwapThreshold {
		recommendations = append(recommendations, "Some swap usage detected - monitor memory trends and consider memory optimization")
	}

	// Disk recommendations
	for _, partition := range metrics.Disk.Partitions {
		if partition.UsagePercent > DiskThresholdV2 {
			recommendations = append(recommendations,
				fmt.Sprintf("High disk usage on %s - clean up files or expand storage", partition.Mountpoint))
		}

		if partition.InodesPercent > InodeThreshold {
			recommendations = append(recommendations,
				fmt.Sprintf("High inode usage on %s - remove unnecessary small files", partition.Mountpoint))
		}
	}

	for _, disk := range metrics.Disk.IOStats {
		if disk.Utilization > 90 {
			recommendations = append(recommendations,
				fmt.Sprintf("High I/O utilization on %s - consider I/O optimization or faster storage", disk.Device))
		}

		if disk.ReadLatencyMS > 100 || disk.WriteLatencyMS > 100 {
			recommendations = append(recommendations,
				fmt.Sprintf("High I/O latency on %s - check disk health and consider SSD upgrade", disk.Device))
		}
	}

	// Network recommendations
	for _, iface := range metrics.Network.Interfaces {
		if iface.ErrorRate > NetworkErrorRate {
			recommendations = append(recommendations,
				fmt.Sprintf("High error rate on %s - check network hardware and cables", iface.Name))
		}

		if iface.DropRate > NetworkErrorRate {
			recommendations = append(recommendations,
				fmt.Sprintf("High drop rate on %s - investigate network congestion", iface.Name))
		}
	}

	if metrics.Network.TimeWaitConnections > 1000 {
		recommendations = append(recommendations, "High TIME_WAIT connections - tune TCP settings or optimize connection handling")
	}

	// Process recommendations
	if metrics.ProcessStats.Zombie > ZombieProcessLimit {
		recommendations = append(recommendations, "Zombie processes detected - investigate parent processes and fix cleanup issues")
	}

	if metrics.ProcessStats.FDsUsagePercent > FDThreshold {
		recommendations = append(recommendations, "High file descriptor usage - check for file descriptor leaks in applications")
	}

	// Load recommendations
	if metrics.LoadAverage.ZombieProcs > 0 {
		recommendations = append(recommendations, "Clean up zombie processes to improve system stability")
	}

	if metrics.LoadAverage.BlockedProcs > 10 {
		recommendations = append(recommendations, "Many blocked processes detected - investigate I/O bottlenecks")
	}

	// Temperature recommendations
	for i, temp := range metrics.CPU.Temperature {
		if temp > 80.0 {
			recommendations = append(recommendations,
				fmt.Sprintf("High CPU temperature on core %d (%.1f°C) - check cooling system", i, temp))
		}
	}

	// General performance recommendations
	if len(metrics.Performance.Bottlenecks) > 0 {
		recommendations = append(recommendations, "Performance bottlenecks identified - prioritize addressing critical and high-impact issues")
	}

	if metrics.SystemHealth.Score < 60 {
		recommendations = append(recommendations, "System health score is low - address critical alerts and warnings immediately")
	}

	metrics.Recommendations = recommendations
}

// generateAlerts generates system alerts based on thresholds
func (o *SysloadV2Options) generateAlerts(metrics *SystemMetricsV2) {
	var alerts []Alert

	// CPU alerts
	if metrics.CPU.Usage > 95 {
		alerts = append(alerts, Alert{
			Level:        "critical",
			Component:    "cpu",
			Metric:       "usage",
			CurrentValue: metrics.CPU.Usage,
			Threshold:    95.0,
			Message:      "Critical CPU usage detected",
			Impact:       "System performance severely degraded",
			Action:       "Immediately identify and terminate or optimize high-CPU processes",
			Timestamp:    metrics.Timestamp,
		})
	} else if metrics.CPU.Usage > CPUThresholdV2 {
		alerts = append(alerts, Alert{
			Level:        "warning",
			Component:    "cpu",
			Metric:       "usage",
			CurrentValue: metrics.CPU.Usage,
			Threshold:    CPUThresholdV2,
			Message:      "High CPU usage detected",
			Impact:       "System performance may be impacted",
			Action:       "Monitor CPU-intensive processes and optimize if needed",
			Timestamp:    metrics.Timestamp,
		})
	}

	if metrics.CPU.IOWaitPercent > 50 {
		alerts = append(alerts, Alert{
			Level:        "critical",
			Component:    "cpu",
			Metric:       "iowait",
			CurrentValue: metrics.CPU.IOWaitPercent,
			Threshold:    50.0,
			Message:      "Critical I/O wait time",
			Impact:       "System responsiveness severely affected",
			Action:       "Investigate disk I/O bottlenecks immediately",
			Timestamp:    metrics.Timestamp,
		})
	}

	// Memory alerts
	if metrics.Memory.UsagePercent > 95 {
		alerts = append(alerts, Alert{
			Level:        "critical",
			Component:    "memory",
			Metric:       "usage",
			CurrentValue: metrics.Memory.UsagePercent,
			Threshold:    95.0,
			Message:      "Critical memory usage",
			Impact:       "Risk of out-of-memory conditions",
			Action:       "Free memory immediately or restart memory-intensive processes",
			Timestamp:    metrics.Timestamp,
		})
	}

	if metrics.Memory.SwapPercent > 80 {
		alerts = append(alerts, Alert{
			Level:        "critical",
			Component:    "memory",
			Metric:       "swap",
			CurrentValue: metrics.Memory.SwapPercent,
			Threshold:    80.0,
			Message:      "Heavy swap usage",
			Impact:       "Severe performance degradation due to swapping",
			Action:       "Add physical memory or optimize memory usage",
			Timestamp:    metrics.Timestamp,
		})
	}

	// Disk alerts
	for _, partition := range metrics.Disk.Partitions {
		if partition.UsagePercent > 95 {
			alerts = append(alerts, Alert{
				Level:        "critical",
				Component:    "disk",
				Metric:       "space",
				CurrentValue: partition.UsagePercent,
				Threshold:    95.0,
				Message:      fmt.Sprintf("Critical disk space on %s", partition.Mountpoint),
				Impact:       "Risk of disk full, applications may fail",
				Action:       "Free disk space immediately or expand storage",
				Timestamp:    metrics.Timestamp,
			})
		}
	}

	// Load alerts
	loadThreshold := float64(metrics.CPU.Cores) * 3.0
	if metrics.LoadAverage.Load1 > loadThreshold {
		alerts = append(alerts, Alert{
			Level:        "critical",
			Component:    "load",
			Metric:       "load1",
			CurrentValue: metrics.LoadAverage.Load1,
			Threshold:    loadThreshold,
			Message:      "Critical system load",
			Impact:       "System may become unresponsive",
			Action:       "Reduce system load or add more CPU cores",
			Timestamp:    metrics.Timestamp,
		})
	}

	// Process alerts
	if metrics.ProcessStats.Zombie > ZombieProcessLimit*3 {
		alerts = append(alerts, Alert{
			Level:        "warning",
			Component:    "process",
			Metric:       "zombies",
			CurrentValue: float64(metrics.ProcessStats.Zombie),
			Threshold:    float64(ZombieProcessLimit * 3),
			Message:      "Many zombie processes detected",
			Impact:       "Resource waste and potential system instability",
			Action:       "Investigate and fix parent processes",
			Timestamp:    metrics.Timestamp,
		})
	}

	// Network alerts
	for _, iface := range metrics.Network.Interfaces {
		if iface.ErrorRate > 10.0 {
			alerts = append(alerts, Alert{
				Level:        "warning",
				Component:    "network",
				Metric:       "error_rate",
				CurrentValue: iface.ErrorRate,
				Threshold:    10.0,
				Message:      fmt.Sprintf("High error rate on %s", iface.Name),
				Impact:       "Network performance degradation",
				Action:       "Check network hardware and configuration",
				Timestamp:    metrics.Timestamp,
			})
		}
	}

	metrics.Alerts = alerts
}

// generateOptimizationTips generates performance optimization tips
func (o *SysloadV2Options) generateOptimizationTips(metrics *SystemMetricsV2) []string {
	var tips []string

	// CPU optimization tips
	if metrics.CPU.Usage > CPUThresholdV2 {
		tips = append(tips, "Use 'top' or 'htop' to identify CPU-intensive processes")
		tips = append(tips, "Consider process priority adjustment with 'nice' or 'renice'")
		tips = append(tips, "Evaluate CPU affinity settings for multi-core optimization")
	}

	if metrics.CPU.ContextSwitches > ContextSwitchRate {
		tips = append(tips, "Reduce number of running processes if possible")
		tips = append(tips, "Consider using fewer threads or async I/O patterns")
	}

	// Memory optimization tips
	if metrics.Memory.UsagePercent > MemoryThresholdV2 {
		tips = append(tips, "Use 'free -h' and 'vmstat' to monitor memory patterns")
		tips = append(tips, "Identify memory leaks with tools like 'valgrind' or 'mtrace'")
		tips = append(tips, "Tune application memory settings and garbage collection")
	}

	if metrics.Memory.SwapPercent > 0 {
		tips = append(tips, "Consider tuning swappiness: echo 10 > /proc/sys/vm/swappiness")
		tips = append(tips, "Monitor memory usage patterns to prevent future swapping")
	}

	// Disk optimization tips
	diskHighUtil := false
	for _, disk := range metrics.Disk.IOStats {
		if disk.Utilization > 80 {
			diskHighUtil = true
			break
		}
	}

	if diskHighUtil {
		tips = append(tips, "Use 'iotop' to identify I/O-intensive processes")
		tips = append(tips, "Consider I/O scheduling optimization (deadline, cfq, noop)")
		tips = append(tips, "Implement read/write caching strategies")
		tips = append(tips, "Consider RAID configuration for better I/O performance")
	}

	// Network optimization tips
	networkIssues := false
	for _, iface := range metrics.Network.Interfaces {
		if iface.ErrorRate > 1.0 || iface.DropRate > 1.0 {
			networkIssues = true
			break
		}
	}

	if networkIssues {
		tips = append(tips, "Check network interface statistics with 'ip -s link'")
		tips = append(tips, "Tune network buffer sizes and TCP window scaling")
		tips = append(tips, "Monitor network traffic patterns with 'iftop' or 'nethogs'")
	}

	if metrics.Network.TimeWaitConnections > 1000 {
		tips = append(tips, "Tune TCP TIME_WAIT settings: net.ipv4.tcp_tw_reuse=1")
		tips = append(tips, "Optimize application connection pooling")
	}

	// General system tips
	if metrics.SystemHealth.Score < 80 {
		tips = append(tips, "Schedule regular system maintenance and updates")
		tips = append(tips, "Implement monitoring and alerting for proactive issue detection")
		tips = append(tips, "Document baseline performance metrics for comparison")
	}

	if len(metrics.LoadAverage.Issues) > 0 {
		tips = append(tips, "Use 'ps aux --sort=-%cpu' to find top CPU consumers")
		tips = append(tips, "Consider load balancing across multiple systems")
	}

	return tips
}

// collectSystemLimitsDetailed collects detailed system limits
func (o *SysloadV2Options) collectSystemLimitsDetailed(limits *SystemLimits) {
	// File limits
	if data, err := ioutil.ReadFile("/proc/sys/fs/file-max"); err == nil {
		if max, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
			limits.MaxFiles = max
		}
	}

	if data, err := ioutil.ReadFile("/proc/sys/fs/file-nr"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 1 {
			if used, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
				limits.CurrentFiles = used
			}
		}
	}

	// Process limits
	if data, err := ioutil.ReadFile("/proc/sys/kernel/pid_max"); err == nil {
		if max, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
			limits.MaxProcesses = max
		}
	}

	// Thread limits
	if data, err := ioutil.ReadFile("/proc/sys/kernel/threads-max"); err == nil {
		if max, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
			limits.MaxThreads = max
		}
	}

	// Calculate usage percentages
	if limits.MaxFiles > 0 {
		limits.FileUsagePercent = float64(limits.CurrentFiles) / float64(limits.MaxFiles) * 100
	}

	if limits.MaxProcesses > 0 {
		limits.ProcessUsagePercent = float64(limits.CurrentProcesses) / float64(limits.MaxProcesses) * 100
	}

	if limits.MaxThreads > 0 {
		limits.ThreadUsagePercent = float64(limits.CurrentThreads) / float64(limits.MaxThreads) * 100
	}
}

// analyzeResourcePressure analyzes system resource pressure
func (o *SysloadV2Options) analyzeResourcePressure(pressure *ResourcePressure, metrics *SystemMetricsV2) {
	// Analyze CPU pressure
	pressure.CPUPressure = o.analyzePressureComponent("cpu", metrics)

	// Analyze memory pressure
	pressure.MemoryPressure = o.analyzePressureComponent("memory", metrics)

	// Analyze I/O pressure
	pressure.IOPressure = o.analyzePressureComponent("io", metrics)

	// Determine overall pressure
	pressureLevels := []string{
		pressure.CPUPressure.Level,
		pressure.MemoryPressure.Level,
		pressure.IOPressure.Level,
	}

	criticalCount := 0
	highCount := 0
	mediumCount := 0

	for _, level := range pressureLevels {
		switch level {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		case "medium":
			mediumCount++
		}
	}

	if criticalCount > 0 {
		pressure.OverallPressure = "critical"
	} else if highCount > 0 {
		pressure.OverallPressure = "high"
	} else if mediumCount > 0 {
		pressure.OverallPressure = "medium"
	} else {
		pressure.OverallPressure = "low"
	}
}

// analyzePressureComponent analyzes pressure for a specific component
func (o *SysloadV2Options) analyzePressureComponent(component string, metrics *SystemMetricsV2) PressureInfo {
	info := PressureInfo{}

	switch component {
	case "cpu":
		usage := metrics.CPU.Usage
		loadRatio := metrics.LoadAverage.Load1 / float64(metrics.CPU.Cores)

		info.Some10s = usage
		info.Some60s = usage * 0.9 // Simulate historical data
		info.Some300s = usage * 0.8

		if loadRatio > 2.0 {
			info.Level = "critical"
		} else if loadRatio > 1.5 {
			info.Level = "high"
		} else if loadRatio > 1.0 {
			info.Level = "medium"
		} else {
			info.Level = "low"
		}

	case "memory":
		usage := metrics.Memory.UsagePercent

		info.Some10s = usage
		info.Some60s = usage * 0.95
		info.Some300s = usage * 0.9

		if usage > 95 {
			info.Level = "critical"
		} else if usage > 85 {
			info.Level = "high"
		} else if usage > 70 {
			info.Level = "medium"
		} else {
			info.Level = "low"
		}

	case "io":
		avgUtil := 0.0
		deviceCount := 0
		for _, disk := range metrics.Disk.IOStats {
			avgUtil += disk.Utilization
			deviceCount++
		}
		if deviceCount > 0 {
			avgUtil /= float64(deviceCount)
		}

		info.Some10s = avgUtil
		info.Some60s = avgUtil * 0.9
		info.Some300s = avgUtil * 0.8

		if avgUtil > 90 {
			info.Level = "critical"
		} else if avgUtil > 70 {
			info.Level = "high"
		} else if avgUtil > 50 {
			info.Level = "medium"
		} else {
			info.Level = "low"
		}
	}

	return info
}

// analyzeStabilityIndicators analyzes system stability indicators
func (o *SysloadV2Options) analyzeStabilityIndicators(stability *StabilityMetrics, metrics *SystemMetricsV2) {
	stability.UptimeDays = metrics.Uptime.Hours() / 24

	// Check for zombie processes
	stability.OOMKills = uint64(metrics.ProcessStats.Zombie) // Simplified

	// Analyze memory usage patterns for potential leaks
	if metrics.Memory.UsagePercent > 90 && metrics.Memory.SwapPercent > 0 {
		stability.MemoryLeaks = 1
	}

	// Check for deadlocks (simplified detection)
	if metrics.LoadAverage.BlockedProcs > 20 {
		stability.DeadlockDetections = 1
	}

	// Calculate stability score
	score := 100

	if stability.OOMKills > 0 {
		score -= 20
	}
	if stability.MemoryLeaks > 0 {
		score -= 15
	}
	if stability.DeadlockDetections > 0 {
		score -= 10
	}
	if stability.UptimeDays < 1 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	stability.StabilityScore = score
}

// displayMetrics displays the collected metrics in the specified format
func (o *SysloadV2Options) displayMetrics(metrics *SystemMetricsV2) error {
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
func (o *SysloadV2Options) formatJSON(metrics *SystemMetricsV2) (string, error) {
	jsonBytes, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// formatTable formats metrics as a comprehensive table
func (o *SysloadV2Options) formatTable(metrics *SystemMetricsV2) (string, error) {
	var sb strings.Builder

	// Header with timestamp and hostname
	sb.WriteString(fmt.Sprintf("%s System Diagnostics Report - %s\n",
		emoji.ChartIncreasing, metrics.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Host: %s | Uptime: %s | Kernel: %s\n\n",
		metrics.Hostname,
		o.formatDuration(metrics.Uptime),
		metrics.KernelVersion))

	// System Health Overview
	sb.WriteString(o.formatSystemHealth(&metrics.SystemHealth))
	sb.WriteString("\n")

	// Alerts Section
	if len(metrics.Alerts) > 0 {
		sb.WriteString(o.formatAlerts(metrics.Alerts))
		sb.WriteString("\n")
	}

	// System Overview
	sb.WriteString(fmt.Sprintf("%s System Overview\n", emoji.DesktopComputer))
	sb.WriteString(strings.Repeat("=", 60) + "\n")

	// CPU metrics
	sb.WriteString(o.formatCPUMetricsV2(&metrics.CPU))
	sb.WriteString("\n")

	// Memory metrics
	sb.WriteString(o.formatMemoryMetricsV2(&metrics.Memory))
	sb.WriteString("\n")

	// Load metrics
	sb.WriteString(o.formatLoadMetricsV2(&metrics.LoadAverage))
	sb.WriteString("\n")

	// Disk metrics
	sb.WriteString(o.formatDiskMetricsV2(&metrics.Disk))
	sb.WriteString("\n")

	// Network metrics (if not skipped)
	if !o.SkipNetwork && len(metrics.Network.Interfaces) > 0 {
		sb.WriteString(o.formatNetworkMetrics(&metrics.Network))
		sb.WriteString("\n")
	}

	// Process statistics
	sb.WriteString(o.formatProcessStats(&metrics.ProcessStats))
	sb.WriteString("\n")

	// Top processes
	sb.WriteString(o.formatProcessTable("Top CPU Processes", metrics.TopCPU, "cpu"))
	sb.WriteString("\n")

	sb.WriteString(o.formatProcessTable("Top Memory Processes", metrics.TopMemory, "memory"))
	sb.WriteString("\n")

	sb.WriteString(o.formatProcessTable("Top I/O Processes", metrics.TopIO, "io"))
	sb.WriteString("\n")

	// Detailed diagnostics
	if o.Detailed {
		sb.WriteString(o.formatDetailedDiagnostics(metrics))
	}

	// Recommendations
	if len(metrics.Recommendations) > 0 {
		sb.WriteString(o.formatRecommendations(metrics.Recommendations))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// formatSystemHealth formats system health information
func (o *SysloadV2Options) formatSystemHealth(health *SystemHealth) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s System Health Status\n", emoji.Hospital))
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	// Status with color
	var statusColor *color.Color
	var statusEmoji string

	switch health.Status {
	case "healthy":
		statusColor = color.New(color.FgGreen, color.Bold)
		statusEmoji = emoji.CheckMark.String()
	case "warning":
		statusColor = color.New(color.FgYellow, color.Bold)
		statusEmoji = emoji.Warning.String()
	case "critical":
		statusColor = color.New(color.FgRed, color.Bold)
		statusEmoji = emoji.CrossMark.String()
	default:
		statusColor = color.New(color.FgWhite)
		statusEmoji = emoji.QuestionMark.String()
	}

	sb.WriteString(fmt.Sprintf("Overall Status: %s %s (Score: %d/100)\n",
		statusEmoji, statusColor.Sprintf(strings.ToUpper(health.Status)), health.Score))

	// Critical alerts
	if len(health.CriticalAlerts) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Critical Alerts:\n", emoji.MedicalSymbol))
		for _, alert := range health.CriticalAlerts {
			sb.WriteString(fmt.Sprintf("  • %s\n", color.RedString(alert)))
		}
	}

	// Warnings
	if len(health.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Warnings:\n", emoji.Warning))
		for _, warning := range health.Warnings {
			sb.WriteString(fmt.Sprintf("  • %s\n", color.YellowString(warning)))
		}
	}

	// Issues
	if len(health.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Issues:\n", emoji.ExclamationMark))
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
	}

	return sb.String()
}

// formatAlerts formats system alerts
func (o *SysloadV2Options) formatAlerts(alerts []Alert) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Active Alerts\n", emoji.MedicalSymbol))
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	// Group alerts by level
	critical := []Alert{}
	warnings := []Alert{}

	for _, alert := range alerts {
		switch alert.Level {
		case "critical":
			critical = append(critical, alert)
		case "warning":
			warnings = append(warnings, alert)
		}
	}

	// Display critical alerts
	if len(critical) > 0 {
		sb.WriteString(fmt.Sprintf("%s Critical Alerts:\n", emoji.SosButton))
		for _, alert := range critical {
			sb.WriteString(fmt.Sprintf("  %s %s: %s (%.1f > %.1f)\n",
				emoji.Fire,
				color.RedString(strings.ToUpper(alert.Component)),
				alert.Message,
				alert.CurrentValue,
				alert.Threshold))
			sb.WriteString(fmt.Sprintf("    Impact: %s\n", alert.Impact))
			sb.WriteString(fmt.Sprintf("    Action: %s\n", alert.Action))
		}
		sb.WriteString("\n")
	}

	// Display warnings
	if len(warnings) > 0 {
		sb.WriteString(fmt.Sprintf("%s Warnings:\n", emoji.Warning))
		for _, alert := range warnings {
			sb.WriteString(fmt.Sprintf("  %s %s: %s (%.1f > %.1f)\n",
				emoji.ExclamationMark,
				color.YellowString(strings.ToUpper(alert.Component)),
				alert.Message,
				alert.CurrentValue,
				alert.Threshold))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatCPUMetricsV2 formats CPU metrics
func (o *SysloadV2Options) formatCPUMetricsV2(cpu *CPUMetricsV2) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s CPU Information\n", emoji.Gear))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// Overall CPU usage with color
	cpuColor := color.New(color.FgGreen)
	if cpu.IsAbnormal {
		cpuColor = color.New(color.FgRed, color.Bold)
	}

	sb.WriteString(fmt.Sprintf("Overall Usage: %s\n",
		cpuColor.Sprintf("%.1f%%", cpu.Usage)))
	sb.WriteString(fmt.Sprintf("Cores:         %d physical, %d logical\n",
		cpu.Cores, cpu.LogicalCores))

	if cpu.CPUFreq > 0 {
		sb.WriteString(fmt.Sprintf("Frequency:     %.0f MHz\n", cpu.CPUFreq))
	}

	// CPU breakdown
	sb.WriteString("\nCPU Breakdown:\n")
	sb.WriteString(fmt.Sprintf("  User:        %.1f%%\n", cpu.UserPercent))
	sb.WriteString(fmt.Sprintf("  System:      %.1f%%\n", cpu.SystemPercent))
	sb.WriteString(fmt.Sprintf("  I/O Wait:    %.1f%%\n", cpu.IOWaitPercent))
	sb.WriteString(fmt.Sprintf("  Idle:        %.1f%%\n", cpu.IdlePercent))

	if cpu.StealPercent > 0 {
		sb.WriteString(fmt.Sprintf("  Steal:       %.1f%%\n", cpu.StealPercent))
	}

	// Load average
	loadColor := color.New(color.FgGreen)
	loadThreshold := float64(cpu.Cores) * LoadThresholdV2
	if cpu.LoadAvg1 > loadThreshold {
		loadColor = color.New(color.FgRed, color.Bold)
	}

	sb.WriteString(fmt.Sprintf("\nLoad Average:  %s\n",
		loadColor.Sprintf("%.2f, %.2f, %.2f",
			cpu.LoadAvg1, cpu.LoadAvg5, cpu.LoadAvg15)))

	// System activity
	if cpu.ContextSwitches > 0 {
		sb.WriteString(fmt.Sprintf("Context Switches: %s/sec\n",
			o.formatNumber(cpu.ContextSwitches)))
	}
	if cpu.Interrupts > 0 {
		sb.WriteString(fmt.Sprintf("Interrupts:    %s/sec\n",
			o.formatNumber(cpu.Interrupts)))
	}

	// Temperature
	if len(cpu.Temperature) > 0 {
		sb.WriteString("Temperature:   ")
		for i, temp := range cpu.Temperature {
			tempColor := color.New(color.FgGreen)
			if temp > 80.0 {
				tempColor = color.New(color.FgRed)
			} else if temp > 70.0 {
				tempColor = color.New(color.FgYellow)
			}
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(tempColor.Sprintf("%.1f°C", temp))
		}
		sb.WriteString("\n")
	}

	// Issues
	if len(cpu.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Issues:\n", emoji.Warning))
		for _, issue := range cpu.Issues {
			sb.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
	}

	return sb.String()
}

// formatMemoryMetricsV2 formats memory metrics
func (o *SysloadV2Options) formatMemoryMetricsV2(memory *MemoryMetricsV2) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Memory Information\n", emoji.Brain))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// Memory usage with color
	memColor := color.New(color.FgGreen)
	if memory.IsAbnormal {
		memColor = color.New(color.FgRed, color.Bold)
	}

	sb.WriteString(fmt.Sprintf("Memory Usage:  %s / %.1fGB (%s)\n",
		memColor.Sprintf("%.1fGB", memory.UsedGB),
		memory.TotalGB,
		memColor.Sprintf("%.1f%%", memory.UsagePercent)))

	sb.WriteString(fmt.Sprintf("Available:     %.1fGB\n", memory.AvailableGB))
	sb.WriteString(fmt.Sprintf("Cached:        %.1fGB\n", memory.CachedGB))
	sb.WriteString(fmt.Sprintf("Buffers:       %.1fGB\n", memory.BuffersGB))

	// Swap information
	if memory.SwapTotalGB > 0 {
		swapColor := color.New(color.FgGreen)
		if memory.SwapPercent > SwapThreshold {
			swapColor = color.New(color.FgRed, color.Bold)
		} else if memory.SwapPercent > 0 {
			swapColor = color.New(color.FgYellow)
		}

		sb.WriteString(fmt.Sprintf("Swap Usage:    %s / %.1fGB (%s)\n",
			swapColor.Sprintf("%.1fGB", memory.SwapUsedGB),
			memory.SwapTotalGB,
			swapColor.Sprintf("%.1f%%", memory.SwapPercent)))
	}

	// Memory activity
	if memory.PageFaults > 0 {
		sb.WriteString(fmt.Sprintf("Page Faults:   %s (Major: %s)\n",
			o.formatNumber(memory.PageFaults),
			o.formatNumber(memory.MajorPageFaults)))
	}

	if memory.SwapIns > 0 || memory.SwapOuts > 0 {
		sb.WriteString(fmt.Sprintf("Swap I/O:      In: %s, Out: %s\n",
			o.formatNumber(memory.SwapIns),
			o.formatNumber(memory.SwapOuts)))
	}

	// Issues
	if len(memory.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Issues:\n", emoji.Warning))
		for _, issue := range memory.Issues {
			sb.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
	}

	return sb.String()
}

// formatNumber formats large numbers with appropriate units
func (o *SysloadV2Options) formatNumber(num uint64) string {
	if num >= 1000000000 {
		return fmt.Sprintf("%.1fB", float64(num)/1000000000)
	} else if num >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	}
	return fmt.Sprintf("%d", num)
}

// formatDuration formats duration in human-readable format
func (o *SysloadV2Options) formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// formatBytes formats bytes in human-readable format
func (o *SysloadV2Options) formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatPercentage formats percentage with color coding
func (o *SysloadV2Options) formatPercentage(value, threshold float64) string {
	var c *color.Color
	if value >= threshold*1.2 {
		c = color.New(color.FgRed, color.Bold)
	} else if value >= threshold {
		c = color.New(color.FgYellow)
	} else {
		c = color.New(color.FgGreen)
	}
	return c.Sprintf("%.1f%%", value)
}

// getColorForValue returns appropriate color based on value and thresholds
func (o *SysloadV2Options) getColorForValue(value, warning, critical float64) *color.Color {
	if value >= critical {
		return color.New(color.FgRed, color.Bold)
	} else if value >= warning {
		return color.New(color.FgYellow)
	}
	return color.New(color.FgGreen)
}

// truncateString truncates string to specified length with ellipsis
func (o *SysloadV2Options) truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

// formatUptime formats uptime in a readable format
func (o *SysloadV2Options) formatUptime(uptime time.Duration) string {
	totalSeconds := int(uptime.Seconds())
	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d minutes, %d seconds", minutes, seconds)
	} else {
		return fmt.Sprintf("%d seconds", seconds)
	}
}

// isTerminal checks if the output is a terminal
func (o *SysloadV2Options) isTerminal() bool {
	if o.SaveToFile != "" {
		return false
	}

	// Check if stdout is a terminal
	file, ok := o.Out.(*os.File)
	if !ok {
		return false
	}

	stat, err := file.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode() & os.ModeCharDevice) != 0
}

// ensureColorSupport ensures color support is properly configured
func (o *SysloadV2Options) ensureColorSupport() {
	if o.NoColor || !o.isTerminal() {
		color.NoColor = true
	} else {
		// Check for color support
		term := os.Getenv("TERM")
		if term == "" || term == "dumb" {
			color.NoColor = true
		}
	}
}

// validateSystemSupport validates that the system supports required features
func (o *SysloadV2Options) validateSystemSupport() error {
	// Check if /proc is available (Linux-specific features)
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		return fmt.Errorf("this tool requires /proc filesystem (Linux/Unix systems)")
	}

	// Check if we can read basic system information
	if _, err := os.Stat("/proc/stat"); os.IsNotExist(err) {
		return fmt.Errorf("cannot access /proc/stat - insufficient permissions or unsupported system")
	}

	if _, err := os.Stat("/proc/meminfo"); os.IsNotExist(err) {
		return fmt.Errorf("cannot access /proc/meminfo - insufficient permissions or unsupported system")
	}

	return nil
}

// Emergency error handling and cleanup functions
func (o *SysloadV2Options) handlePanic() {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "%s System monitoring encountered an error: %v\n", emoji.SosButton, r)
		fmt.Fprintf(os.Stderr, "Please report this issue with your system details.\n")
		os.Exit(1)
	}
}

// cleanup performs cleanup operations
func (o *SysloadV2Options) cleanup() {
	// Reset color settings
	color.NoColor = false

	// Any other cleanup operations can be added here
}

// printDebugInfo prints debug information if enabled
func (o *SysloadV2Options) printDebugInfo(format string, args ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// Version information
const (
	Version     = "2.0.0"
	BuildDate   = "2024-01-15"
	Description = "Advanced system diagnostics and performance troubleshooting tool"
)

// printVersion prints version information
func printVersion() {
	fmt.Printf("osbuilder sysload %s (built %s)\n", Version, BuildDate)
	fmt.Printf("%s\n", Description)
	fmt.Printf("https://github.com/your-org/osbuilder\n")
}

// init function for package initialization
func init() {
	// Set up color support detection
	color.NoColor = false

	// Set up emoji support
	// emoji.Println("System diagnostics tool initializing...")
}

// Example usage function for documentation
func ExampleUsage() {
	fmt.Println("Example usage:")
	fmt.Println("  osbuilder sysload                    # Basic system overview")
	fmt.Println("  osbuilder sysload --detailed         # Comprehensive diagnostics")
	fmt.Println("  osbuilder sysload --watch            # Continuous monitoring")
	fmt.Println("  osbuilder sysload -o json            # JSON output")
	fmt.Println("  osbuilder sysload --top 10           # Show top 10 processes")
	fmt.Println("  osbuilder sysload --save-to-file report.json  # Save to file")
}

// connectionTypeToString converts connection type from uint32 to string
func (o *SysloadV2Options) connectionTypeToString(connType uint32) string {
	switch connType {
	case 1:
		return "tcp"
	case 2:
		return "udp"
	case 3:
		return "tcp6"
	case 4:
		return "udp6"
	case 5:
		return "unix"
	case 6:
		return "unixgram"
	case 7:
		return "unixpacket"
	default:
		return fmt.Sprintf("unknown(%d)", connType)
	}
}

// addressFamilyToString converts address family from uint32 to string
func (o *SysloadV2Options) addressFamilyToString(family uint32) string {
	switch family {
	case 1:
		return "unix"
	case 2:
		return "inet"
	case 10:
		return "inet6"
	case 16:
		return "netlink"
	case 17:
		return "packet"
	default:
		return fmt.Sprintf("family_%d", family)
	}
}

// formatLoadMetricsV2 formats load average metrics
func (o *SysloadV2Options) formatLoadMetricsV2(load *LoadMetricsV2) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Load & Process Information\n", emoji.BalanceScale))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// Load averages with per-core ratios
	sb.WriteString(fmt.Sprintf("Load Average:  %.2f, %.2f, %.2f\n",
		load.Load1, load.Load5, load.Load15))
	sb.WriteString(fmt.Sprintf("Per Core:      %.2f, %.2f, %.2f\n",
		load.LoadPerCore1, load.LoadPerCore5, load.LoadPerCore15))

	// CPU saturation
	saturationColor := color.New(color.FgGreen)
	if load.CPUSaturation > 100 {
		saturationColor = color.New(color.FgRed, color.Bold)
	} else if load.CPUSaturation > 80 {
		saturationColor = color.New(color.FgYellow)
	}

	sb.WriteString(fmt.Sprintf("CPU Saturation: %s\n",
		saturationColor.Sprintf("%.1f%%", load.CPUSaturation)))

	// Process counts
	if load.TotalProcs > 0 {
		sb.WriteString(fmt.Sprintf("\nProcess States:\n"))
		sb.WriteString(fmt.Sprintf("  Total:       %d\n", load.TotalProcs))
		sb.WriteString(fmt.Sprintf("  Running:     %d\n", load.RunningProcs))
		sb.WriteString(fmt.Sprintf("  Blocked:     %d\n", load.BlockedProcs))

		if load.ZombieProcs > 0 {
			zombieColor := color.New(color.FgYellow)
			if load.ZombieProcs > ZombieProcessLimit {
				zombieColor = color.New(color.FgRed, color.Bold)
			}
			sb.WriteString(fmt.Sprintf("  Zombies:     %s\n",
				zombieColor.Sprintf("%d", load.ZombieProcs)))
		}
	}

	// Issues
	if len(load.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Issues:\n", emoji.Warning))
		for _, issue := range load.Issues {
			sb.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
	}

	return sb.String()
}

// formatDiskMetricsV2 formats disk metrics
func (o *SysloadV2Options) formatDiskMetricsV2(disk *DiskMetricsV2) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Disk Information\n", emoji.FloppyDisk))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// Partition usage
	sb.WriteString("Disk Usage:\n")
	for _, partition := range disk.Partitions {
		break
		usageColor := color.New(color.FgGreen)
		if len(partition.Issues) > 0 {
			usageColor = color.New(color.FgRed, color.Bold)
		} else if partition.UsagePercent > DiskThresholdV2 {
			usageColor = color.New(color.FgYellow)
		}

		sb.WriteString(fmt.Sprintf("  %-20s %s / %.1fGB (%s) %s\n",
			partition.Mountpoint,
			usageColor.Sprintf("%.1fGB", partition.UsedGB),
			partition.TotalGB,
			usageColor.Sprintf("%.1f%%", partition.UsagePercent),
			partition.Filesystem))

		// Show inode usage if high
		if partition.InodesPercent > 50 {
			inodeColor := color.New(color.FgGreen)
			if partition.InodesPercent > InodeThreshold {
				inodeColor = color.New(color.FgRed, color.Bold)
			} else if partition.InodesPercent > 70 {
				inodeColor = color.New(color.FgYellow)
			}
			sb.WriteString(fmt.Sprintf("    Inodes: %s\n",
				inodeColor.Sprintf("%.1f%% used", partition.InodesPercent)))
		}

		// Show issues
		if len(partition.Issues) > 0 {
			for _, issue := range partition.Issues {
				sb.WriteString(fmt.Sprintf("    %s %s\n", emoji.Warning, issue))
			}
		}
	}

	// I/O statistics
	if len(disk.IOStats) > 0 {
		sb.WriteString(fmt.Sprintf("\nDisk I/O Summary:\n"))
		sb.WriteString(fmt.Sprintf("  Total Read:    %.1f MB (%s IOPS)\n",
			disk.TotalReadMB, o.formatNumber(disk.TotalReadIOPS)))
		sb.WriteString(fmt.Sprintf("  Total Write:   %.1f MB (%s IOPS)\n",
			disk.TotalWriteMB, o.formatNumber(disk.TotalWriteIOPS)))

		// Show per-device stats if detailed
		if o.Detailed && len(disk.IOStats) <= 5 {
			sb.WriteString(fmt.Sprintf("\nPer-Device I/O:\n"))
			for _, ioStat := range disk.IOStats {
				utilColor := color.New(color.FgGreen)
				if ioStat.Utilization > 90 {
					utilColor = color.New(color.FgRed, color.Bold)
				} else if ioStat.Utilization > 70 {
					utilColor = color.New(color.FgYellow)
				}

				sb.WriteString(fmt.Sprintf("  %-10s Util: %s",
					ioStat.Device,
					utilColor.Sprintf("%.1f%%", ioStat.Utilization)))

				if ioStat.ReadLatencyMS > 0 || ioStat.WriteLatencyMS > 0 {
					sb.WriteString(fmt.Sprintf(" | Latency: R:%.1fms W:%.1fms",
						ioStat.ReadLatencyMS, ioStat.WriteLatencyMS))
				}
				sb.WriteString("\n")

				// Show device issues
				if len(ioStat.Issues) > 0 {
					for _, issue := range ioStat.Issues {
						sb.WriteString(fmt.Sprintf("    %s %s\n", emoji.Warning, issue))
					}
				}
			}
		}
	}

	// Overall issues
	if len(disk.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Issues:\n", emoji.Warning))
		for _, issue := range disk.Issues {
			sb.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
	}

	return sb.String()
}

// formatNetworkMetrics formats network metrics
func (o *SysloadV2Options) formatNetworkMetrics(network *NetworkMetrics) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Network Information\n", emoji.GlobeWithMeridians))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// Overall network statistics
	sb.WriteString(fmt.Sprintf("Network Traffic:\n"))
	sb.WriteString(fmt.Sprintf("  RX: %.2f MB (%s packets)\n",
		float64(network.TotalRxBytes)/1024/1024, o.formatNumber(network.TotalRxPackets)))
	sb.WriteString(fmt.Sprintf("  TX: %.2f MB (%s packets)\n",
		float64(network.TotalTxBytes)/1024/1024, o.formatNumber(network.TotalTxPackets)))

	// Error statistics
	if network.TotalRxErrors > 0 || network.TotalTxErrors > 0 {
		sb.WriteString(fmt.Sprintf("  Errors: RX:%d TX:%d\n",
			network.TotalRxErrors, network.TotalTxErrors))
	}

	if network.TotalRxDropped > 0 || network.TotalTxDropped > 0 {
		sb.WriteString(fmt.Sprintf("  Dropped: RX:%d TX:%d\n",
			network.TotalRxDropped, network.TotalTxDropped))
	}

	// Connection statistics
	if network.ActiveConnections > 0 {
		sb.WriteString(fmt.Sprintf("\nConnections:\n"))
		sb.WriteString(fmt.Sprintf("  Active:      %d\n", network.ActiveConnections))

		if network.TimeWaitConnections > 0 {
			timeWaitColor := color.New(color.FgGreen)
			if network.TimeWaitConnections > 1000 {
				timeWaitColor = color.New(color.FgYellow)
			}
			if network.TimeWaitConnections > 5000 {
				timeWaitColor = color.New(color.FgRed, color.Bold)
			}
			sb.WriteString(fmt.Sprintf("  TIME_WAIT:   %s\n",
				timeWaitColor.Sprintf("%d", network.TimeWaitConnections)))
		}
	}

	// Interface details (show only problematic or important interfaces)
	problematicInterfaces := []NetworkInterface{}
	for _, iface := range network.Interfaces {
		if len(iface.Issues) > 0 || iface.ErrorRate > 0 || iface.Name == "eth0" || iface.Name == "wlan0" {
			problematicInterfaces = append(problematicInterfaces, iface)
		}
	}

	if len(problematicInterfaces) > 0 && o.Detailed {
		sb.WriteString(fmt.Sprintf("\nInterface Details:\n"))
		for _, iface := range problematicInterfaces {
			if iface.Name == "lo" {
				continue // Skip loopback unless there are issues
			}

			statusIcon := emoji.CheckMark
			if !iface.IsUp {
				statusIcon = emoji.CrossMark
			}

			sb.WriteString(fmt.Sprintf("  %s %-10s", statusIcon, iface.Name))

			if len(iface.Addrs) > 0 {
				sb.WriteString(fmt.Sprintf(" %s", iface.Addrs[0]))
			}

			if iface.ErrorRate > 0 || iface.DropRate > 0 {
				sb.WriteString(fmt.Sprintf(" | Err:%.2f%% Drop:%.2f%%",
					iface.ErrorRate, iface.DropRate))
			}
			sb.WriteString("\n")

			// Show interface issues
			if len(iface.Issues) > 0 {
				for _, issue := range iface.Issues {
					sb.WriteString(fmt.Sprintf("    %s %s\n", emoji.Warning, issue))
				}
			}
		}
	}

	// Listening ports (show only if detailed and not too many)
	if o.Detailed && len(network.ListenPorts) > 0 && len(network.ListenPorts) <= 20 {
		sb.WriteString(fmt.Sprintf("\nListening Ports:\n"))
		for _, port := range network.ListenPorts {
			sb.WriteString(fmt.Sprintf("  %s/%d", port.Protocol, port.Port))
			if port.Process != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", port.Process))
			}
			sb.WriteString("\n")
		}
	}

	// Overall issues
	if len(network.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Issues:\n", emoji.Warning))
		for _, issue := range network.Issues {
			sb.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
	}

	return sb.String()
}

// formatProcessStats formats process statistics
func (o *SysloadV2Options) formatProcessStats(stats *ProcessStats) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Process Statistics\n", emoji.Gear))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	sb.WriteString(fmt.Sprintf("Total Processes: %d\n", stats.Total))
	sb.WriteString(fmt.Sprintf("  Running:       %d\n", stats.Running))
	sb.WriteString(fmt.Sprintf("  Sleeping:      %d\n", stats.Sleeping))

	if stats.Zombie > 0 {
		zombieColor := color.New(color.FgYellow)
		if stats.Zombie > ZombieProcessLimit {
			zombieColor = color.New(color.FgRed, color.Bold)
		}
		sb.WriteString(fmt.Sprintf("  Zombies:       %s\n",
			zombieColor.Sprintf("%d", stats.Zombie)))
	}

	if stats.Stopped > 0 {
		sb.WriteString(fmt.Sprintf("  Stopped:       %d\n", stats.Stopped))
	}

	if stats.Wait > 0 {
		sb.WriteString(fmt.Sprintf("  Waiting:       %d\n", stats.Wait))
	}

	// Resource usage summary
	sb.WriteString(fmt.Sprintf("\nResource Usage:\n"))
	if stats.HighCPUProcesses > 0 {
		sb.WriteString(fmt.Sprintf("  High CPU:      %d processes\n", stats.HighCPUProcesses))
	}
	if stats.HighMemProcesses > 0 {
		sb.WriteString(fmt.Sprintf("  High Memory:   %d processes\n", stats.HighMemProcesses))
	}
	if stats.HighIOProcesses > 0 {
		sb.WriteString(fmt.Sprintf("  High I/O:      %d processes\n", stats.HighIOProcesses))
	}

	// System resources
	if stats.ThreadsTotal > 0 {
		sb.WriteString(fmt.Sprintf("Total Threads:   %s\n", o.formatNumber(uint64(stats.ThreadsTotal))))
	}

	// File descriptor usage
	if stats.FDsLimit > 0 {
		fdColor := color.New(color.FgGreen)
		if stats.FDsUsagePercent > FDThreshold {
			fdColor = color.New(color.FgRed, color.Bold)
		} else if stats.FDsUsagePercent > 60 {
			fdColor = color.New(color.FgYellow)
		}

		sb.WriteString(fmt.Sprintf("File Descriptors: %s / %s (%s)\n",
			o.formatNumber(uint64(stats.FDsTotal)),
			o.formatNumber(uint64(stats.FDsLimit)),
			fdColor.Sprintf("%.1f%%", stats.FDsUsagePercent)))
	}

	// Issues
	if len(stats.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s Issues:\n", emoji.Warning))
		for _, issue := range stats.Issues {
			sb.WriteString(fmt.Sprintf("  • %s\n", issue))
		}
	}

	return sb.String()
}

// formatProcessTable formats a table of processes
func (o *SysloadV2Options) formatProcessTable(title string, processes []ProcessInfoV2, sortBy string) string {
	if len(processes) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s\n", emoji.BarChart, title))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// Table header
	switch sortBy {
	case "cpu":
		sb.WriteString(fmt.Sprintf("%-8s %-12s %6s %8s %-15s %s\n",
			"PID", "USER", "CPU%", "MEM(MB)", "COMMAND", "STATUS"))
	case "memory":
		sb.WriteString(fmt.Sprintf("%-8s %-12s %8s %6s %-15s %s\n",
			"PID", "USER", "MEM(MB)", "CPU%", "COMMAND", "STATUS"))
	case "io":
		sb.WriteString(fmt.Sprintf("%-8s %-12s %10s %6s %-15s %s\n",
			"PID", "USER", "I/O(MB)", "CPU%", "COMMAND", "STATUS"))
	}
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	// Process rows
	for _, proc := range processes {
		// Color code based on resource usage and issues
		var nameColor *color.Color = color.New(color.FgWhite)
		if len(proc.Issues) > 0 {
			nameColor = color.New(color.FgRed)
		} else if proc.CPUPercent > CPUThresholdV2 || proc.MemoryPercent > 20 {
			nameColor = color.New(color.FgYellow)
		}

		// Truncate username and command for display
		username := proc.Username
		if len(username) > 12 {
			username = username[:9] + "..."
		}

		command := proc.Command
		if len(command) > 15 {
			command = command[:12] + "..."
		}

		status := proc.Status
		if status == "Z" {
			status = color.RedString("ZOMBIE")
		} else if status == "D" {
			status = color.YellowString("BLOCK")
		}

		switch sortBy {
		case "cpu":
			sb.WriteString(fmt.Sprintf("%-8d %-12s %6.1f %8.1f %-15s %s",
				proc.PID, username, proc.CPUPercent, proc.MemoryMB,
				nameColor.Sprint(command), status))
		case "memory":
			sb.WriteString(fmt.Sprintf("%-8d %-12s %8.1f %6.1f %-15s %s",
				proc.PID, username, proc.MemoryMB, proc.CPUPercent,
				nameColor.Sprint(command), status))
		case "io":
			ioTotal := proc.IOReadMB + proc.IOWriteMB
			sb.WriteString(fmt.Sprintf("%-8d %-12s %10.1f %6.1f %-15s %s",
				proc.PID, username, ioTotal, proc.CPUPercent,
				nameColor.Sprint(command), status))
		}

		// Show issues inline
		if len(proc.Issues) > 0 {
			sb.WriteString(fmt.Sprintf(" %s", emoji.Warning))
		}

		sb.WriteString("\n")

		// Show detailed issues if present
		if len(proc.Issues) > 0 && o.Detailed {
			for _, issue := range proc.Issues {
				sb.WriteString(fmt.Sprintf("    %s %s\n", emoji.RightArrow, issue))
			}
		}
	}

	return sb.String()
}

// formatDetailedDiagnostics formats detailed diagnostic information
func (o *SysloadV2Options) formatDetailedDiagnostics(metrics *SystemMetricsV2) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Detailed Diagnostics\n", emoji.Microscope))
	sb.WriteString(strings.Repeat("=", 60) + "\n")

	// Performance analysis
	if metrics.Performance.PerformanceScore > 0 {
		sb.WriteString(fmt.Sprintf("%s Performance Analysis\n", emoji.Rocket))
		sb.WriteString(strings.Repeat("-", 30) + "\n")

		scoreColor := color.New(color.FgGreen)
		if metrics.Performance.PerformanceScore < 60 {
			scoreColor = color.New(color.FgRed, color.Bold)
		} else if metrics.Performance.PerformanceScore < 80 {
			scoreColor = color.New(color.FgYellow)
		}

		sb.WriteString(fmt.Sprintf("Performance Score: %s/100\n",
			scoreColor.Sprintf("%d", metrics.Performance.PerformanceScore)))

		// Response times
		sb.WriteString(fmt.Sprintf("\nResponse Times:\n"))
		sb.WriteString(fmt.Sprintf("  CPU Scheduling: %.1fms\n", metrics.Performance.ResponseTime.AvgCPUScheduling))
		sb.WriteString(fmt.Sprintf("  Disk I/O:       %.1fms\n", metrics.Performance.ResponseTime.AvgDiskIO))
		sb.WriteString(fmt.Sprintf("  Memory Access:  %.1fms\n", metrics.Performance.ResponseTime.AvgMemoryAccess))
		sb.WriteString(fmt.Sprintf("  Network I/O:    %.1fms\n", metrics.Performance.ResponseTime.AvgNetworkIO))

		// Throughput
		sb.WriteString(fmt.Sprintf("\nThroughput:\n"))
		if metrics.Performance.Throughput.CPUInstructionsPerSec > 0 {
			sb.WriteString(fmt.Sprintf("  CPU Instructions: %s/sec\n",
				o.formatNumber(metrics.Performance.Throughput.CPUInstructionsPerSec)))
		}
		if metrics.Performance.Throughput.DiskIOPS > 0 {
			sb.WriteString(fmt.Sprintf("  Disk IOPS:        %s\n",
				o.formatNumber(metrics.Performance.Throughput.DiskIOPS)))
		}
		if metrics.Performance.Throughput.NetworkPPS > 0 {
			sb.WriteString(fmt.Sprintf("  Network PPS:      %s\n",
				o.formatNumber(metrics.Performance.Throughput.NetworkPPS)))
		}

		// Bottlenecks
		if len(metrics.Performance.Bottlenecks) > 0 {
			sb.WriteString(fmt.Sprintf("\n%s Identified Bottlenecks:\n", emoji.StopSign))
			for _, bottleneck := range metrics.Performance.Bottlenecks {
				severityColor := color.New(color.FgGreen)
				switch bottleneck.Severity {
				case "critical":
					severityColor = color.New(color.FgRed, color.Bold)
				case "high":
					severityColor = color.New(color.FgRed)
				case "medium":
					severityColor = color.New(color.FgYellow)
				}

				sb.WriteString(fmt.Sprintf("  %s %s (%s impact: %.1f%%)\n",
					emoji.TriangularFlag,
					bottleneck.Description,
					severityColor.Sprint(strings.ToUpper(bottleneck.Severity)),
					bottleneck.Impact))
				sb.WriteString(fmt.Sprintf("    Cause: %s\n", bottleneck.Cause))
				sb.WriteString(fmt.Sprintf("    Solution: %s\n", bottleneck.Solution))
			}
		}

		sb.WriteString("\n")
	}

	// Resource pressure analysis
	sb.WriteString(fmt.Sprintf("%s Resource Pressure\n", emoji.Volcano))
	sb.WriteString(strings.Repeat("-", 30) + "\n")

	pressureColor := color.New(color.FgGreen)
	switch metrics.SystemHealth.ResourcePressure.OverallPressure {
	case "critical":
		pressureColor = color.New(color.FgRed, color.Bold)
	case "high":
		pressureColor = color.New(color.FgRed)
	case "medium":
		pressureColor = color.New(color.FgYellow)
	}

	sb.WriteString(fmt.Sprintf("Overall Pressure: %s\n",
		pressureColor.Sprint(strings.ToUpper(metrics.SystemHealth.ResourcePressure.OverallPressure))))

	sb.WriteString(fmt.Sprintf("CPU Pressure:     %s (%.1f%% avg)\n",
		strings.ToUpper(metrics.SystemHealth.ResourcePressure.CPUPressure.Level),
		metrics.SystemHealth.ResourcePressure.CPUPressure.Some60s))
	sb.WriteString(fmt.Sprintf("Memory Pressure:  %s (%.1f%% avg)\n",
		strings.ToUpper(metrics.SystemHealth.ResourcePressure.MemoryPressure.Level),
		metrics.SystemHealth.ResourcePressure.MemoryPressure.Some60s))
	sb.WriteString(fmt.Sprintf("I/O Pressure:     %s (%.1f%% avg)\n",
		strings.ToUpper(metrics.SystemHealth.ResourcePressure.IOPressure.Level),
		metrics.SystemHealth.ResourcePressure.IOPressure.Some60s))

	// System stability
	sb.WriteString(fmt.Sprintf("\n%s System Stability\n", emoji.Shield))
	sb.WriteString(strings.Repeat("-", 30) + "\n")

	stability := metrics.SystemHealth.StabilityIndicators
	stabilityColor := color.New(color.FgGreen)
	if stability.StabilityScore < 60 {
		stabilityColor = color.New(color.FgRed, color.Bold)
	} else if stability.StabilityScore < 80 {
		stabilityColor = color.New(color.FgYellow)
	}

	sb.WriteString(fmt.Sprintf("Stability Score:  %s/100\n",
		stabilityColor.Sprintf("%d", stability.StabilityScore)))
	sb.WriteString(fmt.Sprintf("Uptime:           %.1f days\n", stability.UptimeDays))

	if stability.OOMKills > 0 {
		sb.WriteString(fmt.Sprintf("OOM Kills:        %d\n", stability.OOMKills))
	}
	if stability.MemoryLeaks > 0 {
		sb.WriteString(fmt.Sprintf("Memory Leaks:     %d suspected\n", stability.MemoryLeaks))
	}
	if stability.DeadlockDetections > 0 {
		sb.WriteString(fmt.Sprintf("Deadlocks:        %d detected\n", stability.DeadlockDetections))
	}

	// System limits
	sb.WriteString(fmt.Sprintf("\n%s System Limits\n", emoji.Pushpin))
	sb.WriteString(strings.Repeat("-", 30) + "\n")

	limits := metrics.SystemHealth.SystemLimits
	if limits.MaxFiles > 0 {
		sb.WriteString(fmt.Sprintf("Max Files:        %s (%.1f%% used)\n",
			o.formatNumber(limits.MaxFiles), limits.FileUsagePercent))
	}
	if limits.MaxProcesses > 0 {
		sb.WriteString(fmt.Sprintf("Max Processes:    %s (%.1f%% used)\n",
			o.formatNumber(limits.MaxProcesses), limits.ProcessUsagePercent))
	}
	if limits.MaxThreads > 0 {
		sb.WriteString(fmt.Sprintf("Max Threads:      %s (%.1f%% used)\n",
			o.formatNumber(limits.MaxThreads), limits.ThreadUsagePercent))
	}

	sb.WriteString("\n")
	return sb.String()
}

// formatRecommendations formats system recommendations
func (o *SysloadV2Options) formatRecommendations(recommendations []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s System Recommendations\n", emoji.LightBulb))
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	for i, recommendation := range recommendations {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, recommendation))
	}

	return sb.String()
}
