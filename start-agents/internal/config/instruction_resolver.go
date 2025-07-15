package config

import (
	"fmt"
	"path/filepath"
	"sync"
)

// InstructionResolverInterface defines instruction resolver interface
type InstructionResolverInterface interface {
	// ResolveInstructionPath resolves instruction file path for specified role
	ResolveInstructionPath(role string) (string, error)

	// ResolveInstructionPathWithEnv resolves instruction file path with environment
	ResolveInstructionPathWithEnv(role, environment string) (string, error)

	// GetAvailableRoles gets list of available roles
	GetAvailableRoles() []string

	// ValidateInstructionPaths validates all instruction paths
	ValidateInstructionPaths() *ValidationResult
}

// InstructionResolver resolves instruction file paths
type InstructionResolver struct {
	config       *TeamConfig
	pathResolver PathResolverInterface
	validator    InstructionValidatorInterface
	cache        map[string]string
	cacheMutex   sync.RWMutex
}

// NewInstructionResolver creates a new instruction resolver
func NewInstructionResolver(config *TeamConfig) *InstructionResolver {
	return &InstructionResolver{
		config:       config,
		pathResolver: NewPathResolver(config.InstructionsDir),
		validator:    NewInstructionValidator(config.StrictValidation),
		cache:        make(map[string]string),
	}
}

// ResolveInstructionPath resolves instruction file path for role
func (ir *InstructionResolver) ResolveInstructionPath(role string) (string, error) {
	return ir.ResolveInstructionPathWithEnv(role, ir.config.Environment)
}

// ResolveInstructionPathWithEnv 環境指定でinstructionファイルパスを解決
func (ir *InstructionResolver) ResolveInstructionPathWithEnv(role, environment string) (string, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s:%s", role, environment)
	if cached := ir.getCachedPath(cacheKey); cached != "" {
		return cached, nil
	}

	// Resolution order: environment config -> base config -> legacy config -> default -> fallback
	resolvedPath, err := ir.resolvePathWithFallback(role, environment)
	if err != nil {
		return "", err
	}

	// Save to cache
	ir.setCachedPath(cacheKey, resolvedPath)

	return resolvedPath, nil
}

// resolvePathWithFallback resolves path with fallback (5 stages)
func (ir *InstructionResolver) resolvePathWithFallback(role, environment string) (string, error) {
	// 1. Check environment-specific config
	if envPath := ir.getEnvironmentSpecificPath(role, environment); envPath != "" {
		if resolved, err := ir.pathResolver.ResolvePath(envPath); err == nil {
			if ir.validator.ValidateFileExists(resolved) {
				return resolved, nil
			}
		}
	}

	// 2. Check base config
	if basePath := ir.getBasePath(role); basePath != "" {
		if resolved, err := ir.pathResolver.ResolvePath(basePath); err == nil {
			if ir.validator.ValidateFileExists(resolved) {
				return resolved, nil
			}
		}
	}

	// 3. Check legacy config (backward compatibility)
	if legacyPath := ir.getLegacyPath(role); legacyPath != "" {
		fullPath := filepath.Join(ir.config.InstructionsDir, legacyPath)
		if resolved, err := ir.pathResolver.ResolvePath(fullPath); err == nil {
			if ir.validator.ValidateFileExists(resolved) {
				return resolved, nil
			}
		}
	}

	// 4. Use default config
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

	// 5. Final fallback: return path even if file doesn't exist
	return ir.pathResolver.ResolvePath(fullPath)
}

// getEnvironmentSpecificPath gets environment-specific path
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

// getBasePath gets base path
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

// getLegacyPath gets legacy configuration path
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

// getDefaultPath gets default path
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

// getCachedPath gets path from cache
func (ir *InstructionResolver) getCachedPath(key string) string {
	ir.cacheMutex.RLock()
	defer ir.cacheMutex.RUnlock()
	return ir.cache[key]
}

// setCachedPath saves path to cache
func (ir *InstructionResolver) setCachedPath(key, path string) {
	ir.cacheMutex.Lock()
	defer ir.cacheMutex.Unlock()
	ir.cache[key] = path
}

// ClearCache clears the cache
func (ir *InstructionResolver) ClearCache() {
	ir.cacheMutex.Lock()
	defer ir.cacheMutex.Unlock()
	ir.cache = make(map[string]string)
}
