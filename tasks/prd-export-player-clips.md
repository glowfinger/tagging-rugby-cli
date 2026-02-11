# PRD: Export Player Tackle Clips with ffmpeg

## Introduction

Add the ability to export individual tackle clips organized by player from within the TUI. This feature extracts video clips for each tackle using ffmpeg, creating separate MP4 files organized in player-named folders. Clips use the timing data from the database (clip start/end if available, or tackle timestamp with padding), with a minimum of 4 seconds before the tackle start to provide context.

## Goals

- Enable coaches and team members to quickly share specific plays for review
- Organize exported clips by player for easy distribution and analysis
- Use existing timing data from the database to ensure accurate clip boundaries
- Provide a seamless export workflow from within the TUI
- Generate compatible MP4 files that work across all devices and platforms

## User Stories

### US-001: Add export clips keybinding to TUI
**Description:** As a user, I want to trigger clip export from the TUI with a keybinding so I can quickly generate clips without leaving my workflow.

**Acceptance Criteria:**
- [ ] Add keybinding (e.g., `E` for Export) to main TUI view
- [ ] Keybinding triggers export flow for current video
- [ ] Add keybinding to help overlay (`?` screen)
- [ ] Update MEMORY.md with export keybinding pattern
- [ ] CGO_ENABLED=0 go vet ./... passes

### US-002: Query tackles with player and timing data
**Description:** As a developer, I need to query all tackles with player information and timing data so I can generate clips for each player.

**Acceptance Criteria:**
- [ ] Create SQL query to fetch tackles with player names, timestamps, clip start/end times
- [ ] Query joins note_tackles with note_timing and note_details tables
- [ ] Query filters out tackles without player data
- [ ] Query orders by player name, then timestamp
- [ ] Add query to db/sql/ directory with embedded SQL
- [ ] CGO_ENABLED=0 go vet ./... passes

### US-003: Calculate clip timestamps with 4-second minimum
**Description:** As a developer, I need to calculate accurate clip start/end times using database timing data with a 4-second minimum before the tackle.

**Acceptance Criteria:**
- [ ] If clip_start and clip_end exist in database, use those values
- [ ] If only tackle timestamp exists, calculate: start = timestamp - 4 seconds (minimum), end = timestamp + 10 seconds (or duration from database)
- [ ] Ensure start time never goes below 0
- [ ] Ensure end time never exceeds video duration
- [ ] Function signature: `CalculateClipBounds(timestamp, clipStart, clipEnd, videoDuration float64) (start, end float64)`
- [ ] CGO_ENABLED=0 go vet ./... passes

### US-004: Execute ffmpeg to extract clips as H.264 MP4
**Description:** As a developer, I need to execute ffmpeg commands to extract video clips in a compatible format.

**Acceptance Criteria:**
- [ ] Use ffmpeg command: `ffmpeg -ss {start} -i {input} -t {duration} -c:v libx264 -c:a aac -preset fast {output}`
- [ ] Start time (-ss) positioned before input (-i) for faster seeking
- [ ] Duration (-t) calculated as `end - start`
- [ ] Output codec: H.264 video, AAC audio
- [ ] Preset: fast (balance between speed and file size)
- [ ] Handle ffmpeg errors and return meaningful error messages
- [ ] CGO_ENABLED=0 go vet ./... passes

### US-005: Organize clips in player-named folders
**Description:** As a user, I want clips organized in folders by player name so I can easily find and share specific player highlights.

**Acceptance Criteria:**
- [ ] Create output directory structure: `{video-basename}-clips/{player-name}/`
- [ ] Sanitize player names for filesystem (replace spaces with underscores, remove special chars)
- [ ] Clip filename format: `{player-name}_{timestamp}_{category}.mp4`
- [ ] Example: `rugby-game-clips/Toby/Toby_0-26-42_tackle.mp4`
- [ ] Create directories if they don't exist (os.MkdirAll)
- [ ] Handle players with no name (use "Unknown" folder)
- [ ] CGO_ENABLED=0 go vet ./... passes

### US-006: Show export progress in TUI
**Description:** As a user, I want to see export progress so I know the operation is working and how long it will take.

**Acceptance Criteria:**
- [ ] Display status message when export starts: "Exporting {count} clips for {n} players..."
- [ ] Show progress indicator (e.g., "Processing clip {current}/{total}...")
- [ ] Display success message with output directory path when complete
- [ ] Show error message if export fails (with reason)
- [ ] Status messages appear in command result area (existing 3-second display mechanism)
- [ ] Allow user to continue using TUI while export runs in background
- [ ] CGO_ENABLED=0 go vet ./... passes

### US-007: Check dependencies and handle errors
**Description:** As a user, I want clear error messages if ffmpeg is not installed or if export fails so I know how to fix the issue.

**Acceptance Criteria:**
- [ ] Check if ffmpeg is installed before starting export (exec.LookPath("ffmpeg"))
- [ ] If ffmpeg not found, show error: "ffmpeg not found. Install with: brew install ffmpeg (macOS) or apt install ffmpeg (Linux)"
- [ ] If video file is missing, show error with file path
- [ ] If no tackles found for video, show: "No tackles with player data found for this video"
- [ ] If disk write fails, show error with directory path
- [ ] If ffmpeg command fails, show error with stderr output
- [ ] CGO_ENABLED=0 go vet ./... passes

## Functional Requirements

- FR-1: Add export keybinding to TUI that triggers clip export for the current video
- FR-2: Query database for all tackles with player names and timing data
- FR-3: Calculate clip boundaries using database timing data with 4-second minimum before tackle
- FR-4: Execute ffmpeg to extract clips as H.264 MP4 files
- FR-5: Organize clips in player-named folders: `{video-basename}-clips/{player}/`
- FR-6: Show export progress and status messages in TUI
- FR-7: Check for ffmpeg installation and provide clear error messages
- FR-8: Handle edge cases: no player data, video file missing, disk write failures

## Non-Goals

- No video player integration (clips are just exported files)
- No clip preview before export
- No batch export for multiple videos (only current video)
- No custom video encoding settings (uses standard H.264/AAC)
- No clip editing or trimming within the TUI
- No automatic upload or sharing to external services
- No export of notes (only tackles)

## Technical Considerations

- **ffmpeg dependency**: Users must have ffmpeg installed. Add to `doctor` command checks.
- **Database schema**: Use existing tables (note_tackles, note_timing, note_details) - no schema changes needed
- **Timing data**: Clip start/end stored in note_timing table. If null, calculate from tackle timestamp.
- **File paths**: Video file path stored in m.videoPath. Output directory created relative to video file.
- **Player names**: Sanitize for filesystem compatibility (replace spaces, remove `/\:*?"<>|`)
- **Background execution**: Export can run in background goroutine, updating status via tea.Cmd messages
- **Existing patterns**: Follow mpv package pattern for external command execution (exec.Command)

## Design Considerations

### Directory Structure
```
rugby-game.mp4
rugby-game-clips/
  ├── Toby/
  │   ├── Toby_0-26-42_tackle.mp4
  │   ├── Toby_0-44-07_tackle.mp4
  │   └── Toby_0-50-03_tackle.mp4
  ├── Owen/
  │   ├── Owen_0-27-02_tackle.mp4
  │   └── Owen_0-51-33_tackle.mp4
  └── Sam_G/
      └── Sam_G_0-46-44_tackle.mp4
```

### Filename Format
- `{player}_{H-MM-SS}_{category}.mp4`
- Example: `Toby_0-26-42_tackle.mp4`
- Time format matches existing timeutil.FormatTime() output (with `:` replaced by `-`)

### ffmpeg Command
```bash
ffmpeg -ss 26.0 -i rugby-game.mp4 -t 15.0 -c:v libx264 -c:a aac -preset fast Toby_0-26-42_tackle.mp4
```

## Success Metrics

- Users can export clips with a single keypress from the TUI
- Exported clips are organized clearly by player
- Clips include sufficient context (4+ seconds before tackle)
- MP4 files play correctly on all major platforms
- Export completes without errors for typical game footage

## Open Questions

None - requirements are clear and scope is well-defined.
