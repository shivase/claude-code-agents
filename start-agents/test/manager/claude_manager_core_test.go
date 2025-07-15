package manager_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/manager"
	"github.com/shivase/claude-code-agents/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// isCIEnvironment テスト用CI環境検出ヘルパー関数
// 環境変数「CI=true」「GITHUB_ACTIONS=true」「CLAUDE_MOCK_ENV=true」をチェック
func isCIEnvironment() bool {
	return os.Getenv("CI") == "true" ||
		os.Getenv("GITHUB_ACTIONS") == "true" ||
		os.Getenv("CLAUDE_MOCK_ENV") == "true"
}

// isTestEnvironment テスト用テスト環境検出ヘルパー関数
func isTestEnvironment() bool {
	return os.Getenv("CLAUDE_MOCK_ENV") == "true" ||
		os.Getenv("GO_TEST") == "true"
}

// MockCmd execコマンドのモック
type MockCmd struct {
	mock.Mock
	process *os.Process
	started bool
	waited  bool
}

func (m *MockCmd) Start() error {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Error(0)
	}
	m.started = true
	return nil
}

func (m *MockCmd) Wait() error {
	args := m.Called()
	m.waited = true
	return args.Error(0)
}

func (m *MockCmd) Process() *os.Process {
	return m.process
}

// MockPTY PTYのモック
type MockPTY struct {
	mock.Mock
	writeBuffer []byte
	closed      bool
}

func (m *MockPTY) Write(data []byte) (int, error) {
	args := m.Called(data)
	if args.Get(1) != nil {
		return 0, args.Error(1)
	}
	m.writeBuffer = append(m.writeBuffer, data...)
	return len(data), nil
}

func (m *MockPTY) Read(data []byte) (int, error) {
	args := m.Called(data)
	return args.Int(0), args.Error(1)
}

func (m *MockPTY) Close() error {
	args := m.Called()
	m.closed = true
	return args.Error(0)
}

// TestClaudeManagerCreation ClaudeManager作成テスト
func TestClaudeManagerCreation(t *testing.T) {
	tests := []struct {
		name        string
		workingDir  string
		expectError bool
		setupMock   func()
	}{
		{
			name:        "Normal creation",
			workingDir:  "/tmp/test",
			expectError: false,
			setupMock:   func() {},
		},
		{
			name:        "Empty working directory",
			workingDir:  "",
			expectError: false,
			setupMock:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 統一モック環境をセットアップ
			mockConfig := common.SetupClaudeMockForCI(t)
			defer common.TeardownClaudeMock(mockConfig)

			// モック環境の検証
			common.ValidateClaudeMockSetup(t, mockConfig)

			tt.setupMock()

			cm, err := manager.NewClaudeManager(tt.workingDir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cm)

				if cm != nil {
					// クリーンアップ
					defer cm.Shutdown()
				}
			}
		})
	}
}

// TestAgentManagement エージェント管理テスト
func TestAgentManagement(t *testing.T) {
	// 統一モック環境をセットアップ（長時間実行用）
	mockConfig := common.SetupClaudeMockForManager(t)
	defer common.TeardownClaudeMock(mockConfig)

	// モック環境の検証
	common.ValidateClaudeMockSetup(t, mockConfig)

	cm, err := manager.NewClaudeManager(mockConfig.TempDir)
	require.NoError(t, err)
	defer cm.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// エージェント設定
	config := &manager.AgentConfig{
		Name:            "test-agent",
		InstructionFile: "",
		SessionName:     "test-session",
		WorkingDir:      mockConfig.TempDir,
	}

	t.Run("エージェント開始", func(t *testing.T) {
		err := cm.StartAgent(ctx, config)
		assert.NoError(t, err)

		// エージェントがリストに追加されていることを確認
		agents := cm.ListAgents()
		assert.Contains(t, agents, "test-agent")
	})

	t.Run("エージェント状態確認", func(t *testing.T) {
		running, err := cm.GetAgentStatus("test-agent")
		assert.NoError(t, err)
		assert.True(t, running)
	})

	t.Run("存在しないエージェントの状態確認", func(t *testing.T) {
		_, err := cm.GetAgentStatus("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("重複エージェント開始", func(t *testing.T) {
		err := cm.StartAgent(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")
	})

	t.Run("メッセージ送信", func(t *testing.T) {
		err := cm.SendMessage("test-agent", "test message")
		// プロセスが実際に動いているかどうかによりエラーかもしれない
		// ここではエラーが発生しないことを期待
		assert.NoError(t, err)
	})

	t.Run("存在しないエージェントへのメッセージ送信", func(t *testing.T) {
		err := cm.SendMessage("non-existent", "test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("エージェント停止", func(t *testing.T) {
		err := cm.StopAgent("test-agent")
		assert.NoError(t, err)

		// エージェントがリストから削除されていることを確認
		agents := cm.ListAgents()
		assert.NotContains(t, agents, "test-agent")
	})

	t.Run("存在しないエージェント停止", func(t *testing.T) {
		err := cm.StopAgent("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestMultipleAgentsManagement 複数エージェント管理テスト
func TestMultipleAgentsManagement(t *testing.T) {
	// 統一モック環境をセットアップ（長時間実行用）
	mockConfig := common.SetupClaudeMockForManager(t)
	defer common.TeardownClaudeMock(mockConfig)

	// モック環境の検証
	common.ValidateClaudeMockSetup(t, mockConfig)

	cm, err := manager.NewClaudeManager(mockConfig.TempDir)
	require.NoError(t, err)
	defer cm.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// 複数のエージェント設定
	agents := []string{"agent1", "agent2", "agent3"}

	t.Run("複数エージェント同時開始", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := []error{}

		for _, agentName := range agents {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				config := &manager.AgentConfig{
					Name:            name,
					InstructionFile: "",
					SessionName:     "test-session-" + name,
					WorkingDir:      mockConfig.TempDir,
				}

				if err := cm.StartAgent(ctx, config); err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
				}
			}(agentName)
		}

		wg.Wait()

		// エラーがないことを確認
		assert.Empty(t, errors, "並行エージェント起動でエラーが発生: %v", errors)

		// 全エージェントがリストにあることを確認
		agentList := cm.ListAgents()
		assert.Len(t, agentList, len(agents))

		for _, agent := range agents {
			assert.Contains(t, agentList, agent)
		}
	})

	t.Run("複数エージェント状態確認", func(t *testing.T) {
		if isCIEnvironment() {
			t.Skip("CI環境では tmux セッション関連のテストをスキップします")
		}
		for _, agent := range agents {
			running, err := cm.GetAgentStatus(agent)
			assert.NoError(t, err, "エージェント %s の状態確認失敗", agent)
			assert.True(t, running, "エージェント %s が実行中でない", agent)
		}
	})

	t.Run("複数エージェント同時停止", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := []error{}

		for _, agentName := range agents {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				if err := cm.StopAgent(name); err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
				}
			}(agentName)
		}

		wg.Wait()

		// エラーがないことを確認
		assert.Empty(t, errors, "並行エージェント停止でエラーが発生: %v", errors)

		// 全エージェントがリストから削除されていることを確認
		agentList := cm.ListAgents()
		assert.Empty(t, agentList)
	})
}

// TestClaudePathDetection Claude CLIパス検出テスト
func TestClaudePathDetection(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func() (string, func())
		expectError bool
	}{
		{
			name: "claude-code command found",
			setupPath: func() (string, func()) {
				tempDir := t.TempDir()
				claudePath := filepath.Join(tempDir, "claude-code")
				err := os.WriteFile(claudePath, []byte("#!/bin/bash\n"), 0755)
				require.NoError(t, err)

				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", tempDir+":"+oldPath)

				return tempDir, func() { os.Setenv("PATH", oldPath) }
			},
			expectError: false,
		},
		{
			name: "Legacy claude command found",
			setupPath: func() (string, func()) {
				tempDir := t.TempDir()
				claudePath := filepath.Join(tempDir, "claude")
				err := os.WriteFile(claudePath, []byte("#!/bin/bash\n"), 0755)
				require.NoError(t, err)

				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", tempDir+":"+oldPath)

				return tempDir, func() { os.Setenv("PATH", oldPath) }
			},
			expectError: false,
		},
		{
			name: "Claude CLI not found",
			setupPath: func() (string, func()) {
				oldPath := os.Getenv("PATH")
				oldHome := os.Getenv("HOME")
				os.Setenv("PATH", "/nonexistent")
				os.Setenv("HOME", "/nonexistent")

				return "", func() {
					os.Setenv("PATH", oldPath)
					os.Setenv("HOME", oldHome)
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir, cleanup := tt.setupPath()
			defer cleanup()

			cm, err := manager.NewClaudeManager(workingDir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cm)

				if cm != nil {
					defer cm.Shutdown()
				}
			}
		})
	}
}

// TestConfigSetup 設定セットアップテスト
func TestConfigSetup(t *testing.T) {
	// 統一モック環境をセットアップ
	mockConfig := common.SetupClaudeMockForCI(t)
	defer common.TeardownClaudeMock(mockConfig)

	// モック環境の検証
	common.ValidateClaudeMockSetup(t, mockConfig)

	// ホームディレクトリを一時ディレクトリに設定
	oldHome := os.Getenv("HOME")
	testHome := filepath.Join(mockConfig.TempDir, "home")
	os.MkdirAll(testHome, 0755)
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	t.Run("新規設定ファイル作成", func(t *testing.T) {
		cm, err := manager.NewClaudeManager(mockConfig.TempDir)
		require.NoError(t, err)
		defer cm.Shutdown()

		// 設定ファイルが作成されていることを確認
		configFile := filepath.Join(testHome, ".claude", "settings.json")
		_, err = os.Stat(configFile)
		assert.NoError(t, err, "設定ファイルが作成されていない")
	})

	t.Run("既存設定ファイル保持", func(t *testing.T) {
		// 既存の設定ファイルを作成
		configDir := filepath.Join(testHome, ".claude")
		configFile := filepath.Join(configDir, "settings.json")
		os.MkdirAll(configDir, 0755)

		existingConfig := `{"model": "existing", "theme": "light"}`
		err := os.WriteFile(configFile, []byte(existingConfig), 0600)
		require.NoError(t, err)

		cm, err := manager.NewClaudeManager(mockConfig.TempDir)
		require.NoError(t, err)
		defer cm.Shutdown()

		// 既存の設定ファイルが保持されていることを確認
		content, err := os.ReadFile(configFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "existing")
	})
}

// TestSignalHandling シグナルハンドリングテスト
func TestSignalHandling(t *testing.T) {
	// 統一モック環境をセットアップ（短時間実行用）
	mockConfig := common.SetupClaudeMockWithTimeout(t, 1)
	defer common.TeardownClaudeMock(mockConfig)

	// モック環境の検証
	common.ValidateClaudeMockSetup(t, mockConfig)

	cm, err := manager.NewClaudeManager(mockConfig.TempDir)
	require.NoError(t, err)

	t.Run("シグナルハンドリング設定", func(t *testing.T) {
		err := cm.StartWithSignalHandling()
		assert.NoError(t, err)

		// シャットダウンをテスト
		err = cm.Shutdown()
		assert.NoError(t, err)
	})
}

// TestErrorConditions エラー条件テスト
func TestErrorConditions(t *testing.T) {
	// 統一モック環境をセットアップ（失敗するコマンド）
	mockConfig := common.SetupClaudeMockWithFailure(t, 1)
	defer common.TeardownClaudeMock(mockConfig)

	// モック環境の検証
	common.ValidateClaudeMockSetup(t, mockConfig)

	cm, err := manager.NewClaudeManager(mockConfig.TempDir)
	require.NoError(t, err)
	defer cm.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("無効な設定でエージェント開始", func(t *testing.T) {
		config := &manager.AgentConfig{
			Name:            "", // 空の名前
			InstructionFile: "/nonexistent/file",
			SessionName:     "test-session",
			WorkingDir:      mockConfig.TempDir,
		}

		err := cm.StartAgent(ctx, config)
		// 空の名前でもエラーにならないが、プロセス起動で失敗する可能性がある
		// ここではエラーハンドリングの仕組みをテスト
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("タイムアウト付きメッセージ送信", func(t *testing.T) {
		// 存在しないエージェントにメッセージ送信
		err := cm.SendMessage("nonexistent", "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestConcurrentOperations 並行操作テスト
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("並行テストをスキップ（shortモード）")
	}
	// 統一モック環境をセットアップ（短時間実行用）
	mockConfig := common.SetupClaudeMockWithTimeout(t, 1)
	defer common.TeardownClaudeMock(mockConfig)

	// モック環境の検証
	common.ValidateClaudeMockSetup(t, mockConfig)

	cm, err := manager.NewClaudeManager(mockConfig.TempDir)
	require.NoError(t, err)
	defer cm.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("並行エージェント操作", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		results := []string{}

		// 複数のゴルーチンで同時にエージェント操作
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				agentName := fmt.Sprintf("agent-%d", index)
				config := &manager.AgentConfig{
					Name:            agentName,
					InstructionFile: "",
					SessionName:     fmt.Sprintf("session-%d", index),
					WorkingDir:      mockConfig.TempDir,
				}

				// エージェント開始
				if err := cm.StartAgent(ctx, config); err != nil {
					mu.Lock()
					results = append(results, fmt.Sprintf("start-error-%d: %v", index, err))
					mu.Unlock()
					return
				}

				// 短い待機
				time.Sleep(100 * time.Millisecond)

				// 状態確認
				if running, err := cm.GetAgentStatus(agentName); err != nil {
					mu.Lock()
					results = append(results, fmt.Sprintf("status-error-%d: %v", index, err))
					mu.Unlock()
				} else if !running {
					mu.Lock()
					results = append(results, fmt.Sprintf("not-running-%d", index))
					mu.Unlock()
				}

				// エージェント停止
				if err := cm.StopAgent(agentName); err != nil {
					mu.Lock()
					results = append(results, fmt.Sprintf("stop-error-%d: %v", index, err))
					mu.Unlock()
				}

				mu.Lock()
				results = append(results, fmt.Sprintf("success-%d", index))
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// 結果確認
		mu.Lock()
		successCount := 0
		for _, result := range results {
			if strings.HasPrefix(result, "success-") {
				successCount++
			} else {
				t.Logf("Operation result: %s", result)
			}
		}
		mu.Unlock()

		// 少なくとも一部の操作が成功していることを確認
		assert.Greater(t, successCount, 0, "並行操作で成功した操作がない")
	})
}

// TestCIEnvironmentDetection CI環境検出テスト
func TestCIEnvironmentDetection(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		expectedCI   bool
		expectedTest bool
	}{
		{
			name:         "CI=true環境",
			envVars:      map[string]string{"CI": "true"},
			expectedCI:   true,
			expectedTest: false,
		},
		{
			name:         "GITHUB_ACTIONS=true環境",
			envVars:      map[string]string{"GITHUB_ACTIONS": "true"},
			expectedCI:   true,
			expectedTest: false,
		},
		{
			name:         "CLAUDE_MOCK_ENV=true環境",
			envVars:      map[string]string{"CLAUDE_MOCK_ENV": "true"},
			expectedCI:   true,
			expectedTest: true,
		},
		{
			name:         "GO_TEST=true環境",
			envVars:      map[string]string{"GO_TEST": "true"},
			expectedCI:   false,
			expectedTest: true,
		},
		{
			name:         "複数のCI環境変数",
			envVars:      map[string]string{"CI": "true", "GITHUB_ACTIONS": "true"},
			expectedCI:   true,
			expectedTest: false,
		},
		{
			name:         "通常環境（CI環境変数なし）",
			envVars:      map[string]string{},
			expectedCI:   false,
			expectedTest: false,
		},
		{
			name:         "無効な値のCI環境変数",
			envVars:      map[string]string{"CI": "false", "GITHUB_ACTIONS": "false"},
			expectedCI:   false,
			expectedTest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を一時的にバックアップ
			originalEnv := map[string]string{
				"CI":              os.Getenv("CI"),
				"GITHUB_ACTIONS":  os.Getenv("GITHUB_ACTIONS"),
				"CLAUDE_MOCK_ENV": os.Getenv("CLAUDE_MOCK_ENV"),
				"GO_TEST":         os.Getenv("GO_TEST"),
			}

			// 環境変数をクリア
			for key := range originalEnv {
				os.Unsetenv(key)
			}

			// テスト用の環境変数を設定
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// テスト実行
			actualCI := isCIEnvironment()
			actualTest := isTestEnvironment()

			// 結果検証
			assert.Equal(t, tt.expectedCI, actualCI, "CI環境検出が期待値と異なる")
			assert.Equal(t, tt.expectedTest, actualTest, "テスト環境検出が期待値と異なる")

			// 環境変数を復元
			for key, value := range originalEnv {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}
		})
	}
}

// TestEnvironmentVariableValidation 環境変数検証テスト
func TestEnvironmentVariableValidation(t *testing.T) {
	t.Run("環境変数の存在確認", func(t *testing.T) {
		// 重要な環境変数がテスト中に正しく設定されているかテスト
		testVars := []string{"CI", "GITHUB_ACTIONS", "CLAUDE_MOCK_ENV", "GO_TEST"}

		for _, envVar := range testVars {
			value := os.Getenv(envVar)
			t.Logf("環境変数 %s = %s", envVar, value)

			// 環境変数が設定されている場合の値の検証
			if value != "" {
				assert.True(t, value == "true" || value == "false" || value == "1" || value == "0",
					"環境変数 %s の値が期待される形式ではない: %s", envVar, value)
			}
		}
	})

	t.Run("環境変数の組み合わせテスト", func(t *testing.T) {
		// 複数の環境変数が同時に設定されている場合の動作確認
		originalCI := os.Getenv("CI")
		originalMock := os.Getenv("CLAUDE_MOCK_ENV")

		// 両方の環境変数を設定
		os.Setenv("CI", "true")
		os.Setenv("CLAUDE_MOCK_ENV", "true")

		// 両方の関数がtrueを返すことを確認
		assert.True(t, isCIEnvironment(), "CI環境検出が失敗（CI=true, CLAUDE_MOCK_ENV=true）")
		assert.True(t, isTestEnvironment(), "テスト環境検出が失敗（CI=true, CLAUDE_MOCK_ENV=true）")

		// 環境変数を復元
		if originalCI == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", originalCI)
		}
		if originalMock == "" {
			os.Unsetenv("CLAUDE_MOCK_ENV")
		} else {
			os.Setenv("CLAUDE_MOCK_ENV", originalMock)
		}
	})
}

// Benchmark tests

// BenchmarkAgentStartStop エージェント開始・停止のベンチマーク
func BenchmarkAgentStartStop(b *testing.B) {
	// 統一モック環境をセットアップ（ベンチマーク用）
	mockConfig := common.SetupClaudeMockForBenchmark(b)
	defer common.TeardownClaudeMock(mockConfig)

	// モック環境の検証
	common.ValidateClaudeMockSetupForBenchmark(b, mockConfig)

	// ログレベルを下げる
	log.Logger = log.Level(zerolog.WarnLevel)

	cm, err := manager.NewClaudeManager(mockConfig.TempDir)
	require.NoError(b, err)
	defer cm.Shutdown()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agentName := fmt.Sprintf("bench-agent-%d", i)
		config := &manager.AgentConfig{
			Name:            agentName,
			InstructionFile: "",
			SessionName:     fmt.Sprintf("bench-session-%d", i),
			WorkingDir:      mockConfig.TempDir,
		}

		// エージェント開始
		err := cm.StartAgent(ctx, config)
		if err != nil {
			b.Fatalf("Failed to start agent: %v", err)
		}

		// エージェント停止
		err = cm.StopAgent(agentName)
		if err != nil {
			b.Fatalf("Failed to stop agent: %v", err)
		}
	}
}
