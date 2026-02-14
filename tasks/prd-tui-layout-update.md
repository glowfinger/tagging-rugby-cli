# PRD: TUI Layout Update — Containerize Sections & Remove Dividers

## Introduction

Update the TUI layout to create a cleaner, more consistent visual design. All content sections will be wrapped in `RenderInfoBox`-style containers (╭─ Title ─╮ rounded borders), column dividers (`│`) will be removed so columns sit flush, the Playback card will be simplified to use the same single-line header style as Selected Tag, an MPV video status indicator will be added, and the Summary section will move from column 3 to column 1.

## Goals

- Consistent visual language: every content section uses the same `RenderInfoBox` border style
- Cleaner column separation: remove all `│` dividers between columns; columns sit flush
- Simplified Playback card: replace the 3-line tab header + internal dividers with a single-line `RenderInfoBox` header
- Add "Video: Open/Closed" indicator to Playback reflecting actual mpv process state
- Move Summary from column 3 to column 1 (below Playback, above Selected Tag)

## Current Layout (Before)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ StatusBar (full width)                                                          │
├──────────────│──────────────────────│──────────────────│────────────────────────┤
│ Col1 (30)    │ Col2 (flex)          │ Col3 (40)        │ Col4 (30)              │
│              │                      │                  │                        │
│  ┌┤Playback├┐│                      │ Live Stats       │  ┌┤ Playback ├┐       │
│  │ ⏸ Paused ││  Notes List          │                  │  │ Play [Space]│       │
│  ├──────────┤│  (plain table)       │ Summary          │  └────────────┘       │
│  │ Time:... ││                      │  Notes: 5        │  ┌┤Navigation├┐       │
│  ├──────────┤│                      │  Tackles: 3      │  │ Prev [J/↑] │       │
│  │ Overlay  ││                      │  Total: 8        │  └────────────┘       │
│  └──────────┘│                      │                  │  ┌┤ Views ├┐          │
│              │                      │ Event Distrib.   │  │ Stats [S] │        │
│ ╭─Selected──╮│                      │  (bar graph)     │  └───────────┘        │
│ │ #1 Note   ││                      │                  │                        │
│ │ @ 0:01:23 ││                      │ Tackle Stats     │                        │
│ ╰───────────╯│                      │  (table)         │                        │
├──────────────┴──────────────────────┴──────────────────┴────────────────────────┤
│ Timeline (full width)                                                           │
│ CommandInput (full width)                                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Target Layout (After)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ StatusBar (full width)                                                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Col1 (30)     Col2 (flex)            Col3 (40)          Col4 (30)              │
│                                                                                │
│ ╭─Playback──╮                       ╭─Event Distrib.─╮ ┌┤ Playback ├┐        │
│ │ ⏸ Paused  │ ╭─Notes──────────╮   │  (bar graph)   │ │ Play [Space]│        │
│ │ Step: 30s │ │  HDR  TIME ... │   │                │ └────────────┘        │
│ │ Time: ... │ │  #1   0:01 ... │   ╰────────────────╯ ┌┤Navigation├┐        │
│ │ Overlay:  │ │  #2   0:02 ... │                       │ Prev [J/↑] │        │
│ │ Video: Open│ │  ...           │   ╭─Tackle Stats───╮ └────────────┘        │
│ ╰───────────╯ │                │   │  Player Tot ... │ ┌┤ Views ├┐           │
│               ╰────────────────╯   │  TOTAL   8  ... │ │ Stats [S] │         │
│ ╭─Summary───╮                      ╰────────────────╯ └───────────┘         │
│ │ Notes: 5  │                                                                │
│ │ Tackles: 3│                                                                │
│ │ Total: 8  │                                                                │
│ ╰───────────╯                                                                │
│                                                                                │
│ ╭─Selected──╮                                                                 │
│ │ #1 Note   │                                                                 │
│ │ @ 0:01:23 │                                                                 │
│ ╰───────────╯                                                                 │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Timeline (full width)                                                           │
│ CommandInput (full width)                                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
```

Key differences:
- **No `│` column dividers** — columns sit flush
- **Playback** uses `RenderInfoBox` style (╭─ Playback ─╮), no internal dividers, all content lines listed plainly
- **Video: Open/Closed** line added to Playback, reflecting actual mpv process state
- **Summary** moved from column 3 to column 1 (between Playback and Selected Tag), wrapped in `RenderInfoBox`
- **Notes List** (column 2) wrapped in `RenderInfoBox` container
- **Event Distribution** (column 3) wrapped in `RenderInfoBox` container
- **Tackle Stats** (column 3) wrapped in `RenderInfoBox` container
- Column 4 controls unchanged (already use bordered boxes)

## User Stories

### US-001: Replace Playback tab-header with RenderInfoBox style
**Description:** As a user, I want the Playback card to use the same clean single-line header style as Selected Tag so the UI looks consistent.

**Acceptance Criteria:**
- [ ] `RenderMiniPlayer` renders using `RenderInfoBox("Playback", contentLines, width)` instead of the 3-line tab header (lines 1/2/3) and internal `├───┤` dividers
- [ ] Content lines are: play/pause + step size, time position, overlay status, video status (see US-002) — all plain rows, no horizontal dividers between them
- [ ] Mute icon still appears on the status line when muted
- [ ] The card is still centered horizontally when `termWidth > cardWidth`
- [ ] The `showWarning` mini-player warning line still works
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

**Notes:** The current `RenderMiniPlayer` in `tui/components/controls.go` builds a 3-line tab header (line1: ` ┌──┐`, line2: `┌┤ Playback ├┐`, line3: `│└──┘└───┐`) and uses `├───┤` dividers between content rows. Replace all of this with a call to `RenderInfoBox` which produces the simpler `╭─ Playback ─╮` style. Remove the two internal divider lines so all content rows are contiguous.

### US-002: Add "Video: Open/Closed" line to Playback card
**Description:** As a user, I want to see whether the mpv video player is currently open so I know if my video is loaded.

**Acceptance Criteria:**
- [ ] A new `VideoOpen bool` field is added to `StatusBarState` in `tui/components/statusbar.go`
- [ ] `RenderMiniPlayer` includes a "Video: Open" or "Video: Closed" line after the overlay line
- [ ] The `VideoOpen` field is set to `true` when the mpv process is running and `false` when it is not
- [ ] The field is updated in `tui.go` wherever `m.statusBar` is populated (check mpv cmd state: `m.mpvCmd != nil` and process is alive)
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

**Notes:** The mpv process handle is stored in the `Model` (look for `mpvCmd` or similar). Check whether the process is running to determine Open/Closed. The status line should use the same `textStyle` as other Playback content lines.

### US-003: Move Summary from column 3 to column 1
**Description:** As a user, I want to see the tag summary (Notes/Tackles/Total counts) in column 1 below Playback so all status info is in one place.

**Acceptance Criteria:**
- [ ] `renderColumn1` renders a Summary section between Playback and Selected Tag
- [ ] Summary is wrapped in `RenderInfoBox("Summary", contentLines, width)` using the same style as Selected Tag
- [ ] Summary content shows: `Notes: N`, `Tackles: N`, `Total: N` (same data as current column 3 summary)
- [ ] `renderColumn1` receives access to `m.notesList.Items` to compute counts (it already has access via `m.notesList`)
- [ ] The Summary section is removed from `StatsPanel` in `tui/components/statspanel.go` (column 3 starts directly with Event Distribution)
- [ ] The "Live Stats" title at the top of `StatsPanel` is removed (no longer needed since sections are individually containerized)
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

**Notes:** Count logic: iterate `m.notesList.Items`, count `ItemTypeNote` vs `ItemTypeTackle`. Use the same `infoStyle` (LightLavender foreground) for the count lines. In `StatsPanel`, remove the "Live Stats" title, "Summary" header, the three count lines, and the blank line after — so the function starts directly with Event Distribution.

### US-004: Wrap Notes List in RenderInfoBox container
**Description:** As a user, I want the notes list to have a visible border and title so it's clearly delineated.

**Acceptance Criteria:**
- [ ] `renderColumn2` wraps the `NotesList` output in `RenderInfoBox("Notes", notesLines, width)`
- [ ] The height passed to `NotesList` is reduced to account for the 2 extra lines consumed by the InfoBox border (top + bottom)
- [ ] Scrolling and auto-scroll behavior still work correctly
- [ ] Empty state ("No notes yet") renders inside the box
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

**Notes:** `RenderInfoBox` adds 2 lines (top border + bottom border). So pass `height - 2` to `NotesList` for the inner content height. The NotesList output (multi-line string) needs to be split into `contentLines []string` for `RenderInfoBox`. The outer `layout.Container` call remains to guarantee exact column dimensions.

### US-005: Wrap Event Distribution and Tackle Stats in RenderInfoBox containers
**Description:** As a user, I want each stats section to have its own bordered container so the column 3 layout matches the containerized style used everywhere else.

**Acceptance Criteria:**
- [ ] `StatsPanel` wraps the Event Distribution bar graph in `RenderInfoBox("Event Distribution", lines, width)`
- [ ] `StatsPanel` wraps the Tackle Stats table in `RenderInfoBox("Tackle Stats", lines, width)`
- [ ] The "Event Distribution" and "Tackle Stats" plain-text headers (currently rendered by `summaryHeaderStyle`) are removed (the InfoBox title replaces them)
- [ ] The "Live Stats" title and blank line at the top of StatsPanel are removed
- [ ] Bar graph and tackle table content render correctly inside the boxes
- [ ] Empty states ("No events yet", "No tackle data") render inside their respective boxes
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

**Notes:** Build each section's content lines as a `[]string`, then call `RenderInfoBox(title, lines, width)`. Join the two boxes with a blank line between them. The Summary section will already be gone (US-003).

### US-006: Remove column dividers from JoinColumns
**Description:** As a user, I want the columns to sit flush without `│` separators for a cleaner look.

**Acceptance Criteria:**
- [ ] `JoinColumns` in `tui/layout/columns.go` no longer inserts the purple `│` border string between columns
- [ ] Columns are concatenated directly (flush, no separator character)
- [ ] `ComputeColumnWidths` border width calculations are updated: the `borders` variable that currently accounts for separator characters (1 per separator) is set to 0 in all branches
- [ ] Column 2 gains the extra width that was previously consumed by separator characters
- [ ] Layout does not break at any of the responsive breakpoints (1-col, 2-col, 3-col, 4-col)
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

**Notes:** In `ComputeColumnWidths`, `borders` is currently set to 2 or 3 depending on column count. Set all `borders` assignments to 0. In `JoinColumns`, replace the `borderStr` join separator with an empty string `""`. This gives columns 2-3 more cells of width at every breakpoint.

## Functional Requirements

- FR-1: `RenderMiniPlayer` must use `RenderInfoBox("Playback", ...)` for its border/header, removing the 3-line tab header and 2 internal `├───┤` dividers
- FR-2: `StatusBarState` must include `VideoOpen bool`; `RenderMiniPlayer` must render "Video: Open" or "Video: Closed" based on this field
- FR-3: `renderColumn1` must render Playback → Summary → Selected Tag (top to bottom), with Summary in a `RenderInfoBox`
- FR-4: `StatsPanel` must not render Summary, "Live Stats" title, or plain-text section headers
- FR-5: `renderColumn2` must wrap `NotesList` output in `RenderInfoBox("Notes", ...)`
- FR-6: `StatsPanel` must wrap Event Distribution in `RenderInfoBox("Event Distribution", ...)`
- FR-7: `StatsPanel` must wrap Tackle Stats in `RenderInfoBox("Tackle Stats", ...)`
- FR-8: `JoinColumns` must concatenate columns flush (no `│` separator)
- FR-9: `ComputeColumnWidths` must set `borders = 0` in all code paths
- FR-10: All existing responsive breakpoints (1/2/3/4-column) must continue to work correctly

## Non-Goals

- No changes to column 4 (control boxes already have borders)
- No changes to StatusBar, Timeline, or CommandInput components
- No changes to responsive breakpoint thresholds or column width constants
- No changes to form overlays or full-screen overlays (StatsView, HelpOverlay)
- No color palette changes
- No new keybindings

## Technical Considerations

- `RenderInfoBox` already exists in `tui/components/controls.go` and handles border rendering — reuse it as-is
- `RenderMiniPlayer` centering logic (padding when `termWidth > cardWidth`) must be preserved after the refactor
- Removing column borders in `ComputeColumnWidths` frees 2-3 cells of width — column 2 absorbs this automatically since it's flex-width
- The mpv process state can be checked via the existing `m.mpvCmd` field (or equivalent) — look for `exec.Cmd` pointer and check `cmd.Process` / `cmd.ProcessState`

## Success Metrics

- All content sections have consistent `RenderInfoBox` borders
- No `│` column separators visible between columns
- Playback card is visually simpler (single-line header, no internal dividers)
- Video open/closed status is visible and accurate
- Summary counts visible in column 1 without scrolling
- No layout misalignment at any terminal width

## Open Questions

- Should the Summary box in column 1 show any additional stats beyond Notes/Tackles/Total counts?
- Should the "Video: Open/Closed" text use color coding (e.g., green for Open, red for Closed)?
