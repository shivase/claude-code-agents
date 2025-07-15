package cmd

import (
	"testing"

	"github.com/shivase/claude-code-agents/internal/cmd"
	"github.com/stretchr/testify/assert"
)

func TestParseArguments_SimpleValidation(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedSession   string
		expectedResetMode bool
		expectError       bool
	}{
		{
			name:              "Basic session name",
			args:              []string{"ai-teams"},
			expectedSession:   "ai-teams",
			expectedResetMode: false,
			expectError:       false,
		},
		{
			name:              "No arguments",
			args:              []string{},
			expectedSession:   "",
			expectedResetMode: false,
			expectError:       false,
		},
		{
			name:              "With reset flag",
			args:              []string{"--reset", "test-session"},
			expectedSession:   "test-session",
			expectedResetMode: true,
			expectError:       false,
		},
		{
			name:              "Debug flag error",
			args:              []string{"--debug"},
			expectedSession:   "",
			expectedResetMode: false,
			expectError:       true,
		},
		{
			name:              "Verbose flag",
			args:              []string{"--verbose", "test-session"},
			expectedSession:   "test-session",
			expectedResetMode: false,
			expectError:       false,
		},
		{
			name:              "Silent flag",
			args:              []string{"--silent", "test-session"},
			expectedSession:   "test-session",
			expectedResetMode: false,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionName, resetMode, err := cmd.ParseArguments(tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSession, sessionName)
				assert.Equal(t, tt.expectedResetMode, resetMode)
			}
		})
	}
}

func TestParseArguments_EdgeCases(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedSession   string
		expectedResetMode bool
		expectError       bool
	}{
		{
			name:              "Empty session name",
			args:              []string{""},
			expectedSession:   "",
			expectedResetMode: false,
			expectError:       false,
		},
		{
			name:              "Session name with special characters",
			args:              []string{"test-session_123"},
			expectedSession:   "test-session_123",
			expectedResetMode: false,
			expectError:       false,
		},
		{
			name:              "Japanese session name",
			args:              []string{"テスト-セッション"},
			expectedSession:   "テスト-セッション",
			expectedResetMode: false,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionName, resetMode, err := cmd.ParseArguments(tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSession, sessionName)
				assert.Equal(t, tt.expectedResetMode, resetMode)
			}
		})
	}
}

func TestParseArguments_ComplexCombinations(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedSession   string
		expectedResetMode bool
		expectError       bool
	}{
		{
			name:              "Reset and session name",
			args:              []string{"--reset", "complex-session"},
			expectedSession:   "complex-session",
			expectedResetMode: true,
			expectError:       false,
		},
		{
			name:              "Verbose, reset and session name",
			args:              []string{"--verbose", "--reset", "verbose-session"},
			expectedSession:   "verbose-session",
			expectedResetMode: true,
			expectError:       false,
		},
		{
			name:              "Silent and session name",
			args:              []string{"--silent", "silent-session"},
			expectedSession:   "silent-session",
			expectedResetMode: false,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionName, resetMode, err := cmd.ParseArguments(tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSession, sessionName)
				assert.Equal(t, tt.expectedResetMode, resetMode)
			}
		})
	}
}

// Benchmark test
func BenchmarkParseArguments_Simple(b *testing.B) {
	args := []string{"test-session"}
	for i := 0; i < b.N; i++ {
		cmd.ParseArguments(args)
	}
}

func BenchmarkParseArguments_Complex(b *testing.B) {
	args := []string{"--verbose", "--reset", "complex-session"}
	for i := 0; i < b.N; i++ {
		cmd.ParseArguments(args)
	}
}
