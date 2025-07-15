package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shivase/cloud-code-agents/send-agent/internal"
)

// Command definition
var (
	rootCmd = &cobra.Command{
		Use:   "send-agent [agent] [message]",
		Short: "🚀 AI Agent Message Sending System",
		Long: `🚀 AI Agent Message Sending System

A tool for sending messages to AI agents running on tmux sessions.
Supports both integrated monitoring screen and individual session modes.

Available agents:
  po      - Product Owner (Product Manager)
  manager - Project Manager (Flexible team management)
  dev1    - Execution Agent 1 (Flexible role assignment)
  dev2    - Execution Agent 2 (Flexible role assignment)
  dev3    - Execution Agent 3 (Flexible role assignment)
  dev4    - Execution Agent 4 (Flexible role assignment)`,
		Example: `  send-agent --session myproject manager "Please start a new project"
  send-agent --session ai-team dev1 "[As Marketing Lead] Please conduct market research"
  send-agent --reset dev1 "[As Data Analyst] Please create a report"
  send-agent manager "message"  (use default session)
  send-agent list myproject      (list agents in myproject session)
  send-agent list-sessions       (show all sessions)`,
		Args: cobra.ExactArgs(2),
		RunE: executeMainCommand,
	}

	listCmd = &cobra.Command{
		Use:   "list [session-name]",
		Short: "Display list of agents in specified session",
		Args:  cobra.ExactArgs(1),
		RunE:  executeListCommand,
	}

	listSessionsCmd = &cobra.Command{
		Use:   "list-sessions",
		Short: "Display list of all available sessions",
		RunE:  executeListSessionsCommand,
	}
)

func init() {
	rootCmd.Flags().StringP("session", "s", "", "Use specified session name")
	rootCmd.Flags().BoolP("reset", "r", false, "Clear previous role definition and send new instruction")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(listSessionsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
}

// Command execution functions
func executeMainCommand(cmd *cobra.Command, args []string) error {
	agent := args[0]
	message := args[1]
	sessionName, _ := cmd.Flags().GetString("session")
	resetContext, _ := cmd.Flags().GetBool("reset")

	if !internal.IsValidAgent(agent) {
		return fmt.Errorf("invalid agent name '%s'", agent)
	}

	if sessionName == "" {
		detectedSession, err := internal.DetectDefaultSession()
		if err != nil {
			return fmt.Errorf("no available AI agent sessions found\n💡 List sessions: %s list-sessions\n💡 Create new session: start-ai-agent [session-name]", cmd.Root().Name())
		}
		sessionName = detectedSession
		fmt.Printf("🔍 Using default session '%s'\n", sessionName)
	}

	sender := &internal.MessageSender{
		SessionName:  sessionName,
		Agent:        agent,
		Message:      message,
		ResetContext: resetContext,
	}

	return sender.Send()
}

func executeListCommand(cmd *cobra.Command, args []string) error {
	sessionName := args[0]
	manager := &internal.SessionManager{}
	return manager.ShowAgentsForSession(sessionName)
}

func executeListSessionsCommand(cmd *cobra.Command, args []string) error {
	manager := &internal.SessionManager{}
	return manager.ListAllSessions()
}
