package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// 事前コンパイルされた正規表現（パフォーマンス最適化）
var (
	agentSessionRegex = regexp.MustCompile(`-(po|manager|dev\d+)$`)
)

// TmuxManagerImpl tmux操作管理
type TmuxManagerImpl struct {
	sessionName string
	layout      string
}

// NewTmuxManager tmux管理の作成
func NewTmuxManager(sessionName string) *TmuxManagerImpl {
	return &TmuxManagerImpl{
		sessionName: sessionName,
		layout:      "integrated", // "integrated" or "individual"
	}
}

// SessionExists セッションの存在確認
func (tm *TmuxManagerImpl) SessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// ListSessions セッション一覧の取得
func (tm *TmuxManagerImpl) ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var sessions []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			sessions = append(sessions, strings.TrimSpace(line))
		}
	}

	return sessions, nil
}

// CreateSession セッションの作成
func (tm *TmuxManagerImpl) CreateSession(sessionName string) error {
	if tm.SessionExists(sessionName) {
		return fmt.Errorf("session %s already exists", sessionName)
	}

	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session %s: %w", sessionName, err)
	}

	return nil
}

// KillSession セッションの削除
func (tm *TmuxManagerImpl) KillSession(sessionName string) error {
	if !tm.SessionExists(sessionName) {
		return nil // セッションが存在しない場合はエラーとしない
	}

	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session %s: %w", sessionName, err)
	}

	return nil
}

// AttachSession セッションへの接続
func (tm *TmuxManagerImpl) AttachSession(sessionName string) error {
	if !tm.SessionExists(sessionName) {
		return fmt.Errorf("session %s does not exist", sessionName)
	}

	// tmux attach-sessionを実行（非対話的に）
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// 接続エラーの場合はセッション状態を確認
		if tm.SessionExists(sessionName) {
			log.Warn().Str("session", sessionName).Err(err).Msg("Session exists but attach failed")
			return fmt.Errorf("session %s exists but attach failed: %w", sessionName, err)
		}
		return fmt.Errorf("failed to attach to session %s: %w", sessionName, err)
	}

	return nil
}

// CreateIntegratedLayout 統合監視画面レイアウトの作成（動的dev数対応）
func (tm *TmuxManagerImpl) CreateIntegratedLayout(sessionName string, devCount int) error {
	// セッションが存在しない場合は作成
	if !tm.SessionExists(sessionName) {
		if err := tm.CreateSession(sessionName); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
	}

	// ウィンドウ名を設定
	if err := tm.RenameWindow(sessionName, sessionName); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}

	// 動的ペイン構成の作成（PO + Manager + Dev数）
	totalPanes := 2 + devCount

	// 分割のタイミングで少し待機を入れる
	sleep := func() {
		time.Sleep(50 * time.Millisecond)
	}

	// 左側にPO/Manager、右側にDev専用のレイアウトを作成
	// 1. 最初のペインを左右分割（左側 | 右側）
	if err := tm.SplitWindow(sessionName, "-h"); err != nil {
		return fmt.Errorf("failed to split window horizontally: %w", err)
	}
	sleep()

	// 2. 左側（ペイン1）を上下分割（PO | Manager）
	if err := tm.SplitWindow(sessionName+":1.1", "-v"); err != nil {
		return fmt.Errorf("failed to split left pane vertically: %w", err)
	}
	sleep()

	// 3. 右側（ペイン3）を開発者用に分割
	// 最初の開発者はペイン3を使用
	// 2番目以降の開発者のために、常に最初の右側ペイン（ペイン3）を分割
	for i := 2; i <= devCount; i++ {
		// 常にペイン3を分割して等間隔にする
		target := fmt.Sprintf("%s:1.3", sessionName)
		if err := tm.SplitWindow(target, "-v"); err != nil {
			return fmt.Errorf("failed to split dev pane %d: %w", i, err)
		}
		sleep()
	}

	// ペインサイズの調整
	if err := tm.AdjustPaneSizes(sessionName, devCount); err != nil {
		return fmt.Errorf("failed to adjust pane sizes: %w", err)
	}

	// ペインタイトルの設定
	if err := tm.SetPaneTitles(sessionName, devCount); err != nil {
		return fmt.Errorf("failed to set pane titles: %w", err)
	}

	log.Info().Str("session", sessionName).Int("dev_count", devCount).Int("total_panes", totalPanes).Msg("Dynamic integrated layout created successfully")
	return nil
}

// SetupClaudeInPanes 各ペインでClaude CLI自動起動とインストラクション送信
func (tm *TmuxManagerImpl) SetupClaudeInPanes(sessionName string, claudeCLIPath string, instructionsDir string, devCount int) error {

	// 動的ペイン設定マップ（ペイン番号 → エージェント名）
	paneAgentMap := make(map[string]string)
	paneAgentMap["1"] = "po"
	paneAgentMap["2"] = "manager"
	for i := 1; i <= devCount; i++ {
		paneAgentMap[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("dev%d", i)
	}

	for pane, agent := range paneAgentMap {

		// 各ペインでClaude CLIを起動
		if err := tm.startClaudeInPane(sessionName, pane, agent, claudeCLIPath); err != nil {
			log.Error().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to start Claude CLI in pane")
			return fmt.Errorf("failed to start Claude CLI in pane %s (%s): %w", pane, agent, err)
		}

		// ペイン間の起動間隔を5秒に設定
		time.Sleep(5 * time.Second)
	}

	// Claude CLI起動完了を待機
	time.Sleep(2 * time.Second)

	// インストラクションファイルを送信
	for pane, agent := range paneAgentMap {
		if err := tm.sendInstructionToPane(sessionName, pane, agent, instructionsDir); err != nil {
			log.Warn().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to send instruction to pane (non-critical)")
			// インストラクション送信の失敗は警告レベル（継続可能）
		}

		// ペイン間でインストラクション送信間隔を2秒に設定
		time.Sleep(2 * time.Second)
	}

	return nil
}

// SetupClaudeInPanesWithConfig 設定ファイルを使用した各ペインでClaude CLI自動起動とインストラクション送信
func (tm *TmuxManagerImpl) SetupClaudeInPanesWithConfig(sessionName string, claudeCLIPath string, instructionsDir string, config interface{}, devCount int) error {
	// TeamConfigインターフェースを使用してrole別instructionsファイルを取得
	type InstructionConfig interface {
		GetPOInstructionFile() string
		GetManagerInstructionFile() string
		GetDevInstructionFile() string
	}

	// configがInstructionConfigを実装しているかチェック
	var instructionConfig InstructionConfig
	if ic, ok := config.(InstructionConfig); ok {
		instructionConfig = ic
	}

	// 動的ペイン設定マップ（ペイン番号 → エージェント名）
	paneAgentMap := make(map[string]string)
	paneAgentMap["1"] = "po"
	paneAgentMap["2"] = "manager"
	for i := 1; i <= devCount; i++ {
		paneAgentMap[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("dev%d", i)
	}

	for pane, agent := range paneAgentMap {
		// 各ペインでClaude CLIを起動
		if err := tm.startClaudeInPane(sessionName, pane, agent, claudeCLIPath); err != nil {
			log.Error().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to start Claude CLI in pane")
			return fmt.Errorf("failed to start Claude CLI in pane %s (%s): %w", pane, agent, err)
		}

		// ペイン間の起動間隔を5秒に設定
		time.Sleep(5 * time.Second)
	}

	// Claude CLI起動完了を待機
	time.Sleep(2 * time.Second)

	// インストラクションファイルを送信
	for pane, agent := range paneAgentMap {
		if err := tm.SendInstructionToPaneWithConfig(sessionName, pane, agent, instructionsDir, instructionConfig); err != nil {
			log.Warn().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to send instruction to pane (non-critical)")
			// インストラクション送信の失敗は警告レベル（継続可能）
		}

		// ペイン間でインストラクション送信間隔を2秒に設定
		time.Sleep(2 * time.Second)
	}

	return nil
}

// startClaudeInPane 指定ペインでClaude CLIを起動
func (tm *TmuxManagerImpl) startClaudeInPane(sessionName, pane, _ /* agent */, claudeCLIPath string) error {
	// ペインが存在するか確認
	if err := tm.WaitForPaneReady(sessionName, pane, 5*time.Second); err != nil {
		return fmt.Errorf("pane %s not ready: %w", pane, err)
	}

	// Claude CLI起動コマンドを作成
	claudeCommand := fmt.Sprintf("%s --dangerously-skip-permissions", claudeCLIPath)

	// ペインにClaude CLI起動コマンドを送信
	if err := tm.SendKeysWithEnter(sessionName, pane, claudeCommand); err != nil {
		return fmt.Errorf("failed to send Claude CLI command to pane: %w", err)
	}

	return nil
}

// sendInstructionToPane 指定ペインにインストラクションファイルを送信（強化版）
func (tm *TmuxManagerImpl) sendInstructionToPane(sessionName, pane, agent, instructionsDir string) error {
	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("📤 インストラクション送信開始")

	// インストラクションファイルのパスを決定
	var instructionFile string
	switch agent {
	case "po":
		instructionFile = filepath.Join(instructionsDir, "po.md")
	case "manager":
		instructionFile = filepath.Join(instructionsDir, "manager.md")
	case "dev1", "dev2", "dev3", "dev4":
		instructionFile = filepath.Join(instructionsDir, "developer.md")
	default:
		log.Error().Str("agent", agent).Msg("❌ 未知のエージェントタイプ")
		return fmt.Errorf("unknown agent type: %s", agent)
	}

	log.Info().Str("instruction_file", instructionFile).Msg("📁 インストラクションファイルパス決定")

	// インストラクションファイルの存在確認（強化版）
	fileInfo, err := os.Stat(instructionFile)
	if os.IsNotExist(err) {
		log.Warn().Str("instruction_file", instructionFile).Msg("⚠️ インストラクションファイルが存在しません（スキップ）")
		return nil // ファイルが存在しない場合はスキップ（エラーではない）
	}
	if err != nil {
		log.Error().Str("instruction_file", instructionFile).Err(err).Msg("❌ ファイル情報取得エラー")
		return fmt.Errorf("failed to stat instruction file: %w", err)
	}
	if fileInfo.Size() == 0 {
		log.Warn().Str("instruction_file", instructionFile).Msg("⚠️ インストラクションファイルが空です（スキップ）")
		return nil
	}

	log.Info().Str("instruction_file", instructionFile).Int64("file_size", fileInfo.Size()).Msg("✅ ファイル存在確認完了")

	// Claude CLI準備完了待機（強化版）
	if err := tm.waitForClaudeReady(sessionName, pane, 10*time.Second); err != nil {
		log.Warn().Str("session", sessionName).Str("pane", pane).Err(err).Msg("⚠️ Claude CLI準備待機タイムアウト（続行）")
	}

	// catコマンドでインストラクションファイルを送信（リトライ機能付き）
	catCommand := fmt.Sprintf("cat \"%s\"", instructionFile)

	for attempt := 1; attempt <= 3; attempt++ {
		log.Info().Str("command", catCommand).Int("attempt", attempt).Msg("📋 catコマンド送信中")

		if err := tm.SendKeysWithEnter(sessionName, pane, catCommand); err != nil {
			log.Warn().Err(err).Int("attempt", attempt).Msg("⚠️ catコマンド送信失敗")
			if attempt == 3 {
				return fmt.Errorf("failed to send instruction file after 3 attempts: %w", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// catコマンド実行完了を待機
		time.Sleep(2 * time.Second)
		log.Info().Int("attempt", attempt).Msg("✅ catコマンド送信成功")
		break
	}

	// Claude CLI実行のためのEnter送信（最適化版）
	time.Sleep(1 * time.Second)
	log.Info().Msg("🔄 Claude CLI実行のためのEnter送信")

	for i := 0; i < 3; i++ {
		if err := tm.SendKeysToPane(sessionName, pane, "C-m"); err != nil {
			log.Warn().Err(err).Int("attempt", i+1).Msg("⚠️ Enter送信エラー")
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("✅ インストラクション送信完了")
	return nil
}

// CreateIndividualLayout 個別セッション方式の作成
func (tm *TmuxManagerImpl) CreateIndividualLayout(sessionName string, devCount int) error {
	agents := []string{"po", "manager"}
	for i := 1; i <= devCount; i++ {
		agents = append(agents, fmt.Sprintf("dev%d", i))
	}

	for _, agent := range agents {
		agentSession := fmt.Sprintf("%s-%s", sessionName, agent)

		if err := tm.CreateSession(agentSession); err != nil {
			return fmt.Errorf("failed to create session for %s: %w", agent, err)
		}

		// ウィンドウ名を設定
		if err := tm.RenameWindow(agentSession, agentSession); err != nil {
			return fmt.Errorf("failed to rename window for %s: %w", agent, err)
		}
	}

	log.Info().Str("session", sessionName).Msg("Individual layout created successfully")
	return nil
}

// SplitWindow ウィンドウの分割
func (tm *TmuxManagerImpl) SplitWindow(target, direction string) error {
	cmd := exec.Command("tmux", "split-window", direction, "-t", target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux command failed: split-window %s -t %s (output: %s)", direction, target, string(output))
	}
	return nil
}

// RenameWindow ウィンドウ名の変更
func (tm *TmuxManagerImpl) RenameWindow(sessionName, windowName string) error {
	cmd := exec.Command("tmux", "rename-window", "-t", sessionName, windowName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}
	return nil
}

// AdjustPaneSizes ペインサイズの調整（動的dev数対応・等間隔実装）
func (tm *TmuxManagerImpl) AdjustPaneSizes(sessionName string, devCount int) error {
	totalPanes := 2 + devCount // PO + Manager + Dev数

	// devCount=0の場合のゼロ除算保護
	if devCount <= 0 {
		log.Warn().Str("session", sessionName).Int("dev_count", devCount).Msg("dev数が0以下のため、ペインサイズ調整をスキップ")
		return fmt.Errorf("devCount must be greater than 0, got: %d", devCount)
	}

	// 左側（PO/Manager）を全体の50%、右側（Dev）を50%に設定
	leftSidePercentage := 50

	log.Info().Str("session", sessionName).Int("dev_count", devCount).Int("total_panes", totalPanes).Msg("等間隔ペイン分割開始")

	// ウィンドウサイズを取得
	windowWidth, windowHeight, err := tm.getWindowSize(sessionName)
	if err != nil {
		log.Warn().Err(err).Msg("ウィンドウサイズ取得失敗、デフォルト値を使用")
		windowWidth = 120
		windowHeight = 40
	}

	// ウィンドウサイズの妥当性チェック
	if windowHeight <= 0 {
		log.Warn().Int("window_height", windowHeight).Msg("無効なウィンドウ高さ、デフォルト値を使用")
		windowHeight = 40
	}

	// 左側の幅を計算（全体の50%）
	leftWidth := (windowWidth * leftSidePercentage) / 100

	// 1. 左右分割の調整（左側50%, 右側50%）
	time.Sleep(100 * time.Millisecond)
	leftCmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.1", sessionName), "-x", fmt.Sprintf("%d", leftWidth)) // #nosec G204
	if err := leftCmd.Run(); err != nil {
		log.Warn().Str("pane", "1").Int("width", leftWidth).Err(err).Msg("左側ペイン調整失敗")
	}

	// 2. 左側の上下分割調整（PO/Manager 50%ずつ）
	time.Sleep(100 * time.Millisecond)
	poHeight := windowHeight / 2
	poCmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.1", sessionName), "-y", fmt.Sprintf("%d", poHeight)) // #nosec G204
	if err := poCmd.Run(); err != nil {
		log.Warn().Str("pane", "PO").Int("height", poHeight).Err(err).Msg("PO/Manager分割調整失敗")
	}

	// 3. 右側の開発者ペインを等間隔で調整（ゼロ除算保護済み）
	// devCountは既に0以下でないことが確認済み
	devPaneHeight := windowHeight / devCount

	// 各開発者ペインの高さを設定
	for i := 1; i <= devCount; i++ {
		paneNumber := i + 2 // PO(1), Manager(2)の後は3から

		// 各ペインの高さを均等に設定
		time.Sleep(100 * time.Millisecond)
		cmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.%d", sessionName, paneNumber), "-y", fmt.Sprintf("%d", devPaneHeight)) // #nosec G204
		if err := cmd.Run(); err != nil {
			log.Warn().Str("pane", fmt.Sprintf("%d", paneNumber)).Int("height", devPaneHeight).Err(err).Msg("ペイン等間隔リサイズ失敗")
		} else {
			log.Debug().Str("pane", fmt.Sprintf("%d", paneNumber)).Int("height", devPaneHeight).Msg("ペイン等間隔リサイズ成功")
		}
	}

	// 4. 最後に左右の幅を再調整（50%ずつを維持）
	time.Sleep(100 * time.Millisecond)
	finalLeftCmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.1", sessionName), "-x", fmt.Sprintf("%d", leftWidth)) // #nosec G204
	if err := finalLeftCmd.Run(); err != nil {
		log.Warn().Err(err).Msg("最終的な左右幅調整失敗")
	}

	log.Info().Str("session", sessionName).Int("dev_count", devCount).Msg("等間隔ペイン分割完了")
	return nil
}

// SetPaneTitles ペインタイトルの設定（動的dev数対応）
func (tm *TmuxManagerImpl) SetPaneTitles(sessionName string, devCount int) error {
	// ペインタイトルを表示するように設定
	cmd := exec.Command("tmux", "set-option", "-t", sessionName, "pane-border-status", "top")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set pane border status: %w", err)
	}

	cmd = exec.Command("tmux", "set-option", "-t", sessionName, "pane-border-format", "#T")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set pane border format: %w", err)
	}

	// 自動リネームを無効化
	cmd = exec.Command("tmux", "set-window-option", "-t", sessionName, "automatic-rename", "off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable automatic rename: %w", err)
	}

	cmd = exec.Command("tmux", "set-window-option", "-t", sessionName, "allow-rename", "off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable allow rename: %w", err)
	}

	// 各ペインのタイトル設定（動的dev数対応）
	titles := make(map[string]string)
	titles["1"] = "PO"      // 左上
	titles["2"] = "Manager" // 左下

	// 動的に開発者タイトルを設定
	for i := 1; i <= devCount; i++ {
		paneNumber := fmt.Sprintf("%d", i+2)
		titles[paneNumber] = fmt.Sprintf("Dev%d", i)
	}

	for pane, title := range titles {
		target := fmt.Sprintf("%s:1.%s", sessionName, pane)
		cmd = exec.Command("tmux", "select-pane", "-t", target, "-T", title) // #nosec G204
		if err := cmd.Run(); err != nil {
			log.Warn().Str("pane", target).Str("title", title).Err(err).Msg("Failed to set pane title")
		}
	}

	return nil
}

// GetPaneCount ペイン数の取得
func (tm *TmuxManagerImpl) GetPaneCount(sessionName string) (int, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName)
	output, err := cmd.Output()
	if err != nil {
		log.Debug().Str("session", sessionName).Err(err).Msg("Failed to get pane count")
		return 0, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	paneCount := len(lines)

	log.Debug().Str("session", sessionName).Int("pane_count", paneCount).Msg("GetPaneCount result")
	return paneCount, nil
}

// GetPaneList ペイン一覧の取得
func (tm *TmuxManagerImpl) GetPaneList(sessionName string) ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_index}:#{pane_title}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var panes []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			panes = append(panes, strings.TrimSpace(line))
		}
	}

	return panes, nil
}

// SendKeysToPane ペインにキーを送信
func (tm *TmuxManagerImpl) SendKeysToPane(sessionName, pane, keys string) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys) // #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys to pane %s: %w", target, err)
	}
	return nil
}

// SendKeysWithEnter ペインにキーを送信（Enter付き）
func (tm *TmuxManagerImpl) SendKeysWithEnter(sessionName, pane, keys string) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "C-m") // #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys with enter to pane %s: %w", target, err)
	}
	return nil
}

// GetAITeamSessions AIチーム関連セッションの取得
func (tm *TmuxManagerImpl) GetAITeamSessions(expectedPaneCount int) (map[string][]string, error) {
	sessions, err := tm.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	result := map[string][]string{
		"integrated": {},
		"individual": {},
		"other":      {},
	}

	for _, session := range sessions {
		// 統合監視画面の判定（動的ペイン数構成）
		paneCount, err := tm.GetPaneCount(session)
		log.Debug().Str("session", session).Int("pane_count", paneCount).Int("expected_pane_count", expectedPaneCount).Err(err).Msg("Session analysis")

		switch {
		case err == nil && paneCount == expectedPaneCount:
			result["integrated"] = append(result["integrated"], session)
			log.Debug().Str("session", session).Msg("Added as integrated session")
		case agentSessionRegex.MatchString(session):
			// 個別セッション方式の判定
			baseName := agentSessionRegex.ReplaceAllString(session, "")
			if !containsString(result["individual"], baseName) {
				result["individual"] = append(result["individual"], baseName)
				log.Debug().Str("session", session).Str("base_name", baseName).Msg("Added as individual session")
			}
		default:
			// 数字だけのセッション（「1」等）や既存のAIセッションの可能性があるかチェック
			if err == nil && paneCount >= 1 {
				// 数字だけのセッション名や短い名前のセッションは潜在的なAIセッション
				if len(session) <= 3 || strings.Contains(session, "ai") || strings.Contains(session, "claude") {
					result["integrated"] = append(result["integrated"], session)
					log.Debug().Str("session", session).Msg("Added as potential AI session")
				} else {
					result["other"] = append(result["other"], session)
					log.Debug().Str("session", session).Msg("Added as other session")
				}
			} else {
				result["other"] = append(result["other"], session)
				log.Debug().Str("session", session).Msg("Added as other session")
			}
		}
	}

	return result, nil
}

// FindDefaultAISession デフォルトAIセッションの検出
func (tm *TmuxManagerImpl) FindDefaultAISession(expectedPaneCount int) (string, error) {
	aiSessions, err := tm.GetAITeamSessions(expectedPaneCount)
	if err != nil {
		return "", fmt.Errorf("failed to get AI team sessions: %w", err)
	}

	// 統合監視画面セッションを優先
	if len(aiSessions["integrated"]) > 0 {
		return aiSessions["integrated"][0], nil
	}

	// 個別セッション方式の場合
	if len(aiSessions["individual"]) > 0 {
		return aiSessions["individual"][0], nil
	}

	// AIセッションが見つからない場合もセッションを探す
	sessions, err := tm.ListSessions()
	if err != nil {
		return "ai-teams", err
	}

	// 潜在的なAIセッションを検出（数字だけのセッション名や短い名前）
	for _, session := range sessions {
		paneCount, err := tm.GetPaneCount(session)
		if err != nil {
			continue
		}
		// 数字だけのセッション名や短い名前、AI関連キーワードのセッションをチェック
		if paneCount >= 1 && (len(session) <= 3 ||
			strings.Contains(session, "ai") ||
			strings.Contains(session, "claude") ||
			strings.Contains(session, "agent")) {
			return session, nil
		}
	}

	// 最終的にデフォルト値を返す
	return "ai-teams", nil
}

// DetectActiveAISession アクティブなAIセッションの検出
func (tm *TmuxManagerImpl) DetectActiveAISession(expectedPaneCount int) (string, string, error) {
	aiSessions, err := tm.GetAITeamSessions(expectedPaneCount)
	if err != nil {
		return "", "", fmt.Errorf("failed to get AI team sessions: %w", err)
	}

	// 統合監視画面セッションを優先
	if len(aiSessions["integrated"]) > 0 {
		return aiSessions["integrated"][0], "integrated", nil
	}

	// 個別セッション方式の場合
	if len(aiSessions["individual"]) > 0 {
		return aiSessions["individual"][0], "individual", nil
	}

	// AIセッションが見つからない場合
	return "", "", fmt.Errorf("no active AI sessions found")
}

// DeleteAITeamSessions AIチーム関連セッションの削除
func (tm *TmuxManagerImpl) DeleteAITeamSessions(sessionName string, devCount int) error {
	log.Info().Str("session", sessionName).Msg("Deleting AI team sessions")

	deletedCount := 0

	// 統合監視画面の場合
	expectedPaneCount := 2 + devCount
	if tm.SessionExists(sessionName) {
		paneCount, err := tm.GetPaneCount(sessionName)
		switch {
		case err == nil && paneCount == expectedPaneCount:
			log.Info().Str("session", sessionName).Int("pane_count", paneCount).Msg("Deleting integrated session")
			if err := tm.KillSession(sessionName); err != nil {
				return fmt.Errorf("failed to delete integrated session: %w", err)
			}
			deletedCount++
		default:
			log.Info().Str("session", sessionName).Int("pane_count", paneCount).Msg("Deleting general session")
			if err := tm.KillSession(sessionName); err != nil {
				return fmt.Errorf("failed to delete general session: %w", err)
			}
			deletedCount++
		}
	}

	// 個別セッション方式の場合
	agents := []string{"po", "manager"}
	for i := 1; i <= devCount; i++ {
		agents = append(agents, fmt.Sprintf("dev%d", i))
	}
	for _, agent := range agents {
		agentSession := fmt.Sprintf("%s-%s", sessionName, agent)
		if tm.SessionExists(agentSession) {
			log.Info().Str("session", agentSession).Msg("Deleting individual session")
			if err := tm.KillSession(agentSession); err != nil {
				return fmt.Errorf("failed to delete individual session %s: %w", agentSession, err)
			}
			deletedCount++
		}
	}

	if deletedCount == 0 {
		return fmt.Errorf("no sessions found for %s", sessionName)
	}

	log.Info().Str("session", sessionName).Int("deleted_count", deletedCount).Msg("AI team sessions deleted")
	return nil
}

// containsString スライス内の文字列の存在確認
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// WaitForPaneReady ペインの準備完了待機
func (tm *TmuxManagerImpl) WaitForPaneReady(sessionName, pane string, timeout time.Duration) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	start := time.Now()

	for time.Since(start) < timeout {
		// ペインの存在確認
		cmd := exec.Command("tmux", "list-panes", "-t", sessionName)
		output, err := cmd.Output()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// ペインが存在するかチェック
		if strings.Contains(string(output), pane) {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for pane %s to be ready", target)
}

// waitForClaudeReady Claude CLI準備完了待機（新規実装）
func (tm *TmuxManagerImpl) waitForClaudeReady(sessionName, pane string, timeout time.Duration) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	start := time.Now()

	log.Info().Str("target", target).Dur("timeout", timeout).Msg("🔄 Claude CLI準備完了待機開始")

	for time.Since(start) < timeout {
		// ペインの内容を取得してClaude CLIが準備完了かチェック
		cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p") // #nosec G204
		output, err := cmd.Output()
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		paneContent := string(output)

		// Claude CLIが起動完了した場合の典型的な出力パターンをチェック
		if strings.Contains(paneContent, "claude") ||
			strings.Contains(paneContent, ">") ||
			strings.Contains(paneContent, "$") ||
			len(strings.TrimSpace(paneContent)) > 10 {
			log.Info().Str("target", target).Msg("✅ Claude CLI準備完了検知")
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Warn().Str("target", target).Dur("elapsed", time.Since(start)).Msg("⚠️ Claude CLI準備完了待機タイムアウト")
	return fmt.Errorf("timeout waiting for Claude CLI to be ready in pane %s", target)
}

// GetSessionInfo セッション情報の取得
func (tm *TmuxManagerImpl) GetSessionInfo(sessionName string, expectedPaneCount int) (map[string]interface{}, error) {
	if !tm.SessionExists(sessionName) {
		return nil, fmt.Errorf("session %s does not exist", sessionName)
	}

	info := map[string]interface{}{
		"name":   sessionName,
		"exists": true,
	}

	// ペイン数の取得
	paneCount, err := tm.GetPaneCount(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pane count: %w", err)
	}
	info["pane_count"] = paneCount

	// ペイン一覧の取得
	panes, err := tm.GetPaneList(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pane list: %w", err)
	}
	info["panes"] = panes

	// セッションタイプの判定
	if paneCount == expectedPaneCount {
		info["type"] = "integrated"
	} else {
		info["type"] = "general"
	}

	return info, nil
}

// getWindowSize ウィンドウのサイズを取得
func (tm *TmuxManagerImpl) getWindowSize(sessionName string) (int, int, error) {
	// 幅を取得
	widthCmd := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{window_width}")
	widthOutput, err := widthCmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get window width: %w", err)
	}

	width, err := strconv.Atoi(strings.TrimSpace(string(widthOutput)))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse window width: %w", err)
	}

	// 高さを取得
	heightCmd := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{window_height}")
	heightOutput, err := heightCmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get window height: %w", err)
	}

	height, err := strconv.Atoi(strings.TrimSpace(string(heightOutput)))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse window height: %w", err)
	}

	log.Debug().Str("session", sessionName).Int("width", width).Int("height", height).Msg("Window size retrieved")
	return width, height, nil
}
