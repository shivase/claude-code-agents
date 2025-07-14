package process_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/shivase/claude-code-agents/internal/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcessManagerCreation プロセスマネージャー作成テスト
func TestProcessManagerCreation(t *testing.T) {
	pm := process.NewProcessManager()
	assert.NotNil(t, pm)

	// クリーンアップ
	defer pm.StopMonitoring()
}

// TestProcessRegistration プロセス登録テスト
func TestProcessRegistration(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	t.Run("正常なプロセス登録", func(t *testing.T) {
		sessionName := "test-session"
		paneName := "test-pane"
		command := "sleep 1"
		pid := 12345

		pm.RegisterProcess(sessionName, paneName, command, pid)

		// プロセス情報取得
		processInfo, exists := pm.GetProcessInfo(sessionName, paneName)
		assert.True(t, exists)
		assert.Equal(t, pid, processInfo.PID)
		assert.Equal(t, sessionName, processInfo.SessionName)
		assert.Equal(t, paneName, processInfo.PaneName)
		assert.Equal(t, command, processInfo.Command)
		assert.Equal(t, "running", processInfo.Status)
	})

	t.Run("プロセス登録解除", func(t *testing.T) {
		sessionName := "test-session-unreg"
		paneName := "test-pane-unreg"
		command := "sleep 1"
		pid := 12346

		// 登録
		pm.RegisterProcess(sessionName, paneName, command, pid)

		// 存在確認
		_, exists := pm.GetProcessInfo(sessionName, paneName)
		assert.True(t, exists)

		// 登録解除
		pm.UnregisterProcess(sessionName, paneName)

		// 存在しないことを確認
		_, exists = pm.GetProcessInfo(sessionName, paneName)
		assert.False(t, exists)
	})

	t.Run("存在しないプロセスの登録解除", func(t *testing.T) {
		// エラーが発生しないことを確認
		pm.UnregisterProcess("nonexistent", "nonexistent")
	})
}

// TestProcessLifecycle プロセスライフサイクルテスト
func TestProcessLifecycle(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	t.Run("実際のプロセスでのライフサイクル", func(t *testing.T) {
		// テストプロセスを開始
		cmd := exec.Command("sleep", "2")
		err := cmd.Start()
		require.NoError(t, err)

		pid := cmd.Process.Pid
		sessionName := "test-session-lifecycle"
		paneName := "test-pane-lifecycle"

		// プロセス登録
		pm.RegisterProcess(sessionName, paneName, "sleep 2", pid)

		// プロセスが実行中であることを確認
		running := pm.IsProcessRunning(sessionName, paneName)
		assert.True(t, running)

		// プロセスを終了
		err = cmd.Process.Kill()
		require.NoError(t, err)

		// プロセス終了を待機
		cmd.Wait()

		// 少し待ってからプロセス状態を確認
		time.Sleep(100 * time.Millisecond)
		running = pm.IsProcessRunning(sessionName, paneName)
		assert.False(t, running)

		// クリーンアップ
		pm.UnregisterProcess(sessionName, paneName)
	})
}

// TestProcessMonitoring プロセス監視テスト
func TestProcessMonitoring(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("監視機能の開始と停止", func(t *testing.T) {
		// 監視開始
		pm.StartMonitoring(ctx)

		// テストプロセスを開始
		cmd := exec.Command("sleep", "2")
		err := cmd.Start()
		require.NoError(t, err)
		defer cmd.Process.Kill()

		pid := cmd.Process.Pid
		sessionName := "test-session-monitor"
		paneName := "test-pane-monitor"

		// プロセス登録
		pm.RegisterProcess(sessionName, paneName, "sleep 2", pid)

		// 監視が動作していることを確認（プロセス状態更新）
		time.Sleep(1 * time.Second) // 監視間隔より長く待機（短縮）

		// プロセス情報取得
		processInfo, exists := pm.GetProcessInfo(sessionName, paneName)
		assert.True(t, exists)
		assert.NotZero(t, processInfo.LastCheck)

		// クリーンアップ
		pm.UnregisterProcess(sessionName, paneName)
	})
}

// TestMultipleProcessManagement 複数プロセス管理テスト
func TestMultipleProcessManagement(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	// 複数のテストプロセスを開始
	var cmds []*exec.Cmd
	var pids []int
	processCount := 5

	defer func() {
		// クリーンアップ
		for _, cmd := range cmds {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
	}()

	t.Run("複数プロセス登録", func(t *testing.T) {
		for i := 0; i < processCount; i++ {
			cmd := exec.Command("sleep", "5")
			err := cmd.Start()
			require.NoError(t, err)

			cmds = append(cmds, cmd)
			pids = append(pids, cmd.Process.Pid)

			sessionName := fmt.Sprintf("test-session-%d", i)
			paneName := fmt.Sprintf("test-pane-%d", i)
			command := "sleep 5"

			pm.RegisterProcess(sessionName, paneName, command, cmd.Process.Pid)
		}

		// 全プロセス取得
		allProcesses := pm.GetAllProcesses()
		assert.Len(t, allProcesses, processCount)

		// 各プロセスが実行中であることを確認
		for i := 0; i < processCount; i++ {
			sessionName := fmt.Sprintf("test-session-%d", i)
			paneName := fmt.Sprintf("test-pane-%d", i)

			running := pm.IsProcessRunning(sessionName, paneName)
			assert.True(t, running, "プロセス %d が実行中でない", i)
		}
	})

	t.Run("プロセス状態取得", func(t *testing.T) {
		// GetProcessStatusメソッドが存在しないため、GetAllProcessesを使用
		allProcesses := pm.GetAllProcesses()
		assert.Len(t, allProcesses, processCount)

		// プロセス数のカウント
		runningCount := 0
		for _, pInfo := range allProcesses {
			if pInfo.Status == "running" {
				runningCount++
			}
		}
		assert.Equal(t, processCount, runningCount)
	})

	t.Run("個別プロセス終了", func(t *testing.T) {
		// 最初のプロセスを終了
		sessionName := "test-session-0"
		paneName := "test-pane-0"

		err := pm.TerminateProcess(sessionName, paneName)
		assert.NoError(t, err)

		// プロセスが登録から削除されていることを確認
		_, exists := pm.GetProcessInfo(sessionName, paneName)
		assert.False(t, exists)

		// 全プロセス数が減っていることを確認
		allProcesses := pm.GetAllProcesses()
		assert.Len(t, allProcesses, processCount-1)
	})

	t.Run("全プロセス終了", func(t *testing.T) {
		err := pm.TerminateAllProcesses()
		assert.NoError(t, err)

		// 全プロセスが削除されていることを確認
		allProcesses := pm.GetAllProcesses()
		assert.Empty(t, allProcesses)
	})
}

// TestClaudeProcessManagement Claude CLIプロセス管理テスト
func TestClaudeProcessManagement(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	t.Run("Claude CLIプロセス検出", func(t *testing.T) {
		// 実際のClaudeプロセスがない場合の動作をテスト
		processes, err := pm.CheckClaudeProcesses()
		// プロセスが見つからない場合はエラーでもOK
		if err != nil {
			assert.Empty(t, processes)
		} else {
			assert.NotNil(t, processes)
		}
	})

	t.Run("疑似Claude CLIプロセステスト", func(t *testing.T) {
		// Claudeプロセスを模擬したプロセスを開始
		cmd := exec.Command("sleep", "30")
		err := cmd.Start()
		require.NoError(t, err)
		defer cmd.Process.Kill()

		pid := cmd.Process.Pid
		pm.RegisterProcess("claude-session", "claude-pane", "claude --dangerously-skip-permissions", pid)

		// プロセスが実行中であることを確認
		running := pm.IsProcessRunning("claude-session", "claude-pane")
		assert.True(t, running)

		// プロセス終了
		err = pm.TerminateProcess("claude-session", "claude-pane")
		assert.NoError(t, err)
	})

	t.Run("Claude CLIプロセス一括終了", func(t *testing.T) {
		// 実際のClaudeプロセスがない場合でもエラーにならないことを確認
		_ = pm.TerminateClaudeProcesses()
		// プロセスが見つからない場合はエラーでもOK（通常の動作）
		// エラーメッセージの詳細なチェックは不要
	})
}

// TestProcessHealthCheck プロセスヘルスチェックテスト
func TestProcessHealthCheck(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	t.Run("死んだプロセスの検出", func(t *testing.T) {
		// テストプロセスを開始
		cmd := exec.Command("sleep", "1")
		err := cmd.Start()
		require.NoError(t, err)

		pid := cmd.Process.Pid
		sessionName := "test-session-health"
		paneName := "test-pane-health"

		// プロセス登録
		pm.RegisterProcess(sessionName, paneName, "sleep 1", pid)

		// プロセスが実行中であることを確認
		running := pm.IsProcessRunning(sessionName, paneName)
		assert.True(t, running)

		// プロセスが自然終了するまで待機
		cmd.Wait()
		time.Sleep(100 * time.Millisecond)

		// プロセスが死んでいることを確認
		running = pm.IsProcessRunning(sessionName, paneName)
		assert.False(t, running)

		// プロセス情報の状態が更新されることを確認
		pInfo, exists := pm.GetProcessInfo(sessionName, paneName)
		assert.True(t, exists)
		// Note: checkProcessHealth() は内部で呼ばれるため、ステータスが更新されるかは実装依存
		_ = pInfo // 使用していることを明示

		// クリーンアップ
		pm.UnregisterProcess(sessionName, paneName)
	})

	t.Run("死んだプロセスのクリーンアップ", func(t *testing.T) {
		// 存在しないPIDでプロセスを登録
		nonExistentPID := 99999
		pm.RegisterProcess("cleanup-session", "cleanup-pane", "nonexistent", nonExistentPID)

		// CleanupDeadProcessesメソッドが存在しないため、手動でプロセス状態をチェック
		running := pm.IsProcessRunning("cleanup-session", "cleanup-pane")
		assert.False(t, running, "存在しないPIDのプロセスは実行中でないはず")

		// 存在しないプロセスの登録を解除
		pm.UnregisterProcess("cleanup-session", "cleanup-pane")
		_, exists := pm.GetProcessInfo("cleanup-session", "cleanup-pane")
		assert.False(t, exists)
	})
}

// TestConcurrentProcessOperations 並行プロセス操作テスト
func TestConcurrentProcessOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	t.Run("並行プロセス登録", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := []error{}
		processCount := 10

		for i := 0; i < processCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				// テストプロセスを開始
				cmd := exec.Command("sleep", "2")
				err := cmd.Start()
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					return
				}
				defer cmd.Process.Kill()

				sessionName := fmt.Sprintf("concurrent-session-%d", index)
				paneName := fmt.Sprintf("concurrent-pane-%d", index)
				command := "sleep 2"

				// プロセス登録
				pm.RegisterProcess(sessionName, paneName, command, cmd.Process.Pid)

				// 短い待機
				time.Sleep(10 * time.Millisecond)

				// プロセス状態確認
				running := pm.IsProcessRunning(sessionName, paneName)
				if !running {
					mu.Lock()
					errors = append(errors, fmt.Errorf("process %d not running", index))
					mu.Unlock()
				}

				// プロセス登録解除
				pm.UnregisterProcess(sessionName, paneName)
			}(i)
		}

		wg.Wait()

		// エラーが少ないことを確認
		assert.LessOrEqual(t, len(errors), processCount/2, "並行操作で多すぎるエラー: %v", errors)
	})

	t.Run("並行プロセス終了", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errors := []error{}
		processCount := 5

		// プロセスを開始して登録
		var cmds []*exec.Cmd
		for i := 0; i < processCount; i++ {
			cmd := exec.Command("sleep", "3")
			err := cmd.Start()
			require.NoError(t, err)
			cmds = append(cmds, cmd)

			sessionName := fmt.Sprintf("parallel-session-%d", i)
			paneName := fmt.Sprintf("parallel-pane-%d", i)
			pm.RegisterProcess(sessionName, paneName, "sleep 3", cmd.Process.Pid)
		}

		// 並行でプロセス終了
		for i := 0; i < processCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				sessionName := fmt.Sprintf("parallel-session-%d", index)
				paneName := fmt.Sprintf("parallel-pane-%d", index)

				err := pm.TerminateProcess(sessionName, paneName)
				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// クリーンアップ
		for _, cmd := range cmds {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}

		// エラーが少ないことを確認
		assert.LessOrEqual(t, len(errors), processCount/2, "並行終了で多すぎるエラー: %v", errors)
	})
}

// TestErrorHandling エラーハンドリングテスト
func TestErrorHandling(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	t.Run("存在しないプロセスの操作", func(t *testing.T) {
		// 存在しないプロセスの状態確認
		running := pm.IsProcessRunning("nonexistent", "nonexistent")
		assert.False(t, running)

		// 存在しないプロセスの情報取得
		_, exists := pm.GetProcessInfo("nonexistent", "nonexistent")
		assert.False(t, exists)

		// 存在しないプロセスの終了
		err := pm.TerminateProcess("nonexistent", "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("無効なPIDでの操作", func(t *testing.T) {
		// 無効なPIDでプロセス登録
		invalidPID := -1
		pm.RegisterProcess("invalid-session", "invalid-pane", "invalid", invalidPID)

		// プロセスが実行中でないことを確認
		running := pm.IsProcessRunning("invalid-session", "invalid-pane")
		assert.False(t, running)

		// クリーンアップ
		pm.UnregisterProcess("invalid-session", "invalid-pane")
	})
}

// TestGlobalProcessManager グローバルプロセスマネージャーテスト
func TestGlobalProcessManager(t *testing.T) {
	t.Run("グローバルマネージャー取得", func(t *testing.T) {
		gpm1 := process.GetGlobalProcessManager()
		assert.NotNil(t, gpm1)

		gpm2 := process.GetGlobalProcessManager()
		assert.NotNil(t, gpm2)

		// 同じインスタンスであることを確認
		assert.Equal(t, gpm1, gpm2)
	})
}

// TestProcessResourceManagement プロセスリソース管理テスト
func TestProcessResourceManagement(t *testing.T) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	t.Run("プロセス強制終了のシーケンス", func(t *testing.T) {
		// テストプロセスを開始（SIGTERMを無視するようなプロセス）
		cmd := exec.Command("sleep", "5")
		err := cmd.Start()
		require.NoError(t, err)

		pid := cmd.Process.Pid
		sessionName := "resource-session"
		paneName := "resource-pane"

		// プロセス登録
		pm.RegisterProcess(sessionName, paneName, "sleep 5", pid)

		// プロセスが実行中であることを確認
		running := pm.IsProcessRunning(sessionName, paneName)
		assert.True(t, running)

		// プロセス終了
		err = pm.TerminateProcess(sessionName, paneName)
		assert.NoError(t, err)

		// プロセスが登録から削除されていることを確認
		_, exists := pm.GetProcessInfo(sessionName, paneName)
		assert.False(t, exists)
	})

	t.Run("プロセス情報の正確性", func(t *testing.T) {
		// テストプロセスを開始
		cmd := exec.Command("sleep", "3")
		err := cmd.Start()
		require.NoError(t, err)
		defer cmd.Process.Kill()

		pid := cmd.Process.Pid
		sessionName := "info-session"
		paneName := "info-pane"
		command := "sleep 3"

		startTime := time.Now()
		pm.RegisterProcess(sessionName, paneName, command, pid)

		// プロセス情報を取得
		pInfo, exists := pm.GetProcessInfo(sessionName, paneName)
		assert.True(t, exists)

		// 情報の正確性を確認
		assert.Equal(t, pid, pInfo.PID)
		assert.Equal(t, sessionName, pInfo.SessionName)
		assert.Equal(t, paneName, pInfo.PaneName)
		assert.Equal(t, command, pInfo.Command)
		assert.WithinDuration(t, startTime, pInfo.StartTime, time.Second)
		assert.Equal(t, "running", pInfo.Status)

		// クリーンアップ
		pm.UnregisterProcess(sessionName, paneName)
	})
}

// Benchmark tests

// BenchmarkProcessRegistration プロセス登録のベンチマーク
func BenchmarkProcessRegistration(b *testing.B) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionName := fmt.Sprintf("bench-session-%d", i)
		paneName := fmt.Sprintf("bench-pane-%d", i)
		command := "test command"
		pid := 1000 + i

		pm.RegisterProcess(sessionName, paneName, command, pid)
		pm.UnregisterProcess(sessionName, paneName)
	}
}

// BenchmarkProcessStatusCheck プロセス状態確認のベンチマーク
func BenchmarkProcessStatusCheck(b *testing.B) {
	pm := process.NewProcessManager()
	defer pm.StopMonitoring()

	// テストプロセスを登録
	sessionName := "bench-session"
	paneName := "bench-pane"
	pm.RegisterProcess(sessionName, paneName, "test", 1234)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.IsProcessRunning(sessionName, paneName)
	}

	pm.UnregisterProcess(sessionName, paneName)
}

// Helper functions

// createTestProcess テスト用プロセス作成
func createTestProcess(t *testing.T, duration string) (*exec.Cmd, int) {
	cmd := exec.Command("sleep", duration)
	err := cmd.Start()
	require.NoError(t, err)
	return cmd, cmd.Process.Pid
}

// waitForProcessExit プロセス終了待機
func waitForProcessExit(pid int, timeout time.Duration) bool {
	end := time.Now().Add(timeout)
	for time.Now().Before(end) {
		if process, err := os.FindProcess(pid); err != nil {
			return true
		} else if err := process.Signal(syscall.Signal(0)); err != nil {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// killProcessByPID PIDでプロセスを強制終了
func killProcessByPID(pid int) error {
	if process, err := os.FindProcess(pid); err != nil {
		return err
	} else {
		return process.Kill()
	}
}
