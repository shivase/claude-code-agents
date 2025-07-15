package auth

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// AuthBackupManager manages authentication backup
type AuthBackupManager struct {
	homeDir   string
	claudeDir string
	backupDir string
	retention time.Duration
}

// NewAuthBackupManager creates authentication backup manager
func NewAuthBackupManager() (*AuthBackupManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	backupDir := fmt.Sprintf("/tmp/claude_auth_backup_%d", time.Now().Unix())

	return &AuthBackupManager{
		homeDir:   homeDir,
		claudeDir: claudeDir,
		backupDir: backupDir,
		retention: 24 * time.Hour,
	}, nil
}

// BackupIDEAuth backs up IDE integration authentication info
func (abm *AuthBackupManager) BackupIDEAuth() error {
	log.Info().Str("backup_dir", abm.backupDir).Msg("💾 Starting IDE authentication backup")

	// Create backup directory
	if err := os.MkdirAll(abm.backupDir, 0750); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Check IDE integration directory existence
	ideDir := filepath.Join(abm.claudeDir, "ide")
	if _, err := os.Stat(ideDir); os.IsNotExist(err) {
		log.Info().Msg("ℹ️ IDE directory not found, skipping backup")
		return nil
	}

	// Backup IDE integration info
	backupIdeDir := filepath.Join(abm.backupDir, "ide")
	if err := abm.copyDir(ideDir, backupIdeDir); err != nil {
		return fmt.Errorf("failed to backup IDE directory: %w", err)
	}

	log.Info().Msg("✅ IDE authentication backup successful")
	return nil
}

// RestoreIDEAuth restores IDE integration authentication info
func (abm *AuthBackupManager) RestoreIDEAuth() error {
	log.Info().Str("backup_dir", abm.backupDir).Msg("💾 Starting IDE authentication restore")

	// Check backup directory existence
	backupIdeDir := filepath.Join(abm.backupDir, "ide")
	if _, err := os.Stat(backupIdeDir); os.IsNotExist(err) {
		log.Info().Msg("ℹ️ IDE backup not found, skipping restore")
		return nil
	}

	// Restore IDE integration info
	ideDir := filepath.Join(abm.claudeDir, "ide")
	if err := abm.copyDir(backupIdeDir, ideDir); err != nil {
		return fmt.Errorf("failed to restore IDE directory: %w", err)
	}

	log.Info().Msg("✅ IDE authentication restore successful")
	return nil
}

// CleanupBackup removes backup directory
func (abm *AuthBackupManager) CleanupBackup() error {
	if _, err := os.Stat(abm.backupDir); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(abm.backupDir); err != nil {
		return fmt.Errorf("failed to cleanup backup directory: %w", err)
	}

	log.Info().Str("backup_dir", abm.backupDir).Msg("✅ Backup cleanup completed")
	return nil
}

// copyDir recursively copies directory
func (abm *AuthBackupManager) copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Process each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy directories
			if err := abm.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := abm.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a file
func (abm *AuthBackupManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			log.Warn().Err(err).Msg("⚠️ Failed to close source file")
		}
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode()) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			log.Warn().Err(err).Msg("⚠️ Failed to close destination file")
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// ConfigProtector protects configuration
type ConfigProtector struct {
	claudeDir string
	lockFile  string
}

// NewConfigProtector creates configuration protection system
func NewConfigProtector() (*ConfigProtector, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	lockFile := filepath.Join(claudeDir, "config.lock")

	return &ConfigProtector{
		claudeDir: claudeDir,
		lockFile:  lockFile,
	}, nil
}

// ProtectExistingConfig protects existing configuration
func (cp *ConfigProtector) ProtectExistingConfig() error {
	settingsFile := filepath.Join(cp.claudeDir, "settings.json")

	// Check existing configuration
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		log.Info().Msg("ℹ️ Existing Claude configuration not found")
		return nil
	}

	// Validate configuration file content
	if err := cp.ValidateConfig(); err != nil {
		return fmt.Errorf("existing config validation failed: %w", err)
	}

	// Create lock file
	if err := os.WriteFile(cp.lockFile, []byte("protected"), 0600); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	log.Info().Msg("✅ Protected existing Claude configuration")
	return nil
}

// ValidateConfig validates configuration file
func (cp *ConfigProtector) ValidateConfig() error {
	settingsFile := filepath.Join(cp.claudeDir, "settings.json")

	// Check file existence
	info, err := os.Stat(settingsFile)
	if err != nil {
		return fmt.Errorf("config file not found: %w", err)
	}

	// Check file size
	if info.Size() == 0 {
		return fmt.Errorf("config file is empty")
	}

	// Check file readability - path normalization and directory traversal prevention
	cleanPath := filepath.Clean(settingsFile)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("config file path contains directory traversal")
	}
	// Test readability
	testFile, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("config file is not readable: %w", err)
	}
	defer func() {
		if err := testFile.Close(); err != nil {
			log.Warn().Err(err).Msg("⚠️ Failed to close test file")
		}
	}()

	log.Info().Msg("✅ Configuration file validation successful")
	return nil
}

// IsConfigProtected checks if configuration is protected
func (cp *ConfigProtector) IsConfigProtected() bool {
	_, err := os.Stat(cp.lockFile)
	return err == nil
}

// UnlockConfig unlocks configuration protection
func (cp *ConfigProtector) UnlockConfig() error {
	if !cp.IsConfigProtected() {
		return nil
	}

	if err := os.Remove(cp.lockFile); err != nil {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	log.Info().Msg("✅ Configuration protection unlocked")
	return nil
}

// AuthManager integrates authentication management
type AuthManager struct {
	backup    *AuthBackupManager
	protector *ConfigProtector
}

// NewAuthManager creates integrated authentication manager
func NewAuthManager() (*AuthManager, error) {
	backup, err := NewAuthBackupManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create backup manager: %w", err)
	}

	protector, err := NewConfigProtector()
	if err != nil {
		return nil, fmt.Errorf("failed to create config protector: %w", err)
	}

	return &AuthManager{
		backup:    backup,
		protector: protector,
	}, nil
}

// ProtectAndBackup protects and backs up authentication
func (am *AuthManager) ProtectAndBackup() error {
	log.Info().Msg("🔒 Starting authentication protection and backup")

	// Protect existing configuration
	if err := am.protector.ProtectExistingConfig(); err != nil {
		return fmt.Errorf("failed to protect existing config: %w", err)
	}

	// Backup IDE authentication info
	if err := am.backup.BackupIDEAuth(); err != nil {
		return fmt.Errorf("failed to backup IDE auth: %w", err)
	}

	log.Info().Msg("✅ Authentication protection and backup completed")
	return nil
}

// RestoreAndCleanup restores authentication and cleans up
func (am *AuthManager) RestoreAndCleanup() error {
	log.Info().Msg("🔄 Starting authentication restore and cleanup")

	// Restore IDE authentication info
	if err := am.backup.RestoreIDEAuth(); err != nil {
		log.Error().Err(err).Msg("❌ IDE authentication restore failed")
		// Treat restore failure as warning and continue
	}

	// Cleanup backup
	if err := am.backup.CleanupBackup(); err != nil {
		log.Error().Err(err).Msg("❌ Backup cleanup failed")
		// Treat cleanup failure as warning and continue
	}

	// Unlock configuration protection
	if err := am.protector.UnlockConfig(); err != nil {
		log.Error().Err(err).Msg("❌ Configuration unlock failed")
		// Treat unlock failure as warning and continue
	}

	log.Info().Msg("✅ Authentication restore and cleanup completed")
	return nil
}
