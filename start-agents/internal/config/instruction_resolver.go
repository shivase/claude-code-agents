package config

import (
	"fmt"
	"path/filepath"
	"sync"
)

// InstructionResolverInterface instruction解決インターフェース
type InstructionResolverInterface interface {
	// ResolveInstructionPath 指定されたロールのinstructionファイルパスを解決
	ResolveInstructionPath(role string) (string, error)

	// ResolveInstructionPathWithEnv 環境指定でinstructionファイルパスを解決
	ResolveInstructionPathWithEnv(role, environment string) (string, error)

	// GetAvailableRoles 利用可能なロール一覧を取得
	GetAvailableRoles() []string

	// ValidateInstructionPaths すべてのinstructionパスを検証
	ValidateInstructionPaths() *ValidationResult
}

// InstructionResolver instruction解決器
type InstructionResolver struct {
	config       *TeamConfig
	pathResolver PathResolverInterface
	validator    InstructionValidatorInterface
	cache        map[string]string
	cacheMutex   sync.RWMutex
}

// NewInstructionResolver 新しいinstruction解決器を作成
func NewInstructionResolver(config *TeamConfig) *InstructionResolver {
	return &InstructionResolver{
		config:       config,
		pathResolver: NewPathResolver(config.InstructionsDir),
		validator:    NewInstructionValidator(config.StrictValidation),
		cache:        make(map[string]string),
	}
}

// ResolveInstructionPath ロール指定でinstructionファイルパスを解決
func (ir *InstructionResolver) ResolveInstructionPath(role string) (string, error) {
	return ir.ResolveInstructionPathWithEnv(role, ir.config.Environment)
}

// ResolveInstructionPathWithEnv 環境指定でinstructionファイルパスを解決
func (ir *InstructionResolver) ResolveInstructionPathWithEnv(role, environment string) (string, error) {
	// キャッシュチェック
	cacheKey := fmt.Sprintf("%s:%s", role, environment)
	if cached := ir.getCachedPath(cacheKey); cached != "" {
		return cached, nil
	}

	// 解決順序：環境設定 → 基本設定 → 既存設定 → デフォルト → フォールバック
	resolvedPath, err := ir.resolvePathWithFallback(role, environment)
	if err != nil {
		return "", err
	}

	// キャッシュに保存
	ir.setCachedPath(cacheKey, resolvedPath)

	return resolvedPath, nil
}

// resolvePathWithFallback フォールバック付きパス解決（5段階）
func (ir *InstructionResolver) resolvePathWithFallback(role, environment string) (string, error) {
	// 1. 環境別設定をチェック
	if envPath := ir.getEnvironmentSpecificPath(role, environment); envPath != "" {
		if resolved, err := ir.pathResolver.ResolvePath(envPath); err == nil {
			if ir.validator.ValidateFileExists(resolved) {
				return resolved, nil
			}
		}
	}

	// 2. 基本設定をチェック
	if basePath := ir.getBasePath(role); basePath != "" {
		if resolved, err := ir.pathResolver.ResolvePath(basePath); err == nil {
			if ir.validator.ValidateFileExists(resolved) {
				return resolved, nil
			}
		}
	}

	// 3. 既存設定をチェック（後方互換性）
	if legacyPath := ir.getLegacyPath(role); legacyPath != "" {
		fullPath := filepath.Join(ir.config.InstructionsDir, legacyPath)
		if resolved, err := ir.pathResolver.ResolvePath(fullPath); err == nil {
			if ir.validator.ValidateFileExists(resolved) {
				return resolved, nil
			}
		}
	}

	// 4. デフォルト設定を使用
	defaultPath := ir.getDefaultPath(role)
	fallbackDir := ir.config.FallbackInstructionDir
	if fallbackDir == "" {
		fallbackDir = ir.config.InstructionsDir
	}

	fullPath := filepath.Join(fallbackDir, defaultPath)
	if resolved, err := ir.pathResolver.ResolvePath(fullPath); err == nil {
		if ir.validator.ValidateFileExists(resolved) {
			return resolved, nil
		}
	}

	// 5. 最終フォールバック：ファイルが存在しなくてもパスを返す
	return ir.pathResolver.ResolvePath(fullPath)
}

// getEnvironmentSpecificPath 環境別パス取得
func (ir *InstructionResolver) getEnvironmentSpecificPath(role, environment string) string {
	if ir.config.InstructionConfig == nil || environment == "" {
		return ""
	}

	envConfig, exists := ir.config.InstructionConfig.Environments[environment]
	if !exists {
		return ""
	}

	switch role {
	case "po":
		return envConfig.POInstructionPath
	case "manager":
		return envConfig.ManagerInstructionPath
	case "dev", "dev1", "dev2", "dev3", "dev4":
		return envConfig.DevInstructionPath
	default:
		return ""
	}
}

// getBasePath 基本パス取得
func (ir *InstructionResolver) getBasePath(role string) string {
	if ir.config.InstructionConfig == nil {
		return ""
	}

	switch role {
	case "po":
		return ir.config.InstructionConfig.Base.POInstructionPath
	case "manager":
		return ir.config.InstructionConfig.Base.ManagerInstructionPath
	case "dev", "dev1", "dev2", "dev3", "dev4":
		return ir.config.InstructionConfig.Base.DevInstructionPath
	default:
		return ""
	}
}

// getLegacyPath 既存設定パス取得
func (ir *InstructionResolver) getLegacyPath(role string) string {
	switch role {
	case "po":
		return ir.config.POInstructionFile
	case "manager":
		return ir.config.ManagerInstructionFile
	case "dev", "dev1", "dev2", "dev3", "dev4":
		return ir.config.DevInstructionFile
	default:
		return ""
	}
}

// getDefaultPath デフォルトパス取得
func (ir *InstructionResolver) getDefaultPath(role string) string {
	extension := ".md"
	if ir.config.InstructionConfig != nil &&
		ir.config.InstructionConfig.Global.DefaultExtension != "" {
		extension = ir.config.InstructionConfig.Global.DefaultExtension
	}

	switch role {
	case "po":
		return "po" + extension
	case "manager":
		return "manager" + extension
	case "dev", "dev1", "dev2", "dev3", "dev4":
		return "developer" + extension
	default:
		return role + extension
	}
}

// GetAvailableRoles 利用可能なロール一覧を取得
func (ir *InstructionResolver) GetAvailableRoles() []string {
	return []string{"po", "manager", "dev", "dev1", "dev2", "dev3", "dev4"}
}

// ValidateInstructionPaths すべてのinstructionパスを検証
func (ir *InstructionResolver) ValidateInstructionPaths() *ValidationResult {
	return ir.validator.ValidateConfig(ir.config)
}

// getCachedPath キャッシュからパスを取得
func (ir *InstructionResolver) getCachedPath(key string) string {
	ir.cacheMutex.RLock()
	defer ir.cacheMutex.RUnlock()
	return ir.cache[key]
}

// setCachedPath キャッシュにパスを保存
func (ir *InstructionResolver) setCachedPath(key, path string) {
	ir.cacheMutex.Lock()
	defer ir.cacheMutex.Unlock()
	ir.cache[key] = path
}

// ClearCache キャッシュをクリア
func (ir *InstructionResolver) ClearCache() {
	ir.cacheMutex.Lock()
	defer ir.cacheMutex.Unlock()
	ir.cache = make(map[string]string)
}
