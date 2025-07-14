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

// ValidatePathsDetailed パス検証機能の詳細確認
func ValidatePathsDetailed() []string {
	var errors []string

	// Claude CLI実行可能ファイルの検証
	claudePath := findClaudeExecutableHelper()
	if claudePath == "" {
		errors = append(errors, "Claude CLI実行可能ファイルが見つかりません")
	} else {
		fmt.Printf("   ✅ Claude CLI: %s\n", claudePath)
	}

	// ディレクトリ解決器を使用したパス検証
	dirResolver := utils.GetGlobalDirectoryResolver()
	dirInfo := dirResolver.GetDirectoryInfo()

	fmt.Printf("   📂 プロジェクトルート: %s\n", dirInfo["project_root"])
	fmt.Printf("   📂 作業ディレクトリ: %s\n", dirInfo["original_working_dir"])

	// 必要ディレクトリの確認
	homeDir, _ := os.UserHomeDir()
	requiredDirs := []string{
		filepath.Join(homeDir, ".claude"),
		filepath.Join(homeDir, ".claude", "claude-code-agents"),
		filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"),
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); err != nil {
			errors = append(errors, fmt.Sprintf("必要ディレクトリが見つかりません: %s", dir))
		} else {
			fmt.Printf("   ✅ ディレクトリ確認: %s\n", dir)
		}
	}

	return errors
}

// ValidateConfigurationDetailed 設定ファイル確認機能の詳細確認
func ValidateConfigurationDetailed() []string {
	var errors []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		errors = append(errors, "ホームディレクトリの取得に失敗")
		return errors
	}

	// ~/.claude/settings.json の確認
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		errors = append(errors, "Claude設定ファイル(settings.json)が見つかりません")
	} else {
		// ファイルサイズと内容の基本チェック
		info, err := os.Stat(settingsPath)
		if err != nil || info.Size() == 0 {
			errors = append(errors, "Claude設定ファイル(settings.json)が空です")
		} else {
			fmt.Printf("   ✅ settings.json: %s (%d bytes)\n", settingsPath, info.Size())
		}
	}

	// ~/.claude/claude.json の確認
	claudeJsonPath := filepath.Join(homeDir, ".claude", "claude.json")
	if _, err := os.Stat(claudeJsonPath); err != nil {
		errors = append(errors, "Claude認証ファイル(claude.json)が見つかりません")
	} else {
		info, err := os.Stat(claudeJsonPath)
		if err != nil || info.Size() == 0 {
			errors = append(errors, "Claude認証ファイル(claude.json)が空です")
		} else {
			fmt.Printf("   ✅ claude.json: %s (%d bytes)\n", claudeJsonPath, info.Size())
		}
	}

	return errors
}

// ValidateAuthenticationDetailed Claude認証チェック機能の詳細確認
func ValidateAuthenticationDetailed() []string {
	var warnings []string

	// 認証マネージャーを使用した認証確認
	authManager := auth.NewClaudeAuthManager()

	// 設定ファイル確認
	if err := authManager.CheckSettingsFile(); err != nil {
		warnings = append(warnings, fmt.Sprintf("設定ファイル確認失敗: %v", err))
	} else {
		fmt.Printf("   ✅ 設定ファイル確認完了\n")
	}

	// 認証状態確認
	authStatus, err := authManager.CheckAuthenticationStatus()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("認証状態確認失敗: %v", err))
	} else {
		if authStatus.IsAuthenticated {
			if authStatus.UserID != "" {
				fmt.Printf("   ✅ 認証済み (UserID: %s...)\n", authStatus.UserID[:8])
			}
			if authStatus.OAuthAccount != nil {
				if email, exists := authStatus.OAuthAccount["emailAddress"]; exists {
					fmt.Printf("   ✅ OAuth認証済み: %v\n", email)
				}
			}
		} else {
			warnings = append(warnings, "Claude認証が完了していません")
		}
	}

	return warnings
}

// ValidateEnvironmentDetailed システム環境確認機能の詳細確認
func ValidateEnvironmentDetailed() []string {
	var errors []string

	// OS情報確認
	fmt.Printf("   🖥️ OS: %s\n", runtime.GOOS)
	fmt.Printf("   🏗️ アーキテクチャ: %s\n", runtime.GOARCH)

	// 権限確認
	homeDir, _ := os.UserHomeDir()
	claudeDir := filepath.Join(homeDir, ".claude")

	// .claudeディレクトリの書き込み権限確認
	testFile := filepath.Join(claudeDir, "test_write")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		errors = append(errors, fmt.Sprintf(".claudeディレクトリへの書き込み権限がありません: %v", err))
	} else {
		_ = os.Remove(testFile) // テストファイル削除
		fmt.Printf("   ✅ ディレクトリ書き込み権限: 正常\n")
	}

	// 依存関係確認
	dependencies := []string{"tmux"}
	for _, dep := range dependencies {
		if path, err := exec.LookPath(dep); err != nil {
			errors = append(errors, fmt.Sprintf("依存関係 '%s' が見つかりません", dep))
		} else {
			fmt.Printf("   ✅ 依存関係 %s: %s\n", dep, path)
		}
	}

	// 環境変数確認
	shell := os.Getenv("SHELL")
	if shell == "" {
		errors = append(errors, "SHELL環境変数が設定されていません")
	} else {
		fmt.Printf("   ✅ SHELL: %s\n", shell)
	}

	return errors
}

// DisplaySolutionsForErrors エラーに対する解決策表示
func DisplaySolutionsForErrors(errors []string) {
	for _, err := range errors {
		switch {
		case strings.Contains(err, "Claude CLI実行可能ファイル"):
			fmt.Println("   → Claude CLIをインストールしてください")
			fmt.Println("      curl -fsSL https://anthropic.com/claude/install.sh | sh")
		case strings.Contains(err, "必要ディレクトリ"):
			fmt.Println("   → 必要なディレクトリを作成してください")
			fmt.Println("      mkdir -p ~/.claude/claude-code-agents/instructions")
		case strings.Contains(err, "settings.json"):
			fmt.Println("   → Claude CLIを起動して初期設定を完了してください")
			fmt.Println("      claude")
		case strings.Contains(err, "claude.json"):
			fmt.Println("   → Claude CLIにログインしてください")
			fmt.Println("      claude")
		case strings.Contains(err, "tmux"):
			fmt.Println("   → tmuxをインストールしてください")
			fmt.Println("      macOS: brew install tmux")
			fmt.Println("      Ubuntu: sudo apt install tmux")
		case strings.Contains(err, "書き込み権限"):
			fmt.Println("   → ディレクトリの権限を確認してください")
			fmt.Println("      chmod 750 ~/.claude")
		case strings.Contains(err, "SHELL環境変数"):
			fmt.Println("   → SHELL環境変数を設定してください")
			fmt.Println("      export SHELL=/bin/bash")
		}
	}
}

// DisplaySolutionsForWarnings 警告に対する推奨事項表示
func DisplaySolutionsForWarnings(warnings []string) {
	for _, warning := range warnings {
		switch {
		case strings.Contains(warning, "認証が完了していません"):
			fmt.Println("   → Claude CLIにログインすることを推奨します")
			fmt.Println("      claude")
		case strings.Contains(warning, "認証状態確認失敗"):
			fmt.Println("   → Claude CLIの再インストールを検討してください")
		case strings.Contains(warning, "設定ファイル確認失敗"):
			fmt.Println("   → Claude CLIの設定を再作成してください")
		}
	}
}

// findClaudeExecutableHelper Claude CLI実行可能ファイルを検索
func findClaudeExecutableHelper() string {
	// 一般的なパスを順番に確認
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

	// PATHから検索
	if path, err := exec.LookPath("claude"); err == nil {
		return path
	}

	return ""
}
