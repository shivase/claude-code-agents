package internal

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// Tmux related utility functions

func GetTmuxSessions() ([]Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tmux session list: %v", err)
	}

	var sessions []Session
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			sessions = append(sessions, Session{Name: line})
		}
	}

	return sessions, nil
}

func HasSession(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

func GetPaneCount(sessionName string) (int, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get pane count: %v", err)
	}
	return len(strings.Split(strings.TrimSpace(string(output)), "\n")), nil
}

func GetPanes(sessionName string) ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_index}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pane list: %v", err)
	}

	var panes []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			panes = append(panes, line)
		}
	}
	return panes, nil
}

func ShowPanes(sessionName string) error {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "  Pane #{pane_index}: #{pane_title}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to display pane status: %v", err)
	}
	fmt.Print(string(output))
	return nil
}

func TmuxSendKeys(target, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys)
	return cmd.Run()
}

func DetectDefaultSession() (string, error) {
	sessions, err := GetTmuxSessions()
	if err != nil || len(sessions) == 0 {
		return "", fmt.Errorf("no tmux sessions found")
	}

	// Prioritize integrated monitoring screen sessions (6 panes)
	for _, session := range sessions {
		paneCount, err := GetPaneCount(session.Name)
		if err != nil {
			continue
		}
		if paneCount == IntegratedSessionPaneCount {
			return session.Name, nil
		}
	}

	// Detect potential AI sessions (numeric session names or short names)
	for _, session := range sessions {
		paneCount, err := GetPaneCount(session.Name)
		if err != nil {
			continue
		}
		// Check sessions with numeric names, short names, or AI-related keywords
		if paneCount >= 1 && (len(session.Name) <= 3 ||
			strings.Contains(session.Name, "ai") ||
			strings.Contains(session.Name, "claude") ||
			strings.Contains(session.Name, "agent")) {
			return session.Name, nil
		}
	}

	// Find base names for individual session mode
	individualSessions := map[string]bool{}
	re := regexp.MustCompile(`-(po|manager|dev[1-4])$`)
	for _, session := range sessions {
		if re.MatchString(session.Name) {
			baseName := re.ReplaceAllString(session.Name, "")
			individualSessions[baseName] = true
		}
	}

	if len(individualSessions) > 0 {
		var baseNames []string
		for baseName := range individualSessions {
			baseNames = append(baseNames, baseName)
		}
		sort.Strings(baseNames)
		return baseNames[0], nil
	}

	return "", fmt.Errorf("no AI agent related sessions found")
}
