package cmd

// 共通機能とデータ構造の定義

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

// 共通設定構造体
type CommonConfig struct {
	HomeDir         string
	ConfigDir       string
	WorkingDir      string
	ClaudeCLIPath   string
	InstructionsDir string
	LogLevel        string
	Verbose         bool
}

// グローバル設定インスタンス
var globalConfig *CommonConfig

// グローバル設定変数
var (
	globalConfigDir string
	globalLogLevel  = "info"
	globalVerbose   = false
)

// AgentConfig エージェント設定
type AgentConfig struct {
	Name            string
	InstructionFile string
	WorkingDir      string
}

// GetCommonConfig 共通設定の取得
func GetCommonConfig() *CommonConfig {
	if globalConfig == nil {
		globalConfig = &CommonConfig{}
		if err := globalConfig.Initialize(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize common config")
		}
	}
	return globalConfig
}

// Initialize 設定の初期化
func (c *CommonConfig) Initialize() error {
	// ホームディレクトリの設定
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	c.HomeDir = homeDir

	// 統一された設定ディレクトリパス
	claudeDir := filepath.Join(homeDir, ".claude")
	claudCodeAgentsDir := filepath.Join(claudeDir, "claude-code-agents")

	// 設定ディレクトリの設定
	if globalConfigDir != "" {
		c.ConfigDir = globalConfigDir
	} else {
		c.ConfigDir = claudCodeAgentsDir
	}

	// 作業ディレクトリの設定
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	c.WorkingDir = workingDir

	// Claude CLIパスの設定
	c.ClaudeCLIPath = filepath.Join(claudeDir, "local", "claude")

	// インストラクションディレクトリの設定
	c.InstructionsDir = filepath.Join(claudCodeAgentsDir, "instructions")

	// ログレベルの設定
	c.LogLevel = globalLogLevel
	c.Verbose = globalVerbose

	return nil
}

// GetConfigPath 設定ファイルパスの取得
func (c *CommonConfig) GetConfigPath() string {
	return filepath.Join(c.ConfigDir, "manager.json")
}

// GetTeamConfigPath チーム設定ファイルパスの取得
func (c *CommonConfig) GetTeamConfigPath() string {
	return filepath.Join(c.ConfigDir, "agents.conf")
}

// GetSessionName tmuxセッション名の取得（動的検出）
func (c *CommonConfig) GetSessionName() string {
	// tmuxManagerを使用してアクティブなAIセッションを検出
	tmuxManager := tmux.NewTmuxManager("")
	if sessionName, err := tmuxManager.FindDefaultAISession(6); err == nil {
		return sessionName
	}

	// 検出できない場合はデフォルト値を返す
	return "ai-teams"
}

// GetLogPath ログファイルパスの取得
func (c *CommonConfig) GetLogPath() string {
	return filepath.Join(c.ConfigDir, "logs", "manager.log")
}

// 共通のエラーハンドリング

// CommandError コマンドエラー構造体
type CommandError struct {
	Command string
	Err     error
	Code    int
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command '%s' failed: %v (exit code: %d)", e.Command, e.Err, e.Code)
}

// WrapError エラーのラップ
func WrapError(command string, err error, code int) *CommandError {
	return &CommandError{
		Command: command,
		Err:     err,
		Code:    code,
	}
}

// 共通のバリデーション関数

// ValidateAgentName エージェント名の検証
func ValidateAgentName(agentName string) error {
	validAgents := map[string]bool{
		"po":      true,
		"manager": true,
		"dev1":    true,
		"dev2":    true,
		"dev3":    true,
		"dev4":    true,
	}

	if !validAgents[agentName] {
		return fmt.Errorf("invalid agent name '%s'. Valid agents: po, manager, dev1, dev2, dev3, dev4", agentName)
	}

	return nil
}

// ValidateMessage メッセージの検証
func ValidateMessage(message string) error {
	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	if len(message) > 4096 {
		return fmt.Errorf("message too long (max 4096 characters)")
	}

	return nil
}

// ValidateConfig 設定の検証
func ValidateConfig() error {
	config := GetCommonConfig()

	// Claude CLIの存在確認
	if _, err := os.Stat(config.ClaudeCLIPath); os.IsNotExist(err) {
		return fmt.Errorf("claude CLI not found at %s", config.ClaudeCLIPath)
	}

	// インストラクションディレクトリの存在確認
	if _, err := os.Stat(config.InstructionsDir); os.IsNotExist(err) {
		return fmt.Errorf("instructions directory not found at %s", config.InstructionsDir)
	}

	return nil
}

// 共通のヘルパー関数

// EnsureDir ディレクトリの存在確認と作成
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0750)
	}
	return nil
}

// IsProcessRunning プロセスの実行状態確認
func IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// プロセスにシグナル0を送信して存在確認（Unix系のみ）
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// FormatDuration 期間のフォーマット
func FormatDuration(duration time.Duration) string {
	switch {
	case duration < time.Millisecond:
		return fmt.Sprintf("%.2fμs", float64(duration.Nanoseconds())/1000)
	case duration < time.Second:
		return fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1e6)
	default:
		return fmt.Sprintf("%.2fs", duration.Seconds())
	}
}

// TruncateString 文字列の切り詰め
func TruncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}

	if maxLen <= 3 {
		return str[:maxLen]
	}

	return str[:maxLen-3] + "..."
}

// 共通のステータス管理

// SystemStatus システムステータス
type SystemStatus struct {
	IsRunning     bool            `json:"is_running"`
	StartTime     time.Time       `json:"start_time"`
	Uptime        time.Duration   `json:"uptime"`
	AgentStatuses map[string]bool `json:"agent_statuses"`
	LastUpdate    time.Time       `json:"last_update"`
	Version       string          `json:"version"`
	PID           int             `json:"pid"`
}

// NewSystemStatus 新しいシステムステータスの作成
func NewSystemStatus() *SystemStatus {
	return &SystemStatus{
		IsRunning:     false,
		AgentStatuses: make(map[string]bool),
		LastUpdate:    time.Now(),
		Version:       "1.0.0",
		PID:           os.Getpid(),
	}
}

// UpdateAgentStatus エージェントステータスの更新
func (s *SystemStatus) UpdateAgentStatus(agentName string, isRunning bool) {
	s.AgentStatuses[agentName] = isRunning
	s.LastUpdate = time.Now()
}

// GetAgentStatus エージェントステータスの取得
func (s *SystemStatus) GetAgentStatus(agentName string) bool {
	return s.AgentStatuses[agentName]
}

// GetAllAgentStatuses 全エージェントステータスの取得
func (s *SystemStatus) GetAllAgentStatuses() map[string]bool {
	statuses := make(map[string]bool)
	for agent, status := range s.AgentStatuses {
		statuses[agent] = status
	}

	return statuses
}

// Start システム開始
func (s *SystemStatus) Start() {
	s.IsRunning = true
	s.StartTime = time.Now()
	s.LastUpdate = time.Now()
}

// Stop システム停止
func (s *SystemStatus) Stop() {
	s.IsRunning = false
	s.LastUpdate = time.Now()
}

// GetUptime 稼働時間の取得
func (s *SystemStatus) GetUptime() time.Duration {
	if !s.IsRunning {
		return 0
	}

	return time.Since(s.StartTime)
}

// 共通のメッセージフォーマット

// FormatAgentStatus エージェントステータスのフォーマット
func FormatAgentStatus(agentName string, isRunning bool) string {
	statusIcon := "❌"
	statusText := "停止中"

	if isRunning {
		statusIcon = "✅"
		statusText = "実行中"
	}

	return fmt.Sprintf("%s %s: %s", statusIcon, agentName, statusText)
}

// FormatMessage メッセージのフォーマット
func FormatMessage(sender, recipient, message string) string {
	timestamp := time.Now().Format("15:04:05")
	return fmt.Sprintf("[%s] %s -> %s: %s", timestamp, sender, recipient, message)
}

// FormatError エラーのフォーマット
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	return fmt.Sprintf("❌ エラー: %v", err)
}

// FormatSuccess 成功メッセージのフォーマット
func FormatSuccess(message string) string {
	return fmt.Sprintf("✅ %s", message)
}

// FormatWarning 警告メッセージのフォーマット
func FormatWarning(message string) string {
	return fmt.Sprintf("⚠️ %s", message)
}

// FormatInfo 情報メッセージのフォーマット
func FormatInfo(message string) string {
	return fmt.Sprintf("ℹ️ %s", message)
}

// 共通のリソース管理

// ResourceManager リソース管理
type ResourceManager struct {
	cleanupFuncs []func() error
}

// NewResourceManager 新しいリソース管理の作成
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		cleanupFuncs: make([]func() error, 0),
	}
}

// AddCleanup クリーンアップ関数の追加
func (rm *ResourceManager) AddCleanup(cleanup func() error) {
	rm.cleanupFuncs = append(rm.cleanupFuncs, cleanup)
}

// Cleanup 全リソースのクリーンアップ
func (rm *ResourceManager) Cleanup() error {
	var errors []error

	// 逆順でクリーンアップ実行
	for i := len(rm.cleanupFuncs) - 1; i >= 0; i-- {
		if err := rm.cleanupFuncs[i](); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// グローバルリソース管理インスタンス
var globalResourceManager *ResourceManager

// GetResourceManager グローバルリソース管理の取得
func GetResourceManager() *ResourceManager {
	if globalResourceManager == nil {
		globalResourceManager = NewResourceManager()
	}
	return globalResourceManager
}
