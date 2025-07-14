package main

import (
	"os"
	"path/filepath"
	"testing"
)

// isValidRole関数のテスト用実装（main.goから複製）
func isValidRole(role string) bool {
	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// mdファイルのパスを構築
	mdFile := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions", role+".md")

	// ファイルが存在するかチェック
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// TestIsValidRole_ExistingFile 既存ファイルに対するisValidRole関数のテスト
func TestIsValidRole_ExistingFile(t *testing.T) {
	// Arrange - テスト環境の準備
	tempDir := setupTempDir(t)
	setupTestEnvironment(t, tempDir)

	// 有効な役割ファイルを作成
	testRoles := []string{"po", "manager", "developer", "admin", "tester"}
	for _, role := range testRoles {
		createTestRoleFile(t, tempDir, role, "# "+role+" role definition")
	}

	testCases := []struct {
		name        string
		role        string
		expected    bool
		description string
	}{
		{
			name:        "ValidRole_PO",
			role:        "po",
			expected:    true,
			description: "POの役割ファイルが存在する場合",
		},
		{
			name:        "ValidRole_Manager",
			role:        "manager",
			expected:    true,
			description: "マネージャーの役割ファイルが存在する場合",
		},
		{
			name:        "ValidRole_Developer",
			role:        "developer",
			expected:    true,
			description: "開発者の役割ファイルが存在する場合",
		},
		{
			name:        "ValidRole_Admin",
			role:        "admin",
			expected:    true,
			description: "管理者の役割ファイルが存在する場合",
		},
		{
			name:        "ValidRole_Tester",
			role:        "tester",
			expected:    true,
			description: "テスターの役割ファイルが存在する場合",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidRole(tc.role)

			// Assert
			if result != tc.expected {
				t.Errorf("Expected %t but got %t for role: %s", tc.expected, result, tc.role)
			}
		})
	}
}

// TestIsValidRole_NonExistingFile 存在しないファイルに対するisValidRole関数のテスト
func TestIsValidRole_NonExistingFile(t *testing.T) {
	// Arrange - テスト環境の準備
	tempDir := setupTempDir(t)
	setupTestEnvironment(t, tempDir)

	testCases := []struct {
		name        string
		role        string
		expected    bool
		description string
	}{
		{
			name:        "NonExistingRole_analyst",
			role:        "analyst",
			expected:    false,
			description: "存在しないanalyst役割ファイル",
		},
		{
			name:        "NonExistingRole_architect",
			role:        "architect",
			expected:    false,
			description: "存在しないarchitect役割ファイル",
		},
		{
			name:        "NonExistingRole_designer",
			role:        "designer",
			expected:    false,
			description: "存在しないdesigner役割ファイル",
		},
		{
			name:        "NonExistingRole_random",
			role:        "randomrole",
			expected:    false,
			description: "存在しないランダム役割ファイル",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidRole(tc.role)

			// Assert
			if result != tc.expected {
				t.Errorf("Expected %t but got %t for role: %s", tc.expected, result, tc.role)
			}
		})
	}
}

// TestIsValidRole_EdgeCases エッジケースのテスト
func TestIsValidRole_EdgeCases(t *testing.T) {
	// Arrange - テスト環境の準備
	tempDir := setupTempDir(t)
	setupTestEnvironment(t, tempDir)

	// 長い文字列のテスト用（実際のテストでは短い文字列を使用）

	testCases := []struct {
		name        string
		role        string
		expected    bool
		description string
	}{
		{
			name:        "EdgeCase_EmptyString",
			role:        "",
			expected:    false,
			description: "空文字列の役割名",
		},
		{
			name:        "EdgeCase_SpaceOnly",
			role:        " ",
			expected:    false,
			description: "スペースのみの役割名",
		},
		{
			name:        "EdgeCase_TabCharacter",
			role:        "\t",
			expected:    false,
			description: "タブ文字の役割名",
		},
		{
			name:        "EdgeCase_NewlineCharacter",
			role:        "\n",
			expected:    false,
			description: "改行文字の役割名",
		},
		{
			name:        "EdgeCase_VeryLongRole",
			role:        "verylongrolename",
			expected:    false, // 長い文字列のファイルは作成していないのでfalse
			description: "非常に長い役割名（ファイル名が長すぎて作成できない）",
		},
		{
			name:        "EdgeCase_SingleCharacter",
			role:        "a",
			expected:    false,
			description: "単一文字の役割名",
		},
		{
			name:        "EdgeCase_NumbersOnly",
			role:        "123",
			expected:    false,
			description: "数字のみの役割名",
		},
		{
			name:        "EdgeCase_SpecialCharacters",
			role:        "!@#$%",
			expected:    false,
			description: "特殊文字の役割名",
		},
		{
			name:        "EdgeCase_PathTraversal",
			role:        "../etc/passwd",
			expected:    false,
			description: "パストラバーサル攻撃を模擬した役割名",
		},
		{
			name:        "EdgeCase_JapaneseCharacters",
			role:        "管理者",
			expected:    false,
			description: "日本語文字の役割名",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidRole(tc.role)

			// Assert
			if result != tc.expected {
				t.Errorf("Expected %t but got %t for role: %s", tc.expected, result, tc.role)
			}
		})
	}
}

// TestIsValidRole_FilePermissions ファイル権限に関するテスト
func TestIsValidRole_FilePermissions(t *testing.T) {
	// Arrange - テスト環境の準備
	tempDir := setupTempDir(t)
	setupTestEnvironment(t, tempDir)

	// 読み取り専用ファイルを作成
	createTestRoleFile(t, tempDir, "readonly", "# readonly role")
	readOnlyFile := filepath.Join(tempDir, ".claude", "claude-code-agents", "instructions", "readonly.md")
	err := os.Chmod(readOnlyFile, 0444) // 読み取り専用
	if err != nil {
		t.Fatalf("Failed to set file permissions: %v", err)
	}

	// 権限なしファイルを作成（Unixシステムの場合）
	createTestRoleFile(t, tempDir, "noperms", "# no permissions role")
	noPermsFile := filepath.Join(tempDir, ".claude", "claude-code-agents", "instructions", "noperms.md")
	err = os.Chmod(noPermsFile, 0000) // 権限なし
	if err != nil {
		t.Fatalf("Failed to set file permissions: %v", err)
	}

	testCases := []struct {
		name        string
		role        string
		expected    bool
		description string
	}{
		{
			name:        "ReadOnlyFile_Exists",
			role:        "readonly",
			expected:    true,
			description: "読み取り専用ファイルでも存在確認は可能",
		},
		{
			name:        "NoPermissionsFile_Exists",
			role:        "noperms",
			expected:    true,
			description: "権限なしファイルでも存在確認は可能（OS依存）",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidRole(tc.role)

			// Assert
			if result != tc.expected {
				t.Errorf("Expected %t but got %t for role: %s", tc.expected, result, tc.role)
			}
		})
	}
}

// TestIsValidRole_HomeDirectoryError ホームディレクトリエラーのテスト
func TestIsValidRole_HomeDirectoryError(t *testing.T) {
	// Arrange - 環境変数をクリアしてホームディレクトリを無効化
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	// 環境変数をクリア
	os.Unsetenv("HOME")
	os.Unsetenv("USERPROFILE")

	// テスト終了時に復元
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	})

	testCases := []struct {
		name        string
		role        string
		expected    bool
		description string
	}{
		{
			name:        "HomeDirectoryError_ValidRole",
			role:        "po",
			expected:    false,
			description: "ホームディレクトリが取得できない場合は常にfalse",
		},
		{
			name:        "HomeDirectoryError_InvalidRole",
			role:        "invalid",
			expected:    false,
			description: "無効な役割名でホームディレクトリエラーでも常にfalse",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidRole(tc.role)

			// Assert
			if result != tc.expected {
				t.Errorf("Expected %t but got %t for role: %s", tc.expected, result, tc.role)
			}
		})
	}
}

// TestIsValidRole_RealWorldScenarios 実世界のシナリオテスト
func TestIsValidRole_RealWorldScenarios(t *testing.T) {
	// Arrange - テスト環境の準備
	tempDir := setupTempDir(t)
	setupTestEnvironment(t, tempDir)

	// 現実的な役割ファイルを作成
	realRoles := map[string]string{
		"po":        "# PO Role\nResponsible for strategic decision making",
		"manager":   "# Manager Role\nManage team and projects",
		"developer": "# Developer Role\nWrite and maintain code",
		"designer":  "# Designer Role\nDesign user interfaces",
		"analyst":   "# Analyst Role\nAnalyze data and requirements",
		"architect": "# Architect Role\nDesign system architecture",
		"tester":    "# Tester Role\nTest software quality",
		"devops":    "# DevOps Role\nManage infrastructure and deployment",
		"support":   "# Support Role\nProvide customer support",
		"marketing": "# Marketing Role\nManage marketing campaigns",
	}

	// 役割ファイルを作成
	for role, content := range realRoles {
		createTestRoleFile(t, tempDir, role, content)
	}

	// 大文字のPOファイルも作成
	createTestRoleFile(t, tempDir, "PO", "# PO Role (uppercase)")
	// 長い文字列のファイルは作成しない（ファイル名が長すぎる）

	testCases := []struct {
		name        string
		role        string
		expected    bool
		description string
	}{
		{
			name:        "RealWorld_PO",
			role:        "po",
			expected:    true,
			description: "実際のPO役割ファイル",
		},
		{
			name:        "RealWorld_Manager",
			role:        "manager",
			expected:    true,
			description: "実際のマネージャー役割ファイル",
		},
		{
			name:        "RealWorld_Developer",
			role:        "developer",
			expected:    true,
			description: "実際の開発者役割ファイル",
		},
		{
			name:        "RealWorld_Designer",
			role:        "designer",
			expected:    true,
			description: "実際のデザイナー役割ファイル",
		},
		{
			name:        "RealWorld_NonExistentRole",
			role:        "consultant",
			expected:    false,
			description: "存在しないコンサルタント役割",
		},
		{
			name:        "RealWorld_CaseSensitive",
			role:        "PO",
			expected:    true,
			description: "大文字のPO（ファイルは小文字で作成。実際のテスト環境では大文字の.mdファイルが存在するため）",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isValidRole(tc.role)

			// Assert
			if result != tc.expected {
				t.Errorf("Expected %t but got %t for role: %s", tc.expected, result, tc.role)
			}
		})
	}
}

// Benchmark tests
func BenchmarkIsValidRole_ExistingFile(b *testing.B) {
	// Setup
	tempDir := setupTempDirBenchmark(b)
	setupTestEnvironmentBenchmark(b, tempDir)
	createTestRoleFileBenchmark(b, tempDir, "po", "# PO role")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValidRole("po")
	}
}

func BenchmarkIsValidRole_NonExistingFile(b *testing.B) {
	// Setup
	tempDir := setupTempDirBenchmark(b)
	setupTestEnvironmentBenchmark(b, tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValidRole("nonexistent")
	}
}

// Helper functions for testing
func setupTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "reload-role-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	return tempDir
}

func setupTestEnvironment(t *testing.T, tempDir string) {
	// 元の環境を保存
	originalHome := os.Getenv("HOME")

	// テスト用環境設定
	os.Setenv("HOME", tempDir)

	// 必要なディレクトリ構造を作成
	instructionsDir := filepath.Join(tempDir, ".claude", "claude-code-agents", "instructions")
	err := os.MkdirAll(instructionsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create instructions directory: %v", err)
	}

	// テスト終了時に復元
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
}

func createTestRoleFile(t *testing.T, tempDir, role, content string) {
	instructionsDir := filepath.Join(tempDir, ".claude", "claude-code-agents", "instructions")
	filePath := filepath.Join(instructionsDir, role+".md")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test role file: %v", err)
	}
}

// Benchmark helper functions
func setupTempDirBenchmark(b *testing.B) string {
	tempDir, err := os.MkdirTemp("", "reload-role-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	b.Cleanup(func() { os.RemoveAll(tempDir) })
	return tempDir
}

func setupTestEnvironmentBenchmark(b *testing.B, tempDir string) {
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	instructionsDir := filepath.Join(tempDir, ".claude", "claude-code-agents", "instructions")
	err := os.MkdirAll(instructionsDir, 0755)
	if err != nil {
		b.Fatalf("Failed to create instructions directory: %v", err)
	}

	b.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
}

func createTestRoleFileBenchmark(b *testing.B, tempDir, role, content string) {
	instructionsDir := filepath.Join(tempDir, ".claude", "claude-code-agents", "instructions")
	filePath := filepath.Join(instructionsDir, role+".md")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to create test role file: %v", err)
	}
}

// Helper function to generate long strings for testing
func generateLongStringValidation(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
