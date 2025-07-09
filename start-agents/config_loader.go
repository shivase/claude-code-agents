package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/rs/zerolog/log"
)

// TeamConfig AI Team設定構造体
type TeamConfig struct {
	// Path Configurations
	ClaudeCLIPath   string
	InstructionsDir string
	WorkingDir      string
	ConfigDir       string
	LogFile         string

	// System Settings
	MaxProcesses        int
	MaxMemoryMB         int64
	MaxCPUPercent       float64
	LogLevel            string
	HealthCheckInterval time.Duration
	MaxRestartAttempts  int

	// Tmux Settings
	SessionName   string
	DefaultLayout string
	AutoAttach    bool
	PaneCount     int

	// Authentication Settings
	AuthCheckInterval time.Duration
	AuthBackupDir     string
	IDEBackupEnabled  bool

	// Performance Settings
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
	RestartDelay    time.Duration
	ProcessTimeout  time.Duration

	// Command Names
	SendCommand string
	BinaryName  string
}

// LoadTeamConfigFromPath 設定ファイルの読み込み
func LoadTeamConfigFromPath(configPath string) (*TeamConfig, error) {
	homeDir, _ := os.UserHomeDir()
	
	// ディレクトリ解決器を使用して最適な作業ディレクトリを取得
	resolver := GetGlobalDirectoryResolver()
	optimalWorkingDir := resolver.GetOptimalWorkingDirectory()

	// 統一された設定ディレクトリパス
	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "clahd-code-agents")

	// デフォルト設定
	config := &TeamConfig{
		ClaudeCLIPath:       filepath.Join(claudeDir, "local", "claude"),
		InstructionsDir:     filepath.Join(claudCodeAgentsDir, "instructions"),
		WorkingDir:          optimalWorkingDir,
		ConfigDir:           claudCodeAgentsDir,
		LogFile:             filepath.Join(claudCodeAgentsDir, "logs", "manager.log"),
		AuthBackupDir:       filepath.Join(claudCodeAgentsDir, "auth_backup"),
		MaxProcesses:        4,
		MaxMemoryMB:         1024,
		MaxCPUPercent:       80.0,
		LogLevel:            "info",
		HealthCheckInterval: 30 * time.Second,
		MaxRestartAttempts:  3,
		SessionName:         "ai-teams",
		DefaultLayout:       "integrated",
		AutoAttach:          false,
		PaneCount:           6,
		AuthCheckInterval:   30 * time.Minute,
		IDEBackupEnabled:    true,
		StartupTimeout:      10 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		RestartDelay:        5 * time.Second,
		ProcessTimeout:      30 * time.Second,
		SendCommand:         "send-agent",
		BinaryName:          "claude-code-agents",
	}

	// 設定ファイルが存在しない場合はデフォルト設定を返す
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	// 設定ファイルの読み込み
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// コメントや空行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// KEY=VALUE形式の解析
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 設定値の適用
		switch key {
		case "CLAUDE_CLI_PATH":
			config.ClaudeCLIPath = value
		case "INSTRUCTIONS_DIR":
			config.InstructionsDir = value
		case "WORKING_DIR":
			// WorkingDirはディレクトリ解決器で最適化されるためスキップ
			log.Debug().Str("config_working_dir", value).Str("optimal_working_dir", optimalWorkingDir).Msg("WorkingDir override by directory resolver")
		case "CONFIG_DIR":
			config.ConfigDir = value
		case "LOG_FILE":
			config.LogFile = value
		case "LOG_LEVEL":
			config.LogLevel = value
		case "SESSION_NAME":
			config.SessionName = value
		case "DEFAULT_LAYOUT":
			config.DefaultLayout = value
		case "AUTH_BACKUP_DIR":
			config.AuthBackupDir = value
		case "SEND_COMMAND":
			config.SendCommand = value
		case "BINARY_NAME":
			config.BinaryName = value
		case "AUTO_ATTACH":
			config.AutoAttach = value == "true"
		case "IDE_BACKUP_ENABLED":
			config.IDEBackupEnabled = value == "true"
		case "HEALTH_CHECK_INTERVAL":
			if duration, err := time.ParseDuration(value); err == nil {
				config.HealthCheckInterval = duration
			}
		case "AUTH_CHECK_INTERVAL":
			if duration, err := time.ParseDuration(value); err == nil {
				config.AuthCheckInterval = duration
			}
		case "STARTUP_TIMEOUT":
			if duration, err := time.ParseDuration(value); err == nil {
				config.StartupTimeout = duration
			}
		case "SHUTDOWN_TIMEOUT":
			if duration, err := time.ParseDuration(value); err == nil {
				config.ShutdownTimeout = duration
			}
		case "RESTART_DELAY":
			if duration, err := time.ParseDuration(value); err == nil {
				config.RestartDelay = duration
			}
		case "PROCESS_TIMEOUT":
			if duration, err := time.ParseDuration(value); err == nil {
				config.ProcessTimeout = duration
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 設定ファイル読み込み後もディレクトリ依存問題を修正
	if resolveErr := resolver.FixDirectoryDependentPaths(config); resolveErr != nil {
		log.Warn().Err(resolveErr).Msg("Failed to fix directory dependent paths")
	}

	return config, nil
}

// GetUnifiedConfigPaths 統一された設定パスを取得
func GetUnifiedConfigPaths() *ConfigPaths {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// エラーの場合は相対パスを返す
		return &ConfigPaths{
			ClaudeDir:          ".claude",
			CloudCodeAgentsDir: ".claud-code-agents",
			TeamConfigPath:     ".claud-code-agents.conf",
			MainConfigPath:     "manager.json",
			LogsDir:            "logs",
			InstructionsDir:    "instructions",
			AuthBackupDir:      "auth_backup",
		}
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claud-code-agents")

	return &ConfigPaths{
		ClaudeDir:          claudeDir,
		CloudCodeAgentsDir: claudCodeAgentsDir,
		TeamConfigPath:     filepath.Join(claudCodeAgentsDir, "agents.conf"),
		MainConfigPath:     filepath.Join(claudCodeAgentsDir, "manager.json"),
		LogsDir:            filepath.Join(claudCodeAgentsDir, "logs"),
		InstructionsDir:    filepath.Join(claudCodeAgentsDir, "instructions"),
		AuthBackupDir:      filepath.Join(claudCodeAgentsDir, "auth_backup"),
		ClaudeCLIPath:      filepath.Join(claudeDir, "local", "claude"),
	}
}

// ConfigPaths 設定パス構造体
type ConfigPaths struct {
	ClaudeDir          string
	CloudCodeAgentsDir string
	TeamConfigPath     string
	MainConfigPath     string
	LogsDir            string
	InstructionsDir    string
	AuthBackupDir      string
	ClaudeCLIPath      string
}

// EnsureDirectories 必要なディレクトリを作成
func (cp *ConfigPaths) EnsureDirectories() error {
	directories := []string{
		cp.CloudCodeAgentsDir,
		cp.LogsDir,
		cp.InstructionsDir,
		cp.AuthBackupDir,
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// LoadUnifiedConfig 統一された設定読み込み
func LoadUnifiedConfig() (*UnifiedConfig, error) {
	paths := GetUnifiedConfigPaths()
	
	// 必要なディレクトリを作成
	if err := paths.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to ensure directories: %w", err)
	}

	// TeamConfigの読み込み
	teamConfig, err := LoadTeamConfigFromPath(paths.TeamConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load team config: %w", err)
	}

	// MainConfigの読み込み
	mainConfig, err := LoadConfig(paths.MainConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load main config: %w", err)
	}

	// CommonConfigの読み込み
	commonConfig := GetCommonConfig()

	return &UnifiedConfig{
		Paths:      paths,
		Team:       teamConfig,
		Main:       mainConfig,
		Common:     commonConfig,
	}, nil
}

// UnifiedConfig 統一された設定構造体
type UnifiedConfig struct {
	Paths  *ConfigPaths
	Team   *TeamConfig
	Main   *Config
	Common *CommonConfig
}

// GetEffectiveConfig 有効な設定値を取得（優先順位: Team > Main > Common）
func (uc *UnifiedConfig) GetEffectiveConfig() *EffectiveConfig {
	return &EffectiveConfig{
		MaxProcesses:        uc.Team.MaxProcesses,
		MaxMemoryMB:         uc.Team.MaxMemoryMB,
		MaxCPUPercent:       uc.Team.MaxCPUPercent,
		LogLevel:            uc.Team.LogLevel,
		ClaudeCLIPath:       uc.Team.ClaudeCLIPath,
		InstructionsDir:     uc.Team.InstructionsDir,
		WorkingDir:          uc.Team.WorkingDir,
		ConfigDir:           uc.Team.ConfigDir,
		LogFile:             uc.Team.LogFile,
		AuthBackupDir:       uc.Team.AuthBackupDir,
		StartupTimeout:      uc.Team.StartupTimeout,
		ShutdownTimeout:     uc.Team.ShutdownTimeout,
		ProcessTimeout:      uc.Team.ProcessTimeout,
		RestartDelay:        uc.Team.RestartDelay,
		HealthCheckInterval: uc.Team.HealthCheckInterval,
		AuthCheckInterval:   uc.Team.AuthCheckInterval,
		MaxRestartAttempts:  uc.Team.MaxRestartAttempts,
		SessionName:         uc.Team.SessionName,
		DefaultLayout:       uc.Team.DefaultLayout,
		AutoAttach:          uc.Team.AutoAttach,
		PaneCount:           uc.Team.PaneCount,
		IDEBackupEnabled:    uc.Team.IDEBackupEnabled,
		SendCommand:         uc.Team.SendCommand,
		BinaryName:          uc.Team.BinaryName,
	}
}

// EffectiveConfig 有効な設定値
type EffectiveConfig struct {
	MaxProcesses        int
	MaxMemoryMB         int64
	MaxCPUPercent       float64
	LogLevel            string
	ClaudeCLIPath       string
	InstructionsDir     string
	WorkingDir          string
	ConfigDir           string
	LogFile             string
	AuthBackupDir       string
	StartupTimeout      time.Duration
	ShutdownTimeout     time.Duration
	ProcessTimeout      time.Duration
	RestartDelay        time.Duration
	HealthCheckInterval time.Duration
	AuthCheckInterval   time.Duration
	MaxRestartAttempts  int
	SessionName         string
	DefaultLayout       string
	AutoAttach          bool
	PaneCount           int
	IDEBackupEnabled    bool
	SendCommand         string
	BinaryName          string
}

// GetTeamConfigPath 設定ファイルのパスを取得
func GetTeamConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// エラーの場合はカレントディレクトリのデフォルトファイルを返す
		return ".claud-code-agents.conf"
	}

	// 統一された設定ディレクトリパス
	configDir := filepath.Join(homeDir, ".claude", "claud-code-agents")
	
	// 優先順位1: claud-code-agentsディレクトリ内の設定ファイル
	claudConfigPath := filepath.Join(configDir, "agents.conf")
	if _, err := os.Stat(claudConfigPath); err == nil {
		return claudConfigPath
	}

	// 優先順位2: ホームディレクトリの設定ファイル
	homeConfigPath := filepath.Join(homeDir, ".claud-code-agents.conf")
	if _, err := os.Stat(homeConfigPath); err == nil {
		return homeConfigPath
	}

	// 優先順位3: カレントディレクトリの設定ファイル
	currentConfigPath := ".claud-code-agents.conf"
	if _, err := os.Stat(currentConfigPath); err == nil {
		return currentConfigPath
	}

	// デフォルトパス（統一されたディレクトリ内）
	return claudConfigPath
}

// LoadTeamConfig 設定ファイルの読み込み（パラメータなし版）
func LoadTeamConfig() (*TeamConfig, error) {
	configPath := GetTeamConfigPath()
	return LoadTeamConfigFromPath(configPath)
}
