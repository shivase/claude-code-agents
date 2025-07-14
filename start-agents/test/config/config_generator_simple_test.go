package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConfigGenerator - ConfigGenerator作成の簡単なテスト
func TestNewConfigGenerator(t *testing.T) {
	generator := config.NewConfigGenerator()
	assert.NotNil(t, generator)
}

// TestGenerateConfigBasic - 基本的な設定ファイル生成テスト
func TestGenerateConfigBasic(t *testing.T) {
	// 一時ディレクトリを作成
	tempHome, err := os.MkdirTemp("", "config_gen_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// HOME環境変数を一時的に変更
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	generator := config.NewConfigGenerator()

	templateContent := `# Test Configuration
CLAUDE_CLI_PATH=~/.claude/local/claude
INSTRUCTIONS_DIR=~/.claude/claude-code-agents/instructions
SESSION_NAME=test-session
DEV_COUNT=2`

	err = generator.GenerateConfig(templateContent)
	assert.NoError(t, err)

	// 生成されたファイルを確認
	configPath := filepath.Join(tempHome, ".claude", "claude-code-agents", "agents.conf")
	assert.FileExists(t, configPath)

	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test Configuration")
	assert.Contains(t, string(content), "SESSION_NAME=test-session")
}

// TestForceGenerateConfig - 強制上書き機能のテスト
func TestForceGenerateConfig(t *testing.T) {
	// 一時ディレクトリを作成
	tempHome, err := os.MkdirTemp("", "config_force_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// HOME環境変数を一時的に変更
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// 設定ディレクトリを作成
	configDir := filepath.Join(tempHome, ".claude", "claude-code-agents")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// 既存ファイルを作成
	configPath := filepath.Join(configDir, "agents.conf")
	err = os.WriteFile(configPath, []byte("# Existing config"), 0644)
	require.NoError(t, err)

	generator := config.NewConfigGenerator()

	templateContent := `# New Configuration
CLAUDE_CLI_PATH=~/.claude/local/claude
SESSION_NAME=new-session`

	// 強制上書きでファイル生成
	err = generator.ForceGenerateConfig(templateContent)
	assert.NoError(t, err)

	// 新しい内容が書き込まれていることを確認
	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "New Configuration")
	assert.Contains(t, string(content), "SESSION_NAME=new-session")
	assert.NotContains(t, string(content), "Existing config")
}

// TestGenerateConfigWithExistingFile - 既存ファイルがある場合のテスト
func TestGenerateConfigWithExistingFile(t *testing.T) {
	// 一時ディレクトリを作成
	tempHome, err := os.MkdirTemp("", "config_existing_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// HOME環境変数を一時的に変更
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// 設定ディレクトリを作成
	configDir := filepath.Join(tempHome, ".claude", "claude-code-agents")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// 既存ファイルを作成
	configPath := filepath.Join(configDir, "agents.conf")
	originalContent := "# Original config"
	err = os.WriteFile(configPath, []byte(originalContent), 0644)
	require.NoError(t, err)

	generator := config.NewConfigGenerator()

	templateContent := `# New Configuration
CLAUDE_CLI_PATH=~/.claude/local/claude`

	// 通常の生成（既存ファイルがあるとエラーになるはず）
	err = generator.GenerateConfig(templateContent)
	assert.Error(t, err)

	// 元のファイルが保持されていることを確認
	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Original config")
}
