# PRD: Notes List in Bordered Container and Remove Column Separator

## Introduction

Two related layout changes to clean up the column layout:

1. **Notes list container**: Column 2 currently renders the notes list as a bare table with no visual boundary. Wrapping it in a bordered container with a tab-style "Notes" header matches the bordered style used in column 1 (mini player, selected tag, control boxes).

2. **Remove column separator**: The `│` vertical separator character drawn between columns is no longer needed once each column's content is in its own bordered container — the boxes themselves provide clear visual separation. Removing the separator also reclaims 1-2 characters of horizontal space.

## Goals

- Wrap the notes list in a bordered container with a "Notes" tab header filling column 2
- Remove the `│` vertical separator lines between columns
- Reclaim the 1-2 characters previously used by separators for column content

## User Stories

### US-001: Wrap notes list in bordered container
**Description:** As a user, I want the notes list in column 2 to have a bordered container matching the rest of the UI so all sections look visually consistent.

**Acceptance Criteria:**
- [ ] The notes list in `renderColumn2()` is wrapped in a bordered box with a tab-style "Notes" header (same `┌┤ Notes ├┐` style as control boxes)
- [ ] The table header row (ID, Time, Category, Text) renders inside the box below the tab header
- [ ] All data rows render inside the box
- [ ] The box fills the full column 2 width and height
- [ ] The box bottom border sits at the bottom of the column height
- [ ] Selected row highlighting still works correctly inside the box
- [ ] Empty state ("No notes or tackles for this video") still renders correctly inside the box
- [ ] Scrolling behavior is unchanged — rows scroll within the box
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-002: Remove column separator lines
**Description:** As a developer, I need to remove the `│` vertical separators between columns since bordered containers now provide visual separation.

**Acceptance Criteria:**
- [ ] The `borderStr` variable (purple `│`) is removed from the `View()` layout assembly
- [ ] Three-column row assembly changes from `c1+borderStr+c2+borderStr+c3` to `c1+c2+c3`
- [ ] Two-column row assembly changes from `c1+borderStr+c2` to `c1+c2`
- [ ] `usableWidth` calculation no longer subtracts border characters (was `m.width - 2` for 3-col, `m.width - 1` for 2-col — now just `m.width`)
- [ ] The reclaimed 1-2 characters are distributed to column widths
- [ ] No visual gap or overlap between columns — columns fill the full terminal width
- [ ] Typecheck/vet passes

## Functional Requirements

- FR-1: Column 2 renders its content inside a bordered box with a "Notes" tab header
- FR-2: The bordered box fills the column width and height, with the table content inside
- FR-3: The `│` vertical separators between columns are removed
- FR-4: Column width calculations use the full terminal width (no border character subtraction)
- FR-5: All existing notes list functionality (selection, scrolling, auto-scroll, highlighting) works inside the container

## Non-Goals

- No changes to column 3 (stats panel) — it does not get a container in this story
- No changes to notes list data, columns, or formatting
- No changes to the selection or scrolling logic
- No changes to column 1 content

## Design Considerations

- Reuse the existing box-drawing pattern from `RenderControlBox` / `RenderMiniPlayer` or the generic `RenderInfoBox` helper (if extracted by the selected tag container PRD)
- The notes list box is unique in that it has a fixed height (fills the column) with scrolling content — the box border should frame the full column height, not just the content rows
- The table header row inside the box acts as a secondary header below the tab — keep the underlined styling to distinguish it from data rows

## Technical Considerations

- `renderColumn2()` currently returns the raw output of `components.NotesList()`. The bordered container can either:
  - Be added inside `renderColumn2()` by wrapping the NotesList output in a box
  - Or be added inside `components.NotesList()` itself so it's self-contained
  - Wrapping in `renderColumn2()` is simpler and keeps `NotesList` reusable without forced borders
- The box height must match `colHeight` exactly so `normalizeLines()` doesn't truncate or pad incorrectly. The box itself should produce exactly `colHeight` lines (tab header 3 lines + content rows + bottom border)
- The NotesList table height needs to shrink to account for box chrome: `listHeight = colHeight - 4` (3 lines for tab header + 1 line for bottom border)
- Removing the `borderStr` from row assembly is straightforward — just concatenate columns directly. But ensure `padToWidth` still sizes each column correctly so they tile without gaps

## Success Metrics

- Column 2 has a consistent bordered look matching column 1
- No `│` separators visible between columns
- Notes list scrolling and selection work unchanged inside the container

## Open Questions

- Should column 3 (stats panel) also get a bordered container for consistency, or is that a separate task?
