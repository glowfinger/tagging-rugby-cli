# PRD: Dynamic Notes List Height

## Introduction

The notes list component (`NotesList`) currently renders a fixed 10-row table regardless of available vertical space. This wastes screen real estate on tall terminals and clips content on short ones. This feature makes the notes list dynamically fill all available height passed by the parent layout, improving responsiveness when the terminal is resized vertically.

## Goals

- Remove the hardcoded `tableRows = 10` constant from `noteslist.go`
- Derive visible row count entirely from the `height` parameter already passed to `NotesList()`
- Ensure the notes list responds correctly when the terminal is resized vertically
- Update `TUI-ARCHITECTURE.md` NotesList section to reflect dynamic height behaviour

## User Stories

### US-001: Remove fixed row constant and use dynamic height
**Description:** As a user, I want the notes list to fill all available vertical space so I can see more notes on taller terminals without manual scrolling.

**Acceptance Criteria:**
- [ ] The `tableRows` constant is removed from `noteslist.go`
- [ ] `NotesList()` calculates visible rows from the `height` parameter (subtracting 1 for the header row)
- [ ] All references to `tableRows` inside `NotesList()`, `scrollToCurrentTime()`, and scroll-offset clamping logic use the computed row count instead
- [ ] No minimum height constraint — use whatever height is available (even 0 or negative values should not panic)
- [ ] `CGO_ENABLED=0 go vet ./...` passes
- [ ] `CGO_ENABLED=0 go build ./...` passes

### US-002: Update TUI-ARCHITECTURE.md
**Description:** As a developer, I want the architecture doc to accurately describe the NotesList component so I can trust it as a reference.

**Acceptance Criteria:**
- [ ] The NotesList section in `tui/TUI-ARCHITECTURE.md` no longer mentions "fixed-row" or "10-row"
- [ ] The description states that row count is derived from the `height` parameter
- [ ] No other sections are modified

## Functional Requirements

- FR-1: Remove `const tableRows = 10` from `tui/components/noteslist.go`
- FR-2: Compute `visibleRows := height - 1` (1 line reserved for the header) at the top of `NotesList()`
- FR-3: Replace every occurrence of `tableRows` with `visibleRows` in `NotesList()` — the render loop, empty-state padding, and scroll-offset clamping
- FR-4: Pass `visibleRows` into `scrollToCurrentTime()` (replace the method's current use of `tableRows`)
- FR-5: Guard against non-positive `visibleRows` — if `visibleRows <= 0`, return only the header (or empty string)
- FR-6: Update the NotesList subsection of `tui/TUI-ARCHITECTURE.md` to say the component derives its row count from the `height` parameter

## Non-Goals

- No changes to column width calculation or responsive column hiding
- No changes to the `Container` component or layout system
- No changes to how `renderColumn2` computes or passes the `height` value
- No minimum height enforcement — the caller is responsible for providing a sensible height
- No changes to any TUI-ARCHITECTURE.md sections other than NotesList

## Technical Considerations

- The `height` parameter is already passed through from `renderColumn2` → `NotesList` — no new plumbing needed
- `scrollToCurrentTime` is a method on `*NotesListState` that currently references `tableRows` directly; it will need a `visibleRows int` parameter added
- The empty-state branch also pads to `tableRows` — update it to use `visibleRows`

## Success Metrics

- Notes list fills the full column height at any terminal size
- Vertical resize immediately adjusts the number of visible rows
- No layout misalignment (columns stay the same height as before)

## Open Questions

- None — scope is well-defined.
