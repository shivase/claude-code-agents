package auth

// AuthProviderInterface Claude認証プロバイダーのインターフェース
type AuthProviderInterface interface {
	// CheckAuth 認証状態を確認
	CheckAuth() error
	// CheckSettingsFile 設定ファイルを確認
	CheckSettingsFile() error
	// IsReady Claude CLIが使用準備完了かチェック
	IsReady() bool
	// GetPath Claude CLIのパスを取得
	GetPath() string
	// ValidateSetup セットアップの包括的な検証
	ValidateSetup() error
}

// PreAuthCheckerInterface 事前認証チェッカーのインターフェース
type PreAuthCheckerInterface interface {
	// CheckAuthenticationBeforeStart 開始前の認証確認
	CheckAuthenticationBeforeStart() error
}
