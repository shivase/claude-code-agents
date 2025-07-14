package cmd

import (
	"fmt"
	"time"

	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/shivase/claude-code-agents/internal/utils"
)

// DisplayConfigCommand 設定情報詳細表示コマンド
func DisplayConfigCommand() error {
	fmt.Println("🔧 AI Teams System - 設定情報詳細表示")
	fmt.Println("=========================================")

	// 1. 統一設定の読み込み
	unifiedConfig, err := config.LoadUnifiedConfig()
	if err != nil {
		fmt.Printf("⚠️ 統一設定の読み込みに失敗: %v\n", err)
		fmt.Println("📝 基本設定情報のみ表示します")

		// フォールバック：基本設定のみ表示
		displayBasicConfigFallback()
		return nil
	}

	// 2. TeamConfig 全設定値の詳細表示
	fmt.Println("\n📁 TeamConfig - チーム設定")
	fmt.Println("---------------------------")
	displayTeamConfig(unifiedConfig.Team)

	// 3. CommonConfig 全設定値の詳細表示
	fmt.Println("\n⚙️ CommonConfig - 共通設定")
	fmt.Println("----------------------------")
	fmt.Println("   ⚠️ CommonConfig は削除されました（import cycle解決のため）")

	// 4. パス情報の完全表示
	fmt.Println("\n📂 Path Configuration - パス設定")
	fmt.Println("----------------------------------")
	displayPathConfiguration(unifiedConfig.Paths)

	// 5. システム設定の詳細表示
	fmt.Println("\n🖥️ System Settings - システム設定")
	fmt.Println("-----------------------------------")
	displaySystemSettings(unifiedConfig.Team)

	// 6. 認証設定の詳細表示
	fmt.Println("\n🔐 Authentication Settings - 認証設定")
	fmt.Println("--------------------------------------")
	displayAuthenticationSettings(unifiedConfig.Team)

	// 7. 設定ファイルの存在確認と検証結果表示
	fmt.Println("\n📋 Configuration File Validation - 設定ファイル検証")
	fmt.Println("----------------------------------------------------")
	displayConfigurationValidation(unifiedConfig.Paths)

	// 8. 有効設定値の表示
	fmt.Println("\n✅ Effective Configuration - 有効設定値")
	fmt.Println("----------------------------------------")
	fmt.Println("   有効設定値の表示は実装中です")

	// 9. ディレクトリ解決情報
	fmt.Println("\n📁 Directory Resolution - ディレクトリ解決")
	fmt.Println("------------------------------------------")
	resolver := utils.GetGlobalDirectoryResolver()
	resolver.DisplayDirectoryInfo()

	fmt.Println("=========================================")
	fmt.Printf("🕐 設定表示完了時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return nil
}

// displayTeamConfig TeamConfig詳細表示
func displayTeamConfig(teamConfig *config.TeamConfig) {
	if teamConfig == nil {
		fmt.Println("⚠️ TeamConfig が読み込まれていません")
		return
	}

	fmt.Printf("   Claude CLI Path:      %s\n", teamConfig.ClaudeCLIPath)
	fmt.Printf("   Instructions Dir:     %s\n", teamConfig.InstructionsDir)
	fmt.Printf("   Working Dir:          %s\n", teamConfig.WorkingDir)
	fmt.Printf("   Config Dir:           %s\n", teamConfig.ConfigDir)
	fmt.Printf("   Log File:             %s\n", teamConfig.LogFile)
	fmt.Printf("   Auth Backup Dir:      %s\n", teamConfig.AuthBackupDir)
	fmt.Printf("   Max Processes:        %d\n", teamConfig.MaxProcesses)
	fmt.Printf("   Max Memory (MB):      %d\n", teamConfig.MaxMemoryMB)
	fmt.Printf("   Max CPU Percent:      %.1f%%\n", teamConfig.MaxCPUPercent)
	fmt.Printf("   Log Level:            %s\n", teamConfig.LogLevel)
	fmt.Printf("   Session Name:         %s\n", teamConfig.SessionName)
	fmt.Printf("   Default Layout:       %s\n", teamConfig.DefaultLayout)
	fmt.Printf("   Auto Attach:          %t\n", teamConfig.AutoAttach)
	fmt.Printf("   Pane Count:           %d\n", teamConfig.PaneCount)
	fmt.Printf("   IDE Backup Enabled:   %t\n", teamConfig.IDEBackupEnabled)
	fmt.Printf("   Send Command:         %s\n", teamConfig.SendCommand)
	fmt.Printf("   Binary Name:          %s\n", teamConfig.BinaryName)
	fmt.Printf("   Health Check Interval: %s\n", teamConfig.HealthCheckInterval)
	fmt.Printf("   Auth Check Interval:  %s\n", teamConfig.AuthCheckInterval)
	fmt.Printf("   Startup Timeout:      %s\n", teamConfig.StartupTimeout)
	fmt.Printf("   Shutdown Timeout:     %s\n", teamConfig.ShutdownTimeout)
	fmt.Printf("   Restart Delay:        %s\n", teamConfig.RestartDelay)
	fmt.Printf("   Process Timeout:      %s\n", teamConfig.ProcessTimeout)
	fmt.Printf("   Max Restart Attempts: %d\n", teamConfig.MaxRestartAttempts)
}

// displayPathConfiguration パス設定詳細表示
func displayPathConfiguration(paths *config.ConfigPaths) {
	if paths == nil {
		fmt.Println("⚠️ Path Configuration が読み込まれていません")
		return
	}

	fmt.Printf("   Claude Dir:           %s\n", paths.ClaudeDir)
	fmt.Printf("   Cloud Code Agents Dir: %s\n", paths.CloudCodeAgentsDir)
	fmt.Printf("   Team Config Path:     %s\n", paths.TeamConfigPath)
	fmt.Printf("   Main Config Path:     %s\n", paths.MainConfigPath)
	fmt.Printf("   Logs Dir:             %s\n", paths.LogsDir)
	fmt.Printf("   Instructions Dir:     %s\n", paths.InstructionsDir)
	fmt.Printf("   Auth Backup Dir:      %s\n", paths.AuthBackupDir)
	fmt.Printf("   Claude CLI Path:      %s\n", paths.ClaudeCLIPath)
}

// displaySystemSettings システム設定詳細表示
func displaySystemSettings(teamConfig *config.TeamConfig) {
	if teamConfig == nil {
		fmt.Println("⚠️ System Settings が読み込まれていません")
		return
	}

	fmt.Printf("   最大プロセス数:       %d\n", teamConfig.MaxProcesses)
	fmt.Printf("   最大メモリ使用量:     %d MB\n", teamConfig.MaxMemoryMB)
	fmt.Printf("   最大CPU使用率:        %.1f%%\n", teamConfig.MaxCPUPercent)
	fmt.Printf("   ヘルスチェック間隔:   %s\n", teamConfig.HealthCheckInterval)
	fmt.Printf("   最大再起動試行回数:   %d\n", teamConfig.MaxRestartAttempts)
	fmt.Printf("   プロセスタイムアウト: %s\n", teamConfig.ProcessTimeout)
	fmt.Printf("   起動タイムアウト:     %s\n", teamConfig.StartupTimeout)
	fmt.Printf("   終了タイムアウト:     %s\n", teamConfig.ShutdownTimeout)
	fmt.Printf("   再起動遅延:           %s\n", teamConfig.RestartDelay)
}

// displayAuthenticationSettings 認証設定詳細表示
func displayAuthenticationSettings(teamConfig *config.TeamConfig) {
	if teamConfig == nil {
		fmt.Println("⚠️ Authentication Settings が読み込まれていません")
		return
	}

	fmt.Printf("   認証チェック間隔:     %s\n", teamConfig.AuthCheckInterval)
	fmt.Printf("   認証バックアップ:     %s\n", teamConfig.AuthBackupDir)
	fmt.Printf("   Claude CLI Path:      %s\n", teamConfig.ClaudeCLIPath)
}

// displayConfigurationValidation 設定ファイル検証
func displayConfigurationValidation(paths *config.ConfigPaths) {
	if paths == nil {
		fmt.Println("⚠️ Path Configuration が読み込まれていません")
		return
	}

	fmt.Printf("   Team Config:          %s", paths.TeamConfigPath)
	if utils.ValidatePath(paths.TeamConfigPath) {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ❌")
	}

	fmt.Printf("   Instructions Dir:     %s", paths.InstructionsDir)
	if utils.ValidatePath(paths.InstructionsDir) {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ❌")
	}

	fmt.Printf("   Claude CLI:           %s", paths.ClaudeCLIPath)
	if utils.IsExecutable(utils.ExpandPathSafe(paths.ClaudeCLIPath)) {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ❌")
	}
}

// displayBasicConfigFallback 基本設定フォールバック表示
func displayBasicConfigFallback() {
	fmt.Println("\n📁 基本設定情報")
	fmt.Println("--------------")

	// 基本的な設定情報のみ表示
	configPath := config.GetDefaultTeamConfigPath()
	fmt.Printf("   設定ファイルパス:     %s\n", configPath)

	if utils.ValidatePath(configPath) {
		fmt.Println("   設定ファイル状態:     ✅ 存在")
	} else {
		fmt.Println("   設定ファイル状態:     ❌ 不在")
	}
}

// DisplaySessionConfigCommand セッション設定詳細表示コマンド
func DisplaySessionConfigCommand(sessionName string) error {
	fmt.Printf("🔧 セッション設定詳細表示: %s\n", sessionName)
	fmt.Println("=====================================")

	// セッション固有の設定情報を表示
	fmt.Printf("   セッション名:         %s\n", sessionName)
	fmt.Printf("   表示時刻:             %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return nil
}
