package tmux

import (
	"fmt"
	"os"
)

// TmuxDetectionError is an error when execution inside tmux environment is detected
type TmuxDetectionError struct {
	IsInsideTmux bool
	SessionInfo  string
	PaneInfo     string
}

// Error returns the error message
func (e *TmuxDetectionError) Error() string {
	return "execution inside tmux environment was detected"
}

// IsInsideTmux detects whether execution is inside tmux environment
func IsInsideTmux() (bool, *TmuxDetectionError) {
	tmuxEnv := os.Getenv("TMUX")
	tmuxPane := os.Getenv("TMUX_PANE")

	// Check for TMUX environment variable
	if tmuxEnv != "" {
		return true, &TmuxDetectionError{
			IsInsideTmux: true,
			SessionInfo:  tmuxEnv,
			PaneInfo:     tmuxPane,
		}
	}

	// Additional check using TMUX_PANE environment variable
	if tmuxPane != "" {
		return true, &TmuxDetectionError{
			IsInsideTmux: true,
			SessionInfo:  tmuxEnv,
			PaneInfo:     tmuxPane,
		}
	}

	return false, nil
}

// PrintErrorMessage displays user-friendly error message
func PrintErrorMessage(debugMode bool, err *TmuxDetectionError) {
	if !debugMode {
		// Basic error message
		_, _ = fmt.Fprintf(os.Stderr, "❌ Error: This command cannot be executed from inside tmux.\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "Please exit tmux or run from a different terminal.\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "💡 Solutions:\n")
		_, _ = fmt.Fprintf(os.Stderr, "  1. Detach tmux with Ctrl+B, D\n")
		_, _ = fmt.Fprintf(os.Stderr, "  2. Exit tmux with 'exit'\n")
		_, _ = fmt.Fprintf(os.Stderr, "  3. Open a new terminal window\n")
	} else {
		// Detailed message in debug mode
		_, _ = fmt.Fprintf(os.Stderr, "❌ Error: Execution inside tmux environment detected\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "[Debug Information]\n")
		_, _ = fmt.Fprintf(os.Stderr, "  TMUX: %s\n", err.SessionInfo)
		_, _ = fmt.Fprintf(os.Stderr, "  TMUX_PANE: %s\n", err.PaneInfo)

		// Get session name if possible
		sessionName := os.Getenv("TMUX_SESSION")
		if sessionName != "" {
			_, _ = fmt.Fprintf(os.Stderr, "  Session name: %s\n", sessionName)
		}

		_, _ = fmt.Fprintf(os.Stderr, "\nThis command manages tmux sessions,\n")
		_, _ = fmt.Fprintf(os.Stderr, "so it must be run from outside tmux environment.\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "💡 Solutions:\n")
		_, _ = fmt.Fprintf(os.Stderr, "  1. Detach tmux with Ctrl+B, D\n")
		_, _ = fmt.Fprintf(os.Stderr, "  2. Exit tmux with 'exit'\n")
		_, _ = fmt.Fprintf(os.Stderr, "  3. Open a new terminal window\n")
	}
}
