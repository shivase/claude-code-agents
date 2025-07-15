package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Pre-compiled regular expressions (performance optimization)
var (
	agentSessionRegex = regexp.MustCompile(`-(po|manager|dev\d+)$`)
)

// TmuxManagerImpl manages tmux operations
type TmuxManagerImpl struct {
	sessionName string
	layout      string
}

// NewTmuxManager creates a new tmux manager
func NewTmuxManager(sessionName string) *TmuxManagerImpl {
	return &TmuxManagerImpl{
		sessionName: sessionName,
		layout:      "integrated", // "integrated" or "individual"
	}
}

// SessionExists checks if a session exists
func (tm *TmuxManagerImpl) SessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// ListSessions retrieves session list
func (tm *TmuxManagerImpl) ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var sessions []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			sessions = append(sessions, strings.TrimSpace(line))
		}
	}

	return sessions, nil
}

// CreateSession creates a new session
func (tm *TmuxManagerImpl) CreateSession(sessionName string) error {
	if tm.SessionExists(sessionName) {
		return fmt.Errorf("session %s already exists", sessionName)
	}

	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session %s: %w", sessionName, err)
	}

	return nil
}

// KillSession deletes a session
func (tm *TmuxManagerImpl) KillSession(sessionName string) error {
	if !tm.SessionExists(sessionName) {
		return nil // No error if session doesn't exist
	}

	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session %s: %w", sessionName, err)
	}

	return nil
}

// AttachSession attaches to a session
func (tm *TmuxManagerImpl) AttachSession(sessionName string) error {
	if !tm.SessionExists(sessionName) {
		return fmt.Errorf("session %s does not exist", sessionName)
	}

	// Execute tmux attach-session (non-interactively)
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Check session state on connection error
		if tm.SessionExists(sessionName) {
			log.Warn().Str("session", sessionName).Err(err).Msg("Session exists but attach failed")
			return fmt.Errorf("session %s exists but attach failed: %w", sessionName, err)
		}
		return fmt.Errorf("failed to attach to session %s: %w", sessionName, err)
	}

	return nil
}

// CreateIntegratedLayout creates integrated monitoring screen layout (supports dynamic dev count)
func (tm *TmuxManagerImpl) CreateIntegratedLayout(sessionName string, devCount int) error {
	// Create session if it doesn't exist
	if !tm.SessionExists(sessionName) {
		if err := tm.CreateSession(sessionName); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Set window name
	if err := tm.RenameWindow(sessionName, sessionName); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}

	// Create dynamic pane configuration (PO + Manager + Dev count)
	totalPanes := 2 + devCount

	// Add a small delay for split timing
	sleep := func() {
		time.Sleep(50 * time.Millisecond)
	}

	// Create layout with PO/Manager on left, Dev on right
	// 1. Split first pane horizontally (left | right)
	if err := tm.SplitWindow(sessionName, "-h"); err != nil {
		return fmt.Errorf("failed to split window horizontally: %w", err)
	}
	sleep()

	// 2. Split left side (pane 1) vertically (PO | Manager)
	if err := tm.SplitWindow(sessionName+":1.1", "-v"); err != nil {
		return fmt.Errorf("failed to split left pane vertically: %w", err)
	}
	sleep()

	// 3. Split right side (pane 3) for developers
	// First developer uses pane 3
	// For subsequent developers, always split the first right pane (pane 3)
	for i := 2; i <= devCount; i++ {
		// Always split pane 3 to maintain equal spacing
		target := fmt.Sprintf("%s:1.3", sessionName)
		if err := tm.SplitWindow(target, "-v"); err != nil {
			return fmt.Errorf("failed to split dev pane %d: %w", i, err)
		}
		sleep()
	}

	// Adjust pane sizes
	if err := tm.AdjustPaneSizes(sessionName, devCount); err != nil {
		return fmt.Errorf("failed to adjust pane sizes: %w", err)
	}

	// Set pane titles
	if err := tm.SetPaneTitles(sessionName, devCount); err != nil {
		return fmt.Errorf("failed to set pane titles: %w", err)
	}

	log.Info().Str("session", sessionName).Int("dev_count", devCount).Int("total_panes", totalPanes).Msg("Dynamic integrated layout created successfully")
	return nil
}

// SetupClaudeInPanes starts Claude CLI automatically and sends instructions in each pane
func (tm *TmuxManagerImpl) SetupClaudeInPanes(sessionName string, claudeCLIPath string, instructionsDir string, devCount int) error {

	// Dynamic pane configuration map (pane number → agent name)
	paneAgentMap := make(map[string]string)
	paneAgentMap["1"] = "po"
	paneAgentMap["2"] = "manager"
	for i := 1; i <= devCount; i++ {
		paneAgentMap[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("dev%d", i)
	}

	for pane, agent := range paneAgentMap {

		// Start Claude CLI in each pane
		if err := tm.startClaudeInPane(sessionName, pane, agent, claudeCLIPath); err != nil {
			log.Error().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to start Claude CLI in pane")
			return fmt.Errorf("failed to start Claude CLI in pane %s (%s): %w", pane, agent, err)
		}

		// Set 5 second interval between pane startups
		time.Sleep(5 * time.Second)
	}

	// Wait for Claude CLI startup to complete
	time.Sleep(2 * time.Second)

	// Send instruction files
	for pane, agent := range paneAgentMap {
		if err := tm.sendInstructionToPane(sessionName, pane, agent, instructionsDir); err != nil {
			log.Warn().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to send instruction to pane (non-critical)")
			// Instruction send failure is warning level (can continue)
		}

		// Set 2 second interval between instruction sends
		time.Sleep(2 * time.Second)
	}

	return nil
}

// SetupClaudeInPanesWithConfig starts Claude CLI automatically and sends instructions in each pane using configuration
func (tm *TmuxManagerImpl) SetupClaudeInPanesWithConfig(sessionName string, claudeCLIPath string, instructionsDir string, config interface{}, devCount int) error {
	// Get role-specific instructions files using TeamConfig interface
	type InstructionConfig interface {
		GetPOInstructionFile() string
		GetManagerInstructionFile() string
		GetDevInstructionFile() string
	}

	// Check if config implements InstructionConfig
	var instructionConfig InstructionConfig
	if ic, ok := config.(InstructionConfig); ok {
		instructionConfig = ic
	}

	// Dynamic pane configuration map (pane number → agent name)
	paneAgentMap := make(map[string]string)
	paneAgentMap["1"] = "po"
	paneAgentMap["2"] = "manager"
	for i := 1; i <= devCount; i++ {
		paneAgentMap[fmt.Sprintf("%d", i+2)] = fmt.Sprintf("dev%d", i)
	}

	for pane, agent := range paneAgentMap {
		// Start Claude CLI in each pane
		if err := tm.startClaudeInPane(sessionName, pane, agent, claudeCLIPath); err != nil {
			log.Error().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to start Claude CLI in pane")
			return fmt.Errorf("failed to start Claude CLI in pane %s (%s): %w", pane, agent, err)
		}

		// Set 5 second interval between pane startups
		time.Sleep(5 * time.Second)
	}

	// Wait for Claude CLI startup to complete
	time.Sleep(2 * time.Second)

	// Send instruction files
	for pane, agent := range paneAgentMap {
		if err := tm.SendInstructionToPaneWithConfig(sessionName, pane, agent, instructionsDir, instructionConfig); err != nil {
			log.Warn().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Err(err).Msg("Failed to send instruction to pane (non-critical)")
			// Instruction send failure is warning level (can continue)
		}

		// Set 2 second interval between instruction sends
		time.Sleep(2 * time.Second)
	}

	return nil
}

// startClaudeInPane starts Claude CLI in specified pane
func (tm *TmuxManagerImpl) startClaudeInPane(sessionName, pane, _ /* agent */, claudeCLIPath string) error {
	// Check if pane exists
	if err := tm.WaitForPaneReady(sessionName, pane, 5*time.Second); err != nil {
		return fmt.Errorf("pane %s not ready: %w", pane, err)
	}

	// Create Claude CLI start command
	claudeCommand := fmt.Sprintf("%s --dangerously-skip-permissions", claudeCLIPath)

	// Send Claude CLI start command to pane
	if err := tm.SendKeysWithEnter(sessionName, pane, claudeCommand); err != nil {
		return fmt.Errorf("failed to send Claude CLI command to pane: %w", err)
	}

	return nil
}

// sendInstructionToPane sends instruction file to specified pane (enhanced version)
func (tm *TmuxManagerImpl) sendInstructionToPane(sessionName, pane, agent, instructionsDir string) error {
	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("📤 Starting instruction sending")

	// Determine instruction file path
	var instructionFile string
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

	log.Info().Str("instruction_file", instructionFile).Msg("📁 Instruction file path determined")

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

	log.Info().Str("instruction_file", instructionFile).Int64("file_size", fileInfo.Size()).Msg("✅ File existence verified")

	// Wait for Claude CLI to be ready (enhanced version)
	if err := tm.waitForClaudeReady(sessionName, pane, 10*time.Second); err != nil {
		log.Warn().Str("session", sessionName).Str("pane", pane).Err(err).Msg("⚠️ Claude CLI readiness wait timeout (continuing)")
	}

	// Send instruction file using cat command (with retry functionality)
	catCommand := fmt.Sprintf("cat \"%s\"", instructionFile)

	for attempt := 1; attempt <= 3; attempt++ {
		log.Info().Str("command", catCommand).Int("attempt", attempt).Msg("📋 Sending cat command")

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
		log.Info().Int("attempt", attempt).Msg("✅ Cat command sent successfully")
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

	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("✅ Instruction sending completed")
	return nil
}

// CreateIndividualLayout creates individual session layout
func (tm *TmuxManagerImpl) CreateIndividualLayout(sessionName string, devCount int) error {
	agents := []string{"po", "manager"}
	for i := 1; i <= devCount; i++ {
		agents = append(agents, fmt.Sprintf("dev%d", i))
	}

	for _, agent := range agents {
		agentSession := fmt.Sprintf("%s-%s", sessionName, agent)

		if err := tm.CreateSession(agentSession); err != nil {
			return fmt.Errorf("failed to create session for %s: %w", agent, err)
		}

		// Set window name
		if err := tm.RenameWindow(agentSession, agentSession); err != nil {
			return fmt.Errorf("failed to rename window for %s: %w", agent, err)
		}
	}

	log.Info().Str("session", sessionName).Msg("Individual layout created successfully")
	return nil
}

// SplitWindow splits a window
func (tm *TmuxManagerImpl) SplitWindow(target, direction string) error {
	cmd := exec.Command("tmux", "split-window", direction, "-t", target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux command failed: split-window %s -t %s (output: %s)", direction, target, string(output))
	}
	return nil
}

// RenameWindow renames a window
func (tm *TmuxManagerImpl) RenameWindow(sessionName, windowName string) error {
	cmd := exec.Command("tmux", "rename-window", "-t", sessionName, windowName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rename window: %w", err)
	}
	return nil
}

// AdjustPaneSizes adjusts pane sizes (supports dynamic dev count with equal spacing)
func (tm *TmuxManagerImpl) AdjustPaneSizes(sessionName string, devCount int) error {
	totalPanes := 2 + devCount // PO + Manager + Dev count

	// Protection against division by zero when devCount=0
	if devCount <= 0 {
		log.Warn().Str("session", sessionName).Int("dev_count", devCount).Msg("Skipping pane size adjustment as dev count is 0 or less")
		return fmt.Errorf("devCount must be greater than 0, got: %d", devCount)
	}

	// Set left side (PO/Manager) to 50%, right side (Dev) to 50%
	leftSidePercentage := 50

	log.Info().Str("session", sessionName).Int("dev_count", devCount).Int("total_panes", totalPanes).Msg("Starting equal spacing pane division")

	// Get window size
	windowWidth, windowHeight, err := tm.getWindowSize(sessionName)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get window size, using default values")
		windowWidth = 120
		windowHeight = 40
	}

	// Check window size validity
	if windowHeight <= 0 {
		log.Warn().Int("window_height", windowHeight).Msg("Invalid window height, using default value")
		windowHeight = 40
	}

	// Calculate left side width (50% of total)
	leftWidth := (windowWidth * leftSidePercentage) / 100

	// 1. Adjust left-right split (left 50%, right 50%)
	time.Sleep(100 * time.Millisecond)
	leftCmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.1", sessionName), "-x", fmt.Sprintf("%d", leftWidth)) // #nosec G204
	if err := leftCmd.Run(); err != nil {
		log.Warn().Str("pane", "1").Int("width", leftWidth).Err(err).Msg("Failed to adjust left pane")
	}

	// 2. Adjust left side vertical split (PO/Manager 50% each)
	time.Sleep(100 * time.Millisecond)
	poHeight := windowHeight / 2
	poCmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.1", sessionName), "-y", fmt.Sprintf("%d", poHeight)) // #nosec G204
	if err := poCmd.Run(); err != nil {
		log.Warn().Str("pane", "PO").Int("height", poHeight).Err(err).Msg("Failed to adjust PO/Manager split")
	}

	// 3. Adjust right side developer panes with equal spacing (division by zero protected)
	// devCount is already confirmed to be greater than 0
	devPaneHeight := windowHeight / devCount

	// Set height for each developer pane
	for i := 1; i <= devCount; i++ {
		paneNumber := i + 2 // After PO(1), Manager(2), starts from 3

		// Set each pane height equally
		time.Sleep(100 * time.Millisecond)
		cmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.%d", sessionName, paneNumber), "-y", fmt.Sprintf("%d", devPaneHeight)) // #nosec G204
		if err := cmd.Run(); err != nil {
			log.Warn().Str("pane", fmt.Sprintf("%d", paneNumber)).Int("height", devPaneHeight).Err(err).Msg("Failed to resize pane with equal spacing")
		} else {
			log.Debug().Str("pane", fmt.Sprintf("%d", paneNumber)).Int("height", devPaneHeight).Msg("Successfully resized pane with equal spacing")
		}
	}

	// 4. Finally readjust left-right width (maintain 50% each)
	time.Sleep(100 * time.Millisecond)
	finalLeftCmd := exec.Command("tmux", "resize-pane", "-t", fmt.Sprintf("%s:1.1", sessionName), "-x", fmt.Sprintf("%d", leftWidth)) // #nosec G204
	if err := finalLeftCmd.Run(); err != nil {
		log.Warn().Err(err).Msg("Failed to perform final left-right width adjustment")
	}

	log.Info().Str("session", sessionName).Int("dev_count", devCount).Msg("Equal spacing pane division completed")
	return nil
}

// SetPaneTitles sets pane titles (supports dynamic dev count)
func (tm *TmuxManagerImpl) SetPaneTitles(sessionName string, devCount int) error {
	// Configure to display pane titles
	cmd := exec.Command("tmux", "set-option", "-t", sessionName, "pane-border-status", "top")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set pane border status: %w", err)
	}

	cmd = exec.Command("tmux", "set-option", "-t", sessionName, "pane-border-format", "#T")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set pane border format: %w", err)
	}

	// Disable automatic rename
	cmd = exec.Command("tmux", "set-window-option", "-t", sessionName, "automatic-rename", "off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable automatic rename: %w", err)
	}

	cmd = exec.Command("tmux", "set-window-option", "-t", sessionName, "allow-rename", "off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable allow rename: %w", err)
	}

	// Set title for each pane (supports dynamic dev count)
	titles := make(map[string]string)
	titles["1"] = "PO"      // Top left
	titles["2"] = "Manager" // Bottom left

	// Dynamically set developer titles
	for i := 1; i <= devCount; i++ {
		paneNumber := fmt.Sprintf("%d", i+2)
		titles[paneNumber] = fmt.Sprintf("Dev%d", i)
	}

	for pane, title := range titles {
		target := fmt.Sprintf("%s:1.%s", sessionName, pane)
		cmd = exec.Command("tmux", "select-pane", "-t", target, "-T", title) // #nosec G204
		if err := cmd.Run(); err != nil {
			log.Warn().Str("pane", target).Str("title", title).Err(err).Msg("Failed to set pane title")
		}
	}

	return nil
}

// GetPaneCount retrieves pane count
func (tm *TmuxManagerImpl) GetPaneCount(sessionName string) (int, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName)
	output, err := cmd.Output()
	if err != nil {
		log.Debug().Str("session", sessionName).Err(err).Msg("Failed to get pane count")
		return 0, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	paneCount := len(lines)

	log.Debug().Str("session", sessionName).Int("pane_count", paneCount).Msg("GetPaneCount result")
	return paneCount, nil
}

// GetPaneList retrieves pane list
func (tm *TmuxManagerImpl) GetPaneList(sessionName string) ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_index}:#{pane_title}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var panes []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			panes = append(panes, strings.TrimSpace(line))
		}
	}

	return panes, nil
}

// SendKeysToPane sends keys to a pane
func (tm *TmuxManagerImpl) SendKeysToPane(sessionName, pane, keys string) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys) // #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys to pane %s: %w", target, err)
	}
	return nil
}

// SendKeysWithEnter sends keys to a pane with Enter
func (tm *TmuxManagerImpl) SendKeysWithEnter(sessionName, pane, keys string) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "C-m") // #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys with enter to pane %s: %w", target, err)
	}
	return nil
}

// GetAITeamSessions retrieves AI team related sessions
func (tm *TmuxManagerImpl) GetAITeamSessions(expectedPaneCount int) (map[string][]string, error) {
	sessions, err := tm.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	result := map[string][]string{
		"integrated": {},
		"individual": {},
		"other":      {},
	}

	for _, session := range sessions {
		// Determine integrated monitoring screen (dynamic pane count configuration)
		paneCount, err := tm.GetPaneCount(session)
		log.Debug().Str("session", session).Int("pane_count", paneCount).Int("expected_pane_count", expectedPaneCount).Err(err).Msg("Session analysis")

		switch {
		case err == nil && paneCount == expectedPaneCount:
			result["integrated"] = append(result["integrated"], session)
			log.Debug().Str("session", session).Msg("Added as integrated session")
		case agentSessionRegex.MatchString(session):
			// Determine individual session layout
			baseName := agentSessionRegex.ReplaceAllString(session, "")
			if !containsString(result["individual"], baseName) {
				result["individual"] = append(result["individual"], baseName)
				log.Debug().Str("session", session).Str("base_name", baseName).Msg("Added as individual session")
			}
		default:
			// Check for numeric-only sessions (like "1") or potential existing AI sessions
			if err == nil && paneCount >= 1 {
				// Numeric-only or short session names are potential AI sessions
				if len(session) <= 3 || strings.Contains(session, "ai") || strings.Contains(session, "claude") {
					result["integrated"] = append(result["integrated"], session)
					log.Debug().Str("session", session).Msg("Added as potential AI session")
				} else {
					result["other"] = append(result["other"], session)
					log.Debug().Str("session", session).Msg("Added as other session")
				}
			} else {
				result["other"] = append(result["other"], session)
				log.Debug().Str("session", session).Msg("Added as other session")
			}
		}
	}

	return result, nil
}

// FindDefaultAISession finds default AI session
func (tm *TmuxManagerImpl) FindDefaultAISession(expectedPaneCount int) (string, error) {
	aiSessions, err := tm.GetAITeamSessions(expectedPaneCount)
	if err != nil {
		return "", fmt.Errorf("failed to get AI team sessions: %w", err)
	}

	// Prioritize integrated monitoring screen sessions
	if len(aiSessions["integrated"]) > 0 {
		return aiSessions["integrated"][0], nil
	}

	// For individual session layout
	if len(aiSessions["individual"]) > 0 {
		return aiSessions["individual"][0], nil
	}

	// Look for sessions even if no AI sessions found
	sessions, err := tm.ListSessions()
	if err != nil {
		return "ai-teams", err
	}

	// Detect potential AI sessions (numeric-only or short session names)
	for _, session := range sessions {
		paneCount, err := tm.GetPaneCount(session)
		if err != nil {
			continue
		}
		// Check for numeric-only session names, short names, or AI-related keywords
		if paneCount >= 1 && (len(session) <= 3 ||
			strings.Contains(session, "ai") ||
			strings.Contains(session, "claude") ||
			strings.Contains(session, "agent")) {
			return session, nil
		}
	}

	// Finally return default value
	return "ai-teams", nil
}

// DetectActiveAISession detects active AI session
func (tm *TmuxManagerImpl) DetectActiveAISession(expectedPaneCount int) (string, string, error) {
	aiSessions, err := tm.GetAITeamSessions(expectedPaneCount)
	if err != nil {
		return "", "", fmt.Errorf("failed to get AI team sessions: %w", err)
	}

	// Prioritize integrated monitoring screen sessions
	if len(aiSessions["integrated"]) > 0 {
		return aiSessions["integrated"][0], "integrated", nil
	}

	// For individual session layout
	if len(aiSessions["individual"]) > 0 {
		return aiSessions["individual"][0], "individual", nil
	}

	// When no AI sessions found
	return "", "", fmt.Errorf("no active AI sessions found")
}

// DeleteAITeamSessions deletes AI team related sessions
func (tm *TmuxManagerImpl) DeleteAITeamSessions(sessionName string, devCount int) error {
	log.Info().Str("session", sessionName).Msg("Deleting AI team sessions")

	deletedCount := 0

	// For integrated monitoring screen
	expectedPaneCount := 2 + devCount
	if tm.SessionExists(sessionName) {
		paneCount, err := tm.GetPaneCount(sessionName)
		switch {
		case err == nil && paneCount == expectedPaneCount:
			log.Info().Str("session", sessionName).Int("pane_count", paneCount).Msg("Deleting integrated session")
			if err := tm.KillSession(sessionName); err != nil {
				return fmt.Errorf("failed to delete integrated session: %w", err)
			}
			deletedCount++
		default:
			log.Info().Str("session", sessionName).Int("pane_count", paneCount).Msg("Deleting general session")
			if err := tm.KillSession(sessionName); err != nil {
				return fmt.Errorf("failed to delete general session: %w", err)
			}
			deletedCount++
		}
	}

	// For individual session layout
	agents := []string{"po", "manager"}
	for i := 1; i <= devCount; i++ {
		agents = append(agents, fmt.Sprintf("dev%d", i))
	}
	for _, agent := range agents {
		agentSession := fmt.Sprintf("%s-%s", sessionName, agent)
		if tm.SessionExists(agentSession) {
			log.Info().Str("session", agentSession).Msg("Deleting individual session")
			if err := tm.KillSession(agentSession); err != nil {
				return fmt.Errorf("failed to delete individual session %s: %w", agentSession, err)
			}
			deletedCount++
		}
	}

	if deletedCount == 0 {
		return fmt.Errorf("no sessions found for %s", sessionName)
	}

	log.Info().Str("session", sessionName).Int("deleted_count", deletedCount).Msg("AI team sessions deleted")
	return nil
}

// containsString checks if string exists in slice
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// WaitForPaneReady waits for pane to be ready
func (tm *TmuxManagerImpl) WaitForPaneReady(sessionName, pane string, timeout time.Duration) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	start := time.Now()

	for time.Since(start) < timeout {
		// Check if pane exists
		cmd := exec.Command("tmux", "list-panes", "-t", sessionName)
		output, err := cmd.Output()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Check if pane exists
		if strings.Contains(string(output), pane) {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for pane %s to be ready", target)
}

// waitForClaudeReady waits for Claude CLI to be ready (new implementation)
func (tm *TmuxManagerImpl) waitForClaudeReady(sessionName, pane string, timeout time.Duration) error {
	target := fmt.Sprintf("%s:1.%s", sessionName, pane)
	start := time.Now()

	log.Info().Str("target", target).Dur("timeout", timeout).Msg("🔄 Starting Claude CLI readiness wait")

	for time.Since(start) < timeout {
		// Get pane content and check if Claude CLI is ready
		cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p") // #nosec G204
		output, err := cmd.Output()
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		paneContent := string(output)

		// Check typical output patterns when Claude CLI has started
		if strings.Contains(paneContent, "claude") ||
			strings.Contains(paneContent, ">") ||
			strings.Contains(paneContent, "$") ||
			len(strings.TrimSpace(paneContent)) > 10 {
			log.Info().Str("target", target).Msg("✅ Claude CLI readiness detected")
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Warn().Str("target", target).Dur("elapsed", time.Since(start)).Msg("⚠️ Claude CLI readiness wait timeout")
	return fmt.Errorf("timeout waiting for Claude CLI to be ready in pane %s", target)
}

// GetSessionInfo retrieves session information
func (tm *TmuxManagerImpl) GetSessionInfo(sessionName string, expectedPaneCount int) (map[string]interface{}, error) {
	if !tm.SessionExists(sessionName) {
		return nil, fmt.Errorf("session %s does not exist", sessionName)
	}

	info := map[string]interface{}{
		"name":   sessionName,
		"exists": true,
	}

	// Get pane count
	paneCount, err := tm.GetPaneCount(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pane count: %w", err)
	}
	info["pane_count"] = paneCount

	// Get pane list
	panes, err := tm.GetPaneList(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pane list: %w", err)
	}
	info["panes"] = panes

	// Determine session type
	if paneCount == expectedPaneCount {
		info["type"] = "integrated"
	} else {
		info["type"] = "general"
	}

	return info, nil
}

// getWindowSize retrieves window size
func (tm *TmuxManagerImpl) getWindowSize(sessionName string) (int, int, error) {
	// Get width
	widthCmd := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{window_width}")
	widthOutput, err := widthCmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get window width: %w", err)
	}

	width, err := strconv.Atoi(strings.TrimSpace(string(widthOutput)))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse window width: %w", err)
	}

	// Get height
	heightCmd := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{window_height}")
	heightOutput, err := heightCmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get window height: %w", err)
	}

	height, err := strconv.Atoi(strings.TrimSpace(string(heightOutput)))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse window height: %w", err)
	}

	log.Debug().Str("session", sessionName).Int("width", width).Int("height", height).Msg("Window size retrieved")
	return width, height, nil
}
