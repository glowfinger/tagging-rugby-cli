# PRD: Replace Top Players with Tackle Stats Table

## Introduction

Replace the "Top Players" leaderboard section in Column 3's StatsPanel with a full tackle stats table showing all players. The current leaderboard only shows the top 5 players with limited info (name, total, %). The new table displays every player with columns for Player, Total, Completed, Missed, and %, plus a pinned totals row at the top. The Summary and Event Distribution sections remain unchanged.

## Goals

- Show all players' tackle stats at a glance in Column 3 (no arbitrary top-5 limit)
- Include a totals row pinned at the top of the table for quick aggregate view
- Use all available vertical space in the column for player rows
- Maintain the existing Summary and Event Distribution sections above the table

## User Stories

### US-001: Replace Top Players with Tackle Stats Table
**Description:** As a user, I want to see a full tackle stats table in Column 3 so that I can see every player's tackle data without opening the full-screen StatsView.

**Acceptance Criteria:**
- [ ] The "Top Players" section in `StatsPanel` is replaced with a "Tackle Stats" table
- [ ] Table columns: Player, Total, Comp, Miss, %
- [ ] All players from `tackleStats` are displayed (not limited to top 5)
- [ ] Players are sorted by total tackles descending
- [ ] Table uses all remaining vertical space in the column (after Summary and Event Distribution)
- [ ] Empty state shows "No tackle data" when `tackleStats` is empty
- [ ] `CGO_ENABLED=0 go vet ./...` passes

### US-002: Add Pinned Totals Row
**Description:** As a user, I want a totals row pinned at the top of the tackle stats table so I can see aggregate stats regardless of scroll position.

**Acceptance Criteria:**
- [ ] A "TOTAL" row appears as the first row after the table header
- [ ] The totals row sums Total, Completed, and Missed across all players
- [ ] The % column in the totals row shows the overall completion percentage (total completed / (total completed + total missed) * 100)
- [ ] The totals row is visually distinct from player rows (different colour or bold styling)
- [ ] The totals row remains visible even if the player list is long enough to trigger the Container's "↓ More..." truncation
- [ ] `CGO_ENABLED=0 go vet ./...` passes

### US-003: Update TUI-ARCHITECTURE.md
**Description:** As a developer, I want the architecture doc to reflect the new tackle stats table so the documentation stays accurate.

**Acceptance Criteria:**
- [ ] The StatsPanel description in TUI-ARCHITECTURE.md is updated from "bar graph, category counts, leaderboard" to "bar graph, category counts, tackle stats table"
- [ ] The Column 3 description in the column rendering table is updated accordingly
- [ ] No other sections are changed

## Functional Requirements

- FR-1: Remove the "Top Players" leaderboard section from `StatsPanel()` in `statspanel.go`
- FR-2: Add a "Tackle Stats" table section with a header row: `Player  Total  Comp  Miss    %`
- FR-3: Add a totals row immediately after the header, summing all players' Total, Completed, Missed, and computing overall % as `sum(Completed) / (sum(Completed) + sum(Missed)) * 100`
- FR-4: Render all players from `tackleStats`, sorted by total tackles descending (alphabetical name as tiebreaker)
- FR-5: Player name column should truncate long names to fit the available width
- FR-6: The % column shows `-` when a player has zero completed + missed tackles
- FR-7: The table fills all remaining vertical space — no hardcoded row limit
- FR-8: The table is read-only (no scroll, select, or sort interaction) — interaction is a future feature

## Non-Goals

- No interactive scrolling, row selection, or sort toggling in Column 3 (future feature)
- No changes to the full-screen StatsView overlay
- No changes to Summary or Event Distribution sections
- No changes to how `tackleStats` data is computed or passed to `StatsPanel`

## Technical Considerations

- The `StatsPanel` function already receives `tackleStats []PlayerStats` and `height int` — use `height` to determine available rows for the table after rendering Summary and Event Distribution
- Reuse existing `PlayerStats` struct from `statsview.go` (has Total, Completed, Missed, Percentage fields)
- The totals row must be rendered before the player rows so it stays visible when Container truncates overflow with "↓ More..."
- Column widths should be computed relative to the `width` parameter to handle responsive column sizing
- Use existing styles: `styles.Pink` for headers, `styles.LightLavender` for player rows, `styles.Cyan` for the totals row, `styles.Lavender` for dim text

## Success Metrics

- All players visible in Column 3 without needing to open StatsView
- Totals row always visible at top of table
- No layout breakage at any terminal width

## Open Questions

- None — this is a straightforward replacement of one section within an existing component.
