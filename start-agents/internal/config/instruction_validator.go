package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

// InstructionValidatorInterface defines validation interface
type InstructionValidatorInterface interface {
	// ValidateConfig validates entire configuration
	ValidateConfig(config *TeamConfig) *ValidationResult

	// ValidateInstructionPath validates individual path
	ValidateInstructionPath(role, path string) *PathValidationResult

	// ValidateFileExists checks file existence
	ValidateFileExists(path string) bool

	// ValidateFileReadable checks file readability
	ValidateFileReadable(path string) bool
}

// InstructionValidator validates instruction configuration
type InstructionValidator struct {
	strictMode bool
}

// NewInstructionValidator creates a new validator
func NewInstructionValidator(strictMode bool) *InstructionValidator {
	return &InstructionValidator{
		strictMode: strictMode,
	}
}

// ValidationResult represents validation result
type ValidationResult struct {
	IsValid  bool
	Errors   []ValidationError
	Warnings []ValidationWarning
	Info     []ValidationInfo
}

// ValidationError represents validation error
type ValidationError struct {
	Field      string
	Path       string
	Message    string
	Code       string
	Suggestion string
}

// ValidationWarning represents validation warning
type ValidationWarning struct {
	Field      string
	Path       string
	Message    string
	Suggestion string
}

// ValidationInfo represents validation information
type ValidationInfo struct {
	Field   string
	Path    string
	Message string
}

// PathValidationResult represents path validation result
type PathValidationResult struct {
	IsValid      bool
	Exists       bool
	Readable     bool
	Error        error
	ResolvedPath string
}

// InstructionError represents instruction-related error
type InstructionError struct {
	Type    InstructionErrorType
	Role    string
	Path    string
	Message string
	Cause   error
}

// InstructionErrorType represents error type
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

	// Get role list
	roles := []string{"po", "manager", "dev"}

	for _, role := range roles {
		// Resolve and validate path for each role
		if err := iv.validateRoleInstructionPath(role, config, result); err != nil {
			result.IsValid = false
		}
	}

	return result
}

// validateRoleInstructionPath validates instruction path for role
func (iv *InstructionValidator) validateRoleInstructionPath(role string, config *TeamConfig, result *ValidationResult) error {
	// Get path from configuration
	path := iv.getPathForRole(role, config)
	if path == "" {
		// Warn if path not configured
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:      role + "_instruction_path",
			Message:    "instruction path not configured",
			Suggestion: "Configure instruction path for " + role + " role",
		})
		return nil
	}

	// Execute path validation
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

// getPathForRole gets path for role
func (iv *InstructionValidator) getPathForRole(role string, config *TeamConfig) string {
	// Try to get from extended configuration
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

	// Get from legacy configuration (backward compatibility)
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

	// Test path resolution
	if path == "" {
		// Skip if path is empty
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

	// Check file existence
	result.Exists = iv.ValidateFileExists(resolvedPath)

	// Check read permission
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

	// Simple read test
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
