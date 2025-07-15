package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentStruct(t *testing.T) {
	t.Run("AgentCreation", func(t *testing.T) {
		agent := Agent{
			Name:        "test-agent",
			Description: "Test agent description",
		}

		assert.Equal(t, "test-agent", agent.Name)
		assert.Equal(t, "Test agent description", agent.Description)
	})

	t.Run("AgentZeroValue", func(t *testing.T) {
		var agent Agent

		assert.Equal(t, "", agent.Name)
		assert.Equal(t, "", agent.Description)
	})
}

func TestSessionStruct(t *testing.T) {
	t.Run("SessionCreation", func(t *testing.T) {
		session := Session{
			Name:  "test-session",
			Type:  "integrated",
			Panes: 6,
		}

		assert.Equal(t, "test-session", session.Name)
		assert.Equal(t, "integrated", session.Type)
		assert.Equal(t, 6, session.Panes)
	})

	t.Run("SessionZeroValue", func(t *testing.T) {
		var session Session

		assert.Equal(t, "", session.Name)
		assert.Equal(t, "", session.Type)
		assert.Equal(t, 0, session.Panes)
	})
}

func TestSessionManagerStruct(t *testing.T) {
	t.Run("SessionManagerCreation", func(t *testing.T) {
		manager := SessionManager{
			sessions: []Session{
				{Name: "session1", Type: "integrated", Panes: 6},
				{Name: "session2", Type: "individual", Panes: 1},
			},
		}

		assert.Len(t, manager.sessions, 2)
		assert.Equal(t, "session1", manager.sessions[0].Name)
		assert.Equal(t, "session2", manager.sessions[1].Name)
	})

	t.Run("SessionManagerZeroValue", func(t *testing.T) {
		var manager SessionManager

		assert.Nil(t, manager.sessions)
	})
}

func TestMessageSenderStruct(t *testing.T) {
	t.Run("MessageSenderCreation", func(t *testing.T) {
		sender := MessageSender{
			SessionName:  "test-session",
			Agent:        "manager",
			Message:      "test message",
			ResetContext: true,
		}

		assert.Equal(t, "test-session", sender.SessionName)
		assert.Equal(t, "manager", sender.Agent)
		assert.Equal(t, "test message", sender.Message)
		assert.True(t, sender.ResetContext)
	})

	t.Run("MessageSenderZeroValue", func(t *testing.T) {
		var sender MessageSender

		assert.Equal(t, "", sender.SessionName)
		assert.Equal(t, "", sender.Agent)
		assert.Equal(t, "", sender.Message)
		assert.False(t, sender.ResetContext)
	})
}

func TestIsValidAgentFunction(t *testing.T) {
	tests := []struct {
		name     string
		agent    string
		expected bool
	}{
		{"ValidPO", AgentPO, true},
		{"ValidManager", AgentManager, true},
		{"ValidDev1", AgentDev1, true},
		{"ValidDev2", AgentDev2, true},
		{"ValidDev3", AgentDev3, true},
		{"ValidDev4", AgentDev4, true},
		{"InvalidAgent", "invalid", false},
		{"EmptyString", "", false},
		{"CaseSensitive", "PO", false},
		{"CaseSensitive2", "Manager", false},
		{"PartialMatch", "dev", false},
		{"NumericOnly", "1", false},
		{"SpecialChars", "dev@1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAgent(tt.agent)
			assert.Equal(t, tt.expected, result, "IsValidAgent(%s) = %v, want %v", tt.agent, result, tt.expected)
		})
	}
}

func TestFindAgentByNameFunction(t *testing.T) {
	tests := []struct {
		name     string
		agent    string
		expected *Agent
	}{
		{
			"FindPO",
			AgentPO,
			&Agent{AgentPO, "Product Owner (Product Manager)"},
		},
		{
			"FindManager",
			AgentManager,
			&Agent{AgentManager, "Project Manager (Flexible team management)"},
		},
		{
			"FindDev1",
			AgentDev1,
			&Agent{AgentDev1, "Execution Agent 1 (Flexible role assignment)"},
		},
		{
			"FindDev2",
			AgentDev2,
			&Agent{AgentDev2, "Execution Agent 2 (Flexible role assignment)"},
		},
		{
			"FindDev3",
			AgentDev3,
			&Agent{AgentDev3, "Execution Agent 3 (Flexible role assignment)"},
		},
		{
			"FindDev4",
			AgentDev4,
			&Agent{AgentDev4, "Execution Agent 4 (Flexible role assignment)"},
		},
		{
			"NotFoundInvalid",
			"invalid",
			nil,
		},
		{
			"NotFoundEmpty",
			"",
			nil,
		},
		{
			"NotFoundCaseSensitive",
			"PO",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindAgentByName(tt.agent)

			if tt.expected == nil {
				assert.Nil(t, result, "FindAgentByName(%s) should return nil", tt.agent)
			} else {
				assert.NotNil(t, result, "FindAgentByName(%s) should not return nil", tt.agent)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Description, result.Description)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	t.Run("IntegratedSessionPaneCount", func(t *testing.T) {
		assert.Equal(t, 6, IntegratedSessionPaneCount)
	})

	t.Run("DelayConstants", func(t *testing.T) {
		assert.Equal(t, 400, ClearDelay)
		assert.Equal(t, 200, AdditionalClearDelay)
		assert.Equal(t, 300, MessageDelay)
		assert.Equal(t, 500, ExecuteDelay)
	})

	t.Run("AgentConstants", func(t *testing.T) {
		assert.Equal(t, "po", AgentPO)
		assert.Equal(t, "manager", AgentManager)
		assert.Equal(t, "dev1", AgentDev1)
		assert.Equal(t, "dev2", AgentDev2)
		assert.Equal(t, "dev3", AgentDev3)
		assert.Equal(t, "dev4", AgentDev4)
	})
}

func TestAvailableAgents(t *testing.T) {
	t.Run("AvailableAgentsLength", func(t *testing.T) {
		assert.Len(t, AvailableAgents, 6)
	})

	t.Run("AvailableAgentsContent", func(t *testing.T) {
		expectedAgents := []Agent{
			{AgentPO, "Product Owner (Product Manager)"},
			{AgentManager, "Project Manager (Flexible team management)"},
			{AgentDev1, "Execution Agent 1 (Flexible role assignment)"},
			{AgentDev2, "Execution Agent 2 (Flexible role assignment)"},
			{AgentDev3, "Execution Agent 3 (Flexible role assignment)"},
			{AgentDev4, "Execution Agent 4 (Flexible role assignment)"},
		}

		assert.Equal(t, expectedAgents, AvailableAgents)
	})

	t.Run("AvailableAgentsIndividualCheck", func(t *testing.T) {
		// 各エージェントが正しく設定されているかチェック
		agentMap := make(map[string]Agent)
		for _, agent := range AvailableAgents {
			agentMap[agent.Name] = agent
		}

		// POのチェック
		po, exists := agentMap[AgentPO]
		assert.True(t, exists, "PO agent should exist")
		assert.Contains(t, po.Description, "Product Owner")

		// Managerのチェック
		manager, exists := agentMap[AgentManager]
		assert.True(t, exists, "Manager agent should exist")
		assert.Contains(t, manager.Description, "Project Manager")

		// Dev1-4のチェック
		for i := 1; i <= 4; i++ {
			devName := "dev" + string(rune('0'+i))
			dev, exists := agentMap[devName]
			assert.True(t, exists, "Dev%d agent should exist", i)
			assert.Contains(t, dev.Description, "Execution Agent")
		}
	})
}

func TestValidAgentNames(t *testing.T) {
	t.Run("ValidAgentNamesLength", func(t *testing.T) {
		assert.Len(t, ValidAgentNames, 6)
	})

	t.Run("ValidAgentNamesContent", func(t *testing.T) {
		expectedMap := map[string]bool{
			AgentPO: true, AgentManager: true, AgentDev1: true,
			AgentDev2: true, AgentDev3: true, AgentDev4: true,
		}

		assert.Equal(t, expectedMap, ValidAgentNames)
	})

	t.Run("ValidAgentNamesConsistency", func(t *testing.T) {
		// AvailableAgentsとValidAgentNamesの整合性チェック
		for _, agent := range AvailableAgents {
			assert.True(t, ValidAgentNames[agent.Name], "Agent %s should be in ValidAgentNames", agent.Name)
		}

		// ValidAgentNamesの各キーがAvailableAgentsに存在するかチェック
		for name := range ValidAgentNames {
			found := false
			for _, agent := range AvailableAgents {
				if agent.Name == name {
					found = true
					break
				}
			}
			assert.True(t, found, "ValidAgentNames key %s should exist in AvailableAgents", name)
		}
	})
}

func TestAgentNameConsistency(t *testing.T) {
	t.Run("AgentConstantsConsistency", func(t *testing.T) {
		// 定数で定義されたエージェント名がAvailableAgentsに存在するかチェック
		agentConstants := []string{
			AgentPO, AgentManager, AgentDev1, AgentDev2, AgentDev3, AgentDev4,
		}

		for _, constant := range agentConstants {
			found := false
			for _, agent := range AvailableAgents {
				if agent.Name == constant {
					found = true
					break
				}
			}
			assert.True(t, found, "Agent constant %s should exist in AvailableAgents", constant)
		}
	})
}

func TestStructFields(t *testing.T) {
	t.Run("AgentStructFields", func(t *testing.T) {
		agent := Agent{
			Name:        "test",
			Description: "test desc",
		}

		// フィールドが正しく設定されているかチェック
		assert.Equal(t, "test", agent.Name)
		assert.Equal(t, "test desc", agent.Description)

		// フィールドが変更可能かチェック
		agent.Name = "modified"
		agent.Description = "modified desc"
		assert.Equal(t, "modified", agent.Name)
		assert.Equal(t, "modified desc", agent.Description)
	})

	t.Run("SessionStructFields", func(t *testing.T) {
		session := Session{
			Name:  "test",
			Type:  "test-type",
			Panes: 10,
		}

		// フィールドが正しく設定されているかチェック
		assert.Equal(t, "test", session.Name)
		assert.Equal(t, "test-type", session.Type)
		assert.Equal(t, 10, session.Panes)

		// フィールドが変更可能かチェック
		session.Name = "modified"
		session.Type = "modified-type"
		session.Panes = 20
		assert.Equal(t, "modified", session.Name)
		assert.Equal(t, "modified-type", session.Type)
		assert.Equal(t, 20, session.Panes)
	})

	t.Run("MessageSenderStructFields", func(t *testing.T) {
		sender := MessageSender{
			SessionName:  "test",
			Agent:        "test-agent",
			Message:      "test message",
			ResetContext: true,
		}

		// フィールドが正しく設定されているかチェック
		assert.Equal(t, "test", sender.SessionName)
		assert.Equal(t, "test-agent", sender.Agent)
		assert.Equal(t, "test message", sender.Message)
		assert.True(t, sender.ResetContext)

		// フィールドが変更可能かチェック
		sender.SessionName = "modified"
		sender.Agent = "modified-agent"
		sender.Message = "modified message"
		sender.ResetContext = false
		assert.Equal(t, "modified", sender.SessionName)
		assert.Equal(t, "modified-agent", sender.Agent)
		assert.Equal(t, "modified message", sender.Message)
		assert.False(t, sender.ResetContext)
	})
}
