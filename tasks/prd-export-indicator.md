# PRD: Export Indicator in Column 1

## Introduction

Add an Export Indicator box at the bottom of Column 1 in the TUI that gives the user
a real-time at-a-glance view of tackle clip export progress for the active video.
The box shows export status (Ready / Processing / Error / Completed), a completed/total
clip count, and a progress bar — all derived from the `notes` and `note_clips` database
tables, refreshed on the same ~250 ms tick as video position polling.

## Goals

- Surface tackle clip export progress inside the TUI without leaving the screen.
- Show a status string that maps aggregate `note_clips.status` values to one of four
  human-readable states: Ready, Processing, Error, Completed.
- Show a `completed / total` ratio so the user knows how many clips have finished.
- Render a proportional ASCII progress bar (0–100%) matching the completion ratio.
- Refresh on the existing video-position tick so no additional ticker is needed.
- Document the new component in `tui/TUI-ARCHITECTURE.md`.

## User Stories

### US-001: Add ExportIndicatorState and DB query

**Description:** As a developer, I need a data structure and database query to compute
export progress so the TUI can display accurate counts.

**Acceptance Criteria:**
- [ ] New `ExportIndicatorState` struct in `tui/components/exportindicator.go`:
  ```go
  type ExportIndicatorState struct {
      TotalTackles    int
      CompletedClips  int
      PendingClips    int
      ErrorClips      int
  }
  ```
- [ ] New SQL query `db/sql/select_export_progress.sql` that returns
      `(total_tackles, completed_clips, pending_clips, error_clips)` for a given
      video path, filtering `notes.category = 'tackle'` and LEFT JOINing `note_clips`:
  ```sql
  SELECT
      COUNT(DISTINCT n.id)                                         AS total_tackles,
      COUNT(CASE WHEN nc.status = 'completed' THEN 1 END)         AS completed_clips,
      COUNT(CASE WHEN nc.status = 'pending'   THEN 1 END)         AS pending_clips,
      COUNT(CASE WHEN nc.status = 'error'     THEN 1 END)         AS error_clips
  FROM notes n
  INNER JOIN videos v ON v.id = n.video_id
  LEFT JOIN note_clips nc ON nc.note_id = n.id
  WHERE v.path = ? AND n.category = 'tackle';
  ```
- [ ] New exported constant `SelectExportProgressSQL` in `db/functions.go` or a new
      `db/sql.go` embed, using `//go:embed` pattern consistent with the existing codebase.
- [ ] New function `db.QueryExportProgress(database *sql.DB, videoPath string) (ExportIndicatorState, error)`
      (placed in `db/functions.go`) that runs the query and scans into the struct.
  - Note: `ExportIndicatorState` lives in `tui/components/`, so `db.QueryExportProgress`
    should return a plain struct or the four ints separately; the mapping to
    `ExportIndicatorState` happens in the TUI layer. Define a `db.ExportProgress` struct
    in `db/models.go` that mirrors the four fields and return that from the DB function.
- [ ] `CGO_ENABLED=0 go vet ./...` passes.

### US-002: ExportIndicator render component

**Description:** As a developer, I need a pure render function that converts
`ExportIndicatorState` into a 5-line InfoBox (border top + 3 content rows + border bottom)
so Column 1 can include it.

**Acceptance Criteria:**
- [ ] New file `tui/components/exportindicator.go` with:
  - `ExportIndicatorState` struct (as above).
  - `ExportStatus() string` method on `ExportIndicatorState` returning one of:
    - `"Completed"` — `TotalTackles > 0 && CompletedClips == TotalTackles && PendingClips == 0 && ErrorClips == 0`
    - `"Processing"` — `PendingClips > 0`
    - `"Error"` — `ErrorClips > 0 && PendingClips == 0`
    - `"Ready"` — anything else (including zero total)
  - `ExportIndicator(state ExportIndicatorState, width int) string` render function.
- [ ] Row layout inside InfoBox (title `"Export"`):
  - **Row 1:** `Status: <value>` — value coloured per state:
    - Ready → `styles.Lavender`
    - Processing → `styles.Amber`
    - Error → `styles.Red`
    - Completed → `styles.Green`
  - **Row 2:** `Clips:  <completed>/<total>` — both numbers in `styles.LightLavender`
  - **Row 3:** ASCII progress bar filling `innerWidth` chars, using `█` for filled and `░`
    for empty, coloured `styles.Cyan`. Fraction = `CompletedClips / TotalTackles`
    (clamp to [0,1]); when `TotalTackles == 0` show an empty bar.
- [ ] The function calls `RenderInfoBox("Export", contentLines, width, false)`.
- [ ] `innerWidth = width - 4` (2 border chars + 2 padding spaces).
- [ ] `CGO_ENABLED=0 go vet ./...` passes.

### US-003: Wire ExportIndicatorState into the Model

**Description:** As a developer, I need the TUI Model to hold and refresh
`ExportIndicatorState` so the render component has live data.

**Acceptance Criteria:**
- [ ] Add `exportIndicator components.ExportIndicatorState` field to `tui.Model` struct
      in `tui/tui.go`.
- [ ] Add a helper method `(m *Model) refreshExportProgress()` that:
  1. Calls `db.QueryExportProgress(m.db, m.videoPath)` (which returns `db.ExportProgress`).
  2. Maps the result fields into `m.exportIndicator` (`ExportIndicatorState`).
  3. Swallows errors silently (leaves state unchanged) to avoid disrupting the TUI on
     transient DB errors.
- [ ] In the existing video-position poll handler (the `tea.Tick` branch in `Update()`
  that refreshes `m.statusBar.TimePos`), call `m.refreshExportProgress()` so it fires
  at the same ~250 ms cadence.
- [ ] `CGO_ENABLED=0 go vet ./...` passes.

### US-004: Render ExportIndicator at the bottom of Column 1

**Description:** As a user, I want to see the Export Indicator box at the bottom of
Column 1 so I can monitor export progress without leaving the tagging view.

**Acceptance Criteria:**
- [ ] In `renderColumn1` (`tui/columns.go`), append the output of
      `components.ExportIndicator(m.exportIndicator, width)` after the Selected Tag
      detail card (i.e., it is the last item added to `lines`).
- [ ] The box is always rendered (even when no clips exist — shows `Status: Ready`,
      `Clips: 0/0`, and an empty progress bar).
- [ ] The overall Column 1 output is still wrapped in
      `layout.Container{Width: width, Height: height}.Render(...)` so exact dimensions
      are guaranteed (Container truncates overflow with `↓ More...`).
- [ ] `CGO_ENABLED=0 go vet ./...` passes.

### US-005: Update TUI-ARCHITECTURE.md

**Description:** As a developer, I want the architecture doc updated to reflect the new
component and the Model field so the documentation stays accurate.

**Acceptance Criteria:**
- [ ] `tui/TUI-ARCHITECTURE.md` updated:
  - Add `exportindicator.go` to the directory structure listing with a brief comment.
  - Add an **ExportIndicator** section under "Component Contracts" documenting:
    - State struct fields.
    - `ExportStatus()` method and its four return values with trigger conditions.
    - `ExportIndicator(state ExportIndicatorState, width int) string` signature.
    - Row layout description (Status, Clips count, progress bar).
  - Update the Column 1 row in the "Column Rendering" table to include "Export indicator
    (bottom)" in the Content column.
  - Add `exportIndicator ExportIndicatorState` to the Model description in the
    "Rendering Pipeline" section or a new Model Fields section if none exists.
- [ ] `db/sql/select_export_progress.sql` mentioned in architecture doc or in a DB
      section if one is added.

## Functional Requirements

- **FR-1:** `ExportIndicatorState` struct with four int fields: `TotalTackles`,
  `CompletedClips`, `PendingClips`, `ErrorClips`.
- **FR-2:** `db.ExportProgress` struct in `db/models.go` mirroring the same four fields,
  returned by `db.QueryExportProgress`.
- **FR-3:** SQL query filters `notes.category = 'tackle'` joined to `videos` by
  `video_id`, LEFT JOINs `note_clips` on `note_id`, groups to a single aggregate row.
- **FR-4:** `ExportStatus()` maps aggregate DB state → `"Ready"` | `"Processing"` |
  `"Error"` | `"Completed"` with the priority: Completed > Error > Processing > Ready.
- **FR-5:** Progress bar uses `█` (filled) and `░` (empty) characters, spans
  `width - 4` chars, fraction = `CompletedClips / TotalTackles` (0 when total = 0).
- **FR-6:** Status string is colour-coded: Ready=Lavender, Processing=Amber, Error=Red,
  Completed=Green (using existing `tui/styles/styles.go` constants).
- **FR-7:** Export indicator box uses `RenderInfoBox("Export", lines, width, false)` —
  unfocused border styling (Purple).
- **FR-8:** Refresh happens inside the existing `tea.Tick` handler (~250 ms) — no new
  ticker or goroutine added.
- **FR-9:** Export indicator is always visible at the bottom of Column 1 regardless of
  whether any clips exist.
- **FR-10:** `TUI-ARCHITECTURE.md` is updated in the same implementation batch.

## Non-Goals

- No interactive controls inside the Export Indicator box (it is read-only).
- No per-clip breakdown list (only aggregate counts).
- No notification or sound when export completes.
- No export triggering from within the indicator (Ctrl+E export flow is separate).
- No separate refresh ticker — reuses the existing poll tick.
- No colour change on the box border when focused (always unfocused Purple border).

## Design Considerations

### Visual Layout (Column 1 width = 30 chars)

```
┌ Export ───────────────────┐
│ Status:         Processing │
│ Clips:              3/12  │
│ ████░░░░░░░░░░░░░░░░░░░░░ │
└───────────────────────────┘
```

- Title: `"Export"` (left-aligned in InfoBox header style, Pink).
- Row 1 label `"Status:"` left-aligned, value right-aligned within inner width.
- Row 2 label `"Clips:"` left-aligned, value right-aligned within inner width.
- Row 3: full-width progress bar (no label), `innerWidth = width - 4` chars.
- Uses `lipgloss` for colour; no manual ANSI escape codes.

### Component Placement in Column 1

Current Column 1 stack (top → bottom):
1. Video status card (`RenderVideoBox`)
2. Mode indicator (`ModeIndicator`)
3. Summary counts (`RenderInfoBox "Summary"`)
4. Selected Tag detail (`RenderInfoBox "Selected Tag"`) — only when item selected
5. **Export indicator (`ExportIndicator`)** ← NEW, always rendered

The Container wraps all of these and truncates overflow with `↓ More...`.

## Technical Considerations

- **DB query embed pattern:** Follow existing pattern — embed `.sql` file with
  `//go:embed db/sql/select_export_progress.sql` and expose as an exported string
  constant, or inline the SQL string in `db/functions.go` if the embed approach adds
  complexity. Be consistent with adjacent code in `db/`.
- **Error handling:** `refreshExportProgress` must not propagate errors to the Update
  loop — log or silently ignore, leave `m.exportIndicator` at previous value.
- **Progress bar width:** `innerWidth = width - 4`. With Col1Width = 30, this gives
  26 chars for the bar. Use integer arithmetic: `filled = int(fraction * float64(innerWidth))`.
- **Colour rendering:** Use `lipgloss.NewStyle().Foreground(styles.XXX).Render(str)` —
  do not hand-roll ANSI escape codes.
- **No CGO:** All build and vet commands must use `CGO_ENABLED=0`.

## Success Metrics

- Export Indicator is visible in Column 1 during a normal tagging session without any
  user action required.
- Status transitions correctly: Ready when no export is running, Processing when clips
  are pending, Error when any clip errored (no pending), Completed when all tackle clips
  are done.
- Progress bar advances from empty to full as clips complete.
- No layout regression — Column 1 height is still exactly constrained by Container.

## Open Questions

- Should the progress bar colour change when in Error state (e.g., Red instead of Cyan)?
  Current spec: always Cyan. Can be changed as a follow-up.
- If `TotalTackles == 0` but `CompletedClips > 0` (data inconsistency), clamp to
  fraction = 1.0. Agreed — clamp `[0.0, 1.0]`.
