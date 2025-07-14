package auth_test

import (
	"fmt"
	"sync"

	"github.com/stretchr/testify/mock"

	"github.com/shivase/claude-code-agents/internal/auth"
)

// MockAuthProvider - AuthProviderInterfaceのモック実装
type MockAuthProvider struct {
	mock.Mock
	mu sync.RWMutex
}

// CheckAuth - 認証状態確認のモック
func (m *MockAuthProvider) CheckAuth() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called()
	return args.Error(0)
}

// CheckSettingsFile - 設定ファイル確認のモック
func (m *MockAuthProvider) CheckSettingsFile() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called()
	return args.Error(0)
}

// IsReady - Claude CLI準備確認のモック
func (m *MockAuthProvider) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called()
	return args.Bool(0)
}

// GetPath - Claude CLIパス取得のモック
func (m *MockAuthProvider) GetPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called()
	return args.String(0)
}

// ValidateSetup - セットアップ検証のモック
func (m *MockAuthProvider) ValidateSetup() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called()
	return args.Error(0)
}

// CreateTestAuthStatus - テスト用認証状態を作成
func CreateTestAuthStatus(authenticated bool, userID string) *auth.AuthStatus {
	status := &auth.AuthStatus{
		IsAuthenticated: authenticated,
		UserID:          userID,
		LastChecked:     1234567890,
	}

	if authenticated && userID != "" {
		status.OAuthAccount = map[string]interface{}{
			"emailAddress": "test@example.com",
			"provider":     "google",
		}
	}

	return status
}

// MockAuthError - テスト用認証エラー
type MockAuthError struct {
	Code    string
	Message string
}

func (e *MockAuthError) Error() string {
	return fmt.Sprintf("Auth Error [%s]: %s", e.Code, e.Message)
}
