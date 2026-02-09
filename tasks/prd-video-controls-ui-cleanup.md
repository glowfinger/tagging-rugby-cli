# PRD: Better Video Controls & UI Cleanup

## Introduction

The TUI currently uses a limited set of vim-style keybindings for video playback and navigation, and the three-column layout has alignment issues when the terminal is resized. This PRD addresses two focused improvements: (1) adding duplicate/alternative keybindings so arrow keys and punctuation keys work alongside the existing H/J/K/L keys, and (2) fixing the column layout so columns stay inline and the live stats panel (column 3) shrinks gracefully before column 1.

## Goals

- Add arrow key and punctuation key alternatives for all core video/navigation controls
- Fix column alignment so all three columns render inline at the same height
- Implement responsive column sizing where column 3 (live stats) shrinks first when the terminal narrows
- Update controls display and help overlay to reflect the new keybindings

## User Stories

### US-065: Add arrow key alternatives for seek and navigation
**Description:** As a user, I want to use Left/Right arrow keys to seek backward/forward and Up/Down arrow keys to navigate the notes list, so that I have intuitive controls alongside the vim-style H/J/K/L keys.

**Acceptance Criteria:**
- [ ] `Left` arrow key seeks backward by current step size (same as `H`)
- [ ] `Right` arrow key seeks forward by current step size (same as `L`)
- [ ] `Up` arrow key selects previous item in list (same as `J`)
- [ ] `Down` arrow key selects next item in list (same as `K`)
- [ ] Existing H/L/J/K bindings continue to work unchanged
- [ ] Arrow keys do NOT trigger in command mode (command mode already handles left/right for cursor movement)
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

### US-066: Add comma/period alternatives for step size control
**Description:** As a user, I want to use `,` and `.` keys (in addition to `<` and `>`) to decrease/increase step size, so I don't need to hold Shift.

**Acceptance Criteria:**
- [ ] `,` key decreases step size (same as `<`)
- [ ] `.` key increases step size (same as `>`)
- [ ] Existing `<` and `>` bindings continue to work unchanged
- [ ] Keys do NOT trigger in command mode
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

### US-067: Update controls display in column 1 to show new keybindings
**Description:** As a user, I want the controls panel to show all available key alternatives so I can discover them without opening help.

**Acceptance Criteria:**
- [ ] Column 1 controls list shows alternative keys (e.g., "Back [H/Left]", "Forward [L/Right]")
- [ ] Controls display in `renderColumn1()` uses `GetControlGroups()` from `controls.go` instead of a duplicate hardcoded list
- [ ] `GetControlGroups()` is updated with the new shortcut text
- [ ] The form-overlay controls display (`ControlsDisplay()`) also reflects the updated shortcuts
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

### US-068: Update help overlay with new keybindings
**Description:** As a user, I want the help overlay (`?`) to document all available key alternatives so I have a complete reference.

**Acceptance Criteria:**
- [ ] Help overlay Playback section lists arrow keys alongside H/L (e.g., "H / Left" for step backward)
- [ ] Help overlay Navigation section lists arrow keys alongside J/K (e.g., "J / Up" for previous item)
- [ ] Help overlay Playback section lists `,`/`.` alongside `<`/`>` for step size
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

### US-069: Fix three-column layout alignment
**Description:** As a user, I want the three columns to always render inline (side-by-side at equal height) so the layout doesn't break or misalign.

**Acceptance Criteria:**
- [ ] All three columns render at the same height regardless of content length
- [ ] Vertical border characters (`│`) align correctly on every row
- [ ] No blank gaps or overflow between columns
- [ ] Layout is correct at terminal widths of 80, 120, and 160+ columns
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

### US-070: Responsive column widths — column 3 shrinks first
**Description:** As a user, I want the live stats panel (column 3) to shrink before columns 1 and 2 when the terminal is narrow, so that the most important content (controls and notes list) stays usable.

**Acceptance Criteria:**
- [ ] At wide terminals (120+ cols): columns split roughly equal thirds
- [ ] At medium terminals (80-119 cols): column 3 shrinks to a minimum width (e.g., 15-20 chars) before columns 1 and 2 start shrinking
- [ ] Column 3 hides entirely if terminal width is too narrow to fit it (below ~90 cols), leaving a two-column layout
- [ ] Columns 1 and 2 share remaining space equally when column 3 is hidden
- [ ] Layout recalculates on `WindowSizeMsg`
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

## Functional Requirements

- FR-1: In the normal-mode key handler (`Update()`), add cases for `"left"`, `"right"`, `"up"`, `"down"` that map to seek backward, seek forward, select previous, and select next respectively
- FR-2: In the normal-mode key handler, add cases for `","` and `"."` that map to `decreaseStepSize()` and `increaseStepSize()` respectively
- FR-3: Arrow keys must NOT interfere with command mode input — `left`/`right` are already handled in `handleCommandInput()` for cursor movement; `up`/`down` should be ignored in command mode
- FR-4: Update `GetControlGroups()` in `controls.go` to show combined shortcut text (e.g., `"H/←"`, `"L/→"`, `"J/↑"`, `"K/↓"`, `",/<"`, `"./>"`  or similar concise format)
- FR-5: Refactor `renderColumn1()` to use `GetControlGroups()` instead of its own hardcoded controls slice
- FR-6: Update help overlay keybinding definitions to include the alternative keys
- FR-7: Replace the equal-thirds column width calculation in `View()` with a responsive algorithm: allocate a minimum width to column 3, then distribute remaining space to columns 1 and 2
- FR-8: When terminal width minus borders is less than a threshold (e.g., 88 cols), hide column 3 entirely and split between columns 1 and 2

## Non-Goals

- No changes to the actual mpv playback functionality or IPC protocol
- No new keybindings beyond arrow keys and comma/period
- No changes to forms, database, or command mode behavior
- No changes to the stats view or its keybindings
- No refactoring of the timeline or status bar components

## Technical Considerations

- The key handler switch in `Update()` uses `msg.String()` which returns `"left"`, `"right"`, `"up"`, `"down"` for arrow keys, and `","`, `"."` for punctuation
- Command mode already captures `"left"` and `"right"` for cursor movement — the normal-mode arrow key cases will only trigger when `commandInput.Active` is false (command mode is checked first)
- Column width calculation happens in `View()` around lines 1438-1443 — this is the main location for responsive sizing changes
- The `padToWidth()` helper uses `lipgloss.Width()` which correctly handles ANSI escape sequences
- `renderColumn1()` currently has its own hardcoded controls list (lines 1528-1543) that duplicates `GetControlGroups()` — consolidating removes this duplication

## Success Metrics

- All six vim-style keys (H/J/K/L/</>) and their alternatives (arrows/,/.) perform identical actions
- Three-column layout stays aligned at all terminal widths from 80 to 200+ columns
- Column 3 gracefully disappears at narrow widths without breaking layout
- Controls display and help overlay are accurate and complete

## Open Questions

- Should arrow keys also work in the stats view for J/K navigation? (Currently stats view only uses J/K)
- What exact column width threshold should trigger hiding column 3? (Suggested: ~88 cols total)
