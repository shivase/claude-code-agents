package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Config represents application settings
type Config struct {
	// Process control
	MaxProcesses    int           `json:"max_processes"`
	RestartDelay    time.Duration `json:"restart_delay"`
	ProcessTimeout  time.Duration `json:"process_timeout"`
	StartupTimeout  time.Duration `json:"startup_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// Resource control
	MaxMemoryMB   int64   `json:"max_memory_mb"`
	MaxCPUPercent float64 `json:"max_cpu_percent"`

	// Authentication settings
	AuthCheckInterval time.Duration `json:"auth_check_interval"`

	// Logging settings
	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`

	// Path settings
	ClaudePath      string `json:"claude_path"`
	InstructionsDir string `json:"instructions_dir"`

	// Monitoring settings
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MaxRestartAttempts  int           `json:"max_restart_attempts"`

	// Developer settings
	DevCount int `json:"dev_count"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	// Unified configuration directory path
	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claude-code-agents")

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
		DevCount:            4,
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create with default settings if config file doesn't exist
		if err := SaveConfig(config, configPath); err != nil {
			return nil, err
		}
		return config, nil
	}

	// Path normalization and directory traversal prevention
	cleanPath := filepath.Clean(configPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("config path contains directory traversal")
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Create directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// ResourceMonitor monitors system resources
type ResourceMonitor struct {
	config *Config
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(config *Config) *ResourceMonitor {
	return &ResourceMonitor{
		config: config,
	}
}

// CheckMemoryUsage checks memory usage
func (rm *ResourceMonitor) CheckMemoryUsage() (bool, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	const mb = uint64(1024 * 1024)
	usedMB := memStats.Sys / mb

	// Prevent integer overflow
	if rm.config.MaxMemoryMB < 0 {
		return false, fmt.Errorf("invalid max memory configuration: %d", rm.config.MaxMemoryMB)
	}
	if rm.config.MaxMemoryMB == 0 {
		return true, nil // No limit
	}
	// Safe conversion check
	maxMemoryMB := uint64(rm.config.MaxMemoryMB) // #nosec G115
	return usedMB <= maxMemoryMB, nil
}

// CheckCPUUsage checks CPU usage (simplified version)
func (rm *ResourceMonitor) CheckCPUUsage() (bool, error) {
	// In actual implementation, perform CPU sampling
	// Here we simply return true
	return true, nil
}

// HealthChecker performs health checks
type HealthChecker struct {
	config *Config
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *Config) *HealthChecker {
	return &HealthChecker{
		config: config,
	}
}

// CheckClaudeHealth checks Claude CLI health
func (hc *HealthChecker) CheckClaudeHealth(claudePath string) error {
	// Basic health check for Claude CLI
	if _, err := os.Stat(claudePath); err != nil {
		return err
	}

	// Simple execution check
	// In actual implementation, check Claude CLI responsiveness
	return nil
}

// CheckAuthHealth checks authentication health
func (hc *HealthChecker) CheckAuthHealth() error {
	homeDir, _ := os.UserHomeDir()
	configFile := filepath.Join(homeDir, ".claude", "settings.json")

	if _, err := os.Stat(configFile); err != nil {
		return err
	}

	return nil
}
