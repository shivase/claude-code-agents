package logger

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// StartupPhase manages startup phases
type StartupPhase struct {
	Name      string
	StartTime time.Time
	Details   map[string]interface{}
}

// StartupLogger interface for managing startup logs
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

// DefaultStartupLogger default startup log implementation
type DefaultStartupLogger struct{}

// NewStartupLogger creates new startup logger instance
func NewStartupLogger() StartupLogger {
	return &DefaultStartupLogger{}
}

// LogSystemInit logs system initialization
func (sl *DefaultStartupLogger) LogSystemInit(phase string, details map[string]interface{}) {
	log.Info().
		Str("category", "startup").
		Str("phase", phase).
		Interface("details", details).
		Msg("🚀 System initialization")
}

// LogConfigLoad logs configuration file loading
func (sl *DefaultStartupLogger) LogConfigLoad(configPath string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["config_path"] = configPath

	log.Info().
		Str("category", "startup").
		Str("phase", "config_load").
		Interface("details", details).
		Msg("📋 Configuration file loading")
}

// LogInstructionConfig logs instruction configuration information
func (sl *DefaultStartupLogger) LogInstructionConfig(instructionInfo map[string]interface{}, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	// Integrate instruction information into details
	for key, value := range instructionInfo {
		details[key] = value
	}

	log.Info().
		Str("category", "startup").
		Str("phase", "instruction_config").
		Interface("details", details).
		Msg("📝 Instruction configuration verification")
}

// LogEnvironmentInfo logs environment information
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
		Msg("🔍 Environment information verification")
}

// LogTmuxSetup logs tmux setup
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
		Msg("🖥️  tmux session setup")
}

// LogClaudeStart logs Claude CLI startup
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
		Msg("🤖 Claude CLI startup")
}

// LogStartupComplete logs startup completion
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
		Msg("✅ Startup completed")
}

// LogStartupError logs startup error
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
		Msg("❌ Startup error")
}

// BeginPhase begins startup phase
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
		Msg("🔄 Startup phase began")

	return sp
}

// Complete completes phase
func (sp *StartupPhase) Complete() {
	duration := time.Since(sp.StartTime)

	log.Info().
		Str("category", "startup").
		Str("phase", sp.Name).
		Str("status", "completed").
		Dur("duration", duration).
		Interface("details", sp.Details).
		Msg("✅ Startup phase completed")
}

// CompleteWithError completes phase with error
func (sp *StartupPhase) CompleteWithError(err error) {
	duration := time.Since(sp.StartTime)

	log.Error().
		Str("category", "startup").
		Str("phase", sp.Name).
		Str("status", "failed").
		Dur("duration", duration).
		Interface("details", sp.Details).
		Err(err).
		Msg("❌ Startup phase failed")
}

// Helper functions

// LogStartupProgress displays startup progress
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
		Msg("📊 Startup progress")
}

// LogStartupDebug debug startup log
func LogStartupDebug(phase string, message string, details map[string]interface{}) {
	log.Debug().
		Str("category", "startup").
		Str("phase", phase).
		Interface("details", details).
		Msg(fmt.Sprintf("🔍 %s", message))
}

// LogStartupWarning startup warning log
func LogStartupWarning(phase string, message string, details map[string]interface{}) {
	log.Warn().
		Str("category", "startup").
		Str("phase", phase).
		Interface("details", details).
		Msg(fmt.Sprintf("⚠️  %s", message))
}

// Global variable (for easy access)
var defaultStartupLogger = NewStartupLogger()

// LogSystemInit package-level convenience function
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
