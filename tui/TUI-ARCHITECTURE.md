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
  components/
    statusbar.go      # StatusBarState, StatusBar()
    timeline.go       # Timeline() — progress bar with event markers
    commandinput.go   # CommandInputState, CommandInput() — : command mode
    noteslist.go      # ListItem, NotesListState, NotesList() — scrollable tag table
    controls.go       # ControlGroup, GetControlGroups(), RenderControlBox()
    statspanel.go     # StatsPanel() — stats summary, event distribution, tackle stats table
    statsview.go      # StatsViewState, PlayerStats, StatsView() — full-screen stats overlay
    help.go           # HelpOverlay() — keybinding reference overlay
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
1. Early returns (quitting, error, help overlay, stats view, mini player)
2. Form overlays (confirm discard, note form, tackle form)
3. Normal multi-column layout:
   a. StatusBar          — full width, 1 line
   b. Columns            — responsive 2/3/4-col grid
   c. Timeline           — full width, 2 lines (progress bar + markers)
   d. CommandInput       — full width, 1 line
```

The final output is: `statusBar + "\n" + columnsView + "\n" + timeline + "\n" + commandInput`

### Column Rendering

Each column is rendered independently by a method on `*Model`:

| Column | Method | Content |
|--------|--------|---------|
| 1 | `renderColumn1(width, height)` | Playback status, selected tag detail |
| 2 | `renderColumn2(width, height)` | Scrollable notes/tackles table |
| 3 | `renderColumn3(width, height)` | Live stats panel (bar graph, tackle stats table) |
| 4 | `renderColumn4(width, height)` | Keybinding control groups (Playback, Navigation, Views) |

Each method wraps its output in `layout.Container{Width, Height}.Render(...)` to
guarantee exact dimensions. The containerized columns are then joined by
`layout.JoinColumns()` with purple `│` separators.

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

### ComputeColumnWidths(termWidth int) (col1, col2, col3, col4 int, showCol2, showCol3, showCol4 bool)

Responsive column width calculation. Constants: `Col1Width = 30`, `Col3Width = 40`, `Col4Width = 30`, `ColMinWidth = 30`, `Col4ShowThreshold = 170`. Column 3 is fixed at 40 cells; Column 2 gets all remaining space.

| Terminal Width | Layout | Column Sizing |
|---------------|--------|---------------|
| >= 170 | 4-column | Col 1 = 30 (fixed), Col 3 = 40 (fixed), Col 4 = 30 (fixed), Col 2 = all remaining space |
| 102 - 169 | 3-column | Col 1 = 30 (fixed), Col 3 = 40 (fixed), Col 2 = all remaining space |
| 61 - 101 | 2-column | Col 1 = 30 (fixed), Col 3 hidden, Col 2 = all remaining space |
| <= 60 | 1-column | Col 1 = 30 (fixed) only |

Hide order: Col 4 first (< 170), then Col 3 (when Col 2 would fall below 30 cells), then Col 2 (when < 30 cells). Col 1 always visible.

### JoinColumns(columns []string, widths []int, height int) string

Joins pre-rendered column strings side by side with purple `│` border separators.
Splits each column into lines, then assembles rows by concatenating corresponding
lines from each column. Columns should already be containerized for exact
dimensions, but includes a fallback for out-of-bounds rows.

## Component Contracts

Each component in `tui/components/` follows the pattern:
- **State struct** — holds all data needed for rendering (e.g., `StatusBarState`, `NotesListState`)
- **Render function** — pure function taking state + dimensions, returning a string
- No Bubble Tea `Model` interface — state is owned by the parent `tui.Model`

### StatusBar (`statusbar.go`)

- **State:** `StatusBarState{Paused, Muted, TimePos, Duration, StepSize, OverlayEnabled}`
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
- **Signature:** `NotesList(state NotesListState, width, height int, currentTimePos float64) string`
- Renders: dynamically-sized scrollable table of notes and tackles (visible row count derived from height parameter), auto-scrolls to current timestamp
- `ListItem` struct: `{ID, Type, TimestampSeconds, Text, Starred, Category, Player, Team}`

### Controls (`controls.go`)

- **Signature:** `GetControlGroups() []ControlGroup` — returns keybinding groups
- **Signature:** `RenderControlBox(group ControlGroup, width int) string` — renders bordered box
- `ControlGroup{Name, SubGroups [][]Control}` — sub-groups separated by dividers

### StatsPanel (`statspanel.go`)

- **Signature:** `StatsPanel(tackleStats []PlayerStats, items []ListItem, width, height int) string`
- Renders: stats summary, bar graph of event distribution, tackle stats table

### StatsView (`statsview.go`)

- **State:** `StatsViewState{Active, Stats []PlayerStats, SortColumn, SortAscending, SelectedRow, ScrollOffset}`
- **Signature:** `StatsView(state StatsViewState, width, height int) string`
- Renders: full-screen sortable stats table overlay

### HelpOverlay (`help.go`)

- **Signature:** `HelpOverlay(width, height int) string`
- Renders: full-screen keybinding reference grouped by function

## Forms Integration (`tui/forms/`)

Forms use the [huh](https://github.com/charmbracelet/huh) library with a custom theme.

### Integration Pattern

1. **Store** `*huh.Form` + result struct in `Model` (nil when form is inactive)
2. **Open:** Call `form.Init()`, return its `tea.Cmd`
3. **Delegate:** In `Update()`, forward ALL messages to the form (not just `KeyMsg` — huh
   needs cursor blink, focus messages, etc.)
4. **Submit:** Check `form.State == huh.StateCompleted` to read bound result values
5. **Cancel:** Check `form.State == huh.StateAborted` to handle Esc
6. **Close:** Set form pointer to `nil` to deactivate

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
| `Purple` | `#5C4F4B` | Borders, dim accents, column separators |
| `BrightPurple` | `#724D7C` | Highlights, focus states |
| `Lavender` | `#AEA47A` | Secondary text |
| `LightLavender` | `#F3DBB2` | Primary text |
| `Pink` | `#D33061` | Headers, special elements |
| `Cyan` | `#3097C6` | Information, interactive elements |
| `Amber` | `#CC8B3F` | Sub-headers |
| `Red` | `#AC3835` | Warnings, errors |
| `Green` | `#A6A75D` | Success messages |

Pre-defined styles: `Background`, `Panel`, `Border`, `Highlight`, `PrimaryText`,
`SecondaryText`, `Warning`, `Success`.
