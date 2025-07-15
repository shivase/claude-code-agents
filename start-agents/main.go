package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shivase/claude-code-agents/internal/cmd"
	"github.com/shivase/claude-code-agents/internal/logger"
	"github.com/shivase/claude-code-agents/internal/tmux"
)

func main() {
	// Check debug mode first
	args := os.Args[1:]
	debugMode := false
	for _, arg := range args {
		if arg == "--debug" || arg == "-d" {
			debugMode = true
			break
		}
	}

	// Check if running inside tmux environment
	isInTmux, tmuxErr := tmux.IsInsideTmux()
	if isInTmux {
		tmux.PrintErrorMessage(debugMode, tmuxErr)
		os.Exit(1)
	}
	logLevel := "info"
	if debugMode {
		logLevel = "debug"
	}
	cmd.InitializeMainSystem(logLevel)

	// Startup begin log
	startTime := time.Now()
	startupPhase := logger.BeginPhase("application_startup", map[string]interface{}{
		"debug_mode": debugMode,
		"log_level":  logLevel,
		"args":       args,
	})

	sessionName, _, err := cmd.ParseArguments(args)
	if err != nil && err.Error() == "debug flag processed by main" {
		filteredArgs := []string{}
		for _, arg := range args {
			if arg != "--debug" && arg != "-d" {
				filteredArgs = append(filteredArgs, arg)
			}
		}
		sessionName, _, err = cmd.ParseArguments(filteredArgs)
	}
	if err != nil {
		logger.LogStartupError("argument_parsing", err, nil)
		startupPhase.CompleteWithError(err)
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if sessionName == "" {
		sessionErr := fmt.Errorf("session name not specified")
		logger.LogStartupError("session_validation", sessionErr, nil)
		startupPhase.CompleteWithError(sessionErr)
		fmt.Println("❌ Error: Please specify a session name")
		cmd.ShowUsage()
		os.Exit(1)
	}
	if !cmd.IsValidSessionName(sessionName) {
		validationErr := fmt.Errorf("invalid session name: %s", sessionName)
		logger.LogStartupError("session_validation", validationErr, nil)
		startupPhase.CompleteWithError(validationErr)
		fmt.Println("❌ Error: Invalid session name")
		os.Exit(1)
	}

	// System launch phase
	logger.LogTmuxSetup(sessionName, 6, map[string]interface{}{
		"session_name": sessionName,
	})

	if err := cmd.LaunchSystem(sessionName); err != nil {
		logger.LogStartupError("system_launch", err, nil)
		startupPhase.CompleteWithError(err)
		_, _ = fmt.Fprintf(os.Stderr, "Launch error: %v\n", err)
		os.Exit(1)
	}

	// Startup complete log
	totalTime := time.Since(startTime)
	logger.LogStartupComplete(totalTime, map[string]interface{}{
		"session_name": sessionName,
		"success":      true,
	})
	startupPhase.Complete()
}
