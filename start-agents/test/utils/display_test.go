package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/shivase/claude-code-agents/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// テスト用のヘルパー関数：標準出力をキャプチャする
func captureStdout(f func()) (string, error) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan bool)

	go func() {
		defer close(done)
		_, _ = io.Copy(&buf, r)
	}()

	f()

	_ = w.Close()
	os.Stdout = originalStdout
	<-done

	return buf.String(), nil
}

func TestSetVerboseLogging(t *testing.T) {
	// 初期状態をリセット
	utils.SetVerboseLogging(false)
	utils.SetSilentMode(false)

	tests := []struct {
		name            string
		verbose         bool
		expectedVerbose bool
		expectedSilent  bool
	}{
		{
			name:            "詳細ログ有効化",
			verbose:         true,
			expectedVerbose: true,
			expectedSilent:  false,
		},
		{
			name:            "詳細ログ無効化",
			verbose:         false,
			expectedVerbose: false,
			expectedSilent:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.SetVerboseLogging(tt.verbose)
			assert.Equal(t, tt.expectedVerbose, utils.IsVerboseLogging())
			assert.Equal(t, tt.expectedSilent, utils.IsSilentMode())
		})
	}

	t.Run("詳細ログがサイレントモードを無効化", func(t *testing.T) {
		utils.SetSilentMode(true)
		assert.True(t, utils.IsSilentMode())

		utils.SetVerboseLogging(true)
		assert.True(t, utils.IsVerboseLogging())
		assert.False(t, utils.IsSilentMode())
	})
}

func TestSetSilentMode(t *testing.T) {
	// 初期状態をリセット
	utils.SetVerboseLogging(false)
	utils.SetSilentMode(false)

	tests := []struct {
		name            string
		silent          bool
		expectedVerbose bool
		expectedSilent  bool
	}{
		{
			name:            "サイレントモード有効化",
			silent:          true,
			expectedVerbose: false,
			expectedSilent:  true,
		},
		{
			name:            "サイレントモード無効化",
			silent:          false,
			expectedVerbose: false,
			expectedSilent:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.SetSilentMode(tt.silent)
			assert.Equal(t, tt.expectedVerbose, utils.IsVerboseLogging())
			assert.Equal(t, tt.expectedSilent, utils.IsSilentMode())
		})
	}

	t.Run("サイレントモードが詳細ログを無効化", func(t *testing.T) {
		utils.SetVerboseLogging(true)
		assert.True(t, utils.IsVerboseLogging())

		utils.SetSilentMode(true)
		assert.False(t, utils.IsVerboseLogging())
		assert.True(t, utils.IsSilentMode())
	})
}

func TestDisplayProgress(t *testing.T) {
	// 初期状態をリセット
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		output, err := captureStdout(func() {
			utils.DisplayProgress("テスト操作", "進行中です")
		})

		require.NoError(t, err)
		assert.Contains(t, output, "🔄")
		assert.Contains(t, output, "テスト操作")
		assert.Contains(t, output, "進行中です")
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)

		output, err := captureStdout(func() {
			utils.DisplayProgress("テスト操作", "進行中です")
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		// テスト後にリセット
		utils.SetSilentMode(false)
	})
}

func TestDisplaySuccess(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		output, err := captureStdout(func() {
			utils.DisplaySuccess("テスト操作", "完了しました")
		})

		require.NoError(t, err)
		assert.Contains(t, output, "✅")
		assert.Contains(t, output, "テスト操作")
		assert.Contains(t, output, "完了しました")
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)

		output, err := captureStdout(func() {
			utils.DisplaySuccess("テスト操作", "完了しました")
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestDisplayError(t *testing.T) {
	t.Run("エラーメッセージ表示", func(t *testing.T) {
		testErr := fmt.Errorf("テストエラー")

		output, err := captureStdout(func() {
			utils.DisplayError("テスト操作", testErr)
		})

		require.NoError(t, err)
		assert.Contains(t, output, "❌")
		assert.Contains(t, output, "テスト操作")
		assert.Contains(t, output, "テストエラー")
	})

	t.Run("サイレントモードでも表示される", func(t *testing.T) {
		utils.SetSilentMode(true)
		testErr := fmt.Errorf("テストエラー")

		output, err := captureStdout(func() {
			utils.DisplayError("テスト操作", testErr)
		})

		require.NoError(t, err)
		assert.Contains(t, output, "❌")
		assert.Contains(t, output, "テスト操作")
		assert.Contains(t, output, "テストエラー")

		utils.SetSilentMode(false)
	})
}

func TestDisplayInfo(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		output, err := captureStdout(func() {
			utils.DisplayInfo("テスト操作", "情報メッセージ")
		})

		require.NoError(t, err)
		assert.Contains(t, output, "ℹ️")
		assert.Contains(t, output, "テスト操作")
		assert.Contains(t, output, "情報メッセージ")
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)

		output, err := captureStdout(func() {
			utils.DisplayInfo("テスト操作", "情報メッセージ")
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestDisplayWarning(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		output, err := captureStdout(func() {
			utils.DisplayWarning("テスト操作", "警告メッセージ")
		})

		require.NoError(t, err)
		assert.Contains(t, output, "⚠️")
		assert.Contains(t, output, "テスト操作")
		assert.Contains(t, output, "警告メッセージ")
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)

		output, err := captureStdout(func() {
			utils.DisplayWarning("テスト操作", "警告メッセージ")
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestDisplayStartupBanner(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		output, err := captureStdout(func() {
			utils.DisplayStartupBanner()
		})

		require.NoError(t, err)
		assert.Contains(t, output, "🚀 AI Teams System")
		assert.Contains(t, output, "Claude Code Agents")
		assert.Contains(t, output, "Version: 1.0.0")
		assert.Contains(t, output, "Runtime: Go")
		assert.Contains(t, output, runtime.Version())
		assert.Contains(t, output, fmt.Sprintf("Platform: %s/%s", runtime.GOOS, runtime.GOARCH))
		assert.Contains(t, output, "Start Time:")
		assert.Contains(t, output, "=====================================")
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)

		output, err := captureStdout(func() {
			utils.DisplayStartupBanner()
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestDisplayLauncherStart(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		output, err := captureStdout(func() {
			utils.DisplayLauncherStart()
		})

		require.NoError(t, err)
		assert.Contains(t, output, "🚀 システムランチャー開始")
		assert.Contains(t, output, "=====================================")
		// タイムスタンプフォーマットの確認
		timePattern := time.Now().Format("15:04")[:4] // HH:MM部分だけチェック
		assert.Contains(t, output, timePattern)
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)

		output, err := captureStdout(func() {
			utils.DisplayLauncherStart()
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestDisplayLauncherProgress(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		output, err := captureStdout(func() {
			utils.DisplayLauncherProgress()
		})

		require.NoError(t, err)
		assert.Contains(t, output, "🔄 システム初期化中...")
		// タイムスタンプフォーマットの確認
		timePattern := time.Now().Format("15:04")[:4] // HH:MM部分だけチェック
		assert.Contains(t, output, timePattern)
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)

		output, err := captureStdout(func() {
			utils.DisplayLauncherProgress()
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestDisplayConfig(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("map形式の設定表示", func(t *testing.T) {
		config := map[string]interface{}{
			"dev_count":  4,
			"log_level":  "info",
			"debug_mode": true,
		}

		output, err := captureStdout(func() {
			utils.DisplayConfig(config, "test-session")
		})

		require.NoError(t, err)
		assert.Contains(t, output, "📋 設定情報")
		assert.Contains(t, output, "セッション名: test-session")
		assert.Contains(t, output, "dev_count")
		assert.Contains(t, output, "log_level")
		assert.Contains(t, output, "debug_mode")
	})

	t.Run("非map形式の設定", func(t *testing.T) {
		config := "非マップ設定"

		output, err := captureStdout(func() {
			utils.DisplayConfig(config, "test-session")
		})

		require.NoError(t, err)
		assert.Contains(t, output, "📋 設定情報")
		assert.Contains(t, output, "セッション名: test-session")
		// 設定詳細は表示されない
		assert.NotContains(t, output, "非マップ設定")
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)
		config := map[string]interface{}{"test": "value"}

		output, err := captureStdout(func() {
			utils.DisplayConfig(config, "test-session")
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestDisplayValidationResults(t *testing.T) {
	utils.SetSilentMode(false)

	t.Run("通常モードでの表示", func(t *testing.T) {
		config := map[string]interface{}{"test": "value"}

		output, err := captureStdout(func() {
			utils.DisplayValidationResults(config)
		})

		require.NoError(t, err)
		assert.Contains(t, output, "🔍 検証結果")
		assert.Contains(t, output, "✅ Claude CLI: 利用可能")
		assert.Contains(t, output, "✅ インストラクション: 準備完了")
		assert.Contains(t, output, "✅ 作業ディレクトリ: アクセス可能")
	})

	t.Run("サイレントモードでの非表示", func(t *testing.T) {
		utils.SetSilentMode(true)
		config := map[string]interface{}{"test": "value"}

		output, err := captureStdout(func() {
			utils.DisplayValidationResults(config)
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		utils.SetSilentMode(false)
	})
}

func TestFormatPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "空のパス",
			path:     "",
			expected: "<empty>",
		},
		{
			name:     "ホームディレクトリ",
			path:     homeDir,
			expected: "~",
		},
		{
			name:     "ホームディレクトリ内のパス",
			path:     filepath.Join(homeDir, "documents", "test.txt"),
			expected: "~/documents/test.txt",
		},
		{
			name:     "絶対パス",
			path:     "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "相対パス",
			path:     "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.txt")
	err := os.WriteFile(existingFile, []byte("test"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "空のパス",
			path:     "",
			expected: false,
		},
		{
			name:     "存在するファイル",
			path:     existingFile,
			expected: true,
		},
		{
			name:     "存在するディレクトリ",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "存在しないパス",
			path:     filepath.Join(tmpDir, "nonexistent.txt"),
			expected: false,
		},
		{
			name:     "チルダパス（ホームディレクトリ）",
			path:     "~",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidatePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandPathOld(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "チルダパス",
			path:     "~/test",
			expected: filepath.Join(homeDir, "test"),
		},
		{
			name:     "通常のパス",
			path:     "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "空のパス",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ExpandPathOld(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	// 実行可能ファイルを作成
	executableFile := filepath.Join(tmpDir, "executable")
	err := os.WriteFile(executableFile, []byte("#!/bin/bash\necho test"), 0755)
	require.NoError(t, err)

	// 非実行可能ファイルを作成
	nonExecutableFile := filepath.Join(tmpDir, "non_executable.txt")
	err = os.WriteFile(nonExecutableFile, []byte("test content"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "実行可能ファイル",
			path:     executableFile,
			expected: true,
		},
		{
			name:     "非実行可能ファイル",
			path:     nonExecutableFile,
			expected: false,
		},
		{
			name:     "存在しないファイル",
			path:     filepath.Join(tmpDir, "nonexistent"),
			expected: false,
		},
		{
			name:     "ディレクトリ（実行可能）",
			path:     tmpDir,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsExecutable(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDisplayFunctionsIntegration(t *testing.T) {
	// 初期状態をリセット
	utils.SetVerboseLogging(false)
	utils.SetSilentMode(false)

	t.Run("表示モードの切り替えテスト", func(t *testing.T) {
		// 通常モード
		output1, err := captureStdout(func() {
			utils.DisplayProgress("テスト", "通常モード")
		})
		require.NoError(t, err)
		assert.Contains(t, output1, "テスト")

		// サイレントモードに切り替え
		utils.SetSilentMode(true)
		output2, err := captureStdout(func() {
			utils.DisplayProgress("テスト", "サイレントモード")
		})
		require.NoError(t, err)
		assert.Empty(t, output2)

		// 詳細モードに切り替え（サイレントモードが自動的に無効になる）
		utils.SetVerboseLogging(true)
		output3, err := captureStdout(func() {
			utils.DisplayProgress("テスト", "詳細モード")
		})
		require.NoError(t, err)
		assert.Contains(t, output3, "テスト")
		assert.True(t, utils.IsVerboseLogging())
		assert.False(t, utils.IsSilentMode())
	})
}

// ベンチマークテスト
func BenchmarkDisplayProgress(b *testing.B) {
	// 標準出力を無効化してベンチマークを実行
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() {
		os.Stdout = originalStdout
	}()

	utils.SetSilentMode(false)

	for i := 0; i < b.N; i++ {
		utils.DisplayProgress("ベンチマークテスト", "実行中")
	}
}

func BenchmarkFormatPath(b *testing.B) {
	homeDir, _ := os.UserHomeDir()
	testPath := filepath.Join(homeDir, "test", "benchmark", "path.txt")

	for i := 0; i < b.N; i++ {
		utils.FormatPath(testPath)
	}
}

func BenchmarkValidatePath(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < b.N; i++ {
		utils.ValidatePath(tmpDir)
	}
}
