package internal_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shivase/cloud-code-agents/send-agent/internal"
)

// tmux関数の基本的な機能テスト
// 注意: これらのテストは実際のtmuxが動作していることを前提とする

func TestGetTmuxSessions_Structure(t *testing.T) {
	t.Run("戻り値の構造確認", func(t *testing.T) {
		sessions, err := internal.GetTmuxSessions()

		// tmuxが存在しない場合はエラーが返されることを確認
		if err != nil {
			assert.Contains(t, err.Error(), "tmuxセッション一覧の取得に失敗しました")
			return
		}

		// 正常に取得できる場合は各セッションにNameが設定されていることを確認
		for _, session := range sessions {
			assert.NotEmpty(t, session.Name, "セッション名が空でない")
		}
	})
}

func TestHasSession_Behavior(t *testing.T) {
	t.Run("存在しないセッションの場合", func(t *testing.T) {
		// 存在しないセッション名での動作確認
		result := internal.HasSession("non-existing-session-12345")
		assert.False(t, result, "存在しないセッションはfalseを返す")
	})
}

func TestGetPaneCount_Structure(t *testing.T) {
	t.Run("存在しないセッションの場合", func(t *testing.T) {
		count, err := internal.GetPaneCount("non-existing-session-12345")
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "ペイン数の取得に失敗しました")
	})
}

func TestGetPanes_Structure(t *testing.T) {
	t.Run("存在しないセッションの場合", func(t *testing.T) {
		panes, err := internal.GetPanes("non-existing-session-12345")
		assert.Error(t, err)
		assert.Nil(t, panes)
		assert.Contains(t, err.Error(), "ペイン一覧の取得に失敗しました")
	})
}

func TestShowPanes_Structure(t *testing.T) {
	t.Run("存在しないセッションの場合", func(t *testing.T) {
		err := internal.ShowPanes("non-existing-session-12345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ペイン状態の表示に失敗しました")
	})
}

func TestTmuxSendKeys_Structure(t *testing.T) {
	t.Run("存在しないセッションの場合", func(t *testing.T) {
		err := internal.TmuxSendKeys("non-existing-session-12345:0", "echo hello")
		// send-keysは存在しないセッションでもエラーを返すことを確認
		assert.Error(t, err)
	})
}

func TestDetectDefaultSession_Structure(t *testing.T) {
	t.Run("戻り値の構造確認", func(t *testing.T) {
		session, err := internal.DetectDefaultSession()

		// tmuxが存在しない場合はエラーが返されることを確認
		if err != nil {
			assert.Contains(t, err.Error(), "tmuxセッションが見つかりません")
			return
		}

		// 正常に取得できる場合はセッション名が空でないことを確認
		assert.NotEmpty(t, session, "セッション名が空でない")
	})
}

// 統合テスト用のヘルパー関数
func TestIntegratedSessionDetection(t *testing.T) {
	// 正規表現パターンのテスト
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
			// 正規表現のテスト
			re := regexp.MustCompile(`-(po|manager|dev[1-4])$`)
			result := re.MatchString(tt.session)
			assert.Equal(t, tt.expected, result)
		})
	}
}
