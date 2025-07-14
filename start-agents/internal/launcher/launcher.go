package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/auth"
	"github.com/shivase/claude-code-agents/internal/process"
	"github.com/shivase/claude-code-agents/internal/tmux"
	"github.com/shivase/claude-code-agents/internal/utils"
)

// ValidateEnvironment システム環境の検証
func ValidateEnvironment() error {
	log.Info().Msg("Validating environment...")

	// Claude CLIパスの設定（デフォルト）
	claudePath := findClaudeExecutableHelper()
	if claudePath == "" {
		log.Error().Msg("❌ Claude CLI検証失敗 Claude CLIが見つかりません")
		return fmt.Errorf("claude CLI not found")
	}

	// Claude認証状態の確認（設定ファイル確認のみ）
	claudeAuth := auth.NewClaudeAuthManager()
	if authStatus, err := claudeAuth.CheckAuthenticationStatus(); err != nil {
		return fmt.Errorf("claude authentication check failed: %w", err)
	} else if !authStatus.IsAuthenticated {
		log.Warn().Msg("Claude認証が完了していません")
	}
	log.Info().Msg("✅ Claude設定ファイル確認完了")

	// Claude CLIパス情報を表示
	log.Info().Str("claude_path", claudePath).Msg("✅ Claude CLI検証完了")

	// 必要なディレクトリの確認
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("必要ディレクトリ確認", "必要なディレクトリの存在を確認中...")
	}
	if !checkRequiredDirectories() {
		utils.DisplayError("必要ディレクトリ確認失敗", fmt.Errorf("required directories not found"))
		return fmt.Errorf("required directories not found")
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("必要ディレクトリ確認完了", "必要なディレクトリが全て確認されました")
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
		filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions"),
	}

	for _, dir := range requiredDirs {
		expandedDir := expandPathHelper(dir)
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
	tmuxManager *tmux.TmuxManagerImpl
}

// NewSystemLauncher 新しいシステムランチャーを作成
func NewSystemLauncher(config *LauncherConfig) (*SystemLauncher, error) {
	if config == nil {
		return nil, fmt.Errorf("launcher config is required")
	}

	// Claude CLIパスが指定されていない場合は自動検出
	if config.ClaudePath == "" {
		config.ClaudePath = findClaudeExecutableHelper()
		if config.ClaudePath == "" {
			return nil, fmt.Errorf("claude CLI not found")
		}
	}

	// tmuxManagerを初期化
	tmuxManager := tmux.NewTmuxManager(config.SessionName)

	return &SystemLauncher{
		config:      config,
		tmuxManager: tmuxManager,
	}, nil
}

// Launch システムを起動
func (sl *SystemLauncher) Launch() error {
	log.Info().Str("session", sl.config.SessionName).Msg("Starting system launcher")

	// 統一フォーマットで起動情報を表示
	log.Info().Msg("📌 システムランチャー起動")
	log.Info().Msg("-------------------------------------")
	log.Info().Str("layout", sl.config.Layout).Msg("ℹ️ 起動モード選択")

	// 既存のClaude CLIプロセスをクリーンアップ
	if utils.IsVerboseLogging() {
		log.Info().Msg("🔄 プロセスクリーンアップ 既存のClaude CLIプロセスをクリーンアップ中")
		log.Info().Msg("✅ プロセスクリーンアップ完了 既存のClaude CLIプロセスをクリーンアップしました")
	}

	// レイアウトに応じて起動方法を選択
	switch sl.config.Layout {
	case "individual":
		log.Info().Msg("ℹ️ 個別セッション起動 個別セッション方式でシステムを起動します")
		log.Info().Msg("🔄 個別セッション起動 個別セッション方式でシステムを起動中")
		return sl.startIndividualSessions()
	case "integrated":
		fallthrough
	default:
		return sl.startIntegratedMonitor()
	}
}

// startIndividualSessions 個別セッション方式で起動
func (sl *SystemLauncher) startIndividualSessions() error {
	log.Info().Msg("Starting individual sessions...")

	if utils.IsVerboseLogging() {
		utils.DisplayProgress("個別セッション起動", "個別セッション方式でシステムを起動中...")
	}

	// 既存セッションのクリーンアップ
	if sl.config.Reset {
		if utils.IsVerboseLogging() {
			utils.DisplayProgress("セッションクリーンアップ", "既存セッションをクリーンアップ中...")
		}
		sl.cleanupIndividualSessions()
		if utils.IsVerboseLogging() {
			utils.DisplaySuccess("セッションクリーンアップ完了", "既存セッションのクリーンアップが完了しました")
		}
	}

	// 各エージェントのセッションを作成
	agents := []string{"po", "manager", "dev1", "dev2", "dev3", "dev4"}
	for _, agent := range agents {
		sessionName := fmt.Sprintf("%s-%s", sl.config.SessionName, agent)

		if sl.tmuxManager.SessionExists(sessionName) {
			if utils.IsVerboseLogging() {
				utils.DisplayInfo("セッション存在確認", fmt.Sprintf("セッション %s は既に存在します", sessionName))
			}
			log.Info().Str("session", sessionName).Msg("Session already exists")
			continue
		}

		if utils.IsVerboseLogging() {
			utils.DisplayProgress("エージェントセッション作成", fmt.Sprintf("%s エージェントのセッションを作成中...", agent))
		}
		if err := sl.createAgentSession(sessionName, agent); err != nil {
			utils.DisplayError("エージェントセッション作成失敗", fmt.Errorf("failed to create session %s: %w", sessionName, err))
			return fmt.Errorf("failed to create session %s: %w", sessionName, err)
		}
		if utils.IsVerboseLogging() {
			utils.DisplaySuccess("エージェントセッション作成完了", fmt.Sprintf("%s エージェントのセッションが作成されました", agent))
		}
	}

	log.Info().Msg("Individual sessions started successfully")
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("個別セッション起動完了", "全ての個別セッションが正常に起動しました")
	}
	return nil
}

// startIntegratedMonitor 統合監視画面方式で起動
func (sl *SystemLauncher) startIntegratedMonitor() error {
	log.Info().Msg("Starting integrated monitor...")

	log.Info().Msg("ℹ️ 統合監視起動 統合監視画面方式でシステムを起動します")
	log.Info().Msg("🔄 統合監視起動 統合監視画面方式でシステムを起動中")

	utils.DisplayLauncherStart()
	utils.DisplayLauncherProgress()

	// 既存セッションの確認
	if sl.tmuxManager.SessionExists(sl.config.SessionName) {
		if sl.config.Reset {
			if utils.IsVerboseLogging() {
				utils.DisplayProgress("既存セッション削除", "既存セッションを削除中...")
			}
			if err := sl.tmuxManager.KillSession(sl.config.SessionName); err != nil {
				log.Warn().Err(err).Str("session", sl.config.SessionName).Msg("Failed to kill existing session")
			}
			time.Sleep(2 * time.Second)
			if utils.IsVerboseLogging() {
				utils.DisplaySuccess("既存セッション削除完了", "既存セッションが削除されました")
			}
		} else {
			if utils.IsVerboseLogging() {
				utils.DisplayInfo("既存セッション接続", fmt.Sprintf("既存セッション %s に接続します", sl.config.SessionName))
			}
			log.Info().Str("session", sl.config.SessionName).Msg("Attaching to existing session")
			return sl.tmuxManager.AttachSession(sl.config.SessionName)
		}
	}

	// 新しいセッションを作成
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("新規セッション作成", "新しいtmuxセッションを作成中...")
	}
	if err := sl.tmuxManager.CreateSession(sl.config.SessionName); err != nil {
		utils.DisplayError("新規セッション作成失敗", err)
		return fmt.Errorf("failed to create session: %w", err)
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("新規セッション作成完了", fmt.Sprintf("セッション %s が作成されました", sl.config.SessionName))
	}

	// 統合レイアウトを作成
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("統合レイアウト作成", "6ペイン統合レイアウトを作成中...")
	}
	if err := sl.createIntegratedLayout(); err != nil {
		utils.DisplayError("統合レイアウト作成失敗", err)
		return fmt.Errorf("failed to create integrated layout: %w", err)
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("統合レイアウト作成完了", "6ペイン統合レイアウトが作成されました")
	}

	// 各ペインにエージェントを配置
	log.Info().Msg("🔄 エージェント配置 6ペインにエージェントを配置中")
	sl.setupAgentsInPanes()
	log.Info().Msg("✅ エージェント配置完了 全てのエージェントが正常に配置されました")

	// セッションに接続
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("セッション接続", "セッションに接続中...")
	}
	if err := sl.tmuxManager.AttachSession(sl.config.SessionName); err != nil {
		utils.DisplayError("セッション接続失敗", err)
		return err
	}

	return nil
}

// createIntegratedLayout 統合レイアウトを作成（claude.shと同じ構成）
func (sl *SystemLauncher) createIntegratedLayout() error {
	sessionName := sl.config.SessionName

	// 6ペイン構成を段階的に作成（claude.shと同じ構成）
	log.Info().Msg("Creating 6-pane layout (claude.sh compatible)...")
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("6ペインレイアウト作成", "claude.shと同じ6ペイン構成を段階的に作成中...")
	}

	// 1. 左右分割（左側、右側）
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("左右分割", "左右分割を作成中...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -h -t %s", sessionName)); err != nil {
		utils.DisplayError("左右分割失敗", err)
		return fmt.Errorf("failed to create horizontal split: %w", err)
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("左右分割完了", "左右分割が作成されました (ペイン0,1)")
	}
	log.Debug().Msg("✓ Horizontal split created (panes 0,1)")

	// 2. 左側を上下分割（上: PO、下: Manager）
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("左側上下分割", "左側ペインを上下分割中（PO/Manager）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.1", sessionName)); err != nil {
		utils.DisplayError("左側上下分割失敗", err)
		return fmt.Errorf("failed to split left pane: %w", err)
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("左側上下分割完了", "左側ペインが上下分割されました (PO/Manager)")
	}
	log.Debug().Msg("✓ Left pane split for PO/Manager (panes 0,1,2)")

	// 3. 右側を上下分割（上: Dev1、下: 残り）
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("右側上下分割", "右側ペインを上下分割中（Dev1/残り）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.2", sessionName)); err != nil {
		utils.DisplayError("右側上下分割失敗", err)
		return fmt.Errorf("failed to split right pane: %w", err)
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("右側上下分割完了", "右側ペインが上下分割されました (Dev1/残り)")
	}
	log.Debug().Msg("✓ Right pane split for Dev1 (panes 0,1,2,3)")

	// 4. 右下をさらに分割（Dev2用）
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("右下分割", "右下ペインをさらに分割中（Dev2用）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.3", sessionName)); err != nil {
		utils.DisplayError("右下分割失敗", err)
		return fmt.Errorf("failed to split bottom right pane for Dev2: %w", err)
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("右下分割完了", "右下ペインが分割されました (Dev2用)")
	}
	log.Debug().Msg("✓ Bottom right split for Dev2 (panes 0,1,2,3,4)")

	// 5. 最後のペインを分割（Dev3用）
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("最終分割", "最後のペインを分割中（Dev3用）...")
	}
	if err := sl.executeCommand(fmt.Sprintf("split-window -v -t %s:1.4", sessionName)); err != nil {
		utils.DisplayError("最終分割失敗", err)
		return fmt.Errorf("failed to split last pane for Dev3: %w", err)
	}
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("最終分割完了", "最終分割が完了しました (Dev3用)")
	}
	log.Debug().Msg("✓ Final split for Dev3 (panes 0,1,2,3,4,5)")

	log.Info().Msg("6-pane layout created successfully (claude.sh compatible)")

	// レイアウト最適化
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("レイアウト最適化", "ペインレイアウトを最適化中...")
	}
	sl.optimizeLayout()
	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("レイアウト最適化完了", "ペインレイアウトが最適化されました")
	}

	return nil
}

// optimizeLayout レイアウトを最適化（claude.shと同じ構成）
func (sl *SystemLauncher) optimizeLayout() {
	sessionName := sl.config.SessionName

	// ペインサイズを最適化（左側50%、右側50%に正確に分割）
	resizeCommands := []string{
		// 最初に左側ペイン（PO+Manager）を50%に設定
		fmt.Sprintf("resize-pane -t %s:1.1 -p 50", sessionName), // 左側全体を50%に
		// 左側内部でPOとManagerを上下均等分割
		fmt.Sprintf("resize-pane -t %s:1.2 -p 50", sessionName), // Manager を左側の50%に
		// 右側のペイン群は自動的に残りの50%を占有する
		// 右側内部でDev1-4を均等分割（25%ずつ）
		fmt.Sprintf("resize-pane -t %s:1.4 -p 25", sessionName), // Dev2
		fmt.Sprintf("resize-pane -t %s:1.5 -p 25", sessionName), // Dev3
		fmt.Sprintf("resize-pane -t %s:1.6 -p 25", sessionName), // Dev4
	}

	for _, cmd := range resizeCommands {
		if err := sl.executeCommand(cmd); err != nil {
			log.Warn().Err(err).Str("cmd", cmd).Msg("Failed to execute resize command")
		}
	}

	// Claude CLI表示最適化のためのtmux設定
	optimizationCommands := []string{
		// ペインタイトルの設定
		fmt.Sprintf("set-option -t %s pane-border-status top", sessionName),
		fmt.Sprintf("set-option -t %s pane-border-format \"#T\"", sessionName),
		// ウィンドウ名を最上部に表示し、セッション名を表示
		fmt.Sprintf("set-option -t %s status-position top", sessionName),
		fmt.Sprintf("set-window-option -t %s window-status-format \" %s \"", sessionName, sessionName),
		fmt.Sprintf("set-window-option -t %s window-status-current-format \" [%s] \"", sessionName, sessionName),
		fmt.Sprintf("set-window-option -t %s automatic-rename off", sessionName),
		fmt.Sprintf("set-window-option -t %s allow-rename off", sessionName),
		// ウィンドウ名をセッション名に設定
		fmt.Sprintf("rename-window -t %s \"%s\"", sessionName, sessionName),
		// 左右のペインサイズを完全に50:50に保持する設定
		fmt.Sprintf("set-window-option -t %s main-pane-width 50%%", sessionName),
		fmt.Sprintf("set-window-option -t %s main-pane-height 100%%", sessionName),
		// Claude CLI表示最適化（サイズ問題を解決）
		fmt.Sprintf("set-option -t %s default-terminal \"screen-256color\"", sessionName),
		// ペインの境界線を最小化
		fmt.Sprintf("set-option -t %s pane-border-lines simple", sessionName),
		// 履歴バッファサイズを増加
		fmt.Sprintf("set-option -t %s history-limit 50000", sessionName),
		// ウィンドウサイズの自動調整を無効化
		fmt.Sprintf("set-window-option -t %s aggressive-resize off", sessionName),
		// ペインの同期を無効化
		fmt.Sprintf("set-window-option -t %s synchronize-panes off", sessionName),
	}

	for _, cmd := range optimizationCommands {
		if err := sl.executeCommand(cmd); err != nil {
			log.Warn().Err(err).Str("cmd", cmd).Msg("Failed to execute optimization command")
		}
	}

	// 画面の再描画を強制
	if err := sl.executeCommand(fmt.Sprintf("refresh-client -t %s", sessionName)); err != nil {
		log.Warn().Err(err).Msg("Failed to refresh client")
	}

	// ペインを同期させて表示を更新
	if err := sl.executeCommand(fmt.Sprintf("synchronize-panes -t %s -d", sessionName)); err != nil {
		log.Warn().Err(err).Msg("Failed to synchronize panes")
	}
}

// setupAgentsInPanes 各ペインにエージェントを配置（claude.shと同じ構成）
func (sl *SystemLauncher) setupAgentsInPanes() {
	// claude.shと同じ構成: 左側にPO/Manager、右側にDev1-Dev4
	agents := []struct {
		pane int
		name string
		file string
	}{
		{1, "PO", "po.md"},           // 左上
		{2, "Manager", "manager.md"}, // 左下
		{3, "Dev1", "developer.md"},  // 右上
		{4, "Dev2", "developer.md"},  // 右上中
		{5, "Dev3", "developer.md"},  // 右下中
		{6, "Dev4", "developer.md"},  // 右下
	}

	// 順次実行（並列実行を避けるため）
	for i, agent := range agents {
		log.Info().Str("agent", agent.name).Int("current", i+1).Int("total", len(agents)).Msg("🚀 エージェント開始")

		if utils.IsVerboseLogging() {
			utils.DisplayProgress("エージェント配置", fmt.Sprintf("%s エージェントをペイン%dに配置中...", agent.name, agent.pane))
		}

		// ペインタイトルを設定
		paneTarget := fmt.Sprintf("%s:1.%d", sl.config.SessionName, agent.pane)
		if err := sl.executeCommand(fmt.Sprintf("select-pane -t %s -T %s", paneTarget, agent.name)); err != nil {
			log.Warn().Err(err).Msgf("Failed to set pane title for %s", agent.name)
		}

		sl.setupAgent(agent.pane, agent.name, agent.file)

		if utils.IsVerboseLogging() {
			utils.DisplaySuccess("エージェント配置完了", fmt.Sprintf("%s エージェントがペイン%dに配置されました", agent.name, agent.pane))
		}

		log.Info().Str("agent", agent.name).Int("current", i+1).Int("total", len(agents)).Msg("✅ エージェント完了")

		// エージェント間に少し待機時間を入れる（リソース競合を避けるため）
		if i < len(agents)-1 { // 最後のエージェント以外
			log.Info().Msg("⏳ 次のエージェント準備中")
			time.Sleep(2 * time.Second)
		}
	}
}

// setupAgent エージェントをペインにセットアップ
func (sl *SystemLauncher) setupAgent(pane int, name, instructionFile string) {
	sessionName := sl.config.SessionName
	paneTarget := fmt.Sprintf("%s:1.%d", sessionName, pane)

	// ペインタイトルを設定
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("ペインタイトル設定", fmt.Sprintf("%s のペインタイトルを設定中...", name))
	}
	if err := sl.executeCommand(fmt.Sprintf("select-pane -t %s -T %s", paneTarget, name)); err != nil {
		log.Warn().Err(err).Msgf("Failed to set pane title for %s", name)
	}

	// 作業ディレクトリに移動
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("作業ディレクトリ移動", fmt.Sprintf("%s の作業ディレクトリに移動中...", name))
	}
	if err := sl.sendKeys(paneTarget, fmt.Sprintf("cd '%s'", sl.config.WorkingDir)); err != nil {
		log.Warn().Err(err).Msg("Failed to send cd command")
	}

	// ペインサイズ設定は環境変数経由でClaude CLI起動時に行われるため、
	// ここでは追加のコマンド送信を行わない
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("ペインサイズ確認", fmt.Sprintf("%s のペインサイズを確認中...", name))

		// ログ出力用にペインサイズを取得（コマンド送信はしない）
		cmd := exec.Command("tmux", "display-message", "-t", paneTarget, "-p", "#{pane_width}x#{pane_height}") // #nosec G204
		sizeOutput, err := cmd.Output()
		if err == nil {
			size := strings.TrimSpace(string(sizeOutput))
			log.Info().Str("name", name).Str("size", size).Msg("ℹ️ ペインサイズ")
		}
	}

	// 既存のClaude CLIプロセスをチェック・終了
	pm := process.GetGlobalProcessManager()
	if claudeProcesses, err := pm.CheckClaudeProcesses(); err == nil && len(claudeProcesses) > 0 {
		if utils.IsVerboseLogging() {
			utils.DisplayProgress("プロセスクリーンアップ", fmt.Sprintf("%s の既存プロセスをクリーンアップ中...", name))
		}
		if err := pm.TerminateClaudeProcesses(); err != nil {
			log.Warn().Err(err).Msg("Failed to terminate Claude processes")
		}
		time.Sleep(1 * time.Second)
	}

	// Claude CLIを起動
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("Claude CLI起動", fmt.Sprintf("%s のClaude CLIを起動中...", name))
	}
	homeDir, _ := os.UserHomeDir()

	// tmuxペインサイズを取得して環境変数に設定
	widthCmd := exec.Command("tmux", "display-message", "-t", paneTarget, "-p", "#{pane_width}")   // #nosec G204
	heightCmd := exec.Command("tmux", "display-message", "-t", paneTarget, "-p", "#{pane_height}") // #nosec G204

	widthOutput, _ := widthCmd.Output()
	heightOutput, _ := heightCmd.Output()

	// ペインサイズを取得（デフォルト値を使用）
	width := strings.TrimSpace(string(widthOutput))
	height := strings.TrimSpace(string(heightOutput))

	if width == "" {
		width = "80"
		log.Debug().Msg("Using default width: 80")
	}
	if height == "" {
		height = "24"
		log.Debug().Msg("Using default height: 24")
	}

	// サイズ情報をログに記録
	log.Debug().Str("width", width).Str("height", height).Msg("Pane size configured")

	// Claude CLIを直接起動（scriptコマンドを使用せず、サイズ問題を解決）
	claudeCmd := fmt.Sprintf("CLAUDE_CONFIG_DIR=\"%s\" \"%s\" --dangerously-skip-permissions",
		filepath.Join(homeDir, ".claude"), sl.config.ClaudePath)
	if err := sl.sendKeys(paneTarget, claudeCmd); err != nil {
		log.Warn().Err(err).Msg("Failed to send claude command")
	}

	// Claude CLIの起動を待機
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("Claude CLI起動待機", fmt.Sprintf("%s のClaude CLI起動を待機中...", name))
	}

	// Claude CLIが完全に起動するまで待機
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		if utils.IsVerboseLogging() {
			utils.DisplayProgress("起動確認", fmt.Sprintf("%s のClaude CLI起動確認中... (%d/10)", name, i+1))
		}
	}

	// Claude CLI起動後にレイアウトを強制リセット
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("レイアウト初期化", fmt.Sprintf("%s のClaude CLI表示を初期化中...", name))
	}

	// Claude CLI起動後にサイズ調整を実行（tmuxコマンドで実行）
	sl.optimizeClaudeCLIDisplay(name)

	// インストラクションファイルを送信
	if instructionFile != "" {
		if utils.IsVerboseLogging() {
			utils.DisplayProgress("インストラクション送信", fmt.Sprintf("%s にインストラクションファイルを送信中...", name))
		}

		// Claude起動設定を作成してインストラクション送信機能を使用
		claudeLauncher := NewClaudeLauncher(&LauncherConfig{
			SessionName:     sl.config.SessionName,
			ClaudePath:      sl.config.ClaudePath,
			WorkingDir:      sl.config.WorkingDir,
			InstructionsDir: sl.config.InstructionsDir,
		})

		if err := claudeLauncher.SendInstructionToAgent(paneTarget, instructionFile); err != nil {
			if utils.IsVerboseLogging() {
				utils.DisplayError("インストラクション送信失敗", fmt.Errorf("failed to send instruction to %s: %w", name, err))
			}
			// インストラクション送信の失敗は致命的ではないので続行
		} else {
			if utils.IsVerboseLogging() {
				utils.DisplaySuccess("インストラクション送信完了", fmt.Sprintf("%s にインストラクションを送信しました", name))
			}
		}
	}

	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("エージェント設定完了", fmt.Sprintf("%s の設定が完了しました", name))
	}
}

// createAgentSession エージェントのセッションを作成
func (sl *SystemLauncher) createAgentSession(sessionName, agent string) error {
	// セッションを作成
	if err := sl.tmuxManager.CreateSession(sessionName); err != nil {
		return err
	}

	// ウィンドウ名を設定
	if err := sl.executeCommand(fmt.Sprintf("rename-window -t %s %s", sessionName, sessionName)); err != nil {
		log.Warn().Err(err).Msg("Failed to rename tmux window")
	}

	// 作業ディレクトリに移動
	if err := sl.sendKeys(sessionName, fmt.Sprintf("cd '%s'", sl.config.WorkingDir)); err != nil {
		log.Warn().Err(err).Msg("Failed to send cd command to session")
	}

	// 注意: インストラクションファイルの選択と送信は従来の設定で処理される

	// 既存のClaude CLIプロセスをチェック・終了
	pm := process.GetGlobalProcessManager()
	if claudeProcesses, err := pm.CheckClaudeProcesses(); err == nil && len(claudeProcesses) > 0 {
		if err := pm.TerminateClaudeProcesses(); err != nil {
			log.Warn().Err(err).Msg("Failed to terminate Claude processes")
		}
		time.Sleep(1 * time.Second)
	}

	// Claude CLIを起動
	homeDir, _ := os.UserHomeDir()

	// Claude CLIを直接起動（scriptコマンドを使用せず、サイズ問題を解決）
	claudeCmd := fmt.Sprintf("CLAUDE_CONFIG_DIR=\"%s\" \"%s\" --dangerously-skip-permissions",
		filepath.Join(homeDir, ".claude"), sl.config.ClaudePath)
	if err := sl.sendKeys(sessionName, claudeCmd); err != nil {
		log.Warn().Err(err).Msg("Failed to send Claude CLI command")
	}

	// Claude CLIの起動を待機
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
	}

	// Claude CLI起動後にサイズ調整を実行（tmuxコマンドで実行）
	sl.optimizeClaudeCLIDisplay(agent)

	// 注意: インストラクションファイルの送信は従来の設定で処理される

	return nil
}

// executeCommand tmuxコマンドを実行
func (sl *SystemLauncher) executeCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	log.Debug().Str("command", cmd).Msg("Executing tmux command")

	execCmd := exec.Command("tmux", parts...) // #nosec G204
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
	agents := []string{"po", "manager", "dev1", "dev2", "dev3", "dev4"}
	for _, agent := range agents {
		sessionName := fmt.Sprintf("%s-%s", sl.config.SessionName, agent)
		if err := sl.tmuxManager.KillSession(sessionName); err != nil {
			log.Warn().Err(err).Msgf("Failed to kill session %s", sessionName)
		}
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
	claudePath := findClaudeExecutableHelper()
	if claudePath == "" {
		return fmt.Errorf("claude CLI not found")
	}
	if !isExecutableHelper(claudePath) {
		return fmt.Errorf("claude CLI is not executable")
	}
	log.Info().Msg("✓ Claude CLI execution test passed")

	// 設定ファイルテスト
	if !checkClaudeConfig() {
		return fmt.Errorf("claude configuration test failed")
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

	instructionsDir := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions")
	files := []string{"po.md", "manager.md", "developer.md"}

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
	tmuxManager := tmux.NewTmuxManager("test")
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

// optimizeClaudeCLIDisplay Claude CLIの表示を最適化（tmuxコマンドで実行）
func (sl *SystemLauncher) optimizeClaudeCLIDisplay(name string) {
	if utils.IsVerboseLogging() {
		utils.DisplayProgress("Claude CLI表示最適化", fmt.Sprintf("%s のClaude CLI表示を最適化中...", name))
	}

	// scriptコマンドを削除したため、特別な最適化は不要
	// Claude CLIが自動的にペインサイズを認識する

	if utils.IsVerboseLogging() {
		utils.DisplaySuccess("Claude CLI表示最適化完了", fmt.Sprintf("%s のClaude CLI表示が最適化されました", name))
	}
}

// ヘルパー関数
func expandPathHelper(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		// ~/foo -> /home/user/foo (正しい展開)
		if len(path) == 1 {
			return homeDir
		}
		if path[1] == '/' {
			return filepath.Join(homeDir, path[2:])
		}
		// ~user形式は未対応
		return path
	}
	return path
}

func findClaudeExecutableHelper() string {
	// 動的npm パスの検出を最初に試す
	if npmPath := detectNpmClaudeCodePathHelper(); npmPath != "" {
		return npmPath
	}

	// Claude CLIの一般的なパスを検索（claude-codeコマンドを優先）
	commonPaths := []string{
		// npm グローバルインストール（claude-code）
		"/usr/local/lib/node_modules/@anthropic-ai/claude-code/cli.js",
		"/opt/homebrew/lib/node_modules/@anthropic-ai/claude-code/cli.js",
		"/usr/local/lib/node_modules/@anthropic/claude-code/bin/claude-code",
		"/opt/homebrew/lib/node_modules/@anthropic/claude-code/bin/claude-code",
		// 従来のclaudeコマンド
		filepath.Join(os.Getenv("HOME"), ".claude", "local", "claude"),
		"/usr/local/bin/claude",
		"/usr/bin/claude",
		"/opt/claude/bin/claude",
	}

	for _, path := range commonPaths {
		if isExecutableHelper(path) {
			return path
		}
	}

	// PATHから検索（claude-codeを優先）
	if claudePath, err := exec.LookPath("claude-code"); err == nil {
		return claudePath
	}

	// PATHから従来のclaudeコマンドを検索
	if claudePath, err := exec.LookPath("claude"); err == nil {
		return claudePath
	}

	return ""
}

// detectNpmClaudeCodePathHelper - npm グローバルインストールパスの動的検出
func detectNpmClaudeCodePathHelper() string {
	// npm root -g でグローバルインストールパスを取得
	cmd := exec.Command("npm", "root", "-g")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	npmRoot := strings.TrimSpace(string(output))
	if npmRoot == "" {
		return ""
	}

	// 複数のパッケージ名を試す
	candidatePaths := []string{
		// @anthropic-ai/claude-code (実際のパッケージ名)
		filepath.Join(npmRoot, "@anthropic-ai", "claude-code", "cli.js"),
		// @anthropic/claude-code (将来の可能性)
		filepath.Join(npmRoot, "@anthropic", "claude-code", "bin", "claude-code"),
		filepath.Join(npmRoot, "@anthropic", "claude-code", "cli.js"),
	}

	// パスの存在確認
	for _, claudeCodePath := range candidatePaths {
		if _, err := os.Stat(claudeCodePath); err == nil {
			return claudeCodePath
		}
	}

	return ""
}

func isExecutableHelper(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	// 実行可能かどうかをテスト（claude-codeとclaudeの両方をサポート）
	if err := exec.Command(path, "--version").Run(); err != nil {
		// --versionが失敗した場合は--helpを試す
		if err := exec.Command(path, "--help").Run(); err != nil {
			return false
		}
	}

	return true
}
