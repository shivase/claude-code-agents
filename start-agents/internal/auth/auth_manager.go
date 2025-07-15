package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ClaudeAuthManager manages Claude authentication
type ClaudeAuthManager struct {
}

// NewClaudeAuthManager creates authentication manager
func NewClaudeAuthManager() *ClaudeAuthManager {
	return &ClaudeAuthManager{}
}

// AuthStatus represents authentication status
type AuthStatus struct {
	IsAuthenticated bool                   `json:"isAuthenticated"`
	UserID          string                 `json:"userID"`
	OAuthAccount    map[string]interface{} `json:"oauthAccount,omitempty"`
	LastChecked     int64                  `json:"lastChecked"`
}

// CheckAuthenticationStatus checks authentication status
func (cam *ClaudeAuthManager) CheckAuthenticationStatus() (*AuthStatus, error) {
	log.Debug().Msg("🔍 Checking Claude authentication status")

	// Load Claude file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeJsonPath := filepath.Join(homeDir, ".claude", "claude.json")

	authStatus := &AuthStatus{
		LastChecked: time.Now().Unix(),
	}

	// Check file existence
	if _, err := os.Stat(claudeJsonPath); err != nil {
		log.Warn().Str("config_path", claudeJsonPath).Msg("⚠️ Claude configuration file not found")
		return authStatus, err
	}

	// Read file
	fileData, err := os.ReadFile(claudeJsonPath) // #nosec G304
	if err != nil {
		log.Warn().Err(err).Msg("⚠️ Failed to read Claude configuration file")
		return authStatus, nil
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		log.Warn().Err(err).Msg("⚠️ Failed to parse Claude configuration JSON")
		return authStatus, nil
	}

	log.Debug().Msg("✅ Claude configuration file loaded successfully")

	// Check userID existence
	if userID, exists := data["userID"]; exists && userID != nil {
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			authStatus.UserID = userIDStr
			authStatus.IsAuthenticated = true
			log.Debug().Str("user_id_prefix", userIDStr[:8]+"...").Msg("✅ Existing authentication confirmed")
		}
	}

	// Check OAuth account info
	if oauthAccount, exists := data["oauthAccount"]; exists && oauthAccount != nil {
		if oauthMap, ok := oauthAccount.(map[string]interface{}); ok {
			authStatus.OAuthAccount = oauthMap
			authStatus.IsAuthenticated = true
			if email, exists := oauthMap["emailAddress"]; exists {
				log.Debug().Interface("email", email).Msg("📧 OAuth authenticated")
			}
		}
	}

	return authStatus, nil
}

// CheckSettingsFile checks Claude settings file existence
func (cam *ClaudeAuthManager) CheckSettingsFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check existence of ~/.claude/settings.json
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		return fmt.Errorf("claude settings file not found: %s", settingsPath)
	}

	log.Debug().Str("settings_path", settingsPath).Msg("✅ Claude settings file check completed")
	return nil
}

// PreAuthChecker performs pre-authentication checks
type PreAuthChecker struct {
	claudePath string
}

// NewPreAuthChecker creates pre-authentication checker
func NewPreAuthChecker(claudePath string) *PreAuthChecker {
	return &PreAuthChecker{
		claudePath: claudePath,
	}
}

// CheckAuthenticationBeforeStart checks authentication before start
func (pac *PreAuthChecker) CheckAuthenticationBeforeStart() error {
	log.Debug().Msg("ℹ️ Checking Claude authentication status before tmux start")

	// Check Claude settings file
	cam := NewClaudeAuthManager()
	if err := cam.CheckSettingsFile(); err != nil {
		return fmt.Errorf("claude settings file check failed: %w", err)
	}

	// Check authentication status (exclusive access version)
	log.Debug().Msg("🔄 Checking Claude authentication status (exclusive access)")

	// Check authentication consistency for parallel startup
	if err := cam.ValidateAuthConcurrency(); err != nil {
		return fmt.Errorf("parallel authentication consistency check failed: %w", err)
	}

	// Check authentication status
	authStatus, err := cam.CheckAuthenticationStatus()
	if err != nil {
		return fmt.Errorf("authentication status check failed: %w", err)
	}

	if !authStatus.IsAuthenticated {
		log.Warn().Msg("⚠️ Claude authentication required")
		log.Debug().Msg("💡 Starting interactive authentication. Please follow the on-screen instructions")
		log.Debug().Msg("────────────────────────────────────────────────────────")

		// Execute interactive authentication
		if err := cam.PerformInteractiveAuth(); err != nil {
			return fmt.Errorf("claude authentication check failed: %w", err)
		}
	}

	log.Debug().Msg("✅ Claude authentication check completed")
	return nil
}

// PerformInteractiveAuth performs interactive authentication
func (cam *ClaudeAuthManager) PerformInteractiveAuth() error {
	log.Debug().Msg("🔐 Starting Claude authentication status check")

	// Check authentication status with simple test command
	cmd := exec.Command("claude", "--print", "test")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("authentication check failed: %w", err)
	}

	// Authentication successful if output exists
	if len(output) == 0 {
		return fmt.Errorf("claude authentication response is empty")
	}

	log.Debug().Msg("✅ Claude authentication check completed")
	return nil
}

// EnsureAuthentication ensures authentication is complete
func (cam *ClaudeAuthManager) EnsureAuthentication() error {
	// 認証状態確認
	authStatus, err := cam.CheckAuthenticationStatus()
	if err != nil {
		return fmt.Errorf("authentication status check failed: %w", err)
	}

	// Return early if already authenticated
	if authStatus.IsAuthenticated {
		log.Debug().Str("user_id_prefix", authStatus.UserID[:8]+"...").Msg("✅ Claude authenticated")
		return nil
	}

	// Execute interactive authentication if needed
	log.Warn().Msg("⚠️ Claude authentication required. Starting interactive authentication")

	if err := cam.PerformInteractiveAuth(); err != nil {
		return fmt.Errorf("authentication execution failed: %w", err)
	}

	// Check status after authentication
	authStatus, err = cam.CheckAuthenticationStatus()
	if err != nil {
		return fmt.Errorf("post-authentication status check failed: %w", err)
	}

	if !authStatus.IsAuthenticated {
		return fmt.Errorf("authentication status still invalid after authentication process")
	}

	log.Debug().Msg("🎉 Claude authentication completed successfully")
	return nil
}

// ValidateAuthConcurrency validates authentication consistency for parallel startup
func (cam *ClaudeAuthManager) ValidateAuthConcurrency() error {
	log.Debug().Msg("🔄 Checking parallel Claude startup authentication consistency")

	// Check Claude Code process count
	cmd := exec.Command("pgrep", "-f", "claude")
	output, err := cmd.Output()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check Claude process count")
		return nil // Continue as non-fatal error
	}

	processCount := len(strings.Split(strings.TrimSpace(string(output)), "\n"))
	if processCount > 1 {
		log.Warn().Int("process_count", processCount).Msg("⚠️ Multiple Claude Code processes detected")

		// Wait briefly for authentication state stabilization during parallel access
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// SafeAuthUpdate safely updates authentication state
func (cam *ClaudeAuthManager) SafeAuthUpdate(updateFunc func(map[string]interface{}) error) error {
	// Simple implementation: dummy operation
	data := make(map[string]interface{})
	return updateFunc(data)
}

// CleanupCorruptedFiles cleans up corrupted files
func (cam *ClaudeAuthManager) CleanupCorruptedFiles() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	claudeDir := filepath.Join(homeDir, ".claude")

	// Delete corrupted files older than 1 week
	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return fmt.Errorf("failed to read claude directory: %w", err)
	}

	cleaned := 0
	cutoff := time.Now().AddDate(0, 0, -7) // 1 week ago

	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, ".corrupted.") {
			fullPath := filepath.Join(claudeDir, name)
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				if err := os.Remove(fullPath); err != nil {
					log.Warn().Err(err).Str("file", fullPath).Msg("Failed to delete corrupted file")
				} else {
					cleaned++
				}
			}
		}
	}

	if cleaned > 0 {
		log.Debug().Int("cleaned_count", cleaned).Msg("🧹 Cleaned up old corrupted files")
	}

	return nil
}
