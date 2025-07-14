package server

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

// MessageClient - tmuxベースのメッセージクライアント
type MessageClient struct {
	sessionName string
	tmuxManager *tmux.TmuxManagerImpl
}

// NewMessageClient - クライアント作成（自動セッション検出対応）
func NewMessageClient(sessionName string) *MessageClient {
	tmuxManager := tmux.NewTmuxManager(sessionName)

	// セッション名が空の場合は自動検出
	if sessionName == "" {
		expectedPaneCount := 6 // デフォルト値（PO + Manager + 4 Dev）
		if detectedSession, err := tmuxManager.FindDefaultAISession(expectedPaneCount); err == nil {
			sessionName = detectedSession
		} else {
			sessionName = "ai-teams" // フォールバック
		}
	}

	// 指定されたセッションが存在しない場合、自動検出を試行
	if !tmuxManager.SessionExists(sessionName) {
		expectedPaneCount := 6 // デフォルト値（PO + Manager + 4 Dev）
		if detectedSession, err := tmuxManager.FindDefaultAISession(expectedPaneCount); err == nil {
			sessionName = detectedSession
		}
	}

	return &MessageClient{
		sessionName: sessionName,
		tmuxManager: tmuxManager,
	}
}

// SendMessage - tmuxベースのメッセージ送信
func (mc *MessageClient) SendMessage(agent, message string) error {
	paneIndex, err := mc.getAgentPaneIndex(agent)
	if err != nil {
		return fmt.Errorf("failed to get pane index for agent %s: %w", agent, err)
	}

	// tmuxペインにメッセージを送信
	if err := mc.tmuxManager.SendKeysWithEnter(mc.sessionName, strconv.Itoa(paneIndex), message); err != nil {
		return fmt.Errorf("failed to send message to agent %s: %w", agent, err)
	}

	log.Info().Str("agent", agent).Str("message", message).Msg("Message sent to agent")
	return nil
}

// ListAgents - tmuxベースのエージェント一覧取得
func (mc *MessageClient) ListAgents() ([]string, error) {
	panes, err := mc.tmuxManager.GetPaneList(mc.sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pane list: %w", err)
	}

	var agents []string
	for _, pane := range panes {
		parts := strings.Split(pane, ":")
		if len(parts) >= 2 {
			agentName := strings.TrimSpace(parts[1])
			if agentName != "" {
				agents = append(agents, agentName)
			}
		}
	}

	return agents, nil
}

// GetStatus - tmuxベースのエージェント状態取得
func (mc *MessageClient) GetStatus(agent string) (bool, error) {
	paneIndex, err := mc.getAgentPaneIndex(agent)
	if err != nil {
		return false, fmt.Errorf("failed to get pane index for agent %s: %w", agent, err)
	}

	// tmuxペインが存在するかチェック
	cmd := exec.Command("tmux", "list-panes", "-t", mc.sessionName) // #nosec G204
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check pane status: %w", err)
	}

	// ペインが存在するかチェック
	return strings.Contains(string(output), fmt.Sprintf("%d:", paneIndex)), nil
}

// IsServerRunning - tmuxセッションの実行状態確認
func (mc *MessageClient) IsServerRunning() bool {
	return mc.tmuxManager.SessionExists(mc.sessionName)
}

// GetSessionName - セッション名の取得
func (mc *MessageClient) GetSessionName() string {
	return mc.sessionName
}

// CheckConnection - tmuxセッション接続テスト（柔軟な検出）
func (mc *MessageClient) CheckConnection() error {
	// まず自動検出を試行（セッションが存在しない場合、または存在してもAIセッションでない場合）
	if !mc.tmuxManager.SessionExists(mc.sessionName) {
		detectedSession, sessionType, err := mc.tmuxManager.DetectActiveAISession(6)
		if err != nil {
			return fmt.Errorf("no active AI sessions found: %w", err)
		}

		mc.sessionName = detectedSession
		log.Info().Str("session", detectedSession).Str("type", sessionType).Msg("Auto-detected AI session")
	} else {
		// 既存セッションがAIセッションかどうかを判定
		paneCount, err := mc.tmuxManager.GetPaneCount(mc.sessionName)
		if err != nil {
			return fmt.Errorf("failed to get pane count for session %s: %w", mc.sessionName, err)
		}

		// 6ペインまたは1ペインでない場合、他のAIセッションを検出
		if paneCount != 6 && paneCount != 1 {
			detectedSession, sessionType, err := mc.tmuxManager.DetectActiveAISession(6)
			if err != nil {
				return fmt.Errorf("session %s has %d panes (not AI session) and no other AI sessions found: %w", mc.sessionName, paneCount, err)
			}

			mc.sessionName = detectedSession
			log.Info().Str("session", detectedSession).Str("type", sessionType).Msg("Auto-detected AI session (current session not AI)")
		}
	}

	// 最終的なペイン数確認
	paneCount, err := mc.tmuxManager.GetPaneCount(mc.sessionName)
	if err != nil {
		return fmt.Errorf("failed to get pane count for session %s: %w", mc.sessionName, err)
	}

	// 統合監視画面（6ペイン）または個別セッション（1ペイン）を許可
	if paneCount != 6 && paneCount != 1 {
		return fmt.Errorf("expected 6 panes (integrated) or 1 pane (individual) but found %d in session %s", paneCount, mc.sessionName)
	}

	return nil
}

// getAgentPaneIndex - エージェント名からペインインデックスを取得（claude.shと同じ構成）
func (mc *MessageClient) getAgentPaneIndex(agent string) (int, error) {
	switch strings.ToLower(agent) {
	case "po":
		return 1, nil // 左上
	case "manager":
		return 2, nil // 左下
	case "dev1":
		return 3, nil // 右上
	case "dev2":
		return 4, nil // 右上中
	case "dev3":
		return 5, nil // 右下中
	case "dev4":
		return 6, nil // 右下
	default:
		return -1, fmt.Errorf("unknown agent: %s", agent)
	}
}
