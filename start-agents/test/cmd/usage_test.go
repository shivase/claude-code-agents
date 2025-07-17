package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/shivase/claude-code-agents/internal/cmd"
	"github.com/stretchr/testify/assert"
)

// Test helper function: captures standard output
func captureStdout(f func()) (string, error) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan bool)

	go func() {
		defer close(done)
		io.Copy(&buf, r)
	}()

	f()

	w.Close()
	os.Stdout = originalStdout
	<-done

	return buf.String(), nil
}

func TestShowUsage(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check basic help contents
	t.Run("Check basic information", func(t *testing.T) {
		assert.Contains(t, output, "AI Parallel Development Team - Integrated Launch System")
		assert.Contains(t, output, "Usage:")
		assert.Contains(t, output, "claude-code-agents")
	})

	// Check argument descriptions
	t.Run("Check argument descriptions", func(t *testing.T) {
		assert.Contains(t, output, "Arguments:")
		assert.Contains(t, output, "session-name")
		assert.Contains(t, output, "tmux session name")
	})

	// Check option descriptions
	t.Run("Check option descriptions", func(t *testing.T) {
		assert.Contains(t, output, "Options:")
		assert.Contains(t, output, "--reset")
		assert.Contains(t, output, "--verbose")
		assert.Contains(t, output, "--debug")
		assert.Contains(t, output, "--silent")
		assert.Contains(t, output, "--help")

		// Check short options
		assert.Contains(t, output, "-v")
		assert.Contains(t, output, "-d")
		assert.Contains(t, output, "-s")
	})

	// Check management command descriptions
	t.Run("Check management command descriptions", func(t *testing.T) {
		assert.Contains(t, output, "Management Commands:")
		assert.Contains(t, output, "--list")
		assert.Contains(t, output, "--delete")
		assert.Contains(t, output, "--delete-all")
		assert.Contains(t, output, "--show-config")
		assert.Contains(t, output, "--config")
		assert.Contains(t, output, "--generate-config")
		assert.Contains(t, output, "--init [ja|en]")
		assert.Contains(t, output, "--doctor")
		assert.Contains(t, output, "--force")
	})

	// Check usage examples
	t.Run("Check usage examples", func(t *testing.T) {
		assert.Contains(t, output, "Examples:")
		assert.Contains(t, output, "claude-code-agents myproject")
		assert.Contains(t, output, "claude-code-agents ai-team")
		assert.Contains(t, output, "claude-code-agents myproject --reset")
		assert.Contains(t, output, "claude-code-agents myproject --verbose")
		assert.Contains(t, output, "claude-code-agents myproject --silent")
		assert.Contains(t, output, "claude-code-agents --list")
		assert.Contains(t, output, "claude-code-agents --delete myproject")
		assert.Contains(t, output, "claude-code-agents --delete-all")
		assert.Contains(t, output, "claude-code-agents --show-config")
		assert.Contains(t, output, "claude-code-agents --config ai-team")
		assert.Contains(t, output, "claude-code-agents --generate-config")
		assert.Contains(t, output, "claude-code-agents --generate-config --force")
		assert.Contains(t, output, "claude-code-agents --init ja")
		assert.Contains(t, output, "claude-code-agents --init en")
		assert.Contains(t, output, "claude-code-agents --init ja --force")
		assert.Contains(t, output, "claude-code-agents --doctor")
	})

	// Check environment variable descriptions
	t.Run("Check environment variable descriptions", func(t *testing.T) {
		assert.Contains(t, output, "Environment Variables:")
		assert.Contains(t, output, "VERBOSE=true")
		assert.Contains(t, output, "SILENT=true")
	})
}

func TestShowUsage_OutputFormat(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("Check output format", func(t *testing.T) {
		// Check for emojis
		assert.Contains(t, output, "🚀")

		// セクションの区切りの確認
		lines := strings.Split(output, "\n")
		assert.True(t, len(lines) > 10, "Should have sufficient lines")

		// Check for appropriate section separators with empty lines
		hasEmptyLines := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				hasEmptyLines = true
				break
			}
		}
		assert.True(t, hasEmptyLines, "Should have empty lines between sections")
	})

	t.Run("Check command line format", func(t *testing.T) {
		// Check that command line examples are in correct format
		commandExamples := []string{
			"claude-code-agents myproject",
			"claude-code-agents ai-team",
			"claude-code-agents myproject --reset",
			"claude-code-agents myproject --verbose",
			"claude-code-agents myproject --silent",
			"claude-code-agents --list",
			"claude-code-agents --delete myproject",
			"claude-code-agents --delete-all",
			"claude-code-agents --show-config",
			"claude-code-agents --config ai-team",
			"claude-code-agents --generate-config",
			"claude-code-agents --generate-config --force",
			"claude-code-agents --init ja",
			"claude-code-agents --init en",
			"claude-code-agents --init ja --force",
			"claude-code-agents --doctor",
		}

		for _, example := range commandExamples {
			assert.Contains(t, output, example, "Command example '%s' should be included", example)
		}
	})
}

func TestShowUsage_EnglishContent(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("Check English descriptions", func(t *testing.T) {
		// Check for major English description texts
		englishTexts := []string{
			"Integrated Launch System",
			"Usage",
			"Arguments",
			"Options",
			"Management Commands",
			"Examples",
			"Environment Variables",
			"session-name",
			"Enable verbose logging",
			"Enable debug logging",
			"Silent mode",
			"Show this help",
			"Show running AI team sessions",
			"Delete specified session",
			"Delete all AI team sessions",
			"Show configuration summary",
			"Show detailed configuration",
			"Generate configuration file template",
			"Overwrite existing files",
			"Initialize system (create directories and config files)",
			"Overwrite existing files during initialization",
			"Run system health check",
		}

		for _, text := range englishTexts {
			assert.Contains(t, output, text, "English text '%s' should be included", text)
		}
	})
}

func TestShowUsage_OptionConsistency(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("Check option consistency", func(t *testing.T) {
		// Check correspondence between long and short options
		optionPairs := map[string]string{
			"--verbose": "-v",
			"--debug":   "-d",
			"--silent":  "-s",
		}

		for longOpt, shortOpt := range optionPairs {
			assert.Contains(t, output, longOpt, "Long option '%s' should be included", longOpt)
			assert.Contains(t, output, shortOpt, "Short option '%s' should be included", shortOpt)

			// Check that long and short options are on the same line
			lines := strings.Split(output, "\n")
			found := false
			for _, line := range lines {
				if strings.Contains(line, longOpt) && strings.Contains(line, shortOpt) {
					found = true
					break
				}
			}
			assert.True(t, found, "Options '%s' and '%s' should be on the same line", longOpt, shortOpt)
		}
	})

	t.Run("Check required and optional arguments", func(t *testing.T) {
		// Check required argument notation
		assert.Contains(t, output, "<session-name>")
		assert.Contains(t, output, "(required)")

		// Check optional argument notation
		assert.Contains(t, output, "[options]")
		assert.Contains(t, output, "[management-commands]")
		assert.Contains(t, output, "[name]")
		assert.Contains(t, output, "[session]")
	})
}

func TestShowUsage_CompleteCoverage(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("Check complete coverage", func(t *testing.T) {
		// Check that all options defined in parser.go are included
		allOptions := []string{
			"--help",
			"--verbose",
			"--debug",
			"--silent",
			"--list",
			"--delete",
			"--delete-all",
			"--show-config",
			"--config",
			"--generate-config",
			"--init",
			"--doctor",
			"--reset",
			"--force",
			"-h",
			"-v",
			"-d",
			"-s",
		}

		for _, option := range allOptions {
			assert.Contains(t, output, option, "Option '%s' should be included in help", option)
		}
	})

	t.Run("Check environment variable coverage", func(t *testing.T) {
		envVars := []string{
			"VERBOSE=true",
			"SILENT=true",
		}

		for _, envVar := range envVars {
			assert.Contains(t, output, envVar, "Environment variable '%s' should be included in help", envVar)
		}
	})
}

func TestShowUsage_Structure(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("Check help structure", func(t *testing.T) {
		lines := strings.Split(output, "\n")

		// Check section order
		sectionOrder := []string{
			"AI Parallel Development Team - Integrated Launch System",
			"Usage:",
			"Arguments:",
			"Options:",
			"Management Commands:",
			"Examples:",
			"Environment Variables:",
		}

		lastIndex := -1
		for _, section := range sectionOrder {
			currentIndex := -1
			for i, line := range lines {
				if strings.Contains(line, section) {
					currentIndex = i
					break
				}
			}

			assert.NotEqual(t, -1, currentIndex, "Section '%s' should be found", section)
			assert.Greater(t, currentIndex, lastIndex, "Section '%s' should be in correct order", section)
			lastIndex = currentIndex
		}
	})

	t.Run("Check indentation", func(t *testing.T) {
		lines := strings.Split(output, "\n")

		// Check that option description lines are properly indented
		optionLines := []string{}
		for _, line := range lines {
			if strings.Contains(line, "--") && !strings.HasPrefix(strings.TrimSpace(line), "claude-code-agents") {
				optionLines = append(optionLines, line)
			}
		}

		assert.Greater(t, len(optionLines), 0, "Option description lines should exist")

		for _, line := range optionLines {
			// Check that indented with at least 2 spaces
			assert.True(t, strings.HasPrefix(line, "  "), "Option line '%s' should be properly indented", strings.TrimSpace(line))
		}
	})
}

// Performance test
func BenchmarkShowUsage(b *testing.B) {
	// Disable standard output and run benchmark
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() {
		os.Stdout = originalStdout
	}()

	for i := 0; i < b.N; i++ {
		cmd.ShowUsage()
	}
}
