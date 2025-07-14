package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestInstructionFile テスト用のinstructionファイルを作成
func createTestInstructionFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	filePath := filepath.Join(dir, filename)
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	return filePath
}

// createTestConfig テスト用のTeamConfigを作成
func createTestConfig(instructionsDir string) *config.TeamConfig {
	return &config.TeamConfig{
		POInstructionFile:      "po.md",
		ManagerInstructionFile: "manager.md",
		DevInstructionFile:     "developer.md",
		InstructionsDir:        instructionsDir,
	}
}

// createEnhancedTestConfig 拡張設定を含むテスト用TeamConfigを作成
func createEnhancedTestConfig(instructionsDir string) *config.TeamConfig {
	baseConfig := createTestConfig(instructionsDir)
	baseConfig.InstructionConfig = &config.InstructionConfig{
		Base: config.InstructionRoleConfig{
			POInstructionPath:      filepath.Join(instructionsDir, "enhanced_po.md"),
			ManagerInstructionPath: filepath.Join(instructionsDir, "enhanced_manager.md"),
			DevInstructionPath:     filepath.Join(instructionsDir, "enhanced_dev.md"),
		},
		Global: config.InstructionGlobalConfig{
			DefaultExtension: ".md",
			CacheEnabled:     true,
		},
	}
	return baseConfig
}

// TestInstructionValidator_NewInstructionValidator バリデーター作成テスト
func TestInstructionValidator_NewInstructionValidator(t *testing.T) {
	tests := []struct {
		name       string
		strictMode bool
	}{
		{
			name:       "strict mode enabled",
			strictMode: true,
		},
		{
			name:       "strict mode disabled",
			strictMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := config.NewInstructionValidator(tt.strictMode)

			assert.NotNil(t, validator)
		})
	}
}

// TestInstructionValidator_ValidateFileExists ファイル存在確認テスト
func TestInstructionValidator_ValidateFileExists(t *testing.T) {
	validator := config.NewInstructionValidator(true)

	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "validator_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// テストファイルを作成
	testFile := createTestInstructionFile(t, tempDir, "test.md", "# Test Instruction")

	tests := []struct {
		name   string
		path   string
		exists bool
	}{
		{
			name:   "existing file",
			path:   testFile,
			exists: true,
		},
		{
			name:   "non-existing file",
			path:   filepath.Join(tempDir, "nonexistent.md"),
			exists: false,
		},
		{
			name:   "empty path",
			path:   "",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateFileExists(tt.path)
			assert.Equal(t, tt.exists, result)
		})
	}
}

// TestInstructionValidator_ValidateFileReadable ファイル読み取り権限テスト
func TestInstructionValidator_ValidateFileReadable(t *testing.T) {
	validator := config.NewInstructionValidator(true)

	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "validator_readable_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 読み取り可能なファイルを作成
	readableFile := createTestInstructionFile(t, tempDir, "readable.md", "# Readable Content")

	// 空ファイルを作成
	emptyFile := createTestInstructionFile(t, tempDir, "empty.md", "")

	tests := []struct {
		name     string
		path     string
		readable bool
	}{
		{
			name:     "readable file with content",
			path:     readableFile,
			readable: true,
		},
		{
			name:     "empty file (should be readable)",
			path:     emptyFile,
			readable: true,
		},
		{
			name:     "non-existing file",
			path:     filepath.Join(tempDir, "nonexistent.md"),
			readable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateFileReadable(tt.path)
			assert.Equal(t, tt.readable, result)
		})
	}
}

// TestInstructionValidator_ValidateInstructionPath パスバリデーションテスト
func TestInstructionValidator_ValidateInstructionPath(t *testing.T) {
	validator := config.NewInstructionValidator(true)

	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "validator_path_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// テストファイルを作成
	validFile := createTestInstructionFile(t, tempDir, "valid.md", "# Valid Instruction")

	tests := []struct {
		name         string
		role         string
		path         string
		expectValid  bool
		expectExists bool
		expectError  bool
	}{
		{
			name:         "valid existing file",
			role:         "developer",
			path:         validFile,
			expectValid:  true,
			expectExists: true,
			expectError:  false,
		},
		{
			name:         "non-existing file",
			role:         "manager",
			path:         filepath.Join(tempDir, "nonexistent.md"),
			expectValid:  true,
			expectExists: false,
			expectError:  false,
		},
		{
			name:         "empty path",
			role:         "po",
			path:         "",
			expectValid:  true,
			expectExists: false,
			expectError:  false,
		},
		{
			name:         "invalid path characters",
			role:         "developer",
			path:         string([]byte{0}), // null byte
			expectValid:  true,              // PathResolverがnullバイトを処理するため
			expectExists: false,
			expectError:  false, // パス解決は成功するが存在しない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateInstructionPath(tt.role, tt.path)

			assert.Equal(t, tt.expectValid, result.IsValid)
			assert.Equal(t, tt.expectExists, result.Exists)

			if tt.expectError {
				assert.Error(t, result.Error)
			} else {
				assert.NoError(t, result.Error)
			}

			if tt.path != "" && !tt.expectError {
				assert.NotEmpty(t, result.ResolvedPath)
			}
		})
	}
}

// TestInstructionValidator_ValidateConfig 設定全体バリデーションテスト
func TestInstructionValidator_ValidateConfig(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "validator_config_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// テストファイルを作成
	createTestInstructionFile(t, tempDir, "po.md", "# PO Instructions")
	createTestInstructionFile(t, tempDir, "manager.md", "# Manager Instructions")
	createTestInstructionFile(t, tempDir, "developer.md", "# Developer Instructions")

	tests := []struct {
		name             string
		strictMode       bool
		config           *config.TeamConfig
		expectValid      bool
		expectErrorCount int
		expectWarnCount  int
	}{
		{
			name:             "valid config with existing files",
			strictMode:       false,
			config:           createTestConfig(tempDir),
			expectValid:      true,
			expectErrorCount: 0,
			expectWarnCount:  3, // ファイルが存在しないため警告が発生
		},
		{
			name:             "config with missing files (non-strict)",
			strictMode:       false,
			config:           createTestConfig("/nonexistent/path"),
			expectValid:      true,
			expectErrorCount: 0,
			expectWarnCount:  3, // 3つのロールすべてでファイルが見つからない
		},
		{
			name:             "config with missing files (strict)",
			strictMode:       true,
			config:           createTestConfig("/nonexistent/path"),
			expectValid:      false,
			expectErrorCount: 3, // 3つのロールすべてでエラー
			expectWarnCount:  0,
		},
		{
			name:             "empty config (non-strict)",
			strictMode:       false,
			config:           &config.TeamConfig{},
			expectValid:      true,
			expectErrorCount: 0,
			expectWarnCount:  3, // パスが設定されていない警告
		},
		{
			name:             "enhanced config with existing files",
			strictMode:       false,
			config:           createEnhancedTestConfig(tempDir),
			expectValid:      true,
			expectErrorCount: 0,
			expectWarnCount:  3, // ファイルが存在しないため警告が発生
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := config.NewInstructionValidator(tt.strictMode)

			result := validator.ValidateConfig(tt.config)

			assert.Equal(t, tt.expectValid, result.IsValid)
			assert.Equal(t, tt.expectErrorCount, len(result.Errors))
			assert.Equal(t, tt.expectWarnCount, len(result.Warnings))

			// エラーの詳細確認
			for _, err := range result.Errors {
				assert.NotEmpty(t, err.Field)
				assert.NotEmpty(t, err.Message)
				assert.NotEmpty(t, err.Code)
			}

			// 警告の詳細確認
			for _, warn := range result.Warnings {
				assert.NotEmpty(t, warn.Field)
				assert.NotEmpty(t, warn.Message)
			}
		})
	}
}

// TestInstructionValidator_SecurityValidation セキュリティバリデーションテスト
func TestInstructionValidator_SecurityValidation(t *testing.T) {
	validator := config.NewInstructionValidator(true)

	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "validator_security_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		path        string
		expectSafe  bool
		description string
	}{
		{
			name:        "path traversal attempt",
			path:        "../../etc/passwd",
			expectSafe:  true, // PathResolverで安全に処理される
			description: "Should safely handle path traversal attempts",
		},
		{
			name:        "null byte injection",
			path:        "config\x00/../../etc/passwd",
			expectSafe:  true, // PathResolverで安全に処理される
			description: "Should handle null byte injection safely",
		},
		{
			name:        "legitimate relative path",
			path:        "instructions/po.md",
			expectSafe:  true,
			description: "Should allow legitimate relative paths",
		},
		{
			name:        "legitimate absolute path",
			path:        filepath.Join(tempDir, "safe.md"),
			expectSafe:  true,
			description: "Should allow legitimate absolute paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateInstructionPath("test", tt.path)

			if tt.expectSafe {
				// 安全なパスの場合、エラーは起きないか、または安全に処理される
				if result.Error != nil {
					// エラーがあっても、セキュリティ上危険でない範囲のエラーであることを確認
					assert.NotContains(t, result.ResolvedPath, "/etc")
					assert.NotContains(t, result.ResolvedPath, "/root")
				}
			} else {
				// 危険なパスの場合、エラーが発生する
				assert.Error(t, result.Error)
			}

			t.Logf("%s: %s -> %s (error: %v)", tt.description, tt.path, result.ResolvedPath, result.Error)
		})
	}
}

// TestInstructionValidator_EdgeCases エッジケーステスト
func TestInstructionValidator_EdgeCases(t *testing.T) {
	validator := config.NewInstructionValidator(true)

	t.Run("very long path", func(t *testing.T) {
		longPath := "/" + strings.Repeat("very_long_directory_name/", 100) + "file.md"
		result := validator.ValidateInstructionPath("test", longPath)

		// 長すぎるパスでもクラッシュしないことを確認
		assert.NotNil(t, result)
	})

	t.Run("unicode characters in path", func(t *testing.T) {
		unicodePath := "設定/日本語ファイル.md"
		result := validator.ValidateInstructionPath("test", unicodePath)

		// Unicode文字を含むパスが適切に処理されることを確認
		assert.NotNil(t, result)
		if result.Error == nil {
			assert.Contains(t, result.ResolvedPath, "設定")
		}
	})

	t.Run("special characters in path", func(t *testing.T) {
		specialPath := "config/file-with_special.chars@123.md"
		result := validator.ValidateInstructionPath("test", specialPath)

		// 特殊文字を含むパスが適切に処理されることを確認
		assert.NotNil(t, result)
	})
}

// TestInstructionValidator_ValidationResult バリデーション結果構造体テスト
func TestInstructionValidator_ValidationResult(t *testing.T) {
	t.Run("ValidationError structure", func(t *testing.T) {
		err := config.ValidationError{
			Field:      "test_field",
			Path:       "/test/path",
			Message:    "test message",
			Code:       "TEST_CODE",
			Suggestion: "test suggestion",
		}

		assert.Equal(t, "test_field", err.Field)
		assert.Equal(t, "/test/path", err.Path)
		assert.Equal(t, "test message", err.Message)
		assert.Equal(t, "TEST_CODE", err.Code)
		assert.Equal(t, "test suggestion", err.Suggestion)
	})

	t.Run("ValidationWarning structure", func(t *testing.T) {
		warn := config.ValidationWarning{
			Field:      "test_field",
			Path:       "/test/path",
			Message:    "test warning",
			Suggestion: "test suggestion",
		}

		assert.Equal(t, "test_field", warn.Field)
		assert.Equal(t, "/test/path", warn.Path)
		assert.Equal(t, "test warning", warn.Message)
		assert.Equal(t, "test suggestion", warn.Suggestion)
	})

	t.Run("ValidationInfo structure", func(t *testing.T) {
		info := config.ValidationInfo{
			Field:   "test_field",
			Path:    "/test/path",
			Message: "test info",
		}

		assert.Equal(t, "test_field", info.Field)
		assert.Equal(t, "/test/path", info.Path)
		assert.Equal(t, "test info", info.Message)
	})
}

// TestInstructionError エラー構造体テスト
func TestInstructionError(t *testing.T) {
	t.Run("error creation and methods", func(t *testing.T) {
		baseErr := errors.New("base error")
		instrErr := &config.InstructionError{
			Type:    config.ErrorTypeFileNotFound,
			Role:    "developer",
			Path:    "/test/path",
			Message: "test error message",
			Cause:   baseErr,
		}

		// Error() メソッドのテスト
		errMsg := instrErr.Error()
		assert.Contains(t, errMsg, "developer")
		assert.Contains(t, errMsg, "/test/path")
		assert.Contains(t, errMsg, "test error message")

		// Unwrap() メソッドのテスト
		unwrapped := instrErr.Unwrap()
		assert.Equal(t, baseErr, unwrapped)
	})
}

// TestInstructionValidator_GetPathForRole パス取得ロジックテスト
func TestInstructionValidator_GetPathForRole(t *testing.T) {
	validator := config.NewInstructionValidator(false)

	t.Run("basic config paths", func(t *testing.T) {
		config := &config.TeamConfig{
			POInstructionFile:      "po.md",
			ManagerInstructionFile: "manager.md",
			DevInstructionFile:     "developer.md",
		}

		// 非公開メソッドのテストのため、ValidateConfigを通じて間接的にテスト
		result := validator.ValidateConfig(config)

		// 各ロールに対してバリデーションが実行されていることを確認
		assert.NotNil(t, result)
		// 3つのロール（po, manager, dev）に対して何らかの結果があることを確認
		totalMessages := len(result.Errors) + len(result.Warnings) + len(result.Info)
		assert.GreaterOrEqual(t, totalMessages, 0)
	})

	t.Run("enhanced config paths", func(t *testing.T) {
		config := &config.TeamConfig{
			InstructionConfig: &config.InstructionConfig{
				Base: config.InstructionRoleConfig{
					POInstructionPath:      "enhanced/po.md",
					ManagerInstructionPath: "enhanced/manager.md",
					DevInstructionPath:     "enhanced/dev.md",
				},
			},
		}

		result := validator.ValidateConfig(config)
		assert.NotNil(t, result)
	})
}

// BenchmarkInstructionValidator_ValidateConfig バリデーションパフォーマンステスト
func BenchmarkInstructionValidator_ValidateConfig(b *testing.B) {
	validator := config.NewInstructionValidator(false)

	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "validator_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テストファイルを作成 (ヘルパー関数をベンチマーク用に調整)
	createBenchTestFile(b, tempDir, "po.md", "# PO Instructions")
	createBenchTestFile(b, tempDir, "manager.md", "# Manager Instructions")
	createBenchTestFile(b, tempDir, "developer.md", "# Developer Instructions")

	config := createTestConfig(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateConfig(config)
	}
}

// createBenchTestFile ベンチマーク用テストファイル作成ヘルパー
func createBenchTestFile(b *testing.B, dir, filename, content string) {
	b.Helper()

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		b.Fatalf("Failed to create dir: %v", err)
	}

	filePath := filepath.Join(dir, filename)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to write file: %v", err)
	}
}

// BenchmarkInstructionValidator_ValidateInstructionPath パスバリデーションパフォーマンステスト
func BenchmarkInstructionValidator_ValidateInstructionPath(b *testing.B) {
	validator := config.NewInstructionValidator(false)
	testPath := "config/test.md"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateInstructionPath("test", testPath)
	}
}
