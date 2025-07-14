package logger

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// StartupPhase 起動フェーズの管理構造体
type StartupPhase struct {
	Name      string
	StartTime time.Time
	Details   map[string]interface{}
}

// StartupLogger 起動ログの管理インターフェース
type StartupLogger interface {
	LogSystemInit(phase string, details map[string]interface{})
	LogConfigLoad(configPath string, details map[string]interface{})
	LogInstructionConfig(instructionInfo map[string]interface{}, details map[string]interface{})
	LogEnvironmentInfo(envInfo map[string]interface{}, debugMode bool)
	LogTmuxSetup(sessionName string, paneCount int, details map[string]interface{})
	LogClaudeStart(agent string, paneID string, details map[string]interface{})
	LogStartupComplete(totalTime time.Duration, details map[string]interface{})
	LogStartupError(phase string, err error, recovery map[string]interface{})
	BeginPhase(phase string, details map[string]interface{}) *StartupPhase
}

// DefaultStartupLogger デフォルトの起動ログ実装
type DefaultStartupLogger struct{}

// NewStartupLogger 新しい起動ログインスタンスを作成
func NewStartupLogger() StartupLogger {
	return &DefaultStartupLogger{}
}

// LogSystemInit システム初期化ログ
func (sl *DefaultStartupLogger) LogSystemInit(phase string, details map[string]interface{}) {
	log.Info().
		Str("category", "startup").
		Str("phase", phase).
		Interface("details", details).
		Msg("🚀 システム初期化")
}

// LogConfigLoad 設定ファイル読み込みログ
func (sl *DefaultStartupLogger) LogConfigLoad(configPath string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["config_path"] = configPath

	log.Info().
		Str("category", "startup").
		Str("phase", "config_load").
		Interface("details", details).
		Msg("📋 設定ファイル読み込み")
}

// LogInstructionConfig instruction設定情報ログ
func (sl *DefaultStartupLogger) LogInstructionConfig(instructionInfo map[string]interface{}, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	// instruction情報をdetailsに統合
	for key, value := range instructionInfo {
		details[key] = value
	}

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_config").
		Interface("details", details).
		Msg("📝 instruction設定確認")
}

// LogEnvironmentInfo 環境情報ログ
func (sl *DefaultStartupLogger) LogEnvironmentInfo(envInfo map[string]interface{}, debugMode bool) {
	logLevel := log.Info()
	if debugMode {
		logLevel = log.Debug()
	}

	logLevel.
		Str("category", "startup").
		Str("phase", "environment_check").
		Bool("debug_mode", debugMode).
		Interface("environment", envInfo).
		Msg("🔍 環境情報確認")
}

// LogTmuxSetup tmuxセットアップログ
func (sl *DefaultStartupLogger) LogTmuxSetup(sessionName string, paneCount int, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["session_name"] = sessionName
	details["pane_count"] = paneCount

	log.Info().
		Str("category", "startup").
		Str("phase", "tmux_setup").
		Interface("details", details).
		Msg("🖥️  tmuxセッション設定")
}

// LogClaudeStart Claude CLI起動ログ
func (sl *DefaultStartupLogger) LogClaudeStart(agent string, paneID string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["agent"] = agent
	details["pane_id"] = paneID

	log.Info().
		Str("category", "startup").
		Str("phase", "claude_start").
		Interface("details", details).
		Msg("🤖 Claude CLI起動")
}

// LogStartupComplete 起動完了ログ
func (sl *DefaultStartupLogger) LogStartupComplete(totalTime time.Duration, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["total_time"] = totalTime.String()
	details["total_time_ms"] = totalTime.Milliseconds()

	log.Info().
		Str("category", "startup").
		Str("phase", "complete").
		Interface("details", details).
		Msg("✅ 起動完了")
}

// LogStartupError 起動エラーログ
func (sl *DefaultStartupLogger) LogStartupError(phase string, err error, recovery map[string]interface{}) {
	fields := map[string]interface{}{
		"category": "startup",
		"phase":    phase,
		"error":    err.Error(),
	}

	if recovery != nil {
		fields["recovery_info"] = recovery
	}

	log.Error().
		Interface("details", fields).
		Err(err).
		Msg("❌ 起動エラー")
}

// BeginPhase 起動フェーズ開始
func (sl *DefaultStartupLogger) BeginPhase(phase string, details map[string]interface{}) *StartupPhase {
	sp := &StartupPhase{
		Name:      phase,
		StartTime: time.Now(),
		Details:   details,
	}

	log.Info().
		Str("category", "startup").
		Str("phase", phase).
		Str("status", "started").
		Interface("details", details).
		Msg("🔄 起動フェーズ開始")

	return sp
}

// Complete フェーズ完了
func (sp *StartupPhase) Complete() {
	duration := time.Since(sp.StartTime)

	log.Info().
		Str("category", "startup").
		Str("phase", sp.Name).
		Str("status", "completed").
		Dur("duration", duration).
		Interface("details", sp.Details).
		Msg("✅ 起動フェーズ完了")
}

// CompleteWithError フェーズエラー完了
func (sp *StartupPhase) CompleteWithError(err error) {
	duration := time.Since(sp.StartTime)

	log.Error().
		Str("category", "startup").
		Str("phase", sp.Name).
		Str("status", "failed").
		Dur("duration", duration).
		Interface("details", sp.Details).
		Err(err).
		Msg("❌ 起動フェーズ失敗")
}

// ヘルパー関数

// LogStartupProgress 起動プログレス表示
func LogStartupProgress(phase string, progress int, total int) {
	percentage := float64(progress) / float64(total) * 100

	fields := map[string]interface{}{
		"progress":   progress,
		"total":      total,
		"percentage": fmt.Sprintf("%.1f%%", percentage),
	}

	log.Info().
		Str("category", "startup").
		Str("phase", phase).
		Interface("details", fields).
		Msg("📊 起動プログレス")
}

// LogStartupDebug デバッグ用起動ログ
func LogStartupDebug(phase string, message string, details map[string]interface{}) {
	log.Debug().
		Str("category", "startup").
		Str("phase", phase).
		Interface("details", details).
		Msg(fmt.Sprintf("🔍 %s", message))
}

// LogStartupWarning 起動警告ログ
func LogStartupWarning(phase string, message string, details map[string]interface{}) {
	log.Warn().
		Str("category", "startup").
		Str("phase", phase).
		Interface("details", details).
		Msg(fmt.Sprintf("⚠️  %s", message))
}

// グローバル変数（簡単なアクセス用）
var defaultStartupLogger = NewStartupLogger()

// LogSystemInit パッケージレベルの便利関数
func LogSystemInit(phase string, details map[string]interface{}) {
	defaultStartupLogger.LogSystemInit(phase, details)
}

func LogConfigLoad(configPath string, details map[string]interface{}) {
	defaultStartupLogger.LogConfigLoad(configPath, details)
}

func LogInstructionConfig(instructionInfo map[string]interface{}, details map[string]interface{}) {
	defaultStartupLogger.LogInstructionConfig(instructionInfo, details)
}

func LogEnvironmentInfo(envInfo map[string]interface{}, debugMode bool) {
	defaultStartupLogger.LogEnvironmentInfo(envInfo, debugMode)
}

func LogTmuxSetup(sessionName string, paneCount int, details map[string]interface{}) {
	defaultStartupLogger.LogTmuxSetup(sessionName, paneCount, details)
}

func LogClaudeStart(agent string, paneID string, details map[string]interface{}) {
	defaultStartupLogger.LogClaudeStart(agent, paneID, details)
}

func LogStartupComplete(totalTime time.Duration, details map[string]interface{}) {
	defaultStartupLogger.LogStartupComplete(totalTime, details)
}

func LogStartupError(phase string, err error, recovery map[string]interface{}) {
	defaultStartupLogger.LogStartupError(phase, err, recovery)
}

func BeginPhase(phase string, details map[string]interface{}) *StartupPhase {
	return defaultStartupLogger.BeginPhase(phase, details)
}
