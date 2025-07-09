# Claude Code Agents - Integrated Multi-Agent System

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-green.svg)](https://github.com/shivase/cloud-code-agents/actions)
[![Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen.svg)](https://github.com/shivase/cloud-code-agents/actions)

A comprehensive integrated multi-agent system specialized for Claude Code environments, implemented in Go for high performance.

📖 [日本語README](../README.md)

## Overview

**Claude Code Agents** is an integrated multi-agent system designed to streamline AI development work in Claude Code environments. Multiple AI agents collaborate to accomplish tasks efficiently, providing an enterprise-grade solution.

## 🚀 Key Features

### Integrated Multi-Agent Management
- **Manager-Agent** system (CEO, Manager, Developer)
- **Real-time inter-agent communication** (send-agent functionality)
- **Session management** for work continuity
- **Automatic load balancing** for efficient task distribution

### Advanced Execution Control
- **PTY control** for complete terminal emulation
- **Graceful shutdown** mechanisms for safe termination
- **Process monitoring** and automatic recovery
- **Resource management** (memory/CPU usage control)

### Enterprise-Ready
- **Secure authentication handling** (`~/.claude/settings.json`)
- **Structured logging** for detailed operation tracking
- **Configurable timeouts** and retry mechanisms
- **Concurrent execution control** for high performance

### Developer-Friendly
- **Rich command-line options** for flexible usage
- **JSON configuration files** for easy customization
- **Detailed error messages** and debugging information
- **Hot reload** for immediate configuration changes

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Claude Code Agents                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  Agent (CEO)    │  │ Agent (Manager) │  │Agent (Developer)│ │
│  │     🎯          │  │      📋         │  │       💻        │ │
│  │  ┌───────────┐  │  │  ┌───────────┐  │  │  ┌───────────┐  │ │
│  │  │   PTY     │  │  │  │   PTY     │  │  │  │   PTY     │  │ │
│  │  │ Terminal  │  │  │  │ Terminal  │  │  │  │ Terminal  │  │ │
│  │  └───────────┘  │  │  └───────────┘  │  │  └───────────┘  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Message System  │  │ Resource Monitor│  │ Health Checker  │ │
│  │  (send-agent)   │  │   (Memory/CPU)  │  │  (Auth/Claude)  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │Session Manager  │  │ Logger System   │  │ Signal Handler  │ │
│  │  (tmux/config)  │  │   (zerolog)     │  │  (Graceful)     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 📦 Installation

### Prerequisites
- Go 1.21 or later
- Claude Code CLI (latest version)
- tmux (for session management)

### Quick Start
```bash
# 1. Clone the repository
git clone https://github.com/shivase/cloud-code-agents.git
cd cloud-code-agents

# 2. Install dependencies and build
make setup

# 3. Install binaries
make install

# 4. Initialize configuration
make init-config
```

## 🎯 Usage

### Basic Usage

#### System Startup
```bash
# Start agent system
claude-code-agents start

# Check system status
claude-code-agents status

# List agents
claude-code-agents list

# Stop system
claude-code-agents stop
```

#### Agent Communication
```bash
# Send tasks to agents
send-agent ceo "Please start a new project"
send-agent manager "Please divide tasks among the development team"
send-agent dev1 "Please implement API endpoints"

# Specify a particular session
send-agent --session myproject dev1 "【As Frontend Lead】Please implement UI design"
```

### Available Agents

- **ceo** - Chief Executive Officer (Overall supervision)
- **manager** - Project Manager (Flexible team management)
- **dev1** - Execution Agent 1 (Flexible role assignment)
- **dev2** - Execution Agent 2 (Flexible role assignment)
- **dev3** - Execution Agent 3 (Flexible role assignment)
- **dev4** - Execution Agent 4 (Flexible role assignment)

### Advanced Usage
```bash
# Start with specific layout
claude-code-agents start --layout individual

# Debug mode
claude-code-agents --verbose start

# Custom configuration directory
claude-code-agents --config-dir /path/to/config start

# Session management
send-agent list-sessions
send-agent list myproject
```

## ⚙️ Configuration

### Configuration File: `~/.claude/agents-config.json`

```json
{
  "system": {
    "max_sessions": 10,
    "session_timeout": "30m",
    "health_check_interval": "30s",
    "working_dir": "/tmp/claude-agents",
    "auto_attach": true,
    "default_layout": "integrated"
  },
  "claude": {
    "cli_path": "/usr/local/bin/claude",
    "instructions_dir": "./configs/instructions",
    "auth_check_interval": "30m"
  },
  "logging": {
    "level": "info",
    "file": "./logs/agents.log",
    "structured": true
  },
  "performance": {
    "max_memory_mb": 1024,
    "max_cpu_percent": 80.0,
    "max_restart_attempts": 3
  }
}
```

### Environment Variables
```bash
export CLAUDE_AGENTS_LOG_LEVEL=debug
export CLAUDE_AGENTS_CONFIG_DIR=/path/to/config
export CLAUDE_AGENTS_VERBOSE=true
```

## 📁 Project Structure

```
cloud-code-agents/
├── cmd/
│   ├── manager/           # Agent management system
│   ├── send-agent/        # Agent communication system
│   └── claude-teams/      # Team management system
├── internal/
│   ├── auth/              # Authentication features
│   ├── config/            # Configuration management
│   ├── command/           # Command execution
│   ├── session/           # Session management
│   └── server/            # Server functionality
├── pkg/
│   ├── types/             # Common type definitions
│   └── agent/             # Agent functionality
├── configs/               # Configuration files
├── docs/                  # Documentation
├── logs/                  # Log files
├── go.mod                 # Go module definition
└── Makefile              # Build system
```

## 🔄 Comparison with Existing Solutions

| Feature | Traditional Scripts | Claude Code Agents | Improvement |
|---------|-------------------|------------------|-------------|
| **Multi-Agent** | Single execution | Multiple concurrent | 10x efficiency |
| **Process Management** | tmux+script | PTY+Go runtime | Stability & control |
| **Authentication** | Manual management | Auto-protection & recovery | Enhanced security |
| **Error Handling** | Simple logs | Structured logging+monitoring | Better operations |
| **Resource Control** | No limits | Memory/CPU monitoring | System stability |
| **Parallel Processing** | Sequential | Concurrent Goroutines | Performance boost |
| **Configuration** | Hard-coded | JSON configuration | Flexibility & maintainability |
| **Message Communication** | Manual input | Automated sending system | Efficiency improvement |

## 🛠️ Development

### Development Environment Setup
```bash
# Build development environment
make setup

# Run tests
make test          # Unit tests
make test-race     # Race condition detection
make test-coverage # Coverage measurement

# Code quality
make lint          # Static analysis
make fmt           # Format code
make vet           # Go vet
```

### CI/CD
```bash
# Local CI execution
make ci-local

# Release build
make release

# Package creation
make package
```

## 🔧 Troubleshooting

### Common Issues

#### Claude CLI Not Found
```bash
# Check path
which claude

# Specify path in configuration
{
  "claude": {
    "cli_path": "/usr/local/bin/claude"
  }
}
```

#### Authentication Errors
```bash
# Check authentication status
claude auth status

# Re-authenticate
claude auth login

# Verify settings file
ls -la ~/.claude/settings.json
```

#### Session Management Issues
```bash
# List sessions
send-agent list-sessions

# Stop problematic session
claude-code-agents stop

# Force reset
claude-code-agents start --reset
```

## 📊 Feature List

### System Management Features
- Agent startup/shutdown/monitoring
- Session management (integrated/individual)
- Resource monitoring and control
- Automatic recovery functionality

### Message Communication Features
- Inter-agent communication
- Automated message sending
- Message history logging
- Automatic session detection

### Configuration & Authentication Features
- Configuration file management
- Authentication status monitoring
- Configuration hot reload
- Environment variable support

## 🏆 Key Implementation Highlights

### 1. Thread-Safe Agent Management
```go
type Agent struct {
    ID          string
    Role        types.Role
    Session     string
    MessageChan chan types.Message
    mu          sync.RWMutex
    status      types.AgentStatus
}
```

### 2. Concurrent-Safe Data Structures
```go
type SafeMap struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

func (sm *SafeMap) Set(key string, value interface{}) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    sm.data[key] = value
}
```

### 3. Advanced PTY Control
```go
func startAgentWithPTY(agentID string) (*Agent, error) {
    cmd := exec.Command("claude", "--dangerously-skip-permissions")
    cmd.Env = append(os.Environ(), "TERM=xterm-256color")
    
    pty, err := pty.Start(cmd)
    if err != nil {
        return nil, fmt.Errorf("PTY start failed: %v", err)
    }
    
    return &Agent{
        ID:  agentID,
        PTY: pty,
        cmd: cmd,
    }, nil
}
```

### 4. Graceful Shutdown
```go
func (s *System) Shutdown() error {
    s.cancel() // Cancel all contexts
    
    // Wait for all agents to terminate
    for _, agent := range s.agents {
        if err := agent.Stop(); err != nil {
            s.logger.Error().Err(err).Msg("Agent stop failed")
        }
    }
    
    return nil
}
```

## 🤝 Community

- [GitHub Issues](https://github.com/shivase/cloud-code-agents/issues) - Bug reports & feature requests
- [Discussions](https://github.com/shivase/cloud-code-agents/discussions) - Q&A & discussions
- [Wiki](https://github.com/shivase/cloud-code-agents/wiki) - Detailed documentation

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details.

## 🙏 Acknowledgments

- [Claude Code](https://claude.ai/code) - Amazing AI development environment
- [creack/pty](https://github.com/creack/pty) - PTY control library
- [rs/zerolog](https://github.com/rs/zerolog) - High-performance logging library
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- Go community - Powerful development ecosystem

---

*Built with ❤️ for the Claude Code community*