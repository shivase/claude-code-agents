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

// SystemLoadInfo システム負荷情報
type SystemLoadInfo struct {
	LoadAvg1Min  float64 `json:"load_avg_1min"`
	LoadAvg5Min  float64 `json:"load_avg_5min"`
	LoadAvg15Min float64 `json:"load_avg_15min"`
	CPUCores     int     `json:"cpu_cores"`
	MemoryGB     float64 `json:"memory_gb"`
	Processes    int     `json:"processes"`
	Threads      int     `json:"threads"`
}

// SystemOptimizer システム最適化機能
type SystemOptimizer struct {
	config   *config.TeamConfig
	loadInfo *SystemLoadInfo
}

// NewSystemOptimizer 新しいシステム最適化インスタンスを作成
func NewSystemOptimizer(config *config.TeamConfig) *SystemOptimizer {
	return &SystemOptimizer{
		config: config,
	}
}

// GetSystemLoadInfo システム負荷情報の取得
func (so *SystemOptimizer) GetSystemLoadInfo() (*SystemLoadInfo, error) {
	// uptime コマンドで負荷平均を取得
	cmd := exec.Command("uptime")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get system load: %w", err)
	}

	loadInfo := &SystemLoadInfo{
		CPUCores: runtime.NumCPU(),
	}

	// メモリ情報を取得
	if memSize, err := getMemorySize(); err == nil {
		loadInfo.MemoryGB = float64(memSize) / (1024 * 1024 * 1024)
	}

	// uptime 出力から負荷平均を解析
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

	// プロセス情報を取得
	if processes, err := getProcessCount(); err == nil {
		loadInfo.Processes = processes
	}

	so.loadInfo = loadInfo
	return loadInfo, nil
}

// IsHighLoadCondition 高負荷状態かどうかを判定
func (so *SystemOptimizer) IsHighLoadCondition() bool {
	if so.loadInfo == nil {
		return false
	}

	// 負荷平均がCPUコア数の80%を超える場合は高負荷とみなす
	threshold := float64(so.loadInfo.CPUCores) * 0.8
	return so.loadInfo.LoadAvg1Min > threshold
}

// OptimizeSystemLoad システム負荷を最適化
func (so *SystemOptimizer) OptimizeSystemLoad() error {
	if !so.IsHighLoadCondition() {
		log.Info().Msg("System load is within normal range")
		return nil
	}

	log.Warn().Float64("load_avg", so.loadInfo.LoadAvg1Min).Int("cpu_cores", so.loadInfo.CPUCores).Msg("High system load detected")

	// 高負荷時の最適化処理
	so.optimizeClaudeProcesses()

	return nil
}

// optimizeClaudeProcesses claudeプロセスの最適化
func (so *SystemOptimizer) optimizeClaudeProcesses() {
	log.Info().Msg("Claude process optimization - feature disabled")
}

// LimitProcessResources プロセスのリソース制限を設定
func (so *SystemOptimizer) LimitProcessResources(pid int) error {
	// プロセスのリソース制限を設定
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// CPU時間制限を設定（例: 300秒）
	if err := syscall.Setrlimit(syscall.RLIMIT_CPU, &syscall.Rlimit{
		Cur: 300,
		Max: 300,
	}); err != nil {
		log.Warn().Err(err).Int("pid", pid).Msg("Failed to set CPU time limit")
	}

	// メモリ制限を設定（例: 1GB）
	if err := syscall.Setrlimit(syscall.RLIMIT_AS, &syscall.Rlimit{
		Cur: 1024 * 1024 * 1024,
		Max: 1024 * 1024 * 1024,
	}); err != nil {
		log.Warn().Err(err).Int("pid", pid).Msg("Failed to set memory limit")
	}

	_ = process // プロセスハンドルを使用
	return nil
}

// MonitorSystemLoad システム負荷の監視
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

				// 自動最適化を実行
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

// GenerateSystemReport システム負荷レポートの生成
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

		// 負荷状態の評価
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

// ヘルパー関数

// getMemorySize システムメモリサイズを取得
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

// getProcessCount プロセス数を取得
func getProcessCount() (int, error) {
	cmd := exec.Command("ps", "ax")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get process count: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	return len(lines) - 1, nil // ヘッダー行を除く
}
