package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ValidateEnvironment システム環境の検証
func ValidateEnvironment() error {
	log.Info().Msg("Validating environment...")

	// Claude CLIパスの検証
	if IsVerboseLogging() {
		displayProgress("Claude CLI検証", "Claude CLIのパスを検証中...")
	}
	claudePath := findClaudeExecutable()
	if claudePath == "" {
		displayError("Claude CLI検証失敗", fmt.Errorf("Claude CLI not found"))
		return fmt.Errorf("Claude CLI not found")
	}
	// Claude CLIパスを簡素化された形式で表示
	displayClaudePath(claudePath)
	if IsVerboseLogging() {
		displaySuccess("Claude CLI検証完了", fmt.Sprintf("Claude CLI を発見: %s", claudePath))
	}

	// Claude CLIが実行可能かテスト
	if IsVerboseLogging() {
		displayProgress("Claude CLI実行テスト", "Claude CLIの実行可能性をテスト中...")
	}
	if !isExecutable(claudePath) {
		displayError("Claude CLI実行テスト失敗", fmt.Errorf("Claude CLI is not executable: %s", claudePath))
		return fmt.Errorf("Claude CLI is not executable: %s", claudePath)
	}
	if IsVerboseLogging() {
		displaySuccess("Claude CLI実行テスト完了", "Claude CLIが正常に実行可能です")
	}

	// 設定ファイルの確認
	if IsVerboseLogging() {
		displayProgress("Claude設定確認", "Claude設定ファイルを確認中...")
	}
	if !checkClaudeConfig() {
		displayError("Claude設定確認失敗", fmt.Errorf("Claude configuration not found or invalid"))
		return fmt.Errorf("Claude configuration not found or invalid")
	}
	if IsVerboseLogging() {
		displaySuccess("Claude設定確認完了", "Claude設定ファイルが正常に確認されました")
	}

	// 必要なディレクトリの確認
	if IsVerboseLogging() {
		displayProgress("必要ディレクトリ確認", "必要なディレクトリの存在を確認中...")
	}
	if !checkRequiredDirectories() {
		displayError("必要ディレクトリ確認失敗", fmt.Errorf("Required directories not found"))
		return fmt.Errorf("Required directories not found")
	}
	if IsVerboseLogging() {
		displaySuccess("必要ディレクトリ確認完了", "必要なディレクトリが全て確認されました")
	}

	log.Info().Msg("Environment validation completed")
	return nil
}



// checkClaudeConfig Claude設定ファイルの確認
func checkClaudeConfig() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	configPath := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(configPath); err != nil {
		return false
	}

	// ファイルが空でないかチェック
	info, err := os.Stat(configPath)
	if err != nil || info.Size() == 0 {
		return false
	}

	return true
}

// checkRequiredDirectories 必要なディレクトリの確認
func checkRequiredDirectories() bool {
	homeDir, _ := os.UserHomeDir()

	requiredDirs := []string{
		 filepath.Join(homeDir, ".claude","claud-code-agents","instructions"),
	}

	for _, dir := range requiredDirs {
		expandedDir := expandPath(dir)
		if _, err := os.Stat(expandedDir); err != nil {
			log.Error().Str("dir", dir).Str("expanded", expandedDir).Msg("Required directory not found")
			return false
		}
	}

	return true
}

// LauncherConfig システム起動設定
type LauncherConfig struct {
	SessionName     string
	Layout          string
	Reset           bool
	WorkingDir      string
	InstructionsDir string
	ClaudePath      string
}

// SystemLauncher システムランチャー
type SystemLauncher struct {
	config      *LauncherConfig
	tmuxManager *TmuxManager
}

// NewSystemLauncher 新しいシステムランチャーを作成
func NewSystemLauncher(config *LauncherConfig) (*SystemLauncher, error) {
	if config == nil {
		return nil, fmt.Errorf("launcher config is required")
	}

	// Claude CLIパスが指定されていない場合は自動検出
	if config.ClaudePath == "" {
		config.ClaudePath = findClaudeExecutable()
		if config.ClaudePath == "" {
			return nil, fmt.Errorf("Claude CLI not found")
		}
	}

	// tmuxManagerを初期化
	tmuxManager := NewTmuxManager(config.SessionName)

	return &SystemLauncher{
		config:      config,
		tmuxManager: tmuxManager,
	}, nil
}

// Launch システムを起動
func (sl *SystemLauncher) Launch() error {
	log.Info().Str("session", sl.config.SessionName).Msg("Starting system launcher")

	displaySubHeader("システムランチャー起動")
	displayInfo("起動モード選択", fmt.Sprintf("レイアウト: %s", sl.config.Layout))

	// レイアウトに応じて起動方法を選択
	switch sl.config.Layout {
	case "individual":
		displayInfo("個別セッション起動", "個別セッション方式でシステムを起動します")
		return sl.startIndividualSessions()
	case "integrated":
		fallthrough
	default:
		displayInfo("統合監視起動", "統合監視画面方式でシステムを起動します")
		return sl.startIntegratedMonitor()
	}
}

// startIndividualSessions 個別セッション方式で起動
func (sl *SystemLauncher) startIndividualSessions() error {
	log.Info().Msg("Starting individual sessions...")

	if IsVerboseLogging() {
		displayProgress("個別セッション起動", "個別セッション方式でシステムを起動中...")
	}

	// 既存セッションのクリーンアップ
	if sl.config.Reset {
		if IsVerboseLogging() {
			displayProgress("セッションクリーンアップ", "既存セッションをクリーンアップ中...")
		}
		sl.cleanupIndividualSessions()
		if IsVerboseLogging() {
			displaySuccess("セッションクリーンアップ完了", "既存セッションのクリーンアップが完了しました")
		}
	}

	// 各エージェントのセッションを作成
	agents := []string{"ceo", "manager", "dev1", "dev2", "dev3", "dev4"}
	for _, agent := range agents {
		sessionName := fmt.Sprintf("%s-%s", sl.config.SessionName, agent)
		
		if sl.tmuxManager.SessionExists(sessionName) {
			if IsVerboseLogging() {
				displayInfo("セッション存在確認", fmt.Sprintf("セッション %s は既に存在します", sessionName))
			}
			log.Info().Str("session", sessionName).Msg("Session already exists")
			continue
		}

		if IsVerboseLogging() {
			displayProgress("エージェントセッション作成", fmt.Sprintf("%s エージェントのセッションを作成中...", agent))
		}
		if err := sl.createAgentSession(sessionName, agent); err != nil {
			displayError("エージェントセッション作成失敗", fmt.Errorf("failed to create session %s: %w", sessionName, err))
			return fmt.Errorf("failed to create session %s: %w", sessionName, err)
		}
		if IsVerboseLogging() {
			displaySuccess("エージェントセッション作成完了", fmt.Sprintf("%s エージェントのセッションが作成されました", agent))
		}
	}

	log.Info().Msg("Individual sessions started successfully")
	if IsVerboseLogging() {
		displaySuccess("個別セッション起動完了", "全ての個別セッションが正常に起動しました")
	}
	return nil
}

// startIntegratedMonitor 統合監視画面方式で起動
func (sl *SystemLauncher) startIntegratedMonitor() error {
	log.Info().Msg("Starting integrated monitor...")

	displayLauncherStart()
	displayLauncherProgress()

	// 既存セッションの確認
	if sl.tmuxManager.SessionExists(sl.config.SessionName) {
		if sl.config.Reset {
			if IsVerboseLogging() {
				displayProgress("既存セッション削除", "既存セッションを削除中...")
			}
			sl.tmuxManager.KillSession(sl.config.SessionName)
			time.Sleep(2 * time.Second)
			if IsVerboseLogging() {
				displaySuccess("既存セッション削除完了", "既存セッションが削除されました")
			}
		} else {
			if IsVerboseLogging() {
				displayInfo("既存セッション接続", fmt.Sprintf("既存セッション %s に接続します", sl.config.SessionName))
			}
			log.Info().Str("session", sl.config.SessionName).Msg("Attaching to existing session")
			return sl.tmuxManager.AttachSession(sl.config.SessionName)
		}
	}

	// 新しいセッションを作成
	if IsVerboseLogging() {
		displayProgress("新規セッション作成", "新しいtmuxセッションを作成中...")
	}
	if err := sl.tmuxManager.CreateSession(sl.config.SessionName); err != nil {
		displayError("新規セッション作成失敗", err)
		return fmt.Errorf("failed to create session: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("新規セッション作成完了", fmt.Sprintf("セッション %s が作成されました", sl.config.SessionName))
	}

	// 統合レイアウトを作成
	if IsVerboseLogging() {
		displayProgress("統合レイアウト作成", "6ペイン統合レイアウトを作成中...")
	}
	if err := sl.createIntegratedLayout(); err != nil {
		displayError("統合レイアウト作成失敗", err)
		return fmt.Errorf("failed to create integrated layout: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("統合レイアウト作成完了", "6ペイン統合レイアウトが作成されました")
	}

	// 各ペインにエージェントを配置
	displayAgentDeployment()
	if err := sl.setupAgentsInPanes(); err != nil {
		displayError("エージェント配置失敗", err)
		return fmt.Errorf("failed to setup agents: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("エージェント配置完了", "全てのエージェントが正常に配置されました")
	}

	// セッションに接続
	if IsVerboseLogging() {
		displayProgress("セッション接続", "セッションに接続中...")
	}
	if err := sl.tmuxManager.AttachSession(sl.config.SessionName); err != nil {
		displayError("セッション接続失敗", err)
		return err
	}

	return nil
}

// createIntegratedLayout 統合レイアウトを作成（claude.shと同じ構成）
func (sl *SystemLauncher) createIntegratedLayout() error {
	sessionName := sl.config.SessionName

	// 6ペイン構成を段階的に作成（claude.shと同じ構成）
	log.Info().Msg("Creating 6-pane layout (claude.sh compatible)...")
	if IsVerboseLogging() {
		displayProgress("6ペインレイアウト作成", "claude.shと同じ6ペイン構成を段階的に作成中...")
	}

	// 1. 左右分割（左側、右側）
	if IsVerboseLogging() {
		displayProgress("左右分割", "左右分割を作成中...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -h -t %s", sessionName)); err != nil {
		displayError("左右分割失敗", err)
		return fmt.Errorf("failed to create horizontal split: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("左右分割完了", "左右分割が作成されました (ペイン0,1)")
	}
	log.Debug().Msg("✓ Horizontal split created (panes 0,1)")

	// 2. 左側を上下分割（上: CEO、下: Manager）
	if IsVerboseLogging() {
		displayProgress("左側上下分割", "左側ペインを上下分割中（CEO/Manager）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.1", sessionName)); err != nil {
		displayError("左側上下分割失敗", err)
		return fmt.Errorf("failed to split left pane: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("左側上下分割完了", "左側ペインが上下分割されました (CEO/Manager)")
	}
	log.Debug().Msg("✓ Left pane split for CEO/Manager (panes 0,1,2)")

	// 3. 右側を上下分割（上: Dev1、下: 残り）
	if IsVerboseLogging() {
		displayProgress("右側上下分割", "右側ペインを上下分割中（Dev1/残り）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.3", sessionName)); err != nil {
		displayError("右側上下分割失敗", err)
		return fmt.Errorf("failed to split right pane: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("右側上下分割完了", "右側ペインが上下分割されました (Dev1/残り)")
	}
	log.Debug().Msg("✓ Right pane split for Dev1 (panes 0,1,2,3)")

	// 4. 右下をさらに分割（Dev2用）
	if IsVerboseLogging() {
		displayProgress("右下分割", "右下ペインをさらに分割中（Dev2用）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.4", sessionName)); err != nil {
		displayError("右下分割失敗", err)
		return fmt.Errorf("failed to split bottom right pane for Dev2: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("右下分割完了", "右下ペインが分割されました (Dev2用)")
	}
	log.Debug().Msg("✓ Bottom right split for Dev2 (panes 0,1,2,3,4)")

	// 5. 最後のペインを分割（Dev3用）
	if IsVerboseLogging() {
		displayProgress("最終分割", "最後のペインを分割中（Dev3用）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.5", sessionName)); err != nil {
		displayError("最終分割失敗", err)
		return fmt.Errorf("failed to split last pane for Dev3: %w", err)
	}
	if IsVerboseLogging() {
		displaySuccess("最終分割完了", "最終分割が完了しました (Dev3用)")
	}
	log.Debug().Msg("✓ Final split for Dev3 (panes 0,1,2,3,4,5)")

	log.Info().Msg("6-pane layout created successfully (claude.sh compatible)")

	// レイアウト最適化
	if IsVerboseLogging() {
		displayProgress("レイアウト最適化", "ペインレイアウトを最適化中...")
	}
	sl.optimizeLayout()
	if IsVerboseLogging() {
		displaySuccess("レイアウト最適化完了", "ペインレイアウトが最適化されました")
	}

	return nil
}

// optimizeLayout レイアウトを最適化（claude.shと同じ構成）
func (sl *SystemLauncher) optimizeLayout() {
	sessionName := sl.config.SessionName

	// 右側のDev1-Dev4のペインを等間隔に調整（claude.shと同じ構成）
	resizeCommands := []string{
		fmt.Sprintf("resize-pane -t %s:1.3 -p 25", sessionName), // Dev1
		fmt.Sprintf("resize-pane -t %s:1.4 -p 25", sessionName), // Dev2
		fmt.Sprintf("resize-pane -t %s:1.5 -p 25", sessionName), // Dev3
		fmt.Sprintf("resize-pane -t %s:1.6 -p 25", sessionName), // Dev4
	}

	for _, cmd := range resizeCommands {
		sl.executeCommand(cmd)
	}

	// ペインタイトルの設定
	sl.executeCommand(fmt.Sprintf("set-option -t %s pane-border-status top", sessionName))
	sl.executeCommand(fmt.Sprintf("set-option -t %s pane-border-format \"#T\"", sessionName))
	sl.executeCommand(fmt.Sprintf("set-window-option -t %s automatic-rename off", sessionName))
	sl.executeCommand(fmt.Sprintf("set-window-option -t %s allow-rename off", sessionName))
}

// setupAgentsInPanes 各ペインにエージェントを配置（claude.shと同じ構成）
func (sl *SystemLauncher) setupAgentsInPanes() error {
	// claude.shと同じ構成: 左側にCEO/Manager、右側にDev1-Dev4
	agents := []struct {
		pane int
		name string
		file string
	}{
		{1, "CEO", "ceo.md"},       // 左上
		{2, "Manager", "manager.md"}, // 左下
		{3, "Dev1", "developer.md"}, // 右上
		{4, "Dev2", "developer.md"}, // 右上中
		{5, "Dev3", "developer.md"}, // 右下中
		{6, "Dev4", "developer.md"}, // 右下
	}

	for _, agent := range agents {
		if IsVerboseLogging() {
			displayProgress("エージェント配置", fmt.Sprintf("%s エージェントをペイン%dに配置中...", agent.name, agent.pane))
		}
		if err := sl.setupAgent(agent.pane, agent.name, agent.file); err != nil {
			displayError("エージェント配置失敗", fmt.Errorf("failed to setup agent %s: %w", agent.name, err))
			return fmt.Errorf("failed to setup agent %s: %w", agent.name, err)
		}
		if IsVerboseLogging() {
			displaySuccess("エージェント配置完了", fmt.Sprintf("%s エージェントがペイン%dに配置されました", agent.name, agent.pane))
		}
	}

	return nil
}

// setupAgent エージェントをペインにセットアップ
func (sl *SystemLauncher) setupAgent(pane int, name, instructionFile string) error {
	sessionName := sl.config.SessionName
	paneTarget := fmt.Sprintf("%s:1.%d", sessionName, pane)

	// ペインタイトルを設定
	if IsVerboseLogging() {
		displayProgress("ペインタイトル設定", fmt.Sprintf("%s のペインタイトルを設定中...", name))
	}
	sl.executeCommand(fmt.Sprintf("select-pane -t %s -T %s", paneTarget, name))

	// 作業ディレクトリに移動
	if IsVerboseLogging() {
		displayProgress("作業ディレクトリ移動", fmt.Sprintf("%s の作業ディレクトリに移動中...", name))
	}
	sl.sendKeys(paneTarget, fmt.Sprintf("cd '%s'", sl.config.WorkingDir))

	// Claude CLIを起動
	if IsVerboseLogging() {
		displayProgress("Claude CLI起動", fmt.Sprintf("%s のClaude CLIを起動中...", name))
	}
	claudeCmd := fmt.Sprintf("script -q /dev/null \"%s\" --dangerously-skip-permissions", sl.config.ClaudePath)
	sl.sendKeys(paneTarget, claudeCmd)

	// 少し待ってからインストラクションファイルを送信
	time.Sleep(2 * time.Second)

	// インストラクションファイルのパス
	if IsVerboseLogging() {
		displayProgress("インストラクション送信", fmt.Sprintf("%s のインストラクションファイルを送信中...", name))
	}
	instructionPath := filepath.Join(sl.config.InstructionsDir, instructionFile)
	sl.sendKeys(paneTarget, fmt.Sprintf("cat \"%s\"", instructionPath))

	// プロンプト状態に戻す
	time.Sleep(1 * time.Second)
	sl.sendKeys(paneTarget, "")

	if IsVerboseLogging() {
		displaySuccess("エージェント設定完了", fmt.Sprintf("%s の設定が完了しました", name))
	}
	return nil
}

// createAgentSession エージェントのセッションを作成
func (sl *SystemLauncher) createAgentSession(sessionName, agent string) error {
	// セッションを作成
	if err := sl.tmuxManager.CreateSession(sessionName); err != nil {
		return err
	}

	// ウィンドウ名を設定
	sl.executeCommand(fmt.Sprintf("rename-window -t %s %s", sessionName, sessionName))

	// 作業ディレクトリに移動
	sl.sendKeys(sessionName, fmt.Sprintf("cd '%s'", sl.config.WorkingDir))

	// インストラクションファイルの選択
	var instructionFile string
	switch agent {
	case "ceo":
		instructionFile = "ceo.md"
	case "manager":
		instructionFile = "manager.md"
	default:
		instructionFile = "developer.md"
	}

	// Claude CLIを起動
	claudeCmd := fmt.Sprintf("script -q /dev/null \"%s\" --dangerously-skip-permissions", sl.config.ClaudePath)
	sl.sendKeys(sessionName, claudeCmd)

	// インストラクションファイルを送信
	time.Sleep(2 * time.Second)
	instructionPath := filepath.Join(sl.config.InstructionsDir, instructionFile)
	sl.sendKeys(sessionName, fmt.Sprintf("cat \"%s\"", instructionPath))

	// プロンプト状態に戻す
	time.Sleep(1 * time.Second)
	sl.sendKeys(sessionName, "")

	return nil
}

// executeCommand tmuxコマンドを実行
func (sl *SystemLauncher) executeCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	log.Debug().Str("command", cmd).Msg("Executing tmux command")
	
	execCmd := exec.Command("tmux", parts...)
	if output, err := execCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux command failed: %s (output: %s)", cmd, string(output))
	}

	return nil
}

// sendKeys tmuxペインにキーを送信
func (sl *SystemLauncher) sendKeys(target, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "C-m")
	return cmd.Run()
}

// cleanupIndividualSessions 個別セッションをクリーンアップ
func (sl *SystemLauncher) cleanupIndividualSessions() {
	agents := []string{"ceo", "manager", "dev1", "dev2", "dev3", "dev4"}
	for _, agent := range agents {
		sessionName := fmt.Sprintf("%s-%s", sl.config.SessionName, agent)
		sl.tmuxManager.KillSession(sessionName)
	}
}

// RunIntegrationTests 統合テストを実行
func RunIntegrationTests() error {
	log.Info().Msg("Starting integration tests...")

	// 環境検証テスト
	if err := ValidateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}
	log.Info().Msg("✓ Environment validation passed")

	// tmux接続テスト
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux is not available: %w", err)
	}
	log.Info().Msg("✓ Tmux availability test passed")

	// Claude CLI実行テスト
	claudePath := findClaudeExecutable()
	if claudePath == "" {
		return fmt.Errorf("Claude CLI not found")
	}
	if !isExecutable(claudePath) {
		return fmt.Errorf("Claude CLI is not executable")
	}
	log.Info().Msg("✓ Claude CLI execution test passed")

	// 設定ファイルテスト
	if !checkClaudeConfig() {
		return fmt.Errorf("Claude configuration test failed")
	}
	log.Info().Msg("✓ Claude configuration test passed")

	// インストラクションファイルテスト
	if !checkInstructionFiles() {
		return fmt.Errorf("instruction files test failed")
	}
	log.Info().Msg("✓ Instruction files test passed")

	// セッション作成・削除テスト
	if err := testSessionOperations(); err != nil {
		return fmt.Errorf("session operations test failed: %w", err)
	}
	log.Info().Msg("✓ Session operations test passed")

	log.Info().Msg("All integration tests passed successfully")
	return nil
}

// checkInstructionFiles インストラクションファイルの確認
func checkInstructionFiles() bool {
	homeDir, _ := os.UserHomeDir()

	instructionsDir := filepath.Join(homeDir, ".claude","claud-code-agents","instructions")
	files := []string{"ceo.md", "manager.md", "developer.md"}

	for _, file := range files {
		filePath := filepath.Join(instructionsDir, file)
		if _, err := os.Stat(filePath); err != nil {
			log.Error().Str("file", filePath).Msg("Instruction file not found")
			return false
		}
	}

	return true
}

// testSessionOperations セッション操作のテスト
func testSessionOperations() error {
	tmuxManager := NewTmuxManager("test")
	testSessionName := "test-session-" + fmt.Sprintf("%d", time.Now().Unix())

	// セッション作成テスト
	if err := tmuxManager.CreateSession(testSessionName); err != nil {
		return fmt.Errorf("failed to create test session: %w", err)
	}

	// セッション存在確認
	if !tmuxManager.SessionExists(testSessionName) {
		return fmt.Errorf("test session not found after creation")
	}

	// セッション削除テスト
	if err := tmuxManager.KillSession(testSessionName); err != nil {
		return fmt.Errorf("failed to kill test session: %w", err)
	}

	// セッション削除確認
	time.Sleep(1 * time.Second)
	if tmuxManager.SessionExists(testSessionName) {
		return fmt.Errorf("test session still exists after deletion")
	}

	return nil
}
