package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/shivase/claude-code-agents/internal/logger"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

//go:embed instructions
var InstructionsFS embed.FS

var (
	// GlobalLogLevel global configuration flag
	GlobalLogLevel string

	// MainInitialized initialization management
	MainInitialized   bool
	LoggerInitialized bool
)

// InitializeMainSystem main system initialization process integration
func InitializeMainSystem(logLevel string) {
	// Skip if already initialized
	if MainInitialized {
		return
	}

	// Initialize log system
	GlobalLogLevel = logLevel
	InitLogger()

	// Set initialization flag
	MainInitialized = true
}

// InitLogger log system initialization
func InitLogger() {
	if LoggerInitialized {
		return
	}

	_, err := zerolog.ParseLevel(GlobalLogLevel)
	if err != nil {
		// fallback to error level if parsing fails
		GlobalLogLevel = "error"
	}

	logger.InitConsoleLogger(GlobalLogLevel)

	log.Debug().Msg("Commands package initialized: Command functions are available")

	LoggerInitialized = true
}

// IsValidSessionName session name validation
func IsValidSessionName(name string) bool {
	if name == "" {
		return false
	}

	for _, char := range name {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '-' && char != '_' {
			return false
		}
	}
	return true
}

// ListAISessions session list display function
func ListAISessions() error {
	fmt.Println("🤖 Session List")
	fmt.Println("==================================")

	tmuxManager := tmux.NewTmuxManager("ai-teams")
	sessions, err := tmuxManager.ListSessions()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("📭 No AI team sessions currently running")
			return nil
		}
		return fmt.Errorf("tmux session retrieval error: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("📭 No sessions currently running")
	} else {
		fmt.Printf("🚀 Running sessions: %d\n", len(sessions))
		for i, session := range sessions {
			fmt.Printf("  %d. %s\n", i+1, session)
		}
	}

	return nil
}

// DeleteAISession delete specified session
func DeleteAISession(sessionName string) error {
	if sessionName == "" {
		fmt.Println("❌ Error: Please specify the session name to delete")
		fmt.Println("Usage: ./claude-code-agents --delete [session-name]")
		fmt.Println("Session list: ./claude-code-agents --list")
		return fmt.Errorf("session name not specified")
	}

	fmt.Printf("🗑️ Deleting session: %s\n", sessionName)

	tmuxManager := tmux.NewTmuxManager(sessionName)
	if !tmuxManager.SessionExists(sessionName) {
		fmt.Printf("⚠️ Session '%s' does not exist\n", sessionName)
		return nil
	}

	if err := tmuxManager.KillSession(sessionName); err != nil {
		return fmt.Errorf("session deletion error: %w", err)
	}

	fmt.Printf("✅ Session '%s' deleted\n", sessionName)
	return nil
}

// DeleteAllAISessions delete all sessions
func DeleteAllAISessions() error {
	fmt.Println("🗑️ Deleting All AI Team Sessions")
	fmt.Println("==============================")

	tmuxManager := tmux.NewTmuxManager("ai-teams")
	sessions, err := tmuxManager.ListSessions()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("📭 No AI team sessions currently running")
			return nil
		}
		return fmt.Errorf("tmux session retrieval error: %w", err)
	}

	var aiSessions []string

	// Extract AI team related sessions
	for _, session := range sessions {
		if strings.Contains(session, "ai-") || strings.Contains(session, "claude-") ||
			strings.Contains(session, "dev-") || strings.Contains(session, "agent-") {
			aiSessions = append(aiSessions, session)
		}
	}

	if len(aiSessions) == 0 {
		fmt.Println("📭 No AI team sessions to delete")
		return nil
	}

	fmt.Printf("🎯 Sessions to delete: %d\n", len(aiSessions))
	for i, session := range aiSessions {
		fmt.Printf("  %d. %s\n", i+1, session)
	}

	// Delete each session
	deletedCount := 0
	for _, session := range aiSessions {
		sessionManager := tmux.NewTmuxManager(session)
		if err := sessionManager.KillSession(session); err != nil {
			fmt.Printf("⚠️ Failed to delete session '%s': %v\n", session, err)
		} else {
			deletedCount++
			fmt.Printf("✅ Session '%s' deleted\n", session)
		}
	}

	fmt.Printf("\n🎉 Deleted %d AI team sessions\n", deletedCount)
	return nil
}

// LaunchSystem system launch function
func LaunchSystem(sessionName string) error {
	fmt.Printf("🚀 System startup: %s\n", sessionName)

	// Load configuration file
	configPath := config.GetDefaultTeamConfigPath()
	configLoader := config.NewTeamConfigLoader(configPath)
	teamConfig, err := configLoader.LoadTeamConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration file: %w", err)
	}

	// Configuration load completion log
	if GlobalLogLevel == "debug" || GlobalLogLevel == "info" {
		logger.LogConfigLoad(configPath, map[string]interface{}{
			"config_path":      configPath,
			"dev_count":        teamConfig.DevCount,
			"session_name":     teamConfig.SessionName,
			"instructions_dir": teamConfig.InstructionsDir,
		})
	}

	// Collect and display instruction file information
	instructionInfo := gatherInstructionInfo(teamConfig)
	envInfo := gatherEnvironmentInfo(teamConfig)

	// Instruction configuration information log
	if GlobalLogLevel == "debug" || GlobalLogLevel == "info" {
		logger.LogInstructionConfig(instructionInfo, map[string]interface{}{
			"config_loaded": true,
			"role_count":    len(instructionInfo),
		})
	}

	// Environment information log (get debug mode from global)
	debugMode := GlobalLogLevel == "debug"
	if debugMode || GlobalLogLevel == "info" {
		logger.LogEnvironmentInfo(envInfo, debugMode)
	}

	// Basic tmux management operations
	tmuxManager := tmux.NewTmuxManager(sessionName)

	// Check existing session
	if tmuxManager.SessionExists(sessionName) {
		fmt.Printf("🔄 Connecting to existing session '%s'\n", sessionName)
		return tmuxManager.AttachSession(sessionName)
	}

	// Create new session
	fmt.Printf("📝 Creating new session '%s'\n", sessionName)
	if err := tmuxManager.CreateSession(sessionName); err != nil {
		return fmt.Errorf("session creation failed: %w", err)
	}

	// Create integrated layout (dynamic dev count support)
	fmt.Println("🎛️ Creating integrated layout...")
	if err := tmuxManager.CreateIntegratedLayout(sessionName, teamConfig.DevCount); err != nil {
		return fmt.Errorf("integrated layout creation failed: %w", err)
	}

	// Claude CLI automatic startup process (configuration file support)
	fmt.Println("🤖 Starting Claude CLI in each pane...")
	if err := tmuxManager.SetupClaudeInPanesWithConfig(sessionName, teamConfig.ClaudeCLIPath, teamConfig.InstructionsDir, teamConfig, teamConfig.DevCount); err != nil {
		fmt.Printf("⚠️ Claude CLI automatic startup failed: %v\n", err)
		fmt.Printf("Please start Claude CLI manually: %s --dangerously-skip-permissions\n", teamConfig.ClaudeCLIPath)
		// Fallback: try conventional method
		fmt.Println("🔄 Fallback: retrying with conventional method...")
		if err := tmuxManager.SetupClaudeInPanes(sessionName, teamConfig.ClaudeCLIPath, teamConfig.InstructionsDir, teamConfig.DevCount); err != nil {
			fmt.Printf("⚠️ Fallback startup also failed: %v\n", err)
		} else {
			fmt.Println("✅ Fallback startup successful")
		}
	} else {
		fmt.Println("✅ Claude CLI automatic startup completed (configuration file support)")
	}

	fmt.Printf("✅ Session '%s' preparation completed\n", sessionName)
	return tmuxManager.AttachSession(sessionName)
}

// InitializeSystemCommand system initialization command
func InitializeSystemCommand(forceOverwrite bool, language string) error {
	fmt.Println("🚀 Claude Code Agents System Initialization")
	fmt.Println("=====================================")

	if forceOverwrite {
		fmt.Println("⚠️ Force overwrite mode is enabled")
	}

	// Directory creation process
	if err := createSystemDirectories(forceOverwrite); err != nil {
		return fmt.Errorf("directory creation failed: %w", err)
	}

	// Configuration file generation process
	if err := generateInitialConfig(forceOverwrite); err != nil {
		return fmt.Errorf("configuration file generation failed: %w", err)
	}

	// Copy instruction files based on language
	if err := CopyInstructionFiles(language, forceOverwrite); err != nil {
		return fmt.Errorf("instruction files copy failed: %w", err)
	}

	// Display success message
	displayInitializationSuccess()

	return nil
}

// createSystemDirectories system directory creation
func createSystemDirectories(forceOverwrite bool) error {
	fmt.Println("📁 Creating directory structure...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// List of directories to create
	directories := []struct {
		path        string
		description string
	}{
		{filepath.Join(homeDir, ".claude"), "Claude base directory"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents"), "Claude Code Agents directory"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"), "Instructions directory"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents", "auth_backup"), "Authentication backup directory"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents", "logs"), "Log directory"},
	}

	for _, dir := range directories {
		fmt.Printf("  📂 %s: %s\n", dir.description, dir.path)

		// Check if directory already exists
		if _, err := os.Stat(dir.path); err == nil {
			if !forceOverwrite {
				fmt.Printf("     ✅ Already exists (skipped)\n")
				continue
			}
			fmt.Printf("     ⚠️ Already exists but continuing (force mode)\n")
		}

		// Create directory
		if err := os.MkdirAll(dir.path, 0750); err != nil {
			return fmt.Errorf("directory creation failed %s: %w", dir.path, err)
		}
		fmt.Printf("     ✅ Creation completed\n")
	}

	return nil
}

// CopyInstructionFiles copy instruction files based on language
func CopyInstructionFiles(language string, forceOverwrite bool) error {
	fmt.Printf("📋 Copying instruction files for language: %s\n", language)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	targetDir := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions")

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// List of instruction files to copy
	instructionFiles := []string{"po.md", "manager.md", "developer.md"}

	for _, filename := range instructionFiles {
		sourcePath := filepath.Join("instructions", language, filename)
		targetFile := filepath.Join(targetDir, filename)

		fmt.Printf("  📄 %s: embedded %s -> %s\n", filename, sourcePath, targetFile)

		// Read embedded file
		data, err := InstructionsFS.ReadFile(sourcePath)
		if err != nil {
			fmt.Printf("     ⚠️ Embedded file not found (skipped)\n")
			continue
		}

		// Check if target file already exists
		if _, err := os.Stat(targetFile); err == nil {
			if !forceOverwrite {
				fmt.Printf("     ✅ Already exists (skipped)\n")
				continue
			}
			fmt.Printf("     ⚠️ Already exists but overwriting (force mode)\n")
		}

		// Write file
		if err := os.WriteFile(targetFile, data, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
		fmt.Printf("     ✅ Copy completed\n")
	}

	return nil
}

// generateInitialConfig initial configuration file generation
func generateInitialConfig(forceOverwrite bool) error {
	fmt.Println("⚙️ Generating configuration file...")

	// Generate configuration file using ConfigGenerator
	configGenerator := config.NewConfigGenerator()

	// Create template for configuration file generation
	templateContent := generateConfigTemplate()

	var err error
	if forceOverwrite {
		err = configGenerator.ForceGenerateConfig(templateContent)
	} else {
		err = configGenerator.GenerateConfig(templateContent)
	}

	if err != nil {
		return fmt.Errorf("configuration file generation failed: %w", err)
	}

	fmt.Println("  ✅ agents.conf configuration file created")
	return nil
}

// generateConfigTemplate configuration file template generation
func generateConfigTemplate() string {
	return `# Claude Code Agents Configuration File
# This file was automatically generated during system initialization

# Path Configurations
CLAUDE_CLI_PATH=~/.claude/local/claude
INSTRUCTIONS_DIR=~/.claude/claude-code-agents/instructions
CONFIG_DIR=~/.claude/claude-code-agents
LOG_FILE=~/.claude/claude-code-agents/logs/manager.log
AUTH_BACKUP_DIR=~/.claude/claude-code-agents/auth_backup

# System Settings
LOG_LEVEL=info

# Tmux Settings
SESSION_NAME=ai-teams
DEFAULT_LAYOUT=integrated
AUTO_ATTACH=false
IDE_BACKUP_ENABLED=true

# Commands
SEND_COMMAND=send-agent
BINARY_NAME=claude-code-agents

# Developer Settings
DEV_COUNT=4

# Role-based Instructions
PO_INSTRUCTION_FILE=po.md
MANAGER_INSTRUCTION_FILE=manager.md
DEV_INSTRUCTION_FILE=developer.md

# Timeout Settings
HEALTH_CHECK_INTERVAL=30s
AUTH_CHECK_INTERVAL=30m
STARTUP_TIMEOUT=10s
SHUTDOWN_TIMEOUT=15s
RESTART_DELAY=5s
PROCESS_TIMEOUT=30s

# === Extended Instruction Configuration ===
# To enable dynamic instruction settings, edit the following configuration

# Environment Settings
# ENVIRONMENT=development
# STRICT_VALIDATION=false
# FALLBACK_INSTRUCTION_DIR=~/.claude/claude-code-agents/fallback

# Extended instruction settings (configurable in JSON format)
# For detailed configuration, refer to documentation/instruction-config.md
`
}

// displayInitializationSuccess display initialization success message
func displayInitializationSuccess() {
	homeDir, _ := os.UserHomeDir()

	fmt.Println()
	fmt.Println("🎉 System initialization completed!")
	fmt.Println("=" + strings.Repeat("=", 38))
	fmt.Println()
	fmt.Println("📂 Created directories:")
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents"))
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "auth_backup"))
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "logs"))
	fmt.Println()
	fmt.Println("📝 Created files:")
	fmt.Printf("  • %s/agents.conf\n", filepath.Join(homeDir, ".claude", "claude-code-agents"))
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("  1. Place instruction files:")
	fmt.Printf("     • %s/po.md\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Printf("     • %s/manager.md\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Printf("     • %s/developer.md\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Println("  2. Check system health:")
	fmt.Println("     ./claude-code-agents --doctor")
	fmt.Println("  3. Authenticate with Claude CLI:")
	fmt.Println("     claude auth")
	fmt.Println("  4. Start the system:")
	fmt.Println("     ./claude-code-agents ai-teams")
	fmt.Println()
}

// GenerateConfigCommand configuration file generation command
func GenerateConfigCommand(forceOverwrite bool) error {
	fmt.Println("⚙️ Configuration File Generation")
	fmt.Println("====================")

	if forceOverwrite {
		fmt.Println("⚠️ Force overwrite mode is enabled")
	}

	// Configuration file generation process
	if err := generateInitialConfig(forceOverwrite); err != nil {
		return fmt.Errorf("configuration file generation failed: %w", err)
	}

	fmt.Println("✅ Configuration file generation completed")

	homeDir, _ := os.UserHomeDir()
	fmt.Printf("📝 Generated file: %s/agents.conf\n", filepath.Join(homeDir, ".claude", "claude-code-agents"))
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("  1. Review and edit the configuration file")
	fmt.Println("  2. Check system health: ./claude-code-agents --doctor")

	return nil
}

// gatherInstructionInfo collect instruction file information
func gatherInstructionInfo(config *config.TeamConfig) map[string]interface{} {
	info := make(map[string]interface{})

	// Basic instruction file information
	info["po_instruction_file"] = config.POInstructionFile
	info["manager_instruction_file"] = config.ManagerInstructionFile
	info["dev_instruction_file"] = config.DevInstructionFile
	info["instructions_directory"] = config.InstructionsDir

	// Check if extended instruction settings exist
	if config.InstructionConfig != nil {
		info["enhanced_config_enabled"] = true
		info["base_config"] = map[string]interface{}{
			"po_path":      config.InstructionConfig.Base.POInstructionPath,
			"manager_path": config.InstructionConfig.Base.ManagerInstructionPath,
			"dev_path":     config.InstructionConfig.Base.DevInstructionPath,
		}

		// Environment-specific settings
		if len(config.InstructionConfig.Environments) > 0 {
			info["environment_configs"] = len(config.InstructionConfig.Environments)
			envNames := make([]string, 0, len(config.InstructionConfig.Environments))
			for envName := range config.InstructionConfig.Environments {
				envNames = append(envNames, envName)
			}
			info["available_environments"] = envNames
		}

		// Global settings
		if config.InstructionConfig.Global.DefaultExtension != "" {
			info["default_extension"] = config.InstructionConfig.Global.DefaultExtension
		}
		if config.InstructionConfig.Global.CacheEnabled {
			info["cache_enabled"] = true
			info["cache_ttl"] = config.InstructionConfig.Global.CacheTTL.String()
		}
	} else {
		info["enhanced_config_enabled"] = false
	}

	// Fallback settings
	if config.FallbackInstructionDir != "" {
		info["fallback_directory"] = config.FallbackInstructionDir
	}

	// Environment settings
	if config.Environment != "" {
		info["current_environment"] = config.Environment
	}

	// Validation settings
	info["strict_validation"] = config.StrictValidation

	return info
}

// gatherEnvironmentInfo collect environment information
func gatherEnvironmentInfo(config *config.TeamConfig) map[string]interface{} {
	info := make(map[string]interface{})

	// System information
	info["claude_cli_path"] = config.ClaudeCLIPath
	info["working_directory"] = config.WorkingDir
	info["config_directory"] = config.ConfigDir
	info["log_file"] = config.LogFile
	info["session_name"] = config.SessionName
	info["dev_count"] = config.DevCount

	// tmux settings
	info["tmux_layout"] = config.DefaultLayout
	info["auto_attach"] = config.AutoAttach
	info["pane_count"] = config.PaneCount

	// Timeout settings
	info["startup_timeout"] = config.StartupTimeout.String()
	info["shutdown_timeout"] = config.ShutdownTimeout.String()
	info["process_timeout"] = config.ProcessTimeout.String()

	// Resource settings
	info["max_processes"] = config.MaxProcesses
	info["max_memory_mb"] = config.MaxMemoryMB
	info["max_cpu_percent"] = config.MaxCPUPercent

	// Monitoring settings
	info["health_check_interval"] = config.HealthCheckInterval.String()
	info["auth_check_interval"] = config.AuthCheckInterval.String()

	return info
}
