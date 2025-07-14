package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shivase/claude-code-agents/internal/auth"
)

// TestNewClaudeAuthManager - 認証マネージャー作成のテスト
func TestNewClaudeAuthManager(t *testing.T) {
	manager := auth.NewClaudeAuthManager()
	assert.NotNil(t, manager)
}

// TestNewPreAuthChecker - 事前認証チェッカー作成のテスト
func TestNewPreAuthChecker(t *testing.T) {
	claudePath := "/test/path"
	checker := auth.NewPreAuthChecker(claudePath)
	assert.NotNil(t, checker)
	// claudePathはプライベートフィールドなので直接アクセスできない
}
