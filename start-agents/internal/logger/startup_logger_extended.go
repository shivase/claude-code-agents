package logger

import (
	"time"

	"github.com/rs/zerolog/log"
)

// InstructionLogger InstructionSender インストラクション送信ログのインターフェース
type InstructionLogger interface {
	LogInstructionLoad(instructionsDir string, details map[string]interface{})
	LogInstructionSend(agent string, instructionFile string, details map[string]interface{})
	LogInstructionProgress(agent string, status string, progress map[string]interface{})
	LogInstructionError(agent string, instructionFile string, err error, recovery map[string]interface{})
	LogInstructionBatch(agents []string, details map[string]interface{})
	BeginInstructionPhase(batchName string, agentCount int) *InstructionPhase
}

// InstructionPhase インストラクション送信フェーズの管理構造体
type InstructionPhase struct {
	BatchName  string
	StartTime  time.Time
	AgentCount int
	Completed  int
	Failed     int
	Details    map[string]interface{}
}

// ExtendedStartupLogger 拡張された起動ログ実装
type ExtendedStartupLogger struct {
	*DefaultStartupLogger
}

// NewExtendedStartupLogger 拡張起動ログインスタンスを作成
func NewExtendedStartupLogger() *ExtendedStartupLogger {
	return &ExtendedStartupLogger{
		DefaultStartupLogger: &DefaultStartupLogger{},
	}
}

// LogInstructionLoad インストラクションディレクトリ読み込みログ
func (esl *ExtendedStartupLogger) LogInstructionLoad(instructionsDir string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["instructions_dir"] = instructionsDir

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_load").
		Interface("details", details).
		Msg("📋 インストラクションディレクトリ読み込み")
}

// LogInstructionSend 個別インストラクション送信ログ
func (esl *ExtendedStartupLogger) LogInstructionSend(agent string, instructionFile string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["agent"] = agent
	details["instruction_file"] = instructionFile

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_send").
		Interface("details", details).
		Msg("📤 インストラクション送信")
}

// LogInstructionProgress インストラクション送信進捗ログ
func (esl *ExtendedStartupLogger) LogInstructionProgress(agent string, status string, progress map[string]interface{}) {
	if progress == nil {
		progress = make(map[string]interface{})
	}
	progress["agent"] = agent
	progress["status"] = status

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_progress").
		Interface("details", progress).
		Msg("📊 インストラクション送信進捗")
}

// LogInstructionError インストラクション送信エラーログ
func (esl *ExtendedStartupLogger) LogInstructionError(agent string, instructionFile string, err error, recovery map[string]interface{}) {
	fields := map[string]interface{}{
		"category":         "startup",
		"phase":            "instruction_send",
		"agent":            agent,
		"instruction_file": instructionFile,
		"error":            err.Error(),
	}

	if recovery != nil {
		fields["recovery_info"] = recovery
	}

	log.Error().
		Interface("details", fields).
		Err(err).
		Msg("❌ インストラクション送信エラー")
}

// LogInstructionBatch バッチインストラクション送信ログ
func (esl *ExtendedStartupLogger) LogInstructionBatch(agents []string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["agents"] = agents
	details["agent_count"] = len(agents)

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_batch").
		Interface("details", details).
		Msg("📦 バッチインストラクション送信開始")
}

// BeginInstructionPhase インストラクション送信フェーズ開始
func (esl *ExtendedStartupLogger) BeginInstructionPhase(batchName string, agentCount int) *InstructionPhase {
	ip := &InstructionPhase{
		BatchName:  batchName,
		StartTime:  time.Now(),
		AgentCount: agentCount,
		Completed:  0,
		Failed:     0,
		Details: map[string]interface{}{
			"batch_name":  batchName,
			"agent_count": agentCount,
		},
	}

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_batch").
		Str("status", "started").
		Interface("details", ip.Details).
		Msg("🔄 インストラクション送信バッチ開始")

	return ip
}

// RecordSuccess 成功記録
func (ip *InstructionPhase) RecordSuccess(agent string, instructionFile string) {
	ip.Completed++

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_batch").
		Str("status", "agent_completed").
		Str("agent", agent).
		Str("instruction_file", instructionFile).
		Int("completed", ip.Completed).
		Int("remaining", ip.AgentCount-ip.Completed-ip.Failed).
		Msg("✅ エージェントインストラクション送信完了")
}

// RecordFailure 失敗記録
func (ip *InstructionPhase) RecordFailure(agent string, instructionFile string, err error) {
	ip.Failed++

	log.Error().
		Str("category", "startup").
		Str("phase", "instruction_batch").
		Str("status", "agent_failed").
		Str("agent", agent).
		Str("instruction_file", instructionFile).
		Int("failed", ip.Failed).
		Int("remaining", ip.AgentCount-ip.Completed-ip.Failed).
		Err(err).
		Msg("❌ エージェントインストラクション送信失敗")
}

// Complete バッチ完了
func (ip *InstructionPhase) Complete() {
	duration := time.Since(ip.StartTime)
	successRate := float64(ip.Completed) / float64(ip.AgentCount) * 100

	finalDetails := map[string]interface{}{
		"batch_name":   ip.BatchName,
		"agent_count":  ip.AgentCount,
		"completed":    ip.Completed,
		"failed":       ip.Failed,
		"success_rate": successRate,
		"duration":     duration.String(),
		"duration_ms":  duration.Milliseconds(),
	}

	if ip.Failed > 0 {
		log.Warn().
			Str("category", "startup").
			Str("phase", "instruction_batch").
			Str("status", "completed_with_errors").
			Interface("details", finalDetails).
			Msg("⚠️ インストラクション送信バッチ完了（一部失敗）")
	} else {
		log.Info().
			Str("category", "startup").
			Str("phase", "instruction_batch").
			Str("status", "completed").
			Interface("details", finalDetails).
			Msg("✅ インストラクション送信バッチ完了")
	}
}

// CompleteWithError バッチエラー完了
func (ip *InstructionPhase) CompleteWithError(err error) {
	duration := time.Since(ip.StartTime)

	finalDetails := map[string]interface{}{
		"batch_name":  ip.BatchName,
		"agent_count": ip.AgentCount,
		"completed":   ip.Completed,
		"failed":      ip.Failed,
		"duration":    duration.String(),
		"error":       err.Error(),
	}

	log.Error().
		Str("category", "startup").
		Str("phase", "instruction_batch").
		Str("status", "failed").
		Interface("details", finalDetails).
		Err(err).
		Msg("❌ インストラクション送信バッチ失敗")
}

// ConfigIntegration 設定統合ログ機能
type ConfigIntegration interface {
	LogConfigValidation(configPath string, validationResults map[string]interface{})
	LogConfigMerge(sources []string, mergeResults map[string]interface{})
	LogConfigSchema(schemaVersion string, schemaDetails map[string]interface{})
	LogConfigBackwardCompatibility(version string, compatibilityInfo map[string]interface{})
}

// LogConfigValidation 設定検証ログ
func (esl *ExtendedStartupLogger) LogConfigValidation(configPath string, validationResults map[string]interface{}) {
	if validationResults == nil {
		validationResults = make(map[string]interface{})
	}
	validationResults["config_path"] = configPath

	log.Info().
		Str("category", "startup").
		Str("phase", "config_validation").
		Interface("details", validationResults).
		Msg("🔍 設定ファイル検証")
}

// LogConfigMerge 設定マージログ
func (esl *ExtendedStartupLogger) LogConfigMerge(sources []string, mergeResults map[string]interface{}) {
	if mergeResults == nil {
		mergeResults = make(map[string]interface{})
	}
	mergeResults["sources"] = sources
	mergeResults["source_count"] = len(sources)

	log.Info().
		Str("category", "startup").
		Str("phase", "config_merge").
		Interface("details", mergeResults).
		Msg("🔗 設定ファイルマージ")
}

// LogConfigSchema 設定スキーマログ
func (esl *ExtendedStartupLogger) LogConfigSchema(schemaVersion string, schemaDetails map[string]interface{}) {
	if schemaDetails == nil {
		schemaDetails = make(map[string]interface{})
	}
	schemaDetails["schema_version"] = schemaVersion

	log.Info().
		Str("category", "startup").
		Str("phase", "config_schema").
		Interface("details", schemaDetails).
		Msg("📋 設定スキーマ適用")
}

// LogConfigBackwardCompatibility 後方互換性ログ
func (esl *ExtendedStartupLogger) LogConfigBackwardCompatibility(version string, compatibilityInfo map[string]interface{}) {
	if compatibilityInfo == nil {
		compatibilityInfo = make(map[string]interface{})
	}
	compatibilityInfo["compatibility_version"] = version

	log.Info().
		Str("category", "startup").
		Str("phase", "config_compatibility").
		Interface("details", compatibilityInfo).
		Msg("🔄 後方互換性処理")
}
