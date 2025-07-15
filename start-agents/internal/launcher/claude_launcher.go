package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/process"
	"github.com/shivase/claude-code-agents/internal/tmux"
	"github.com/shivase/claude-code-agents/internal/utils"
)

// ClaudeLauncher Claude CLI起動用のヘルパー
type ClaudeLauncher struct {
	config      *LauncherConfig
	tmuxManager *tmux.TmuxManagerImpl
}

// NewClaudeLauncher Claude起動ヘルパーを作成
func NewClaudeLauncher(config *LauncherConfig) *ClaudeLauncher {
	return &ClaudeLauncher{
		config:      config,
		tmuxManager: tmux.NewTmuxManager(config.SessionName),
	}
}

// LaunchClaude 指定されたペインまたはセッションでClaude CLIを起動
func (cl *ClaudeLauncher) LaunchClaude(target string) error {
	// プロセスマネージャーを取得
	pm := process.GetGlobalProcessManager()

	// 既存のClaude CLIプロセスをクリーンアップ
	if err := pm.TerminateClaudeProcesses(); err != nil {
		log.Warn().Err(err).Msg("Failed to cleanup existing Claude processes")
	}

	// OAuth認証チェックは環境検証時に完了済みのためスキップ
	log.Info().Str("target", target).Msg("📋 認証チェックをスキップ（環境検証時に完了済み）")

	// Claude CLIコマンドを構築（環境変数で設定を制御）
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Claude CLIを起動（既存認証を使用）
	configDir := filepath.Join(homeDir, ".claude")
	claudeCmd := fmt.Sprintf("CLAUDE_CONFIG_DIR=\"%s\" \"%s\" --dangerously-skip-permissions",
		configDir, cl.config.ClaudePath)

	// tmux環境で既存認証を強制使用するための環境変数を設定
	envSetCmd := fmt.Sprintf("export CLAUDE_CONFIG_DIR=\"%s\"", configDir)
	cmd := exec.Command("tmux", "send-keys", "-t", target, envSetCmd, "C-m") // #nosec G204
	if err := cmd.Run(); err != nil {
		log.Warn().Err(err).Msg("⚠️ 環境変数設定警告")
	}
	time.Sleep(500 * time.Millisecond) // 環境変数設定の反映待機

	// Claude CLI設定ファイルの状態確認
	if _, err := os.Stat(settingsPath); err != nil {
		log.Warn().Str("settings_path", settingsPath).Msg("⚠️ Claude設定ファイル確認が見つかりません")
	} else {
		log.Info().Str("settings_path", settingsPath).Msg("✅ Claude設定ファイル確認を使用")
	}

	// claude.jsonファイルの作成を防ぐための環境変数設定
	claudeJsonPath := filepath.Join(homeDir, ".claude.json")
	if _, err := os.Stat(claudeJsonPath); err == nil {
		log.Warn().Str("claude_json_path", claudeJsonPath).Msg("⚠️ 非推奨ファイル検出（推奨: 削除またはリネーム）")
	}

	// 統合監視画面の場合
	if strings.Contains(target, ":") {
		// ペイン形式 (session:pane)
		return cl.launchInPane(target, claudeCmd)
	} else {
		// セッション形式
		return cl.launchInSession(target, claudeCmd)
	}
}

// launchInPane ペインでClaude CLIを起動
func (cl *ClaudeLauncher) launchInPane(paneTarget, claudeCmd string) error {
	log.Info().Str("pane", paneTarget).Msg("Launching Claude CLI in pane")

	// ペインにClaude CLIを送信
	cmd := exec.Command("tmux", "send-keys", "-t", paneTarget, claudeCmd, "C-m")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send Claude CLI command to pane %s: %w", paneTarget, err)
	}

	// 起動待機とサイズ調整
	time.Sleep(3 * time.Second)

	// Claude CLI起動後にサイズ調整を実行（tmuxコマンドで実行）
	cl.optimizeClaudeCLIDisplay()

	// プロセス登録
	if claudeProcesses, err := process.GetGlobalProcessManager().CheckClaudeProcesses(); err == nil && len(claudeProcesses) > 0 {
		latestProcess := claudeProcesses[len(claudeProcesses)-1]
		sessionName := strings.Split(paneTarget, ":")[0]
		paneName := strings.Split(paneTarget, ":")[1]
		process.GetGlobalProcessManager().RegisterProcess(sessionName, paneName, claudeCmd, latestProcess.PID)
		log.Info().Int("pid", latestProcess.PID).Str("pane", paneTarget).Msg("Claude CLI process registered")
	}

	return nil
}

// launchInSession セッションでClaude CLIを起動
func (cl *ClaudeLauncher) launchInSession(sessionName, claudeCmd string) error {
	log.Info().Str("session", sessionName).Msg("Launching Claude CLI in session")

	// セッションにClaude CLIを送信
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, claudeCmd, "C-m")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send Claude CLI command to session %s: %w", sessionName, err)
	}

	// 起動待機
	time.Sleep(3 * time.Second)

	// プロセス登録
	if claudeProcesses, err := process.GetGlobalProcessManager().CheckClaudeProcesses(); err == nil && len(claudeProcesses) > 0 {
		latestProcess := claudeProcesses[len(claudeProcesses)-1]
		process.GetGlobalProcessManager().RegisterProcess(sessionName, "main", claudeCmd, latestProcess.PID)
		log.Info().Int("pid", latestProcess.PID).Str("session", sessionName).Msg("Claude CLI process registered")
	}

	return nil
}

// StartAllAgents 全エージェントでClaude CLIを起動
func (cl *ClaudeLauncher) StartAllAgents() error {
	log.Info().Msg("Starting Claude CLI for all agents")

	utils.DisplayInfo("Claude CLI Batch Launch", "Starting system")

	// 統合監視画面の場合
	if cl.tmuxManager.SessionExists(cl.config.SessionName) {
		paneCount, err := cl.tmuxManager.GetPaneCount(cl.config.SessionName)
		if err == nil && paneCount == 6 {
			utils.DisplayInfo("Integrated Monitoring Mode", "Launching Claude CLI in 6 panes")
			return cl.startIntegratedAgents()
		}
	}

	// For individual session mode
	utils.DisplayInfo("Individual Session Mode", "Launching Claude CLI in individual sessions")
	return cl.startIndividualAgents()
}

// startIntegratedAgents 統合監視画面の各ペインでClaude CLIを起動（認証競合防止のため順次実行）
func (cl *ClaudeLauncher) startIntegratedAgents() error {
	agents := []struct {
		pane int
		name string
		file string
	}{
		{1, "PO", "po.md"},
		{2, "Manager", "manager.md"},
		{3, "Dev1", "developer.md"},
		{4, "Dev2", "developer.md"},
		{5, "Dev3", "developer.md"},
		{6, "Dev4", "developer.md"},
	}

	// 認証ファイル競合を防ぐため、順次実行に変更
	for i, agent := range agents {
		paneTarget := fmt.Sprintf("%s:1.%d", cl.config.SessionName, agent.pane)

		utils.DisplayProgress("Claude CLI起動", fmt.Sprintf("%s (ペイン%d) - %d/%d", agent.name, agent.pane, i+1, len(agents)))

		if err := cl.LaunchClaude(paneTarget); err != nil {
			utils.DisplayError("Claude CLI起動失敗", fmt.Errorf("failed to start Claude CLI for %s: %w", agent.name, err))
			return err
		}

		// Claude CLI起動後の安定化待機（OAuth認証競合とファイルアクセス競合防止）
		time.Sleep(5 * time.Second)

		// インストラクションファイルを送信
		utils.DisplayProgress("インストラクション送信", fmt.Sprintf("%s にインストラクションを送信中...", agent.name))

		if err := cl.SendInstructionToAgent(paneTarget, agent.file); err != nil {
			utils.DisplayError("インストラクション送信失敗", fmt.Errorf("failed to send instruction to %s: %w", agent.name, err))
			// インストラクション送信の失敗は致命的ではないので続行
		} else {
			utils.DisplaySuccess("インストラクション送信完了", fmt.Sprintf("%s にインストラクションを送信しました", agent.name))
		}

		// 次のエージェント起動前の待機（OAuth認証とファイルアクセス間隔確保）
		time.Sleep(3 * time.Second)

		utils.DisplaySuccess("Claude CLI起動完了", fmt.Sprintf("%s でClaude CLIが起動しました", agent.name))
	}

	utils.DisplaySuccess("全エージェント起動完了", "全てのエージェントでClaude CLIが起動しました")
	return nil
}

// startIndividualAgents 個別セッションでClaude CLIを起動（認証競合防止のため順次実行）
func (cl *ClaudeLauncher) startIndividualAgents() error {
	agents := []string{"po", "manager", "dev1", "dev2", "dev3", "dev4"}

	// 認証ファイル競合を防ぐため、順次実行に変更
	for i, agent := range agents {
		sessionName := fmt.Sprintf("%s-%s", cl.config.SessionName, agent)

		if !cl.tmuxManager.SessionExists(sessionName) {
			utils.DisplayInfo("セッション確認", fmt.Sprintf("セッション %s が存在しません", sessionName))
			continue
		}

		utils.DisplayProgress("Claude CLI起動", fmt.Sprintf("%s セッション - %d/%d", sessionName, i+1, len(agents)))

		if err := cl.LaunchClaude(sessionName); err != nil {
			utils.DisplayError("Claude CLI起動失敗", fmt.Errorf("failed to start Claude CLI for %s: %w", sessionName, err))
			return err
		}

		// Claude CLI起動後の安定化待機（OAuth認証競合とファイルアクセス競合防止）
		time.Sleep(5 * time.Second)

		// インストラクションファイルを送信
		var instructionFile string
		switch agent {
		case "po":
			instructionFile = "po.md"
		case "manager":
			instructionFile = "manager.md"
		default:
			instructionFile = "developer.md"
		}

		utils.DisplayProgress("インストラクション送信", fmt.Sprintf("%s にインストラクションを送信中...", agent))

		if err := cl.SendInstructionToAgent(sessionName, instructionFile); err != nil {
			utils.DisplayError("インストラクション送信失敗", fmt.Errorf("failed to send instruction to %s: %w", agent, err))
			// インストラクション送信の失敗は致命的ではないので続行
		} else {
			utils.DisplaySuccess("インストラクション送信完了", fmt.Sprintf("%s にインストラクションを送信しました", agent))
		}

		// 次のエージェント起動前の待機（OAuth認証とファイルアクセス間隔確保）
		time.Sleep(3 * time.Second)

		utils.DisplaySuccess("Claude CLI起動完了", fmt.Sprintf("%s でClaude CLIが起動しました", sessionName))
	}

	utils.DisplaySuccess("全エージェント起動完了", "全てのエージェントでClaude CLIが起動しました")
	return nil
}

// SendInstructionToAgent エージェントにインストラクションを送信
func (cl *ClaudeLauncher) SendInstructionToAgent(target, instructionFile string) error {
	log.Info().Str("instruction_file", instructionFile).Str("target", target).Msg("📤 Starting instruction sending")

	// Determine if target is pane format (session:pane) or session format
	if strings.Contains(target, ":") {
		// For pane format, use sendInstructionToPaneWithConfig
		parts := strings.Split(target, ":")
		sessionName := parts[0]
		pane := parts[1]

		// Estimate agent name (from pane number)
		var agent string
		switch {
		case strings.Contains(pane, ".1"):
			agent = "po"
		case strings.Contains(pane, ".2"):
			agent = "manager"
		case strings.Contains(pane, ".3"):
			agent = "dev1"
		case strings.Contains(pane, ".4"):
			agent = "dev2"
		case strings.Contains(pane, ".5"):
			agent = "dev3"
		case strings.Contains(pane, ".6"):
			agent = "dev4"
		default:
			// Estimate agent name from instructionFile as default
			switch instructionFile {
			case "po.md":
				agent = "po"
			case "manager.md":
				agent = "manager"
			default:
				agent = "dev1"
			}
		}

		// Use tmux manager's configuration-based sending function
		return cl.tmuxManager.SendInstructionToPaneWithConfig(sessionName, pane, agent, cl.config.InstructionsDir, nil)
	}

	// For session format, execute conventional processing
	instructionPath := filepath.Join(cl.config.InstructionsDir, instructionFile)

	// Check file existence
	if _, err := os.Stat(instructionPath); err != nil {
		log.Error().Str("instruction_path", instructionPath).Msg("❌ Instruction file check failed")
		return fmt.Errorf("instruction file not found: %s", instructionPath)
	}

	log.Info().Str("target", target).Str("file", instructionFile).Msg("Sending instruction to agent")

	// Read file contents
	_, err := os.ReadFile(instructionPath) // #nosec G304
	if err != nil {
		log.Error().Str("instruction_path", instructionPath).Msg("❌ File read failed")
		return fmt.Errorf("failed to read instruction file: %w", err)
	}

	// Send instructions file content (utilize Claude CLI's Read function)
	// Send file path to Claude CLI to use Read function
	readCmd := fmt.Sprintf("cat \"%s\"", instructionPath)

	log.Info().Str("read_cmd", readCmd).Msg("📋 Sending instruction read command")

	// Send cat command
	cmd := exec.Command("tmux", "send-keys", "-t", target, readCmd, "C-m") // #nosec G204
	if err := cmd.Run(); err != nil {
		log.Warn().Err(err).Msg("⚠️ Instruction read command sending error")
		return fmt.Errorf("failed to send instruction read command: %w", err)
	}

	// Wait for cat command execution
	time.Sleep(2 * time.Second)
	log.Info().Msg("📋 Instruction reading completed")

	// Send Enter key reliably to execute cat command result in Claude CLI
	time.Sleep(2 * time.Second) // Ensure Claude CLI preparation time

	log.Info().Msg("🔄 Starting Enter sending for Claude CLI execution")

	// Send Enter multiple times to put Claude CLI in execution state
	for i := 0; i < 5; i++ {
		cmd = exec.Command("tmux", "send-keys", "-t", target, "C-m")
		if err := cmd.Run(); err != nil {
			log.Warn().Err(err).Int("attempt", i+1).Msg("⚠️ Enter sending error")
		}
		time.Sleep(500 * time.Millisecond) // Extend interval between each Enter
	}

	log.Info().Str("target", target).Msg("✅ Instruction sending completed")
	return nil
}

// GetClaudeStartCommand gets command for launching Claude CLI
func (cl *ClaudeLauncher) GetClaudeStartCommand() string {
	homeDir, _ := os.UserHomeDir()
	return fmt.Sprintf("CLAUDE_CONFIG_DIR=\"%s\" \"%s\" --dangerously-skip-permissions",
		filepath.Join(homeDir, ".claude"), cl.config.ClaudePath)
}

// optimizeClaudeCLIDisplay optimizes Claude CLI display (simplified by removing script command)
func (cl *ClaudeLauncher) optimizeClaudeCLIDisplay() {
	log.Info().Msg("✅ Claude CLI display optimization: Automatically displayed at optimal size due to script command removal")

	// Since script command was removed, Claude CLI automatically recognizes pane size
	// No special optimization processing required
}
