package config

// ConfigLoaderInterface 設定読み込みのインターフェース
type ConfigLoaderInterface interface {
	// LoadTeamConfig チーム設定を読み込み
	LoadTeamConfig() (*TeamConfig, error)
	// SaveTeamConfig チーム設定を保存
	SaveTeamConfig(*TeamConfig) error
	// GetTeamConfigPath チーム設定ファイルパスを取得
	GetTeamConfigPath() string
}

// === 動的instruction機能は既に実装済み ===
// インターフェースと構造体はinstruction_resolver.goとinstruction_validator.goに定義されています

// ConfigGeneratorInterface 設定生成のインターフェース
type ConfigGeneratorInterface interface {
	// GenerateConfig 設定ファイルを生成
	GenerateConfig(forceOverwrite bool) error
	// ValidateConfig 設定ファイルを検証
	ValidateConfig() error
}
