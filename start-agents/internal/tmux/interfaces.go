package tmux

import "time"

// TmuxManagerInterface tmux操作管理のインターフェース
type TmuxManagerInterface interface {
	// SessionExists セッションの存在確認
	SessionExists(sessionName string) bool
	// ListSessions セッション一覧の取得
	ListSessions() ([]string, error)
	// CreateSession セッションの作成
	CreateSession(sessionName string) error
	// KillSession セッションの削除
	KillSession(sessionName string) error
	// AttachSession セッションへの接続
	AttachSession(sessionName string) error
	// CreateIntegratedLayout 統合監視画面レイアウトの作成
	CreateIntegratedLayout(sessionName string, devCount int) error
	// CreateIndividualLayout 個別セッション方式の作成
	CreateIndividualLayout(sessionName string) error
	// SplitWindow ウィンドウの分割
	SplitWindow(target, direction string) error
	// RenameWindow ウィンドウ名の変更
	RenameWindow(sessionName, windowName string) error
	// AdjustPaneSizes ペインサイズの調整
	AdjustPaneSizes(sessionName string, devCount int) error
	// SetPaneTitles ペインタイトルの設定
	SetPaneTitles(sessionName string, devCount int) error
	// GetPaneCount ペイン数の取得
	GetPaneCount(sessionName string) (int, error)
	// GetPaneList ペイン一覧の取得
	GetPaneList(sessionName string) ([]string, error)
	// SendKeysToPane ペインにキーを送信
	SendKeysToPane(sessionName, pane, keys string) error
	// SendKeysWithEnter ペインにキーを送信（Enter付き）
	SendKeysWithEnter(sessionName, pane, keys string) error
	// GetAITeamSessions AIチーム関連セッションの取得
	GetAITeamSessions(expectedPaneCount int) (map[string][]string, error)
	// FindDefaultAISession デフォルトAIセッションの検出
	FindDefaultAISession(expectedPaneCount int) (string, error)
	// DetectActiveAISession アクティブなAIセッションの検出
	DetectActiveAISession(expectedPaneCount int) (string, string, error)
	// DeleteAITeamSessions AIチーム関連セッションの削除
	DeleteAITeamSessions(sessionName string, devCount int) error
	// WaitForPaneReady ペインの準備完了待機
	WaitForPaneReady(sessionName, pane string, timeout time.Duration) error
	// GetSessionInfo セッション情報の取得
	GetSessionInfo(sessionName string) (map[string]interface{}, error)
	// SendInstructionToPaneWithConfig 設定ファイルを使用してインストラクションファイルを送信
	SendInstructionToPaneWithConfig(sessionName, pane, agent, instructionsDir string, config interface{}) error
}
