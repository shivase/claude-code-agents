# send-agent

[![CI](https://github.com/shivase/send-agent/workflows/CI/badge.svg)](https://github.com/shivase/send-agent/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/shivase/send-agent)](https://goreportcard.com/report/github.com/shivase/send-agent)
[![codecov](https://codecov.io/gh/shivase/send-agent/branch/main/graph/badge.svg)](https://codecov.io/gh/shivase/send-agent)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

🚀 AI Agent Message Sending System

A Go application for sending messages to AI agents running in tmux sessions. Supports both integrated monitoring screen and individual session modes.

📖 [日本語README](../README.md)

## Overview

This tool is a command-line utility for efficiently sending messages to AI agents running within tmux sessions. It manages multiple AI agents and automatically sends and executes messages to each agent.

## Available Agents

- **ceo** - Chief Executive Officer (Overall supervision)
- **manager** - Project Manager (Flexible team management)
- **dev1** - Execution Agent 1 (Flexible role assignment)
- **dev2** - Execution Agent 2 (Flexible role assignment)
- **dev3** - Execution Agent 3 (Flexible role assignment)
- **dev4** - Execution Agent 4 (Flexible role assignment)

## Features

- Unified management through integrated monitoring screen (6 panes)
- Agent management through individual session mode
- Automatic message sending and execution
- Message history logging
- Automatic session detection

## Installation

### Prerequisites

- Go 1.21 or later
- tmux

### Build and Install

```bash
# Clone the repository
git clone <repository-url>
cd send-agent

# Build and install
make install
```

## Usage

### Basic Usage

```bash
# Send message using default session
send-agent manager "Please start a new project"

# Send message specifying a particular session
send-agent --session myproject dev1 "【As Marketing Lead】Please conduct market research"

# Use short form to specify session
send-agent -s ai-team ceo "Please check the progress of the entire team"
```

### Session Management

```bash
# Display list of available sessions
send-agent list-sessions

# Display agent list for a specific session
send-agent list myproject
```

### Session Types

#### Integrated Monitoring Screen
- Manages AI agents in a 6-pane integrated screen
- Each agent is assigned to a fixed pane

#### Individual Session Mode
- Individual session for each agent
- Format: `<base-name>-<agent-name>`

## Configuration

### Directory Structure

```
send-agent/
├── main.go          # Main application
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
├── Makefile         # Build configuration
├── build/           # Built binaries
└── logs/            # Sending logs
    └── communication.log
```

### Configurable Constants

```go
const (
    IntegratedSessionPaneCount = 6    // Number of panes for integrated monitoring
    LogDir                     = "logs"
    LogFile                    = "communication.log"
    
    // Wait times for message sending (milliseconds)
    ClearDelay           = 400
    AdditionalClearDelay = 200
    MessageDelay         = 300
    ExecuteDelay         = 500
)
```

## Development

### Development Commands

```bash
# Development build (with race detector)
make dev-build

# Run tests
make test

# Format code
make fmt

# Release build
make release

# Show help
make help
```

### Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework

## Logging

Sent messages are automatically logged to `logs/communication.log`.

```
[2024-01-01 12:00:00] → manager: "Please start a new project"
[2024-01-01 12:05:00] → dev1: "【As Marketing Lead】Please conduct market research"
```

## Troubleshooting

### Common Issues

1. **Session not found**
   ```bash
   send-agent list-sessions
   ```

2. **Panes not recognized correctly**
   - Check the number of panes in the tmux session
   - Integrated monitoring screen must have 6 panes

3. **Message sending fails**
   - Ensure tmux session is running
   - Verify the target agent is valid

## License

MIT License - See [LICENSE](../LICENSE) file for details.

## Contributing

Pull requests and issue reports are welcome.

## Author

Please check the project metadata for developer information.