package common

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// SetupClaudeMockForBenchmark ベンチマーク用のClaude CLIモックをセットアップ
func SetupClaudeMockForBenchmark(b *testing.B) *ClaudeMockConfig {
	b.Helper()

	// ベンチマーク用の高速モック（最小限の処理）
	var commands []string
	if runtime.GOOS == "windows" {
		commands = []string{
			"@echo off",
			"rem Benchmark mock claude cli",
			"exit 0",
		}
	} else {
		commands = []string{
			"#!/bin/bash",
			"# Benchmark mock claude cli",
			"sleep 0.01", // 極短時間
			"exit 0",
		}
	}

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

	// 一時ディレクトリを作成
	tempDir := b.TempDir()

	// Claude CLIのダミーファイルを作成
	claudePath := filepath.Join(tempDir, "claude")
	err := os.WriteFile(claudePath, []byte(scriptContent), 0600)
	require.NoError(b, err)
	// 実行可能権限を追加
	//nolint:gosec // G302: テスト環境での実行可能ファイル作成のため必要
	err = os.Chmod(claudePath, 0700)
	require.NoError(b, err)

	// Windows環境では.exeファイルも作成
	if runtime.GOOS == "windows" {
		claudeExePath := filepath.Join(tempDir, "claude.exe")
		err := os.WriteFile(claudeExePath, []byte(scriptContent), 0600)
		require.NoError(b, err)
		// 実行可能権限を追加
		//nolint:gosec // G302: テスト環境での実行可能ファイル作成のため必要
		err = os.Chmod(claudeExePath, 0700)
		require.NoError(b, err)
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

// ValidateClaudeMockSetupForBenchmark ベンチマーク用モック環境の検証
func ValidateClaudeMockSetupForBenchmark(b *testing.B, config *ClaudeMockConfig) {
	b.Helper()

	// Claude CLIファイルが存在することを確認
	claudePath := filepath.Join(config.TempDir, "claude")
	_, err := os.Stat(claudePath)
	require.NoError(b, err, "Claude mock file should exist for benchmark")

	// ファイルが実行可能であることを確認
	info, err := os.Stat(claudePath)
	require.NoError(b, err)

	mode := info.Mode()
	if runtime.GOOS != "windows" {
		require.True(b, mode&0111 != 0, "Claude mock file should be executable for benchmark")
	}

	// PATHにテストディレクトリが含まれていることを確認
	currentPath := os.Getenv("PATH")
	require.Contains(b, currentPath, config.TempDir, "PATH should contain temp directory for benchmark")
}
