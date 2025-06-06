# logflow - Multi-Source Log Viewer

A TUI application for viewing logs from multiple sources simultaneously during development.

## Project Structure

```
logflow/
├── cmd/
│   └── logflow/
│       └── main.go
├── internal/
│   ├── ipc/
│   │   ├── server.go      # Unix socket server
│   │   ├── client.go      # Unix socket client  
│   │   └── protocol.go    # Message protocol
│   ├── ui/
│   │   ├── app.go         # Main TUI application
│   │   ├── layout.go      # Layout management
│   │   ├── pane.go        # Individual log panes
│   │   └── keybindings.go # Key handling
│   ├── sources/
│   │   ├── pipe.go        # Stdin pipe source
│   │   ├── docker.go      # Docker logs source
│   │   └── podman.go      # Podman logs source
│   └── log/
│       ├── entry.go       # Log entry types
│       ├── parser.go      # Log level parsing
│       └── buffer.go      # Log buffering
├── pkg/
│   └── types/
│       └── types.go       # Shared types
├── go.mod
├── go.sum
├── README.md
└── Makefile
```

## Installation

```bash
git clone https://github.com/Yriskit-ai/logflow
cd logflow
go mod tidy
go build -o bin/logflow cmd/logflow/main.go
```

## Usage

```bash
# Start the TUI dashboard
logflow

# In other terminals, pipe logs to dashboard
python app.py | logflow --source backend
npm run dev | logflow --source frontend
podman logs -f redis | logflow --source redis

# Or attach directly to containers
logflow --docker redis-container --source redis
logflow --podman postgres-dev --source db
```

## Key Features

- **Multi-pane viewing**: See logs from multiple sources simultaneously
- **Flexible layouts**: Horizontal, vertical, and auto-grid layouts
- **Zoom mode**: Focus on a single source with full-screen view
- **Smart search**: Search within a pane or across all sources
- **Log level filtering**: Filter by ERROR, WARN, INFO, DEBUG
- **Real-time streaming**: Live log updates with pause/resume
- **Container integration**: Direct Docker and Podman log support

## Key Bindings

### Navigation
- `1-9`: Jump to numbered pane
- `Tab/Shift+Tab`: Cycle through panes
- `h/j/k/l`: Vim-style pane navigation

### Layout & View
- `l`: Cycle layouts (horizontal → vertical → auto-grid)
- `z`: Zoom into focused pane
- `Z`: Zoom out to multi-pane view

### Search & Filter
- `/`: Search current pane
- `Ctrl+/` or `?`: Global search across all panes
- `e/w/i/a`: Filter by log level (Error/Warning/Info/All)

### Control
- `Space`: Pause/resume focused pane
- `f`: Toggle follow mode (auto-scroll)
- `c`: Clear focused pane
- `x`: Export logs
- `q`: Quit

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    logflow daemon                           │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │   TUI App   │◄───│Log Manager  │◄───│  IPC Server │     │
│  │             │    │             │    │             │     │
│  │ - Rendering │    │ - Buffering │    │ - Unix Socket│     │
│  │ - Input     │    │ - Filtering │    │ - Message    │     │
│  │ - Layout    │    │ - Routing   │    │   Protocol   │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
└─────────────────────────────────────────┬───────────────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
            ┌───────▼────────┐    ┌───────▼────────┐    ┌───────▼────────┐
            │ Source Process │    │ Source Process │    │ Source Process │
            │  --source      │    │  --source      │    │  --source      │
            │   backend      │    │   frontend     │    │    redis       │
            └────────────────┘    └────────────────┘    └────────────────┘
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details
