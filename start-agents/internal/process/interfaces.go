package process

import (
	"fmt"
	"time"
)

// ProcessManagerInterface プロセス管理インターフェース
type ProcessManagerInterface interface {
	// GetClaudeProcesses クローデプロセスの取得
	GetClaudeProcesses() ([]*ProcessInfo, error)
	// CheckClaudeProcesses クローデプロセスの確認
	CheckClaudeProcesses() ([]*ProcessInfo, error)
	// KillClaudeProcesses クローデプロセスの強制終了
	KillClaudeProcesses() error
	// WaitForClaudeProcesses クローデプロセスの終了待機
	WaitForClaudeProcesses(timeout time.Duration) error
	// GetProcessInfo プロセス情報の取得
	GetProcessInfo(pid int) (*ProcessInfo, error)
	// IsProcessRunning プロセスの実行確認
	IsProcessRunning(pid int) bool
	// KillProcess プロセスの終了
	KillProcess(pid int) error
	// MonitorProcesses プロセスの監視
	MonitorProcesses() error
	// GetProcessesByName 名前でプロセスを検索
	GetProcessesByName(name string) ([]*ProcessInfo, error)
	// GetProcessCounts プロセス数の取得
	GetProcessCounts() (map[string]int, error)
	// CleanupProcesses プロセスのクリーンアップ
	CleanupProcesses() error
}

// ProcessInfo プロセス情報
type ProcessInfo struct {
	PID         int       `json:"pid"`
	Name        string    `json:"name"`
	Command     string    `json:"command"`
	StartTime   time.Time `json:"start_time"`
	CPUPercent  string    `json:"cpu_percent"`
	MemoryUsage string    `json:"memory_usage"`
	Status      string    `json:"status"`
	SessionName string    `json:"session_name"`
	PaneName    string    `json:"pane_name"`
	LastCheck   time.Time `json:"last_check"`
}

// String プロセス情報を文字列で表示
func (p *ProcessInfo) String() string {
	return fmt.Sprintf("PID: %d, Name: %s, Command: %s, Start: %s, CPU: %s, Memory: %s, Status: %s",
		p.PID, p.Name, p.Command, p.StartTime.Format("2006-01-02 15:04:05"), p.CPUPercent, p.MemoryUsage, p.Status)
}

// FileLockInterface ファイルロック管理インターフェース
type FileLockInterface interface {
	// Lock ファイルをロック
	Lock() error
	// Unlock ファイルをアンロック
	Unlock() error
	// IsLocked ロック状態を確認
	IsLocked() bool
}
