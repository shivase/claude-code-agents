package internal

// Constant definitions
const (
	IntegratedSessionPaneCount = 6

	ClearDelay           = 400
	AdditionalClearDelay = 200
	MessageDelay         = 300
	ExecuteDelay         = 500

	AgentPO      = "po"
	AgentManager = "manager"
	AgentDev1    = "dev1"
	AgentDev2    = "dev2"
	AgentDev3    = "dev3"
	AgentDev4    = "dev4"
)

type Agent struct {
	Name        string
	Description string
}

type Session struct {
	Name  string
	Type  string
	Panes int
}

type SessionManager struct {
	sessions []Session
}

type MessageSender struct {
	SessionName  string
	Agent        string
	Message      string
	ResetContext bool
}

var AvailableAgents = []Agent{
	{AgentPO, "Product Owner (Product Manager)"},
	{AgentManager, "Project Manager (Flexible team management)"},
	{AgentDev1, "Execution Agent 1 (Flexible role assignment)"},
	{AgentDev2, "Execution Agent 2 (Flexible role assignment)"},
	{AgentDev3, "Execution Agent 3 (Flexible role assignment)"},
	{AgentDev4, "Execution Agent 4 (Flexible role assignment)"},
}

var ValidAgentNames = map[string]bool{
	AgentPO: true, AgentManager: true, AgentDev1: true,
	AgentDev2: true, AgentDev3: true, AgentDev4: true,
}
