# PRD: Search & Navigate Notes

## Introduction

Add a search component to column 2 that lets users search and navigate through notes in-place. Rather than filtering the list, search highlights all matches and lets users cycle through them. A new focus system with Tab/Shift+Tab manages focus between Video, Search, and Notes panels. A mode indicator in column 1 shows the current focus and input mode. Vim-style row navigation commands allow jumping to specific rows by number.

## Goals

- Provide in-place search across note text, ID, player, and outcome fields
- Highlight all matches in the notes list with a distinct color for the current match
- Allow cycling through matches with Tab while search is focused
- Add a focus system (Video / Search / Notes) with pink-bordered focused panel
- Display a mode indicator in column 1 showing focus and input mode
- Support Vim-style row navigation (`<number>G`, `0`, `$`)
- Show row numbers alongside DB IDs in the notes list
- Remove left/right arrow keybindings from video seek (keep H/L only)

## User Stories

### US-001: Search input component
**Description:** As a user, I want a search input above the notes list so that I can search for specific notes.

**Acceptance Criteria:**
- [ ] Search input renders at the top of column 2, above the notes list InfoBox
- [ ] Search input width matches column 2 width
- [ ] Input shows a cursor and accepts text input when focused
- [ ] Match indicator at the end of the input shows `[current/total]` (e.g., `[2/5]`)
- [ ] When the input is empty the match indicator is hidden
- [ ] Typing `:` as the first character switches to command mode; the input prefix changes from `/` to `:`
- [ ] Deleting the `:` prefix (backspace on empty command) switches back to search mode
- [ ] Pressing Escape clears input and returns focus to Notes
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-002: Search highlighting in notes list
**Description:** As a user, I want to see all matching rows highlighted so I can see search results in context.

**Acceptance Criteria:**
- [ ] All rows remain visible (no filtering) when a search is active
- [ ] Rows matching the search query are highlighted in a distinct color (e.g., `styles.Amber` background)
- [ ] The "current match" row has a different highlight color (e.g., `styles.Pink` background) to distinguish it from other matches
- [ ] Search matches against: note text, DB ID, player name, and outcome/category fields
- [ ] Search is case-insensitive
- [ ] When search input is cleared, all highlighting is removed
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-003: Cycle through matches
**Description:** As a user, I want to cycle through search matches so I can quickly jump between results.

**Acceptance Criteria:**
- [ ] When search input is focused, Tab moves to the next match and updates the current match indicator
- [ ] Shift+Tab moves to the previous match
- [ ] Cycling wraps around (last match → first match, and vice versa)
- [ ] The notes list auto-scrolls to keep the current match visible
- [ ] The current match index updates in the match indicator `[current/total]`
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-004: Focus system with Tab/Shift+Tab
**Description:** As a user, I want to Tab between panels so I can control which panel receives input.

**Acceptance Criteria:**
- [ ] Three focus targets: Video, Search, Notes
- [ ] Tab cycles focus: Video → Search → Notes → Video
- [ ] Shift+Tab cycles in reverse: Video → Notes → Search → Video
- [ ] The focused panel's InfoBox border renders in `styles.Pink` instead of `styles.Purple`
- [ ] When Search is focused, keystrokes go to the search input (except Tab/Shift+Tab/Escape)
- [ ] When Search is focused, Tab cycles through matches (not focus); use Escape then Tab to change focus
- [ ] When Notes is focused, J/K navigate notes, Vim commands work
- [ ] When Video is focused, H/L seek, Space toggles play/pause, etc.
- [ ] Default focus is Notes on startup
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-005: Mode indicator component in column 1
**Description:** As a user, I want to see which panel is focused and what input mode is active.

**Acceptance Criteria:**
- [ ] Mode indicator renders in column 1, below the video box (above or replacing the summary box area)
- [ ] Left-aligned label: `Focus:` followed by the focused panel name (Video, Search, Notes)
- [ ] Right-aligned label: current mode (Normal, Search, Command)
- [ ] Mode reflects: "Search" when typing in search, "Command" when `:` prefix is active, "Normal" otherwise
- [ ] Uses full column 1 width, rendered inside an InfoBox
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-006: Row numbers in notes list
**Description:** As a user, I want each note row to show both its DB ID and a sequential row number.

**Acceptance Criteria:**
- [ ] Each row displays a 1-indexed row number prefixed with `#` (e.g., `#1`, `#2`, `#3`)
- [ ] The DB ID column remains as-is (shows the database ID)
- [ ] Header row updated to include `#` column label
- [ ] Column widths adjusted to accommodate the new `#` column (row number width: 5 chars)
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-007: Vim-style row navigation
**Description:** As a user, I want to jump to specific rows using Vim-style commands when Notes is focused.

**Acceptance Criteria:**
- [ ] `<number>G` jumps to row number `<number>` (1-indexed), e.g., `5G` goes to row #5
- [ ] `G` alone (no number prefix) jumps to the last row
- [ ] `0` jumps to the first row (row #1)
- [ ] `$` jumps to the last row
- [ ] `gg` jumps to the first row
- [ ] Number input accumulates (e.g., typing `1` then `2` then `G` goes to row #12)
- [ ] Number buffer clears after the command executes or after a timeout/Escape
- [ ] Out-of-range numbers clamp to valid range (first or last row)
- [ ] Notes list auto-scrolls to show the target row
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-008: Remove left/right arrow from video seek
**Description:** As a user, I want arrow keys freed up so they don't conflict with future navigation patterns.

**Acceptance Criteria:**
- [ ] Left arrow key no longer seeks backward in video
- [ ] Right arrow key no longer seeks forward in video
- [ ] `H` and `L` still work for seeking backward/forward
- [ ] Help overlay and control groups updated to remove left/right references
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

### US-009: Update TUI-ARCHITECTURE.md
**Description:** As a developer, I want the architecture doc to reflect the new search, focus, and mode indicator systems.

**Acceptance Criteria:**
- [ ] Document the SearchInput component (state struct, render signature, mode switching)
- [ ] Document the focus system (FocusTarget enum, Tab/Shift+Tab behavior)
- [ ] Document the ModeIndicator component (placement in column 1, rendering)
- [ ] Document the Vim navigation system (number buffer, commands)
- [ ] Update the column 2 rendering section to include search input above notes
- [ ] Update the column 1 rendering section to include mode indicator
- [ ] Update keybindings section (remove left/right arrow from seek)
- [ ] Typecheck/vet passes (`CGO_ENABLED=0`)

## Functional Requirements

- FR-1: Add `SearchInputState` struct with fields: `Active bool`, `Input string`, `CursorPos int`, `Mode string` ("search" or "command"), `Matches []int` (indices into Items), `CurrentMatch int`
- FR-2: Add `FocusTarget` type with constants `FocusVideo`, `FocusSearch`, `FocusNotes`; store `focus FocusTarget` in `Model`
- FR-3: Render search input as a bordered input field spanning column 2 width, above the notes InfoBox; height = 3 lines (top border, input line, bottom border)
- FR-4: Search matches against `ListItem.Text`, `ListItem.ID` (as string), `ListItem.Player`, and `ListItem.Category` fields, case-insensitive
- FR-5: Highlight matching rows with `styles.Amber` background; current match with `styles.Pink` background
- FR-6: Match indicator format: `[M/N]` right-aligned in the search input line, where M = current match (1-indexed), N = total matches
- FR-7: Tab in search mode cycles to next match (wrapping); Shift+Tab to previous match
- FR-8: Tab in normal mode (non-search-focused) cycles focus: Video → Search → Notes → Video; Shift+Tab reverses
- FR-9: Typing `:` as first character in search input switches `Mode` to "command"; backspacing the `:` reverts to "search"
- FR-10: ModeIndicator component renders a single InfoBox in column 1 with left-aligned `Focus: <name>` and right-aligned `<mode>`
- FR-11: Add `RowNumber` (1-indexed sequential) display column to NotesList, separate from DB `ID`
- FR-12: Vim navigation commands (`<number>G`, `gg`, `0`, `$`) operate on 1-indexed row numbers, only when Notes is focused
- FR-13: Accumulate digit keypresses into a number buffer; `G` executes jump; buffer clears on non-digit/non-G input or Escape
- FR-14: Remove `"left"` and `"right"` key handlers from video seek in `Update()`
- FR-15: Focused panel's InfoBox renders with `styles.Pink` border color instead of `styles.Purple`

## Non-Goals

- No fuzzy/regex search — simple substring matching only
- No persistent search history or saved searches
- No search across multiple videos or database-wide search
- No reordering or sorting of notes based on search relevance
- No search-and-replace functionality

## Design Considerations

### Column 2 Layout (with search)
```
╭─ Search ─────────────────────────╮  ← Pink border when focused
│ /query text             [2/5]    │
╰──────────────────────────────────╯
╭─ Notes ──────────────────────────╮  ← Pink border when focused
│  #  ID     Time      Cat   Text  │
│  1  42   0:01:23   tackle  ...   │  ← normal row
│  2  43   0:01:45   note    ...   │  ← amber highlight (match)
│  3  44   0:02:10   tackle  ...   │  ← pink highlight (current match)
│  4  45   0:02:30   note    ...   │  ← normal row
╰──────────────────────────────────╯
```

### Column 1 Layout (with mode indicator)
```
╭─ Video ──────────────────╮
│  ▶ Playing               │
│  0:05:23 / 1:30:00       │
╰──────────────────────────╯
╭─ Mode ───────────────────╮
│ Focus: Notes    Search   │
╰──────────────────────────╯
╭─ Summary ────────────────╮
│  Notes:   12             │
│  Tackles: 8              │
│  Total:   20             │
╰──────────────────────────╯
```

### Search Input Modes
- **Search mode** (default): prefix `/`, searches as you type, Tab cycles matches
- **Command mode**: prefix `:`, triggered by typing `:` as first char, Enter executes command

### Focus Border Colors
- Focused panel: `styles.Pink` (`#D33061`) border
- Unfocused panel: `styles.Purple` (`#5C4F4B`) border (existing default)

## Technical Considerations

- Search input component follows existing component contract: state struct + pure render function
- `RenderInfoBox` needs a `focused bool` parameter (or a variant) to toggle border color
- Column 2 height must be split: 3 lines for search box + remaining for notes list
- Vim number buffer stored in `Model` as `numberBuffer string`; cleared on command execution or mode change
- Search re-evaluation triggered on every keystroke in search input (rebuild `Matches` slice)
- Focus state must be checked in `Update()` to route key messages to the correct handler

## Success Metrics

- User can find a specific note by searching text/ID/player in under 3 seconds
- Match cycling with Tab provides instant visual feedback of current match position
- Focus indicator clearly shows which panel is active at all times
- Vim navigation allows jumping to any row with minimal keystrokes

## Open Questions

- Should search persist when focus moves away from the search input, or clear automatically?
- Should the number buffer have a visible indicator (e.g., show accumulated digits in the status bar)?
- Should `Enter` in search mode jump to the current match and switch focus to Notes?
