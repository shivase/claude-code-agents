package internal

import (
	"fmt"
	"regexp"
	"sort"
)

// セッション管理メソッド

func (sm *SessionManager) ListAllSessions() error {
	fmt.Println("📋 利用可能なAIエージェントセッション一覧:")
	fmt.Println("==================================")

	sessions, err := GetTmuxSessions()
	if err != nil {
		return fmt.Errorf("tmuxセッションの取得に失敗しました: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Println("❌ 起動中のtmuxセッションがありません")
		return nil
	}

	integratedSessions, individualSessions := sm.categorizeSession(sessions)

	sm.displayIntegratedSessions(integratedSessions)
	sm.displayIndividualSessions(individualSessions)

	if len(integratedSessions) == 0 && len(individualSessions) == 0 {
		fmt.Println()
		fmt.Println("ℹ️ AIエージェント関連のセッションが見つかりませんでした")
		fmt.Println("💡 新しいセッションを作成: start-ai-agent [セッション名]")
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
		fmt.Println("📺 統合監視画面セッション:")
		for _, session := range sessions {
			fmt.Printf("  🎯 %s (6ペイン統合画面)\n", session.Name)
			fmt.Printf("    使用例: send-agent --session %s po \"メッセージ\"\n", session.Name)
		}
	}
}

func (sm *SessionManager) displayIndividualSessions(sessions map[string]bool) {
	if len(sessions) > 0 {
		fmt.Println()
		fmt.Println("🔄 個別セッション方式:")
		var baseNames []string
		for baseName := range sessions {
			baseNames = append(baseNames, baseName)
		}
		sort.Strings(baseNames)
		for _, baseName := range baseNames {
			fmt.Printf("  📋 %s グループ\n", baseName)
			fmt.Printf("    使用例: send-agent --session %s manager \"メッセージ\"\n", baseName)
		}
	}
}

func (sm *SessionManager) ShowAgentsForSession(sessionName string) error {
	fmt.Printf("📋 AIエージェントメンバー一覧 (セッション: %s):\n", sessionName)
	fmt.Println("==================================================")

	if HasSession(sessionName) {
		return sm.showIntegratedSessionAgents(sessionName)
	}

	return sm.showIndividualSessionAgents(sessionName)
}

func (sm *SessionManager) showIntegratedSessionAgents(sessionName string) error {
	paneCount, err := GetPaneCount(sessionName)
	if err != nil {
		return fmt.Errorf("セッション '%s' の情報取得に失敗しました: %v", sessionName, err)
	}

	if paneCount == IntegratedSessionPaneCount {
		fmt.Printf("🎯 統合監視画面（%s）使用中:\n", sessionName)
		sm.displayAgentPaneMapping()
		fmt.Println()
		fmt.Println("現在のペイン状態:")
		return ShowPanes(sessionName)
	}

	return fmt.Errorf("セッション '%s' は統合監視画面形式ではありません", sessionName)
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
		fmt.Printf("🔄 個別セッション方式（%s）:\n", sessionName)
		for _, agentName := range foundSessions {
			agent := FindAgentByName(agentName)
			if agent != nil {
				fmt.Printf("  %s → %s-%s (%s)\n",
					agentName, sessionName, agentName, agent.Description)
			}
		}
		return nil
	}

	return fmt.Errorf("セッション '%s' に関連するAIエージェントセッションが見つかりません\n💡 利用可能なセッション一覧: send-agent list-sessions", sessionName)
}

func (sm *SessionManager) displayAgentPaneMapping() {
	agentPaneMap := map[string]int{
		AgentPO: 0, AgentManager: 1, AgentDev1: 2,
		AgentDev2: 3, AgentDev3: 4, AgentDev4: 5,
	}

	for _, agent := range AvailableAgents {
		paneIndex := agentPaneMap[agent.Name]
		fmt.Printf("  %s → ペイン%d (%s)\n", agent.Name, paneIndex, agent.Description)
	}
}
