package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ClaudeMockConfig Claude CLIモック設定
type ClaudeMockConfig struct {
	ScriptContent string // モックスクリプトの内容
	TempDir       string // 一時ディレクトリパス
	OriginalPath  string // 元のPATH環境変数
	CleanupFunc   func() // クリーンアップ関数
}

// SetupClaudeMock setupClaudeMock Claude CLIのモック環境をセットアップ
func SetupClaudeMock(t *testing.T, scriptContent string) *ClaudeMockConfig {
	t.Helper()

	// デフォルトのスクリプト内容（プロセスを生かし続ける）
	if scriptContent == "" {
		if runtime.GOOS == "windows" {
			scriptContent = "@echo off\necho mock claude cli\n:loop\nping -n 2 127.0.0.1 >nul\ngoto loop\n"
		} else {
			scriptContent = "#!/bin/bash\necho 'mock claude cli'\n# Keep process alive for tests\nwhile true; do\n  sleep 1\ndone\n"
		}
	}

	// 一時ディレクトリを作成
	tempDir := t.TempDir()

	// Claude CLIのダミーファイルを作成（claude-codeも作成）
	claudePath := filepath.Join(tempDir, "claude")
	err := os.WriteFile(claudePath, []byte(scriptContent), 0600)
	require.NoError(t, err)
	// 実行可能権限を追加
	//nolint:gosec // G302: テスト環境での実行可能ファイル作成のため必要
	err = os.Chmod(claudePath, 0700)
	require.NoError(t, err)

	// claude-codeコマンドも作成（新しいCLIコマンド名）
	claudeCodePath := filepath.Join(tempDir, "claude-code")
	err = os.WriteFile(claudeCodePath, []byte(scriptContent), 0600)
	require.NoError(t, err)
	//nolint:gosec // G302: テスト環境での実行可能ファイル作成のため必要
	err = os.Chmod(claudeCodePath, 0700)
	require.NoError(t, err)

	// Windows環境では.exeファイルも作成
	if runtime.GOOS == "windows" {
		claudeExePath := filepath.Join(tempDir, "claude.exe")
		// デフォルトスクリプトがWindowsのものかチェック
		windowsScript := scriptContent
		if !strings.Contains(scriptContent, "@echo off") {
			windowsScript = "@echo off\necho mock claude cli\n:loop\nping -n 2 127.0.0.1 >nul\ngoto loop\n"
		}
		err := os.WriteFile(claudeExePath, []byte(windowsScript), 0600)
		require.NoError(t, err)
		// 実行可能権限を追加
		//nolint:gosec // G302: テスト環境での実行可能ファイル作成のため必要
		err = os.Chmod(claudeExePath, 0700)
		require.NoError(t, err)

		// claude-code.exeも作成
		claudeCodeExePath := filepath.Join(tempDir, "claude-code.exe")
		err = os.WriteFile(claudeCodeExePath, []byte(windowsScript), 0600)
		require.NoError(t, err)
		//nolint:gosec // G302: テスト環境での実行可能ファイル作成のため必要
		err = os.Chmod(claudeCodeExePath, 0700)
		require.NoError(t, err)
	}

	// 元のPATH環境変数を保存
	originalPath := os.Getenv("PATH")

	// PATHにテストディレクトリを追加
	var newPath string
	if runtime.GOOS == "windows" {
		newPath = tempDir + ";" + originalPath
	} else {
		newPath = tempDir + ":" + originalPath
	}
	_ = os.Setenv("PATH", newPath)

	// クリーンアップ関数を定義
	cleanupFunc := func() {
		_ = os.Setenv("PATH", originalPath)
	}

	return &ClaudeMockConfig{
		ScriptContent: scriptContent,
		TempDir:       tempDir,
		OriginalPath:  originalPath,
		CleanupFunc:   cleanupFunc,
	}
}

// SetupClaudeMockWithCustomCommand カスタムコマンドでClaude CLIモックをセットアップ
func SetupClaudeMockWithCustomCommand(t *testing.T, commands []string) *ClaudeMockConfig {
	t.Helper()

	var scriptContent string
	if runtime.GOOS == "windows" {
		scriptContent = "@echo off\n"
		for _, cmd := range commands {
			scriptContent += cmd + "\n"
		}
	} else {
		scriptContent = "#!/bin/bash\n"
		for _, cmd := range commands {
			scriptContent += cmd + "\n"
		}
	}

	return SetupClaudeMock(t, scriptContent)
}

// SetupClaudeMockWithTimeout タイムアウト付きClaude CLIモックをセットアップ
func SetupClaudeMockWithTimeout(t *testing.T, timeoutSeconds int) *ClaudeMockConfig {
	t.Helper()

	var commands []string
	if runtime.GOOS == "windows" {
		commands = []string{
			"echo mock claude cli with timeout",
			fmt.Sprintf("timeout /t %d >nul", timeoutSeconds),
		}
	} else {
		commands = []string{
			"echo 'mock claude cli with timeout'",
			fmt.Sprintf("sleep %d", timeoutSeconds),
		}
	}

	return SetupClaudeMockWithCustomCommand(t, commands)
}

// SetupClaudeMockWithFailure 失敗するClaude CLIモックをセットアップ
func SetupClaudeMockWithFailure(t *testing.T, exitCode int) *ClaudeMockConfig {
	t.Helper()

	var commands []string
	if runtime.GOOS == "windows" {
		commands = []string{
			"echo mock claude cli failure",
			fmt.Sprintf("exit %d", exitCode),
		}
	} else {
		commands = []string{
			"echo 'mock claude cli failure'",
			fmt.Sprintf("exit %d", exitCode),
		}
	}

	return SetupClaudeMockWithCustomCommand(t, commands)
}

// TeardownClaudeMock Claude CLIモック環境をクリーンアップ
func TeardownClaudeMock(config *ClaudeMockConfig) {
	if config != nil && config.CleanupFunc != nil {
		config.CleanupFunc()
	}
}

// SetupClaudeMockForCI CI環境用の安定したClaude CLIモックをセットアップ
func SetupClaudeMockForCI(t *testing.T) *ClaudeMockConfig {
	t.Helper()

	// CLAUDE_MOCK_ENV環境変数を設定（テスト用モックであることを示す）
	t.Setenv("CLAUDE_MOCK_ENV", "true")

	// CI環境では即座に終了するモックを使用
	var commands []string
	if runtime.GOOS == "windows" {
		commands = []string{
			"echo CI mock claude cli",
			"exit 0",
		}
	} else {
		commands = []string{
			"echo 'CI mock claude cli'",
			"exit 0",
		}
	}

	return SetupClaudeMockWithCustomCommand(t, commands)
}

// SetupClaudeMockForManager Manager用のClaude CLIモックをセットアップ
func SetupClaudeMockForManager(t *testing.T) *ClaudeMockConfig {
	t.Helper()

	// CLAUDE_MOCK_ENV環境変数を設定（テスト用モックであることを示す）
	t.Setenv("CLAUDE_MOCK_ENV", "true")

	var commands []string
	if runtime.GOOS == "windows" {
		commands = []string{
			"echo Manager mock claude cli",
			"exit 0",
		}
	} else {
		commands = []string{
			"echo 'Manager mock claude cli'",
			"exit 0",
		}
	}

	return SetupClaudeMockWithCustomCommand(t, commands)
}

// ValidateClaudeMockSetup モック環境が正しくセットアップされているかを検証
func ValidateClaudeMockSetup(t *testing.T, config *ClaudeMockConfig) {
	t.Helper()

	// Claude CLIファイルが存在することを確認
	claudePath := filepath.Join(config.TempDir, "claude")
	_, err := os.Stat(claudePath)
	require.NoError(t, err, "Claude mock file should exist")

	// ファイルが実行可能であることを確認
	info, err := os.Stat(claudePath)
	require.NoError(t, err)

	mode := info.Mode()
	if runtime.GOOS != "windows" {
		require.True(t, mode&0111 != 0, "Claude mock file should be executable")
	}

	// PATHにテストディレクトリが含まれていることを確認
	currentPath := os.Getenv("PATH")
	require.Contains(t, currentPath, config.TempDir, "PATH should contain temp directory")
}
