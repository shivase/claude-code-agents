package cmd

import (
	"fmt"
	"time"

	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/shivase/claude-code-agents/internal/utils"
)

// DisplayConfigCommand displays detailed configuration information
func DisplayConfigCommand() error {
	fmt.Println("🔧 AI Teams System - Configuration Details")
	fmt.Println("=========================================")

	// 1. Load unified configuration
	unifiedConfig, err := config.LoadUnifiedConfig()
	if err != nil {
		fmt.Printf("⚠️ Failed to load unified configuration: %v\n", err)
		fmt.Println("📝 Displaying basic configuration only")

		// Fallback: display basic configuration only
		displayBasicConfigFallback()
		return nil
	}

	// 2. Display all TeamConfig values in detail
	fmt.Println("\n📁 TeamConfig - Team Settings")
	fmt.Println("---------------------------")
	displayTeamConfig(unifiedConfig.Team)

	// 3. Display all CommonConfig values in detail
	fmt.Println("\n⚙️ CommonConfig - Common Settings")
	fmt.Println("----------------------------")
	fmt.Println("   ⚠️ CommonConfig has been removed (to resolve import cycle)")

	// 4. Display complete path information
	fmt.Println("\n📂 Path Configuration - Path Settings")
	fmt.Println("----------------------------------")
	displayPathConfiguration(unifiedConfig.Paths)

	// 5. Display system settings in detail
	fmt.Println("\n🖥️ System Settings")
	fmt.Println("-----------------------------------")
	displaySystemSettings(unifiedConfig.Team)

	// 6. Display authentication settings in detail
	fmt.Println("\n🔐 Authentication Settings")
	fmt.Println("--------------------------------------")
	displayAuthenticationSettings(unifiedConfig.Team)

	// 7. Display configuration file existence check and validation results
	fmt.Println("\n📋 Configuration File Validation")
	fmt.Println("----------------------------------------------------")
	displayConfigurationValidation(unifiedConfig.Paths)

	// 8. Display effective configuration values
	fmt.Println("\n✅ Effective Configuration")
	fmt.Println("----------------------------------------")
	fmt.Println("   Effective configuration display is under implementation")

	// 9. Directory resolution information
	fmt.Println("\n📁 Directory Resolution")
	fmt.Println("------------------------------------------")
	resolver := utils.GetGlobalDirectoryResolver()
	resolver.DisplayDirectoryInfo()

	fmt.Println("=========================================")
	fmt.Printf("🕐 Configuration display completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return nil
}

// displayTeamConfig displays TeamConfig details
func displayTeamConfig(teamConfig *config.TeamConfig) {
	if teamConfig == nil {
		fmt.Println("⚠️ TeamConfig not loaded")
		return
	}

	fmt.Printf("   Claude CLI Path:      %s\n", teamConfig.ClaudeCLIPath)
	fmt.Printf("   Instructions Dir:     %s\n", teamConfig.InstructionsDir)
	fmt.Printf("   Working Dir:          %s\n", teamConfig.WorkingDir)
	fmt.Printf("   Config Dir:           %s\n", teamConfig.ConfigDir)
	fmt.Printf("   Log File:             %s\n", teamConfig.LogFile)
	fmt.Printf("   Auth Backup Dir:      %s\n", teamConfig.AuthBackupDir)
	fmt.Printf("   Max Processes:        %d\n", teamConfig.MaxProcesses)
	fmt.Printf("   Max Memory (MB):      %d\n", teamConfig.MaxMemoryMB)
	fmt.Printf("   Max CPU Percent:      %.1f%%\n", teamConfig.MaxCPUPercent)
	fmt.Printf("   Log Level:            %s\n", teamConfig.LogLevel)
	fmt.Printf("   Session Name:         %s\n", teamConfig.SessionName)
	fmt.Printf("   Default Layout:       %s\n", teamConfig.DefaultLayout)
	fmt.Printf("   Auto Attach:          %t\n", teamConfig.AutoAttach)
	fmt.Printf("   Pane Count:           %d\n", teamConfig.PaneCount)
	fmt.Printf("   IDE Backup Enabled:   %t\n", teamConfig.IDEBackupEnabled)
	fmt.Printf("   Send Command:         %s\n", teamConfig.SendCommand)
	fmt.Printf("   Binary Name:          %s\n", teamConfig.BinaryName)
	fmt.Printf("   Health Check Interval: %s\n", teamConfig.HealthCheckInterval)
	fmt.Printf("   Auth Check Interval:  %s\n", teamConfig.AuthCheckInterval)
	fmt.Printf("   Startup Timeout:      %s\n", teamConfig.StartupTimeout)
	fmt.Printf("   Shutdown Timeout:     %s\n", teamConfig.ShutdownTimeout)
	fmt.Printf("   Restart Delay:        %s\n", teamConfig.RestartDelay)
	fmt.Printf("   Process Timeout:      %s\n", teamConfig.ProcessTimeout)
	fmt.Printf("   Max Restart Attempts: %d\n", teamConfig.MaxRestartAttempts)
}

// displayPathConfiguration displays path configuration details
func displayPathConfiguration(paths *config.ConfigPaths) {
	if paths == nil {
		fmt.Println("⚠️ Path Configuration not loaded")
		return
	}

	fmt.Printf("   Claude Dir:           %s\n", paths.ClaudeDir)
	fmt.Printf("   Cloud Code Agents Dir: %s\n", paths.CloudCodeAgentsDir)
	fmt.Printf("   Team Config Path:     %s\n", paths.TeamConfigPath)
	fmt.Printf("   Main Config Path:     %s\n", paths.MainConfigPath)
	fmt.Printf("   Logs Dir:             %s\n", paths.LogsDir)
	fmt.Printf("   Instructions Dir:     %s\n", paths.InstructionsDir)
	fmt.Printf("   Auth Backup Dir:      %s\n", paths.AuthBackupDir)
	fmt.Printf("   Claude CLI Path:      %s\n", paths.ClaudeCLIPath)
}

// displaySystemSettings displays system settings details
func displaySystemSettings(teamConfig *config.TeamConfig) {
	if teamConfig == nil {
		fmt.Println("⚠️ System Settings not loaded")
		return
	}

	fmt.Printf("   Max Processes:        %d\n", teamConfig.MaxProcesses)
	fmt.Printf("   Max Memory Usage:     %d MB\n", teamConfig.MaxMemoryMB)
	fmt.Printf("   Max CPU Usage:        %.1f%%\n", teamConfig.MaxCPUPercent)
	fmt.Printf("   Health Check Interval: %s\n", teamConfig.HealthCheckInterval)
	fmt.Printf("   Max Restart Attempts: %d\n", teamConfig.MaxRestartAttempts)
	fmt.Printf("   Process Timeout:      %s\n", teamConfig.ProcessTimeout)
	fmt.Printf("   Startup Timeout:      %s\n", teamConfig.StartupTimeout)
	fmt.Printf("   Shutdown Timeout:     %s\n", teamConfig.ShutdownTimeout)
	fmt.Printf("   Restart Delay:        %s\n", teamConfig.RestartDelay)
}

// displayAuthenticationSettings displays authentication settings details
func displayAuthenticationSettings(teamConfig *config.TeamConfig) {
	if teamConfig == nil {
		fmt.Println("⚠️ Authentication Settings not loaded")
		return
	}

	fmt.Printf("   Auth Check Interval:  %s\n", teamConfig.AuthCheckInterval)
	fmt.Printf("   Auth Backup Dir:      %s\n", teamConfig.AuthBackupDir)
	fmt.Printf("   Claude CLI Path:      %s\n", teamConfig.ClaudeCLIPath)
}

// displayConfigurationValidation validates configuration files
func displayConfigurationValidation(paths *config.ConfigPaths) {
	if paths == nil {
		fmt.Println("⚠️ Path Configuration not loaded")
		return
	}

	fmt.Printf("   Team Config:          %s", paths.TeamConfigPath)
	if utils.ValidatePath(paths.TeamConfigPath) {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ❌")
	}

	fmt.Printf("   Instructions Dir:     %s", paths.InstructionsDir)
	if utils.ValidatePath(paths.InstructionsDir) {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ❌")
	}

	fmt.Printf("   Claude CLI:           %s", paths.ClaudeCLIPath)
	if utils.IsExecutable(utils.ExpandPathSafe(paths.ClaudeCLIPath)) {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ❌")
	}
}

// displayBasicConfigFallback displays basic configuration fallback
func displayBasicConfigFallback() {
	fmt.Println("\n📁 Basic Configuration Information")
	fmt.Println("--------------")

	// Display only basic configuration information
	configPath := config.GetDefaultTeamConfigPath()
	fmt.Printf("   Config File Path:     %s\n", configPath)

	if utils.ValidatePath(configPath) {
		fmt.Println("   Config File Status:   ✅ Exists")
	} else {
		fmt.Println("   Config File Status:   ❌ Not Found")
	}
}

// DisplaySessionConfigCommand displays session configuration details
func DisplaySessionConfigCommand(sessionName string) error {
	fmt.Printf("🔧 Session Configuration Details: %s\n", sessionName)
	fmt.Println("=====================================")

	// Display session-specific configuration information
	fmt.Printf("   Session Name:         %s\n", sessionName)
	fmt.Printf("   Display Time:         %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return nil
}
