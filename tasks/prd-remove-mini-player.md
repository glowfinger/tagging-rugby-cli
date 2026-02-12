# PRD: Remove Mini Player Mode & Revise Responsive Layout

## Introduction

Remove the standalone mini player fallback that appears when the terminal is narrow (< 80 cells). Instead, always show the default multi-column layout with revised responsive breakpoints. The typical terminal width is > 200 cells, so the layout should be optimized for wide terminals while gracefully degrading by truncating and hiding columns as width decreases. Column 1 must always remain visible.

## Goals

- Eliminate the mini player narrow-mode entirely (the standalone centered card shown at < 80 width)
- Remove the key-filtering that suppresses non-playback keys in narrow mode
- Revise responsive column breakpoints and sizing for the new layout rules
- Keep Column 1 always visible regardless of terminal width

## User Stories

### US-001: Remove mini player narrow-mode fallback
**Description:** As a user, I want the full column layout to always display so that I never get stuck in a limited mini player view.

**Acceptance Criteria:**
- [ ] The `RenderMiniPlayer` standalone view (narrow terminal fallback at < 80 width) is no longer triggered in `View()`
- [ ] The key-filtering block in `Update()` that suppresses non-playback keys when `width < MinTerminalWidth` is removed
- [ ] All keybindings work at any terminal width
- [ ] The `MinTerminalWidth` constant is removed or repurposed
- [ ] `RenderMiniPlayer` function in `tui/components/controls.go` can be removed if no longer called anywhere (check form overlays too)
- [ ] Typecheck and build pass (`CGO_ENABLED=0 go vet ./...`)

### US-002: Revise responsive column breakpoints and sizing
**Description:** As a user, I want columns to have proper fixed widths and hide progressively as the terminal narrows, so the layout stays readable at any size.

**Acceptance Criteria:**
- [ ] Column 1: fixed at 30 cells, **always visible**
- [ ] Column 2: target width 80 cells; truncates as terminal narrows; hides when it would fall below 30 cells
- [ ] Column 3: target width 60 cells; truncates as terminal narrows; hides when it would fall below 30 cells (hides before Column 2)
- [ ] Column 4: fixed at 30 cells; hidden when terminal width < 170
- [ ] Update `ComputeColumnWidths()` in `tui/layout/columns.go` with new constants and logic
- [ ] Update `View()` in `tui/tui.go` to use the new column visibility flags
- [ ] Layout renders correctly at widths: 220, 200, 170, 140, 100, 60
- [ ] Typecheck and build pass (`CGO_ENABLED=0 go vet ./...`)

### US-003: Clean up form overlay views that reference mini player
**Description:** As a developer, I want to remove mini player references from the form overlay code paths (note form, tackle form, confirm discard) and truncate the forms to fit the available terminal width.

**Acceptance Criteria:**
- [ ] The `confirmDiscardForm`, `noteForm`, and `tackleForm` overlay views in `View()` no longer call `RenderMiniPlayer`
- [ ] Form overlays truncate to fit the current terminal width (no wrapping or overflow)
- [ ] Error view (`m.err != nil`) no longer calls `RenderMiniPlayer`
- [ ] Typecheck and build pass (`CGO_ENABLED=0 go vet ./...`)

## Functional Requirements

- FR-1: Remove the `if m.width < layout.MinTerminalWidth` guard in `View()` that returns `RenderMiniPlayer`
- FR-2: Remove the key-filtering `switch` block in `Update()` that suppresses keys when `width < MinTerminalWidth`
- FR-3: Update layout constants in `tui/layout/columns.go`:
  - `Col1Width = 30` (unchanged, always visible)
  - `Col2TargetWidth = 80` (new — target width for Column 2)
  - `Col3TargetWidth = 60` (new — target width for Column 3)
  - `Col4Width = 30` (unchanged)
  - `Col4ShowThreshold = 170` (changed from 160)
  - `ColMinWidth = 30` (new — minimum width before a column hides)
- FR-4: Revise `ComputeColumnWidths()` so that:
  - Column 4 shows when width >= 170
  - Column 3 gets up to 60 cells, truncates as width shrinks, hides when it would fall below 30 cells
  - Column 2 gets up to 80 cells, truncates as width shrinks, hides when it would fall below 30 cells
  - Column 1 (30 cells) is always shown
- FR-5: Remove or refactor `RenderMiniPlayer` in `tui/components/controls.go` — delete if unused, or keep only the compact card used inside Column 1 (if that's a separate render path)
- FR-6: Update form overlay views to not depend on `RenderMiniPlayer` — truncate forms to fit available terminal width

## Non-Goals

- No changes to the Column 1 mini player card that renders *inside* Column 1 of the default layout — that stays as-is
- No changes to the help overlay, stats view, or timeline
- No changes to Column 1 content (selected item details, etc.)
- No new UI components or layout modes

## Technical Considerations

- The `RenderMiniPlayer` function is used in two contexts: (1) standalone narrow-mode fallback, and (2) form overlay headers. Both need updating, but the compact player card rendered inside `renderColumn1()` is a separate code path and should be preserved
- The `ControlsDisplay` function used in form overlays may also need review if it depends on mini player state
- Column width computation must handle edge cases where the terminal is extremely narrow — Column 1 should still render even if it's the only column

## Success Metrics

- The mini player standalone view never appears regardless of terminal width
- All keybindings work at any terminal width
- Layout is correct and readable at the typical width (> 200 cells)
- Columns hide gracefully: Col4 disappears below 170, Col3 truncates then hides at < 30 cells, Col2 truncates then hides at < 30 cells, Col1 always visible
- Form overlays truncate cleanly at any terminal width
