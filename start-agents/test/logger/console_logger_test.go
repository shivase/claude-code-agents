package logger

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConsoleLogger(t *testing.T) {
	// 元のグローバルレベルを保存
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	tests := []struct {
		name          string
		level         string
		expectedLevel zerolog.Level
	}{
		{
			name:          "Debug level",
			level:         "debug",
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "Info level",
			level:         "info",
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name:          "Warn level",
			level:         "warn",
			expectedLevel: zerolog.WarnLevel,
		},
		{
			name:          "Error level",
			level:         "error",
			expectedLevel: zerolog.ErrorLevel,
		},
		{
			name:          "Fatal level",
			level:         "fatal",
			expectedLevel: zerolog.FatalLevel,
		},
		{
			name:          "Invalid level",
			level:         "invalid",
			expectedLevel: zerolog.InfoLevel, // デフォルト
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.InitConsoleLogger(tt.level)
			assert.Equal(t, tt.expectedLevel, zerolog.GlobalLevel())
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	tests := []struct {
		name          string
		level         string
		expectedLevel zerolog.Level
	}{
		{
			name:          "Debug level setting",
			level:         "debug",
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "Info level setting",
			level:         "info",
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name:          "Invalid level setting",
			level:         "invalid",
			expectedLevel: zerolog.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.SetLogLevel(tt.level)
			assert.Equal(t, tt.expectedLevel, zerolog.GlobalLevel())
		})
	}
}

func TestInitWithDebugFlag(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	// 標準出力をキャプチャ
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		name           string
		debugEnabled   bool
		baseLogLevel   string
		expectedLevel  zerolog.Level
		expectDebugMsg bool
	}{
		{
			name:           "Debug flag enabled",
			debugEnabled:   true,
			baseLogLevel:   "info",
			expectedLevel:  zerolog.DebugLevel,
			expectDebugMsg: true,
		},
		{
			name:           "Debug flag disabled",
			debugEnabled:   false,
			baseLogLevel:   "warn",
			expectedLevel:  zerolog.WarnLevel,
			expectDebugMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.InitWithDebugFlag(tt.debugEnabled, tt.baseLogLevel)
			assert.Equal(t, tt.expectedLevel, zerolog.GlobalLevel())
		})
	}

	// 標準出力を復元
	w.Close()
	os.Stdout = originalStdout

	// 出力内容の確認
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if len(output) > 0 {
		assert.Contains(t, output, "Debug mode enabled")
	}
}

func TestGetCurrentLogLevel(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	tests := []struct {
		name          string
		level         zerolog.Level
		expectedLevel string
	}{
		{
			name:          "デバッグレベル取得",
			level:         zerolog.DebugLevel,
			expectedLevel: "debug",
		},
		{
			name:          "情報レベル取得",
			level:         zerolog.InfoLevel,
			expectedLevel: "info",
		},
		{
			name:          "警告レベル取得",
			level:         zerolog.WarnLevel,
			expectedLevel: "warn",
		},
		{
			name:          "エラーレベル取得",
			level:         zerolog.ErrorLevel,
			expectedLevel: "error",
		},
		{
			name:          "致命的レベル取得",
			level:         zerolog.FatalLevel,
			expectedLevel: "fatal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zerolog.SetGlobalLevel(tt.level)
			assert.Equal(t, tt.expectedLevel, logger.GetCurrentLogLevel())
		})
	}
}

func TestIsDebugEnabled(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	tests := []struct {
		name     string
		level    zerolog.Level
		expected bool
	}{
		{
			name:     "デバッグレベルで有効",
			level:    zerolog.DebugLevel,
			expected: true,
		},
		{
			name:     "情報レベルで無効",
			level:    zerolog.InfoLevel,
			expected: false,
		},
		{
			name:     "警告レベルで無効",
			level:    zerolog.WarnLevel,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zerolog.SetGlobalLevel(tt.level)
			assert.Equal(t, tt.expected, logger.IsDebugEnabled())
		})
	}
}

func TestApplyConfigLogLevel(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	// ログ出力をキャプチャ
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	tests := []struct {
		name           string
		configLogLevel string
		debugOverride  bool
		expectedLevel  string
		expectDebugMsg bool
	}{
		{
			name:           "デバッグオーバーライド有効",
			configLogLevel: "warn",
			debugOverride:  true,
			expectedLevel:  "debug",
			expectDebugMsg: true,
		},
		{
			name:           "設定適用",
			configLogLevel: "error",
			debugOverride:  false,
			expectedLevel:  "error",
			expectDebugMsg: false,
		},
		{
			name:           "空の設定",
			configLogLevel: "",
			debugOverride:  false,
			expectedLevel:  "info", // 変更されない
			expectDebugMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 初期状態を設定
			logger.SetLogLevel("info")
			buf.Reset()

			logger.ApplyConfigLogLevel(tt.configLogLevel, tt.debugOverride)
			assert.Equal(t, tt.expectedLevel, logger.GetCurrentLogLevel())
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{
			name:     "有効なレベル - debug",
			level:    "debug",
			expected: true,
		},
		{
			name:     "有効なレベル - info",
			level:    "info",
			expected: true,
		},
		{
			name:     "有効なレベル - warn",
			level:    "warn",
			expected: true,
		},
		{
			name:     "有効なレベル - error",
			level:    "error",
			expected: true,
		},
		{
			name:     "有効なレベル - fatal",
			level:    "fatal",
			expected: true,
		},
		{
			name:     "無効なレベル",
			level:    "invalid",
			expected: false,
		},
		{
			name:     "空文字列",
			level:    "",
			expected: false,
		},
		{
			name:     "大文字",
			level:    "INFO",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.ValidateLogLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLogLevelPriority(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	tests := []struct {
		name             string
		level            zerolog.Level
		expectedPriority int
	}{
		{
			name:             "デバッグレベル優先度",
			level:            zerolog.DebugLevel,
			expectedPriority: 0,
		},
		{
			name:             "情報レベル優先度",
			level:            zerolog.InfoLevel,
			expectedPriority: 1,
		},
		{
			name:             "警告レベル優先度",
			level:            zerolog.WarnLevel,
			expectedPriority: 2,
		},
		{
			name:             "エラーレベル優先度",
			level:            zerolog.ErrorLevel,
			expectedPriority: 3,
		},
		{
			name:             "致命的レベル優先度",
			level:            zerolog.FatalLevel,
			expectedPriority: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zerolog.SetGlobalLevel(tt.level)
			assert.Equal(t, tt.expectedPriority, logger.GetLogLevelPriority())
		})
	}
}

func TestLogWithError(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	testErr := fmt.Errorf("test error")
	fields := map[string]interface{}{
		"session":  "test-session",
		"count":    42,
		"enabled":  true,
		"duration": 5 * time.Second,
	}

	logger.LogWithError(testErr, "Test error message", fields)

	output := buf.String()
	assert.Contains(t, output, "Test error message")
	assert.Contains(t, output, "test error")
	assert.Contains(t, output, "test-session")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "true")
}

func TestLogDebugWithCondition(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

	tests := []struct {
		name       string
		condition  bool
		debugLevel bool
		expectLog  bool
	}{
		{
			name:       "条件真・デバッグレベル",
			condition:  true,
			debugLevel: true,
			expectLog:  true,
		},
		{
			name:       "条件偽・デバッグレベル",
			condition:  false,
			debugLevel: true,
			expectLog:  false,
		},
		{
			name:       "条件真・情報レベル",
			condition:  true,
			debugLevel: false,
			expectLog:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.debugLevel {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}

			buf.Reset()
			fields := map[string]interface{}{"test": "value"}
			logger.LogDebugWithCondition(tt.condition, "Test debug message", fields)

			output := buf.String()
			if tt.expectLog {
				assert.Contains(t, output, "Test debug message")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogStructured(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	tests := []struct {
		name      string
		level     string
		message   string
		fields    map[string]interface{}
		expectLog bool
	}{
		{
			name:      "情報ログ",
			level:     "info",
			message:   "Test info message",
			fields:    map[string]interface{}{"key": "value"},
			expectLog: true,
		},
		{
			name:      "警告ログ",
			level:     "warn",
			message:   "Test warn message",
			fields:    map[string]interface{}{"count": 123},
			expectLog: true,
		},
		{
			name:      "無効レベル（デフォルト）",
			level:     "invalid",
			message:   "Test default message",
			fields:    map[string]interface{}{"enabled": false},
			expectLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			buf.Reset()

			logger.LogStructured(tt.level, tt.message, tt.fields)

			output := buf.String()
			if tt.expectLog {
				assert.Contains(t, output, tt.message)
			}
		})
	}
}

func TestProgressLoggingFunctions(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	details := map[string]interface{}{
		"operation": "test",
		"progress":  50,
	}

	t.Run("LogProgress", func(t *testing.T) {
		buf.Reset()
		logger.LogProgress("Test operation", details)
		output := buf.String()
		assert.Contains(t, output, "🔄 Test operation")
		assert.Contains(t, output, "test")
		assert.Contains(t, output, "50")
	})

	t.Run("LogSuccess", func(t *testing.T) {
		buf.Reset()
		logger.LogSuccess("Test operation", details)
		output := buf.String()
		assert.Contains(t, output, "✅ Test operation")
	})

	t.Run("LogWarning", func(t *testing.T) {
		buf.Reset()
		logger.LogWarning("Test operation", details)
		output := buf.String()
		assert.Contains(t, output, "⚠️ Test operation")
	})

	t.Run("LogError", func(t *testing.T) {
		buf.Reset()
		testErr := fmt.Errorf("test error")
		logger.LogError("Test operation", testErr, details)
		output := buf.String()
		assert.Contains(t, output, "❌ Test operation")
		assert.Contains(t, output, "test error")
	})

	t.Run("LogError with nil error", func(t *testing.T) {
		buf.Reset()
		logger.LogError("Test operation", nil, details)
		output := buf.String()
		assert.Contains(t, output, "❌ Test operation")
	})

	t.Run("LogError with nil details", func(t *testing.T) {
		buf.Reset()
		testErr := fmt.Errorf("test error")
		logger.LogError("Test operation", testErr, nil)
		output := buf.String()
		assert.Contains(t, output, "❌ Test operation")
		assert.Contains(t, output, "test error")
	})
}

func TestTestLoggerIntegration(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalLevel)
	}()

	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

	err := logger.TestLoggerIntegration()
	assert.NoError(t, err)

	// 元のレベルが復元されていることを確認
	assert.Equal(t, originalLevel, zerolog.GlobalLevel())

	// ログ出力の確認
	output := buf.String()
	assert.Contains(t, output, "Testing log level")
	assert.Contains(t, output, "Logger integration test completed successfully")
}

func TestLogSystemInfo(t *testing.T) {
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	logger.LogSystemInfo()

	output := buf.String()
	assert.Contains(t, output, "Logger system initialized")
	assert.Contains(t, output, "log_level")
	assert.Contains(t, output, "debug_enabled")
	assert.Contains(t, output, "time_format")
}

// エッジケースのテスト
func TestEdgeCases(t *testing.T) {
	t.Run("ログレベル境界値テスト", func(t *testing.T) {
		// 非常に低いレベル
		logger.SetLogLevel("debug")
		assert.Equal(t, "debug", logger.GetCurrentLogLevel())
		assert.True(t, logger.IsDebugEnabled())

		// 非常に高いレベル
		logger.SetLogLevel("fatal")
		assert.Equal(t, "fatal", logger.GetCurrentLogLevel())
		assert.False(t, logger.IsDebugEnabled())
	})

	t.Run("空のフィールドマップテスト", func(t *testing.T) {
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		// 空のフィールドマップでログ関数をテスト
		emptyFields := make(map[string]interface{})

		require.NotPanics(t, func() {
			logger.LogWithError(fmt.Errorf("test"), "test message", emptyFields)
		})

		require.NotPanics(t, func() {
			logger.LogDebugWithCondition(true, "test debug", emptyFields)
		})

		require.NotPanics(t, func() {
			logger.LogStructured("info", "test structured", emptyFields)
		})
	})

	t.Run("nil フィールドマップテスト", func(t *testing.T) {
		var buf bytes.Buffer
		log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		require.NotPanics(t, func() {
			logger.LogWithError(fmt.Errorf("test"), "test message", nil)
		})

		require.NotPanics(t, func() {
			logger.LogDebugWithCondition(true, "test debug", nil)
		})

		require.NotPanics(t, func() {
			logger.LogStructured("info", "test structured", nil)
		})
	})
}

// ベンチマークテスト
func BenchmarkInitConsoleLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logger.InitConsoleLogger("info")
	}
}

func BenchmarkSetLogLevel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logger.SetLogLevel("info")
	}
}

func BenchmarkLogStructured(b *testing.B) {
	fields := map[string]interface{}{
		"operation": "benchmark",
		"iteration": 0,
		"enabled":   true,
	}

	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fields["iteration"] = i
		logger.LogStructured("info", "Benchmark test", fields)
	}
}
