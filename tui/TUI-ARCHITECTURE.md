# TUI Architecture

This document describes the terminal user interface layer of tagging-rugby-cli.
It covers the rendering pipeline, layout system, component contracts, forms
integration, and style system.

## Overview

The TUI is built on:

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — Elm-architecture framework (`Model`, `Update`, `View`)
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — ANSI styling and width measurement
- **[huh](https://github.com/charmbracelet/huh)** — Declarative form components (note input, tackle wizard, confirm dialogs)
- **[charmbracelet/x/ansi](https://github.com/charmbracelet/x)** — Grapheme-aware string truncation (emoji, East-Asian characters)

## Directory Structure

```
tui/
  tui.go              # Model struct, Init, Update, View (orchestration)
  columns.go          # renderColumn1/2/3/4, formatStepSize (column render methods)
  focus.go            # FocusTarget type (FocusVideo, FocusSearch, FocusNotes), cycleFocus()
  components/
    statusbar.go      # StatusBarState, StatusBar()
    timeline.go       # Timeline() — progress bar with event markers
    commandinput.go   # CommandInputState, CommandInput() — : command mode
    noteslist.go      # ListItem, NotesListState, NotesList() — scrollable tag table
    searchinput.go    # SearchInputState, SearchInput() — search/command input with match indicator
    modeindicator.go  # ModeIndicator() — displays current focus and input mode
    controls.go       # ControlGroup, GetControlGroups(), ControlGroupLines(), RenderInfoBox(), RenderControlBox() (Deprecated)
    statspanel.go     # StatsPanel() — stats summary, event distribution, tackle stats table
    statsview.go      # StatsViewState, PlayerStats, StatsView() — sortable stats view (renders in Column 2)
    help.go           # HelpOverlay() — keybinding reference (renders in Column 2)
    exportindicator.go # ExportIndicatorState, ExportIndicator() — tackle clip export progress box
  forms/
    theme.go          # Theme() — custom huh theme matching the Ciapre palette
    noteform.go       # NoteFormResult, NewNoteForm() — note input form
    tackleform.go     # TackleFormResult, NewTackleForm() — multi-step tackle wizard
    forms.go          # NewConfirmDiscardForm() — discard confirmation dialog
  styles/
    styles.go         # Ciapre colour constants and pre-defined Lip Gloss styles
  layout/
    helpers.go        # PadToWidth(), NormalizeLines() — low-level text utilities
    container.go      # Container{Width, Height}.Render() — exact bounding box
    columns.go        # ComputeColumnWidths(), JoinColumns() — responsive multi-column
```

## Rendering Pipeline

`View()` in `tui.go` orchestrates the full screen render each frame:

```
1. Early returns (quitting, error)
2. Normal multi-column layout — forms and overlays render inside Column 2:
   a. StatusBar          — full width, 1 line
   b. Columns            — responsive 2/3/4-col grid
                           (Column 2 shows form/overlay content when active)
   c. Timeline           — full width, 2 lines (progress bar + markers)
   d. CommandInput       — full width, 1 line
```

The final output is: `statusBar + "\n" + columnsView + "\n" + timeline + "\n" + commandInput`

> **Note:** Help overlay, stats view, and all huh forms (note, tackle, confirm discard)
> render **inside Column 2** rather than as full-screen takeovers. See
> [Overlay-in-Column-2 Pattern](#overlay-in-column-2-pattern) below.

### Column Rendering

Each column is rendered independently by a method on `*Model`:

| Column | Method | Content |
|--------|--------|---------|
| 1 | `renderColumn1(width, height)` | Video status, mode indicator, summary counts, selected tag detail, export indicator (bottom) |
| 2 | `renderColumn2(width, height)` | **Conditional:** active form/overlay (note form, tackle form, confirm discard, help overlay, stats view) when any is open; otherwise search input + scrollable notes/tackles table |
| 3 | `renderColumn3(width, height)` | Event distribution bar graph, tackle stats table — **hidden when any form/overlay is active** |
| 4 | `renderColumn4(width, height)` | Keybinding control groups via RenderInfoBox (Playback, Navigation, Views) — remains visible when form/overlay is active |

Each method wraps its output in `layout.Container{Width, Height}.Render(...)` to
guarantee exact dimensions. The containerized columns are then joined by
`layout.JoinColumns()` flush with no separators.

## Layout System (`tui/layout/`)

### PadToWidth(s string, width int) string

Pads or truncates a string to exactly `width` visual columns. Uses
`lipgloss.Width()` for ANSI-aware measurement and `ansi.Truncate()` for
grapheme-aware truncation (handles emoji and East-Asian wide characters).

### NormalizeLines(lines []string, height int) []string

Pads or truncates a string slice to exactly `height` entries. Excess lines
are dropped; missing lines are filled with empty strings.

### Container{Width, Height}.Render(content string) string

Constrains content to an exact `Width x Height` bounding box:
- Splits content on newlines
- When content exceeds `Height`, truncates and replaces the last visible line
  with a styled `↓ More...` scroll indicator (using `styles.Purple`)
- When content has fewer lines than `Height`, appends empty lines
- Each line is padded/truncated to `Width` via `PadToWidth`
- Output is always exactly `Height` lines, each exactly `Width` visual columns

### ComputeColumnWidths(termWidth int, overlayActive bool) (col1, col2, col3, col4 int, showCol2, showCol3, showCol4 bool)

Responsive column width calculation with no border separator overhead (borders = 0). Constants: `Col1Width = 30`, `Col3Width = 40`, `Col4Width = 30`, `ColMinWidth = 30`, `Col4ShowThreshold = 170`. Column 3 is fixed at 40 cells; Column 2 gets all remaining space.

`overlayActive` is `true` when any form or overlay is rendering in Column 2 (`m.noteForm != nil || m.tackleForm != nil || m.confirmDiscardForm != nil || m.showHelp || m.statsView.Active`). When true, Column 3 is always hidden regardless of terminal width so Column 2 gets extra space.

**Normal layout** (`overlayActive == false`):

| Terminal Width | Layout | Column Sizing |
|---------------|--------|---------------|
| >= 170 | 4-column | Col 1 = 30 (fixed), Col 3 = 40 (fixed), Col 4 = 30 (fixed), Col 2 = all remaining space |
| 102 - 169 | 3-column | Col 1 = 30 (fixed), Col 3 = 40 (fixed), Col 2 = all remaining space |
| 61 - 101 | 2-column | Col 1 = 30 (fixed), Col 3 hidden, Col 2 = all remaining space |
| <= 60 | 1-column | Col 1 = 30 (fixed) only |

**Overlay layout** (`overlayActive == true`):

| Terminal Width | Layout | Column Sizing |
|---------------|--------|---------------|
| >= 170 | Col 1 + Col 2 (form) + Col 4 | Col 1 = 30, Col 4 = 30, Col 2 = termWidth − 60 |
| 61 - 169 | Col 1 + Col 2 (form) | Col 1 = 30, Col 2 = termWidth − 30 |
| <= 60 | Col 1 only (forms blocked) | Col 1 = 30 (form/overlay activation is prevented) |

Hide order (normal): Col 4 first (< 170), then Col 3 (when Col 2 would fall below 30 cells), then Col 2 (when < 30 cells). Col 1 always visible.

### JoinColumns(columns []string, widths []int, height int) string

Joins pre-rendered column strings side by side flush with no separators.
Splits each column into lines, then assembles rows by concatenating corresponding
lines from each column. Columns should already be containerized for exact
dimensions, but includes a fallback for out-of-bounds rows.

## Component Contracts

Each component in `tui/components/` follows the pattern:
- **State struct** — holds all data needed for rendering (e.g., `StatusBarState`, `NotesListState`)
- **Render function** — pure function taking state + dimensions, returning a string
- No Bubble Tea `Model` interface — state is owned by the parent `tui.Model`

### StatusBar (`statusbar.go`)

- **State:** `StatusBarState{Paused, Muted, TimePos, Duration, StepSize, OverlayEnabled, VideoOpen}`
- **Signature:** `StatusBar(state StatusBarState, width int) string`
- Renders: play/pause icon, timestamp, duration, step size, mute/overlay indicators

### Timeline (`timeline.go`)

- **Signature:** `Timeline(timePos, duration float64, items []ListItem, width int) string`
- Renders: 2-line progress bar with note/tackle markers at their timestamps

### CommandInput (`commandinput.go`)

- **State:** `CommandInputState{Active, Input, CursorPos, Result, IsError}`
- **Signature:** `CommandInput(state CommandInputState, width int) string`
- Renders: `:` prompt when active, result messages, or help hint

### NotesList (`noteslist.go`)

- **State:** `NotesListState{Items []ListItem, SelectedIndex, ScrollOffset}`
- **Signature:** `NotesList(state NotesListState, width, height int, currentTimePos float64, matches []int, currentMatch int, query string) string`
- Renders: dynamically-sized scrollable table with right-aligned row numbers (1, 2, ...), notes and tackles
- Row number column: 5 chars wide, right-aligned, no `#` prefix (e.g., `  1`, ` 12`, `123`)
- **Inline match highlighting:** matched rows get a subtle `MatchBg` background; the matching substring within each field is highlighted with Amber (match) or Pink (current match) background
- Highlight priority: current match inline > match inline > selected (BrightPurple full row) > default
- `ListItem` struct: `{ID, Type, TimestampSeconds, Text, Starred, Category, Player, Team}`

### SearchInput (`searchinput.go`)

- **State:** `SearchInputState{Input, CursorPos, Mode ("search"|"command"), Matches []int, CurrentMatch}`
- **Signature:** `SearchInput(state SearchInputState, width int, focused bool) string`
- Renders: bordered input box (3 lines) with `/` or `:` prefix, cursor, and [M/N] match indicator
- Mode switching: typing `:` on empty search input switches to command mode; backspace on empty command input switches back
- Methods: `InsertChar`, `Backspace`, `MoveCursorLeft`, `MoveCursorRight`, `Clear`

### ModeIndicator (`modeindicator.go`)

- **Signature:** `ModeIndicator(focusName, mode string, width int) string`
- Renders: 4-line InfoBox (borders + 2 content lines): `Focus: <panel>` on line 1, `Mode: <mode>` on line 2, each with label left-aligned and value right-aligned
- Placed in column 1 between video box and summary box

### Controls (`controls.go`)

- **Signature:** `GetControlGroups() []ControlGroup` — returns keybinding groups
- **Signature:** `RenderInfoBox(title string, contentLines []string, width int, focused bool) string` — generic bordered box; when focused=true, border uses Pink instead of Purple
- **Signature:** `RenderVideoBox(state StatusBarState, width int, showWarning bool, focused bool) string` — renders video status card using `RenderInfoBox` style; focused=true gives Pink border when Video panel has focus
- **Signature:** `ControlGroupLines(group ControlGroup, innerWidth int) []string` — converts a ControlGroup into content lines for RenderInfoBox; sub-groups separated by blank lines
- **Signature:** `RenderControlBox(group ControlGroup, width int) string` — (Deprecated) renders bordered box with square tab header; use RenderInfoBox with ControlGroupLines instead
- `ControlGroup{Name, SubGroups [][]Control}` — sub-groups separated by blank lines (RenderInfoBox) or dividers (deprecated RenderControlBox)

### StatsPanel (`statspanel.go`)

- **Signature:** `StatsPanel(tackleStats []PlayerStats, items []ListItem, width, height int) string`
- Renders: event distribution bar graph and tackle stats table, each wrapped in `RenderInfoBox`

### StatsView (`statsview.go`)

- **State:** `StatsViewState{Active, Stats []PlayerStats, SortColumn, SortAscending, SelectedRow, ScrollOffset}`
- **Signature:** `StatsView(state StatsViewState, width, height int) string`
- Renders: sortable stats table (placed in Column 2 when active)

### HelpOverlay (`help.go`)

- **Signature:** `HelpOverlay(width, height int) string`
- Renders: keybinding reference grouped by function (placed in Column 2 when active)

### ExportIndicator (`exportindicator.go`)

- **State:** `ExportIndicatorState{TotalTackles, CompletedClips, PendingClips, ErrorClips int}`
- **Method:** `ExportStatus() string` — maps aggregate DB state to one of four labels:
  - `"Completed"` — `TotalTackles > 0 && CompletedClips == TotalTackles && PendingClips == 0 && ErrorClips == 0`
  - `"Processing"` — `PendingClips > 0`
  - `"Error"` — `ErrorClips > 0 && PendingClips == 0`
  - `"Ready"` — all other cases (including zero total)
- **Signature:** `ExportIndicator(state ExportIndicatorState, width int) string`
- Renders: `RenderInfoBox("Export", ...)` with 3 content rows:
  - Row 1 — `Status: <value>` — colour-coded: Ready=Lavender, Processing=Amber, Error=Red, Completed=Green
  - Row 2 — `Clips:  <completed>/<total>` — both numbers in LightLavender
  - Row 3 — ASCII progress bar (`█` filled, `░` empty) spanning `width-4` chars, coloured Cyan; fraction = `CompletedClips/TotalTackles` (clamped [0,1]; empty bar when total=0)
- Placed as the last (bottom) item in Column 1; always rendered regardless of clip count.
- Refreshed in the Model via `refreshExportProgress()` called from the existing ~250 ms video-position tick handler.
- DB source: `db.QueryExportProgress(db, videoPath)` → `db.ExportProgress{TotalTackles, CompletedClips, PendingClips, ErrorClips}` backed by `db/sql/select_export_progress.sql`.

## Column 2 Content Replacement Pattern

Forms and interactive views are rendered as full-width, full-height containers that
**replace** Column 2's normal content (search bar + notes list). They are not floating
overlays — they are statically placed inside the Column 2 slot and sized to fill it
exactly via `layout.Container{Width: width, Height: height}.Render(...)`.

This keeps Column 1 (video status, export indicator) and Column 4 (keybinding reference)
visible at all times.

### Items that replace Column 2

| Item | Trigger key | Type |
|------|-------------|------|
| Note form | `N` | huh form |
| Tackle wizard | `T` | huh form |
| Confirm discard | automatic (when editing) | huh form |
| Help | `?` | static render |
| Stats view | `S` | interactive render |

### Activation guard (narrow mode)

Before replacing Column 2, key handlers check that Column 2 is visible (`termWidth >= 61`).
If not, the key press is silently ignored. This prevents rendering a form into a
zero-width or negative-width container.

### `overlayActive` flag

`View()` computes:

```go
overlayActive := m.noteForm != nil || m.tackleForm != nil || m.confirmDiscardForm != nil || m.showHelp || m.statsView.Active
```

and passes it to `ComputeColumnWidths`. This causes Column 3 to hide and Column 2 to
expand into the freed space.

### Column 2 conditional rendering

`renderColumn2(width, height int)` checks active state at the top, in priority order,
and returns the first match wrapped in `layout.Container{Width: width, Height: height}`:

1. `m.confirmDiscardForm != nil` → `Container.Render(m.confirmDiscardForm.View())`
2. `m.noteForm != nil` → `Container.Render(m.noteForm.View())`
3. `m.tackleForm != nil` → `Container.Render(m.tackleForm.View())`
4. `m.showHelp` → `Container.Render(HelpOverlay(width, height))`
5. `m.statsView.Active` → `Container.Render(StatsView(m.statsView, width, height))`
6. Otherwise → search input + notes list (normal content)

Confirm Discard is checked first because both it and its parent form (note or tackle)
may be non-nil simultaneously — Confirm Discard wins the display slot.

### Dismissal

**Esc** is the single key to exit any active form or view. See [Global Keys](#global-keys) below.

---

## Focus System (`tui/focus.go`)

The TUI has a focus system that routes keyboard input to the correct panel.

### FocusTarget Type

`FocusTarget` is an int type with three constants:
- `FocusVideo` (0) — video panel receives playback keys (Space, H/L, Ctrl+H/L, etc.)
- `FocusSearch` (1) — search input receives text input, mode switching, match cycling
- `FocusNotes` (2) — notes list receives navigation keys (J/K, Enter, Vim commands)

Default focus is `FocusNotes`.

### Focus Cycling

- **Tab** cycles forward: Video → Search → Notes → Video
- **Shift+Tab** cycles backward: Video → Notes → Search → Video
- When in FocusSearch with active matches, Tab/Shift+Tab cycle through matches instead
- Tab/Shift+Tab always handled regardless of focus

### Global Keys

`Ctrl+C` (quit) works in all focus modes. The following keys are guarded — they work in FocusVideo and FocusNotes but are passed to the search input in FocusSearch: `?` (help), `S` (stats), `N` (note form), `T` (tackle form)

## Vim Navigation (FocusNotes)

When notes list is focused, Vim-style navigation commands are available:

| Command | Action |
|---------|--------|
| `0` (empty buffer) | Jump to first row |
| `$` | Jump to last row |
| `G` (empty buffer) | Jump to last row |
| `nG` (digits + G) | Jump to row n (1-indexed) |
| `gg` (two g presses) | Jump to first row |
| `J`/`K` | Move up/down one row |

Digit keys accumulate in a number buffer. Any non-digit/non-G key clears the buffer.

## Keybindings

### Global Keys

These keys are handled before any focus-specific handler:

| Key | Action |
|-----|--------|
| `Ctrl+C` | Quit |
| `Esc` | Unified dismiss handler, checked in priority order: 1) `m.confirmDiscardForm != nil` → close it and re-open parent form; 2) `m.noteForm != nil` → trigger huh abort (discard guard may show confirm dialog); 3) `m.tackleForm != nil` → trigger huh abort; 4) `m.showHelp` → set `m.showHelp = false`; 5) `m.statsView.Active` → set `m.statsView.Active = false`; 6) `FocusSearch` → clear search input and return to `FocusNotes`; 7) otherwise → fall through to other handlers (e.g. cancel command mode) |

### Video Focus (FocusVideo)
- `Space` — toggle play/pause
- `H` — seek backward by step size
- `L` — seek forward by step size
- `Ctrl+H` / `Ctrl+L` — frame step backward/forward
- `,`/`<` and `.`/`>` — decrease/increase step size
- `M` — toggle mute
- `O` — toggle overlay

### Search Focus (FocusSearch)
- Printable chars — insert into search input
- `Backspace` — delete character
- `Left`/`Right` — move cursor
- `Escape` — clear search, return to FocusNotes
- `:` (on empty input) — switch to command mode
- `Tab`/`Shift+Tab` — cycle through matches (if any)
- `Enter` (command mode) — execute command

### Notes Focus (FocusNotes)
- `J`/`K` — navigate up/down
- `Enter` — jump to selected item timestamp
- `E` — edit selected tackle
- `X` — delete selected item
- `:` — enter command mode
- Vim commands (see above)

## Forms Integration (`tui/forms/`)

Forms use the [huh](https://github.com/charmbracelet/huh) library with a custom theme.

### Integration Pattern

1. **Store** `*huh.Form` + result struct in `Model` (nil when form is inactive)
2. **Open:** Call `form.Init()`, return its `tea.Cmd`; opening is blocked if terminal width < 61
3. **Delegate:** In `Update()`, forward ALL messages to the form (not just `KeyMsg` — huh
   needs cursor blink, focus messages, etc.)
4. **Submit:** Check `form.State == huh.StateCompleted` to read bound result values
5. **Cancel:** Check `form.State == huh.StateAborted` to handle Esc (huh sets this state when Esc is pressed)
6. **Close:** Set form pointer to `nil` to deactivate
7. **Render:** Form output is placed in Column 2 via `renderColumn2`. The parent `View()` sets
   `overlayActive = true` which hides Column 3 and expands Column 2. Form content is wrapped in
   `layout.Container{Width: col2Width, Height: col2Height}.Render(m.form.View())`.

### Available Forms

| Form | Constructor | Result Type | Purpose |
|------|------------|-------------|---------|
| Note form | `NewNoteForm(timestamp, result)` | `NoteFormResult{Text, Category, Player, Team}` | Create/edit timestamped notes |
| Tackle wizard | `NewTackleForm(timestamp, result)` | `TackleFormResult{Player, Attempt, Outcome, ...}` | Multi-step tackle entry |
| Confirm discard | `NewConfirmDiscardForm(discard)` | `*bool` | Confirm before discarding form data |

### Theme (`theme.go`)

`Theme()` returns a `*huh.Theme` that matches the Ciapre colour palette. It
customizes focused/blurred border styles, title colours, and selection indicators
to blend with the rest of the TUI.

## Styles (`tui/styles/styles.go`)

The colour palette is **Ciapre** (warm, earthy) from the Gogh terminal themes project.

| Constant | Hex | Usage |
|----------|-----|-------|
| `DeepPurple` | `#191C27` | Main background |
| `DarkPurple` | `#181818` | Secondary dark background |
| `Purple` | `#5C4F4B` | Borders, dim accents |
| `BrightPurple` | `#724D7C` | Highlights, focus states |
| `Lavender` | `#AEA47A` | Secondary text |
| `LightLavender` | `#F3DBB2` | Primary text |
| `Pink` | `#D33061` | Headers, special elements |
| `Cyan` | `#3097C6` | Information, interactive elements |
| `Amber` | `#CC8B3F` | Sub-headers |
| `Red` | `#AC3835` | Warnings, errors |
| `Green` | `#A6A75D` | Success messages |
| `MatchBg` | `#2A2D3A` | Subtle background for search-matched rows |

Pre-defined styles: `Background`, `Panel`, `Border`, `Highlight`, `PrimaryText`,
`SecondaryText`, `Warning`, `Success`.
