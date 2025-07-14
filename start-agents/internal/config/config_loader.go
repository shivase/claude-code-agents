package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/utils"
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

	// 開発者設定
	DevCount int

	// Role-based Instructions
	POInstructionFile      string
	ManagerInstructionFile string
	DevInstructionFile     string

	// === 新規追加フィールド ===
	// 拡張instruction設定
	InstructionConfig *InstructionConfig `json:"instruction_config,omitempty"`

	// フォールバック設定
	FallbackInstructionDir string `json:"fallback_instruction_dir,omitempty"`

	// 環境設定
	Environment string `json:"environment,omitempty"` // development, production, etc.

	// バリデーション設定
	StrictValidation bool `json:"strict_validation,omitempty"`
}

// InstructionConfig 拡張instruction設定
type InstructionConfig struct {
	// 基本設定
	Base InstructionRoleConfig `json:"base"`

	// 環境別設定
	Environments map[string]InstructionRoleConfig `json:"environments,omitempty"`

	// グローバル設定
	Global InstructionGlobalConfig `json:"global,omitempty"`
}

// InstructionRoleConfig ロール別instruction設定
type InstructionRoleConfig struct {
	POInstructionPath      string `json:"po_instruction_path,omitempty"`
	ManagerInstructionPath string `json:"manager_instruction_path,omitempty"`
	DevInstructionPath     string `json:"dev_instruction_path,omitempty"`
}

// InstructionGlobalConfig グローバルinstruction設定
type InstructionGlobalConfig struct {
	DefaultExtension string        `json:"default_extension,omitempty"` // .md, .txt
	SearchPaths      []string      `json:"search_paths,omitempty"`
	CacheEnabled     bool          `json:"cache_enabled,omitempty"`
	CacheTTL         time.Duration `json:"cache_ttl,omitempty"`
}

// GetWorkingDir ConfigInterface 実装メソッド
func (tc *TeamConfig) GetWorkingDir() string    { return tc.WorkingDir }
func (tc *TeamConfig) SetWorkingDir(dir string) { tc.WorkingDir = dir }

func (tc *TeamConfig) GetClaudeCLIPath() string     { return tc.ClaudeCLIPath }
func (tc *TeamConfig) SetClaudeCLIPath(path string) { tc.ClaudeCLIPath = path }

func (tc *TeamConfig) GetInstructionsDir() string    { return tc.InstructionsDir }
func (tc *TeamConfig) SetInstructionsDir(dir string) { tc.InstructionsDir = dir }

func (tc *TeamConfig) GetConfigDir() string    { return tc.ConfigDir }
func (tc *TeamConfig) SetConfigDir(dir string) { tc.ConfigDir = dir }

func (tc *TeamConfig) GetLogFile() string     { return tc.LogFile }
func (tc *TeamConfig) SetLogFile(path string) { tc.LogFile = path }

func (tc *TeamConfig) GetAuthBackupDir() string    { return tc.AuthBackupDir }
func (tc *TeamConfig) SetAuthBackupDir(dir string) { tc.AuthBackupDir = dir }

func (tc *TeamConfig) GetDefaultLayout() string          { return tc.DefaultLayout }
func (tc *TeamConfig) GetPaneCount() int                 { return tc.PaneCount }
func (tc *TeamConfig) GetAutoAttach() bool               { return tc.AutoAttach }
func (tc *TeamConfig) GetSessionName() string            { return tc.SessionName }
func (tc *TeamConfig) GetMaxProcesses() int              { return tc.MaxProcesses }
func (tc *TeamConfig) GetLogLevel() string               { return tc.LogLevel }
func (tc *TeamConfig) GetPOInstructionFile() string      { return tc.POInstructionFile }
func (tc *TeamConfig) GetManagerInstructionFile() string { return tc.ManagerInstructionFile }
func (tc *TeamConfig) GetDevInstructionFile() string     { return tc.DevInstructionFile }

// LoadTeamConfigFromPath 設定ファイルの読み込み
func LoadTeamConfigFromPath(configPath string) (*TeamConfig, error) {
	homeDir, _ := os.UserHomeDir()

	// ディレクトリ解決器を使用して最適な作業ディレクトリを取得
	resolver := utils.GetGlobalDirectoryResolver()
	optimalWorkingDir := resolver.GetOptimalWorkingDirectory()

	// 統一された設定ディレクトリパス
	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claude-code-agents")

	// デフォルト設定
	config := &TeamConfig{
		ClaudeCLIPath:          filepath.Join(claudeDir, "local", "claude"),
		InstructionsDir:        filepath.Join(claudCodeAgentsDir, "instructions"),
		WorkingDir:             optimalWorkingDir,
		ConfigDir:              claudCodeAgentsDir,
		LogFile:                filepath.Join(claudCodeAgentsDir, "logs", "manager.log"),
		AuthBackupDir:          filepath.Join(claudCodeAgentsDir, "auth_backup"),
		MaxProcesses:           4,
		MaxMemoryMB:            1024,
		MaxCPUPercent:          80.0,
		LogLevel:               "info",
		HealthCheckInterval:    30 * time.Second,
		MaxRestartAttempts:     3,
		SessionName:            "ai-teams",
		DefaultLayout:          "integrated",
		AutoAttach:             false,
		PaneCount:              6,
		AuthCheckInterval:      30 * time.Minute,
		IDEBackupEnabled:       true,
		StartupTimeout:         10 * time.Second,
		ShutdownTimeout:        15 * time.Second,
		RestartDelay:           5 * time.Second,
		ProcessTimeout:         30 * time.Second,
		SendCommand:            "send-agent",
		BinaryName:             "claude-code-agents",
		DevCount:               4,
		POInstructionFile:      "po.md",
		ManagerInstructionFile: "manager.md",
		DevInstructionFile:     "developer.md",
	}

	// 設定ファイルが存在しない場合はデフォルト設定を返す
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	// 設定ファイルの読み込み - パスの正規化とディレクトリトラバーサル防止
	cleanPath := filepath.Clean(configPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("config path contains directory traversal")
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			// ログに記録するか適切に処理
			_, err := fmt.Fprintf(os.Stderr, "Warning: failed to close config file: %v\n", err)
			if err != nil {
				return
			}
		}
	}()

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
		case "DEV_COUNT":
			if count, err := strconv.Atoi(value); err == nil && count > 0 {
				config.DevCount = count
			}
		case "PO_INSTRUCTION_FILE":
			config.POInstructionFile = value
		case "MANAGER_INSTRUCTION_FILE":
			config.ManagerInstructionFile = value
		case "DEV_INSTRUCTION_FILE":
			config.DevInstructionFile = value
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
			CloudCodeAgentsDir: ".claude-code-agents",
			TeamConfigPath:     ".claude-code-agents.conf",
			MainConfigPath:     "manager.json",
			LogsDir:            "logs",
			InstructionsDir:    "instructions",
			AuthBackupDir:      "auth_backup",
		}
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claude-code-agents")

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
		if err := os.MkdirAll(dir, 0750); err != nil {
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

	return &UnifiedConfig{
		Paths: paths,
		Team:  teamConfig,
		Main:  mainConfig,
	}, nil
}

// UnifiedConfig 統一された設定構造体
type UnifiedConfig struct {
	Paths *ConfigPaths
	Team  *TeamConfig
	Main  *Config
}

// GetEffectiveConfig 有効な設定値を取得（優先順位: Team > Main > Common）
func (uc *UnifiedConfig) GetEffectiveConfig() *EffectiveConfig {
	return &EffectiveConfig{
		MaxProcesses:           uc.Team.MaxProcesses,
		MaxMemoryMB:            uc.Team.MaxMemoryMB,
		MaxCPUPercent:          uc.Team.MaxCPUPercent,
		LogLevel:               uc.Team.LogLevel,
		ClaudeCLIPath:          uc.Team.ClaudeCLIPath,
		InstructionsDir:        uc.Team.InstructionsDir,
		WorkingDir:             uc.Team.WorkingDir,
		ConfigDir:              uc.Team.ConfigDir,
		LogFile:                uc.Team.LogFile,
		AuthBackupDir:          uc.Team.AuthBackupDir,
		StartupTimeout:         uc.Team.StartupTimeout,
		ShutdownTimeout:        uc.Team.ShutdownTimeout,
		ProcessTimeout:         uc.Team.ProcessTimeout,
		RestartDelay:           uc.Team.RestartDelay,
		HealthCheckInterval:    uc.Team.HealthCheckInterval,
		AuthCheckInterval:      uc.Team.AuthCheckInterval,
		MaxRestartAttempts:     uc.Team.MaxRestartAttempts,
		SessionName:            uc.Team.SessionName,
		DefaultLayout:          uc.Team.DefaultLayout,
		AutoAttach:             uc.Team.AutoAttach,
		PaneCount:              uc.Team.PaneCount,
		IDEBackupEnabled:       uc.Team.IDEBackupEnabled,
		SendCommand:            uc.Team.SendCommand,
		BinaryName:             uc.Team.BinaryName,
		DevCount:               uc.Team.DevCount,
		POInstructionFile:      uc.Team.POInstructionFile,
		ManagerInstructionFile: uc.Team.ManagerInstructionFile,
		DevInstructionFile:     uc.Team.DevInstructionFile,
	}
}

// EffectiveConfig 有効な設定値
type EffectiveConfig struct {
	MaxProcesses           int
	MaxMemoryMB            int64
	MaxCPUPercent          float64
	LogLevel               string
	ClaudeCLIPath          string
	InstructionsDir        string
	WorkingDir             string
	ConfigDir              string
	LogFile                string
	AuthBackupDir          string
	StartupTimeout         time.Duration
	ShutdownTimeout        time.Duration
	ProcessTimeout         time.Duration
	RestartDelay           time.Duration
	HealthCheckInterval    time.Duration
	AuthCheckInterval      time.Duration
	MaxRestartAttempts     int
	SessionName            string
	DefaultLayout          string
	AutoAttach             bool
	PaneCount              int
	IDEBackupEnabled       bool
	SendCommand            string
	BinaryName             string
	DevCount               int
	POInstructionFile      string
	ManagerInstructionFile string
	DevInstructionFile     string
}

// GetTeamConfigPath 設定ファイルのパスを取得
func GetTeamConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// エラーの場合はカレントディレクトリのデフォルトファイルを返す
		return ".claude-code-agents.conf"
	}

	// 統一された設定ディレクトリパス
	configDir := filepath.Join(homeDir, ".claude", "claude-code-agents")

	// 優先順位1: claude-code-agentsディレクトリ内の設定ファイル
	claudConfigPath := filepath.Join(configDir, "agents.conf")
	if _, err := os.Stat(claudConfigPath); err == nil {
		return claudConfigPath
	}

	// 優先順位2: ホームディレクトリの設定ファイル
	homeConfigPath := filepath.Join(homeDir, ".claude-code-agents.conf")
	if _, err := os.Stat(homeConfigPath); err == nil {
		return homeConfigPath
	}

	// 優先順位3: カレントディレクトリの設定ファイル
	currentConfigPath := ".claude-code-agents.conf"
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

// GetDefaultTeamConfigPath デフォルトのチーム設定ファイルパスを取得
func GetDefaultTeamConfigPath() string {
	return GetTeamConfigPath()
}

// TeamConfigLoader チーム設定ローダー
type TeamConfigLoader struct {
	configPath          string
	instructionResolver InstructionResolverInterface
	validator           InstructionValidatorInterface
}

// NewTeamConfigLoader 新しいチーム設定ローダーを作成
func NewTeamConfigLoader(configPath string) *TeamConfigLoader {
	return &TeamConfigLoader{
		configPath: configPath,
	}
}

// LoadTeamConfig チーム設定を読み込み（拡張版）
func (tcl *TeamConfigLoader) LoadTeamConfig() (*TeamConfig, error) {
	// 基本設定の読み込み
	config, err := LoadTeamConfigFromPath(tcl.configPath)
	if err != nil {
		return nil, err
	}

	// instruction解決器とバリデーターを初期化
	tcl.instructionResolver = NewInstructionResolver(config)
	tcl.validator = NewInstructionValidator(config.StrictValidation)

	// 必要に応じて設定の正規化
	tcl.normalizeInstructionConfig(config)

	return config, nil
}

// SaveTeamConfig チーム設定を保存
func (tcl *TeamConfigLoader) SaveTeamConfig(config *TeamConfig) error {
	// ディレクトリ作成
	if err := os.MkdirAll(filepath.Dir(tcl.configPath), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 設定ファイルの内容を生成
	content := fmt.Sprintf(`# AI Team Configuration File
# Generated by Claude Code Agents

# Path Configurations
CLAUDE_CLI_PATH=%s
INSTRUCTIONS_DIR=%s
CONFIG_DIR=%s
LOG_FILE=%s
AUTH_BACKUP_DIR=%s

# System Settings
LOG_LEVEL=%s

# Tmux Settings
SESSION_NAME=%s
DEFAULT_LAYOUT=%s
AUTO_ATTACH=%t
IDE_BACKUP_ENABLED=%t

# Commands
SEND_COMMAND=%s
BINARY_NAME=%s

# Developer Settings
DEV_COUNT=%d

# Role-based Instructions
PO_INSTRUCTION_FILE=%s
MANAGER_INSTRUCTION_FILE=%s
DEV_INSTRUCTION_FILE=%s

# Timeout Settings
HEALTH_CHECK_INTERVAL=%s
AUTH_CHECK_INTERVAL=%s
STARTUP_TIMEOUT=%s
SHUTDOWN_TIMEOUT=%s
RESTART_DELAY=%s
PROCESS_TIMEOUT=%s
`,
		config.ClaudeCLIPath,
		config.InstructionsDir,
		config.ConfigDir,
		config.LogFile,
		config.AuthBackupDir,
		config.LogLevel,
		config.SessionName,
		config.DefaultLayout,
		config.AutoAttach,
		config.IDEBackupEnabled,
		config.SendCommand,
		config.BinaryName,
		config.DevCount,
		config.POInstructionFile,
		config.ManagerInstructionFile,
		config.DevInstructionFile,
		config.HealthCheckInterval.String(),
		config.AuthCheckInterval.String(),
		config.StartupTimeout.String(),
		config.ShutdownTimeout.String(),
		config.RestartDelay.String(),
		config.ProcessTimeout.String(),
	)

	return os.WriteFile(tcl.configPath, []byte(content), 0600)
}

// GetDevCount 開発者数を取得
func (tc *TeamConfig) GetDevCount() int {
	return tc.DevCount
}

// SetDevCount 開発者数を設定
func (tc *TeamConfig) SetDevCount(count int) {
	if count > 0 {
		tc.DevCount = count
		// PaneCountも更新（PO + Manager + Dev数）
		tc.PaneCount = 2 + count
	}
}

// GetAgentList 動的エージェントリストを取得
func (tc *TeamConfig) GetAgentList() []string {
	agents := []string{"po", "manager"}
	for i := 1; i <= tc.DevCount; i++ {
		agents = append(agents, fmt.Sprintf("dev%d", i))
	}
	return agents
}

// GetPaneAgentMap 動的ペイン-エージェントマップを取得
func (tc *TeamConfig) GetPaneAgentMap() map[string]string {
	paneMap := make(map[string]string)
	paneMap["1"] = "po"
	paneMap["2"] = "manager"
	for i := 1; i <= tc.DevCount; i++ {
		paneMap[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("dev%d", i)
	}
	return paneMap
}

// GetPaneTitles 動的ペインタイトルマップを取得
func (tc *TeamConfig) GetPaneTitles() map[string]string {
	titles := make(map[string]string)
	titles["1"] = "PO"
	titles["2"] = "Manager"
	for i := 1; i <= tc.DevCount; i++ {
		titles[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("Dev%d", i)
	}
	return titles
}

// === 新規追加メソッド（dynamic instruction機能） ===

// GetInstructionResolver instruction解決器を取得
func (tcl *TeamConfigLoader) GetInstructionResolver() InstructionResolverInterface {
	return tcl.instructionResolver
}

// ResolveInstructionPath instructionファイルパスを解決
func (tcl *TeamConfigLoader) ResolveInstructionPath(role string) (string, error) {
	if tcl.instructionResolver == nil {
		return "", fmt.Errorf("instruction resolver not initialized")
	}
	return tcl.instructionResolver.ResolveInstructionPath(role)
}

// ValidateInstructionConfig instruction設定を検証
func (tcl *TeamConfigLoader) ValidateInstructionConfig() *ValidationResult {
	if tcl.validator == nil {
		return &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{{
				Message: "validator not initialized",
				Code:    "VALIDATOR_NOT_INITIALIZED",
			}},
		}
	}

	config, err := tcl.LoadTeamConfig()
	if err != nil {
		return &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{{
				Message: err.Error(),
				Code:    "CONFIG_LOAD_FAILED",
			}},
		}
	}

	return tcl.validator.ValidateConfig(config)
}

// normalizeInstructionConfig 設定の正規化
func (tcl *TeamConfigLoader) normalizeInstructionConfig(config *TeamConfig) {
	// 既存設定から拡張設定への移行
	if config.InstructionConfig == nil &&
		(config.POInstructionFile != "" ||
			config.ManagerInstructionFile != "" ||
			config.DevInstructionFile != "") {

		// 基本設定として既存値を設定
		config.InstructionConfig = &InstructionConfig{
			Base: InstructionRoleConfig{
				POInstructionPath:      config.POInstructionFile,
				ManagerInstructionPath: config.ManagerInstructionFile,
				DevInstructionPath:     config.DevInstructionFile,
			},
			Global: InstructionGlobalConfig{
				DefaultExtension: ".md",
				CacheEnabled:     true,
				CacheTTL:         5 * time.Minute,
			},
		}
	}

	// フォールバックディレクトリの設定
	if config.FallbackInstructionDir == "" {
		config.FallbackInstructionDir = config.InstructionsDir
	}
}

// GetTeamConfigPath 設定ファイルパスを取得
func (tcl *TeamConfigLoader) GetTeamConfigPath() string {
	return tcl.configPath
}
