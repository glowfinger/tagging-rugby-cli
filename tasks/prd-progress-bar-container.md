# PRD: Progress Bar in Bordered Container

## Introduction

The timeline progress bar currently renders as two bare lines (bar + position indicator) with a `DarkPurple` background spanning full terminal width. It sits between the column layout and the command input bar without any visual framing. This change wraps the progress bar in a bordered container with a tab-style header, matching the bordered style used throughout the rest of the UI. The container spans the full terminal width and shrinks responsively with the screen size.

## Goals

- Wrap the timeline progress bar in a bordered container with a tab-style header
- The container spans full terminal width and shrinks responsively with the screen
- Match the visual style of other bordered containers (control boxes, mini player, notes list)

## User Stories

### US-001: Wrap progress bar in bordered container
**Description:** As a user, I want the progress bar to be in a bordered container matching the rest of the UI so the layout looks consistent and polished.

**Acceptance Criteria:**
- [ ] The timeline renders inside a bordered box with a tab-style "Timeline" header (same `┌┤ Timeline ├┐` box-drawing style as other containers)
- [ ] The progress bar line (filled/unfilled bar with event markers + time display) renders inside the box
- [ ] The position indicator line (▲) renders inside the box below the bar
- [ ] The box spans the full terminal width (`m.width`)
- [ ] The bar width shrinks correctly when the terminal is resized — the bar adapts to the inner box width (total width minus 2 border characters minus 2 padding spaces)
- [ ] Event markers (◆) still display at correct positions on the bar
- [ ] The playhead position (╸ and ▲) still tracks correctly
- [ ] The time display (`H:MM:SS / H:MM:SS`) still shows to the right of the bar
- [ ] At narrow widths (< 20), the timeline still returns empty or a minimal render gracefully
- [ ] Border color uses `styles.Purple`, header text uses `styles.Pink`
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

## Functional Requirements

- FR-1: The timeline is rendered inside a bordered box with a "Timeline" tab header
- FR-2: The box spans full terminal width and adapts when the terminal is resized
- FR-3: The progress bar and position indicator render as content rows inside the box
- FR-4: All existing timeline functionality (fill position, event markers, playhead, time display) works unchanged inside the container
- FR-5: The `colHeight` calculation in `View()` accounts for the increased height of the bordered timeline (tab header adds 3 lines + bottom border adds 1 line = 4 extra lines compared to the current 2-line bare timeline)

## Non-Goals

- No changes to the progress bar logic (fill calculation, marker placement, colors)
- No interactivity on the timeline (no click-to-seek)
- No changes to the command input bar below the timeline

## Design Considerations

- Reuse the same box-drawing pattern as `RenderControlBox` / `RenderMiniPlayer` — tab header (3 lines), content rows, bottom border
- The container has exactly 2 content rows (bar line + indicator line), so total box height = 3 (tab header) + 2 (content) + 1 (bottom border) = 6 lines
- The current timeline is 2 lines; the bordered version will be 6 lines, so `colHeight` in `View()` must be adjusted (subtract 4 more from available height, or adjust the constant from `m.height - 5` to `m.height - 9`)

## Technical Considerations

- The `Timeline()` function in `tui/components/timeline.go` currently receives `width` and uses it directly for bar calculation. With borders, the inner content width is `width - 4` (2 border chars + 2 padding spaces). Options:
  - Add the border wrapping inside `Timeline()` itself — pass `width` as before, let it build the bar at `width - 4` and add borders around it
  - Or wrap externally in `View()` — call `Timeline()` with reduced width, then wrap the output in a box. This is simpler but duplicates box-rendering logic
  - Adding inside `Timeline()` is cleaner since it's self-contained
- The `bgStyle` with `DarkPurple` background currently applied to the bar lines can be removed since the bordered box provides the visual frame
- The `colHeight` adjustment: currently `m.height - 5` (1 status bar + 2 timeline + 1 command input + 1 gap). With bordered timeline at 6 lines: `m.height - 9` (1 status bar + 6 timeline + 1 command input + 1 gap)

## Success Metrics

- Timeline has a consistent bordered look matching the rest of the UI
- Progress bar shrinks correctly at any terminal width
- No loss of timeline functionality (markers, playhead, time display)

## Open Questions

- Should the tab header say "Timeline" or "Progress"?
