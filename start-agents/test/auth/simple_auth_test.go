package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shivase/claude-code-agents/internal/auth"
)

// TestNewClaudeAuthManager - Test creating authentication manager
func TestNewClaudeAuthManager(t *testing.T) {
	manager := auth.NewClaudeAuthManager()
	assert.NotNil(t, manager)
}

// TestNewPreAuthChecker - Test creating pre-authentication checker
func TestNewPreAuthChecker(t *testing.T) {
	claudePath := "/test/path"
	checker := auth.NewPreAuthChecker(claudePath)
	assert.NotNil(t, checker)
	// claudePath is a private field so it cannot be accessed directly
}
