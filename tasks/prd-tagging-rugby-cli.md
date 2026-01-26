# PRD: tagging-rugby-cli

## Introduction

A Go CLI tool for rugby coaches and analysts to review game footage. The tool controls video playback via mpv media player and allows users to create timestamped annotations with categories, player tags, and team associations. All notes are stored in a local SQLite database, enabling filtering and searching during video review sessions.

## Goals

- Provide full control over mpv video playback via IPC socket
- Enable timestamped note-taking with categories and player/team tagging
- Support jumping to specific annotations for quick review
- Allow creation of video clips/segments between timestamps
- Store all data persistently in SQLite, tied to video filename
- Enable filtering and searching of notes by category, player, or time range

## User Stories

### US-001: Initialize a new video session
**Description:** As a coach, I want to open a video file for analysis so that I can start reviewing footage.

**Acceptance Criteria:**
- [ ] Command `tagging-rugby-cli open <video-file>` launches mpv with IPC socket enabled
- [ ] Creates/opens SQLite database in `~/.tagging-rugby-cli/` or configurable location
- [ ] Associates session with video filename for future resume
- [ ] Displays confirmation with video filename and duration
- [ ] Typecheck/lint passes

### US-002: Control video playback
**Description:** As an analyst, I want to control video playback without leaving the CLI so that I can efficiently review footage.

**Acceptance Criteria:**
- [ ] `play` - Resume playback
- [ ] `pause` - Pause playback
- [ ] `seek <time>` - Jump to specific timestamp (supports MM:SS or seconds)
- [ ] `ff [seconds]` - Fast forward (default 10s)
- [ ] `rw [seconds]` - Rewind (default 10s)
- [ ] `speed <multiplier>` - Set playback speed (0.25x to 2x)
- [ ] `frame+` / `frame-` - Step forward/backward one frame
- [ ] `now` - Display current timestamp
- [ ] All commands communicate with mpv via IPC socket
- [ ] Typecheck/lint passes

### US-003: Add a timestamped note
**Description:** As a coach, I want to add a note at the current video timestamp so that I can record observations during review.

**Acceptance Criteria:**
- [ ] Command `note add "<text>"` creates note at current mpv timestamp
- [ ] Optional flags: `--category <cat>` for event type (e.g., try, tackle, turnover)
- [ ] Optional flags: `--player <name>` to tag a player
- [ ] Optional flags: `--team <name>` to associate with a team
- [ ] Note stored in SQLite with video filename, timestamp, text, and metadata
- [ ] Displays confirmation with note ID and timestamp
- [ ] Typecheck/lint passes

### US-004: List notes for current session
**Description:** As an analyst, I want to see all notes for the current video so that I can review my annotations.

**Acceptance Criteria:**
- [ ] Command `note list` displays all notes for current video
- [ ] Output shows: ID, timestamp, category, player, team, text (truncated)
- [ ] Notes sorted by timestamp ascending
- [ ] Formatted as readable table in terminal
- [ ] Typecheck/lint passes

### US-005: Filter and search notes
**Description:** As a coach, I want to filter notes by category, player, or time range so that I can find specific events quickly.

**Acceptance Criteria:**
- [ ] `note list --category <cat>` filters by category
- [ ] `note list --player <name>` filters by player
- [ ] `note list --team <name>` filters by team
- [ ] `note list --from <time> --to <time>` filters by time range
- [ ] Filters can be combined
- [ ] Displays count of matching notes
- [ ] Typecheck/lint passes

### US-006: Jump to a note's timestamp
**Description:** As an analyst, I want to jump to a specific note's timestamp so that I can review that moment in the video.

**Acceptance Criteria:**
- [ ] Command `note goto <id>` seeks mpv to that note's timestamp
- [ ] Displays the note details after seeking
- [ ] Error message if note ID not found
- [ ] Typecheck/lint passes

### US-007: Edit an existing note
**Description:** As a coach, I want to edit a note so that I can correct or add details.

**Acceptance Criteria:**
- [ ] Command `note edit <id> "<new-text>"` updates note text
- [ ] Optional flags to update category, player, team
- [ ] `--timestamp` flag to update timestamp to current playback position
- [ ] Displays updated note after edit
- [ ] Typecheck/lint passes

### US-008: Delete a note
**Description:** As an analyst, I want to delete a note so that I can remove incorrect annotations.

**Acceptance Criteria:**
- [ ] Command `note delete <id>` removes note from database
- [ ] Prompts for confirmation before deletion
- [ ] `--force` flag to skip confirmation
- [ ] Displays confirmation after deletion
- [ ] Typecheck/lint passes

### US-009: Create a clip/segment
**Description:** As a coach, I want to mark a segment of video so that I can reference specific plays.

**Acceptance Criteria:**
- [ ] Command `clip start` marks the start timestamp
- [ ] Command `clip end "<description>"` marks end and saves the segment
- [ ] Alternative: `clip add <start> <end> "<description>"` for manual entry
- [ ] Clips stored in SQLite with start/end timestamps, description
- [ ] Optional `--category`, `--player`, `--team` flags
- [ ] Typecheck/lint passes

### US-010: List and play clips
**Description:** As an analyst, I want to list clips and loop playback of a specific clip.

**Acceptance Criteria:**
- [ ] Command `clip list` shows all clips for current video
- [ ] Output shows: ID, start, end, duration, category, description
- [ ] Command `clip play <id>` seeks to start and sets mpv A-B loop
- [ ] Command `clip stop` clears the A-B loop
- [ ] Typecheck/lint passes

### US-011: Manage categories
**Description:** As a coach, I want to define reusable categories so that annotations are consistent.

**Acceptance Criteria:**
- [ ] Command `category list` shows all defined categories
- [ ] Command `category add <name>` creates a new category
- [ ] Command `category delete <name>` removes a category
- [ ] Default categories pre-populated: try, tackle, turnover, lineout, scrum, penalty, kick
- [ ] Categories stored in SQLite
- [ ] Typecheck/lint passes

### US-012: Resume a previous session
**Description:** As an analyst, I want to resume analysis of a video I previously worked on.

**Acceptance Criteria:**
- [ ] Command `tagging-rugby-cli open <video-file>` detects existing notes for that file
- [ ] Displays count of existing notes/clips
- [ ] All previous notes and clips are available
- [ ] Typecheck/lint passes

### US-013: Interactive TUI mode
**Description:** As a coach, I want an interactive terminal UI so that I can efficiently control playback and take notes.

**Acceptance Criteria:**
- [ ] Command `tagging-rugby-cli open <video-file>` launches Bubbletea TUI
- [ ] TUI displays: status bar, command input, notes list panel
- [ ] Commands can be typed in command input area
- [ ] `:` enters command mode for text commands
- [ ] `q` or `Ctrl+C` gracefully exits TUI
- [ ] TUI built with Bubbletea, Bubbles components, and Lipgloss styling
- [ ] Responsive layout adapts to terminal size
- [ ] Typecheck/lint passes

### US-014: Vim-style keybindings in interactive mode
**Description:** As a coach, I want quick single-key shortcuts for stepping through video so that I can scrub through footage efficiently.

**Acceptance Criteria:**
- [ ] `H` key steps backward by current step size
- [ ] `L` key steps forward by current step size
- [ ] `J` key jumps to previous note/event (by timestamp)
- [ ] `K` key jumps to next note/event (by timestamp)
- [ ] `<` key decreases step size to next smaller value
- [ ] `>` key increases step size to next larger value
- [ ] `Space` key toggles play/pause
- [ ] `M` key toggles mute/unmute
- [ ] `O` key toggles note overlay on mpv screen
- [ ] `?` key displays help screen with all keybindings
- [ ] `S` key opens stats view
- [ ] Step sizes cycle through: 0.1s, 0.5s, 1s, 2s, 5s, 10s, 30s
- [ ] Current step size displayed in prompt or status line
- [ ] Default step size is 1s
- [ ] When jumping to an event, display the note details briefly
- [ ] Keybindings work without pressing Enter (raw input mode)
- [ ] Keybindings only active in interactive mode
- [ ] Can still type normal commands (keybindings don't interfere with text input)
- [ ] Typecheck/lint passes

### US-015: Help screen for keybindings
**Description:** As a user, I want to see a help screen listing all keyboard shortcuts so that I can learn and remember the controls.

**Acceptance Criteria:**
- [ ] `?` key displays help screen in terminal
- [ ] Help screen lists all keybindings with descriptions
- [ ] Help screen grouped by function (navigation, playback, notes, etc.)
- [ ] `help` command also displays the help screen
- [ ] Press any key to dismiss help screen and return to interactive mode
- [ ] Typecheck/lint passes

### US-016: Note overlay on video
**Description:** As a coach, I want to see notes displayed on the video so that I can review annotations without looking away from the footage.

**Acceptance Criteria:**
- [ ] When overlay is enabled, notes near current timestamp display on mpv screen
- [ ] Overlay shows note text, category, and player/team if present
- [ ] Notes displayed when playback is within 2 seconds of note timestamp (configurable)
- [ ] Overlay position configurable (top, bottom, or corner)
- [ ] Overlay does not obscure critical video content (semi-transparent background)
- [ ] `O` key toggles overlay on/off in interactive mode
- [ ] `overlay on` / `overlay off` commands available
- [ ] Overlay state persists for session
- [ ] Typecheck/lint passes

### US-017: Status bar display
**Description:** As a user, I want to see current state information so that I know the current settings at a glance.

**Acceptance Criteria:**
- [ ] TUI displays persistent status bar (top or bottom of screen)
- [ ] Status bar shows: current timestamp, video duration, step size
- [ ] Status bar shows: play/pause icon, mute icon, overlay icon
- [ ] Status bar styled with Lipgloss (distinct background color)
- [ ] Status bar updates in real-time during playback
- [ ] Typecheck/lint passes

### US-018: Track tackles
**Description:** As a coach, I want to record detailed tackle events so that I can analyze defensive performance.

**Acceptance Criteria:**
- [ ] Command `tackle add` creates a tackle event at current timestamp
- [ ] Required field: `--player <name>` - player making the tackle
- [ ] Required field: `--team <name>` - team name
- [ ] Required field: `--attempt <number>` - attempt number
- [ ] Required field: `--outcome <type>` - one of: missed, completed, possible, other
- [ ] Optional field: `--followed <event>` - event that followed the tackle (e.g., turnover, penalty, ruck)
- [ ] Optional field: `--star` - boolean flag to mark as notable
- [ ] Optional field: `--notes "<text>"` - additional notes
- [ ] Optional field: `--zone <zone>` - field zone where tackle occurred
- [ ] Command `tackle list` shows all tackles for current video
- [ ] Command `tackle list --player <name>` filters by player
- [ ] Command `tackle list --outcome <type>` filters by outcome
- [ ] Command `tackle list --star` shows only starred tackles
- [ ] J/K keys also navigate through tackle events
- [ ] Typecheck/lint passes

### US-019: Export player tackle statistics
**Description:** As a coach, I want to export a player's tackle statistics to a text file so that I can share or review their performance outside the CLI.

**Acceptance Criteria:**
- [ ] Command `tackle export --player <name>` exports stats to text file
- [ ] Default output file: `<player-name>-tackles.txt` in current directory
- [ ] Optional `--output <path>` to specify custom file path
- [ ] Export includes: total tackles, completed, missed, possible, other counts
- [ ] Export includes: completion percentage
- [ ] Export includes: list of all tackle events with timestamps, zone, followed, notes
- [ ] Export includes: starred tackles highlighted
- [ ] Optional `--video <path>` to filter to specific video
- [ ] Displays confirmation with file path after export
- [ ] Typecheck/lint passes

### US-020: Export clips as video files
**Description:** As a coach, I want to export clips as separate video files so that I can share specific plays with players or staff.

**Acceptance Criteria:**
- [ ] Command `clip export <id>` exports clip as video file
- [ ] Uses ffmpeg for video extraction (must be installed)
- [ ] Default output: `clip-<id>.mp4` in current directory
- [ ] Optional `--output <path>` to specify custom file path
- [ ] Optional `--format <fmt>` to specify format (mp4, webm, mkv)
- [ ] Preserves original video quality by default (stream copy)
- [ ] Optional `--reencode` flag to re-encode video
- [ ] Command `clip export --all` exports all clips for current video
- [ ] Shows progress during export
- [ ] Error message if ffmpeg not installed
- [ ] Typecheck/lint passes

### US-021: Stats page with player filtering
**Description:** As a coach, I want to view tackle statistics with the ability to filter players in and out so that I can compare performance.

**Acceptance Criteria:**
- [ ] `S` key or `stats` command opens stats view in TUI
- [ ] Stats view shows table of all players with tackle statistics
- [ ] Columns: Player, Total, Completed, Missed, Possible, Completion %, Starred
- [ ] Players listed alphabetically by default
- [ ] Type player name/initials to toggle them in filter (highlighted)
- [ ] `/` enters filter mode for typing player indicators
- [ ] Filtered players shown at top, others greyed out below
- [ ] Multiple players can be filtered in for comparison
- [ ] `Esc` clears all filters
- [ ] `Tab` cycles sort column (Total, Completed, Missed, %)
- [ ] Stats update based on current video or all videos (toggle with `V`)
- [ ] `Backspace` exits stats view and returns to main view
- [ ] Typecheck/lint passes

### US-022: Dependency checks and warnings
**Description:** As a user, I want to be warned if required dependencies are missing so that I know what to install.

**Acceptance Criteria:**
- [ ] On startup, check if mpv is installed (via `which mpv` or similar)
- [ ] On startup, check if ffmpeg is installed (via `which ffmpeg` or similar)
- [ ] If mpv is missing, display warning with install link: https://mpv.io/installation/
- [ ] If ffmpeg is missing, display warning with install link: https://ffmpeg.org/download.html
- [ ] mpv is required - exit with error if not found when opening a video
- [ ] ffmpeg is optional - only warn when attempting clip export
- [ ] Command `tagging-rugby-cli doctor` checks all dependencies and reports status
- [ ] Warnings styled with pink color from palette for visibility
- [ ] Typecheck/lint passes

## Functional Requirements

- FR-1: The CLI must communicate with mpv via JSON IPC protocol over Unix socket
- FR-2: The CLI must launch mpv with `--input-ipc-server` flag pointing to a known socket path
- FR-3: All notes must be stored in SQLite with schema: id, video_path, timestamp_seconds, text, category, player, team, created_at
- FR-4: All clips must be stored in SQLite with schema: id, video_path, start_seconds, end_seconds, description, category, player, team, created_at
- FR-5: Categories must be stored in SQLite with schema: id, name, created_at
- FR-6: All tackles must be stored in SQLite with schema: id, video_path, timestamp_seconds, player, team (string), attempt (integer), outcome (missed|completed|possible|other), followed (string - event that followed), star (boolean), notes, zone, created_at
- FR-7: The database must be created automatically on first run
- FR-8: Video files must be identified by absolute path for consistency
- FR-9: Timestamps must support both MM:SS format and raw seconds for input
- FR-10: Timestamps must display in MM:SS format in output
- FR-11: The CLI must handle mpv not running or connection failures gracefully
- FR-12: The CLI must use cobra for command structure
- FR-13: The CLI must use a config file for database location and default settings

## Design Considerations

### Color Palette
Retro purple 8-color palette for terminal UI:

| Name | Hex | Usage |
|------|-----|-------|
| Deep Purple | `#1a1a2e` | Background |
| Dark Purple | `#16213e` | Panel backgrounds |
| Purple | `#4a347d` | Borders, separators |
| Bright Purple | `#7b2cbf` | Highlights, active elements |
| Lavender | `#c77dff` | Primary text |
| Light Lavender | `#e0aaff` | Secondary text, labels |
| Pink | `#ff6b9d` | Warnings, starred items |
| Cyan | `#64dfdf` | Success, timestamps, counts |

### UI Layout
```
┌──────────────────────────────────────────────────┐
│ ▶ 12:34 / 45:00   Step: 1s   [M] [O]   tagging   │  <- Status bar
├──────────────────────────────────────────────────┤
│                                                  │
│  Notes/Tackles List                              │  <- Main panel
│  - 02:15 [tackle] Player A - completed           │
│  - 05:32 [note] Good defensive line              │
│  > 12:30 [tackle] Player B - missed      ★      │  <- Selected
│                                                  │
├──────────────────────────────────────────────────┤
│ : command input                                  │  <- Command input
└──────────────────────────────────────────────────┘
```

## Non-Goals

- No video transcoding beyond clip export (no format conversion of source video)
- No multi-video playlists or batch processing
- No remote/cloud storage of notes
- No GUI or web interface
- No real-time collaboration features
- No CSV/JSON export for notes/clips (text export for tackle stats only)
- No match/game metadata management beyond video filename

## Technical Considerations

- **Language:** Go 1.21+
- **CLI Framework:** cobra for command parsing
- **TUI Framework:** Bubbletea for interactive terminal UI (with Bubbles components and Lipgloss for styling)
- **Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **mpv IPC:** JSON-based protocol over Unix socket (`/tmp/tagging-rugby-mpv.sock`)
- **Config:** Viper for configuration management, stored in `~/.config/tagging-rugby-cli/config.yaml`
- **Database Location:** Default `~/.local/share/tagging-rugby-cli/data.db`
- **Clip Export:** ffmpeg (external dependency, must be installed by user)

### mpv IPC Commands Used
- `get_property`: time-pos, duration, pause, speed, mute
- `set_property`: pause, speed, mute, ab-loop-a, ab-loop-b
- `seek`: absolute and relative seeking
- `frame-step`, `frame-back-step`: frame-by-frame navigation
- `osd-overlay`: display note text overlays on video

## Success Metrics

- User can add a note in under 3 seconds from observation
- User can jump to any annotated moment in under 2 commands
- All video control accessible without touching mpv window
- Notes persist reliably across sessions
- CLI responds to commands within 100ms

## Open Questions

- Should timestamps include frame numbers for frame-accurate annotation?
- Should additional vim-style keybindings be added?
- Should clips support nested/hierarchical categorization?
- Should the tool support Windows (named pipes instead of Unix sockets)?
