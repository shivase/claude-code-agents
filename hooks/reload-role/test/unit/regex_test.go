package main

import (
	"regexp"
	"testing"
)

// TestReloadRoleRegexMatching /reload-roleコマンドの正規表現マッチングテスト
func TestReloadRoleRegexMatching(t *testing.T) {
	// 正規表現パターン（main.goから抽出）
	reloadRoleRegex := regexp.MustCompile(`^/reload-role\s+([a-zA-Z]+)`)

	testCases := []struct {
		name          string
		input         string
		expectedMatch bool
		expectedRole  string
		description   string
	}{
		// 正常ケース
		{
			name:          "ValidRole_PO",
			input:         "/reload-role po",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "POの役割を正常に認識",
		},
		{
			name:          "ValidRole_Manager",
			input:         "/reload-role manager",
			expectedMatch: true,
			expectedRole:  "manager",
			description:   "マネージャーの役割を正常に認識",
		},
		{
			name:          "ValidRole_Developer",
			input:         "/reload-role developer",
			expectedMatch: true,
			expectedRole:  "developer",
			description:   "開発者の役割を正常に認識",
		},
		{
			name:          "ValidRole_Uppercase",
			input:         "/reload-role PO",
			expectedMatch: true,
			expectedRole:  "PO",
			description:   "大文字の役割名を正常に認識",
		},
		{
			name:          "ValidRole_MixedCase",
			input:         "/reload-role DevOps",
			expectedMatch: true,
			expectedRole:  "DevOps",
			description:   "混合大小文字の役割名を正常に認識",
		},
		{
			name:          "ValidRole_SingleSpace",
			input:         "/reload-role admin",
			expectedMatch: true,
			expectedRole:  "admin",
			description:   "単一スペースでの役割名を正常に認識",
		},
		{
			name:          "ValidRole_MultipleSpaces",
			input:         "/reload-role   tester",
			expectedMatch: true,
			expectedRole:  "tester",
			description:   "複数スペースでも正常に認識",
		},

		// エラーケース
		{
			name:          "InvalidCommand_NoRole",
			input:         "/reload-role",
			expectedMatch: false,
			expectedRole:  "",
			description:   "役割名なしのコマンドは無効",
		},
		{
			name:          "InvalidCommand_NoSlash",
			input:         "reload-role po",
			expectedMatch: false,
			expectedRole:  "",
			description:   "スラッシュなしのコマンドは無効",
		},
		{
			name:          "InvalidCommand_WrongCommand",
			input:         "/restart-role po",
			expectedMatch: false,
			expectedRole:  "",
			description:   "間違ったコマンド名は無効",
		},
		{
			name:          "InvalidRole_WithNumbers",
			input:         "/reload-role dev123",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "数字を含む役割名は正規表現でマッチするが、最初の英字部分のみ抽出",
		},
		{
			name:          "InvalidRole_WithSymbols",
			input:         "/reload-role dev-ops",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "ハイフンを含む役割名は正規表現でマッチするが、最初の英字部分のみ抽出",
		},
		{
			name:          "InvalidRole_WithUnderscore",
			input:         "/reload-role dev_ops",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "アンダースコアを含む役割名は正規表現でマッチするが、最初の英字部分のみ抽出",
		},
		{
			name:          "InvalidRole_WithSpecialChars",
			input:         "/reload-role dev@ops",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "特殊文字を含む役割名は正規表現でマッチするが、最初の英字部分のみ抽出",
		},
		{
			name:          "InvalidRole_EmptyString",
			input:         "/reload-role ",
			expectedMatch: false,
			expectedRole:  "",
			description:   "空の役割名は無効",
		},
		{
			name:          "InvalidRole_OnlySpaces",
			input:         "/reload-role    ",
			expectedMatch: false,
			expectedRole:  "",
			description:   "スペースのみの役割名は無効",
		},

		// エッジケース
		{
			name:          "EdgeCase_ExtraText",
			input:         "/reload-role po extra text",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "追加テキストがあっても最初の役割名を認識",
		},
		{
			name:          "EdgeCase_TabCharacter",
			input:         "/reload-role\tpo",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "タブ文字も\\s+でマッチする（\\s+は空白文字全般）",
		},
		{
			name:          "EdgeCase_NewLine",
			input:         "/reload-role\npo",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "改行文字も\\s+でマッチする（\\s+は空白文字全般）",
		},
		{
			name:          "EdgeCase_LeadingSpaces",
			input:         "   /reload-role po",
			expectedMatch: false,
			expectedRole:  "",
			description:   "行頭スペースがあるコマンドは無効",
		},
		{
			name:          "EdgeCase_TrailingSpaces",
			input:         "/reload-role po   ",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "行末スペースがあっても有効",
		},
		{
			name:          "EdgeCase_VeryLongRole",
			input:         "/reload-role " + generateLongString(100),
			expectedMatch: true,
			expectedRole:  generateLongString(100),
			description:   "非常に長い役割名も正規表現は通る",
		},
		{
			name:          "EdgeCase_SingleChar",
			input:         "/reload-role a",
			expectedMatch: true,
			expectedRole:  "a",
			description:   "単一文字の役割名は有効",
		},
		{
			name:          "EdgeCase_JapaneseChars",
			input:         "/reload-role 管理者",
			expectedMatch: false,
			expectedRole:  "",
			description:   "日本語文字は無効（英字のみ許可）",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			matches := reloadRoleRegex.FindStringSubmatch(tc.input)

			// Assert
			if tc.expectedMatch {
				if len(matches) == 0 {
					t.Errorf("Expected match but got none for input: %s", tc.input)
					return
				}
				if matches[1] != tc.expectedRole {
					t.Errorf("Expected role '%s' but got '%s' for input: %s", tc.expectedRole, matches[1], tc.input)
				}
			} else {
				if len(matches) != 0 {
					t.Errorf("Expected no match but got role '%s' for input: %s", matches[1], tc.input)
				}
			}
		})
	}
}

// TestReloadRoleRegexPattern 正規表現パターン自体のテスト
func TestReloadRoleRegexPattern(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     string
		input       string
		shouldMatch bool
		description string
	}{
		{
			name:        "OriginalPattern_ValidInput",
			pattern:     `^/reload-role\s+([a-zA-Z]+)`,
			input:       "/reload-role po",
			shouldMatch: true,
			description: "オリジナルパターンで正常入力をテスト",
		},
		{
			name:        "OriginalPattern_InvalidInput",
			pattern:     `^/reload-role\s+([a-zA-Z]+)`,
			input:       "/reload-role 123",
			shouldMatch: false,
			description: "オリジナルパターンで無効入力をテスト",
		},
		{
			name:        "CaseInsensitivePattern_UpperCase",
			pattern:     `^/reload-role\s+([a-zA-Z]+)`,
			input:       "/reload-role ADMIN",
			shouldMatch: true,
			description: "大文字の役割名をテスト",
		},
		{
			name:        "CaseInsensitivePattern_LowerCase",
			pattern:     `^/reload-role\s+([a-zA-Z]+)`,
			input:       "/reload-role admin",
			shouldMatch: true,
			description: "小文字の役割名をテスト",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			regex := regexp.MustCompile(tc.pattern)

			// Act
			matches := regex.FindStringSubmatch(tc.input)

			// Assert
			if tc.shouldMatch {
				if len(matches) == 0 {
					t.Errorf("Pattern '%s' should match input '%s' but didn't", tc.pattern, tc.input)
				}
			} else {
				if len(matches) != 0 {
					t.Errorf("Pattern '%s' should not match input '%s' but did", tc.pattern, tc.input)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkReloadRoleRegexMatching(b *testing.B) {
	reloadRoleRegex := regexp.MustCompile(`^/reload-role\s+([a-zA-Z]+)`)
	input := "/reload-role po"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reloadRoleRegex.FindStringSubmatch(input)
	}
}

func BenchmarkReloadRoleRegexCompilation(b *testing.B) {
	pattern := `^/reload-role\s+([a-zA-Z]+)`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		regexp.MustCompile(pattern)
	}
}

// Helper function to generate long strings for testing
func generateLongString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
