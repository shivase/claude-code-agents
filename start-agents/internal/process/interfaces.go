package process

import (
	"fmt"
	"time"
)

// ProcessManagerInterface defines the interface for process management
type ProcessManagerInterface interface {
	// GetClaudeProcesses retrieves all Claude processes
	GetClaudeProcesses() ([]*ProcessInfo, error)
	// CheckClaudeProcesses checks the status of Claude processes
	CheckClaudeProcesses() ([]*ProcessInfo, error)
	// KillClaudeProcesses forcefully terminates all Claude processes
	KillClaudeProcesses() error
	// WaitForClaudeProcesses waits for Claude processes to terminate
	WaitForClaudeProcesses(timeout time.Duration) error
	// GetProcessInfo retrieves information about a specific process
	GetProcessInfo(pid int) (*ProcessInfo, error)
	// IsProcessRunning checks if a process is currently running
	IsProcessRunning(pid int) bool
	// KillProcess terminates a specific process
	KillProcess(pid int) error
	// MonitorProcesses monitors running processes
	MonitorProcesses() error
	// GetProcessesByName searches for processes by name
	GetProcessesByName(name string) ([]*ProcessInfo, error)
	// GetProcessCounts retrieves the count of processes by type
	GetProcessCounts() (map[string]int, error)
	// CleanupProcesses cleans up terminated processes
	CleanupProcesses() error
}

// ProcessInfo contains information about a process
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

// String returns a string representation of the process info
func (p *ProcessInfo) String() string {
	return fmt.Sprintf("PID: %d, Name: %s, Command: %s, Start: %s, CPU: %s, Memory: %s, Status: %s",
		p.PID, p.Name, p.Command, p.StartTime.Format("2006-01-02 15:04:05"), p.CPUPercent, p.MemoryUsage, p.Status)
}

// FileLockInterface defines the interface for file lock management
type FileLockInterface interface {
	// Lock acquires a file lock
	Lock() error
	// Unlock releases a file lock
	Unlock() error
	// IsLocked checks if the file is locked
	IsLocked() bool
}
