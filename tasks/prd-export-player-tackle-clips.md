# PRD: Export Player Tackle Clips

## Introduction

Add a TUI-driven bulk export feature that renders all tackle clips as individual MP4 files using ffmpeg. When the user presses `Ctrl+E`, the system queries all tackle notes across all videos, calculates timing (with a 4-second minimum duration), spawns ffmpeg processes with stream copy (`-c copy`), and tracks progress in the `note_clips` table. A progress bar in Column 1 shows real-time export status.

Clips are organized in a nested folder structure:
```
clips/{video_filename}/{category}/{player}/{start_hhmmss}-{player}-{category}-{outcome}.mp4
```

## Goals

- Export all tackle clips across all videos with a single keypress (`Ctrl+E`)
- Enforce a minimum 4-second clip duration (extend `end` from `start` if needed)
- Track each clip's export state in `note_clips` (started, finished, errored)
- Show a progress bar component in Column 1 during export
- Organize output files in a deterministic folder hierarchy relative to the video
- Reuse existing `buildFfmpegArgs()` and ffmpeg infrastructure from `cmd/clip.go`

## User Stories

### US-001: Query tackle notes for export
**Description:** As the system, I need to gather all tackle notes with their timing, video path, tackle details, and category so I can build the export queue.

**Acceptance Criteria:**
- [ ] New DB query joins `notes`, `note_timing`, `note_videos`, `note_tackles` for `category = 'tackle'`
- [ ] Query returns: `note.id`, `note.category`, `note_timing.start`, `note_timing.end`, `note_videos.path`, `note_tackles.player`, `note_tackles.outcome`
- [ ] Query works across all videos (not filtered to current video)
- [ ] Results are ordered by `note_videos.path`, then `note_timing.start`
- [ ] Typecheck/lint passes (`CGO_ENABLED=0 go vet ./...`)

### US-002: Enforce minimum 4-second clip duration
**Description:** As the system, I need to ensure every exported clip is at least 4 seconds long so that very short tackles still produce usable clips.

**Acceptance Criteria:**
- [ ] If `note_timing.end - note_timing.start < 4.0`, set effective end to `start + 4.0`
- [ ] The adjusted duration is used for ffmpeg export only — do NOT modify the database timing
- [ ] If `note_timing.end` is 0 or NULL, use `start + 4.0` as the effective end
- [ ] Typecheck/lint passes

### US-003: Build clip output path
**Description:** As the system, I need to construct the correct output file path following the naming convention so clips are organized by video, category, and player.

**Acceptance Criteria:**
- [ ] Base directory: `clips/` relative to the video file's directory
- [ ] Folder structure: `clips/{video_filename_without_ext}/{category}/{player}/`
- [ ] File name format: `{start_as_hhmmss}-{player}-{category}-{outcome}.mp4`
- [ ] `start_as_hhmmss` is derived from `note_timing.start` seconds (e.g., 3661.5 → `010101`)
- [ ] Sanitize player name and outcome for filesystem safety (replace spaces/special chars with `_`)
- [ ] Create directories recursively if they don't exist
- [ ] Typecheck/lint passes

### US-004: Track clip export in note_clips table
**Description:** As a user, I want each exported clip tracked in the database so I can see what was exported, when, and whether it succeeded.

**Acceptance Criteria:**
- [ ] Before ffmpeg starts: insert `note_clips` row with `name` = output file path, `started_at` = now, `finished_at` = NULL
- [ ] On success: update row with `finished_at` = now, `duration` = actual clip duration
- [ ] On error: update row with `error_at` = now, `error` = error message
- [ ] Skip notes that already have a `note_clips` row with a non-NULL `finished_at` (already exported)
- [ ] Typecheck/lint passes

### US-005: Run ffmpeg export with stream copy
**Description:** As the system, I need to execute ffmpeg for each clip using stream copy mode for fast export.

**Acceptance Criteria:**
- [ ] Check ffmpeg is available via `deps.CheckFfmpeg()` before starting
- [ ] Reuse or adapt `buildFfmpegArgs()` from `cmd/clip.go` for argument construction
- [ ] Use `-c copy` (stream copy) — no re-encoding
- [ ] Run exports sequentially (one ffmpeg at a time) to avoid I/O contention
- [ ] Exports run in a background goroutine so the TUI remains responsive
- [ ] Send Bubble Tea messages (cmds) back to the TUI to update progress state
- [ ] Typecheck/lint passes

### US-006: Progress bar component in Column 1
**Description:** As a user, I want to see a visual progress bar in Column 1 showing how many clips have been exported so I can monitor the batch export.

**Acceptance Criteria:**
- [ ] New `ExportProgress` component in `tui/components/`
- [ ] Displays: progress bar, percentage, `N/M clips` counter, current clip file name
- [ ] Rendered as a bordered info box in Column 1 (below selected tag detail)
- [ ] Only visible when an export is in progress
- [ ] Shows completion message briefly when all clips finish
- [ ] Shows error count if any clips failed (e.g., `12/15 done, 3 errors`)
- [ ] Wraps in `Container` for exact dimensions (per architecture rules)
- [ ] Typecheck/lint passes

### US-007: Ctrl+E keybinding to trigger export
**Description:** As a user, I want to press `Ctrl+E` to start exporting all tackle clips so I don't have to use the CLI.

**Acceptance Criteria:**
- [ ] `Ctrl+E` triggers the export when no form/overlay is active
- [ ] If an export is already in progress, show a message "Export already in progress" instead of starting a new one
- [ ] Guard: do not trigger if a form, help overlay, stats view, or command input is active
- [ ] Add `Ctrl+E  Export clips` to the Controls keybinding display (Column 4)
- [ ] Typecheck/lint passes

### US-008: Render exported clips in TUI
**Description:** As a user, I want a keybinding to view/list the exported clips so I can verify what was exported.

**Acceptance Criteria:**
- [ ] New keybinding (suggest `C` for "Clips") toggles a clips list view
- [ ] View shows exported clips grouped by video with: file name, player, category, outcome, duration, status
- [ ] Clips with errors shown in red with error message
- [ ] Press the key again or `Esc` to dismiss
- [ ] Typecheck/lint passes

## Functional Requirements

- FR-1: New SQL query to select all tackle notes joined with `note_timing`, `note_videos`, and `note_tackles` across all videos
- FR-2: Minimum clip duration enforced at 4 seconds from `start` — if `end - start < 4.0`, use `start + 4.0`
- FR-3: Output path: `clips/{video_filename_no_ext}/{category}/{player}/{hhmmss}-{player}-{category}-{outcome}.mp4`
- FR-4: Insert `note_clips` row before each export with `name` = file path; update `finished_at` on success or `error_at` + `error` on failure
- FR-5: Skip clips that already have a completed `note_clips` entry (idempotent re-runs)
- FR-6: Use `buildFfmpegArgs()` with `-c copy` (stream copy) for fast export
- FR-7: Run exports sequentially in a background goroutine, sending progress messages to the TUI via `tea.Cmd`
- FR-8: `ExportProgress` component in Column 1 showing progress bar, percentage, clip counter, current file name
- FR-9: `Ctrl+E` triggers export; guarded against duplicate runs and active overlays
- FR-10: New keybinding to view a list of exported clips with status

## Non-Goals

- No re-encoding option in TUI (CLI `clip export --reencode` covers this)
- No per-player or per-video filtering at export time (exports all tackles across all videos)
- No parallel ffmpeg processes (sequential to avoid I/O contention)
- No thumbnail generation
- No clip playback from the export view (use existing `clip play` CLI command)

## Design Considerations

- The progress bar should follow the existing bordered info box pattern (like `RenderInfoBox` in Column 1)
- Use the Ciapre palette: `Green` for completed, `Amber` for in-progress, `Red` for errors
- Progress bar character: `█` for filled, `░` for empty (common terminal pattern)
- The clips list view could reuse the `StatsView` full-screen overlay pattern

## Technical Considerations

- **Existing infrastructure:** `cmd/clip.go` has `buildFfmpegArgs()` — extract to a shared package (e.g., `pkg/ffmpeg/`) or duplicate with minor changes
- **Goroutine communication:** Use a channel or Bubble Tea's `tea.Program.Send()` to push progress updates from the export goroutine to the TUI model
- **Database access:** The export goroutine needs its own `*sql.DB` connection (don't share the TUI's connection across goroutines)
- **File path sanitization:** Replace characters not safe for filenames (`/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`, spaces) with `_`
- **hhmmss formatting:** Convert `note_timing.start` (float64 seconds) to `HHMMSS` string — e.g., `3661.5` → `010101`

## Success Metrics

- All tackle notes produce correctly named clip files in the expected directory structure
- Re-running export skips already-exported clips (idempotent)
- Progress bar accurately reflects export state
- No TUI freezing during export (responsive UI)
- Export completes without errors for valid video files

## Open Questions

- Should the clips folder be relative to the video file location or the current working directory? (PRD assumes relative to the video file's directory)
- Should there be a confirmation dialog before starting a bulk export?
- What happens if the source video file has been moved/deleted since tagging? (Suggest: record error in `note_clips.error` and continue)
