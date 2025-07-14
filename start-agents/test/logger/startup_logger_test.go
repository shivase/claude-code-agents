package logger

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStartupLogger(t *testing.T) {
	t.Run("起動ログインスタンス作成", func(t *testing.T) {
		startupLogger := logger.NewStartupLogger()
		assert.NotNil(t, startupLogger)
		assert.Implements(t, (*logger.StartupLogger)(nil), startupLogger)
	})
}

func TestDefaultStartupLogger_LogSystemInit(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()
	details := map[string]interface{}{
		"version":    "1.0.0",
		"debug_mode": true,
	}

	sl.LogSystemInit("application_startup", details)

	output := buf.String()
	assert.Contains(t, output, "🚀 システム初期化")
	assert.Contains(t, output, "startup")
	assert.Contains(t, output, "application_startup")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "debug_mode")
}

func TestDefaultStartupLogger_LogConfigLoad(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()

	t.Run("詳細情報ありの設定読み込み", func(t *testing.T) {
		buf.Reset()
		details := map[string]interface{}{
			"file_size": 1024,
			"encoding":  "utf-8",
		}

		sl.LogConfigLoad("/path/to/config.yaml", details)

		output := buf.String()
		assert.Contains(t, output, "📋 設定ファイル読み込み")
		assert.Contains(t, output, "config_load")
		assert.Contains(t, output, "/path/to/config.yaml")
		assert.Contains(t, output, "1024")
		assert.Contains(t, output, "utf-8")
	})

	t.Run("詳細情報なしの設定読み込み", func(t *testing.T) {
		buf.Reset()
		sl.LogConfigLoad("/path/to/config.yaml", nil)

		output := buf.String()
		assert.Contains(t, output, "📋 設定ファイル読み込み")
		assert.Contains(t, output, "/path/to/config.yaml")
	})
}

func TestDefaultStartupLogger_LogInstructionConfig(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()

	instructionInfo := map[string]interface{}{
		"po_instruction":      "po.md",
		"manager_instruction": "manager.md",
		"dev_instruction":     "developer.md",
	}

	details := map[string]interface{}{
		"instruction_dir": "/path/to/instructions",
		"total_files":     3,
	}

	sl.LogInstructionConfig(instructionInfo, details)

	output := buf.String()
	assert.Contains(t, output, "📝 instruction設定確認")
	assert.Contains(t, output, "instruction_config")
	assert.Contains(t, output, "po.md")
	assert.Contains(t, output, "manager.md")
	assert.Contains(t, output, "developer.md")
	assert.Contains(t, output, "/path/to/instructions")
}

func TestDefaultStartupLogger_LogEnvironmentInfo(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	sl := logger.NewStartupLogger()
	envInfo := map[string]interface{}{
		"go_version": "1.21",
		"os":         "darwin",
		"arch":       "amd64",
	}

	t.Run("通常モードでの環境情報ログ", func(t *testing.T) {
		buf.Reset()
		sl.LogEnvironmentInfo(envInfo, false)

		output := buf.String()
		assert.Contains(t, output, "🔍 環境情報確認")
		assert.Contains(t, output, "environment_check")
		assert.Contains(t, output, "1.21")
		assert.Contains(t, output, "darwin")
		assert.Contains(t, output, "amd64")
		assert.Contains(t, output, "\"debug_mode\":false")
	})

	t.Run("デバッグモードでの環境情報ログ", func(t *testing.T) {
		buf.Reset()
		sl.LogEnvironmentInfo(envInfo, true)

		output := buf.String()
		assert.Contains(t, output, "🔍 環境情報確認")
		assert.Contains(t, output, "\"debug_mode\":true")
	})
}

func TestDefaultStartupLogger_LogTmuxSetup(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()

	t.Run("詳細情報ありのtmuxセットアップ", func(t *testing.T) {
		buf.Reset()
		details := map[string]interface{}{
			"layout":      "tiled",
			"window_name": "ai-teams",
		}

		sl.LogTmuxSetup("test-session", 6, details)

		output := buf.String()
		assert.Contains(t, output, "🖥️  tmuxセッション設定")
		assert.Contains(t, output, "tmux_setup")
		assert.Contains(t, output, "test-session")
		assert.Contains(t, output, "6")
		assert.Contains(t, output, "tiled")
		assert.Contains(t, output, "ai-teams")
	})

	t.Run("詳細情報なしのtmuxセットアップ", func(t *testing.T) {
		buf.Reset()
		sl.LogTmuxSetup("simple-session", 4, nil)

		output := buf.String()
		assert.Contains(t, output, "🖥️  tmuxセッション設定")
		assert.Contains(t, output, "simple-session")
		assert.Contains(t, output, "4")
	})
}

func TestDefaultStartupLogger_LogClaudeStart(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()

	t.Run("詳細情報ありのClaude CLI起動", func(t *testing.T) {
		buf.Reset()
		details := map[string]interface{}{
			"cli_path":    "/usr/local/bin/claude",
			"config_file": "~/.claude/config.json",
		}

		sl.LogClaudeStart("dev1", "0", details)

		output := buf.String()
		assert.Contains(t, output, "🤖 Claude CLI起動")
		assert.Contains(t, output, "claude_start")
		assert.Contains(t, output, "dev1")
		assert.Contains(t, output, "\"pane_id\":\"0\"")
		assert.Contains(t, output, "/usr/local/bin/claude")
		assert.Contains(t, output, "~/.claude/config.json")
	})

	t.Run("詳細情報なしのClaude CLI起動", func(t *testing.T) {
		buf.Reset()
		sl.LogClaudeStart("manager", "1", nil)

		output := buf.String()
		assert.Contains(t, output, "🤖 Claude CLI起動")
		assert.Contains(t, output, "manager")
		assert.Contains(t, output, "\"pane_id\":\"1\"")
	})
}

func TestDefaultStartupLogger_LogStartupComplete(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()
	totalTime := 5 * time.Second

	t.Run("詳細情報ありの起動完了", func(t *testing.T) {
		buf.Reset()
		details := map[string]interface{}{
			"agents_started": 6,
			"session_name":   "ai-teams",
		}

		sl.LogStartupComplete(totalTime, details)

		output := buf.String()
		assert.Contains(t, output, "✅ 起動完了")
		assert.Contains(t, output, "complete")
		assert.Contains(t, output, "5s")
		assert.Contains(t, output, "5000")
		assert.Contains(t, output, "6")
		assert.Contains(t, output, "ai-teams")
	})

	t.Run("詳細情報なしの起動完了", func(t *testing.T) {
		buf.Reset()
		sl.LogStartupComplete(totalTime, nil)

		output := buf.String()
		assert.Contains(t, output, "✅ 起動完了")
		assert.Contains(t, output, "5s")
		assert.Contains(t, output, "5000")
	})
}

func TestDefaultStartupLogger_LogStartupError(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()
	testErr := fmt.Errorf("テストエラー")

	t.Run("復旧情報ありのエラーログ", func(t *testing.T) {
		buf.Reset()
		recovery := map[string]interface{}{
			"action":   "retry",
			"attempts": 3,
		}

		sl.LogStartupError("claude_start", testErr, recovery)

		output := buf.String()
		assert.Contains(t, output, "❌ 起動エラー")
		assert.Contains(t, output, "claude_start")
		assert.Contains(t, output, "テストエラー")
		assert.Contains(t, output, "retry")
		assert.Contains(t, output, "3")
	})

	t.Run("復旧情報なしのエラーログ", func(t *testing.T) {
		buf.Reset()
		sl.LogStartupError("config_load", testErr, nil)

		output := buf.String()
		assert.Contains(t, output, "❌ 起動エラー")
		assert.Contains(t, output, "config_load")
		assert.Contains(t, output, "テストエラー")
	})
}

func TestDefaultStartupLogger_BeginPhase(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()
	details := map[string]interface{}{
		"operation": "system_init",
		"step":      1,
	}

	phase := sl.BeginPhase("initialization", details)

	require.NotNil(t, phase)
	assert.Equal(t, "initialization", phase.Name)
	assert.Equal(t, details, phase.Details)
	assert.WithinDuration(t, time.Now(), phase.StartTime, time.Second)

	output := buf.String()
	assert.Contains(t, output, "🔄 起動フェーズ開始")
	assert.Contains(t, output, "initialization")
	assert.Contains(t, output, "started")
	assert.Contains(t, output, "system_init")
}

func TestStartupPhase_Complete(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	details := map[string]interface{}{
		"operation": "test_phase",
	}

	phase := &logger.StartupPhase{
		Name:      "test_phase",
		StartTime: time.Now().Add(-100 * time.Millisecond),
		Details:   details,
	}

	phase.Complete()

	output := buf.String()
	assert.Contains(t, output, "✅ 起動フェーズ完了")
	assert.Contains(t, output, "test_phase")
	assert.Contains(t, output, "completed")
	assert.Contains(t, output, "duration")
	assert.Contains(t, output, "test_phase")
}

func TestStartupPhase_CompleteWithError(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	details := map[string]interface{}{
		"operation": "failed_phase",
	}

	phase := &logger.StartupPhase{
		Name:      "failed_phase",
		StartTime: time.Now().Add(-200 * time.Millisecond),
		Details:   details,
	}

	testErr := fmt.Errorf("フェーズ失敗")
	phase.CompleteWithError(testErr)

	output := buf.String()
	assert.Contains(t, output, "❌ 起動フェーズ失敗")
	assert.Contains(t, output, "failed_phase")
	assert.Contains(t, output, "failed")
	assert.Contains(t, output, "duration")
	assert.Contains(t, output, "フェーズ失敗")
}

func TestLogStartupProgress(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	logger.LogStartupProgress("agent_initialization", 3, 6)

	output := buf.String()
	assert.Contains(t, output, "📊 起動プログレス")
	assert.Contains(t, output, "agent_initialization")
	assert.Contains(t, output, "3")
	assert.Contains(t, output, "6")
	assert.Contains(t, output, "50.0%")
}

func TestLogStartupDebug(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	details := map[string]interface{}{
		"debug_info": "詳細デバッグ情報",
	}

	logger.LogStartupDebug("debug_phase", "デバッグメッセージ", details)

	output := buf.String()
	assert.Contains(t, output, "🔍 デバッグメッセージ")
	assert.Contains(t, output, "debug_phase")
	assert.Contains(t, output, "詳細デバッグ情報")
}

func TestLogStartupWarning(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	details := map[string]interface{}{
		"warning_code": "W001",
	}

	logger.LogStartupWarning("config_validation", "設定に警告があります", details)

	output := buf.String()
	assert.Contains(t, output, "⚠️  設定に警告があります")
	assert.Contains(t, output, "config_validation")
	assert.Contains(t, output, "W001")
}

func TestPackageLevelFunctions(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	t.Run("パッケージレベル関数のテスト", func(t *testing.T) {
		// LogSystemInit
		buf.Reset()
		logger.LogSystemInit("app_start", map[string]interface{}{"test": "value"})
		output := buf.String()
		assert.Contains(t, output, "🚀 システム初期化")

		// LogConfigLoad
		buf.Reset()
		logger.LogConfigLoad("/config/test.yaml", nil)
		output = buf.String()
		assert.Contains(t, output, "📋 設定ファイル読み込み")

		// LogInstructionConfig
		buf.Reset()
		instructionInfo := map[string]interface{}{"test_instruction": "test.md"}
		logger.LogInstructionConfig(instructionInfo, nil)
		output = buf.String()
		assert.Contains(t, output, "📝 instruction設定確認")

		// LogEnvironmentInfo
		buf.Reset()
		envInfo := map[string]interface{}{"test_env": "test_value"}
		logger.LogEnvironmentInfo(envInfo, false)
		output = buf.String()
		assert.Contains(t, output, "🔍 環境情報確認")

		// LogTmuxSetup
		buf.Reset()
		logger.LogTmuxSetup("test-session", 4, nil)
		output = buf.String()
		assert.Contains(t, output, "🖥️  tmuxセッション設定")

		// LogClaudeStart
		buf.Reset()
		logger.LogClaudeStart("test-agent", "0", nil)
		output = buf.String()
		assert.Contains(t, output, "🤖 Claude CLI起動")

		// LogStartupComplete
		buf.Reset()
		logger.LogStartupComplete(1*time.Second, nil)
		output = buf.String()
		assert.Contains(t, output, "✅ 起動完了")

		// LogStartupError
		buf.Reset()
		logger.LogStartupError("test_phase", fmt.Errorf("test error"), nil)
		output = buf.String()
		assert.Contains(t, output, "❌ 起動エラー")

		// BeginPhase
		buf.Reset()
		phase := logger.BeginPhase("test_phase", map[string]interface{}{"test": "value"})
		assert.NotNil(t, phase)
		output = buf.String()
		assert.Contains(t, output, "🔄 起動フェーズ開始")
	})
}

func TestStartupLoggerInterface(t *testing.T) {
	t.Run("インターフェース実装確認", func(t *testing.T) {
		sl := logger.NewStartupLogger()
		assert.Implements(t, (*logger.StartupLogger)(nil), sl)

		// 全メソッドの呼び出し確認
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		details := map[string]interface{}{"test": "value"}

		require.NotPanics(t, func() {
			sl.LogSystemInit("test", details)
			sl.LogConfigLoad("test.yaml", details)
			sl.LogInstructionConfig(details, details)
			sl.LogEnvironmentInfo(details, false)
			sl.LogTmuxSetup("test", 1, details)
			sl.LogClaudeStart("test", "0", details)
			sl.LogStartupComplete(1*time.Second, details)
			sl.LogStartupError("test", fmt.Errorf("test"), details)
			phase := sl.BeginPhase("test", details)
			phase.Complete()
		})
	})
}

// ベンチマークテスト
func BenchmarkDefaultStartupLogger_LogSystemInit(b *testing.B) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	sl := logger.NewStartupLogger()
	details := map[string]interface{}{
		"version":    "1.0.0",
		"debug_mode": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.LogSystemInit("benchmark", details)
	}
}

func BenchmarkLogStartupProgress(b *testing.B) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogStartupProgress("benchmark", i%10, 10)
	}
}

func BenchmarkStartupPhase_Complete(b *testing.B) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	details := map[string]interface{}{"benchmark": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		phase := &logger.StartupPhase{
			Name:      "benchmark_phase",
			StartTime: time.Now(),
			Details:   details,
		}
		phase.Complete()
	}
}
