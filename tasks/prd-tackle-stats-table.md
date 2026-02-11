# PRD: Tackle Stats Table

## Introduction

The current "Top Players" section in Column 3's live stats panel shows a simple ranked list of the top 5 players by total tackles. This is limiting — it caps at 5 players, only sorts by total, and doesn't let the user explore the data. This feature replaces that section with a full table inside a bordered container showing all players' tackle stats with sortable columns and a totals row.

## Goals

- Show all players' tackle stats in a table format (not just top 5)
- Display core columns: Player, Total, Completed, Missed, Percentage
- Include a totals row pinned below the header for at-a-glance aggregates
- Allow the user to change sort order via keybindings
- Wrap the table in a RenderInfoBox container with "Tackle Stats" header for visual consistency

## User Stories

### US-102: Replace Top Players with Tackle Stats table
**Description:** As a user, I want to see all players' tackle stats in a table so I can compare performance across the full squad, not just the top 5.

**Acceptance Criteria:**
- [ ] The "Top Players" section in `StatsPanel()` is replaced with a table rendered inside a `RenderInfoBox` container with the header "Tackle Stats"
- [ ] Table has a header row with columns: Player, Tot, Comp, Miss, Pct (or similar short labels that fit the column 3 width)
- [ ] A totals row appears directly below the header row, summing Total, Completed, Missed across all players, and showing the overall percentage
- [ ] The totals row is visually distinct from player rows (e.g. bold, different colour, or a separator line above/below)
- [ ] All players are shown, truncated if they exceed available height (no scrolling)
- [ ] Player names are truncated with ellipsis if they exceed the available name column width
- [ ] The table renders correctly at column 3 widths from 18 to 40+ characters
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

### US-103: Add sort keybindings to Tackle Stats table
**Description:** As a user, I want to change the sort order of the tackle stats table so I can rank players by different metrics.

**Acceptance Criteria:**
- [ ] Keybindings are active in normal mode (not during forms, command input, or stats view) to cycle the sort column
- [ ] A key (e.g. `s` or another unused key) cycles through sort columns: Total (default), Completed, Missed, Percentage, Player (alphabetical)
- [ ] The current sort column is indicated in the table header (e.g. the active column header is highlighted or has an arrow indicator)
- [ ] Sort order is descending by default for numeric columns, ascending for Player name
- [ ] The totals row always stays pinned below the header regardless of sort order
- [ ] The sort state is stored on the Model and persists while the app is running
- [ ] The sort keybinding is added to the help overlay and controls display
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

## Functional Requirements

- FR-1: Replace the "Top Players" section in `StatsPanel()` with a table wrapped in `RenderInfoBox("Tackle Stats", contentLines, width)`
- FR-2: Table header row shows column labels: Player, Tot, Comp, Miss, Pct
- FR-3: Totals row is rendered directly below the header, showing aggregate values across all players
- FR-4: Player rows show: truncated name, total count, completed count, missed count, percentage (formatted as integer with % suffix)
- FR-5: Percentage is calculated as `Completed / (Completed + Missed) * 100`, displayed as "-" when both are zero
- FR-6: All players from `tackleStats` slice are rendered; rows beyond available height are simply not shown (no scroll mechanism)
- FR-7: A sort key in normal mode cycles through sort columns: Total, Completed, Missed, Percentage, Player
- FR-8: Active sort column is visually indicated in the header row
- FR-9: The totals row is exempt from sorting — always pinned at position 1 (below header)
- FR-10: The Summary and Event Distribution sections above the table remain unchanged

## Non-Goals

- No scrolling within the table — if there are more players than available height, extra rows are truncated
- No per-player drill-down or selection from this table (the full stats view already handles that)
- No column resizing or reordering
- No ascending/descending toggle — fixed direction per column type

## Design Considerations

- Reuse `RenderInfoBox` for the bordered container (same pattern as Selected Tag, Controls, etc.)
- Column widths must be calculated dynamically based on available width — Player name gets the remaining space after fixed-width numeric columns
- Numeric columns should be right-aligned for readability
- The totals row could use a different foreground colour (e.g. `styles.Cyan` for totals vs `styles.LightLavender` for player rows) or bold styling
- Consider a thin separator (e.g. a dashed line `─`) between the totals row and the first player row

## Technical Considerations

- `StatsPanel()` already receives `tackleStats []PlayerStats` and `width int` — no new data plumbing needed
- The `PlayerStats` struct already has all needed fields: Player, Total, Completed, Missed, Percentage
- Sort state needs a new field on the Model (e.g. `tackleSortColumn int` or an enum) — pass it to `StatsPanel()` or sort before calling
- `loadTackleStatsForPanel()` is called on each tick — sorting should happen at render time, not at load time, so the source data stays stable
- The sort keybinding must be excluded when forms, command input, stats view, or help are active (same guard pattern as other normal-mode keys)

## Success Metrics

- All players visible at a glance (not capped at 5)
- User can quickly identify top performer by any metric via sort cycling
- Totals row provides instant team-level summary
- Table fits cleanly within Column 3 at all supported widths (18-40+ chars)

## Open Questions

- Which key should be used for sort cycling? `s` is a natural choice but need to verify it's not already bound.
