package cmd

// Common functions and data structure definitions

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

// Common configuration structure
type CommonConfig struct {
	HomeDir         string
	ConfigDir       string
	WorkingDir      string
	ClaudeCLIPath   string
	InstructionsDir string
	LogLevel        string
	Verbose         bool
}

// Global configuration instance
var globalConfig *CommonConfig

// Global configuration variables
var (
	globalConfigDir string
	globalLogLevel  = "error"
	globalVerbose   = false
)

// AgentConfig agent configuration
type AgentConfig struct {
	Name            string
	InstructionFile string
	WorkingDir      string
}

// GetCommonConfig get common configuration
func GetCommonConfig() *CommonConfig {
	if globalConfig == nil {
		globalConfig = &CommonConfig{}
		if err := globalConfig.Initialize(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize common config")
		}
	}
	return globalConfig
}

// Initialize configuration initialization
func (c *CommonConfig) Initialize() error {
	// Set home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	c.HomeDir = homeDir

	// Unified configuration directory path
	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claude-code-agents")

	// Set configuration directory
	if globalConfigDir != "" {
		c.ConfigDir = globalConfigDir
	} else {
		c.ConfigDir = claudCodeAgentsDir
	}

	// Set working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	c.WorkingDir = workingDir

	// Set Claude CLI path
	c.ClaudeCLIPath = filepath.Join(claudeDir, "local", "claude")

	// Set instructions directory
	c.InstructionsDir = filepath.Join(claudCodeAgentsDir, "instructions")

	// Set log level
	c.LogLevel = globalLogLevel
	c.Verbose = globalVerbose

	return nil
}

// GetConfigPath get configuration file path
func (c *CommonConfig) GetConfigPath() string {
	return filepath.Join(c.ConfigDir, "manager.json")
}

// GetTeamConfigPath get team configuration file path
func (c *CommonConfig) GetTeamConfigPath() string {
	return filepath.Join(c.ConfigDir, "agents.conf")
}

// GetSessionName get tmux session name (dynamic detection)
func (c *CommonConfig) GetSessionName() string {
	// Detect active AI session using tmuxManager
	tmuxManager := tmux.NewTmuxManager("")
	if sessionName, err := tmuxManager.FindDefaultAISession(6); err == nil {
		return sessionName
	}

	// Return default value if detection fails
	return "ai-teams"
}

// GetLogPath get log file path
func (c *CommonConfig) GetLogPath() string {
	return filepath.Join(c.ConfigDir, "logs", "manager.log")
}

// Common error handling

// CommandError command error structure
type CommandError struct {
	Command string
	Err     error
	Code    int
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command '%s' failed: %v (exit code: %d)", e.Command, e.Err, e.Code)
}

// WrapError wrap error
func WrapError(command string, err error, code int) *CommandError {
	return &CommandError{
		Command: command,
		Err:     err,
		Code:    code,
	}
}

// Common validation functions

// ValidateAgentName validate agent name
func ValidateAgentName(agentName string) error {
	validAgents := map[string]bool{
		"po":      true,
		"manager": true,
		"dev1":    true,
		"dev2":    true,
		"dev3":    true,
		"dev4":    true,
	}

	if !validAgents[agentName] {
		return fmt.Errorf("invalid agent name '%s'. Valid agents: po, manager, dev1, dev2, dev3, dev4", agentName)
	}

	return nil
}

// ValidateMessage validate message
func ValidateMessage(message string) error {
	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	if len(message) > 4096 {
		return fmt.Errorf("message too long (max 4096 characters)")
	}

	return nil
}

// ValidateConfig validate configuration
func ValidateConfig() error {
	config := GetCommonConfig()

	// Check Claude CLI existence
	if _, err := os.Stat(config.ClaudeCLIPath); os.IsNotExist(err) {
		return fmt.Errorf("claude CLI not found at %s", config.ClaudeCLIPath)
	}

	// Check instructions directory existence
	if _, err := os.Stat(config.InstructionsDir); os.IsNotExist(err) {
		return fmt.Errorf("instructions directory not found at %s", config.InstructionsDir)
	}

	return nil
}

// Common helper functions

// EnsureDir ensure directory exists and create if necessary
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0750)
	}
	return nil
}

// IsProcessRunning check if process is running
func IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to process to check existence (Unix only)
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// FormatDuration format duration
func FormatDuration(duration time.Duration) string {
	switch {
	case duration < time.Millisecond:
		return fmt.Sprintf("%.2fμs", float64(duration.Nanoseconds())/1000)
	case duration < time.Second:
		return fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1e6)
	default:
		return fmt.Sprintf("%.2fs", duration.Seconds())
	}
}

// TruncateString truncate string
func TruncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}

	if maxLen <= 3 {
		return str[:maxLen]
	}

	return str[:maxLen-3] + "..."
}

// Common status management

// SystemStatus system status
type SystemStatus struct {
	IsRunning     bool            `json:"is_running"`
	StartTime     time.Time       `json:"start_time"`
	Uptime        time.Duration   `json:"uptime"`
	AgentStatuses map[string]bool `json:"agent_statuses"`
	LastUpdate    time.Time       `json:"last_update"`
	Version       string          `json:"version"`
	PID           int             `json:"pid"`
}

// NewSystemStatus create new system status
func NewSystemStatus() *SystemStatus {
	return &SystemStatus{
		IsRunning:     false,
		AgentStatuses: make(map[string]bool),
		LastUpdate:    time.Now(),
		Version:       "1.0.0",
		PID:           os.Getpid(),
	}
}

// UpdateAgentStatus update agent status
func (s *SystemStatus) UpdateAgentStatus(agentName string, isRunning bool) {
	s.AgentStatuses[agentName] = isRunning
	s.LastUpdate = time.Now()
}

// GetAgentStatus get agent status
func (s *SystemStatus) GetAgentStatus(agentName string) bool {
	return s.AgentStatuses[agentName]
}

// GetAllAgentStatuses get all agent statuses
func (s *SystemStatus) GetAllAgentStatuses() map[string]bool {
	statuses := make(map[string]bool)
	for agent, status := range s.AgentStatuses {
		statuses[agent] = status
	}

	return statuses
}

// Start system start
func (s *SystemStatus) Start() {
	s.IsRunning = true
	s.StartTime = time.Now()
	s.LastUpdate = time.Now()
}

// Stop system stop
func (s *SystemStatus) Stop() {
	s.IsRunning = false
	s.LastUpdate = time.Now()
}

// GetUptime get uptime
func (s *SystemStatus) GetUptime() time.Duration {
	if !s.IsRunning {
		return 0
	}

	return time.Since(s.StartTime)
}

// Common message format

// FormatAgentStatus format agent status
func FormatAgentStatus(agentName string, isRunning bool) string {
	statusIcon := "❌"
	statusText := "Stopped"

	if isRunning {
		statusIcon = "✅"
		statusText = "Running"
	}

	return fmt.Sprintf("%s %s: %s", statusIcon, agentName, statusText)
}

// FormatMessage format message
func FormatMessage(sender, recipient, message string) string {
	timestamp := time.Now().Format("15:04:05")
	return fmt.Sprintf("[%s] %s -> %s: %s", timestamp, sender, recipient, message)
}

// FormatError format error
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	return fmt.Sprintf("❌ Error: %v", err)
}

// FormatSuccess format success message
func FormatSuccess(message string) string {
	return fmt.Sprintf("✅ %s", message)
}

// FormatWarning format warning message
func FormatWarning(message string) string {
	return fmt.Sprintf("⚠️ %s", message)
}

// FormatInfo format information message
func FormatInfo(message string) string {
	return fmt.Sprintf("ℹ️ %s", message)
}

// Common resource management

// ResourceManager resource manager
type ResourceManager struct {
	cleanupFuncs []func() error
}

// NewResourceManager create new resource manager
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		cleanupFuncs: make([]func() error, 0),
	}
}

// AddCleanup add cleanup function
func (rm *ResourceManager) AddCleanup(cleanup func() error) {
	rm.cleanupFuncs = append(rm.cleanupFuncs, cleanup)
}

// Cleanup cleanup all resources
func (rm *ResourceManager) Cleanup() error {
	var errors []error

	// Execute cleanup in reverse order
	for i := len(rm.cleanupFuncs) - 1; i >= 0; i-- {
		if err := rm.cleanupFuncs[i](); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// Global resource manager instance
var globalResourceManager *ResourceManager

// GetResourceManager get global resource manager
func GetResourceManager() *ResourceManager {
	if globalResourceManager == nil {
		globalResourceManager = NewResourceManager()
	}
	return globalResourceManager
}
