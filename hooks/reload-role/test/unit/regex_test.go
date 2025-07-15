package main

import (
	"regexp"
	"testing"
)

// TestReloadRoleRegexMatching Tests regular expression matching for /reload-role command
func TestReloadRoleRegexMatching(t *testing.T) {
	// Regular expression pattern (extracted from main.go)
	reloadRoleRegex := regexp.MustCompile(`^/reload-role\s+([a-zA-Z]+)`)

	testCases := []struct {
		name          string
		input         string
		expectedMatch bool
		expectedRole  string
		description   string
	}{
		// Normal cases
		{
			name:          "ValidRole_PO",
			input:         "/reload-role po",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "Correctly recognizes PO role",
		},
		{
			name:          "ValidRole_Manager",
			input:         "/reload-role manager",
			expectedMatch: true,
			expectedRole:  "manager",
			description:   "Correctly recognizes manager role",
		},
		{
			name:          "ValidRole_Developer",
			input:         "/reload-role developer",
			expectedMatch: true,
			expectedRole:  "developer",
			description:   "Correctly recognizes developer role",
		},
		{
			name:          "ValidRole_Uppercase",
			input:         "/reload-role PO",
			expectedMatch: true,
			expectedRole:  "PO",
			description:   "Correctly recognizes uppercase role name",
		},
		{
			name:          "ValidRole_MixedCase",
			input:         "/reload-role DevOps",
			expectedMatch: true,
			expectedRole:  "DevOps",
			description:   "Correctly recognizes mixed case role name",
		},
		{
			name:          "ValidRole_SingleSpace",
			input:         "/reload-role admin",
			expectedMatch: true,
			expectedRole:  "admin",
			description:   "Correctly recognizes role name with single space",
		},
		{
			name:          "ValidRole_MultipleSpaces",
			input:         "/reload-role   tester",
			expectedMatch: true,
			expectedRole:  "tester",
			description:   "Correctly recognizes role name with multiple spaces",
		},

		// Error cases
		{
			name:          "InvalidCommand_NoRole",
			input:         "/reload-role",
			expectedMatch: false,
			expectedRole:  "",
			description:   "Command without role name is invalid",
		},
		{
			name:          "InvalidCommand_NoSlash",
			input:         "reload-role po",
			expectedMatch: false,
			expectedRole:  "",
			description:   "Command without slash is invalid",
		},
		{
			name:          "InvalidCommand_WrongCommand",
			input:         "/restart-role po",
			expectedMatch: false,
			expectedRole:  "",
			description:   "Wrong command name is invalid",
		},
		{
			name:          "InvalidRole_WithNumbers",
			input:         "/reload-role dev123",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "Role name with numbers matches regex but only extracts initial alphabetic part",
		},
		{
			name:          "InvalidRole_WithSymbols",
			input:         "/reload-role dev-ops",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "Role name with hyphen matches regex but only extracts initial alphabetic part",
		},
		{
			name:          "InvalidRole_WithUnderscore",
			input:         "/reload-role dev_ops",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "Role name with underscore matches regex but only extracts initial alphabetic part",
		},
		{
			name:          "InvalidRole_WithSpecialChars",
			input:         "/reload-role dev@ops",
			expectedMatch: true,
			expectedRole:  "dev",
			description:   "Role name with special characters matches regex but only extracts initial alphabetic part",
		},
		{
			name:          "InvalidRole_EmptyString",
			input:         "/reload-role ",
			expectedMatch: false,
			expectedRole:  "",
			description:   "Empty role name is invalid",
		},
		{
			name:          "InvalidRole_OnlySpaces",
			input:         "/reload-role    ",
			expectedMatch: false,
			expectedRole:  "",
			description:   "Role name with only spaces is invalid",
		},

		// Edge cases
		{
			name:          "EdgeCase_ExtraText",
			input:         "/reload-role po extra text",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "Recognizes first role name even with extra text",
		},
		{
			name:          "EdgeCase_TabCharacter",
			input:         "/reload-role\tpo",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "Tab character matches with \\s+ (\\s+ matches all whitespace characters)",
		},
		{
			name:          "EdgeCase_NewLine",
			input:         "/reload-role\npo",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "Newline character matches with \\s+ (\\s+ matches all whitespace characters)",
		},
		{
			name:          "EdgeCase_LeadingSpaces",
			input:         "   /reload-role po",
			expectedMatch: false,
			expectedRole:  "",
			description:   "Command with leading spaces is invalid",
		},
		{
			name:          "EdgeCase_TrailingSpaces",
			input:         "/reload-role po   ",
			expectedMatch: true,
			expectedRole:  "po",
			description:   "Command with trailing spaces is still valid",
		},
		{
			name:          "EdgeCase_VeryLongRole",
			input:         "/reload-role " + generateLongString(100),
			expectedMatch: true,
			expectedRole:  generateLongString(100),
			description:   "Very long role name still passes regex",
		},
		{
			name:          "EdgeCase_SingleChar",
			input:         "/reload-role a",
			expectedMatch: true,
			expectedRole:  "a",
			description:   "Single character role name is valid",
		},
		{
			name:          "EdgeCase_JapaneseChars",
			input:         "/reload-role 管理者",
			expectedMatch: false,
			expectedRole:  "",
			description:   "Japanese characters are invalid (only alphabetic characters allowed)",
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

// TestReloadRoleRegexPattern Tests the regular expression pattern itself
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
			description: "Test valid input with original pattern",
		},
		{
			name:        "OriginalPattern_InvalidInput",
			pattern:     `^/reload-role\s+([a-zA-Z]+)`,
			input:       "/reload-role 123",
			shouldMatch: false,
			description: "Test invalid input with original pattern",
		},
		{
			name:        "CaseInsensitivePattern_UpperCase",
			pattern:     `^/reload-role\s+([a-zA-Z]+)`,
			input:       "/reload-role ADMIN",
			shouldMatch: true,
			description: "Test uppercase role name",
		},
		{
			name:        "CaseInsensitivePattern_LowerCase",
			pattern:     `^/reload-role\s+([a-zA-Z]+)`,
			input:       "/reload-role admin",
			shouldMatch: true,
			description: "Test lowercase role name",
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
