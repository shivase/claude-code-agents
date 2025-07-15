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

// TeamConfig represents AI Team configuration structure
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

	// Developer settings
	DevCount int

	// Role-based Instructions
	POInstructionFile      string
	ManagerInstructionFile string
	DevInstructionFile     string

	// === New fields ===
	// Extended instruction configuration
	InstructionConfig *InstructionConfig `json:"instruction_config,omitempty"`

	// Fallback settings
	FallbackInstructionDir string `json:"fallback_instruction_dir,omitempty"`

	// Environment settings
	Environment string `json:"environment,omitempty"` // development, production, etc.

	// Validation settings
	StrictValidation bool `json:"strict_validation,omitempty"`
}

// InstructionConfig represents extended instruction configuration
type InstructionConfig struct {
	// Base configuration
	Base InstructionRoleConfig `json:"base"`

	// Environment-specific configuration
	Environments map[string]InstructionRoleConfig `json:"environments,omitempty"`

	// Global configuration
	Global InstructionGlobalConfig `json:"global,omitempty"`
}

// InstructionRoleConfig represents role-specific instruction configuration
type InstructionRoleConfig struct {
	POInstructionPath      string `json:"po_instruction_path,omitempty"`
	ManagerInstructionPath string `json:"manager_instruction_path,omitempty"`
	DevInstructionPath     string `json:"dev_instruction_path,omitempty"`
}

// InstructionGlobalConfig represents global instruction configuration
type InstructionGlobalConfig struct {
	DefaultExtension string        `json:"default_extension,omitempty"` // .md, .txt
	SearchPaths      []string      `json:"search_paths,omitempty"`
	CacheEnabled     bool          `json:"cache_enabled,omitempty"`
	CacheTTL         time.Duration `json:"cache_ttl,omitempty"`
}

// GetWorkingDir implements ConfigInterface method
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

// LoadTeamConfigFromPath loads configuration from file path
func LoadTeamConfigFromPath(configPath string) (*TeamConfig, error) {
	homeDir, _ := os.UserHomeDir()

	// Get optimal working directory using directory resolver
	resolver := utils.GetGlobalDirectoryResolver()
	optimalWorkingDir := resolver.GetOptimalWorkingDirectory()

	// Unified configuration directory path
	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claude-code-agents")

	// Default configuration
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

	// Return default configuration if config file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	// Load config file - path normalization and directory traversal prevention
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
			// Log or handle appropriately
			_, err := fmt.Fprintf(os.Stderr, "Warning: failed to close config file: %v\n", err)
			if err != nil {
				return
			}
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Apply configuration values
		switch key {
		case "CLAUDE_CLI_PATH":
			config.ClaudeCLIPath = value
		case "INSTRUCTIONS_DIR":
			config.InstructionsDir = value
		case "WORKING_DIR":
			// Skip WorkingDir as it's optimized by directory resolver
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

	// Fix directory dependent paths after loading config file
	if resolveErr := resolver.FixDirectoryDependentPaths(config); resolveErr != nil {
		log.Warn().Err(resolveErr).Msg("Failed to fix directory dependent paths")
	}

	return config, nil
}

// GetUnifiedConfigPaths gets unified configuration paths
func GetUnifiedConfigPaths() *ConfigPaths {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Return relative paths on error
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

// ConfigPaths represents configuration path structure
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

// EnsureDirectories creates necessary directories
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

// LoadUnifiedConfig loads unified configuration
func LoadUnifiedConfig() (*UnifiedConfig, error) {
	paths := GetUnifiedConfigPaths()

	// Create necessary directories
	if err := paths.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Load TeamConfig
	teamConfig, err := LoadTeamConfigFromPath(paths.TeamConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load team config: %w", err)
	}

	// Load MainConfig
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

// UnifiedConfig represents unified configuration structure
type UnifiedConfig struct {
	Paths *ConfigPaths
	Team  *TeamConfig
	Main  *Config
}

// GetEffectiveConfig gets effective configuration values (Priority: Team > Main > Common)
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

// EffectiveConfig represents effective configuration values
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

// GetTeamConfigPath gets configuration file path
func GetTeamConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Return default file in current directory on error
		return ".claude-code-agents.conf"
	}

	// Unified configuration directory path
	configDir := filepath.Join(homeDir, ".claude", "claude-code-agents")

	// Priority 1: Configuration file in claude-code-agents directory
	claudConfigPath := filepath.Join(configDir, "agents.conf")
	if _, err := os.Stat(claudConfigPath); err == nil {
		return claudConfigPath
	}

	// Priority 2: Configuration file in home directory
	homeConfigPath := filepath.Join(homeDir, ".claude-code-agents.conf")
	if _, err := os.Stat(homeConfigPath); err == nil {
		return homeConfigPath
	}

	// Priority 3: Configuration file in current directory
	currentConfigPath := ".claude-code-agents.conf"
	if _, err := os.Stat(currentConfigPath); err == nil {
		return currentConfigPath
	}

	// Default path (in unified directory)
	return claudConfigPath
}

// LoadTeamConfig loads configuration file (no parameter version)
func LoadTeamConfig() (*TeamConfig, error) {
	configPath := GetTeamConfigPath()
	return LoadTeamConfigFromPath(configPath)
}

// GetDefaultTeamConfigPath gets default team configuration file path
func GetDefaultTeamConfigPath() string {
	return GetTeamConfigPath()
}

// TeamConfigLoader loads team configuration
type TeamConfigLoader struct {
	configPath          string
	instructionResolver InstructionResolverInterface
	validator           InstructionValidatorInterface
}

// NewTeamConfigLoader creates a new team configuration loader
func NewTeamConfigLoader(configPath string) *TeamConfigLoader {
	return &TeamConfigLoader{
		configPath: configPath,
	}
}

// LoadTeamConfig loads team configuration (extended version)
func (tcl *TeamConfigLoader) LoadTeamConfig() (*TeamConfig, error) {
	// Base configurationの読み込み
	config, err := LoadTeamConfigFromPath(tcl.configPath)
	if err != nil {
		return nil, err
	}

	// Initialize instruction resolver and validator
	tcl.instructionResolver = NewInstructionResolver(config)
	tcl.validator = NewInstructionValidator(config.StrictValidation)

	// Normalize configuration if necessary
	tcl.normalizeInstructionConfig(config)

	return config, nil
}

// SaveTeamConfig saves team configuration
func (tcl *TeamConfigLoader) SaveTeamConfig(config *TeamConfig) error {
	// Create directory
	if err := os.MkdirAll(filepath.Dir(tcl.configPath), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate configuration file content
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

// GetDevCount gets developer count
func (tc *TeamConfig) GetDevCount() int {
	return tc.DevCount
}

// SetDevCount sets developer count
func (tc *TeamConfig) SetDevCount(count int) {
	if count > 0 {
		tc.DevCount = count
		// Update PaneCount (PO + Manager + Dev count)
		tc.PaneCount = 2 + count
	}
}

// GetAgentList gets dynamic agent list
func (tc *TeamConfig) GetAgentList() []string {
	agents := []string{"po", "manager"}
	for i := 1; i <= tc.DevCount; i++ {
		agents = append(agents, fmt.Sprintf("dev%d", i))
	}
	return agents
}

// GetPaneAgentMap gets dynamic pane-agent map
func (tc *TeamConfig) GetPaneAgentMap() map[string]string {
	paneMap := make(map[string]string)
	paneMap["1"] = "po"
	paneMap["2"] = "manager"
	for i := 1; i <= tc.DevCount; i++ {
		paneMap[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("dev%d", i)
	}
	return paneMap
}

// GetPaneTitles gets dynamic pane title map
func (tc *TeamConfig) GetPaneTitles() map[string]string {
	titles := make(map[string]string)
	titles["1"] = "PO"
	titles["2"] = "Manager"
	for i := 1; i <= tc.DevCount; i++ {
		titles[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("Dev%d", i)
	}
	return titles
}

// === New methods (dynamic instruction feature) ===

// GetInstructionResolver gets instruction resolver
func (tcl *TeamConfigLoader) GetInstructionResolver() InstructionResolverInterface {
	return tcl.instructionResolver
}

// ResolveInstructionPath resolves instruction file path
func (tcl *TeamConfigLoader) ResolveInstructionPath(role string) (string, error) {
	if tcl.instructionResolver == nil {
		return "", fmt.Errorf("instruction resolver not initialized")
	}
	return tcl.instructionResolver.ResolveInstructionPath(role)
}

// ValidateInstructionConfig validates instruction configuration
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

// normalizeInstructionConfig normalizes configuration
func (tcl *TeamConfigLoader) normalizeInstructionConfig(config *TeamConfig) {
	// Migrate from existing to extended configuration
	if config.InstructionConfig == nil &&
		(config.POInstructionFile != "" ||
			config.ManagerInstructionFile != "" ||
			config.DevInstructionFile != "") {

		// Set existing values as base configuration
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

	// Set fallback directory
	if config.FallbackInstructionDir == "" {
		config.FallbackInstructionDir = config.InstructionsDir
	}
}

// GetTeamConfigPath 設定ファイルパスを取得
func (tcl *TeamConfigLoader) GetTeamConfigPath() string {
	return tcl.configPath
}
