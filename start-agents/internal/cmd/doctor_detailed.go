// Package doctor - detailed diagnostic functions
// This file contains the detailed diagnostic functionality for the --doctor command

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shivase/claude-code-agents/internal/auth"
	"github.com/shivase/claude-code-agents/internal/utils"
)

// ValidatePathsDetailed validates paths in detail
func ValidatePathsDetailed() []string {
	var errors []string

	// Validate Claude CLI executable
	claudePath := findClaudeExecutableHelper()
	if claudePath == "" {
		errors = append(errors, "Claude CLI executable not found")
	} else {
		fmt.Printf("   ✅ Claude CLI: %s\n", claudePath)
	}

	// Path validation using directory resolver
	dirResolver := utils.GetGlobalDirectoryResolver()
	dirInfo := dirResolver.GetDirectoryInfo()

	fmt.Printf("   📂 Project root: %s\n", dirInfo["project_root"])
	fmt.Printf("   📂 Working directory: %s\n", dirInfo["original_working_dir"])

	// Check required directories
	homeDir, _ := os.UserHomeDir()
	requiredDirs := []string{
		filepath.Join(homeDir, ".claude"),
		filepath.Join(homeDir, ".claude", "claude-code-agents"),
		filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"),
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); err != nil {
			errors = append(errors, fmt.Sprintf("Required directory not found: %s", dir))
		} else {
			fmt.Printf("   ✅ Directory check: %s\n", dir)
		}
	}

	return errors
}

// ValidateConfigurationDetailed validates configuration files in detail
func ValidateConfigurationDetailed() []string {
	var errors []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		errors = append(errors, "Failed to get home directory")
		return errors
	}

	// Check ~/.claude/settings.json
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		errors = append(errors, "Claude settings file (settings.json) not found")
	} else {
		// Basic check for file size and content
		info, err := os.Stat(settingsPath)
		if err != nil || info.Size() == 0 {
			errors = append(errors, "Claude settings file (settings.json) is empty")
		} else {
			fmt.Printf("   ✅ settings.json: %s (%d bytes)\n", settingsPath, info.Size())
		}
	}

	// Check ~/.claude/claude.json
	claudeJsonPath := filepath.Join(homeDir, ".claude", "claude.json")
	if _, err := os.Stat(claudeJsonPath); err != nil {
		errors = append(errors, "Claude authentication file (claude.json) not found")
	} else {
		info, err := os.Stat(claudeJsonPath)
		if err != nil || info.Size() == 0 {
			errors = append(errors, "Claude authentication file (claude.json) is empty")
		} else {
			fmt.Printf("   ✅ claude.json: %s (%d bytes)\n", claudeJsonPath, info.Size())
		}
	}

	return errors
}

// ValidateAuthenticationDetailed checks Claude authentication in detail
func ValidateAuthenticationDetailed() []string {
	var warnings []string

	// Check authentication using authentication manager
	authManager := auth.NewClaudeAuthManager()

	// Check settings file
	if err := authManager.CheckSettingsFile(); err != nil {
		warnings = append(warnings, fmt.Sprintf("Settings file check failed: %v", err))
	} else {
		fmt.Printf("   ✅ Settings file check completed\n")
	}

	// Check authentication status
	authStatus, err := authManager.CheckAuthenticationStatus()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Authentication status check failed: %v", err))
	} else {
		if authStatus.IsAuthenticated {
			if authStatus.UserID != "" {
				fmt.Printf("   ✅ Authenticated (UserID: %s...)\n", authStatus.UserID[:8])
			}
			if authStatus.OAuthAccount != nil {
				if email, exists := authStatus.OAuthAccount["emailAddress"]; exists {
					fmt.Printf("   ✅ OAuth authenticated: %v\n", email)
				}
			}
		} else {
			warnings = append(warnings, "Claude authentication not completed")
		}
	}

	return warnings
}

// ValidateEnvironmentDetailed checks system environment in detail
func ValidateEnvironmentDetailed() []string {
	var errors []string

	// Check OS information
	fmt.Printf("   🖥️ OS: %s\n", runtime.GOOS)
	fmt.Printf("   🏗️ Architecture: %s\n", runtime.GOARCH)

	// Check permissions
	homeDir, _ := os.UserHomeDir()
	claudeDir := filepath.Join(homeDir, ".claude")

	// Check write permission for .claude directory
	testFile := filepath.Join(claudeDir, "test_write")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		errors = append(errors, fmt.Sprintf("No write permission for .claude directory: %v", err))
	} else {
		_ = os.Remove(testFile) // Remove test file
		fmt.Printf("   ✅ Directory write permission: OK\n")
	}

	// Check dependencies
	dependencies := []string{"tmux"}
	for _, dep := range dependencies {
		if path, err := exec.LookPath(dep); err != nil {
			errors = append(errors, fmt.Sprintf("Dependency '%s' not found", dep))
		} else {
			fmt.Printf("   ✅ Dependency %s: %s\n", dep, path)
		}
	}

	// Check environment variables
	shell := os.Getenv("SHELL")
	if shell == "" {
		errors = append(errors, "SHELL environment variable not set")
	} else {
		fmt.Printf("   ✅ SHELL: %s\n", shell)
	}

	return errors
}

// DisplaySolutionsForErrors displays solutions for errors
func DisplaySolutionsForErrors(errors []string) {
	for _, err := range errors {
		switch {
		case strings.Contains(err, "Claude CLI executable"):
			fmt.Println("   → Please install Claude CLI")
			fmt.Println("      curl -fsSL https://anthropic.com/claude/install.sh | sh")
		case strings.Contains(err, "Required directory"):
			fmt.Println("   → Please create required directories")
			fmt.Println("      mkdir -p ~/.claude/claude-code-agents/instructions")
		case strings.Contains(err, "settings.json"):
			fmt.Println("   → Please start Claude CLI and complete initial setup")
			fmt.Println("      claude")
		case strings.Contains(err, "claude.json"):
			fmt.Println("   → Please log in to Claude CLI")
			fmt.Println("      claude")
		case strings.Contains(err, "tmux"):
			fmt.Println("   → Please install tmux")
			fmt.Println("      macOS: brew install tmux")
			fmt.Println("      Ubuntu: sudo apt install tmux")
		case strings.Contains(err, "write permission"):
			fmt.Println("   → Please check directory permissions")
			fmt.Println("      chmod 750 ~/.claude")
		case strings.Contains(err, "SHELL environment variable"):
			fmt.Println("   → Please set SHELL environment variable")
			fmt.Println("      export SHELL=/bin/bash")
		}
	}
}

// DisplaySolutionsForWarnings displays recommendations for warnings
func DisplaySolutionsForWarnings(warnings []string) {
	for _, warning := range warnings {
		switch {
		case strings.Contains(warning, "authentication not completed"):
			fmt.Println("   → Recommend logging in to Claude CLI")
			fmt.Println("      claude")
		case strings.Contains(warning, "Authentication status check failed"):
			fmt.Println("   → Consider reinstalling Claude CLI")
		case strings.Contains(warning, "Settings file check failed"):
			fmt.Println("   → Please recreate Claude CLI settings")
		}
	}
}

// findClaudeExecutableHelper searches for Claude CLI executable
func findClaudeExecutableHelper() string {
	// Check common paths in order
	paths := []string{
		"~/.claude/local/claude",
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude",
	}

	for _, path := range paths {
		if strings.HasPrefix(path, "~") {
			homeDir, _ := os.UserHomeDir()
			path = strings.Replace(path, "~", homeDir, 1)
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Search from PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path
	}

	return ""
}
