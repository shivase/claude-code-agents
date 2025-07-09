package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// TmuxManager - tmux操作管理
type TmuxManager struct {
	sessionName string
	layout      string
}

// NewTmuxManager - tmux管理の作成
func NewTmuxManager(sessionName string) *TmuxManager {
	return &TmuxManager{
		sessionName: sessionName,
		layout:      "integrated", // "integrated" or "individual"
	}
}

// SessionExists - セッションの存在確認
func (tm *TmuxManager) SessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// ListSessions - セッション一覧の取得
func (tm *TmuxManager) ListSessions() ([]string, error) {
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

// CreateSession - セッションの作成
func (tm *TmuxManager) CreateSession(sessionName string) error {
	if tm.SessionExists(sessionName) {
		return fmt.Errorf("session %s already exists", sessionName)
	}

	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session %s: %w", sessionName, err)
	}

	log.Info().Str("session", sessionName).Msg("tmux session created")
	return nil
}

// KillSession - セッションの削除
func (tm *TmuxManager) KillSession(sessionName string) error {
	if !tm.SessionExists(sessionName) {
		return nil // セッションが存在しない場合はエラーとしない
	}

	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session %s: %w", sessionName, err)
	}

	log.Info().Str("session", sessionName).Msg("tmux session killed")
	return nil
}

// AttachSession - セッションへの接続
func (tm *TmuxManager) AttachSession(sessionName string) error {
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

// CreateIntegratedLayout - 統合監視画面レイアウトの作成（claude.shと同じ構成）
func (tm *TmuxManager) CreateIntegratedLayout(sessionName string) error {
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

	// 6ペイン構成の作成（claude.shと同じ構成）
	log.Info().Str("session", sessionName).Msg("Creating integrated layout with 6 panes (claude.sh compatible)")

	// 分割のタイミングで少し待機を入れる
	sleep := func() {
		time.Sleep(50 * time.Millisecond)
	}

	// 1. まず左右分割（左側、右側）
	if err := tm.SplitWindow(sessionName, "-h"); err != nil {
		return fmt.Errorf("failed to split window horizontally: %w", err)
	}
	sleep()
	log.Info().Str("session", sessionName).Msg("✓ 左右分割完了")

	// 2. 左側を上下分割（上: CEO、下: Manager）
	if err := tm.SplitWindow(sessionName+":1.1", "-v"); err != nil {
		return fmt.Errorf("failed to split left pane vertically: %w", err)
	}
	sleep()
	log.Info().Str("session", sessionName).Msg("✓ 左側を上下分割完了（上: CEO、下: Manager）")

	// 3. 右側を上下分割（上: Dev1、下: 残り）
	if err := tm.SplitWindow(sessionName+":1.3", "-v"); err != nil {
		return fmt.Errorf("failed to split right pane vertically: %w", err)
	}
	sleep()
	log.Info().Str("session", sessionName).Msg("✓ 右側を上下分割完了（上: Dev1、下: 残り）")

	// 4. 右下をさらに分割（Dev2用）
	if err := tm.SplitWindow(sessionName+":1.4", "-v"); err != nil {
		return fmt.Errorf("failed to split right bottom pane: %w", err)
	}
	sleep()
	log.Info().Str("session", sessionName).Msg("✓ 右下を分割完了（Dev2用）")

	// 5. 最後のペインをさらに分割（Dev3用）
	if err := tm.SplitWindow(sessionName+":1.5", "-v"); err != nil {
		return fmt.Errorf("failed to split for dev3: %w", err)
	}
	sleep()
	log.Info().Str("session", sessionName).Msg("✓ 最後のペインを分割完了（Dev3用）")

	// ペインサイズの調整
	if err := tm.AdjustPaneSizes(sessionName); err != nil {
		return fmt.Errorf("failed to adjust pane sizes: %w", err)
	}

	// ペインタイトルの設定
	if err := tm.SetPaneTitles(sessionName); err != nil {
		return fmt.Errorf("failed to set pane titles: %w", err)
	}

	log.Info().Str("session", sessionName).Msg("Integrated layout created successfully (claude.sh compatible)")
	return nil
}

// CreateIndividualLayout - 個別セッション方式の作成
func (tm *TmuxManager) CreateIndividualLayout(sessionName string) error {
	agents := []string{"ceo", "manager", "dev1", "dev2", "dev3", "dev4"}

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

// SplitWindow - ウィンドウの分割
func (tm *TmuxManager) SplitWindow(target, direction string) error {
	cmd := exec.Command("tmux", "split-window", direction, "-t", target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux command failed: split-window %s -t %s (output: %s)", direction, target, string(output))
	}
	return nil
}

// RenameWindow - ウィンドウ名の変更
func (tm *TmuxManager) RenameWindow(sessionName, windowName string) error {
	cmd := exec.Command("tmux", "rename-window", "-t", sessionName, windowName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}
	return nil
}

// AdjustPaneSizes - ペインサイズの調整（claude.shと同じ構成）
func (tm *TmuxManager) AdjustPaneSizes(sessionName string) error {
	// 右側のDev1-Dev4のペインを等間隔に調整（claude.shと同じ構成）
	panes := []string{"3", "4", "5", "6"} // Dev1, Dev2, Dev3, Dev4

	for _, pane := range panes {
		target := fmt.Sprintf("%s:1.%s", sessionName, pane)
		cmd := exec.Command("tmux", "resize-pane", "-t", target, "-p", "25")
		if err := cmd.Run(); err != nil {
			log.Warn().Str("pane", target).Err(err).Msg("Failed to resize pane")
		}
	}

	return nil
}

// SetPaneTitles - ペインタイトルの設定（claude.shと同じ構成）
func (tm *TmuxManager) SetPaneTitles(sessionName string) error {
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

	// 各ペインのタイトル設定（claude.shと同じ構成）
	titles := map[string]string{
		"1": "CEO",     // 左上
		"2": "Manager", // 左下
		"3": "Dev1",    // 右上
		"4": "Dev2",    // 右上中
		"5": "Dev3",    // 右下中
		"6": "Dev4",    // 右下
	}

	for pane, title := range titles {
		target := fmt.Sprintf("%s:1.%s", sessionName, pane)
		cmd = exec.Command("tmux", "select-pane", "-t", target, "-T", title)
		if err := cmd.Run(); err != nil {
			log.Warn().Str("pane", target).Str("title", title).Err(err).Msg("Failed to set pane title")
		}
	}

	return nil
}

// GetPaneCount - ペイン数の取得
func (tm *TmuxManager) GetPaneCount(sessionName string) (int, error) {
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

// GetPaneList - ペイン一覧の取得
func (tm *TmuxManager) GetPaneList(sessionName string) ([]string, error) {
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

// SendKeysToPane - ペインにキーを送信
func (tm *TmuxManager) SendKeysToPane(sessionName, pane, keys string) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys to pane %s: %w", target, err)
	}
	return nil
}

// SendKeysWithEnter - ペインにキーを送信（Enter付き）
func (tm *TmuxManager) SendKeysWithEnter(sessionName, pane, keys string) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "C-m")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys with enter to pane %s: %w", target, err)
	}
	return nil
}

// GetAITeamSessions - AIチーム関連セッションの取得
func (tm *TmuxManager) GetAITeamSessions() (map[string][]string, error) {
	sessions, err := tm.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	result := map[string][]string{
		"integrated": []string{},
		"individual": []string{},
		"other":      []string{},
	}

	for _, session := range sessions {
		// 統合監視画面の判定（6ペイン構成）
		paneCount, err := tm.GetPaneCount(session)
		log.Debug().Str("session", session).Int("pane_count", paneCount).Err(err).Msg("Session analysis")
		
		if err == nil && paneCount == 6 {
			result["integrated"] = append(result["integrated"], session)
			log.Debug().Str("session", session).Msg("Added as integrated session")
		} else if matched, _ := regexp.MatchString(`-(ceo|manager|dev[1-4])$`, session); matched {
			// 個別セッション方式の判定
			baseName := regexp.MustCompile(`-(ceo|manager|dev[1-4])$`).ReplaceAllString(session, "")
			if !containsString(result["individual"], baseName) {
				result["individual"] = append(result["individual"], baseName)
				log.Debug().Str("session", session).Str("base_name", baseName).Msg("Added as individual session")
			}
		} else {
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

// FindDefaultAISession - デフォルトAIセッションの検出
func (tm *TmuxManager) FindDefaultAISession() (string, error) {
	aiSessions, err := tm.GetAITeamSessions()
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
		return "ai-teams", nil
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

// DetectActiveAISession - アクティブなAIセッションの検出
func (tm *TmuxManager) DetectActiveAISession() (string, string, error) {
	aiSessions, err := tm.GetAITeamSessions()
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

// DeleteAITeamSessions - AIチーム関連セッションの削除
func (tm *TmuxManager) DeleteAITeamSessions(sessionName string) error {
	log.Info().Str("session", sessionName).Msg("Deleting AI team sessions")

	deletedCount := 0

	// 統合監視画面の場合
	if tm.SessionExists(sessionName) {
		paneCount, err := tm.GetPaneCount(sessionName)
		if err == nil && paneCount == 6 {
			log.Info().Str("session", sessionName).Msg("Deleting integrated session")
			if err := tm.KillSession(sessionName); err != nil {
				return fmt.Errorf("failed to delete integrated session: %w", err)
			}
			deletedCount++
		} else {
			log.Info().Str("session", sessionName).Msg("Deleting general session")
			if err := tm.KillSession(sessionName); err != nil {
				return fmt.Errorf("failed to delete general session: %w", err)
			}
			deletedCount++
		}
	}

	// 個別セッション方式の場合
	agents := []string{"ceo", "manager", "dev1", "dev2", "dev3", "dev4"}
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

// containsString - スライス内の文字列の存在確認
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// WaitForPaneReady - ペインの準備完了待機
func (tm *TmuxManager) WaitForPaneReady(sessionName, pane string, timeout time.Duration) error {
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

// GetSessionInfo - セッション情報の取得
func (tm *TmuxManager) GetSessionInfo(sessionName string) (map[string]interface{}, error) {
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
	if paneCount == 6 {
		info["type"] = "integrated"
	} else {
		info["type"] = "general"
	}

	return info, nil
}
