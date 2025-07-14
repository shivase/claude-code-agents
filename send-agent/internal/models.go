package internal

// 定数定義
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
	{AgentPO, "プロダクトオーナー（製品責任者）"},
	{AgentManager, "プロジェクトマネージャー（柔軟なチーム管理）"},
	{AgentDev1, "実行エージェント1（柔軟な役割対応）"},
	{AgentDev2, "実行エージェント2（柔軟な役割対応）"},
	{AgentDev3, "実行エージェント3（柔軟な役割対応）"},
	{AgentDev4, "実行エージェント4（柔軟な役割対応）"},
}

var ValidAgentNames = map[string]bool{
	AgentPO: true, AgentManager: true, AgentDev1: true,
	AgentDev2: true, AgentDev3: true, AgentDev4: true,
}
