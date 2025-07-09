package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MessageServer - tmuxベースのメッセージサーバー
type MessageServer struct {
	manager     *ClaudeManager
	sessionName string
	tmuxManager *TmuxManager
	mu          sync.RWMutex
	shutdown    chan struct{}
	running     bool
}

// NewMessageServer - tmuxベースのサーバー作成
func NewMessageServer(manager *ClaudeManager, sessionName string) (*MessageServer, error) {
	tmuxManager := NewTmuxManager(sessionName)

	// tmuxセッションの存在確認
	if !tmuxManager.SessionExists(sessionName) {
		return nil, fmt.Errorf("tmux session %s does not exist", sessionName)
	}

	return &MessageServer{
		manager:     manager,
		sessionName: sessionName,
		tmuxManager: tmuxManager,
		shutdown:    make(chan struct{}),
		running:     false,
	}, nil
}

// Start - サーバー開始（tmuxベース）
func (ms *MessageServer) Start() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.running {
		return
	}

	ms.running = true
	log.Info().Str("session", ms.sessionName).Msg("MessageServer started (tmux-based)")

	// バックグラウンドでヘルスチェック実行
	go ms.healthCheckLoop()
}

// healthCheckLoop - ヘルスチェックループ
func (ms *MessageServer) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ms.shutdown:
			return
		case <-ticker.C:
			ms.performHealthCheck()
		}
	}
}

// performHealthCheck - ヘルスチェック実行
func (ms *MessageServer) performHealthCheck() {
	if !ms.tmuxManager.SessionExists(ms.sessionName) {
		log.Error().Str("session", ms.sessionName).Msg("tmux session disappeared")
		return
	}

	paneCount, err := ms.tmuxManager.GetPaneCount(ms.sessionName)
	if err != nil {
		log.Error().Err(err).Str("session", ms.sessionName).Msg("failed to get pane count")
		return
	}

	if paneCount != 6 {
		log.Warn().Int("pane_count", paneCount).Str("session", ms.sessionName).Msg("unexpected pane count")
	}

	log.Debug().Str("session", ms.sessionName).Int("panes", paneCount).Msg("health check passed")
}

// SendMessage - tmuxベースのメッセージ送信
func (ms *MessageServer) SendMessage(agent, message string) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if !ms.running {
		return fmt.Errorf("message server is not running")
	}

	paneIndex, err := ms.getAgentPaneIndex(agent)
	if err != nil {
		return fmt.Errorf("failed to get pane index for agent %s: %v", agent, err)
	}

	if err := ms.tmuxManager.SendKeysWithEnter(ms.sessionName, fmt.Sprintf("%d", paneIndex), message); err != nil {
		return fmt.Errorf("failed to send message to agent %s: %v", agent, err)
	}

	log.Info().Str("agent", agent).Str("message", message).Msg("Message sent to agent")
	return nil
}

// ListAgents - エージェント一覧取得
func (ms *MessageServer) ListAgents() ([]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if !ms.running {
		return nil, fmt.Errorf("message server is not running")
	}

	panes, err := ms.tmuxManager.GetPaneList(ms.sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pane list: %v", err)
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

// GetAgentStatus - エージェント状態取得
func (ms *MessageServer) GetAgentStatus(agent string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if !ms.running {
		return false, fmt.Errorf("message server is not running")
	}

	paneIndex, err := ms.getAgentPaneIndex(agent)
	if err != nil {
		return false, fmt.Errorf("failed to get pane index for agent %s: %v", agent, err)
	}

	// tmuxペインが存在するかチェック
	panes, err := ms.tmuxManager.GetPaneList(ms.sessionName)
	if err != nil {
		return false, fmt.Errorf("failed to get pane list: %v", err)
	}

	for _, pane := range panes {
		if strings.HasPrefix(pane, fmt.Sprintf("%d:", paneIndex)) {
			return true, nil
		}
	}

	return false, nil
}

// getAgentPaneIndex - エージェント名からペインインデックスを取得
func (ms *MessageServer) getAgentPaneIndex(agent string) (int, error) {
	switch strings.ToLower(agent) {
	case "ceo":
		return 0, nil
	case "manager":
		return 1, nil
	case "dev1":
		return 2, nil
	case "dev2":
		return 3, nil
	case "dev3":
		return 4, nil
	case "dev4":
		return 5, nil
	default:
		return -1, fmt.Errorf("unknown agent: %s", agent)
	}
}

// Stop - サーバー停止
func (ms *MessageServer) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.running {
		return nil
	}

	ms.running = false
	close(ms.shutdown)

	log.Info().Str("session", ms.sessionName).Msg("MessageServer stopped")
	return nil
}

// IsRunning - サーバー実行状態確認
func (ms *MessageServer) IsRunning() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.running
}

// GetSessionName - セッション名取得
func (ms *MessageServer) GetSessionName() string {
	return ms.sessionName
}
