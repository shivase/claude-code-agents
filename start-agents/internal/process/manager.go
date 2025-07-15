package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

// ProcessManagerImpl implements process management functionality (containedctx fixed version)
type ProcessManagerImpl struct {
	processes map[string]*ProcessInfo
	cancel    context.CancelFunc
	mu        sync.RWMutex // Mutex for concurrency safety
}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManagerImpl {
	_, cancel := context.WithCancel(context.Background())
	return &ProcessManagerImpl{
		processes: make(map[string]*ProcessInfo),
		cancel:    cancel,
	}
}

// StartMonitoring starts process monitoring
func (pm *ProcessManagerImpl) StartMonitoring(ctx context.Context) {
	go pm.monitorProcesses(ctx)
}

// StopMonitoring stops process monitoring
func (pm *ProcessManagerImpl) StopMonitoring() {
	pm.cancel()
}

// RegisterProcess registers a process
func (pm *ProcessManagerImpl) RegisterProcess(sessionName, paneName, command string, pid int) {
	key := fmt.Sprintf("%s:%s", sessionName, paneName)

	pm.mu.Lock()
	pm.processes[key] = &ProcessInfo{
		PID:         pid,
		SessionName: sessionName,
		PaneName:    paneName,
		Command:     command,
		StartTime:   time.Now(),
		Status:      "running",
		LastCheck:   time.Now(),
	}
	pm.mu.Unlock()

	log.Info().Str("session", sessionName).Str("pane", paneName).Int("pid", pid).Msg("Process registered")
}

// UnregisterProcess unregisters a process
func (pm *ProcessManagerImpl) UnregisterProcess(sessionName, paneName string) {
	key := fmt.Sprintf("%s:%s", sessionName, paneName)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if process, exists := pm.processes[key]; exists {
		log.Info().Str("session", sessionName).Str("pane", paneName).Int("pid", process.PID).Msg("Process unregistered")
		delete(pm.processes, key)
	}
}

// IsProcessRunning checks if a process is running
func (pm *ProcessManagerImpl) IsProcessRunning(sessionName, paneName string) bool {
	key := fmt.Sprintf("%s:%s", sessionName, paneName)

	pm.mu.RLock()
	process, exists := pm.processes[key]
	pm.mu.RUnlock()

	if exists {
		return pm.isProcessAlive(process.PID)
	}
	return false
}

// GetProcessInfo retrieves process information
func (pm *ProcessManagerImpl) GetProcessInfo(sessionName, paneName string) (*ProcessInfo, bool) {
	key := fmt.Sprintf("%s:%s", sessionName, paneName)

	pm.mu.RLock()
	process, exists := pm.processes[key]
	pm.mu.RUnlock()

	return process, exists
}

// GetAllProcesses retrieves all process information
func (pm *ProcessManagerImpl) GetAllProcesses() map[string]*ProcessInfo {
	result := make(map[string]*ProcessInfo)

	pm.mu.RLock()
	for key, process := range pm.processes {
		result[key] = process
	}
	pm.mu.RUnlock()

	return result
}

// TerminateProcess forcefully terminates a process
func (pm *ProcessManagerImpl) TerminateProcess(sessionName, paneName string) error {
	key := fmt.Sprintf("%s:%s", sessionName, paneName)

	pm.mu.Lock()
	process, exists := pm.processes[key]
	if exists {
		delete(pm.processes, key)
	}
	pm.mu.Unlock()

	if exists {
		if err := pm.killProcess(process.PID); err != nil {
			return fmt.Errorf("failed to terminate process: %w", err)
		}
		log.Info().Str("session", sessionName).Str("pane", paneName).Int("pid", process.PID).Msg("Process terminated")
		return nil
	}
	return fmt.Errorf("process not found: %s:%s", sessionName, paneName)
}

// TerminateAllProcesses forcefully terminates all processes
func (pm *ProcessManagerImpl) TerminateAllProcesses() error {
	// Channel and WaitGroup for parallel processing
	type killResult struct {
		key string
		pid int
		err error
	}

	// Safely copy process information
	pm.mu.RLock()
	processes := make(map[string]*ProcessInfo)
	for key, process := range pm.processes {
		processes[key] = process
	}
	pm.mu.RUnlock()

	resultChan := make(chan killResult, len(processes))
	var wg sync.WaitGroup

	// Send SIGTERM to all processes in parallel
	for key, process := range processes {
		wg.Add(1)
		go func(k string, p *ProcessInfo) {
			defer wg.Done()

			proc, err := os.FindProcess(p.PID)
			if err != nil {
				resultChan <- killResult{key: k, pid: p.PID, err: err}
				return
			}

			// Send SIGTERM
			if err := proc.Signal(os.Interrupt); err != nil {
				// Do not treat as error if already terminated
				if !pm.isProcessAlive(p.PID) {
					resultChan <- killResult{key: k, pid: p.PID, err: nil}
					return
				}
				resultChan <- killResult{key: k, pid: p.PID, err: err}
			} else {
				resultChan <- killResult{key: k, pid: p.PID, err: nil}
			}
		}(key, process)
	}

	// Wait for goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var errors []error
	sigTermResults := make(map[string]int)
	for result := range resultChan {
		if result.err != nil {
			errors = append(errors, fmt.Errorf("failed to send SIGTERM to process %s: %w", result.key, result.err))
		} else {
			sigTermResults[result.key] = result.pid
		}
	}

	// Gradual wait time (CI environment support)
	maxWaitTime := 3 * time.Second
	checkInterval := 100 * time.Millisecond

	// Reduce wait time in CI environment
	if os.Getenv("CI") == "true" || os.Getenv("CLAUDE_MOCK_ENV") == "true" {
		maxWaitTime = 500 * time.Millisecond
		checkInterval = 50 * time.Millisecond
	}

	// Wait for processes to terminate
	deadline := time.Now().Add(maxWaitTime)
	remainingPIDs := make(map[string]int)
	for k, pid := range sigTermResults {
		remainingPIDs[k] = pid
	}

	for time.Now().Before(deadline) && len(remainingPIDs) > 0 {
		for k, pid := range remainingPIDs {
			if !pm.isProcessAlive(pid) {
				log.Info().Str("key", k).Int("pid", pid).Msg("Process terminated gracefully")
				delete(remainingPIDs, k)
			}
		}
		if len(remainingPIDs) > 0 {
			time.Sleep(checkInterval)
		}
	}

	// Send SIGKILL to processes still alive
	for k, pid := range remainingPIDs {
		proc, err := os.FindProcess(pid)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to find process %s for SIGKILL: %w", k, err))
			continue
		}
		if err := proc.Kill(); err != nil {
			// すでに終了している場合はエラーとしない
			if pm.isProcessAlive(pid) {
				errors = append(errors, fmt.Errorf("failed to kill process %s: %w", k, err))
			}
		} else {
			log.Info().Str("key", k).Int("pid", pid).Msg("Process force killed")
		}
	}

	// Clear all processes
	pm.mu.Lock()
	pm.processes = make(map[string]*ProcessInfo)
	pm.mu.Unlock()

	if len(errors) > 0 {
		return fmt.Errorf("some processes failed to terminate: %w", errors[0])
	}
	return nil
}

// CheckClaudeProcesses checks for Claude CLI processes
func (pm *ProcessManagerImpl) CheckClaudeProcesses() ([]ProcessInfo, error) {
	cmd := exec.Command("pgrep", "-f", "claude.*--dangerously-skip-permissions")
	output, err := cmd.Output()
	if err != nil {
		// Not an error if no processes found
		return []ProcessInfo{}, err
	}

	var processes []ProcessInfo
	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, pidStr := range pids {
		if pidStr == "" {
			continue
		}

		// Retrieve process information
		if processInfo, err := pm.getProcessInfoByPID(pidStr); err == nil {
			processes = append(processes, *processInfo)
		}
	}

	return processes, nil
}

// TerminateClaudeProcesses forcefully terminates Claude CLI processes
func (pm *ProcessManagerImpl) TerminateClaudeProcesses() error {
	log.Info().Msg("Terminating Claude CLI processes")

	// Search for Claude processes using pgrep
	processes, err := pm.CheckClaudeProcesses()
	if err != nil {
		return fmt.Errorf("failed to check Claude processes: %w", err)
	}

	if len(processes) == 0 {
		log.Info().Msg("No Claude processes found")
		return nil
	}

	// Terminate each process
	var errors []error
	for _, process := range processes {
		if err := pm.killProcess(process.PID); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate PID %d: %w", process.PID, err))
		} else {
			log.Info().Int("pid", process.PID).Msg("Claude process terminated")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some Claude processes failed to terminate: %w", errors[0])
	}

	log.Info().Int("count", len(processes)).Msg("All Claude processes terminated")
	return nil
}

// monitorProcesses process monitoring loop (receives context as argument)
func (pm *ProcessManagerImpl) monitorProcesses(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Process monitoring stopped")
			return
		case <-ticker.C:
			pm.checkProcessHealth()
		}
	}
}

// checkProcessHealth checks process health
func (pm *ProcessManagerImpl) checkProcessHealth() {
	deadProcesses := 0

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for key, process := range pm.processes {
		if !pm.isProcessAlive(process.PID) {
			if process.Status != "dead" {
				log.Warn().
					Str("key", key).
					Int("pid", process.PID).
					Str("session", process.SessionName).
					Str("pane", process.PaneName).
					Time("start_time", process.StartTime).
					Msg("Dead process detected")
				deadProcesses++
			}
			process.Status = "dead"
		} else {
			process.Status = "running"
		}
		process.LastCheck = time.Now()
	}

	if deadProcesses > 0 {
		log.Debug().Int("dead_count", deadProcesses).Msg("Process health check completed")
	}
}

// isProcessAlive checks if a process is alive
func (pm *ProcessManagerImpl) isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to process to check existence (using syscall.Signal(0))
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}

// killProcess forcefully terminates a process
func (pm *ProcessManagerImpl) killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// First try graceful termination with SIGTERM
	if err := process.Signal(os.Interrupt); err == nil {
		// Gradual wait time (CI environment support)
		maxWaitTime := 3 * time.Second
		checkInterval := 100 * time.Millisecond

		// Reduce wait time in CI environment
		if os.Getenv("CI") == "true" || os.Getenv("CLAUDE_MOCK_ENV") == "true" {
			maxWaitTime = 500 * time.Millisecond
			checkInterval = 50 * time.Millisecond
		}

		// Gradually check for process termination
		deadline := time.Now().Add(maxWaitTime)
		for time.Now().Before(deadline) {
			if !pm.isProcessAlive(pid) {
				return nil
			}
			time.Sleep(checkInterval)
		}
	}

	// Force termination with SIGKILL
	return process.Kill()
}

// getProcessInfoByPID retrieves process information by PID
func (pm *ProcessManagerImpl) getProcessInfoByPID(pidStr string) (*ProcessInfo, error) {
	cmd := exec.Command("ps", "-p", pidStr, "-o", "pid,command")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process info: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid ps output")
	}

	// Extract PID and command
	re := regexp.MustCompile(`^\s*(\d+)\s+(.+)$`)
	matches := re.FindStringSubmatch(lines[1])
	if len(matches) < 3 {
		return nil, fmt.Errorf("failed to parse ps output")
	}

	pid := 0
	if _, err := fmt.Sscanf(matches[1], "%d", &pid); err != nil {
		return nil, fmt.Errorf("failed to parse PID: %w", err)
	}

	return &ProcessInfo{
		PID:       pid,
		Command:   matches[2],
		StartTime: time.Now(), // Exact start time is difficult to obtain
		Status:    "running",
		LastCheck: time.Now(),
	}, nil
}

// GetProcessStatus retrieves process status
func (pm *ProcessManagerImpl) GetProcessStatus() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := map[string]interface{}{
		"total_processes": len(pm.processes),
		"running_count":   0,
		"dead_count":      0,
		"processes":       make([]map[string]interface{}, 0),
	}

	for key, process := range pm.processes {
		processData := map[string]interface{}{
			"key":        key,
			"pid":        process.PID,
			"session":    process.SessionName,
			"pane":       process.PaneName,
			"command":    process.Command,
			"start_time": process.StartTime.Format("2006-01-02 15:04:05"),
			"status":     process.Status,
			"last_check": process.LastCheck.Format("2006-01-02 15:04:05"),
		}

		result["processes"] = append(result["processes"].([]map[string]interface{}), processData)

		if process.Status == "running" {
			result["running_count"] = result["running_count"].(int) + 1
		} else {
			result["dead_count"] = result["dead_count"].(int) + 1
		}
	}

	return result
}

// CleanupDeadProcesses removes dead processes
func (pm *ProcessManagerImpl) CleanupDeadProcesses() int {
	cleanupCount := 0
	keysToDelete := make([]string, 0)

	pm.mu.RLock()
	for key, process := range pm.processes {
		if process.Status == "dead" || !pm.isProcessAlive(process.PID) {
			keysToDelete = append(keysToDelete, key)
			log.Info().Str("key", key).Int("pid", process.PID).Msg("Dead process cleaned up")
		}
	}
	pm.mu.RUnlock()

	pm.mu.Lock()
	for _, key := range keysToDelete {
		delete(pm.processes, key)
		cleanupCount++
	}
	pm.mu.Unlock()

	return cleanupCount
}

// Global process manager instance
var globalProcessManager *ProcessManagerImpl

// GetGlobalProcessManager retrieves the global process manager
func GetGlobalProcessManager() *ProcessManagerImpl {
	if globalProcessManager == nil {
		globalProcessManager = NewProcessManager()
		// Changed to pass context from the caller
	}
	return globalProcessManager
}
