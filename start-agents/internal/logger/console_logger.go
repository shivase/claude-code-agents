package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitConsoleLogger zerologをConsoleWriterで初期化
func InitConsoleLogger(level string) {
	// 構造化ログ表示用のカスタムConsoleWriterの設定
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		NoColor:    false,
		FormatLevel: func(i interface{}) string {
			level := strings.ToUpper(fmt.Sprintf("%s", i))
			switch level {
			case "INFO":
				return "\x1b[32m[INFO]\x1b[0m"
			case "WARN":
				return "\x1b[33m[WARN]\x1b[0m"
			case "ERROR":
				return "\x1b[91m[ERROR]\x1b[0m"
			case "DEBUG":
				return "\x1b[36m[DEBUG]\x1b[0m"
			case "FATAL":
				return "\x1b[35m[FATAL]\x1b[0m"
			default:
				return fmt.Sprintf("[%s]", level)
			}
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("\n  📋 \x1b[1m%s\x1b[0m", i)
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("\n     ├─ \x1b[34m%s\x1b[0m:", i)
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf(" \x1b[37m%s\x1b[0m", i)
		},
		FormatTimestamp: func(i interface{}) string {
			return fmt.Sprintf("[%s]", i)
		},
	}

	// ログレベルの設定
	var logLevel zerolog.Level
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	// グローバルレベルを設定
	zerolog.SetGlobalLevel(logLevel)

	// グローバルロガーの設定
	log.Logger = zerolog.New(consoleWriter).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	// タイムスタンプ設定
	zerolog.TimeFieldFormat = time.RFC3339
}

// SetLogLevel ログレベルの動的変更
func SetLogLevel(level string) {
	var logLevel zerolog.Level
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)
}

// InitWithDebugFlag --debugフラグ対応の初期化
func InitWithDebugFlag(debugEnabled bool, baseLogLevel string) {
	var finalLogLevel string

	if debugEnabled {
		finalLogLevel = "debug"
		fmt.Printf("[DEBUG] Debug mode enabled - log level set to debug\n")
	} else {
		finalLogLevel = baseLogLevel
	}

	InitConsoleLogger(finalLogLevel)

	if debugEnabled {
		log.Debug().Msg("Console logger initialized with debug mode")
	}
}

// GetCurrentLogLevel 現在のログレベルを取得
func GetCurrentLogLevel() string {
	level := zerolog.GlobalLevel()
	switch level {
	case zerolog.DebugLevel:
		return "debug"
	case zerolog.InfoLevel:
		return "info"
	case zerolog.WarnLevel:
		return "warn"
	case zerolog.ErrorLevel:
		return "error"
	case zerolog.FatalLevel:
		return "fatal"
	default:
		return "info"
	}
}

// IsDebugEnabled デバッグモードが有効かチェック
func IsDebugEnabled() bool {
	return zerolog.GlobalLevel() <= zerolog.DebugLevel
}

// LogSystemInfo システム情報をログ出力
func LogSystemInfo() {
	log.Info().
		Str("log_level", GetCurrentLogLevel()).
		Bool("debug_enabled", IsDebugEnabled()).
		Str("time_format", "RFC3339").
		Msg("Logger system initialized")
}

// ApplyConfigLogLevel 設定ファイルからログレベルを適用
func ApplyConfigLogLevel(configLogLevel string, debugOverride bool) {
	if debugOverride {
		SetLogLevel("debug")
		log.Debug().Msg("Debug override applied - config log level ignored")
		return
	}

	if configLogLevel != "" {
		SetLogLevel(configLogLevel)
		log.Info().
			Str("config_log_level", configLogLevel).
			Str("applied_log_level", GetCurrentLogLevel()).
			Msg("Config log level applied")
	}
}

// ValidateLogLevel ログレベルの妥当性チェック
func ValidateLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}

// GetLogLevelPriority ログレベルの優先度を取得（デバッグ用）
func GetLogLevelPriority() int {
	level := zerolog.GlobalLevel()
	switch level {
	case zerolog.DebugLevel:
		return 0
	case zerolog.InfoLevel:
		return 1
	case zerolog.WarnLevel:
		return 2
	case zerolog.ErrorLevel:
		return 3
	case zerolog.FatalLevel:
		return 4
	default:
		return 1 // default to info
	}
}

// LogWithError エラーログと詳細情報の統合出力
func LogWithError(err error, message string, fields map[string]interface{}) {
	logEvent := log.Error().Err(err)

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			logEvent = logEvent.Str(key, v)
		case int:
			logEvent = logEvent.Int(key, v)
		case bool:
			logEvent = logEvent.Bool(key, v)
		case time.Duration:
			logEvent = logEvent.Dur(key, v)
		default:
			logEvent = logEvent.Interface(key, v)
		}
	}

	logEvent.Msg(message)
}

// LogDebugWithCondition 条件付きデバッグログ
func LogDebugWithCondition(condition bool, message string, fields map[string]interface{}) {
	if !condition || !IsDebugEnabled() {
		return
	}

	logEvent := log.Debug()

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			logEvent = logEvent.Str(key, v)
		case int:
			logEvent = logEvent.Int(key, v)
		case bool:
			logEvent = logEvent.Bool(key, v)
		case time.Duration:
			logEvent = logEvent.Dur(key, v)
		default:
			logEvent = logEvent.Interface(key, v)
		}
	}

	logEvent.Msg(message)
}

// LogStructured 構造化ログ出力ヘルパー
func LogStructured(level string, message string, fields map[string]interface{}) {
	var logEvent *zerolog.Event

	switch level {
	case "debug":
		logEvent = log.Debug()
	case "info":
		logEvent = log.Info()
	case "warn":
		logEvent = log.Warn()
	case "error":
		logEvent = log.Error()
	case "fatal":
		logEvent = log.Fatal()
	default:
		logEvent = log.Info()
	}

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			logEvent = logEvent.Str(key, v)
		case int:
			logEvent = logEvent.Int(key, v)
		case bool:
			logEvent = logEvent.Bool(key, v)
		case time.Duration:
			logEvent = logEvent.Dur(key, v)
		default:
			logEvent = logEvent.Interface(key, v)
		}
	}

	logEvent.Msg(message)
}

// LogProgress プログレス表示用構造化ログ
func LogProgress(operation string, details map[string]interface{}) {
	LogStructured("info", fmt.Sprintf("🔄 %s", operation), details)
}

// LogSuccess 成功表示用構造化ログ
func LogSuccess(operation string, details map[string]interface{}) {
	LogStructured("info", fmt.Sprintf("✅ %s", operation), details)
}

// LogWarning 警告表示用構造化ログ
func LogWarning(operation string, details map[string]interface{}) {
	LogStructured("warn", fmt.Sprintf("⚠️ %s", operation), details)
}

// LogError エラー表示用構造化ログ
func LogError(operation string, err error, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	if err != nil {
		details["error"] = err.Error()
	}
	LogStructured("error", fmt.Sprintf("❌ %s", operation), details)
}

// TestLoggerIntegration ログ統合テスト
func TestLoggerIntegration() error {
	originalLevel := GetCurrentLogLevel()

	// 各ログレベルのテスト
	testLevels := []string{"debug", "info", "warn", "error"}

	for _, testLevel := range testLevels {
		SetLogLevel(testLevel)

		log.Info().
			Str("test_level", testLevel).
			Str("current_level", GetCurrentLogLevel()).
			Msg("Testing log level")

		if GetCurrentLogLevel() != testLevel {
			return fmt.Errorf("log level test failed: expected %s, got %s", testLevel, GetCurrentLogLevel())
		}
	}

	// 元のレベルに戻す
	SetLogLevel(originalLevel)

	log.Info().Msg("Logger integration test completed successfully")
	return nil
}
