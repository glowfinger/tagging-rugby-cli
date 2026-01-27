# tagging-rugby-cli

A terminal-based video analysis tool for rugby coaches and analysts. Control video playback via mpv and create timestamped annotations with categories, player tags, and team associations.

## Features

- Full control over mpv video playback via IPC socket
- Timestamped notes with categories, player/team tagging
- Detailed tackle tracking with outcomes and statistics
- Video clip segments with A-B loop playback
- Export clips as video files using ffmpeg
- Interactive TUI with vim-style keybindings
- Note overlay on video during playback
- SQLite database for persistent storage

## Prerequisites

| Dependency | Required | Description |
|------------|----------|-------------|
| Go 1.24+   | Yes      | Build from source |
| mpv        | Yes      | Video playback |
| ffmpeg     | No       | Clip export only |

### Installing Dependencies

**macOS (Homebrew)**
```bash
brew install go mpv ffmpeg
```

**Ubuntu/Debian**
```bash
sudo apt update
sudo apt install golang mpv ffmpeg
```

**Arch Linux**
```bash
sudo pacman -S go mpv ffmpeg
```

**Windows**
1. Install Go from https://go.dev/dl/
2. Install mpv from https://mpv.io/installation/
3. Install ffmpeg from https://ffmpeg.org/download.html
4. Add mpv and ffmpeg to your PATH

## Installation

### Build from Source

```bash
git clone https://github.com/user/tagging-rugby-cli.git
cd tagging-rugby-cli
go build -o tagging-rugby-cli .
```

Move the binary to your PATH:
```bash
sudo mv tagging-rugby-cli /usr/local/bin/
```

### Verify Dependencies

Run the doctor command to check all dependencies are installed:

```bash
tagging-rugby-cli doctor
```

Example output:
```
Checking dependencies...

✓ mpv: OK
✓ ffmpeg: OK

All dependencies are installed!
```

## Quick Start

### TUI Mode (Recommended)

Launch the interactive terminal UI:

```bash
tagging-rugby-cli open -t match.mp4
```

The TUI provides:
- Status bar showing current timestamp, step size, and playback state
- Notes/tackles list panel
- Command input area (press `:` to enter commands)

### CLI Mode

Open a video without TUI (video controls via separate mpv window):

```bash
tagging-rugby-cli open match.mp4
```

Then use CLI commands in another terminal while mpv is running.

## TUI Keybindings

### Playback

| Key | Action |
|-----|--------|
| `Space` | Toggle play/pause |
| `M` | Toggle mute |
| `H` | Step backward (by step size) |
| `L` | Step forward (by step size) |
| `<` | Decrease step size |
| `>` | Increase step size |

Step sizes cycle through: 0.1s, 0.5s, 1s, 2s, 5s, 10s, 30s

### Navigation

| Key | Action |
|-----|--------|
| `J` | Select previous item in list |
| `K` | Select next item in list |
| `Enter` | Jump to selected item's timestamp |

### Views

| Key | Action |
|-----|--------|
| `?` | Show/hide help screen |
| `S` | Open stats view |
| `O` | Toggle note overlay on video |
| `Backspace` | Return to main view |

### Stats View

| Key | Action |
|-----|--------|
| `/` | Enter filter mode (type player name/initials) |
| `Esc` | Clear all filters |
| `Tab` | Cycle sort column |
| `V` | Toggle current video / all videos |
| `J/K` | Navigate player list |

### Commands

| Key | Action |
|-----|--------|
| `:` | Enter command mode |
| `Esc` | Cancel command mode |
| `q` | Quit application |

## CLI Commands

### Notes

Add a timestamped note at the current playback position:

```bash
tagging-rugby-cli note add "Good defensive line"
tagging-rugby-cli note add "Try scored" --category try --player "John Smith" --team "Home"
```

List notes for the current video:

```bash
tagging-rugby-cli note list
tagging-rugby-cli note list --category tackle --player "John Smith"
tagging-rugby-cli note list --from 5:00 --to 10:00
```

Jump to a note's timestamp:

```bash
tagging-rugby-cli note goto 5
```

Edit a note:

```bash
tagging-rugby-cli note edit 5 "Updated text" --category penalty
tagging-rugby-cli note edit 5 "Same text" --timestamp  # Update to current position
```

Delete a note:

```bash
tagging-rugby-cli note delete 5
tagging-rugby-cli note delete 5 --force  # Skip confirmation
```

### Tackles

Record a tackle event:

```bash
tagging-rugby-cli tackle add --player "John Smith" --team "Home" --attempt 1 --outcome completed
tagging-rugby-cli tackle add -p "Jane Doe" -t "Away" -a 2 -o missed --star --notes "Lost footing"
```

Outcome options: `completed`, `missed`, `possible`, `other`

List tackles:

```bash
tagging-rugby-cli tackle list
tagging-rugby-cli tackle list --player "John Smith"
tagging-rugby-cli tackle list --outcome missed --star
```

Export player statistics:

```bash
tagging-rugby-cli tackle export --player "John Smith"
tagging-rugby-cli tackle export -p "John Smith" --output stats.txt
```

### Clips

Mark a clip using start/end workflow:

```bash
tagging-rugby-cli clip start
# ... seek to end position ...
tagging-rugby-cli clip end "Great try"
```

Or add a clip with explicit times:

```bash
tagging-rugby-cli clip add 1:30 2:15 "Lineout play" --category lineout
```

List clips:

```bash
tagging-rugby-cli clip list
```

Play a clip with A-B loop:

```bash
tagging-rugby-cli clip play 3
tagging-rugby-cli clip stop  # Clear loop
```

Export clips as video files:

```bash
tagging-rugby-cli clip export 3
tagging-rugby-cli clip export 3 --output highlight.mp4 --format mp4
tagging-rugby-cli clip export --all --format webm --reencode
```

### Categories

List available categories:

```bash
tagging-rugby-cli category list
```

Add a custom category:

```bash
tagging-rugby-cli category add "dropout"
```

Delete a category:

```bash
tagging-rugby-cli category delete "dropout"
```

Default categories: try, tackle, turnover, lineout, scrum, penalty, kick

## TUI Commands

When in command mode (press `:`), these commands are available:

| Command | Description |
|---------|-------------|
| `note add <text>` | Add note at current timestamp |
| `note list` | Reload notes list |
| `note goto <id>` | Jump to note timestamp |
| `tackle add -p <player> -t <team> -a <num> -o <outcome>` | Add tackle |
| `tackle list` | Reload tackles list |
| `clip start` | Mark clip start |
| `clip end <description>` | Mark clip end and save |
| `clip list` | Show clip count |
| `clip play <id>` | Play clip with A-B loop |
| `clip stop` | Clear A-B loop |
| `pause` / `play` | Control playback |
| `mute` | Toggle mute |
| `seek <time>` | Seek to time (MM:SS or seconds) |
| `speed <multiplier>` | Set playback speed |
| `help` | Show available commands |
| `quit` | Exit application |

## Data Storage

| Data | Location |
|------|----------|
| Database | `~/.local/share/tagging-rugby-cli/data.db` |
| mpv Socket | `/tmp/tagging-rugby-mpv.sock` |

## Technology Stack

- **Language:** Go 1.24
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **TUI Framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea) with [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Database:** SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)
- **Video Player:** [mpv](https://mpv.io/) via JSON IPC protocol
- **Video Export:** [ffmpeg](https://ffmpeg.org/)

## License

MIT
