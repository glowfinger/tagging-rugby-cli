# PRD: Change Time Format to HH:MM:SS

## Introduction

All time displays in the application currently use `MM:SS` format (e.g., `1:22` for 1 minute 22 seconds). This works for short videos but becomes confusing for longer content — `71:22` is harder to read than `1:11:22`. This change updates all time formatting to `HH:MM:SS` (e.g., `1:11:22`) and updates the input parsers to accept `HH:MM:SS` alongside the existing `MM:SS` and raw seconds formats.

## Goals

- Display all timestamps in `H:MM:SS` format across the entire application (TUI and CLI)
- Accept `HH:MM:SS`, `MM:SS`, and raw seconds as input in all time-parsing contexts
- Consolidate duplicate formatting/parsing logic where possible

## User Stories

### US-001: Update display format functions to HH:MM:SS
**Description:** As a user, I want all timestamps displayed as `H:MM:SS` so I can quickly read times for long videos.

**Acceptance Criteria:**
- [ ] `formatTimeString()` in `tui/tui.go` outputs `H:MM:SS` (e.g., `0:01:22`, `1:11:22`)
- [ ] `formatTime()` in `tui/components/statusbar.go` outputs `H:MM:SS`
- [ ] Both functions produce identical output for the same input (they are currently duplicates)
- [ ] Zero seconds displays as `0:00:00`
- [ ] 90 seconds displays as `0:01:30`
- [ ] 4282 seconds displays as `1:11:22`
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-002: Update inline time formatting in CLI commands
**Description:** As a developer, I need to replace all inline `MM:SS` formatting in CLI commands with the updated format.

**Acceptance Criteria:**
- [ ] `cmd/note.go` line ~89: `fmt.Sprintf("%d:%02d", ...)` replaced with `H:MM:SS` format
- [ ] `cmd/note.go` line ~151: note list timestamp display uses `H:MM:SS`
- [ ] `cmd/clip.go` line ~68: clip timestamp display uses `H:MM:SS`
- [ ] `cmd/tackle.go` line ~97: tackle add output uses `H:MM:SS`
- [ ] `cmd/tackle.go` line ~183: tackle list timestamp display uses `H:MM:SS`
- [ ] Consider extracting a shared `FormatTime()` function to avoid duplication between CLI and TUI
- [ ] Typecheck/vet passes

### US-003: Update form headers to HH:MM:SS
**Description:** As a user, I want the tackle and note form headers to show the timestamp in `H:MM:SS` format.

**Acceptance Criteria:**
- [ ] `tui/forms/tackleform.go` `NewTackleForm()`: header shows `"Add Tackle @ H:MM:SS"` (e.g., `"Add Tackle @ 0:05:30"`)
- [ ] `tui/forms/noteform.go` `NewNoteForm()`: header shows `"Add Note @ H:MM:SS"` (e.g., `"Add Note @ 0:05:30"`)
- [ ] Typecheck/vet passes

### US-004: Update input parsers to accept HH:MM:SS
**Description:** As a user, I want to type times as `HH:MM:SS`, `MM:SS`, or raw seconds in seek commands and filter flags so any format works.

**Acceptance Criteria:**
- [ ] `parseTimeToSeconds()` in `tui/tui.go` accepts `HH:MM:SS` (e.g., `"1:11:22"` → `4282.0`)
- [ ] `parseTimeToSeconds()` in `cmd/note.go` accepts `HH:MM:SS`
- [ ] Parsing priority: try `HH:MM:SS` first, then `MM:SS`, then raw seconds
- [ ] `"1:11:22"` parses as 1 hour 11 minutes 22 seconds = `4282.0`
- [ ] `"5:30"` still parses as 5 minutes 30 seconds = `330.0` (backward compatible)
- [ ] `"90"` still parses as 90 seconds (backward compatible)
- [ ] `"1:11:22"` is distinguished from `"11:22"` correctly (3 parts = H:M:S, 2 parts = M:S)
- [ ] Error messages updated from `"expected MM:SS or seconds"` to `"expected HH:MM:SS, MM:SS, or seconds"`
- [ ] Typecheck/vet passes

### US-005: Update notes list time column
**Description:** As a user, I want the time column in the notes list table to show `H:MM:SS`.

**Acceptance Criteria:**
- [ ] `tui/components/noteslist.go` line ~180: `formatTime()` call produces `H:MM:SS`
- [ ] The time column width may need adjusting from 8 to 9 chars to accommodate the longer format (e.g., `1:11:22` = 7 chars vs `71:22` = 5 chars)
- [ ] Column header alignment still looks correct
- [ ] Typecheck/vet passes

## Functional Requirements

- FR-1: All time display throughout the app uses `H:MM:SS` format — hours are not zero-padded, minutes and seconds are zero-padded to 2 digits (e.g., `0:01:30`, `1:11:22`)
- FR-2: The `seek` command accepts `HH:MM:SS`, `MM:SS`, and raw seconds
- FR-3: The `--from` and `--to` filter flags accept `HH:MM:SS`, `MM:SS`, and raw seconds
- FR-4: Input parsing distinguishes formats by colon count: 2 colons = `H:M:S`, 1 colon = `M:S`, 0 colons = seconds
- FR-5: All existing `MM:SS` and raw seconds inputs continue to work (backward compatible)

## Non-Goals

- No changes to how time is stored in the database (still `REAL` seconds)
- No millisecond display (e.g., `1:11:22.5` is out of scope)
- No 12-hour or AM/PM format
- No timezone awareness

## Technical Considerations

- There are currently two duplicate `formatTime` functions: `tui/tui.go:formatTimeString()` and `tui/components/statusbar.go:formatTime()`. Consider consolidating into a single shared function (e.g., in a `tui/util.go` or making the components one the canonical version)
- There are also two duplicate `parseTimeToSeconds` functions: `tui/tui.go` and `cmd/note.go`. Same consolidation opportunity
- The inline formatting in `cmd/note.go`, `cmd/clip.go`, and `cmd/tackle.go` (manual `minutes/60`, `seconds%60` math) should ideally call a shared format function instead of duplicating the logic
- The timeline component (`tui/components/timeline.go`) also calls `formatTime()` — it will pick up the change automatically once the function is updated
- The notes list time column width constant (`timeWidth := 8`) in `noteslist.go` may need to increase to 9 to fit the longer format

## Success Metrics

- All timestamps across TUI and CLI display as `H:MM:SS`
- `seek 1:11:22` works correctly (seeks to 4282 seconds)
- No regressions in existing `MM:SS` and raw seconds input

## Open Questions

- Should the format be `H:MM:SS` (no zero-pad on hours, e.g., `1:11:22`) or `HH:MM:SS` (zero-padded, e.g., `01:11:22`)? This PRD assumes `H:MM:SS` for compactness.
