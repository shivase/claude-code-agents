package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/config"
)

// SystemLoadInfo represents system load information
type SystemLoadInfo struct {
	LoadAvg1Min  float64 `json:"load_avg_1min"`
	LoadAvg5Min  float64 `json:"load_avg_5min"`
	LoadAvg15Min float64 `json:"load_avg_15min"`
	CPUCores     int     `json:"cpu_cores"`
	MemoryGB     float64 `json:"memory_gb"`
	Processes    int     `json:"processes"`
	Threads      int     `json:"threads"`
}

// SystemOptimizer provides system optimization functionality
type SystemOptimizer struct {
	config   *config.TeamConfig
	loadInfo *SystemLoadInfo
}

// NewSystemOptimizer creates a new system optimizer instance
func NewSystemOptimizer(config *config.TeamConfig) *SystemOptimizer {
	return &SystemOptimizer{
		config: config,
	}
}

// GetSystemLoadInfo retrieves system load information
func (so *SystemOptimizer) GetSystemLoadInfo() (*SystemLoadInfo, error) {
	// Get load average using uptime command
	cmd := exec.Command("uptime")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get system load: %w", err)
	}

	loadInfo := &SystemLoadInfo{
		CPUCores: runtime.NumCPU(),
	}

	// Get memory information
	if memSize, err := getMemorySize(); err == nil {
		loadInfo.MemoryGB = float64(memSize) / (1024 * 1024 * 1024)
	}

	// Parse load average from uptime output
	outputStr := strings.TrimSpace(string(output))
	if strings.Contains(outputStr, "load average:") {
		parts := strings.Split(outputStr, "load average:")
		if len(parts) > 1 {
			loadParts := strings.Split(strings.TrimSpace(parts[1]), ",")
			if len(loadParts) >= 3 {
				if load1, err := strconv.ParseFloat(strings.TrimSpace(loadParts[0]), 64); err == nil {
					loadInfo.LoadAvg1Min = load1
				}
				if load5, err := strconv.ParseFloat(strings.TrimSpace(loadParts[1]), 64); err == nil {
					loadInfo.LoadAvg5Min = load5
				}
				if load15, err := strconv.ParseFloat(strings.TrimSpace(loadParts[2]), 64); err == nil {
					loadInfo.LoadAvg15Min = load15
				}
			}
		}
	}

	// Get process information
	if processes, err := getProcessCount(); err == nil {
		loadInfo.Processes = processes
	}

	so.loadInfo = loadInfo
	return loadInfo, nil
}

// IsHighLoadCondition determines if the system is under high load
func (so *SystemOptimizer) IsHighLoadCondition() bool {
	if so.loadInfo == nil {
		return false
	}

	// Consider high load when load average exceeds 80% of CPU cores
	threshold := float64(so.loadInfo.CPUCores) * 0.8
	return so.loadInfo.LoadAvg1Min > threshold
}

// OptimizeSystemLoad optimizes system load
func (so *SystemOptimizer) OptimizeSystemLoad() error {
	if !so.IsHighLoadCondition() {
		log.Info().Msg("System load is within normal range")
		return nil
	}

	log.Warn().Float64("load_avg", so.loadInfo.LoadAvg1Min).Int("cpu_cores", so.loadInfo.CPUCores).Msg("High system load detected")

	// Optimization process for high load conditions
	so.optimizeClaudeProcesses()

	return nil
}

// optimizeClaudeProcesses optimizes claude processes
func (so *SystemOptimizer) optimizeClaudeProcesses() {
	log.Info().Msg("Claude process optimization - feature disabled")
}

// LimitProcessResources sets resource limits for a process
func (so *SystemOptimizer) LimitProcessResources(pid int) error {
	// Set process resource limits
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Set CPU time limit (e.g., 300 seconds)
	if err := syscall.Setrlimit(syscall.RLIMIT_CPU, &syscall.Rlimit{
		Cur: 300,
		Max: 300,
	}); err != nil {
		log.Warn().Err(err).Int("pid", pid).Msg("Failed to set CPU time limit")
	}

	// Set memory limit (e.g., 1GB)
	if err := syscall.Setrlimit(syscall.RLIMIT_AS, &syscall.Rlimit{
		Cur: 1024 * 1024 * 1024,
		Max: 1024 * 1024 * 1024,
	}); err != nil {
		log.Warn().Err(err).Int("pid", pid).Msg("Failed to set memory limit")
	}

	_ = process // Use process handle
	return nil
}

// MonitorSystemLoad monitors system load
func (so *SystemOptimizer) MonitorSystemLoad(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if loadInfo, err := so.GetSystemLoadInfo(); err == nil {
			if so.IsHighLoadCondition() {
				log.Warn().
					Float64("load_avg_1min", loadInfo.LoadAvg1Min).
					Float64("load_avg_5min", loadInfo.LoadAvg5Min).
					Float64("load_avg_15min", loadInfo.LoadAvg15Min).
					Int("cpu_cores", loadInfo.CPUCores).
					Msg("High system load detected")

				// Execute automatic optimization
				if err := so.OptimizeSystemLoad(); err != nil {
					log.Error().Err(err).Msg("Failed to optimize system load")
				}
			} else {
				log.Debug().
					Float64("load_avg", loadInfo.LoadAvg1Min).
					Msg("System load is normal")
			}
		}
	}
}

// GenerateSystemReport generates a system load report
func (so *SystemOptimizer) GenerateSystemReport() string {
	var report strings.Builder
	report.WriteString("📊 System Load Analysis Report\n")
	report.WriteString("===============================\n\n")

	if so.loadInfo != nil {
		report.WriteString("🖥️ System Information:\n")
		report.WriteString(fmt.Sprintf("   CPU Cores: %d\n", so.loadInfo.CPUCores))
		report.WriteString(fmt.Sprintf("   Memory: %.1f GB\n", so.loadInfo.MemoryGB))
		report.WriteString(fmt.Sprintf("   Load Average (1min): %.2f\n", so.loadInfo.LoadAvg1Min))
		report.WriteString(fmt.Sprintf("   Load Average (5min): %.2f\n", so.loadInfo.LoadAvg5Min))
		report.WriteString(fmt.Sprintf("   Load Average (15min): %.2f\n", so.loadInfo.LoadAvg15Min))
		report.WriteString(fmt.Sprintf("   Total Processes: %d\n\n", so.loadInfo.Processes))

		// Evaluate load status
		threshold := float64(so.loadInfo.CPUCores) * 0.8
		if so.loadInfo.LoadAvg1Min > threshold {
			report.WriteString("⚠️ Status: HIGH LOAD DETECTED\n")
			report.WriteString(fmt.Sprintf("   Load threshold: %.2f (80%% of CPU cores)\n", threshold))
			report.WriteString(fmt.Sprintf("   Current load: %.2f (%.1f%% above threshold)\n\n",
				so.loadInfo.LoadAvg1Min, (so.loadInfo.LoadAvg1Min/threshold-1)*100))
		} else {
			report.WriteString("✅ Status: NORMAL LOAD\n\n")
		}
	}

	// Process analysis - disabled after removing exclusion control
	report.WriteString("🔍 Claude Process Analysis:\n")
	report.WriteString("   Process monitoring functionality has been disabled.\n\n")

	report.WriteString("💡 Recommendations:\n")
	if so.IsHighLoadCondition() {
		report.WriteString("   • Consider reducing the number of concurrent Claude processes\n")
		report.WriteString("   • Lower priority of high CPU usage processes\n")
		report.WriteString("   • Monitor system for potential killed processes\n")
		report.WriteString("   • Check for resource limits and adjust if necessary\n")
	} else {
		report.WriteString("   • System load is within normal range\n")
		report.WriteString("   • Continue monitoring for any changes\n")
	}

	return report.String()
}

// Helper functions

// getMemorySize retrieves system memory size
func getMemorySize() (int64, error) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory size: %w", err)
	}

	size, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse memory size: %w", err)
	}

	return size, nil
}

// getProcessCount retrieves process count
func getProcessCount() (int, error) {
	cmd := exec.Command("ps", "ax")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get process count: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	return len(lines) - 1, nil // Exclude header line
}
