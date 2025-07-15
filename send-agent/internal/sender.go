package internal

import (
	"fmt"
	"time"
)

// Helper functions

func IsValidAgent(agent string) bool {
	return ValidAgentNames[agent]
}

func FindAgentByName(name string) *Agent {
	for _, agent := range AvailableAgents {
		if agent.Name == name {
			return &agent
		}
	}
	return nil
}

// Message sending related methods

func (ms *MessageSender) Send() error {
	target, err := ms.determineTarget()
	if err != nil {
		return err
	}

	if err := ms.sendEnhancedMessage(target); err != nil {
		return err
	}

	ms.displaySummary(target)
	return nil
}

func (ms *MessageSender) determineTarget() (string, error) {
	if HasSession(ms.SessionName) {
		return ms.determineIntegratedTarget()
	}
	return ms.determineIndividualTarget()
}

func (ms *MessageSender) determineIntegratedTarget() (string, error) {
	paneCount, err := GetPaneCount(ms.SessionName)
	if err != nil {
		return "", fmt.Errorf("failed to get information for session '%s': %v", ms.SessionName, err)
	}

	if paneCount != IntegratedSessionPaneCount {
		return "", fmt.Errorf("session '%s' is not in integrated monitoring screen format", ms.SessionName)
	}

	fmt.Printf("🎯 Using integrated monitoring screen (%s) to send message\n", ms.SessionName)

	paneIndex := ms.getAgentPaneIndex()
	panes, err := GetPanes(ms.SessionName)
	if err != nil {
		return "", fmt.Errorf("failed to get pane information: %v", err)
	}

	if paneIndex < len(panes) {
		target := fmt.Sprintf("%s.%s", ms.SessionName, panes[paneIndex])
		fmt.Printf("📍 Sending message to %s pane (pane %s)\n", ms.Agent, panes[paneIndex])
		return target, nil
	}

	target := fmt.Sprintf("%s.%d", ms.SessionName, paneIndex)
	fmt.Printf("📍 Sending message to %s pane (pane %d - fallback)\n", ms.Agent, paneIndex)
	return target, nil
}

func (ms *MessageSender) determineIndividualTarget() (string, error) {
	fmt.Printf("🔄 Using individual session mode (%s) to send message\n", ms.SessionName)

	fullSession := ms.SessionName + "-" + ms.Agent
	if !HasSession(fullSession) {
		return "", fmt.Errorf("session '%s' not found", fullSession)
	}

	return fullSession, nil
}

func (ms *MessageSender) getAgentPaneIndex() int {
	agentPaneMap := map[string]int{
		AgentPO: 0, AgentManager: 1, AgentDev1: 2,
		AgentDev2: 3, AgentDev3: 4, AgentDev4: 5,
	}
	return agentPaneMap[ms.Agent]
}

func (ms *MessageSender) sendEnhancedMessage(target string) error {
	fmt.Printf("📤 Sending: sending message to %s...\n", ms.Agent)
	fmt.Printf("🎯 Target: %s\n", target)

	// If context reset is needed
	if ms.ResetContext {
		if err := ms.resetAgentContext(target); err != nil {
			return fmt.Errorf("context reset failed: %v", err)
		}
	}

	// Clear prompt
	fmt.Printf("🧹 Clearing prompt (Ctrl+C)...\n")
	if err := TmuxSendKeys(target, "C-c"); err != nil {
		return fmt.Errorf("prompt clear failed: %v", err)
	}
	time.Sleep(time.Duration(ClearDelay) * time.Millisecond)

	// Additional clear
	fmt.Printf("🧹 Additional clear (Ctrl+U)...\n")
	if err := TmuxSendKeys(target, "C-u"); err != nil {
		return fmt.Errorf("additional clear failed: %v", err)
	}
	time.Sleep(time.Duration(AdditionalClearDelay) * time.Millisecond)

	// Send message
	fmt.Printf("💬 Message sending: \"%s\"\n", ms.Message)
	if err := TmuxSendKeys(target, ms.Message); err != nil {
		return fmt.Errorf("message sending failed: %v", err)
	}
	time.Sleep(time.Duration(MessageDelay) * time.Millisecond)

	// Press Enter
	fmt.Printf("⏎ Sending Enter (C-m)...\n")
	if err := TmuxSendKeys(target, "C-m"); err != nil {
		return fmt.Errorf("Enter sending failed: %v", err)
	}
	time.Sleep(time.Duration(ExecuteDelay) * time.Millisecond)

	fmt.Printf("✅ Sending completed: auto-executed to %s\n", ms.Agent)
	return nil
}

func (ms *MessageSender) resetAgentContext(target string) error {
	fmt.Printf("🔄 Starting context reset...\n")

	resetMessage := "Please forget the previous role definitions and context, and wait for new instructions."

	// Send reset message
	fmt.Printf("💭 Sending reset message: \"%s\"\n", resetMessage)
	if err := TmuxSendKeys(target, resetMessage); err != nil {
		return fmt.Errorf("reset message sending failed: %v", err)
	}
	time.Sleep(time.Duration(MessageDelay) * time.Millisecond)

	// Press Enter
	if err := TmuxSendKeys(target, "C-m"); err != nil {
		return fmt.Errorf("Enter sending failed: %v", err)
	}
	time.Sleep(time.Duration(ExecuteDelay*3) * time.Millisecond) // Wait longer

	fmt.Printf("✅ Context reset completed\n")
	return nil
}

func (ms *MessageSender) displaySummary(target string) {
	fmt.Println()
	fmt.Println("🎯 Message details:")
	fmt.Printf("   Session: %s\n", ms.SessionName)
	fmt.Printf("   Destination: %s (%s)\n", ms.Agent, target)
	fmt.Printf("   Content: \"%s\"\n", ms.Message)
	if ms.ResetContext {
		fmt.Printf("   Context reset: executed\n")
	}
}
