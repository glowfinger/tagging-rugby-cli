# PRD: Column 4 — Controls Panel

## Introduction

Add a fourth column to the TUI layout with a fixed width of 30 cells, positioned as the rightmost column. Move the three control groups (Playback, Navigation, Views) from Column 1 into this new Column 4. This declutters Column 1 (which retains only the mini player card and selected tag detail) and gives the keybinding reference its own dedicated space.

## Goals

- Add a fixed-width Column 4 (30 cells) as the rightmost column in the layout
- Move all three control groups (Playback, Navigation, Views) from Column 1 into Column 4
- Column 1 retains only the mini player card and selected tag detail box
- Column 4 appears only at terminal width >= 160; below that threshold it is hidden
- No changes to Column 2 (notes list) or Column 3 (stats panel) content
- Maintain the existing responsive layout breakpoints for 3-col, 2-col, and mini player modes

## User Stories

### US-001: Update ComputeColumnWidths for 4-column layout
**Description:** As a developer, I need the layout engine to compute widths for 4 columns so the new controls column can be rendered at the correct size.

**Acceptance Criteria:**
- [ ] `ComputeColumnWidths()` returns 4 column widths + a `showCol4` boolean
- [ ] Column 4 has a fixed width of 30 cells (new constant `Col4Width = 30`)
- [ ] `showCol4` is true when terminal width >= 160 (new constant `Col4ShowThreshold = 160`)
- [ ] When `showCol4` is true, account for 3 border separators in usable width (was 2 for 3-col)
- [ ] When `showCol4` is false, existing 3-col / 2-col / mini breakpoints are unchanged
- [ ] Remaining width (after Col1 + Col4 fixed widths + borders) is split evenly between Col2 and Col3
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-002: Add renderColumn4 method
**Description:** As a developer, I need a `renderColumn4` method that renders the three control groups inside a Container so it integrates with the column join pipeline.

**Acceptance Criteria:**
- [ ] New method `(m *Model) renderColumn4(width, height int) string` in `tui/columns.go`
- [ ] Renders the Playback, Navigation, and Views control groups using `components.GetControlGroups()` and `components.RenderControlBox()`
- [ ] Output is wrapped in `layout.Container{Width, Height}.Render(...)` for exact dimensions
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-003: Remove control groups from renderColumn1
**Description:** As a developer, I need to remove the control groups rendering from Column 1 so they don't appear in two places.

**Acceptance Criteria:**
- [ ] `renderColumn1` no longer calls `GetControlGroups()` or `RenderControlBox()`
- [ ] Column 1 renders only: mini player card + selected tag detail box
- [ ] Column 1 still wraps output in `layout.Container{Width, Height}.Render(...)`
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-004: Wire Column 4 into View() rendering
**Description:** As a developer, I need to update the `View()` method to include Column 4 in the column join when the terminal is wide enough.

**Acceptance Criteria:**
- [ ] When `showCol4` is true, `View()` renders 4 columns: `[col1, col2, col3, col4]`
- [ ] When `showCol4` is false, existing 3-col or 2-col rendering is unchanged
- [ ] Column 4 is the rightmost column (appears after Column 3)
- [ ] `JoinColumns()` correctly joins 4 columns with 3 border separators
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-005: Update TUI-ARCHITECTURE.md
**Description:** As a developer, I need the architecture docs to reflect the new 4-column layout so future contributors understand the layout system.

**Acceptance Criteria:**
- [ ] Directory Structure: `columns.go` comment updated to `renderColumn1/2/3/4`
- [ ] Rendering Pipeline step 3b updated from "responsive 2-col or 3-col grid" to "responsive 2/3/4-col grid"
- [ ] Column Rendering table includes Column 4 row: method `renderColumn4(width, height)`, content "Keybinding control groups (Playback, Navigation, Views)"
- [ ] Column 1 row updated: content changed from "Playback status, selected tag detail, control groups" to "Playback status, selected tag detail"
- [ ] `ComputeColumnWidths` function signature updated to show new return values `(col1, col2, col3, col4 int, showCol3, showCol4 bool)`
- [ ] Responsive layout table includes the 4-column row: `>= 160 | 4-column | Col 1 = 30 (fixed), Col 4 = 30 (fixed), remaining split evenly between 2 and 3`
- [ ] Existing 3-column row threshold updated to `90 - 159`
- [ ] No stale references to controls being in Column 1

## Functional Requirements

- FR-1: Add constant `Col4Width = 30` and `Col4ShowThreshold = 160` to `tui/layout/columns.go`
- FR-2: `ComputeColumnWidths()` returns `(col1, col2, col3, col4 int, showCol3, showCol4 bool)`
- FR-3: When `showCol4` is true: usable = termWidth - 3 borders; col1 = 30, col4 = 30; remaining split evenly between col2 and col3
- FR-4: When `showCol4` is false and `showCol3` is true: existing 3-col logic (usable = termWidth - 2 borders)
- FR-5: When both are false: existing 2-col logic (usable = termWidth - 1 border)
- FR-6: `renderColumn4(width, height)` renders control groups wrapped in Container
- FR-7: `renderColumn1` renders only mini player + selected tag detail (no controls)
- FR-8: `View()` conditionally includes Column 4 in the columns slice and widths slice
- FR-9: `JoinColumns()` already supports N columns — no changes needed

## Non-Goals

- No changes to the control groups themselves (Playback, Navigation, Views stay as-is)
- No changes to Column 2 (notes list) or Column 3 (stats panel) content
- No changes to form overlays, help overlay, or stats view
- No changes to mini player mode (< 80 width)
- No reordering or regrouping of keybinding controls

## Technical Considerations

- `JoinColumns()` is already generic (takes `[]string` and `[]int`) — it handles any number of columns with no code changes
- `ComputeColumnWidths` return signature changes from 4 to 6 values — all call sites must be updated
- Column 1 and Column 4 are both fixed at 30 — at the 160 threshold that leaves 160 - 30 - 30 - 3 = 97 cells split between Col2 (48) and Col3 (49)
- The `components.GetControlGroups()` call moves from `renderColumn1` to `renderColumn4` — no new component code needed

## Success Metrics

- Controls appear in Column 4 at terminal width >= 160
- Column 1 is visually cleaner with only player + tag detail
- All existing responsive breakpoints (3-col at 90+, 2-col at 80+, mini at < 80) continue to work
- No layout misalignment — all columns maintain exact height

## Open Questions

- None — all requirements clarified.
