package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ConfigGenerator generates configuration files
type ConfigGenerator struct {
	targetPath string
	backupPath string
}

// NewConfigGenerator creates a new ConfigGenerator instance
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// GenerateConfig generates configuration file
func (cg *ConfigGenerator) GenerateConfig(templateContent string) error {
	// Set target path
	if err := cg.setTargetPath(); err != nil {
		return fmt.Errorf("failed to set target path: %w", err)
	}

	// Ensure directory exists
	if err := cg.ensureDirectory(); err != nil {
		return fmt.Errorf("failed to ensure directory: %w", err)
	}

	// Check existing file
	if err := cg.checkExistingFile(); err != nil {
		return fmt.Errorf("existing file check failed: %w", err)
	}

	// Generate configuration file
	if err := cg.writeConfigFile(templateContent); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Display success message
	cg.displaySuccessMessage()

	return nil
}

// setTargetPath sets the target path
func (cg *ConfigGenerator) setTargetPath() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Unified configuration directory path
	configDir := filepath.Join(homeDir, ".claude", "claude-code-agents")
	cg.targetPath = filepath.Join(configDir, "agents.conf")
	cg.backupPath = filepath.Join(configDir, fmt.Sprintf("agents.conf.backup.%d", time.Now().Unix()))

	log.Debug().
		Str("target_path", cg.targetPath).
		Str("backup_path", cg.backupPath).
		Msg("Target path set")

	return nil
}

// ensureDirectory ensures directory exists and creates if necessary
func (cg *ConfigGenerator) ensureDirectory() error {
	dir := filepath.Dir(cg.targetPath)

	// Create directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Info().Str("directory", dir).Msg("Directory created")
	}

	// Create necessary subdirectories
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

// checkExistingFile checks for existing file and prevents overwrite
func (cg *ConfigGenerator) checkExistingFile() error {
	if _, err := os.Stat(cg.targetPath); err == nil {
		return fmt.Errorf("config file already exists at %s. Use --force to overwrite or manually remove the existing file", cg.targetPath)
	}

	return nil
}

// writeConfigFile writes configuration file
func (cg *ConfigGenerator) writeConfigFile(content string) error {
	// Safe file operation: create temporary file then move
	tempFile := cg.targetPath + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tempFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Move temporary file to final location
	if err := os.Rename(tempFile, cg.targetPath); err != nil {
		// Remove temporary file if move failed
		if err := os.Remove(tempFile); err != nil {
			log.Warn().Err(err).Str("file", tempFile).Msg("Failed to remove temporary file")
		}
		return fmt.Errorf("failed to move temporary file to final location: %w", err)
	}

	log.Info().Str("file", cg.targetPath).Msg("Config file generated successfully")
	return nil
}

// displaySuccessMessage displays success message
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

// ForceGenerateConfig forcefully generates configuration file (overwrite enabled)
func (cg *ConfigGenerator) ForceGenerateConfig(templateContent string) error {
	// Set target path
	if err := cg.setTargetPath(); err != nil {
		return fmt.Errorf("failed to set target path: %w", err)
	}

	// Ensure directory exists
	if err := cg.ensureDirectory(); err != nil {
		return fmt.Errorf("failed to ensure directory: %w", err)
	}

	// Backup existing file
	if err := cg.backupExistingFile(); err != nil {
		return fmt.Errorf("failed to backup existing file: %w", err)
	}

	// Generate configuration file
	if err := cg.writeConfigFile(templateContent); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Display success message
	cg.displayForceSuccessMessage()

	return nil
}

// backupExistingFile backs up existing file
func (cg *ConfigGenerator) backupExistingFile() error {
	if _, err := os.Stat(cg.targetPath); err == nil {
		// Backup existing file
		if err := os.Rename(cg.targetPath, cg.backupPath); err != nil {
			return fmt.Errorf("failed to backup existing file: %w", err)
		}
		log.Info().Str("backup", cg.backupPath).Msg("Existing file backed up")
	}

	return nil
}

// displayForceSuccessMessage displays force generation success message
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

// fileExists checks if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ValidateConfigDirectory validates configuration directory
func (cg *ConfigGenerator) ValidateConfigDirectory() error {
	if cg.targetPath == "" {
		if err := cg.setTargetPath(); err != nil {
			return fmt.Errorf("failed to set target path: %w", err)
		}
	}

	dir := filepath.Dir(cg.targetPath)

	// Check directory existence
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("config directory does not exist: %s", dir)
	}

	// Check write permission
	testFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("no write permission in config directory: %s", dir)
	}
	if err := os.Remove(testFile); err != nil {
		log.Warn().Err(err).Str("file", testFile).Msg("Failed to remove test file")
	}

	return nil
}

// GetConfigInfo gets configuration information
func (cg *ConfigGenerator) GetConfigInfo() (string, bool, error) {
	if err := cg.setTargetPath(); err != nil {
		return "", false, fmt.Errorf("failed to set target path: %w", err)
	}

	exists := fileExists(cg.targetPath)
	return cg.targetPath, exists, nil
}
