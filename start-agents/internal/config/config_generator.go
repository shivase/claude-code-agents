package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ConfigGenerator 設定ファイル生成機能
type ConfigGenerator struct {
	targetPath string
	backupPath string
}

// NewConfigGenerator 新しいConfigGeneratorインスタンスを作成
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// GenerateConfig 設定ファイルを生成
func (cg *ConfigGenerator) GenerateConfig(templateContent string) error {
	// ターゲットパスの設定
	if err := cg.setTargetPath(); err != nil {
		return fmt.Errorf("failed to set target path: %w", err)
	}

	// ディレクトリの自動作成
	if err := cg.ensureDirectory(); err != nil {
		return fmt.Errorf("failed to ensure directory: %w", err)
	}

	// 既存ファイルのチェック
	if err := cg.checkExistingFile(); err != nil {
		return fmt.Errorf("existing file check failed: %w", err)
	}

	// 設定ファイルの生成
	if err := cg.writeConfigFile(templateContent); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 成功メッセージの表示
	cg.displaySuccessMessage()

	return nil
}

// setTargetPath ターゲットパスの設定
func (cg *ConfigGenerator) setTargetPath() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// 統一された設定ディレクトリパス
	configDir := filepath.Join(homeDir, ".claude", "claude-code-agents")
	cg.targetPath = filepath.Join(configDir, "agents.conf")
	cg.backupPath = filepath.Join(configDir, fmt.Sprintf("agents.conf.backup.%d", time.Now().Unix()))

	log.Debug().
		Str("target_path", cg.targetPath).
		Str("backup_path", cg.backupPath).
		Msg("Target path set")

	return nil
}

// ensureDirectory ディレクトリの自動作成
func (cg *ConfigGenerator) ensureDirectory() error {
	dir := filepath.Dir(cg.targetPath)

	// ディレクトリが存在しない場合は作成
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Info().Str("directory", dir).Msg("Directory created")
	}

	// 必要な子ディレクトリも作成
	subdirs := []string{
		"logs",
		"instructions",
		"auth_backup",
	}

	for _, subdir := range subdirs {
		subdirPath := filepath.Join(dir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(subdirPath, 0750); err != nil {
				log.Warn().Str("subdirectory", subdirPath).Err(err).Msg("Failed to create subdirectory")
			} else {
				log.Info().Str("subdirectory", subdirPath).Msg("Subdirectory created")
			}
		}
	}

	return nil
}

// checkExistingFile 既存ファイルの確認と上書き防止
func (cg *ConfigGenerator) checkExistingFile() error {
	if _, err := os.Stat(cg.targetPath); err == nil {
		return fmt.Errorf("config file already exists at %s. Use --force to overwrite or manually remove the existing file", cg.targetPath)
	}

	return nil
}

// writeConfigFile 設定ファイルの書き込み
func (cg *ConfigGenerator) writeConfigFile(content string) error {
	// 安全なファイル操作：一時ファイルを作成してから移動
	tempFile := cg.targetPath + ".tmp"

	// 一時ファイルに書き込み
	if err := os.WriteFile(tempFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// 一時ファイルを最終的な場所に移動
	if err := os.Rename(tempFile, cg.targetPath); err != nil {
		// 移動に失敗した場合は一時ファイルを削除
		if err := os.Remove(tempFile); err != nil {
			log.Warn().Err(err).Str("file", tempFile).Msg("Failed to remove temporary file")
		}
		return fmt.Errorf("failed to move temporary file to final location: %w", err)
	}

	log.Info().Str("file", cg.targetPath).Msg("Config file generated successfully")
	return nil
}

// displaySuccessMessage 成功メッセージの表示
func (cg *ConfigGenerator) displaySuccessMessage() {
	fmt.Println("🎉 Configuration file generated successfully!")
	fmt.Println("=" + strings.Repeat("=", 45))
	fmt.Printf("📁 Location: %s\n", cg.targetPath)
	fmt.Printf("📝 Content: AI Teams configuration template\n")
	fmt.Printf("🔧 Usage: Customize the settings as needed\n")
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Edit the configuration file to match your environment")
	fmt.Println("   2. Run --show-config to verify your settings")
	fmt.Println("   3. Start the AI Teams system with your custom configuration")
	fmt.Println()
}

// ForceGenerateConfig 強制的に設定ファイルを生成（上書き可能）
func (cg *ConfigGenerator) ForceGenerateConfig(templateContent string) error {
	// ターゲットパスの設定
	if err := cg.setTargetPath(); err != nil {
		return fmt.Errorf("failed to set target path: %w", err)
	}

	// ディレクトリの自動作成
	if err := cg.ensureDirectory(); err != nil {
		return fmt.Errorf("failed to ensure directory: %w", err)
	}

	// 既存ファイルのバックアップ
	if err := cg.backupExistingFile(); err != nil {
		return fmt.Errorf("failed to backup existing file: %w", err)
	}

	// 設定ファイルの生成
	if err := cg.writeConfigFile(templateContent); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 成功メッセージの表示
	cg.displayForceSuccessMessage()

	return nil
}

// backupExistingFile 既存ファイルのバックアップ
func (cg *ConfigGenerator) backupExistingFile() error {
	if _, err := os.Stat(cg.targetPath); err == nil {
		// 既存ファイルをバックアップ
		if err := os.Rename(cg.targetPath, cg.backupPath); err != nil {
			return fmt.Errorf("failed to backup existing file: %w", err)
		}
		log.Info().Str("backup", cg.backupPath).Msg("Existing file backed up")
	}

	return nil
}

// displayForceSuccessMessage 強制生成成功メッセージの表示
func (cg *ConfigGenerator) displayForceSuccessMessage() {
	fmt.Println("🎉 Configuration file generated successfully! (Force mode)")
	fmt.Println("=" + strings.Repeat("=", 55))
	fmt.Printf("📁 Location: %s\n", cg.targetPath)
	fmt.Printf("📝 Content: AI Teams configuration template\n")
	fmt.Printf("🔧 Usage: Customize the settings as needed\n")
	if fileExists(cg.backupPath) {
		fmt.Printf("💾 Backup: %s\n", cg.backupPath)
	}
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Edit the configuration file to match your environment")
	fmt.Println("   2. Run --show-config to verify your settings")
	fmt.Println("   3. Start the AI Teams system with your custom configuration")
	fmt.Println()
}

// fileExists ファイルの存在確認
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ValidateConfigDirectory 設定ディレクトリの検証
func (cg *ConfigGenerator) ValidateConfigDirectory() error {
	if cg.targetPath == "" {
		if err := cg.setTargetPath(); err != nil {
			return fmt.Errorf("failed to set target path: %w", err)
		}
	}

	dir := filepath.Dir(cg.targetPath)

	// ディレクトリの存在確認
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("config directory does not exist: %s", dir)
	}

	// 書き込み権限の確認
	testFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("no write permission in config directory: %s", dir)
	}
	if err := os.Remove(testFile); err != nil {
		log.Warn().Err(err).Str("file", testFile).Msg("Failed to remove test file")
	}

	return nil
}

// GetConfigInfo 設定情報の取得
func (cg *ConfigGenerator) GetConfigInfo() (string, bool, error) {
	if err := cg.setTargetPath(); err != nil {
		return "", false, fmt.Errorf("failed to set target path: %w", err)
	}

	exists := fileExists(cg.targetPath)
	return cg.targetPath, exists, nil
}
