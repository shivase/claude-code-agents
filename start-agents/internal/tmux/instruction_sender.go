package tmux

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// SendInstructionToPaneWithConfig sends instruction file using configuration
func (tm *TmuxManagerImpl) SendInstructionToPaneWithConfig(sessionName, pane, agent, instructionsDir string, config interface{}) error {
	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("📤 Starting configuration-based instruction sending")

	// Determine instruction file path from configuration
	var instructionFile string

	// Get role-specific file names from configuration
	type InstructionConfig interface {
		GetPOInstructionFile() string
		GetManagerInstructionFile() string
		GetDevInstructionFile() string
	}

	if ic, ok := config.(InstructionConfig); ok {
		switch agent {
		case "po":
			if ic.GetPOInstructionFile() != "" {
				instructionFile = filepath.Join(instructionsDir, ic.GetPOInstructionFile())
			} else {
				instructionFile = filepath.Join(instructionsDir, "po.md")
			}
		case "manager":
			if ic.GetManagerInstructionFile() != "" {
				instructionFile = filepath.Join(instructionsDir, ic.GetManagerInstructionFile())
			} else {
				instructionFile = filepath.Join(instructionsDir, "manager.md")
			}
		case "dev1", "dev2", "dev3", "dev4":
			if ic.GetDevInstructionFile() != "" {
				instructionFile = filepath.Join(instructionsDir, ic.GetDevInstructionFile())
			} else {
				instructionFile = filepath.Join(instructionsDir, "developer.md")
			}
		default:
			log.Error().Str("agent", agent).Msg("❌ Unknown agent type")
			return fmt.Errorf("unknown agent type: %s", agent)
		}
	} else {
		// Use default file names if configuration is not provided
		switch agent {
		case "po":
			instructionFile = filepath.Join(instructionsDir, "po.md")
		case "manager":
			instructionFile = filepath.Join(instructionsDir, "manager.md")
		case "dev1", "dev2", "dev3", "dev4":
			instructionFile = filepath.Join(instructionsDir, "developer.md")
		default:
			log.Error().Str("agent", agent).Msg("❌ Unknown agent type")
			return fmt.Errorf("unknown agent type: %s", agent)
		}
	}

	log.Info().Str("instruction_file", instructionFile).Msg("📁 Configuration-based instruction file path determined")

	// Verify instruction file exists (enhanced version)
	fileInfo, err := os.Stat(instructionFile)
	if os.IsNotExist(err) {
		log.Warn().Str("instruction_file", instructionFile).Msg("⚠️ Instruction file does not exist (skipping)")
		return nil // Skip if file doesn't exist (not an error)
	}
	if err != nil {
		log.Error().Str("instruction_file", instructionFile).Err(err).Msg("❌ Failed to get file information")
		return fmt.Errorf("failed to stat instruction file: %w", err)
	}
	if fileInfo.Size() == 0 {
		log.Warn().Str("instruction_file", instructionFile).Msg("⚠️ Instruction file is empty (skipping)")
		return nil
	}

	log.Info().Str("instruction_file", instructionFile).Int64("file_size", fileInfo.Size()).Msg("✅ Configuration-based file existence verified")

	// Wait for Claude CLI to be ready (enhanced version)
	if err := tm.waitForClaudeReady(sessionName, pane, 10*time.Second); err != nil {
		log.Warn().Str("session", sessionName).Str("pane", pane).Err(err).Msg("⚠️ Claude CLI readiness wait timeout (continuing)")
	}

	// Send instruction file using cat command (with retry functionality)
	catCommand := fmt.Sprintf("cat \"%s\"", instructionFile)

	for attempt := 1; attempt <= 3; attempt++ {
		log.Info().Str("command", catCommand).Int("attempt", attempt).Msg("📋 Sending configuration-based cat command")

		if err := tm.SendKeysWithEnter(sessionName, pane, catCommand); err != nil {
			log.Warn().Err(err).Int("attempt", attempt).Msg("⚠️ Failed to send cat command")
			if attempt == 3 {
				return fmt.Errorf("failed to send instruction file after 3 attempts: %w", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// Wait for cat command execution to complete
		time.Sleep(2 * time.Second)
		log.Info().Int("attempt", attempt).Msg("✅ Configuration-based cat command sent successfully")
		break
	}

	// Send Enter for Claude CLI execution (optimized version)
	time.Sleep(1 * time.Second)
	log.Info().Msg("🔄 Sending Enter for Claude CLI execution")

	for i := 0; i < 3; i++ {
		if err := tm.SendKeysToPane(sessionName, pane, "C-m"); err != nil {
			log.Warn().Err(err).Int("attempt", i+1).Msg("⚠️ Error sending Enter")
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("✅ Configuration-based instruction sending completed")
	return nil
}
