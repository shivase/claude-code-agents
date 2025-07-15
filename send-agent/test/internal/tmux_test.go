package internal_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shivase/cloud-code-agents/send-agent/internal"
)

// Basic functional tests for tmux functions
// Note: These tests assume that tmux is actually running

func TestGetTmuxSessions_Structure(t *testing.T) {
	t.Run("戻り値の構造確認", func(t *testing.T) {
		sessions, err := internal.GetTmuxSessions()

		// Verify that an error is returned if tmux doesn't exist
		if err != nil {
			assert.Contains(t, err.Error(), "failed to get tmux session list")
			return
		}

		// Verify that each session has a Name set when successfully retrieved
		for _, session := range sessions {
			assert.NotEmpty(t, session.Name, "セッション名が空でない")
		}
	})
}

func TestHasSession_Behavior(t *testing.T) {
	t.Run("Non-existing session", func(t *testing.T) {
		// Test behavior with non-existing session name
		result := internal.HasSession("non-existing-session-12345")
		assert.False(t, result, "Non-existing session should return false")
	})
}

func TestGetPaneCount_Structure(t *testing.T) {
	t.Run("Non-existing session", func(t *testing.T) {
		count, err := internal.GetPaneCount("non-existing-session-12345")
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get pane count")
	})
}

func TestGetPanes_Structure(t *testing.T) {
	t.Run("Non-existing session", func(t *testing.T) {
		panes, err := internal.GetPanes("non-existing-session-12345")
		assert.Error(t, err)
		assert.Nil(t, panes)
		assert.Contains(t, err.Error(), "failed to get pane list")
	})
}

func TestShowPanes_Structure(t *testing.T) {
	t.Run("Non-existing session", func(t *testing.T) {
		err := internal.ShowPanes("non-existing-session-12345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to display pane status")
	})
}

func TestTmuxSendKeys_Structure(t *testing.T) {
	t.Run("Non-existing session", func(t *testing.T) {
		err := internal.TmuxSendKeys("non-existing-session-12345:0", "echo hello")
		// Verify that send-keys returns an error even for non-existing sessions
		assert.Error(t, err)
	})
}

func TestDetectDefaultSession_Structure(t *testing.T) {
	t.Run("戻り値の構造確認", func(t *testing.T) {
		session, err := internal.DetectDefaultSession()

		// Verify that an error is returned if tmux doesn't exist
		if err != nil {
			assert.Contains(t, err.Error(), "no tmux sessions found")
			return
		}

		// Verify that session name is not empty when successfully retrieved
		assert.NotEmpty(t, session, "セッション名が空でない")
	})
}

// Helper functions for integration tests
func TestIntegratedSessionDetection(t *testing.T) {
	// Test regular expression patterns
	tests := []struct {
		name     string
		session  string
		expected bool
	}{
		{"PO agent session", "project-po", true},
		{"Manager agent session", "project-manager", true},
		{"Dev1 agent session", "project-dev1", true},
		{"Dev2 agent session", "project-dev2", true},
		{"Dev3 agent session", "project-dev3", true},
		{"Dev4 agent session", "project-dev4", true},
		{"Normal session", "project-normal", false},
		{"Short session", "p", false},
		{"AI keyword session", "ai-test", false},
		{"Claude keyword session", "claude-test", false},
		{"Agent keyword session", "agent-test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test regular expression
			re := regexp.MustCompile(`-(po|manager|dev[1-4])$`)
			result := re.MatchString(tt.session)
			assert.Equal(t, tt.expected, result)
		})
	}
}
