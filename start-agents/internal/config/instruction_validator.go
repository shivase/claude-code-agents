package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

// InstructionValidatorInterface バリデーションインターフェース
type InstructionValidatorInterface interface {
	// ValidateConfig 設定全体のバリデーション
	ValidateConfig(config *TeamConfig) *ValidationResult

	// ValidateInstructionPath 個別パスのバリデーション
	ValidateInstructionPath(role, path string) *PathValidationResult

	// ValidateFileExists ファイル存在確認
	ValidateFileExists(path string) bool

	// ValidateFileReadable ファイル読み取り可能確認
	ValidateFileReadable(path string) bool
}

// InstructionValidator instruction設定バリデーター
type InstructionValidator struct {
	strictMode bool
}

// NewInstructionValidator 新しいバリデーターを作成
func NewInstructionValidator(strictMode bool) *InstructionValidator {
	return &InstructionValidator{
		strictMode: strictMode,
	}
}

// ValidationResult バリデーション結果
type ValidationResult struct {
	IsValid  bool
	Errors   []ValidationError
	Warnings []ValidationWarning
	Info     []ValidationInfo
}

// ValidationError バリデーションエラー
type ValidationError struct {
	Field      string
	Path       string
	Message    string
	Code       string
	Suggestion string
}

// ValidationWarning バリデーション警告
type ValidationWarning struct {
	Field      string
	Path       string
	Message    string
	Suggestion string
}

// ValidationInfo バリデーション情報
type ValidationInfo struct {
	Field   string
	Path    string
	Message string
}

// PathValidationResult パスバリデーション結果
type PathValidationResult struct {
	IsValid      bool
	Exists       bool
	Readable     bool
	Error        error
	ResolvedPath string
}

// InstructionError instruction関連エラー
type InstructionError struct {
	Type    InstructionErrorType
	Role    string
	Path    string
	Message string
	Cause   error
}

// InstructionErrorType エラー種別
type InstructionErrorType int

const (
	ErrorTypePathResolution InstructionErrorType = iota
	ErrorTypeFileNotFound
	ErrorTypeFileNotReadable
	ErrorTypeInvalidConfig
	ErrorTypeValidationFailed
	ErrorTypeEnvironmentNotFound
)

func (e *InstructionError) Error() string {
	return fmt.Sprintf("instruction error [%s]: %s (path: %s)",
		e.Role, e.Message, e.Path)
}

func (e *InstructionError) Unwrap() error {
	return e.Cause
}

// ValidateConfig 設定全体のバリデーション
func (iv *InstructionValidator) ValidateConfig(config *TeamConfig) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Info:     []ValidationInfo{},
	}

	// ロール一覧を取得
	roles := []string{"po", "manager", "dev"}

	for _, role := range roles {
		// 各ロールのパスを解決してバリデーション
		if err := iv.validateRoleInstructionPath(role, config, result); err != nil {
			result.IsValid = false
		}
	}

	return result
}

// validateRoleInstructionPath ロール別のinstructionパスバリデーション
func (iv *InstructionValidator) validateRoleInstructionPath(role string, config *TeamConfig, result *ValidationResult) error {
	// 設定からパスを取得
	path := iv.getPathForRole(role, config)
	if path == "" {
		// パスが設定されていない場合は警告
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      role + "_instruction_path",
			Message:    "instruction path not configured",
			Suggestion: "Configure instruction path for " + role + " role",
		})
		return nil
	}

	// パスバリデーション実行
	pathResult := iv.ValidateInstructionPath(role, path)

	switch {
	case pathResult.Error != nil:
		result.Errors = append(result.Errors, ValidationError{
			Field:   role + "_instruction_path",
			Path:    pathResult.ResolvedPath,
			Message: pathResult.Error.Error(),
			Code:    "PATH_RESOLUTION_FAILED",
		})
		return pathResult.Error
	case !pathResult.Exists && iv.strictMode:
		result.Errors = append(result.Errors, ValidationError{
			Field:      role + "_instruction_path",
			Path:       pathResult.ResolvedPath,
			Message:    "instruction file not found",
			Code:       "FILE_NOT_FOUND",
			Suggestion: "Create the instruction file or update the path",
		})
		return fmt.Errorf("instruction file not found: %s", pathResult.ResolvedPath)
	case !pathResult.Exists:
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      role + "_instruction_path",
			Path:       pathResult.ResolvedPath,
			Message:    "instruction file not found",
			Suggestion: "Create the instruction file or update the path",
		})
	case !pathResult.Readable:
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      role + "_instruction_path",
			Path:       pathResult.ResolvedPath,
			Message:    "instruction file not readable",
			Suggestion: "Check file permissions",
		})
	default:
		result.Info = append(result.Info, ValidationInfo{
			Field:   role + "_instruction_path",
			Path:    pathResult.ResolvedPath,
			Message: "instruction file found and accessible",
		})
	}

	return nil
}

// getPathForRole ロールに対応するパスを取得
func (iv *InstructionValidator) getPathForRole(role string, config *TeamConfig) string {
	// 拡張設定から取得を試行
	if config.InstructionConfig != nil {
		switch role {
		case "po":
			if config.InstructionConfig.Base.POInstructionPath != "" {
				return config.InstructionConfig.Base.POInstructionPath
			}
		case "manager":
			if config.InstructionConfig.Base.ManagerInstructionPath != "" {
				return config.InstructionConfig.Base.ManagerInstructionPath
			}
		case "dev":
			if config.InstructionConfig.Base.DevInstructionPath != "" {
				return config.InstructionConfig.Base.DevInstructionPath
			}
		}
	}

	// 既存設定から取得（後方互換性）
	switch role {
	case "po":
		return config.POInstructionFile
	case "manager":
		return config.ManagerInstructionFile
	case "dev":
		return config.DevInstructionFile
	default:
		return ""
	}
}

// ValidateInstructionPath 個別パスのバリデーション
func (iv *InstructionValidator) ValidateInstructionPath(role, path string) *PathValidationResult {
	result := &PathValidationResult{
		IsValid: true,
	}

	// パス解決テスト
	if path == "" {
		// パスが空の場合はスキップ
		result.ResolvedPath = ""
		return result
	}

	resolver := NewPathResolver("")
	resolvedPath, err := resolver.ResolvePath(path)
	if err != nil {
		result.IsValid = false
		result.Error = err
		return result
	}

	result.ResolvedPath = resolvedPath

	// ファイル存在確認
	result.Exists = iv.ValidateFileExists(resolvedPath)

	// 読み取り権限確認
	if result.Exists {
		result.Readable = iv.ValidateFileReadable(resolvedPath)
	}

	return result
}

// ValidateFileExists ファイル存在確認
func (iv *InstructionValidator) ValidateFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ValidateFileReadable ファイル読み取り可能確認
func (iv *InstructionValidator) ValidateFileReadable(path string) bool {
	file, err := os.Open(path) // #nosec G304 - path is validated before this call
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Str("path", path).Msg("Failed to close file during readable check")
		}
	}()

	// 簡単な読み取りテスト
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	if err == nil {
		return true
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	return false
}
