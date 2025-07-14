package testutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/stretchr/testify/mock"
)

// TestHelper テストヘルパー構造体
type TestHelper struct {
	TempDir     string
	MockFS      *MockFileSystem
	MockTmux    *MockTmuxManager
	MockProcess *MockProcessManager
	Config      *config.TeamConfig
}

// SetupTest テスト環境のセットアップ
func SetupTest(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "claude-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	helper := &TestHelper{
		TempDir:     tempDir,
		MockFS:      NewMockFileSystem(),
		MockTmux:    NewMockTmuxManager(),
		MockProcess: NewMockProcessManager(),
	}

	// テスト用設定の初期化
	helper.Config = CreateTestConfig()

	return helper
}

// CleanupTest テスト環境のクリーンアップ
func (h *TestHelper) CleanupTest(t *testing.T) {
	if h.TempDir != "" {
		if err := os.RemoveAll(h.TempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir %s: %v", h.TempDir, err)
		}
	}
	if h.MockFS != nil {
		h.MockFS.Reset()
	}
	if h.MockTmux != nil {
		h.MockTmux.Reset()
	}
	if h.MockProcess != nil {
		h.MockProcess.Reset()
	}
}

// CreateTestConfig テスト用設定の生成
func CreateTestConfig(options ...ConfigOption) *config.TeamConfig {
	cfg := &config.TeamConfig{
		ClaudeCLIPath:          "/mock/claude",
		InstructionsDir:        "/mock/instructions",
		WorkingDir:             "/mock/work",
		ConfigDir:              "/mock/config",
		LogFile:                "/mock/logs/test.log",
		AuthBackupDir:          "/mock/auth_backup",
		MaxProcesses:           4,
		MaxMemoryMB:            1024,
		MaxCPUPercent:          80.0,
		LogLevel:               "info",
		HealthCheckInterval:    30 * time.Second,
		MaxRestartAttempts:     3,
		SessionName:            "test-session",
		DefaultLayout:          "integrated",
		AutoAttach:             false,
		PaneCount:              6,
		AuthCheckInterval:      30 * time.Minute,
		IDEBackupEnabled:       true,
		StartupTimeout:         10 * time.Second,
		ShutdownTimeout:        15 * time.Second,
		RestartDelay:           5 * time.Second,
		ProcessTimeout:         30 * time.Second,
		SendCommand:            "send-agent",
		BinaryName:             "claude-code-agents",
		DevCount:               4,
		POInstructionFile:      "po.md",
		ManagerInstructionFile: "manager.md",
		DevInstructionFile:     "developer.md",
	}

	for _, opt := range options {
		opt(cfg)
	}

	return cfg
}

// ConfigOption 設定オプション
type ConfigOption func(*config.TeamConfig)

// WithDevCount 開発者数設定
func WithDevCount(count int) ConfigOption {
	return func(cfg *config.TeamConfig) {
		cfg.DevCount = count
		cfg.PaneCount = 2 + count // PO + Manager + Dev数
	}
}

// WithSessionName セッション名設定
func WithSessionName(name string) ConfigOption {
	return func(cfg *config.TeamConfig) {
		cfg.SessionName = name
	}
}

// WithLogLevel ログレベル設定
func WithLogLevel(level string) ConfigOption {
	return func(cfg *config.TeamConfig) {
		cfg.LogLevel = level
	}
}

// MockFileSystem ファイルシステム操作のモック
type MockFileSystem struct {
	Files        map[string]*MockFile
	Directories  map[string]*MockDirectory
	Permissions  map[string]os.FileMode
	ShouldFailOn map[string]error
	ReadCallLog  []string
	WriteCallLog []string
}

// MockFile ファイルのモック
type MockFile struct {
	Path    string
	Content []byte
	Mode    os.FileMode
	ModTime time.Time
	Exists  bool
}

// MockDirectory ディレクトリのモック
type MockDirectory struct {
	Path     string
	Mode     os.FileMode
	ModTime  time.Time
	Exists   bool
	Children []string
}

// NewMockFileSystem 新しいモックファイルシステムを作成
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:        make(map[string]*MockFile),
		Directories:  make(map[string]*MockDirectory),
		Permissions:  make(map[string]os.FileMode),
		ShouldFailOn: make(map[string]error),
		ReadCallLog:  make([]string, 0),
		WriteCallLog: make([]string, 0),
	}
}

// Reset モックファイルシステムをリセット
func (mfs *MockFileSystem) Reset() {
	mfs.Files = make(map[string]*MockFile)
	mfs.Directories = make(map[string]*MockDirectory)
	mfs.Permissions = make(map[string]os.FileMode)
	mfs.ShouldFailOn = make(map[string]error)
	mfs.ReadCallLog = make([]string, 0)
	mfs.WriteCallLog = make([]string, 0)
}

// MockTmuxManager tmux操作のモック
type MockTmuxManager struct {
	mock.Mock
	Sessions       map[string]*MockSession
	CallLog        []string
	ShouldFailOn   map[string]error
	CommandResults map[string]string
}

// MockSession tmuxセッションのモック
type MockSession struct {
	Name   string
	Panes  []*MockPane
	Active bool
	Layout string
}

// MockPane tmuxペインのモック
type MockPane struct {
	ID      string
	Command string
	Running bool
	Output  []string
}

// NewMockTmuxManager 新しいモックtmuxマネージャーを作成
func NewMockTmuxManager() *MockTmuxManager {
	return &MockTmuxManager{
		Sessions:       make(map[string]*MockSession),
		CallLog:        make([]string, 0),
		ShouldFailOn:   make(map[string]error),
		CommandResults: make(map[string]string),
	}
}

// Reset モックtmuxマネージャーをリセット
func (mtm *MockTmuxManager) Reset() {
	mtm.Mock = mock.Mock{}
	mtm.Sessions = make(map[string]*MockSession)
	mtm.CallLog = make([]string, 0)
	mtm.ShouldFailOn = make(map[string]error)
	mtm.CommandResults = make(map[string]string)
}

// SessionExists セッション存在確認のモック
func (mtm *MockTmuxManager) SessionExists(sessionName string) bool {
	args := mtm.Called(sessionName)
	if args.Bool(0) {
		return true
	}
	_, exists := mtm.Sessions[sessionName]
	return exists
}

// GetPaneCount ペイン数取得のモック
func (mtm *MockTmuxManager) GetPaneCount(sessionName string) (int, error) {
	args := mtm.Called(sessionName)
	if session, exists := mtm.Sessions[sessionName]; exists {
		return len(session.Panes), args.Error(1)
	}
	return args.Int(0), args.Error(1)
}

// GetPaneList ペインリスト取得のモック
func (mtm *MockTmuxManager) GetPaneList(sessionName string) ([]string, error) {
	args := mtm.Called(sessionName)
	result := args.Get(0)
	if result != nil {
		return result.([]string), args.Error(1)
	}

	if session, exists := mtm.Sessions[sessionName]; exists {
		panes := make([]string, len(session.Panes))
		for i, pane := range session.Panes {
			panes[i] = pane.ID
		}
		return panes, args.Error(1)
	}
	return []string{}, args.Error(1)
}

// SendKeysToPane ペインにキーを送信
func (mtm *MockTmuxManager) SendKeysToPane(sessionName, pane, keys string) error {
	args := mtm.Called(sessionName, pane, keys)
	mtm.CallLog = append(mtm.CallLog, "SendKeysToPane:"+sessionName+":"+pane+":"+keys)
	return args.Error(0)
}

// SendKeysWithEnter キー送信のモック
func (mtm *MockTmuxManager) SendKeysWithEnter(sessionName, paneIndex, message string) error {
	args := mtm.Called(sessionName, paneIndex, message)
	mtm.CallLog = append(mtm.CallLog, "SendKeysWithEnter:"+sessionName+":"+paneIndex+":"+message)
	return args.Error(0)
}

// DetectActiveAISession アクティブAIセッション検出のモック
func (mtm *MockTmuxManager) DetectActiveAISession(expectedPanes int) (string, string, error) {
	args := mtm.Called(expectedPanes)
	return args.String(0), args.String(1), args.Error(2)
}

// ListSessions セッション一覧の取得
func (mtm *MockTmuxManager) ListSessions() ([]string, error) {
	args := mtm.Called()
	return args.Get(0).([]string), args.Error(1)
}

// CreateSession セッションの作成
func (mtm *MockTmuxManager) CreateSession(sessionName string) error {
	args := mtm.Called(sessionName)
	return args.Error(0)
}

// KillSession セッションの削除
func (mtm *MockTmuxManager) KillSession(sessionName string) error {
	args := mtm.Called(sessionName)
	return args.Error(0)
}

// AttachSession セッションへの接続
func (mtm *MockTmuxManager) AttachSession(sessionName string) error {
	args := mtm.Called(sessionName)
	return args.Error(0)
}

// CreateIntegratedLayout 統合監視画面レイアウトの作成
func (mtm *MockTmuxManager) CreateIntegratedLayout(sessionName string, devCount int) error {
	args := mtm.Called(sessionName, devCount)
	return args.Error(0)
}

// CreateIndividualLayout 個別セッション方式の作成
func (mtm *MockTmuxManager) CreateIndividualLayout(sessionName string) error {
	args := mtm.Called(sessionName)
	return args.Error(0)
}

// SplitWindow ウィンドウの分割
func (mtm *MockTmuxManager) SplitWindow(target, direction string) error {
	args := mtm.Called(target, direction)
	return args.Error(0)
}

// RenameWindow ウィンドウ名の変更
func (mtm *MockTmuxManager) RenameWindow(sessionName, windowName string) error {
	args := mtm.Called(sessionName, windowName)
	return args.Error(0)
}

// AdjustPaneSizes ペインサイズの調整
func (mtm *MockTmuxManager) AdjustPaneSizes(sessionName string, devCount int) error {
	args := mtm.Called(sessionName, devCount)
	return args.Error(0)
}

// SetPaneTitles ペインタイトルの設定
func (mtm *MockTmuxManager) SetPaneTitles(sessionName string, devCount int) error {
	args := mtm.Called(sessionName, devCount)
	return args.Error(0)
}

// GetAITeamSessions AIチーム関連セッションの取得
func (mtm *MockTmuxManager) GetAITeamSessions(expectedPaneCount int) (map[string][]string, error) {
	args := mtm.Called(expectedPaneCount)
	return args.Get(0).(map[string][]string), args.Error(1)
}

// FindDefaultAISession デフォルトAIセッションの検出
func (mtm *MockTmuxManager) FindDefaultAISession(expectedPaneCount int) (string, error) {
	args := mtm.Called(expectedPaneCount)
	return args.String(0), args.Error(1)
}

// DeleteAITeamSessions AIチーム関連セッションの削除
func (mtm *MockTmuxManager) DeleteAITeamSessions(sessionName string, devCount int) error {
	args := mtm.Called(sessionName, devCount)
	return args.Error(0)
}

// WaitForPaneReady ペインの準備完了待機
func (mtm *MockTmuxManager) WaitForPaneReady(sessionName, pane string, timeout time.Duration) error {
	args := mtm.Called(sessionName, pane, timeout)
	return args.Error(0)
}

// GetSessionInfo セッション情報の取得
func (mtm *MockTmuxManager) GetSessionInfo(sessionName string) (map[string]interface{}, error) {
	args := mtm.Called(sessionName)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// SendInstructionToPaneWithConfig 設定ファイルを使用してインストラクションファイルを送信
func (mtm *MockTmuxManager) SendInstructionToPaneWithConfig(sessionName, pane, agent, instructionsDir string, config interface{}) error {
	args := mtm.Called(sessionName, pane, agent, instructionsDir, config)
	return args.Error(0)
}

// MockProcessManager プロセス管理のモック
type MockProcessManager struct {
	RunningProcesses map[int]*MockProcess
	CommandHistory   []string
	ExitCodes        map[string]int
	Outputs          map[string]string
}

// MockProcess プロセスのモック
type MockProcess struct {
	PID     int
	Command string
	Args    []string
	Status  string
	Output  string
	Error   error
}

// NewMockProcessManager 新しいモックプロセスマネージャーを作成
func NewMockProcessManager() *MockProcessManager {
	return &MockProcessManager{
		RunningProcesses: make(map[int]*MockProcess),
		CommandHistory:   make([]string, 0),
		ExitCodes:        make(map[string]int),
		Outputs:          make(map[string]string),
	}
}

// Reset モックプロセスマネージャーをリセット
func (mpm *MockProcessManager) Reset() {
	mpm.RunningProcesses = make(map[int]*MockProcess)
	mpm.CommandHistory = make([]string, 0)
	mpm.ExitCodes = make(map[string]int)
	mpm.Outputs = make(map[string]string)
}

// CreateTempFile 一時ファイルの作成
func CreateTempFile(t *testing.T, dir, pattern, content string) string {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Logf("Warning: failed to close temp file: %v", err)
		}
	}()

	if content != "" {
		_, err = file.WriteString(content)
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	return file.Name()
}

// CreateTempDir 一時ディレクトリの作成
func CreateTempDir(t *testing.T, pattern string) string {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

// AssertFileExists ファイル存在確認
func AssertFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", path)
	}
}

// AssertFileNotExists ファイル非存在確認
func AssertFileNotExists(t *testing.T, path string) {
	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file %s to not exist", path)
	}
}

// AssertFileContent ファイル内容確認
func AssertFileContent(t *testing.T, path, expectedContent string) {
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Errorf("Failed to read file %s: %v", path, err)
		return
	}
	if string(content) != expectedContent {
		t.Errorf("File %s content mismatch.\nExpected: %s\nActual: %s",
			path, expectedContent, string(content))
	}
}

// CreateTestInstructionFiles テスト用instructionファイルセットの作成
func CreateTestInstructionFiles(t *testing.T, dir string) map[string]string {
	files := map[string]string{
		"po.md":        "# PO Instructions\nThis is the PO instruction file.",
		"manager.md":   "# Manager Instructions\nThis is the Manager instruction file.",
		"developer.md": "# Developer Instructions\nThis is the Developer instruction file.",
	}

	for filename, content := range files {
		filePath := filepath.Join(dir, filename)
		err := os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create instruction file %s: %v", filename, err)
		}
	}

	return files
}

// WithTimeout テストタイムアウト制御
func WithTimeout(t *testing.T, duration time.Duration, fn func()) {
	done := make(chan bool)
	go func() {
		fn()
		done <- true
	}()

	select {
	case <-done:
		return
	case <-time.After(duration):
		t.Errorf("Test timed out after %v", duration)
	}
}
