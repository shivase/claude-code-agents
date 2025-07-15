package tmux

import "time"

// TmuxManagerInterface interface for tmux operation management
type TmuxManagerInterface interface {
	// SessionExists checks if a session exists
	SessionExists(sessionName string) bool
	// ListSessions retrieves session list
	ListSessions() ([]string, error)
	// CreateSession creates a new session
	CreateSession(sessionName string) error
	// KillSession deletes a session
	KillSession(sessionName string) error
	// AttachSession attaches to a session
	AttachSession(sessionName string) error
	// CreateIntegratedLayout creates integrated monitoring screen layout
	CreateIntegratedLayout(sessionName string, devCount int) error
	// CreateIndividualLayout creates individual session layout
	CreateIndividualLayout(sessionName string) error
	// SplitWindow splits a window
	SplitWindow(target, direction string) error
	// RenameWindow renames a window
	RenameWindow(sessionName, windowName string) error
	// AdjustPaneSizes adjusts pane sizes
	AdjustPaneSizes(sessionName string, devCount int) error
	// SetPaneTitles sets pane titles
	SetPaneTitles(sessionName string, devCount int) error
	// GetPaneCount retrieves pane count
	GetPaneCount(sessionName string) (int, error)
	// GetPaneList retrieves pane list
	GetPaneList(sessionName string) ([]string, error)
	// SendKeysToPane sends keys to a pane
	SendKeysToPane(sessionName, pane, keys string) error
	// SendKeysWithEnter sends keys to a pane with Enter
	SendKeysWithEnter(sessionName, pane, keys string) error
	// GetAITeamSessions retrieves AI team related sessions
	GetAITeamSessions(expectedPaneCount int) (map[string][]string, error)
	// FindDefaultAISession finds default AI session
	FindDefaultAISession(expectedPaneCount int) (string, error)
	// DetectActiveAISession detects active AI session
	DetectActiveAISession(expectedPaneCount int) (string, string, error)
	// DeleteAITeamSessions deletes AI team related sessions
	DeleteAITeamSessions(sessionName string, devCount int) error
	// WaitForPaneReady waits for pane to be ready
	WaitForPaneReady(sessionName, pane string, timeout time.Duration) error
	// GetSessionInfo retrieves session information
	GetSessionInfo(sessionName string) (map[string]interface{}, error)
	// SendInstructionToPaneWithConfig sends instruction file using configuration
	SendInstructionToPaneWithConfig(sessionName, pane, agent, instructionsDir string, config interface{}) error
}
