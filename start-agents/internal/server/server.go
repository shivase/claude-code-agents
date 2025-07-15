package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

// MessageServer - tmux-based message server
type MessageServer struct {
	sessionName string
	tmuxManager *tmux.TmuxManagerImpl
	shutdown    chan struct{}
	running     bool
}

// Start - Start server (tmux-based)
func (ms *MessageServer) Start() {
	if ms.running {
		return
	}

	ms.running = true
	log.Info().Str("session", ms.sessionName).Msg("MessageServer started (tmux-based)")

	// Execute health check in background
	go ms.healthCheckLoop()
}

// healthCheckLoop - Health check loop
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

// performHealthCheck - Perform health check
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

// SendMessage - Send message via tmux
func (ms *MessageServer) SendMessage(agent, message string) error {
	if !ms.running {
		return fmt.Errorf("message server is not running")
	}

	paneIndex, err := ms.getAgentPaneIndex(agent)
	if err != nil {
		return fmt.Errorf("failed to get pane index for agent %s: %w", agent, err)
	}

	if err := ms.tmuxManager.SendKeysWithEnter(ms.sessionName, fmt.Sprintf("%d", paneIndex), message); err != nil {
		return fmt.Errorf("failed to send message to agent %s: %w", agent, err)
	}

	log.Info().Str("agent", agent).Str("message", message).Msg("Message sent to agent")
	return nil
}

// ListAgents - Get agent list
func (ms *MessageServer) ListAgents() ([]string, error) {
	if !ms.running {
		return nil, fmt.Errorf("message server is not running")
	}

	panes, err := ms.tmuxManager.GetPaneList(ms.sessionName)
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

// GetAgentStatus - Get agent status
func (ms *MessageServer) GetAgentStatus(agent string) (bool, error) {
	if !ms.running {
		return false, fmt.Errorf("message server is not running")
	}

	paneIndex, err := ms.getAgentPaneIndex(agent)
	if err != nil {
		return false, fmt.Errorf("failed to get pane index for agent %s: %w", agent, err)
	}

	// Check if tmux pane exists
	panes, err := ms.tmuxManager.GetPaneList(ms.sessionName)
	if err != nil {
		return false, fmt.Errorf("failed to get pane list: %w", err)
	}

	for _, pane := range panes {
		if strings.HasPrefix(pane, fmt.Sprintf("%d:", paneIndex)) {
			return true, nil
		}
	}

	return false, nil
}

// getAgentPaneIndex - Get pane index from agent name
func (ms *MessageServer) getAgentPaneIndex(agent string) (int, error) {
	switch strings.ToLower(agent) {
	case "po":
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

// Stop - Stop server
func (ms *MessageServer) Stop() error {
	if !ms.running {
		return nil
	}

	ms.running = false
	close(ms.shutdown)

	log.Info().Str("session", ms.sessionName).Msg("MessageServer stopped")
	return nil
}

// IsRunning - Check if server is running
func (ms *MessageServer) IsRunning() bool {
	return ms.running
}

// GetSessionName - Get session name
func (ms *MessageServer) GetSessionName() string {
	return ms.sessionName
}
