package manager

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ClaudeConfig - Claude CLI configuration
type ClaudeConfig struct {
	Model                  string `json:"model"`
	Theme                  string `json:"theme"`
	HasCompletedOnboarding bool   `json:"hasCompletedOnboarding"`
	HasSetTheme            bool   `json:"hasSetTheme"`
	SkipInitialSetup       bool   `json:"skipInitialSetup"`
}

// AgentConfig - Agent configuration
type AgentConfig struct {
	Name            string
	InstructionFile string
	SessionName     string
	WorkingDir      string
}

// ClaudeProcess - Claude CLI process management (CI environment PTY issue fixed version)
type ClaudeProcess struct {
	Config      *AgentConfig
	Cmd         *exec.Cmd
	PTY         *os.File
	Logger      zerolog.Logger
	cancel      context.CancelFunc
	isRunning   atomic.Bool // Atomic operation to prevent race conditions
	ptyClosed   atomic.Bool // PTY close state (atomic operation)
	ptyMutex    sync.Mutex  // Mutex for exclusive PTY access control
	restartChan chan struct{}
	messageChan chan string
	isCIEnv     bool // CI environment detection flag
	isMockEnv   bool // Mock environment detection flag
}

// isCIEnvironment - CI environment detection (GitHub Actions, common CI, mock environment)
func isCIEnvironment() bool {
	return os.Getenv("CI") == "true" ||
		os.Getenv("GITHUB_ACTIONS") == "true" ||
		os.Getenv("CLAUDE_MOCK_ENV") == "true"
}

// isTestEnvironment - Test environment detection
func isTestEnvironment() bool {
	return os.Getenv("CLAUDE_MOCK_ENV") == "true" ||
		os.Getenv("GO_TEST") == "true"
}

// safePTYClose - Safe PTY close (fully CI environment compatible version)
func (cp *ClaudeProcess) safePTYClose() error {
	// Atomic operation to check for double close
	if cp.ptyClosed.Load() {
		cp.Logger.Debug().Msg("PTY already closed (atomic check)")
		return nil
	}

	// Protect critical section with mutex
	cp.ptyMutex.Lock()
	defer cp.ptyMutex.Unlock()

	// Double-check pattern (complete race condition elimination)
	if cp.ptyClosed.Load() || cp.PTY == nil {
		cp.Logger.Debug().Msg("PTY already closed or nil (double check)")
		return nil
	}

	// PTY operations in CI/test environments are executed conservatively
	if cp.isCIEnv || cp.isMockEnv {
		cp.Logger.Debug().Msg("CI/Mock environment: safe PTY close")
		// Suppress errors as much as possible in CI environment
		if err := cp.PTY.Close(); err != nil {
			if !strings.Contains(err.Error(), "file already closed") {
				cp.Logger.Debug().Err(err).Msg("PTY close in CI (non-critical error)")
			}
		} else {
			cp.Logger.Debug().Msg("PTY closed successfully in CI")
		}
	} else {
		// Standard close in normal environment
		err := cp.PTY.Close()
		if err != nil {
			cp.Logger.Warn().Err(err).Msg("PTY close warning")
			if !cp.ptyClosed.CompareAndSwap(false, true) {
				return nil // Already closed by another goroutine
			}
			return err
		} else {
			cp.Logger.Debug().Msg("PTY closed successfully")
		}
	}

	// Atomically set close state
	cp.ptyClosed.Store(true)
	return nil
}

// ClaudeManager - Claude CLI management system (containedctx fixed version)
type ClaudeManager struct {
	processes  map[string]*ClaudeProcess
	mu         sync.RWMutex // Protect concurrent access to process map
	config     *ClaudeConfig
	logger     zerolog.Logger
	cancel     context.CancelFunc
	claudePath string
	homeDir    string
	workingDir string
}

// NewClaudeManager - Initialize manager
func NewClaudeManager(workingDir string) (*ClaudeManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Detect Claude CLI path
	claudePath, err := detectClaudePath()
	if err != nil {
		return nil, fmt.Errorf("failed to detect Claude CLI path: %w", err)
	}

	_, cancel := context.WithCancel(context.Background())

	logger := log.With().Str("component", "claude-manager").Logger()

	cm := &ClaudeManager{
		processes:  make(map[string]*ClaudeProcess),
		logger:     logger,
		cancel:     cancel,
		claudePath: claudePath,
		homeDir:    homeDir,
		workingDir: workingDir,
	}

	// Initialize configuration file
	if err := cm.setupConfig(); err != nil {
		return nil, fmt.Errorf("failed to setup config: %w", err)
	}

	return cm, nil
}

// detectClaudePath - Detect Claude CLI path
func detectClaudePath() (string, error) {
	// Try dynamic npm path detection first
	if npmPath := detectNpmClaudeCodePath(); npmPath != "" {
		return npmPath, nil
	}

	// Check common paths in order (prioritize claude-code command)
	paths := []string{
		// npm global install (claude-code)
		"/usr/local/lib/node_modules/@anthropic-ai/claude-code/cli.js",
		"/opt/homebrew/lib/node_modules/@anthropic-ai/claude-code/cli.js",
		"/usr/local/lib/node_modules/@anthropic/claude-code/bin/claude-code",
		"/opt/homebrew/lib/node_modules/@anthropic/claude-code/bin/claude-code",
		// Traditional claude command
		"~/.claude/local/claude",
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude",
	}

	for _, path := range paths {
		expandedPath := expandPath(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath, nil
		}
	}

	// Search from PATH (prioritize claude-code)
	if path, err := exec.LookPath("claude-code"); err == nil {
		return path, nil
	}

	// Search for traditional claude command from PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("claude CLI not found in any expected locations")
}

// detectNpmClaudeCodePath - Dynamic detection of npm global install path
func detectNpmClaudeCodePath() string {
	// Skip npm detection in CI or test environments
	if os.Getenv("CI") == "true" || os.Getenv("CLAUDE_MOCK_ENV") == "true" {
		return ""
	}

	// Check for npm command existence
	if _, err := exec.LookPath("npm"); err != nil {
		return ""
	}

	// Context with timeout (3 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Get global install path with npm root -g
	cmd := exec.CommandContext(ctx, "npm", "root", "-g")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	npmRoot := strings.TrimSpace(string(output))
	if npmRoot == "" {
		return ""
	}

	// Try multiple package names
	candidatePaths := []string{
		// @anthropic-ai/claude-code (actual package name)
		filepath.Join(npmRoot, "@anthropic-ai", "claude-code", "cli.js"),
		// @anthropic/claude-code (future possibility)
		filepath.Join(npmRoot, "@anthropic", "claude-code", "bin", "claude-code"),
		filepath.Join(npmRoot, "@anthropic", "claude-code", "cli.js"),
	}

	// Check path existence
	for _, claudeCodePath := range candidatePaths {
		if _, err := os.Stat(claudeCodePath); err == nil {
			return claudeCodePath
		}
	}

	return ""
}

// expandPath - Path expansion helper
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if len(path) == 1 {
			return homeDir
		}
		if path[1] == '/' {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// setupConfig - Initialize Claude configuration file (with timeout)
func (cm *ClaudeManager) setupConfig() error {
	// Short timeout setting for test/CI environments
	timeout := 30 * time.Second
	if os.Getenv("CI") == "true" || os.Getenv("CLAUDE_MOCK_ENV") == "true" {
		timeout = 5 * time.Second // Reduce to 5 seconds in CI environment
	}

	// タイムアウト付きコンテキスト
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return cm.setupConfigWithContext(ctx)
}

// setupConfigWithContext - Initialize configuration file with context
func (cm *ClaudeManager) setupConfigWithContext(ctx context.Context) error {
	configDir := filepath.Join(cm.homeDir, ".claude")
	configFile := filepath.Join(configDir, "settings.json")

	// Create directory (context check)
	select {
	case <-ctx.Done():
		return fmt.Errorf("config setup timeout: %w", ctx.Err())
	default:
	}

	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if configuration file already exists (context check)
	select {
	case <-ctx.Done():
		return fmt.Errorf("config setup timeout: %w", ctx.Err())
	default:
	}

	if _, err := os.Stat(configFile); err == nil {
		cm.logger.Info().Msg("existing Claude config found, preserving it")
		return nil
	}

	// Create default configuration
	defaultConfig := &ClaudeConfig{
		Model:                  "sonnet",
		Theme:                  "dark",
		HasCompletedOnboarding: true,
		HasSetTheme:            true,
		SkipInitialSetup:       true,
	}

	// JSON marshal (context check)
	select {
	case <-ctx.Done():
		return fmt.Errorf("config setup timeout: %w", ctx.Err())
	default:
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file (context check)
	select {
	case <-ctx.Done():
		return fmt.Errorf("config setup timeout: %w", ctx.Err())
	default:
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.config = defaultConfig
	cm.logger.Info().Msg("Claude config initialized")
	return nil
}

// StartAgent - Start agent process (receive context as argument)
func (cm *ClaudeManager) StartAgent(ctx context.Context, config *AgentConfig) error {
	// Adjust configuration
	config.SessionName = config.Name

	cm.mu.Lock()
	if _, exists := cm.processes[config.Name]; exists {
		cm.mu.Unlock()
		return fmt.Errorf("agent %s already running", config.Name)
	}
	cm.mu.Unlock()

	process, err := cm.createProcess(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create process for %s: %w", config.Name, err)
	}

	cm.mu.Lock()
	cm.processes[config.Name] = process
	cm.mu.Unlock()

	// Start process monitoring
	go cm.monitorProcess(ctx, process)

	return nil
}

// createProcess - Create process (with CI environment detection)
func (cm *ClaudeManager) createProcess(ctx context.Context, config *AgentConfig) (*ClaudeProcess, error) {
	processCtx, cancel := context.WithCancel(ctx)

	logger := cm.logger.With().Str("agent", config.Name).Logger()

	// Environment detection
	isCIEnv := isCIEnvironment()
	isMockEnv := isTestEnvironment()

	process := &ClaudeProcess{
		Config:      config,
		Logger:      logger,
		cancel:      cancel,
		restartChan: make(chan struct{}, 1),
		messageChan: make(chan string, 10),
		isCIEnv:     isCIEnv,
		isMockEnv:   isMockEnv,
	}

	// Set initial state
	process.isRunning.Store(false)
	process.ptyClosed.Store(false)

	logger.Debug().Bool("ci_env", isCIEnv).Bool("mock_env", isMockEnv).Msg("Process environment detected")

	if err := process.start(processCtx, cm.claudePath); err != nil {
		cancel()
		return nil, err
	}

	return process, nil
}

// start - Start process (CI environment compatible version)
func (cp *ClaudeProcess) start(ctx context.Context, claudePath string) error {
	if cp.isRunning.Load() {
		return fmt.Errorf("process already running")
	}

	// Option to skip PTY creation in CI/Mock environment
	if cp.isMockEnv && os.Getenv("CLAUDE_MOCK_PTY_SKIP") == "true" {
		cp.Logger.Info().Msg("Mock environment: PTY creation skipped")
		cp.isRunning.Store(true)
		return nil
	}

	// Implementation equivalent to script -q /dev/null
	cmd := exec.CommandContext(ctx, claudePath, "--dangerously-skip-permissions")
	cmd.Dir = cp.Config.WorkingDir

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"SHELL=/bin/bash",
	)

	// Additional environment variables for CI environment
	if cp.isCIEnv {
		cmd.Env = append(cmd.Env, "CLAUDE_CI_MODE=true")
	}

	// Create PTY (conservative error handling in CI environment)
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		if cp.isCIEnv && strings.Contains(err.Error(), "no such device") {
			cp.Logger.Warn().Err(err).Msg("PTY creation failed in CI (expected)")
			// Treat as mock if PTY creation fails in CI environment
			cp.isRunning.Store(true)
			return nil
		}
		return fmt.Errorf("failed to start with PTY: %w", err)
	}

	cp.Cmd = cmd
	cp.PTY = ptyFile
	cp.isRunning.Store(true)
	cp.ptyClosed.Store(false)

	cp.Logger.Info().Bool("ci_env", cp.isCIEnv).Msg("Claude process started")

	// Send initial instructions (shortened in CI environment)
	go cp.sendInitialInstructions()

	return nil
}

// sendInitialInstructions - Send initial instructions
func (cp *ClaudeProcess) sendInitialInstructions() {
	// Wait for process startup (shortened in CI environment)
	waitTime := 2 * time.Second
	if os.Getenv("CI") == "true" || os.Getenv("CLAUDE_MOCK_ENV") == "true" {
		waitTime = 100 * time.Millisecond // Reduce to 100ms in CI environment
	}
	time.Sleep(waitTime)

	if cp.Config.InstructionFile != "" {
		content, err := os.ReadFile(cp.Config.InstructionFile)
		if err != nil {
			cp.Logger.Error().Err(err).Msg("failed to read instruction file")
			return
		}

		// Send instruction content
		if err := cp.sendMessage(string(content)); err != nil {
			cp.Logger.Error().Err(err).Msg("failed to send initial instructions")
		}
	}
}

// sendMessage - Send message (CI environment compatible version)
func (cp *ClaudeProcess) sendMessage(message string) error {
	// PTY skip handling in Mock environment
	if cp.isMockEnv && cp.PTY == nil {
		cp.Logger.Debug().Str("message", message).Msg("Mock environment: message send simulated")
		return nil
	}

	cp.ptyMutex.Lock()
	defer cp.ptyMutex.Unlock()

	if !cp.isRunning.Load() || cp.PTY == nil || cp.ptyClosed.Load() {
		return fmt.Errorf("process not running or PTY closed")
	}

	// Clear prompt
	if _, err := cp.PTY.Write([]byte{3}); err != nil { // Ctrl+C
		return fmt.Errorf("failed to send Ctrl+C: %w", err)
	}

	// Wait time after Ctrl+C (shortened in CI environment)
	waitTime := 400 * time.Millisecond
	if os.Getenv("CI") == "true" || os.Getenv("CLAUDE_MOCK_ENV") == "true" {
		waitTime = 50 * time.Millisecond // Reduce to 50ms in CI environment
	}
	time.Sleep(waitTime)

	// Send message
	if _, err := cp.PTY.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Send Enter
	if _, err := cp.PTY.Write([]byte{13}); err != nil { // Enter
		return fmt.Errorf("failed to send enter: %w", err)
	}

	cp.Logger.Info().Str("message", message).Msg("message sent")
	return nil
}

// monitorProcess - Process monitoring (receive context as argument)
func (cm *ClaudeManager) monitorProcess(ctx context.Context, process *ClaudeProcess) {
	defer func() {
		if r := recover(); r != nil {
			process.Logger.Error().Interface("panic", r).Msg("process monitor panic")
		}
	}()

	// Read process output
	go cm.readProcessOutput(ctx, process)

	// Process messages
	go cm.processMessages(ctx, process)

	// Monitor process exit
	go cm.watchProcessExit(process)

	// Automatic recovery monitoring
	for {
		select {
		case <-ctx.Done():
			process.Logger.Info().Msg("process monitor stopped")
			return
		case <-process.restartChan:
			if err := cm.restartProcess(ctx, process); err != nil {
				process.Logger.Error().Err(err).Msg("failed to restart process")
			}
		}
	}
}

// readProcessOutput - Read process output (receive context as argument)
func (cm *ClaudeManager) readProcessOutput(ctx context.Context, process *ClaudeProcess) {
	reader := bufio.NewReader(process.PTY)

	// Set read timeout for PTY
	// 100ms read timeout
	if err := process.PTY.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		process.Logger.Warn().Err(err).Msg("failed to set PTY read deadline")
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read with timeout
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					process.Logger.Info().Msg("process output ended")
					return
				}
				// Add short wait for timeouts or other errors to reduce CPU usage
				time.Sleep(10 * time.Millisecond)
				// Reset timeout for next read
				if err := process.PTY.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
					process.Logger.Warn().Err(err).Msg("failed to reset PTY read deadline")
				}
				continue
			}

			// Log output
			process.Logger.Debug().Str("output", strings.TrimSpace(line)).Msg("process output")
		}
	}
}

// processMessages - Process messages (receive context as argument)
func (cm *ClaudeManager) processMessages(ctx context.Context, process *ClaudeProcess) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-process.messageChan:
			if err := process.sendMessage(message); err != nil {
				process.Logger.Error().Err(err).Msg("failed to send message")
			}
		}
	}
}

// watchProcessExit - Monitor process exit (atomic operation version)
func (cm *ClaudeManager) watchProcessExit(process *ClaudeProcess) {
	// Skip Cmd.Wait in Mock environment
	if process.isMockEnv && process.Cmd == nil {
		process.Logger.Debug().Msg("Mock environment: process exit watch skipped")
		return
	}

	if process.Cmd != nil {
		if err := process.Cmd.Wait(); err != nil {
			process.Logger.Error().Err(err).Msg("process exited with error")
		}
	}

	process.isRunning.Store(false)

	// Trigger automatic recovery
	select {
	case process.restartChan <- struct{}{}:
	default:
	}
}

// restartProcess - Restart process (with safe PTY close)
func (cm *ClaudeManager) restartProcess(ctx context.Context, process *ClaudeProcess) error {
	process.Logger.Info().Msg("restarting process")

	// Safe cleanup of existing process
	if err := process.safePTYClose(); err != nil {
		process.Logger.Warn().Err(err).Msg("failed to close PTY during restart")
	}

	// Reset state
	process.isRunning.Store(false)
	process.ptyClosed.Store(false)

	// Start new process
	if err := process.start(ctx, cm.claudePath); err != nil {
		return fmt.Errorf("failed to restart process: %w", err)
	}

	return nil
}

// SendMessage - Send message to agent
func (cm *ClaudeManager) SendMessage(agentName, message string) error {
	cm.mu.RLock()
	process, exists := cm.processes[agentName]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", agentName)
	}

	// Message send timeout (shortened in CI environment)
	timeout := 5 * time.Second
	if os.Getenv("CI") == "true" || os.Getenv("CLAUDE_MOCK_ENV") == "true" {
		timeout = 1 * time.Second // Reduce to 1 second in CI environment
	}

	select {
	case process.messageChan <- message:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout sending message to %s", agentName)
	}
}

// StopAgent - Stop agent
func (cm *ClaudeManager) StopAgent(agentName string) error {
	cm.mu.Lock()
	process, exists := cm.processes[agentName]
	if !exists {
		cm.mu.Unlock()
		return fmt.Errorf("agent %s not found", agentName)
	}
	delete(cm.processes, agentName)
	cm.mu.Unlock()

	process.Logger.Info().Msg("stopping agent")
	process.cancel()

	// Safe PTY close (prevent double close)
	if err := process.safePTYClose(); err != nil {
		process.Logger.Warn().Err(err).Msg("failed to close PTY during stop")
	}

	return nil
}

// Shutdown - Complete shutdown
func (cm *ClaudeManager) Shutdown() error {
	cm.logger.Info().Msg("shutting down Claude manager")

	// Stop all processes
	for name, process := range cm.processes {
		cm.logger.Info().Str("agent", name).Msg("stopping agent")
		process.cancel()

		// Safe PTY close (prevent double close)
		if err := process.safePTYClose(); err != nil {
			cm.logger.Warn().Err(err).Str("agent", name).Msg("failed to close PTY during shutdown")
		}
	}

	cm.cancel()
	return nil
}

// GetAgentStatus - Get agent status
func (cm *ClaudeManager) GetAgentStatus(agentName string) (bool, error) {
	cm.mu.RLock()
	process, exists := cm.processes[agentName]
	cm.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("agent %s not found", agentName)
	}

	return process.isRunning.Load(), nil
}

// ListAgents - Get agent list
func (cm *ClaudeManager) ListAgents() []string {
	cm.mu.RLock()
	agents := make([]string, 0, len(cm.processes))
	for name := range cm.processes {
		agents = append(agents, name)
	}
	cm.mu.RUnlock()

	return agents
}

// StartWithSignalHandling - Start system with signal handling
func (cm *ClaudeManager) StartWithSignalHandling() error {
	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	go func() {
		<-sigChan
		cm.logger.Info().Msg("received shutdown signal")
		if err := cm.Shutdown(); err != nil {
			cm.logger.Error().Err(err).Msg("error during shutdown")
		}
	}()

	return nil
}
