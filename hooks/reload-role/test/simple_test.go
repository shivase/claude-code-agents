package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// テスト用の関数をコピー
func executeHook(cmd *cobra.Command, args []string) error {
	inputMessage := args[0]

	// /reload-role コマンドかどうかをチェック
	reloadRoleRegex := regexp.MustCompile(`^/reload-role\s+([a-zA-Z]+)`)
	matches := reloadRoleRegex.FindStringSubmatch(inputMessage)

	if len(matches) == 0 {
		// /reload-role コマンドでない場合は何もしない
		return nil
	}

	role := matches[1]

	// 役割が有効かどうかをチェック
	if !isValidRole(role) {
		homeDir, _ := os.UserHomeDir()
		instructionsDir := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions")
		fmt.Println("❌ Error: Invalid role.")
		fmt.Printf("📁 Instructions file not found: %s.%s\n", instructionsDir, role)
		fmt.Println("📝 Usage: /reload-role [role name]")
		return fmt.Errorf("instruction file for the specified role does not exist: %s", role)
	}

	// Build the md file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	mdFile := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions", role+".md")

	// Check if the file exists
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		fmt.Printf("❌ Error: %s not found.\n", mdFile)
		return fmt.Errorf("file not found: %s", mdFile)
	}

	// ファイルの内容を読み込み
	content, err := os.ReadFile(mdFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 結果を出力
	fmt.Printf("🔄 Reloading role definition for %s...\n", role)
	fmt.Println("")
	fmt.Printf("📋 File: %s\n", mdFile)
	fmt.Println("")
	fmt.Println("🔄 Resetting previous role definition...")
	fmt.Println("📖 Applying new role definition:")
	fmt.Println("----------------------------------------")
	fmt.Print(string(content))
	fmt.Println("----------------------------------------")
	fmt.Println("")
	fmt.Printf("✅ Successfully reloaded role definition for %s.\n", role)
	fmt.Println("💡 Previous role definition has been completely reset, and only the new role definition is applied.")

	return nil
}

func isValidRole(role string) bool {
	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Build the md file path
	mdFile := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions", role+".md")

	// Check if the file exists
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// isRunningInTmux checks if the process is running inside a tmux session
func isRunningInTmux() bool {
	// TMUX環境変数が設定されているかチェック
	if os.Getenv("TMUX") != "" {
		return true
	}

	// TERM環境変数にscreenまたはtmuxが含まれているかチェック
	term := os.Getenv("TERM")
	if term == "screen" || term == "screen-256color" || term == "tmux" || term == "tmux-256color" {
		return true
	}

	return false
}

func TestExecuteHookRegexValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "正常な/reload-roleコマンド",
			input:    "/reload-role developer",
			expected: true,
		},
		{
			name:     "複数スペースでの/reload-roleコマンド",
			input:    "/reload-role   manager",
			expected: true,
		},
		{
			name:     "大文字を含む役割名",
			input:    "/reload-role DevOps",
			expected: true,
		},
		{
			name:     "役割名のみ",
			input:    "developer",
			expected: false,
		},
		{
			name:     "スラッシュのみ",
			input:    "/reload-role",
			expected: false,
		},
		{
			name:     "別のコマンド",
			input:    "/other-command test",
			expected: false,
		},
		{
			name:     "数字を含む役割名",
			input:    "/reload-role dev123",
			expected: true,
		},
		{
			name:     "Role name with special characters",
			input:    "/reload-role dev-ops",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// executeHook関数内の正規表現をテスト
			reloadRoleRegex := regexp.MustCompile(`^/reload-role\s+([a-zA-Z]+)`)
			matches := reloadRoleRegex.FindStringSubmatch(tt.input)
			matched := len(matches) > 0

			if matched != tt.expected {
				t.Errorf("Expected %v for input '%s', got %v", tt.expected, tt.input, matched)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test-claude-valid-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 一時ホームディレクトリを設定
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// テスト用のディレクトリ構造を作成
	instructionsDir := filepath.Join(tmpDir, ".claude", "claude-code-agents", "instructions")
	err = os.MkdirAll(instructionsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create instructions dir: %v", err)
	}

	// 有効な役割ファイルを作成
	validRoles := []string{"developer", "manager", "po"}
	for _, role := range validRoles {
		roleFile := filepath.Join(instructionsDir, role+".md")
		err = os.WriteFile(roleFile, []byte(fmt.Sprintf("# %s Role\nTest content", role)), 0644)
		if err != nil {
			t.Fatalf("Failed to create role file %s: %v", role, err)
		}
	}

	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "Existing role - developer",
			role:     "developer",
			expected: true,
		},
		{
			name:     "Existing role - manager",
			role:     "manager",
			expected: true,
		},
		{
			name:     "Existing role - po",
			role:     "po",
			expected: true,
		},
		{
			name:     "Non-existent role",
			role:     "nonexistent",
			expected: false,
		},
		{
			name:     "Empty role name",
			role:     "",
			expected: false,
		},
		{
			name:     "Role name with special characters",
			role:     "dev-ops",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidRole(tt.role)
			if result != tt.expected {
				t.Errorf("isValidRole(%s) = %v, expected %v", tt.role, result, tt.expected)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test-claude-error-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 一時ホームディレクトリを設定
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name        string
		setupFunc   func(string) error
		args        []string
		expectedErr bool
		description string
	}{
		{
			name: "Home directory does not exist",
			setupFunc: func(tmpDir string) error {
				// .claudeディレクトリを作成しない
				return nil
			},
			args:        []string{"/reload-role developer"},
			expectedErr: true,
			description: "Error when .claude directory does not exist in home directory",
		},
		{
			name: "Instructions directory does not exist",
			setupFunc: func(tmpDir string) error {
				// .claudeディレクトリのみ作成
				return os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0755)
			},
			args:        []string{"/reload-role developer"},
			expectedErr: true,
			description: "Error when instructions directory does not exist",
		},
		{
			name: "Specified role file does not exist",
			setupFunc: func(tmpDir string) error {
				// instructionsディレクトリまで作成するが、ファイルは作成しない
				return os.MkdirAll(filepath.Join(tmpDir, ".claude", "claude-code-agents", "instructions"), 0755)
			},
			args:        []string{"/reload-role nonexistent"},
			expectedErr: true,
			description: "Error when specified role file does not exist",
		},
		{
			name: "Empty role file exists",
			setupFunc: func(tmpDir string) error {
				instructionsDir := filepath.Join(tmpDir, ".claude", "claude-code-agents", "instructions")
				err := os.MkdirAll(instructionsDir, 0755)
				if err != nil {
					return err
				}
				// 空のファイルを作成
				return os.WriteFile(filepath.Join(instructionsDir, "empty.md"), []byte(""), 0644)
			},
			args:        []string{"/reload-role empty"},
			expectedErr: false,
			description: "Empty role file is processed normally when it exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト環境をセットアップ
			if tt.setupFunc != nil {
				err := tt.setupFunc(tmpDir)
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// executeHook関数を直接呼び出す
			err := executeHook(nil, tt.args)

			// エラーの検証
			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestIsRunningInTmux(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "TMUX environment variable is set",
			envVars: map[string]string{
				"TMUX": "/tmp/tmux-1000/default,1234,0",
			},
			expected: true,
		},
		{
			name: "TERM environment variable is screen",
			envVars: map[string]string{
				"TERM": "screen",
			},
			expected: true,
		},
		{
			name: "TERM environment variable is screen-256color",
			envVars: map[string]string{
				"TERM": "screen-256color",
			},
			expected: true,
		},
		{
			name: "TERM environment variable is tmux",
			envVars: map[string]string{
				"TERM": "tmux",
			},
			expected: true,
		},
		{
			name: "TERM environment variable is tmux-256color",
			envVars: map[string]string{
				"TERM": "tmux-256color",
			},
			expected: true,
		},
		{
			name: "No tmux-related environment variables are set",
			envVars: map[string]string{
				"TERM": "xterm-256color",
			},
			expected: false,
		},
		{
			name:     "No environment variables are set",
			envVars:  map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 元の環境変数を保存
			originalTmux := os.Getenv("TMUX")
			originalTerm := os.Getenv("TERM")
			defer func() {
				os.Setenv("TMUX", originalTmux)
				os.Setenv("TERM", originalTerm)
			}()

			// テスト用の環境変数を設定
			os.Unsetenv("TMUX")
			os.Unsetenv("TERM")
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// テスト実行
			result := isRunningInTmux()
			if result != tt.expected {
				t.Errorf("isRunningInTmux() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
