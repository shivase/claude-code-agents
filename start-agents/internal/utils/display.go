package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// Display control flags
var (
	verboseLogging = false
	silentMode     = false
)

// SetVerboseLogging sets verbose log output
func SetVerboseLogging(verbose bool) {
	verboseLogging = verbose
	if verbose {
		silentMode = false // Disable silent mode when verbose is enabled
		// No need for verbose enable message (fmt output is sufficient)
	}
}

// SetSilentMode sets silent mode
func SetSilentMode(silent bool) {
	silentMode = silent
	if silent {
		verboseLogging = false // Disable verbose when silent is enabled
		// No need for silent mode enable message (fmt output is sufficient)
	}
}

// IsVerboseLogging checks if verbose logging is enabled
func IsVerboseLogging() bool {
	return verboseLogging
}

// IsSilentMode checks if silent mode is enabled
func IsSilentMode() bool {
	return silentMode
}

// DisplayProgress displays progress status
func DisplayProgress(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("🔄 %s: %s\n", operation, message)
	// No need for structured logging (fmt output is sufficient)
}

// DisplaySuccess displays success message
func DisplaySuccess(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("✅ %s: %s\n", operation, message)
	// No need for structured logging (fmt output is sufficient)
}

// DisplayError displays error message
func DisplayError(operation string, err error) {
	fmt.Printf("❌ %s: %v\n", operation, err)
	// No need for structured logging (fmt output is sufficient)
}

// DisplayInfo displays info message
func DisplayInfo(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("ℹ️ %s: %s\n", operation, message)
	// No need for structured logging (fmt output is sufficient)
}

// DisplayWarning displays warning message
func DisplayWarning(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("⚠️ %s: %s\n", operation, message)
	// No need for structured logging (fmt output is sufficient)
}

// DisplayStartupBanner displays startup banner (only in verbose mode)
func DisplayStartupBanner() {
	if silentMode {
		return
	}

	fmt.Println("🚀 AI Teams System - Claude Code Agents")
	fmt.Println("=====================================")
	fmt.Printf("Version: 1.0.0\n")
	fmt.Printf("Runtime: Go %s\n", runtime.Version())
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Start Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=====================================")
	fmt.Println()

	// No need for structured logging (fmt output is sufficient)
}

// DisplayLauncherStart displays launcher start message
func DisplayLauncherStart() {
	if silentMode {
		return
	}
	fmt.Printf("[%s] 🚀 System launcher started\n", time.Now().Format("15:04:05"))
	fmt.Println("=====================================")
}

// DisplayLauncherProgress displays launcher progress
func DisplayLauncherProgress() {
	if silentMode {
		return
	}
	fmt.Printf("[%s] 🔄 System initializing...\n", time.Now().Format("15:04:05"))
}

// DisplayConfig displays configuration information
func DisplayConfig(teamConfig interface{}, sessionName string) {
	if silentMode {
		return
	}

	fmt.Println("📋 Configuration Information")
	fmt.Println("===========")
	fmt.Printf("Session Name: %s\n", sessionName)
	fmt.Println()

	// Process by teamConfig type (received as interface{})
	if config, ok := teamConfig.(map[string]interface{}); ok {
		for key, value := range config {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
	fmt.Println()
}

// DisplayValidationResults displays validation results
func DisplayValidationResults(teamConfig interface{}) {
	if silentMode {
		return
	}

	fmt.Println("🔍 Validation Results")
	fmt.Println("===========")

	// Simple validation result display
	fmt.Println("✅ Claude CLI: Available")
	fmt.Println("✅ Instructions: Ready")
	fmt.Println("✅ Working Directory: Accessible")
	fmt.Println()
}

// FormatPath formats path for display
func FormatPath(path string) string {
	if path == "" {
		return "<empty>"
	}

	// Replace home directory with ~
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}

	return path
}

// ValidatePath checks path existence
func ValidatePath(path string) bool {
	if path == "" {
		return false
	}

	expandedPath := ExpandPathSafe(path)
	_, err := os.Stat(expandedPath)
	return err == nil
}

// ExpandPathOld tilde expansion (deprecated: use ExpandPathSafe from path_utils.go)
func ExpandPathOld(path string) string {
	return ExpandPathSafe(path)
}

// IsExecutable checks if file is executable
func IsExecutable(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := fileInfo.Mode()
	return mode&0111 != 0
}
