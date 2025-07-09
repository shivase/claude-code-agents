package main

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
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ClaudeConfig - Claude CLI設定
type ClaudeConfig struct {
	Model                  string `json:"model"`
	Theme                  string `json:"theme"`
	HasCompletedOnboarding bool   `json:"hasCompletedOnboarding"`
	HasSetTheme            bool   `json:"hasSetTheme"`
	SkipInitialSetup       bool   `json:"skipInitialSetup"`
}

// AgentConfig - エージェント設定
type AgentConfig struct {
	Name            string
	InstructionFile string
	SessionName     string
	WorkingDir      string
}

// ClaudeProcess - Claude CLIプロセス管理
type ClaudeProcess struct {
	Config      *AgentConfig
	Cmd         *exec.Cmd
	PTY         *os.File
	Logger      zerolog.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	isRunning   bool
	restartChan chan struct{}
	messageChan chan string
}

// ClaudeManager - Claude CLI管理システム
type ClaudeManager struct {
	processes  map[string]*ClaudeProcess
	config     *ClaudeConfig
	logger     zerolog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	claudePath string
	homeDir    string
	workingDir string
}

// NewClaudeManager - マネージャー初期化
func NewClaudeManager(workingDir string) (*ClaudeManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Claude CLIパスの検出
	claudePath, err := detectClaudePath()
	if err != nil {
		return nil, fmt.Errorf("failed to detect Claude CLI path: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := log.With().Str("component", "claude-manager").Logger()

	cm := &ClaudeManager{
		processes:  make(map[string]*ClaudeProcess),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		claudePath: claudePath,
		homeDir:    homeDir,
		workingDir: workingDir,
	}

	// 設定ファイルの初期化
	if err := cm.setupConfig(); err != nil {
		return nil, fmt.Errorf("failed to setup config: %v", err)
	}

	return cm, nil
}

// detectClaudePath - Claude CLIパスの検出
func detectClaudePath() (string, error) {
	// 一般的なパスを順番に確認
	paths := []string{
		"~/.claude/local/claude",
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// PATHから検索
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("claude CLI not found in any expected locations")
}

// setupConfig - Claude設定ファイルの初期化
func (cm *ClaudeManager) setupConfig() error {
	configDir := filepath.Join(cm.homeDir, ".claude")
	configFile := filepath.Join(configDir, "settings.json")

	// ディレクトリ作成
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// 既存の設定ファイルがあるかチェック
	if _, err := os.Stat(configFile); err == nil {
		cm.logger.Info().Msg("existing Claude config found, preserving it")
		return nil
	}

	// デフォルト設定を作成
	defaultConfig := &ClaudeConfig{
		Model:                  "sonnet",
		Theme:                  "dark",
		HasCompletedOnboarding: true,
		HasSetTheme:            true,
		SkipInitialSetup:       true,
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	cm.config = defaultConfig
	cm.logger.Info().Msg("Claude config initialized")
	return nil
}

// checkAuth - 認証状態確認
func (cm *ClaudeManager) checkAuth() error {
	configFile := filepath.Join(cm.homeDir, ".claude", "settings.json")

	if _, err := os.Stat(configFile); err != nil {
		return fmt.Errorf("Claude config file not found: %v", err)
	}

	// Claude CLIの動作確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cm.claudePath, "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Claude CLI not responding: %v", err)
	}

	return nil
}

// StartAgent - エージェントプロセス開始
func (cm *ClaudeManager) StartAgent(config *AgentConfig) error {
	// 設定を調整
	config.SessionName = config.Name
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.processes[config.Name]; exists {
		return fmt.Errorf("agent %s already running", config.Name)
	}

	process, err := cm.createProcess(config)
	if err != nil {
		return fmt.Errorf("failed to create process for %s: %v", config.Name, err)
	}

	cm.processes[config.Name] = process

	// プロセス監視開始
	go cm.monitorProcess(process)

	return nil
}

// createProcess - プロセス作成
func (cm *ClaudeManager) createProcess(config *AgentConfig) (*ClaudeProcess, error) {
	ctx, cancel := context.WithCancel(cm.ctx)

	logger := cm.logger.With().Str("agent", config.Name).Logger()

	process := &ClaudeProcess{
		Config:      config,
		Logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		restartChan: make(chan struct{}, 1),
		messageChan: make(chan string, 10),
	}

	if err := process.start(cm.claudePath); err != nil {
		cancel()
		return nil, err
	}

	return process, nil
}

// start - プロセス開始
func (cp *ClaudeProcess) start(claudePath string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.isRunning {
		return fmt.Errorf("process already running")
	}

	// script -q /dev/null 相当の実装
	cmd := exec.CommandContext(cp.ctx, claudePath, "--dangerously-skip-permissions")
	cmd.Dir = cp.Config.WorkingDir

	// 環境変数設定
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"SHELL=/bin/bash",
	)

	// PTYの作成
	pty, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start with PTY: %v", err)
	}

	cp.Cmd = cmd
	cp.PTY = pty
	cp.isRunning = true

	cp.Logger.Info().Msg("Claude process started")

	// 初期指示送信
	go cp.sendInitialInstructions()

	return nil
}

// sendInitialInstructions - 初期指示送信
func (cp *ClaudeProcess) sendInitialInstructions() {
	// プロセス起動を待つ
	time.Sleep(2 * time.Second)

	if cp.Config.InstructionFile != "" {
		content, err := os.ReadFile(cp.Config.InstructionFile)
		if err != nil {
			cp.Logger.Error().Err(err).Msg("failed to read instruction file")
			return
		}

		// 指示内容を送信
		cp.sendMessage(string(content))
	}
}

// sendMessage - メッセージ送信
func (cp *ClaudeProcess) sendMessage(message string) error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if !cp.isRunning || cp.PTY == nil {
		return fmt.Errorf("process not running")
	}

	// プロンプトクリア
	if _, err := cp.PTY.Write([]byte{3}); err != nil { // Ctrl+C
		return fmt.Errorf("failed to send Ctrl+C: %v", err)
	}
	time.Sleep(400 * time.Millisecond)

	// メッセージ送信
	if _, err := cp.PTY.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	// Enter送信
	if _, err := cp.PTY.Write([]byte{13}); err != nil { // Enter
		return fmt.Errorf("failed to send enter: %v", err)
	}

	cp.Logger.Info().Str("message", message).Msg("message sent")
	return nil
}

// monitorProcess - プロセス監視
func (cm *ClaudeManager) monitorProcess(process *ClaudeProcess) {
	defer func() {
		if r := recover(); r != nil {
			process.Logger.Error().Interface("panic", r).Msg("process monitor panic")
		}
	}()

	// プロセス出力読み取り
	go cm.readProcessOutput(process)

	// メッセージ処理
	go cm.processMessages(process)

	// プロセス終了監視
	go cm.watchProcessExit(process)

	// 自動復旧監視
	for {
		select {
		case <-process.ctx.Done():
			process.Logger.Info().Msg("process monitor stopped")
			return
		case <-process.restartChan:
			if err := cm.restartProcess(process); err != nil {
				process.Logger.Error().Err(err).Msg("failed to restart process")
			}
		}
	}
}

// readProcessOutput - プロセス出力読み取り
func (cm *ClaudeManager) readProcessOutput(process *ClaudeProcess) {
	reader := bufio.NewReader(process.PTY)

	for {
		select {
		case <-process.ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					process.Logger.Info().Msg("process output ended")
					return
				}
				process.Logger.Error().Err(err).Msg("failed to read process output")
				continue
			}

			// 出力をログに記録
			process.Logger.Debug().Str("output", strings.TrimSpace(line)).Msg("process output")
		}
	}
}

// processMessages - メッセージ処理
func (cm *ClaudeManager) processMessages(process *ClaudeProcess) {
	for {
		select {
		case <-process.ctx.Done():
			return
		case message := <-process.messageChan:
			if err := process.sendMessage(message); err != nil {
				process.Logger.Error().Err(err).Msg("failed to send message")
			}
		}
	}
}

// watchProcessExit - プロセス終了監視
func (cm *ClaudeManager) watchProcessExit(process *ClaudeProcess) {
	if err := process.Cmd.Wait(); err != nil {
		process.Logger.Error().Err(err).Msg("process exited with error")
	}

	process.mu.Lock()
	process.isRunning = false
	process.mu.Unlock()

	// 自動復旧をトリガー
	select {
	case process.restartChan <- struct{}{}:
	default:
	}
}

// restartProcess - プロセス再起動
func (cm *ClaudeManager) restartProcess(process *ClaudeProcess) error {
	process.Logger.Info().Msg("restarting process")

	// 既存プロセスのクリーンアップ
	if process.PTY != nil {
		process.PTY.Close()
	}

	// 新しいプロセスを開始
	if err := process.start(cm.claudePath); err != nil {
		return fmt.Errorf("failed to restart process: %v", err)
	}

	return nil
}

// SendMessage - エージェントにメッセージ送信
func (cm *ClaudeManager) SendMessage(agentName, message string) error {
	cm.mu.RLock()
	process, exists := cm.processes[agentName]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", agentName)
	}

	select {
	case process.messageChan <- message:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending message to %s", agentName)
	}
}

// StopAgent - エージェント停止
func (cm *ClaudeManager) StopAgent(agentName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	process, exists := cm.processes[agentName]
	if !exists {
		return fmt.Errorf("agent %s not found", agentName)
	}

	process.cancel()
	if process.PTY != nil {
		process.PTY.Close()
	}

	delete(cm.processes, agentName)
	return nil
}

// Shutdown - 全体終了
func (cm *ClaudeManager) Shutdown() error {
	cm.logger.Info().Msg("shutting down Claude manager")

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 全プロセス停止
	for name, process := range cm.processes {
		cm.logger.Info().Str("agent", name).Msg("stopping agent")
		process.cancel()
		if process.PTY != nil {
			process.PTY.Close()
		}
	}

	cm.cancel()
	return nil
}

// GetAgentStatus - エージェント状態取得
func (cm *ClaudeManager) GetAgentStatus(agentName string) (bool, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	process, exists := cm.processes[agentName]
	if !exists {
		return false, fmt.Errorf("agent %s not found", agentName)
	}

	process.mu.RLock()
	isRunning := process.isRunning
	process.mu.RUnlock()

	return isRunning, nil
}

// ListAgents - エージェント一覧取得
func (cm *ClaudeManager) ListAgents() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	agents := make([]string, 0, len(cm.processes))
	for name := range cm.processes {
		agents = append(agents, name)
	}

	return agents
}

// StartWithSignalHandling - シグナルハンドリング付きシステム開始
func (cm *ClaudeManager) StartWithSignalHandling() error {
	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 終了シグナル待機
	go func() {
		<-sigChan
		cm.logger.Info().Msg("received shutdown signal")
		cm.Shutdown()
	}()

	return nil
}
