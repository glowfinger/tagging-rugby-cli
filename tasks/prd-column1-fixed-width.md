# PRD: Fix Column 1 Width to 30 Characters

## Introduction

Column 1 (playback status, selected tag detail, controls) currently shares width equally with the other columns, meaning it gets wider than necessary on wide terminals and can be too narrow on medium ones. This change fixes column 1 to a constant 30 characters wide, giving all remaining space to columns 2 and 3 which benefit more from extra width (notes list and stats panel).

## Goals

- Fix column 1 to exactly 30 characters wide in all layout modes
- Give column 2 (and column 3 when visible) the remaining terminal width
- Keep the existing responsive behavior for column 3 (hide below 90 cols, shrink before column 2)

## User Stories

### US-001: Fix column 1 to 30 characters wide
**Description:** As a user, I want column 1 to be a fixed compact width so the notes list and stats panel get maximum space.

**Acceptance Criteria:**
- [ ] Add a constant `col1Width = 30` in the View() layout section of `tui/tui.go`
- [ ] Three-column layout (width >= 90): col1 = 30, col3 gets its share, col2 gets the remainder (`usableWidth - col1Width - col3Width`)
- [ ] Two-column layout (width < 90): col1 = 30, col2 gets the remainder (`usableWidth - col1Width`)
- [ ] Wide terminals (>= 120): col1 = 30, col2 and col3 split the remaining space (col3 can grow larger than `col3MinWidth`)
- [ ] Medium terminals (90-119): col1 = 30, col3 = `col3MinWidth` (18), col2 gets the rest
- [ ] Narrow terminals (80-89): col1 = 30, col2 gets the rest (col3 hidden)
- [ ] The bordered control boxes from `RenderControlBox()` render correctly at width 30
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

## Functional Requirements

- FR-1: Column 1 is always exactly 30 characters wide regardless of terminal width
- FR-2: Column 2 expands to fill remaining space after column 1 and column 3 are allocated
- FR-3: Column 3 responsive behavior is unchanged: hidden below 90 cols, minimum 18 chars, grows at wide terminals
- FR-4: The `minTerminalWidth` (80) still applies — mini player activates below 80 cols

## Non-Goals

- No changes to column 1 content or rendering
- No changes to column 3 hide/show thresholds
- No changes to mini player behavior
- No user-configurable column width

## Technical Considerations

- The change is localized to the column width calculation in `View()` (around lines 1667-1718 in `tui/tui.go`)
- Replace the dynamic `col1Width` calculations with the constant `30`
- The bordered control boxes (Playback, Navigation, Views) are already designed for narrow widths — 30 chars gives 28 inner chars which is enough for `Name [ Shortcut ]` formatting
- At `minTerminalWidth` (80), two-column layout gives col2 = 80 - 1 - 30 = 49 chars, which is comfortable for the notes list

## Success Metrics

- Column 1 renders at exactly 30 characters at any terminal width >= 80
- Notes list (column 2) gets more space on wide terminals than before
- No layout overflow or truncation issues

## Open Questions

- None — this is a straightforward constant change.
