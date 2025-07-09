package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ログレベル設定
var (
	verboseLogging = false
	silentMode     = false
)

// SetVerboseLogging 詳細ログ出力の有効化/無効化
func SetVerboseLogging(verbose bool) {
	verboseLogging = verbose
}

// SetSilentMode サイレントモードの有効化/無効化
func SetSilentMode(silent bool) {
	silentMode = silent
}

// IsVerboseLogging 詳細ログ出力が有効か確認
func IsVerboseLogging() bool {
	return verboseLogging
}

// IsSilentMode サイレントモードが有効か確認
func IsSilentMode() bool {
	return silentMode
}

// displaySimpleMessage 簡素化されたメッセージ表示（コアメッセージのみ）
func displaySimpleMessage(message string) {
	if silentMode {
		return
	}
	fmt.Println(message)
}

// displayClaudePath Claude CLIパスの表示
func displayClaudePath(path string) {
	if silentMode {
		return
	}
	fmt.Printf("Claude Path: %s\n", path)
}

// displayConfigFileLoaded 設定ファイル読み込み完了の表示
func displayConfigFileLoaded(configPath string, content string) {
	if silentMode {
		return
	}
	fmt.Printf("設定ファイルを読み込みました: %s\n", configPath)
	if verboseLogging && content != "" {
		fmt.Printf("設定内容:\n%s\n", content)
	}
	
	// 設定読み込み後の自動表示
	if verboseLogging {
		displayConfigAfterLoad(configPath)
	}
}

// displayConfigAfterLoad 設定ファイル読み込み後の自動表示
func displayConfigAfterLoad(configPath string) {
	if silentMode {
		return
	}
	
	fmt.Println()
	fmt.Println("📋 設定ファイル読み込み後の設定値一覧:")
	fmt.Println(strings.Repeat("=", 40))
	
	// TeamConfig設定の読み込みと表示
	if strings.Contains(configPath, ".agents.conf") {
		if teamConfig, err := LoadTeamConfig(); err == nil {
			displayTeamConfigSummary(teamConfig)
		} else {
			fmt.Printf("⚠️ TeamConfig読み込みエラー: %v\n", err)
		}
	}
	
	// MainConfig設定の読み込みと表示
	if strings.Contains(configPath, "manager.json") {
		if mainConfig, err := LoadConfig(configPath); err == nil {
			displayMainConfigSummary(mainConfig)
		} else {
			fmt.Printf("⚠️ MainConfig読み込みエラー: %v\n", err)
		}
	}
	
	fmt.Println()
}

// displayTeamConfigSummary TeamConfig設定の要約表示
func displayTeamConfigSummary(config *TeamConfig) {
	fmt.Printf("🎭 Team Configuration Summary:\n")
	fmt.Printf("   Session: %s\n", config.SessionName)
	fmt.Printf("   Max Processes: %d\n", config.MaxProcesses)
	fmt.Printf("   Max Memory: %d MB\n", config.MaxMemoryMB)
	fmt.Printf("   Log Level: %s\n", config.LogLevel)
	fmt.Printf("   Claude CLI: %s\n", formatPath(config.ClaudeCLIPath))
	fmt.Printf("   Instructions: %s\n", formatPath(config.InstructionsDir))
}

// displayMainConfigSummary MainConfig設定の要約表示
func displayMainConfigSummary(config *Config) {
	fmt.Printf("⚙️ Main Configuration Summary:\n")
	fmt.Printf("   Max Processes: %d\n", config.MaxProcesses)
	fmt.Printf("   Max Memory: %d MB\n", config.MaxMemoryMB)
	fmt.Printf("   Max CPU: %.1f%%\n", config.MaxCPUPercent)
	fmt.Printf("   Log Level: %s\n", config.LogLevel)
	fmt.Printf("   Claude Path: %s\n", formatPath(config.ClaudePath))
	fmt.Printf("   Instructions: %s\n", formatPath(config.InstructionsDir))
	fmt.Printf("   Process Timeout: %v\n", config.ProcessTimeout)
	fmt.Printf("   Restart Delay: %v\n", config.RestartDelay)
}

// displaySessionName セッション名の表示
func displaySessionName(sessionName string) {
	if silentMode {
		return
	}
	fmt.Printf("セッション名称: %s\n", sessionName)
}

// displayLauncherStart 統合監視起動の表示
func displayLauncherStart() {
	if silentMode {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] ℹ️ 統合監視起動 統合監視画面方式でシステムを起動します\n", timestamp)
}

// displayLauncherProgress 統合監視起動中の表示
func displayLauncherProgress() {
	if silentMode {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] 🔄 統合監視起動 統合監視画面方式でシステムを起動中...\n", timestamp)
}

// displayAgentDeployment エージェント配置の表示
func displayAgentDeployment() {
	if silentMode {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] 🔄 エージェント配置\n", timestamp)
}

// displayConfig 設定情報の表示（統合版）
func displayConfig(config *TeamConfig, sessionName string) {
	// メイン設定も読み込む
	mainConfig, err := LoadConfig(filepath.Join(config.ConfigDir, "manager.json"))
	if err != nil {
		mainConfig = DefaultConfig()
	}

	// 共通設定も読み込む
	_ = GetCommonConfig() // 未使用警告を回避

	fmt.Println("🚀 AI Teams System Configuration")
	fmt.Println("=" + strings.Repeat("=", 35))
	fmt.Println()

	// セッション情報
	fmt.Printf("📋 Session Information:\n")
	fmt.Printf("   Session Name: %s\n", sessionName)
	fmt.Printf("   Layout: %s\n", config.DefaultLayout)
	fmt.Printf("   Pane Count: %d\n", config.PaneCount)
	fmt.Printf("   Auto Attach: %t\n", config.AutoAttach)
	fmt.Println()

	// 設定ファイル情報
	configPath := GetTeamConfigPath()
	mainConfigPath := filepath.Join(config.ConfigDir, "manager.json")
	fmt.Printf("📁 Configuration Files:\n")
	fmt.Printf("   Team Config Path: %s\n", configPath)
	fmt.Printf("   Main Config Path: %s\n", mainConfigPath)
	
	// 設定ファイルの存在確認
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("   Team Config Status: ✅ Found\n")
	} else {
		fmt.Printf("   Team Config Status: ⚠️ Using defaults\n")
	}
	if _, err := os.Stat(mainConfigPath); err == nil {
		fmt.Printf("   Main Config Status: ✅ Found\n")
	} else {
		fmt.Printf("   Main Config Status: ⚠️ Using defaults\n")
	}
	fmt.Println()

	// パス設定
	fmt.Printf("🗂️ Path Configuration:\n")
	fmt.Printf("   Claude CLI Path: %s\n", config.ClaudeCLIPath)
	fmt.Printf("   Instructions Dir: %s\n", config.InstructionsDir)
	fmt.Printf("   Working Dir: %s\n", config.WorkingDir)
	fmt.Printf("   Config Dir: %s\n", config.ConfigDir)
	fmt.Printf("   Log File: %s\n", config.LogFile)
	fmt.Printf("   Auth Backup Dir: %s\n", config.AuthBackupDir)
	fmt.Println()

	// システム設定（統合）
	fmt.Printf("⚙️ System Settings:\n")
	fmt.Printf("   Max Processes: %d (Main: %d, Team: %d)\n", mainConfig.MaxProcesses, mainConfig.MaxProcesses, config.MaxProcesses)
	fmt.Printf("   Max Memory: %d MB (Main: %d MB, Team: %d MB)\n", mainConfig.MaxMemoryMB, mainConfig.MaxMemoryMB, config.MaxMemoryMB)
	fmt.Printf("   Max CPU: %.1f%% (Main: %.1f%%, Team: %.1f%%)\n", mainConfig.MaxCPUPercent, mainConfig.MaxCPUPercent, config.MaxCPUPercent)
	fmt.Printf("   Log Level: %s (Main: %s, Team: %s)\n", mainConfig.LogLevel, mainConfig.LogLevel, config.LogLevel)
	fmt.Printf("   Max Restart Attempts: %d (Main: %d, Team: %d)\n", mainConfig.MaxRestartAttempts, mainConfig.MaxRestartAttempts, config.MaxRestartAttempts)
	fmt.Println()

	// タイムアウト設定（統合）
	fmt.Printf("⏱️ Timeout Settings:\n")
	fmt.Printf("   Startup Timeout: %v (Main: %v, Team: %v)\n", mainConfig.StartupTimeout, mainConfig.StartupTimeout, config.StartupTimeout)
	fmt.Printf("   Shutdown Timeout: %v (Main: %v, Team: %v)\n", mainConfig.ShutdownTimeout, mainConfig.ShutdownTimeout, config.ShutdownTimeout)
	fmt.Printf("   Process Timeout: %v (Main: %v, Team: %v)\n", mainConfig.ProcessTimeout, mainConfig.ProcessTimeout, config.ProcessTimeout)
	fmt.Printf("   Restart Delay: %v (Main: %v, Team: %v)\n", mainConfig.RestartDelay, mainConfig.RestartDelay, config.RestartDelay)
	fmt.Printf("   Health Check Interval: %v (Main: %v, Team: %v)\n", mainConfig.HealthCheckInterval, mainConfig.HealthCheckInterval, config.HealthCheckInterval)
	fmt.Printf("   Auth Check Interval: %v (Main: %v, Team: %v)\n", mainConfig.AuthCheckInterval, mainConfig.AuthCheckInterval, config.AuthCheckInterval)
	fmt.Println()

	// 認証・バックアップ設定
	fmt.Printf("🔐 Authentication & Backup:\n")
	fmt.Printf("   IDE Backup Enabled: %t\n", config.IDEBackupEnabled)
	fmt.Printf("   Send Command: %s\n", config.SendCommand)
	fmt.Printf("   Binary Name: %s\n", config.BinaryName)
	fmt.Println()

	// 環境情報
	fmt.Printf("🌍 Environment Information:\n")
	fmt.Printf("   Current User: %s\n", getCurrentUser())
	fmt.Printf("   Working Directory: %s\n", getActualWorkingDir())
	fmt.Printf("   System Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 設定値の統合表示
	displayConfigComparison(mainConfig, config)

	// パス検証結果の表示
	displayValidationResults(config)

	log.Info().
		Str("session", sessionName).
		Str("layout", config.DefaultLayout).
		Int("pane_count", config.PaneCount).
		Str("config_path", configPath).
		Msg("Configuration displayed")
}

// displayProgress 進捗状況の表示
func displayProgress(step string, details string) {
	if silentMode {
		return
	}
	
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] 🔄 %s\n", timestamp, step)
	if details != "" && verboseLogging {
		fmt.Printf("         %s\n", details)
	}
	
	// 詳細ログ出力時のみログファイルに記録
	if verboseLogging {
		log.Info().
			Str("step", step).
			Str("details", details).
			Msg("Progress update")
	}
}

// displayError エラー発生時の詳細情報表示
func displayError(step string, err error) {
	if silentMode {
		return
	}
	
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] ❌ %s\n", timestamp, step)
	fmt.Printf("         Error: %v\n", err)
	
	// エラーの詳細情報は詳細モードでのみ表示
	if verboseLogging && err != nil {
		fmt.Printf("         Type: %T\n", err)
		
		// エラーメッセージの詳細分析
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "permission denied") {
			fmt.Printf("         💡 Hint: アクセス権限を確認してください\n")
		} else if strings.Contains(errorMsg, "no such file") {
			fmt.Printf("         💡 Hint: ファイルパスが正しいか確認してください\n")
		} else if strings.Contains(errorMsg, "connection refused") {
			fmt.Printf("         💡 Hint: システムが起動していることを確認してください\n")
		} else if strings.Contains(errorMsg, "timeout") {
			fmt.Printf("         💡 Hint: タイムアウト時間を延長することを検討してください\n")
		}
	}
	
	// エラーは常にログファイルに記録
	log.Error().
		Err(err).
		Str("step", step).
		Msg("Error occurred")
}

// displaySuccess 成功時の詳細情報表示
func displaySuccess(step string, details string) {
	if silentMode {
		return
	}
	
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] ✅ %s\n", timestamp, step)
	if details != "" && verboseLogging {
		fmt.Printf("         %s\n", details)
	}
	
	// 詳細ログ出力時のみログファイルに記録
	if verboseLogging {
		log.Info().
			Str("step", step).
			Str("details", details).
			Msg("Success")
	}
}

// displayWarning 警告情報の表示
func displayWarning(step string, details string) {
	if silentMode {
		return
	}
	
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] ⚠️ %s\n", timestamp, step)
	if details != "" && verboseLogging {
		fmt.Printf("         %s\n", details)
	}
	
	// 警告は常にログファイルに記録
	log.Warn().
		Str("step", step).
		Str("details", details).
		Msg("Warning")
}

// displayInfo 情報の表示
func displayInfo(step string, details string) {
	if silentMode {
		return
	}
	
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] ℹ️ %s\n", timestamp, step)
	if details != "" && verboseLogging {
		fmt.Printf("         %s\n", details)
	}
	
	// 詳細ログ出力時のみログファイルに記録
	if verboseLogging {
		log.Info().
			Str("step", step).
			Str("details", details).
			Msg("Info")
	}
}

// displayHeader セクションヘッダーの表示
func displayHeader(title string) {
	fmt.Println()
	fmt.Printf("🎯 %s\n", title)
	fmt.Println(strings.Repeat("=", len(title)+4))
}

// displaySubHeader サブセクションヘッダーの表示
func displaySubHeader(title string) {
	fmt.Printf("\n📌 %s\n", title)
	fmt.Println(strings.Repeat("-", len(title)+4))
}

// displaySeparator セパレーターの表示
func displaySeparator() {
	fmt.Println(strings.Repeat("─", 50))
}

// displayAgentStatus エージェントステータスの表示
func displayAgentStatus(agentName string, status string, isRunning bool) {
	timestamp := time.Now().Format("15:04:05")
	statusIcon := "❌"
	if isRunning {
		statusIcon = "✅"
	}
	
	fmt.Printf("[%s] %s %s: %s\n", timestamp, statusIcon, agentName, status)
	
	log.Info().
		Str("agent", agentName).
		Str("status", status).
		Bool("running", isRunning).
		Msg("Agent status")
}

// displaySystemStatus システム全体のステータス表示
func displaySystemStatus(running bool, agentCount int, sessionName string) {
	fmt.Println()
	fmt.Println("🖥️ System Status Overview")
	fmt.Println("=" + strings.Repeat("=", 25))
	
	systemIcon := "❌"
	systemStatus := "停止中"
	if running {
		systemIcon = "✅"
		systemStatus = "実行中"
	}
	
	fmt.Printf("   %s System: %s\n", systemIcon, systemStatus)
	fmt.Printf("   📊 Active Agents: %d\n", agentCount)
	fmt.Printf("   🎭 Session: %s\n", sessionName)
	fmt.Printf("   🕐 Last Check: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	
	log.Info().
		Bool("running", running).
		Int("agent_count", agentCount).
		Str("session", sessionName).
		Msg("System status displayed")
}

// displayCommandResult コマンド実行結果の表示
func displayCommandResult(command string, output string, err error) {
	timestamp := time.Now().Format("15:04:05")
	
	if err != nil {
		fmt.Printf("[%s] ❌ Command failed: %s\n", timestamp, command)
		fmt.Printf("         Error: %v\n", err)
		if output != "" {
			fmt.Printf("         Output: %s\n", output)
		}
		
		log.Error().
			Err(err).
			Str("command", command).
			Str("output", output).
			Msg("Command execution failed")
	} else {
		fmt.Printf("[%s] ✅ Command executed: %s\n", timestamp, command)
		if output != "" {
			fmt.Printf("         Output: %s\n", output)
		}
		
		log.Info().
			Str("command", command).
			Str("output", output).
			Msg("Command executed successfully")
	}
}

// displayStartupBanner スタートアップバナーの表示
func displayStartupBanner() {
	fmt.Println()
	fmt.Println("🚀 ═══════════════════════════════════════════════════════════════")
	fmt.Println("   AI Teams System - Claude Code Agents")
	fmt.Println("   Version: 1.0.0")
	fmt.Println("   Developed by: Shivase Team")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
}

// displayShutdownBanner シャットダウンバナーの表示
func displayShutdownBanner() {
	fmt.Println()
	fmt.Println("🛑 ═══════════════════════════════════════════════════════════════")
	fmt.Println("   AI Teams System - Shutdown Complete")
	fmt.Println("   Thank you for using Claude Code Agents!")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
}

// ヘルパー関数

// getCurrentUser 現在のユーザー名を取得
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// getActualWorkingDir 実際の作業ディレクトリを取得
func getActualWorkingDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "unknown"
}

// formatPath パスの表示用フォーマット
func formatPath(path string) string {
	if path == "" {
		return "未設定"
	}
	
	// ホームディレクトリの短縮表示
	if homeDir, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(path, homeDir) {
			return strings.Replace(path, homeDir, "~", 1)
		}
	}
	
	return path
}

// formatDuration 時間の表示用フォーマット
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "未設定"
	}
	return d.String()
}

// formatMemory メモリサイズの表示用フォーマット
func formatMemory(mb int64) string {
	if mb == 0 {
		return "未設定"
	}
	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB", float64(mb)/1024.0)
	}
	return fmt.Sprintf("%d MB", mb)
}

// checkPathExists パスの存在確認と表示
func checkPathExists(path string) (bool, string) {
	if path == "" {
		return false, "未設定"
	}
	
	// チルダ展開
	if strings.HasPrefix(path, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	
	if _, err := os.Stat(path); err == nil {
		return true, "✅ 存在"
	}
	return false, "❌ 不在"
}

// displayPathValidation パスの検証結果表示
func displayPathValidation(label string, path string) {
	exists, status := checkPathExists(path)
	icon := "❌"
	if exists {
		icon = "✅"
	}
	
	fmt.Printf("   %s %s: %s (%s)\n", icon, label, formatPath(path), status)
}

// displayValidationResults 検証結果の表示
func displayValidationResults(config *TeamConfig) {
	fmt.Println()
	displayHeader("Path Validation Results")
	
	displayPathValidation("Claude CLI", config.ClaudeCLIPath)
	displayPathValidation("Instructions Directory", config.InstructionsDir)
	displayPathValidation("Working Directory", config.WorkingDir)
	displayPathValidation("Config Directory", config.ConfigDir)
	displayPathValidation("Auth Backup Directory", config.AuthBackupDir)
	
	// ログファイルのディレクトリ確認
	if config.LogFile != "" {
		logDir := filepath.Dir(config.LogFile)
		displayPathValidation("Log Directory", logDir)
	}
	
	fmt.Println()
}

// displayConfigComparison 設定値の比較表示
func displayConfigComparison(mainConfig *Config, teamConfig *TeamConfig) {
	fmt.Println()
	displayHeader("Configuration Comparison")
	
	fmt.Printf("📊 Effective Settings (Main Config vs Team Config):\n")
	fmt.Printf("   Max Processes: %d vs %d\n", mainConfig.MaxProcesses, teamConfig.MaxProcesses)
	fmt.Printf("   Max Memory MB: %d vs %d\n", mainConfig.MaxMemoryMB, teamConfig.MaxMemoryMB)
	fmt.Printf("   Max CPU Percent: %.1f vs %.1f\n", mainConfig.MaxCPUPercent, teamConfig.MaxCPUPercent)
	fmt.Printf("   Log Level: %s vs %s\n", mainConfig.LogLevel, teamConfig.LogLevel)
	fmt.Printf("   Max Restart Attempts: %d vs %d\n", mainConfig.MaxRestartAttempts, teamConfig.MaxRestartAttempts)
	fmt.Printf("   Startup Timeout: %v vs %v\n", mainConfig.StartupTimeout, teamConfig.StartupTimeout)
	fmt.Printf("   Shutdown Timeout: %v vs %v\n", mainConfig.ShutdownTimeout, teamConfig.ShutdownTimeout)
	fmt.Printf("   Process Timeout: %v vs %v\n", mainConfig.ProcessTimeout, teamConfig.ProcessTimeout)
	fmt.Printf("   Restart Delay: %v vs %v\n", mainConfig.RestartDelay, teamConfig.RestartDelay)
	fmt.Printf("   Health Check Interval: %v vs %v\n", mainConfig.HealthCheckInterval, teamConfig.HealthCheckInterval)
	fmt.Printf("   Auth Check Interval: %v vs %v\n", mainConfig.AuthCheckInterval, teamConfig.AuthCheckInterval)
	fmt.Println()

	// 相違点の強調表示
	fmt.Printf("🔍 Configuration Differences:\n")
	highlight := "⚠️"
	same := "✅"
	
	if mainConfig.MaxProcesses != teamConfig.MaxProcesses {
		fmt.Printf("   %s Max Processes: Different\n", highlight)
	} else {
		fmt.Printf("   %s Max Processes: Same\n", same)
	}
	
	if mainConfig.MaxMemoryMB != teamConfig.MaxMemoryMB {
		fmt.Printf("   %s Max Memory: Different\n", highlight)
	} else {
		fmt.Printf("   %s Max Memory: Same\n", same)
	}
	
	if mainConfig.MaxCPUPercent != teamConfig.MaxCPUPercent {
		fmt.Printf("   %s Max CPU: Different\n", highlight)
	} else {
		fmt.Printf("   %s Max CPU: Same\n", same)
	}
	
	if mainConfig.LogLevel != teamConfig.LogLevel {
		fmt.Printf("   %s Log Level: Different\n", highlight)
	} else {
		fmt.Printf("   %s Log Level: Same\n", same)
	}
	
	fmt.Println()
}

// displayAllConfigurations 全設定値の詳細表示
func displayAllConfigurations(sessionName string) {
	// TeamConfig設定の読み込み
	teamConfig, err := LoadTeamConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load team config")
		return
	}

	// 通常の設定表示
	displayConfig(teamConfig, sessionName)

	// 詳細な設定値の表示
	displayDetailedConfig(teamConfig)

	// システム状態の表示
	displaySystemState(teamConfig)
}

// displayDetailedConfig 詳細設定の表示
func displayDetailedConfig(config *TeamConfig) {
	fmt.Println()
	displayHeader("Detailed Configuration")
	
	fmt.Printf("📋 Team Configuration Details:\n")
	fmt.Printf("   Session Name: %s\n", config.SessionName)
	fmt.Printf("   Default Layout: %s\n", config.DefaultLayout)
	fmt.Printf("   Pane Count: %d\n", config.PaneCount)
	fmt.Printf("   Auto Attach: %t\n", config.AutoAttach)
	fmt.Printf("   IDE Backup Enabled: %t\n", config.IDEBackupEnabled)
	fmt.Printf("   Send Command: %s\n", config.SendCommand)
	fmt.Printf("   Binary Name: %s\n", config.BinaryName)
	fmt.Println()

	fmt.Printf("📂 Path Configuration Details:\n")
	fmt.Printf("   Claude CLI Path: %s\n", formatPath(config.ClaudeCLIPath))
	fmt.Printf("   Instructions Dir: %s\n", formatPath(config.InstructionsDir))
	fmt.Printf("   Working Dir: %s\n", formatPath(config.WorkingDir))
	fmt.Printf("   Config Dir: %s\n", formatPath(config.ConfigDir))
	fmt.Printf("   Log File: %s\n", formatPath(config.LogFile))
	fmt.Printf("   Auth Backup Dir: %s\n", formatPath(config.AuthBackupDir))
	fmt.Println()

	fmt.Printf("💾 Resource Configuration Details:\n")
	fmt.Printf("   Max Processes: %d\n", config.MaxProcesses)
	fmt.Printf("   Max Memory: %s\n", formatMemory(config.MaxMemoryMB))
	fmt.Printf("   Max CPU: %.1f%%\n", config.MaxCPUPercent)
	fmt.Printf("   Log Level: %s\n", config.LogLevel)
	fmt.Printf("   Max Restart Attempts: %d\n", config.MaxRestartAttempts)
	fmt.Println()

	fmt.Printf("⏰ Timeout Configuration Details:\n")
	fmt.Printf("   Startup Timeout: %s\n", formatDuration(config.StartupTimeout))
	fmt.Printf("   Shutdown Timeout: %s\n", formatDuration(config.ShutdownTimeout))
	fmt.Printf("   Process Timeout: %s\n", formatDuration(config.ProcessTimeout))
	fmt.Printf("   Restart Delay: %s\n", formatDuration(config.RestartDelay))
	fmt.Printf("   Health Check Interval: %s\n", formatDuration(config.HealthCheckInterval))
	fmt.Printf("   Auth Check Interval: %s\n", formatDuration(config.AuthCheckInterval))
	fmt.Println()
}

// displaySystemState システム状態の表示
func displaySystemState(config *TeamConfig) {
	fmt.Println()
	displayHeader("System State")
	
	fmt.Printf("🖥️ Runtime Information:\n")
	fmt.Printf("   Current User: %s\n", getCurrentUser())
	fmt.Printf("   Working Directory: %s\n", getActualWorkingDir())
	fmt.Printf("   System Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("   Process ID: %d\n", os.Getpid())
	fmt.Printf("   Parent Process ID: %d\n", os.Getppid())
	fmt.Println()

	fmt.Printf("🔧 Environment Variables:\n")
	if home := os.Getenv("HOME"); home != "" {
		fmt.Printf("   HOME: %s\n", home)
	}
	if user := os.Getenv("USER"); user != "" {
		fmt.Printf("   USER: %s\n", user)
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		fmt.Printf("   SHELL: %s\n", shell)
	}
	if term := os.Getenv("TERM"); term != "" {
		fmt.Printf("   TERM: %s\n", term)
	}
	fmt.Println()

	fmt.Printf("📊 Resource Usage:\n")
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("   Memory Usage: %s\n", formatMemory(int64(memStats.Sys/1024/1024)))
	fmt.Printf("   Allocated: %s\n", formatMemory(int64(memStats.Alloc/1024/1024)))
	fmt.Printf("   Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Printf("   CPU Cores: %d\n", runtime.NumCPU())
	fmt.Println()
}