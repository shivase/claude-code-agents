package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// AuthBackupManager - 認証バックアップ管理
type AuthBackupManager struct {
	homeDir   string
	claudeDir string
	backupDir string
	retention time.Duration
}

// NewAuthBackupManager - 認証バックアップマネージャーの作成
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

// BackupIDEAuth - IDE連携認証情報のバックアップ
func (abm *AuthBackupManager) BackupIDEAuth() error {
	log.Info().Str("backup_dir", abm.backupDir).Msg("Starting IDE auth backup")

	// バックアップディレクトリを作成
	if err := os.MkdirAll(abm.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// IDE連携ディレクトリの存在確認
	ideDir := filepath.Join(abm.claudeDir, "ide")
	if _, err := os.Stat(ideDir); os.IsNotExist(err) {
		log.Info().Msg("No IDE directory found, skipping backup")
		return nil
	}

	// IDE連携情報をバックアップ
	backupIdeDir := filepath.Join(abm.backupDir, "ide")
	if err := abm.copyDir(ideDir, backupIdeDir); err != nil {
		return fmt.Errorf("failed to backup IDE directory: %w", err)
	}

	log.Info().Msg("IDE auth backup completed successfully")
	return nil
}

// RestoreIDEAuth - IDE連携認証情報の復元
func (abm *AuthBackupManager) RestoreIDEAuth() error {
	log.Info().Str("backup_dir", abm.backupDir).Msg("Starting IDE auth restore")

	// バックアップディレクトリの存在確認
	backupIdeDir := filepath.Join(abm.backupDir, "ide")
	if _, err := os.Stat(backupIdeDir); os.IsNotExist(err) {
		log.Info().Msg("No IDE backup found, skipping restore")
		return nil
	}

	// IDE連携情報を復元
	ideDir := filepath.Join(abm.claudeDir, "ide")
	if err := abm.copyDir(backupIdeDir, ideDir); err != nil {
		return fmt.Errorf("failed to restore IDE directory: %w", err)
	}

	log.Info().Msg("IDE auth restore completed successfully")
	return nil
}

// CleanupBackup - バックアップディレクトリの削除
func (abm *AuthBackupManager) CleanupBackup() error {
	if _, err := os.Stat(abm.backupDir); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(abm.backupDir); err != nil {
		return fmt.Errorf("failed to cleanup backup directory: %w", err)
	}

	log.Info().Str("backup_dir", abm.backupDir).Msg("Backup cleanup completed")
	return nil
}

// copyDir - ディレクトリの再帰的コピー
func (abm *AuthBackupManager) copyDir(src, dst string) error {
	// ソースディレクトリの情報を取得
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// 目的ディレクトリを作成
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// ディレクトリ内のファイルを読み取り
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// 各エントリを処理
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// ディレクトリの場合は再帰的にコピー
			if err := abm.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// ファイルの場合はコピー
			if err := abm.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile - ファイルコピー
func (abm *AuthBackupManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// ConfigProtector - 設定保護システム
type ConfigProtector struct {
	claudeDir string
	lockFile  string
}

// NewConfigProtector - 設定保護システムの作成
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

// ProtectExistingConfig - 既存設定の保護
func (cp *ConfigProtector) ProtectExistingConfig() error {
	settingsFile := filepath.Join(cp.claudeDir, "settings.json")

	// 既存設定の存在確認
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		log.Info().Msg("No existing Claude config found")
		return nil
	}

	// 設定ファイルの内容確認
	if err := cp.ValidateConfig(); err != nil {
		return fmt.Errorf("existing config validation failed: %w", err)
	}

	// ロックファイルの作成
	if err := os.WriteFile(cp.lockFile, []byte("protected"), 0644); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	log.Info().Msg("Existing Claude config protected")
	return nil
}

// ValidateConfig - 設定ファイルの検証
func (cp *ConfigProtector) ValidateConfig() error {
	settingsFile := filepath.Join(cp.claudeDir, "settings.json")

	// ファイルの存在確認
	info, err := os.Stat(settingsFile)
	if err != nil {
		return fmt.Errorf("config file not found: %w", err)
	}

	// ファイルサイズの確認
	if info.Size() == 0 {
		return fmt.Errorf("config file is empty")
	}

	// ファイルの読み取り可能性確認
	if _, err := os.ReadFile(settingsFile); err != nil {
		return fmt.Errorf("config file is not readable: %w", err)
	}

	log.Info().Msg("Config file validation passed")
	return nil
}

// IsConfigProtected - 設定保護状態の確認
func (cp *ConfigProtector) IsConfigProtected() bool {
	_, err := os.Stat(cp.lockFile)
	return err == nil
}

// UnlockConfig - 設定保護の解除
func (cp *ConfigProtector) UnlockConfig() error {
	if !cp.IsConfigProtected() {
		return nil
	}

	if err := os.Remove(cp.lockFile); err != nil {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	log.Info().Msg("Config protection unlocked")
	return nil
}

// AuthManager - 統合認証マネージャー
type AuthManager struct {
	backup    *AuthBackupManager
	protector *ConfigProtector
}

// NewAuthManager - 統合認証マネージャーの作成
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

// ProtectAndBackup - 認証保護とバックアップ
func (am *AuthManager) ProtectAndBackup() error {
	log.Info().Msg("Starting auth protection and backup")

	// 既存設定の保護
	if err := am.protector.ProtectExistingConfig(); err != nil {
		return fmt.Errorf("failed to protect existing config: %w", err)
	}

	// IDE認証情報のバックアップ
	if err := am.backup.BackupIDEAuth(); err != nil {
		return fmt.Errorf("failed to backup IDE auth: %w", err)
	}

	log.Info().Msg("Auth protection and backup completed")
	return nil
}

// RestoreAndCleanup - 認証復元とクリーンアップ
func (am *AuthManager) RestoreAndCleanup() error {
	log.Info().Msg("Starting auth restore and cleanup")

	// IDE認証情報の復元
	if err := am.backup.RestoreIDEAuth(); err != nil {
		log.Error().Err(err).Msg("Failed to restore IDE auth")
		// 復元失敗は警告として扱い、処理を続行
	}

	// バックアップのクリーンアップ
	if err := am.backup.CleanupBackup(); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup backup")
		// クリーンアップ失敗は警告として扱い、処理を続行
	}

	// 設定保護の解除
	if err := am.protector.UnlockConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to unlock config")
		// アンロック失敗は警告として扱い、処理を続行
	}

	log.Info().Msg("Auth restore and cleanup completed")
	return nil
}
