# PRD: Remove Gaps Between Adjacent Section Boxes

## Introduction

The TUI currently renders blank lines between adjacent bordered boxes in Column 3 (Event Distribution / Tackle Stats) and Column 4 (Playback / Navigation / Views). These gaps waste vertical space and break the visual continuity of the panel. This PRD covers removing those gaps so boxes stack directly against each other.

## Goals

- Remove the blank line between Event Distribution and Tackle Stats in Column 3.
- Remove the blank lines between Playback, Navigation, and Views boxes in Column 4.
- No other layout or styling changes.

## User Stories

### US-001: Remove gap between Event Distribution and Tackle Stats
**Description:** As a user, I want the Event Distribution and Tackle Stats boxes to be adjacent with no blank line between them so that the stats panel uses vertical space efficiently.

**Acceptance Criteria:**
- [ ] `StatsPanel` in `tui/components/statspanel.go` joins `eventBox` and `tackleBox` with a single `"\n"` instead of `"\n\n"`.
- [ ] No blank row appears between the two boxes when running the TUI.
- [ ] `CGO_ENABLED=0 go build ./...` passes without errors.
- [ ] `CGO_ENABLED=0 go vet ./...` passes without errors.

### US-002: Remove gaps between Playback, Navigation, and Views boxes
**Description:** As a user, I want the Playback, Navigation, and Views control boxes in Column 4 to be adjacent with no blank lines between them so that the controls panel uses vertical space efficiently.

**Acceptance Criteria:**
- [ ] `renderColumn4` in `tui/columns.go` no longer appends a blank `""` line between group boxes.
- [ ] The comment `// 1 blank line gap between bordered containers` is removed along with the associated `if` block.
- [ ] No blank rows appear between Playback/Navigation/Views boxes when running the TUI.
- [ ] `CGO_ENABLED=0 go build ./...` passes without errors.
- [ ] `CGO_ENABLED=0 go vet ./...` passes without errors.

## Functional Requirements

- FR-1: `tui/components/statspanel.go:169` — change `"\n\n"` to `"\n"` in the return statement that joins `eventBox` and `tackleBox`.
- FR-2: `tui/columns.go:174-177` — remove the blank-line gap logic between control group boxes in `renderColumn4`:
  ```go
  // Remove this block:
  if i < len(groups)-1 {
      lines = append(lines, "")
  }
  ```

## Non-Goals

- No changes to box borders, padding, or inner content.
- No changes to Column 1, Column 2, or the help overlay.
- No changes to responsive layout logic.

## Technical Considerations

- Both changes are one-liners / small deletions with no side effects.
- The `Container` component in `renderColumn3` and `renderColumn4` already handles height clamping, so removing gaps will simply give the containers more content lines to display — no overflow risk.
- Use `CGO_ENABLED=0` for all Go build and vet commands (required for `modernc.org/sqlite`).

## Success Metrics

- Boxes in Column 3 and Column 4 are visually flush (no blank rows between them).
- No build or vet errors introduced.

## Open Questions

- None. The change locations are unambiguous from the code.
