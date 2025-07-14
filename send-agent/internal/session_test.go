package internal

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SessionManager構造体のテスト
func TestSessionManager_New(t *testing.T) {
	sm := &SessionManager{}
	assert.NotNil(t, sm)
}

// isIndividualSession()関数のテスト
func TestSessionManager_isIndividualSession(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		expected    bool
	}{
		{
			name:        "PO session",
			sessionName: "project-po",
			expected:    true,
		},
		{
			name:        "Manager session",
			sessionName: "project-manager",
			expected:    true,
		},
		{
			name:        "Dev1 session",
			sessionName: "project-dev1",
			expected:    true,
		},
		{
			name:        "Dev2 session",
			sessionName: "project-dev2",
			expected:    true,
		},
		{
			name:        "Dev3 session",
			sessionName: "project-dev3",
			expected:    true,
		},
		{
			name:        "Dev4 session",
			sessionName: "project-dev4",
			expected:    true,
		},
		{
			name:        "Regular session",
			sessionName: "regular-session",
			expected:    false,
		},
		{
			name:        "Invalid agent name",
			sessionName: "project-dev5",
			expected:    false,
		},
		{
			name:        "Partial match",
			sessionName: "project-po-backup",
			expected:    false,
		},
		{
			name:        "Empty session name",
			sessionName: "",
			expected:    false,
		},
	}

	sm := &SessionManager{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.isIndividualSession(tt.sessionName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// extractBaseName()関数のテスト
func TestSessionManager_extractBaseName(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		expected    string
	}{
		{
			name:        "PO session",
			sessionName: "project-po",
			expected:    "project",
		},
		{
			name:        "Manager session",
			sessionName: "project-manager",
			expected:    "project",
		},
		{
			name:        "Dev1 session",
			sessionName: "project-dev1",
			expected:    "project",
		},
		{
			name:        "Complex base name",
			sessionName: "my-complex-project-name-dev2",
			expected:    "my-complex-project-name",
		},
		{
			name:        "Regular session (no change)",
			sessionName: "regular-session",
			expected:    "regular-session",
		},
		{
			name:        "Empty session name",
			sessionName: "",
			expected:    "",
		},
	}

	sm := &SessionManager{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.extractBaseName(tt.sessionName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 正規表現のテスト（個別テスト）
func TestRegexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		matches bool
	}{
		{
			name:    "PO pattern match",
			pattern: `-(po|manager|dev[1-4])$`,
			input:   "project-po",
			matches: true,
		},
		{
			name:    "Manager pattern match",
			pattern: `-(po|manager|dev[1-4])$`,
			input:   "project-manager",
			matches: true,
		},
		{
			name:    "Dev1 pattern match",
			pattern: `-(po|manager|dev[1-4])$`,
			input:   "project-dev1",
			matches: true,
		},
		{
			name:    "Dev4 pattern match",
			pattern: `-(po|manager|dev[1-4])$`,
			input:   "project-dev4",
			matches: true,
		},
		{
			name:    "Invalid dev number",
			pattern: `-(po|manager|dev[1-4])$`,
			input:   "project-dev5",
			matches: false,
		},
		{
			name:    "No agent suffix",
			pattern: `-(po|manager|dev[1-4])$`,
			input:   "project",
			matches: false,
		},
		{
			name:    "Middle match (should not match)",
			pattern: `-(po|manager|dev[1-4])$`,
			input:   "project-po-backup",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.pattern)
			result := re.MatchString(tt.input)
			assert.Equal(t, tt.matches, result)
		})
	}
}
