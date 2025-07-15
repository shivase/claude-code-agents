package server

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

// MessageClient - tmux-based message client
type MessageClient struct {
	sessionName string
	tmuxManager *tmux.TmuxManagerImpl
}

// NewMessageClient - Create client (with automatic session detection)
func NewMessageClient(sessionName string) *MessageClient {
	tmuxManager := tmux.NewTmuxManager(sessionName)

	// Auto-detect if session name is empty
	if sessionName == "" {
		expectedPaneCount := 6 // Default value (PO + Manager + 4 Dev)
		if detectedSession, err := tmuxManager.FindDefaultAISession(expectedPaneCount); err == nil {
			sessionName = detectedSession
		} else {
			sessionName = "ai-teams" // Fallback
		}
	}

	// If specified session doesn't exist, try auto-detection
	if !tmuxManager.SessionExists(sessionName) {
		expectedPaneCount := 6 // Default value (PO + Manager + 4 Dev)
		if detectedSession, err := tmuxManager.FindDefaultAISession(expectedPaneCount); err == nil {
			sessionName = detectedSession
		}
	}

	return &MessageClient{
		sessionName: sessionName,
		tmuxManager: tmuxManager,
	}
}

// SendMessage - Send message via tmux
func (mc *MessageClient) SendMessage(agent, message string) error {
	paneIndex, err := mc.getAgentPaneIndex(agent)
	if err != nil {
		return fmt.Errorf("failed to get pane index for agent %s: %w", agent, err)
	}

	// Send message to tmux pane
	if err := mc.tmuxManager.SendKeysWithEnter(mc.sessionName, strconv.Itoa(paneIndex), message); err != nil {
		return fmt.Errorf("failed to send message to agent %s: %w", agent, err)
	}

	log.Info().Str("agent", agent).Str("message", message).Msg("Message sent to agent")
	return nil
}

// ListAgents - Get agent list via tmux
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

// GetStatus - Get agent status via tmux
func (mc *MessageClient) GetStatus(agent string) (bool, error) {
	paneIndex, err := mc.getAgentPaneIndex(agent)
	if err != nil {
		return false, fmt.Errorf("failed to get pane index for agent %s: %w", agent, err)
	}

	// Check if tmux pane exists
	cmd := exec.Command("tmux", "list-panes", "-t", mc.sessionName) // #nosec G204
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check pane status: %w", err)
	}

	// Check if pane exists
	return strings.Contains(string(output), fmt.Sprintf("%d:", paneIndex)), nil
}

// IsServerRunning - Check tmux session running state
func (mc *MessageClient) IsServerRunning() bool {
	return mc.tmuxManager.SessionExists(mc.sessionName)
}

// GetSessionName - Get session name
func (mc *MessageClient) GetSessionName() string {
	return mc.sessionName
}

// CheckConnection - Check tmux session connection (flexible detection)
func (mc *MessageClient) CheckConnection() error {
	// First try auto-detection (if session doesn't exist or exists but not an AI session)
	if !mc.tmuxManager.SessionExists(mc.sessionName) {
		detectedSession, sessionType, err := mc.tmuxManager.DetectActiveAISession(6)
		if err != nil {
			return fmt.Errorf("no active AI sessions found: %w", err)
		}

		mc.sessionName = detectedSession
		log.Info().Str("session", detectedSession).Str("type", sessionType).Msg("Auto-detected AI session")
	} else {
		// Check if existing session is an AI session
		paneCount, err := mc.tmuxManager.GetPaneCount(mc.sessionName)
		if err != nil {
			return fmt.Errorf("failed to get pane count for session %s: %w", mc.sessionName, err)
		}

		// If not 6 panes or 1 pane, detect other AI sessions
		if paneCount != 6 && paneCount != 1 {
			detectedSession, sessionType, err := mc.tmuxManager.DetectActiveAISession(6)
			if err != nil {
				return fmt.Errorf("session %s has %d panes (not AI session) and no other AI sessions found: %w", mc.sessionName, paneCount, err)
			}

			mc.sessionName = detectedSession
			log.Info().Str("session", detectedSession).Str("type", sessionType).Msg("Auto-detected AI session (current session not AI)")
		}
	}

	// Final pane count check
	paneCount, err := mc.tmuxManager.GetPaneCount(mc.sessionName)
	if err != nil {
		return fmt.Errorf("failed to get pane count for session %s: %w", mc.sessionName, err)
	}

	// Allow integrated monitoring (6 panes) or individual session (1 pane)
	if paneCount != 6 && paneCount != 1 {
		return fmt.Errorf("expected 6 panes (integrated) or 1 pane (individual) but found %d in session %s", paneCount, mc.sessionName)
	}

	return nil
}

// getAgentPaneIndex - Get pane index from agent name (same configuration as claude.sh)
func (mc *MessageClient) getAgentPaneIndex(agent string) (int, error) {
	switch strings.ToLower(agent) {
	case "po":
		return 1, nil // Top left
	case "manager":
		return 2, nil // Bottom left
	case "dev1":
		return 3, nil // Top right
	case "dev2":
		return 4, nil // Top right middle
	case "dev3":
		return 5, nil // Bottom right middle
	case "dev4":
		return 6, nil // Bottom right
	default:
		return -1, fmt.Errorf("unknown agent: %s", agent)
	}
}
