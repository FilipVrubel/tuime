# tuime

A terminal-based time tracking application with Pomodoro timer support.

## Features

- **Pomodoro Timer**: Customizable work sessions with automatic breaks
- **Manual Time Tracking**: Track time spent on activities
- **Activity Management**: Organize your work by activities
- **Statistics & Analytics**: Visualize your productivity with charts
- **TUI Interface**: Clean, keyboard-driven terminal UI built with Bubble Tea
- **Local SQLite Storage**: All data stored locally in your home directory

## Installation

### From Source

```bash
git clone https://github.com/yourusername/tuime.git
cd tuime
go build -o tuime .
./tuime
```

### Using Docker

See [README.docker.md](README.docker.md) for containerized deployment.


### Navigation

- **Tab**: Switch between views (Home, Statistics, Activities, Config)
- **Arrow Keys**: Navigate menus and change time periods in statistics
- **Enter**: Select/activate items
- **Esc**: Go back or exit

### Pomodoro Timer

1. Navigate to "Start Pomodoro"
2. Select an activity (or choose "No Activity")
3. Timer starts automatically
4. Take breaks as prompted (short break after each session, long break after 4 cycles)

### Manual Tracking

1. Navigate to "Start Tracker"
2. Select an activity
3. Press Space to start/pause, Enter to finish

## Configuration

Configuration is stored at `~/.config/tuime/config.json`.

Default settings:
- Work Duration: 25 minutes
- Short Break: 5 minutes
- Long Break: 15 minutes
- Cycles Before Long Break: 4

Modify settings through the Config view in the application.

## Data Storage

All data is stored locally:
- Database: `~/.local/share/tuime/tuime.db`
- Config: `~/.config/tuime/config.json`

## Development

### Prerequisites

- Go 1.24 or higher
- golangci-lint (optional, for linting)

### Build & Test

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Run all checks
make check

# Clean build artifacts
make clean
```

## License

MIT
