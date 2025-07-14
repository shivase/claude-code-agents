package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/shivase/claude-code-agents/internal/logger"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

var (
	// GlobalLogLevel グローバル設定フラグ
	GlobalLogLevel string

	// MainInitialized 初期化管理
	MainInitialized   bool
	LoggerInitialized bool
)

// InitializeMainSystem メインシステムの初期化処理統合化
func InitializeMainSystem(logLevel string) {
	// 既に初期化されている場合はスキップ
	if MainInitialized {
		return
	}

	// ログシステムの初期化
	GlobalLogLevel = logLevel
	InitLogger()

	// 初期化フラグを設定
	MainInitialized = true
}

// InitLogger ログシステムの初期化
func InitLogger() {
	if LoggerInitialized {
		return
	}

	_, err := zerolog.ParseLevel(GlobalLogLevel)
	if err != nil {
		// fallback to info level if parsing fails
		GlobalLogLevel = "info"
	}

	logger.InitConsoleLogger(GlobalLogLevel)

	log.Info().Msg("Commands package initialized: Command functions are available")

	LoggerInitialized = true
}

// IsValidSessionName セッション名のバリデーション
func IsValidSessionName(name string) bool {
	if name == "" {
		return false
	}

	for _, char := range name {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '-' && char != '_' {
			return false
		}
	}
	return true
}

// ListAISessions セッション一覧表示機能
func ListAISessions() error {
	fmt.Println("🤖 セッション一覧")
	fmt.Println("==================================")

	tmuxManager := tmux.NewTmuxManager("ai-teams")
	sessions, err := tmuxManager.ListSessions()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("📭 現在起動中のAIチームセッションはありません")
			return nil
		}
		return fmt.Errorf("tmuxセッション取得エラー: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("📭 現在起動中のセッションはありません")
	} else {
		fmt.Printf("🚀 起動中のセッション: %d個\n", len(sessions))
		for i, session := range sessions {
			fmt.Printf("  %d. %s\n", i+1, session)
		}
	}

	return nil
}

// DeleteAISession 指定したセッションを削除
func DeleteAISession(sessionName string) error {
	if sessionName == "" {
		fmt.Println("❌ エラー: 削除するセッション名を指定してください")
		fmt.Println("使用方法: ./claude-code-agents --delete [セッション名]")
		fmt.Println("セッション一覧: ./claude-code-agents --list")
		return fmt.Errorf("セッション名が指定されていません")
	}

	fmt.Printf("🗑️ セッション削除: %s\n", sessionName)

	tmuxManager := tmux.NewTmuxManager(sessionName)
	if !tmuxManager.SessionExists(sessionName) {
		fmt.Printf("⚠️ セッション '%s' は存在しません\n", sessionName)
		return nil
	}

	if err := tmuxManager.KillSession(sessionName); err != nil {
		return fmt.Errorf("セッション削除エラー: %w", err)
	}

	fmt.Printf("✅ セッション '%s' を削除しました\n", sessionName)
	return nil
}

// DeleteAllAISessions 全セッションを削除
func DeleteAllAISessions() error {
	fmt.Println("🗑️ 全AIチームセッション削除")
	fmt.Println("==============================")

	tmuxManager := tmux.NewTmuxManager("ai-teams")
	sessions, err := tmuxManager.ListSessions()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("📭 現在起動中のAIチームセッションはありません")
			return nil
		}
		return fmt.Errorf("tmuxセッション取得エラー: %w", err)
	}

	var aiSessions []string

	// AIチーム関連のセッションを抽出
	for _, session := range sessions {
		if strings.Contains(session, "ai-") || strings.Contains(session, "claude-") ||
			strings.Contains(session, "dev-") || strings.Contains(session, "agent-") {
			aiSessions = append(aiSessions, session)
		}
	}

	if len(aiSessions) == 0 {
		fmt.Println("📭 削除対象のAIチームセッションはありません")
		return nil
	}

	fmt.Printf("🎯 削除対象セッション: %d個\n", len(aiSessions))
	for i, session := range aiSessions {
		fmt.Printf("  %d. %s\n", i+1, session)
	}

	// 各セッションを削除
	deletedCount := 0
	for _, session := range aiSessions {
		sessionManager := tmux.NewTmuxManager(session)
		if err := sessionManager.KillSession(session); err != nil {
			fmt.Printf("⚠️ セッション '%s' の削除に失敗: %v\n", session, err)
		} else {
			deletedCount++
			fmt.Printf("✅ セッション '%s' を削除しました\n", session)
		}
	}

	fmt.Printf("\n🎉 %d個のAIチームセッションを削除しました\n", deletedCount)
	return nil
}

// LaunchSystem システム起動機能
func LaunchSystem(sessionName string) error {
	fmt.Printf("🚀 システム起動: %s\n", sessionName)

	// 設定ファイルの読み込み
	configPath := config.GetDefaultTeamConfigPath()
	configLoader := config.NewTeamConfigLoader(configPath)
	teamConfig, err := configLoader.LoadTeamConfig()
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	// 設定読み込み完了ログ
	logger.LogConfigLoad(configPath, map[string]interface{}{
		"config_path":      configPath,
		"dev_count":        teamConfig.DevCount,
		"session_name":     teamConfig.SessionName,
		"instructions_dir": teamConfig.InstructionsDir,
	})

	// instructionファイル情報を収集・表示
	instructionInfo := gatherInstructionInfo(teamConfig)
	envInfo := gatherEnvironmentInfo(teamConfig)

	// instruction設定情報ログ
	logger.LogInstructionConfig(instructionInfo, map[string]interface{}{
		"config_loaded": true,
		"role_count":    len(instructionInfo),
	})

	// 環境情報ログ（デバッグモードをグローバルから取得）
	debugMode := GlobalLogLevel == "debug"
	logger.LogEnvironmentInfo(envInfo, debugMode)

	// tmux管理の基本動作
	tmuxManager := tmux.NewTmuxManager(sessionName)

	// 既存セッションの確認
	if tmuxManager.SessionExists(sessionName) {
		fmt.Printf("🔄 既存セッション '%s' に接続します\n", sessionName)
		return tmuxManager.AttachSession(sessionName)
	}

	// 新しいセッションを作成
	fmt.Printf("📝 新しいセッション '%s' を作成します\n", sessionName)
	if err := tmuxManager.CreateSession(sessionName); err != nil {
		return fmt.Errorf("セッション作成失敗: %w", err)
	}

	// 統合レイアウトの作成（動的dev数対応）
	fmt.Println("🎛️ 統合レイアウトを作成中...")
	if err := tmuxManager.CreateIntegratedLayout(sessionName, teamConfig.DevCount); err != nil {
		return fmt.Errorf("統合レイアウト作成失敗: %w", err)
	}

	// Claude CLI自動起動処理（設定ファイル対応）
	fmt.Println("🤖 各ペインでClaude CLIを起動中...")
	if err := tmuxManager.SetupClaudeInPanesWithConfig(sessionName, teamConfig.ClaudeCLIPath, teamConfig.InstructionsDir, teamConfig, teamConfig.DevCount); err != nil {
		fmt.Printf("⚠️ Claude CLI自動起動失敗: %v\n", err)
		fmt.Printf("手動でClaude CLIを起動してください: %s --dangerously-skip-permissions\n", teamConfig.ClaudeCLIPath)
		// フォールバック: 従来の方法を試行
		fmt.Println("🔄 フォールバック: 従来の方法で再試行中...")
		if err := tmuxManager.SetupClaudeInPanes(sessionName, teamConfig.ClaudeCLIPath, teamConfig.InstructionsDir, teamConfig.DevCount); err != nil {
			fmt.Printf("⚠️ フォールバック起動も失敗: %v\n", err)
		} else {
			fmt.Println("✅ フォールバック起動成功")
		}
	} else {
		fmt.Println("✅ Claude CLI自動起動完了（設定ファイル対応）")
	}

	fmt.Printf("✅ セッション '%s' の準備が完了しました\n", sessionName)
	return tmuxManager.AttachSession(sessionName)
}

// InitializeSystemCommand システム初期化コマンド
func InitializeSystemCommand(forceOverwrite bool) error {
	fmt.Println("🚀 Claude Code Agentsシステム初期化")
	fmt.Println("=====================================")

	if forceOverwrite {
		fmt.Println("⚠️ 強制上書きモードが有効です")
	}

	// ディレクトリ作成処理
	if err := createSystemDirectories(forceOverwrite); err != nil {
		return fmt.Errorf("ディレクトリ作成に失敗: %w", err)
	}

	// 設定ファイル生成処理
	if err := generateInitialConfig(forceOverwrite); err != nil {
		return fmt.Errorf("設定ファイル生成に失敗: %w", err)
	}

	// 成功メッセージ表示
	displayInitializationSuccess()

	return nil
}

// createSystemDirectories システムディレクトリの作成
func createSystemDirectories(forceOverwrite bool) error {
	fmt.Println("📁 ディレクトリ構造を作成中...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	// 作成対象ディレクトリ一覧
	directories := []struct {
		path        string
		description string
	}{
		{filepath.Join(homeDir, ".claude"), "Claude基本ディレクトリ"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents"), "Claude Code Agentsディレクトリ"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"), "インストラクションディレクトリ"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents", "auth_backup"), "認証バックアップディレクトリ"},
		{filepath.Join(homeDir, ".claude", "claude-code-agents", "logs"), "ログディレクトリ"},
	}

	for _, dir := range directories {
		fmt.Printf("  📂 %s: %s\n", dir.description, dir.path)

		// ディレクトリが既に存在する場合のチェック
		if _, err := os.Stat(dir.path); err == nil {
			if !forceOverwrite {
				fmt.Printf("     ✅ 既に存在します（スキップ）\n")
				continue
			}
			fmt.Printf("     ⚠️ 既に存在しますが続行します（強制モード）\n")
		}

		// ディレクトリを作成
		if err := os.MkdirAll(dir.path, 0750); err != nil {
			return fmt.Errorf("ディレクトリ作成失敗 %s: %w", dir.path, err)
		}
		fmt.Printf("     ✅ 作成完了\n")
	}

	return nil
}

// generateInitialConfig 初期設定ファイルの生成
func generateInitialConfig(forceOverwrite bool) error {
	fmt.Println("⚙️ 設定ファイルを生成中...")

	// ConfigGeneratorを使用して設定ファイルを生成
	configGenerator := config.NewConfigGenerator()

	// 設定ファイル生成用のテンプレート作成
	templateContent := generateConfigTemplate()

	var err error
	if forceOverwrite {
		err = configGenerator.ForceGenerateConfig(templateContent)
	} else {
		err = configGenerator.GenerateConfig(templateContent)
	}

	if err != nil {
		return fmt.Errorf("設定ファイル生成失敗: %w", err)
	}

	fmt.Println("  ✅ agents.conf設定ファイルが作成されました")
	return nil
}

// generateConfigTemplate 設定ファイルテンプレートの生成
func generateConfigTemplate() string {
	return `# Claude Code Agents 設定ファイル
# このファイルはシステム初期化時に自動生成されました

# Path Configurations
CLAUDE_CLI_PATH=~/.claude/local/claude
INSTRUCTIONS_DIR=~/.claude/claude-code-agents/instructions
CONFIG_DIR=~/.claude/claude-code-agents
LOG_FILE=~/.claude/claude-code-agents/logs/manager.log
AUTH_BACKUP_DIR=~/.claude/claude-code-agents/auth_backup

# System Settings
LOG_LEVEL=info

# Tmux Settings
SESSION_NAME=ai-teams
DEFAULT_LAYOUT=integrated
AUTO_ATTACH=false
IDE_BACKUP_ENABLED=true

# Commands
SEND_COMMAND=send-agent
BINARY_NAME=claude-code-agents

# Developer Settings
DEV_COUNT=4

# Role-based Instructions
PO_INSTRUCTION_FILE=po.md
MANAGER_INSTRUCTION_FILE=manager.md
DEV_INSTRUCTION_FILE=developer.md

# Timeout Settings
HEALTH_CHECK_INTERVAL=30s
AUTH_CHECK_INTERVAL=30m
STARTUP_TIMEOUT=10s
SHUTDOWN_TIMEOUT=15s
RESTART_DELAY=5s
PROCESS_TIMEOUT=30s

# === Extended Instruction Configuration ===
# 動的instruction設定を有効にするには、以下の設定を編集してください

# 環境設定
# ENVIRONMENT=development
# STRICT_VALIDATION=false
# FALLBACK_INSTRUCTION_DIR=~/.claude/claude-code-agents/fallback

# 拡張instruction設定（JSON形式で設定可能）
# 詳細な設定については、documentation/instruction-config.mdを参照してください
`
}

// displayInitializationSuccess 初期化成功メッセージの表示
func displayInitializationSuccess() {
	homeDir, _ := os.UserHomeDir()

	fmt.Println()
	fmt.Println("🎉 システム初期化が完了しました！")
	fmt.Println("=" + strings.Repeat("=", 38))
	fmt.Println()
	fmt.Println("📂 作成されたディレクトリ:")
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents"))
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "auth_backup"))
	fmt.Printf("  • %s\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "logs"))
	fmt.Println()
	fmt.Println("📝 作成されたファイル:")
	fmt.Printf("  • %s/agents.conf\n", filepath.Join(homeDir, ".claude", "claude-code-agents"))
	fmt.Println()
	fmt.Println("💡 次のステップ:")
	fmt.Println("  1. インストラクションファイルを配置してください:")
	fmt.Printf("     • %s/po.md\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Printf("     • %s/manager.md\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Printf("     • %s/developer.md\n", filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"))
	fmt.Println("  2. システムの健全性を確認してください:")
	fmt.Println("     ./claude-code-agents --doctor")
	fmt.Println("  3. Claude CLIで認証を行ってください:")
	fmt.Println("     claude auth")
	fmt.Println("  4. システムを起動してください:")
	fmt.Println("     ./claude-code-agents ai-teams")
	fmt.Println()
}

// GenerateConfigCommand 設定ファイル生成コマンド
func GenerateConfigCommand(forceOverwrite bool) error {
	fmt.Println("⚙️ 設定ファイル生成")
	fmt.Println("====================")

	if forceOverwrite {
		fmt.Println("⚠️ 強制上書きモードが有効です")
	}

	// 設定ファイル生成処理
	if err := generateInitialConfig(forceOverwrite); err != nil {
		return fmt.Errorf("設定ファイル生成に失敗: %w", err)
	}

	fmt.Println("✅ 設定ファイル生成が完了しました")

	homeDir, _ := os.UserHomeDir()
	fmt.Printf("📝 生成されたファイル: %s/agents.conf\n", filepath.Join(homeDir, ".claude", "claude-code-agents"))
	fmt.Println()
	fmt.Println("💡 次のステップ:")
	fmt.Println("  1. 設定ファイルを確認・編集してください")
	fmt.Println("  2. システムの健全性を確認してください: ./claude-code-agents --doctor")

	return nil
}

// gatherInstructionInfo instructionファイル情報を収集
func gatherInstructionInfo(config *config.TeamConfig) map[string]interface{} {
	info := make(map[string]interface{})

	// 基本instructionファイル情報
	info["po_instruction_file"] = config.POInstructionFile
	info["manager_instruction_file"] = config.ManagerInstructionFile
	info["dev_instruction_file"] = config.DevInstructionFile
	info["instructions_directory"] = config.InstructionsDir

	// 拡張instruction設定があるかチェック
	if config.InstructionConfig != nil {
		info["enhanced_config_enabled"] = true
		info["base_config"] = map[string]interface{}{
			"po_path":      config.InstructionConfig.Base.POInstructionPath,
			"manager_path": config.InstructionConfig.Base.ManagerInstructionPath,
			"dev_path":     config.InstructionConfig.Base.DevInstructionPath,
		}

		// 環境別設定
		if len(config.InstructionConfig.Environments) > 0 {
			info["environment_configs"] = len(config.InstructionConfig.Environments)
			envNames := make([]string, 0, len(config.InstructionConfig.Environments))
			for envName := range config.InstructionConfig.Environments {
				envNames = append(envNames, envName)
			}
			info["available_environments"] = envNames
		}

		// グローバル設定
		if config.InstructionConfig.Global.DefaultExtension != "" {
			info["default_extension"] = config.InstructionConfig.Global.DefaultExtension
		}
		if config.InstructionConfig.Global.CacheEnabled {
			info["cache_enabled"] = true
			info["cache_ttl"] = config.InstructionConfig.Global.CacheTTL.String()
		}
	} else {
		info["enhanced_config_enabled"] = false
	}

	// フォールバック設定
	if config.FallbackInstructionDir != "" {
		info["fallback_directory"] = config.FallbackInstructionDir
	}

	// 環境設定
	if config.Environment != "" {
		info["current_environment"] = config.Environment
	}

	// バリデーション設定
	info["strict_validation"] = config.StrictValidation

	return info
}

// gatherEnvironmentInfo 環境情報を収集
func gatherEnvironmentInfo(config *config.TeamConfig) map[string]interface{} {
	info := make(map[string]interface{})

	// システム情報
	info["claude_cli_path"] = config.ClaudeCLIPath
	info["working_directory"] = config.WorkingDir
	info["config_directory"] = config.ConfigDir
	info["log_file"] = config.LogFile
	info["session_name"] = config.SessionName
	info["dev_count"] = config.DevCount

	// tmux設定
	info["tmux_layout"] = config.DefaultLayout
	info["auto_attach"] = config.AutoAttach
	info["pane_count"] = config.PaneCount

	// タイムアウト設定
	info["startup_timeout"] = config.StartupTimeout.String()
	info["shutdown_timeout"] = config.ShutdownTimeout.String()
	info["process_timeout"] = config.ProcessTimeout.String()

	// リソース設定
	info["max_processes"] = config.MaxProcesses
	info["max_memory_mb"] = config.MaxMemoryMB
	info["max_cpu_percent"] = config.MaxCPUPercent

	// 監視設定
	info["health_check_interval"] = config.HealthCheckInterval.String()
	info["auth_check_interval"] = config.AuthCheckInterval.String()

	return info
}
