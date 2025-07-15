package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shivase/claude-code-agents/internal/auth"
	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/shivase/claude-code-agents/internal/utils"
)

// DoctorCommand performs system health check
func DoctorCommand() error {
	fmt.Println("🏥 System Health Check")
	fmt.Println("=====================")
	fmt.Println()

	var overallStatus = true
	var issues []string

	// Basic environment check
	fmt.Println("🔍 Basic Environment Check")
	fmt.Println("-------------------------")

	// Load configuration file
	fmt.Print("📋 Loading configuration file... ")
	configPath := config.GetDefaultTeamConfigPath()
	configLoader := config.NewTeamConfigLoader(configPath)
	teamConfig, err := configLoader.LoadTeamConfig()
	if err != nil {
		fmt.Printf("❌ Failed\n")
		fmt.Printf("   Error: %v\n", err)
		overallStatus = false
		issues = append(issues, "Failed to load configuration file")
	} else {
		fmt.Printf("✅ Success\n")
	}

	if teamConfig == nil {
		fmt.Println("\n❌ Cannot load configuration file, skipping further checks")
		return fmt.Errorf("failed to load configuration file")
	}

	fmt.Println()

	// Path existence check
	fmt.Println("📂 Important Files & Directories Check")
	fmt.Println("-------------------------------------")

	pathChecks := []struct {
		name        string
		path        string
		required    bool
		description string
	}{
		{"Claude CLI executable", teamConfig.ClaudeCLIPath, true, "Required for Claude CLI execution"},
		{"Instructions directory", teamConfig.InstructionsDir, true, "Stores agent instruction files"},
		{"Working directory", teamConfig.WorkingDir, true, "System execution location"},
		{"Config directory", teamConfig.ConfigDir, true, "Stores configuration files"},
		{"Auth backup directory", teamConfig.AuthBackupDir, false, "Backup of authentication info"},
		{"Log directory", filepath.Dir(teamConfig.LogFile), false, "Stores log files"},
	}

	for _, check := range pathChecks {
		fmt.Printf("📍 %s: ", check.name)

		expandedPath := utils.ExpandPathSafe(check.path)
		exists := utils.ValidatePath(check.path)

		switch exists {
		case true:
			fmt.Printf("✅ Exists (%s)\n", utils.FormatPath(check.path))
		case false:
			icon := "❌"
			if !check.required {
				icon = "⚠️"
			}
			fmt.Printf("%s Not found (%s)\n", icon, utils.FormatPath(check.path))

			if check.required {
				overallStatus = false
				issues = append(issues, fmt.Sprintf("%s does not exist: %s", check.name, expandedPath))
			} else {
				issues = append(issues, fmt.Sprintf("Optional: %s does not exist: %s", check.name, expandedPath))
			}
		}
		fmt.Printf("   Description: %s\n", check.description)
		fmt.Println()
	}

	// List instruction files
	fmt.Println("📄 Instruction Files List")
	fmt.Println("------------------------")

	instructionsDir := filepath.Join(os.Getenv("HOME"), ".claude", "claude-code-agents", "instructions")
	files, err := os.ReadDir(instructionsDir)
	if err != nil {
		fmt.Printf("📂 Instructions directory: %s\n", instructionsDir)
		fmt.Printf("⚠️  Directory does not exist or cannot be read\n")
		fmt.Printf("💡 To create: mkdir -p %s\n", instructionsDir)
	} else {
		fmt.Printf("📂 Instructions directory: %s\n", instructionsDir)
		if len(files) == 0 {
			fmt.Printf("📝 Files found: 0\n")
			fmt.Printf("💡 Example role files: po.md, manager.md, developer.md\n")
		} else {
			fmt.Printf("📝 Files found: %d\n", len(files))
			for _, file := range files {
				if !file.IsDir() {
					fmt.Printf("   📄 %s\n", file.Name())
				}
			}
		}
	}

	fmt.Println()

	// Claude CLI executability check
	fmt.Println("🤖 Claude CLI Executability Check")
	fmt.Println("--------------------------------")

	fmt.Print("🔧 Checking execution permissions... ")
	if utils.IsExecutable(utils.ExpandPathSafe(teamConfig.ClaudeCLIPath)) {
		fmt.Printf("✅ Executable\n")
	} else {
		fmt.Printf("❌ Not executable\n")
		overallStatus = false
		issues = append(issues, "Claude CLI does not have execution permission")
	}

	// Claude authentication check (skip auth test to prevent OAuth conflicts)
	fmt.Print("🔐 Claude authentication check... ")
	claudeAuth := auth.NewClaudeAuthManager()
	if err := claudeAuth.CheckSettingsFile(); err != nil {
		fmt.Printf("❌ Failed to verify settings file\n")
		fmt.Printf("   Error: %v\n", err)
		overallStatus = false
		issues = append(issues, "Problem with Claude settings file")
	} else {
		fmt.Printf("✅ Settings file OK (API auth test skipped)\n")
	}

	fmt.Println()

	// Overall result
	fmt.Println("📊 Diagnosis Result")
	fmt.Println("==================")

	if overallStatus {
		fmt.Println("🎉 System is ready to operate normally!")
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Println("   1. Start the system with: claude-code-agents [session-name]")
		fmt.Println("   2. Verify Claude CLI works properly in each pane")
	} else {
		fmt.Println("⚠️ Problems detected in the system")
		fmt.Println()
		fmt.Println("🔧 Issues that need fixing:")
		for i, issue := range issues {
			fmt.Printf("   %d. %s\n", i+1, issue)
		}
		fmt.Println()
		fmt.Println("💡 Solutions:")
		fmt.Println("   1. Create missing files and directories")
		fmt.Println("   2. Verify Claude CLI is installed correctly")
		fmt.Println("   3. Authenticate with 'claude auth' command")
		fmt.Println("   4. Modify configuration file as needed")
	}

	fmt.Println()
	fmt.Printf("Diagnosis completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	if !overallStatus {
		return fmt.Errorf("%d issues detected in the system", len(issues))
	}

	return nil
}

// DoctorDetailedCommand performs detailed system diagnostics (moved from main.go)
func DoctorDetailedCommand() error {
	fmt.Println("🏥 Starting system diagnostics...")
	fmt.Println("=================================")

	var errors []string
	var warnings []string

	// 1. Path validation (executables, config directories)
	fmt.Println("\n📁 Validating paths...")
	if pathErrors := ValidatePathsDetailed(); len(pathErrors) > 0 {
		errors = append(errors, pathErrors...)
	} else {
		fmt.Println("✅ Path validation: OK")
	}

	// 2. Configuration file validation (existence check, validity)
	fmt.Println("\n⚙️ Validating configuration files...")
	if configErrors := ValidateConfigurationDetailed(); len(configErrors) > 0 {
		errors = append(errors, configErrors...)
	} else {
		fmt.Println("✅ Configuration file validation: OK")
	}

	// 3. Claude authentication check (auth status, token validation)
	fmt.Println("\n🔐 Checking Claude authentication status...")
	if authErrors := ValidateAuthenticationDetailed(); len(authErrors) > 0 {
		warnings = append(warnings, authErrors...)
	} else {
		fmt.Println("✅ Claude authentication check: OK")
	}

	// 4. System environment check (OS, permissions, dependencies)
	fmt.Println("\n🖥️ Checking system environment...")
	if envErrors := ValidateEnvironmentDetailed(); len(envErrors) > 0 {
		errors = append(errors, envErrors...)
	} else {
		fmt.Println("✅ System environment check: OK")
	}

	// 5. tmux connection check (maintain legacy functionality)
	fmt.Println("\n🔧 Checking tmux connection...")
	fmt.Print("📺 tmux availability... ")
	if _, err := exec.LookPath("tmux"); err != nil {
		errors = append(errors, "tmux is not installed")
		fmt.Printf("❌ tmux not found\n")
	} else {
		fmt.Printf("✅ tmux available\n")
	}

	// Display detailed diagnosis results
	fmt.Println("\n=================================")
	fmt.Println("🔍 Detailed Diagnosis Results")
	fmt.Println("=================================")

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("🎉 System diagnosis complete - All checks passed")
		fmt.Printf("Diagnosis completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	} else {
		if len(errors) > 0 {
			fmt.Println("\n❌ Problems detected:")
			for i, err := range errors {
				fmt.Printf("   %d. %s\n", i+1, err)
			}
			fmt.Println("\n💡 Solutions:")
			DisplaySolutionsForErrors(errors)
		}

		if len(warnings) > 0 {
			fmt.Println("\n⚠️ Warnings:")
			for i, warning := range warnings {
				fmt.Printf("   %d. %s\n", i+1, warning)
			}
			fmt.Println("\n💡 Recommendations:")
			DisplaySolutionsForWarnings(warnings)
		}

		if len(errors) > 0 {
			fmt.Println("\n❌ Critical issues found. Please apply the solutions above.")
			return fmt.Errorf("%d issues detected during system diagnosis", len(errors))
		} else {
			fmt.Println("\n✅ No critical issues found, but please review the warnings.")
		}
	}

	return nil
}
