package internal

import (
	"fmt"
	"regexp"
	"sort"
)

// Session management methods

func (sm *SessionManager) ListAllSessions() error {
	fmt.Println("📋 Available AI Agent Sessions:")
	fmt.Println("==================================")

	sessions, err := GetTmuxSessions()
	if err != nil {
		return fmt.Errorf("failed to get tmux sessions: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Println("❌ No running tmux sessions")
		return nil
	}

	integratedSessions, individualSessions := sm.categorizeSession(sessions)

	sm.displayIntegratedSessions(integratedSessions)
	sm.displayIndividualSessions(individualSessions)

	if len(integratedSessions) == 0 && len(individualSessions) == 0 {
		fmt.Println()
		fmt.Println("ℹ️ No AI agent related sessions found")
		fmt.Println("💡 Create new session: start-ai-agent [session-name]")
	}

	return nil
}

func (sm *SessionManager) categorizeSession(sessions []Session) ([]Session, map[string]bool) {
	integratedSessions := []Session{}
	individualSessions := map[string]bool{}

	for _, session := range sessions {
		paneCount, err := GetPaneCount(session.Name)
		if err != nil {
			continue
		}

		if paneCount == IntegratedSessionPaneCount {
			integratedSessions = append(integratedSessions, Session{
				Name:  session.Name,
				Type:  "integrated",
				Panes: paneCount,
			})
		} else if sm.isIndividualSession(session.Name) {
			baseName := sm.extractBaseName(session.Name)
			individualSessions[baseName] = true
		}
	}

	return integratedSessions, individualSessions
}

func (sm *SessionManager) isIndividualSession(sessionName string) bool {
	re := regexp.MustCompile(`-(po|manager|dev[1-4])$`)
	return re.MatchString(sessionName)
}

func (sm *SessionManager) extractBaseName(sessionName string) string {
	re := regexp.MustCompile(`-(po|manager|dev[1-4])$`)
	return re.ReplaceAllString(sessionName, "")
}

func (sm *SessionManager) displayIntegratedSessions(sessions []Session) {
	if len(sessions) > 0 {
		fmt.Println()
		fmt.Println("📺 Integrated monitoring screen sessions:")
		for _, session := range sessions {
			fmt.Printf("  🎯 %s (6-pane integrated screen)\n", session.Name)
			fmt.Printf("    Usage: send-agent --session %s po \"message\"\n", session.Name)
		}
	}
}

func (sm *SessionManager) displayIndividualSessions(sessions map[string]bool) {
	if len(sessions) > 0 {
		fmt.Println()
		fmt.Println("🔄 Individual session mode:")
		var baseNames []string
		for baseName := range sessions {
			baseNames = append(baseNames, baseName)
		}
		sort.Strings(baseNames)
		for _, baseName := range baseNames {
			fmt.Printf("  📋 %s group\n", baseName)
			fmt.Printf("    Usage: send-agent --session %s manager \"message\"\n", baseName)
		}
	}
}

func (sm *SessionManager) ShowAgentsForSession(sessionName string) error {
	fmt.Printf("📋 AI Agent Member List (Session: %s):\n", sessionName)
	fmt.Println("==================================================")

	if HasSession(sessionName) {
		return sm.showIntegratedSessionAgents(sessionName)
	}

	return sm.showIndividualSessionAgents(sessionName)
}

func (sm *SessionManager) showIntegratedSessionAgents(sessionName string) error {
	paneCount, err := GetPaneCount(sessionName)
	if err != nil {
		return fmt.Errorf("failed to get information for session '%s': %v", sessionName, err)
	}

	if paneCount == IntegratedSessionPaneCount {
		fmt.Printf("🎯 Using integrated monitoring screen (%s):\n", sessionName)
		sm.displayAgentPaneMapping()
		fmt.Println()
		fmt.Println("Current pane status:")
		return ShowPanes(sessionName)
	}

	return fmt.Errorf("session '%s' is not in integrated monitoring screen format", sessionName)
}

func (sm *SessionManager) showIndividualSessionAgents(sessionName string) error {
	foundSessions := []string{}
	for _, agent := range AvailableAgents {
		fullSession := sessionName + "-" + agent.Name
		if HasSession(fullSession) {
			foundSessions = append(foundSessions, agent.Name)
		}
	}

	if len(foundSessions) > 0 {
		fmt.Printf("🔄 Individual session mode (%s):\n", sessionName)
		for _, agentName := range foundSessions {
			agent := FindAgentByName(agentName)
			if agent != nil {
				fmt.Printf("  %s → %s-%s (%s)\n",
					agentName, sessionName, agentName, agent.Description)
			}
		}
		return nil
	}

	return fmt.Errorf("no AI agent sessions related to session '%s' found\n💡 Available sessions: send-agent list-sessions", sessionName)
}

func (sm *SessionManager) displayAgentPaneMapping() {
	agentPaneMap := map[string]int{
		AgentPO: 0, AgentManager: 1, AgentDev1: 2,
		AgentDev2: 3, AgentDev3: 4, AgentDev4: 5,
	}

	for _, agent := range AvailableAgents {
		paneIndex := agentPaneMap[agent.Name]
		fmt.Printf("  %s → pane %d (%s)\n", agent.Name, paneIndex, agent.Description)
	}
}
