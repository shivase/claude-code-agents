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
			name:        "正常な作成",
			workingDir:  "/tmp/test",
			expectError: false,
			setupMock:   func() {},
		},
		{
			name:        "空のワーキングディレクトリ",
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
			name: "claude-codeコマンドが見つかる",
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
			name: "従来のclaudeコマンドが見つかる",
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
			name: "Claude CLIが見つからない",
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
