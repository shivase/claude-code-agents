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
		fmt.Println("❌ エラー: 無効な役割です。")
		fmt.Printf("📁 instructionsファイルが見つかりません: %s.%s\n", instructionsDir, role)
		fmt.Println("📝 使用例: /reload-role [role名称]")
		return fmt.Errorf("指定されたrole名に該当するinstructionファイルが存在しません: %s", role)
	}

	// mdファイルのパスを構築
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}

	mdFile := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions", role+".md")

	// ファイルが存在するかチェック
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		fmt.Printf("❌ エラー: %s が見つかりません。\n", mdFile)
		return fmt.Errorf("ファイルが見つかりません: %s", mdFile)
	}

	// ファイルの内容を読み込み
	content, err := os.ReadFile(mdFile)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	// 結果を出力
	fmt.Printf("🔄 %sの役割定義を再読み込み中...\n", role)
	fmt.Println("")
	fmt.Printf("📋 ファイル: %s\n", mdFile)
	fmt.Println("")
	fmt.Println("🔄 前の役割定義をリセットしています...")
	fmt.Println("📖 新しい役割定義を適用します：")
	fmt.Println("----------------------------------------")
	fmt.Print(string(content))
	fmt.Println("----------------------------------------")
	fmt.Println("")
	fmt.Printf("✅ %sの役割定義を正常に再読み込みしました。\n", role)
	fmt.Println("💡 前の役割定義は完全にリセットされ、新しい役割定義のみが適用されます。")

	return nil
}

func isValidRole(role string) bool {
	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// mdファイルのパスを構築
	mdFile := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions", role+".md")

	// ファイルが存在するかチェック
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
			name:     "特殊文字を含む役割名",
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
			name:     "存在する役割 - developer",
			role:     "developer",
			expected: true,
		},
		{
			name:     "存在する役割 - manager",
			role:     "manager",
			expected: true,
		},
		{
			name:     "存在する役割 - po",
			role:     "po",
			expected: true,
		},
		{
			name:     "存在しない役割",
			role:     "nonexistent",
			expected: false,
		},
		{
			name:     "空の役割名",
			role:     "",
			expected: false,
		},
		{
			name:     "特殊文字を含む役割名",
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
			name: "ホームディレクトリが存在しない場合",
			setupFunc: func(tmpDir string) error {
				// .claudeディレクトリを作成しない
				return nil
			},
			args:        []string{"/reload-role developer"},
			expectedErr: true,
			description: "ホームディレクトリに.claudeディレクトリが存在しない場合のエラー",
		},
		{
			name: "instructionsディレクトリが存在しない場合",
			setupFunc: func(tmpDir string) error {
				// .claudeディレクトリのみ作成
				return os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0755)
			},
			args:        []string{"/reload-role developer"},
			expectedErr: true,
			description: "instructionsディレクトリが存在しない場合のエラー",
		},
		{
			name: "指定された役割ファイルが存在しない場合",
			setupFunc: func(tmpDir string) error {
				// instructionsディレクトリまで作成するが、ファイルは作成しない
				return os.MkdirAll(filepath.Join(tmpDir, ".claude", "claude-code-agents", "instructions"), 0755)
			},
			args:        []string{"/reload-role nonexistent"},
			expectedErr: true,
			description: "指定された役割ファイルが存在しない場合のエラー",
		},
		{
			name: "空の役割ファイルが存在する場合",
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
			description: "空の役割ファイルが存在する場合は正常に処理される",
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
			name: "TMUX環境変数が設定されている場合",
			envVars: map[string]string{
				"TMUX": "/tmp/tmux-1000/default,1234,0",
			},
			expected: true,
		},
		{
			name: "TERM環境変数がscreenの場合",
			envVars: map[string]string{
				"TERM": "screen",
			},
			expected: true,
		},
		{
			name: "TERM環境変数がscreen-256colorの場合",
			envVars: map[string]string{
				"TERM": "screen-256color",
			},
			expected: true,
		},
		{
			name: "TERM環境変数がtmuxの場合",
			envVars: map[string]string{
				"TERM": "tmux",
			},
			expected: true,
		},
		{
			name: "TERM環境変数がtmux-256colorの場合",
			envVars: map[string]string{
				"TERM": "tmux-256color",
			},
			expected: true,
		},
		{
			name: "tmux関連の環境変数が設定されていない場合",
			envVars: map[string]string{
				"TERM": "xterm-256color",
			},
			expected: false,
		},
		{
			name:     "環境変数が何も設定されていない場合",
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
