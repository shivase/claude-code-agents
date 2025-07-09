package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config - アプリケーション設定
type Config struct {
	// プロセス制御
	MaxProcesses    int           `json:"max_processes"`
	RestartDelay    time.Duration `json:"restart_delay"`
	ProcessTimeout  time.Duration `json:"process_timeout"`
	StartupTimeout  time.Duration `json:"startup_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// リソース制御
	MaxMemoryMB   int64   `json:"max_memory_mb"`
	MaxCPUPercent float64 `json:"max_cpu_percent"`

	// 認証設定
	AuthCheckInterval time.Duration `json:"auth_check_interval"`

	// ログ設定
	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`

	// パス設定
	ClaudePath      string `json:"claude_path"`
	InstructionsDir string `json:"instructions_dir"`

	// 監視設定
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MaxRestartAttempts  int           `json:"max_restart_attempts"`
}

// DefaultConfig - デフォルト設定
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	// 統一された設定ディレクトリパス
	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claud-code-agents")

	return &Config{
		MaxProcesses:        runtime.NumCPU(),
		RestartDelay:        5 * time.Second,
		ProcessTimeout:      30 * time.Second,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		MaxMemoryMB:         1024,
		MaxCPUPercent:       80.0,
		AuthCheckInterval:   30 * time.Minute,
		LogLevel:            "info",
		LogFile:             filepath.Join(claudCodeAgentsDir, "logs", "manager.log"),
		ClaudePath:          filepath.Join(claudeDir, "local", "claude"),
		InstructionsDir:     filepath.Join(claudCodeAgentsDir, "instructions"),
		HealthCheckInterval: 30 * time.Second,
		MaxRestartAttempts:  3,
	}
}

// LoadConfig - 設定ファイル読み込み
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 設定ファイルが存在しない場合はデフォルト設定で作成
		if err := SaveConfig(config, configPath); err != nil {
			return nil, err
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig - 設定ファイル保存
func SaveConfig(config *Config, configPath string) error {
	// ディレクトリ作成
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// ResourceMonitor - リソース監視
type ResourceMonitor struct {
	config *Config
}

// NewResourceMonitor - リソース監視器作成
func NewResourceMonitor(config *Config) *ResourceMonitor {
	return &ResourceMonitor{
		config: config,
	}
}

// CheckMemoryUsage - メモリ使用量チェック
func (rm *ResourceMonitor) CheckMemoryUsage() (bool, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	usedMB := int64(memStats.Sys / 1024 / 1024)

	return usedMB <= rm.config.MaxMemoryMB, nil
}

// CheckCPUUsage - CPU使用量チェック（簡易版）
func (rm *ResourceMonitor) CheckCPUUsage() (bool, error) {
	// 実際の実装では、CPUサンプリングを行う
	// ここでは簡易的にtrueを返す
	return true, nil
}

// HealthChecker - ヘルスチェック
type HealthChecker struct {
	config *Config
}

// NewHealthChecker - ヘルスチェッカー作成
func NewHealthChecker(config *Config) *HealthChecker {
	return &HealthChecker{
		config: config,
	}
}

// CheckClaudeHealth - Claude CLIヘルスチェック
func (hc *HealthChecker) CheckClaudeHealth(claudePath string) error {
	// Claude CLIの基本的なヘルスチェック
	if _, err := os.Stat(claudePath); err != nil {
		return err
	}

	// 簡易的な実行チェック
	// 実際の実装では、Claude CLIの応答性をチェックする
	return nil
}

// CheckAuthHealth - 認証状態ヘルスチェック
func (hc *HealthChecker) CheckAuthHealth() error {
	homeDir, _ := os.UserHomeDir()
	configFile := filepath.Join(homeDir, ".claude", "settings.json")

	if _, err := os.Stat(configFile); err != nil {
		return err
	}

	return nil
}
