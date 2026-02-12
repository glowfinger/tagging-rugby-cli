# PRD: Remove Status Bar (Top Row)

## Introduction

The full-width status bar at the top of the TUI duplicates information already shown in the mini player card in column 1 (play/pause, time, step size, speed, mute). Removing it declutters the UI and gives 1 extra line of vertical space to the column layout. The overlay indicator (currently only in the status bar) will be moved to the mini player card.

## Goals

- Remove the full-width status bar from the top of the main TUI layout
- Add the overlay indicator to the mini player card in column 1 so no information is lost
- Reclaim 1 line of vertical space for the column layout
- Replace the status bar with the mini player card in form/error overlay views
- Keep the `StatusBar` function and `StatusBarState` struct intact (still used as a data container)

## User Stories

### US-001: Add overlay indicator to mini player card
**Description:** As a user, I want to see the overlay status in the playback card so that information isn't lost when the status bar is removed.

**Acceptance Criteria:**
- [ ] Mini player card shows "Overlay: on" when `state.OverlayEnabled` is true
- [ ] Mini player card shows "Overlay: off" when `state.OverlayEnabled` is false
- [ ] The overlay line appears after the mute line (or after time if not muted), before the warning line
- [ ] Overlay indicator always shows (both on and off states), unlike the mute indicator which is conditional
- [ ] Both the fixed-width (column 1) and auto-sized (standalone narrow terminal) mini player variants show the overlay indicator
- [ ] `CGO_ENABLED=0 go vet ./...` passes
- [ ] `CGO_ENABLED=0 go build ./...` passes

### US-002: Remove status bar from main layout and form/error views
**Description:** As a user, I want the status bar removed from the top of the screen so the UI is cleaner and I get more vertical space for content.

**Acceptance Criteria:**
- [ ] The `statusBar` local variable and `components.StatusBar()` call are removed from `View()` in `tui/tui.go`
- [ ] Main layout assembly changes from `statusBar + "\n" + columnsView + "\n" + timeline + "\n" + commandInput` to `columnsView + "\n" + timeline + "\n" + commandInput`
- [ ] Form overlay views (confirm, note form, tackle form) show `components.RenderMiniPlayer(m.statusBar, 0, false)` at the top instead of the status bar
- [ ] Error view shows `components.RenderMiniPlayer(m.statusBar, 0, true)` at the top instead of the status bar
- [ ] `colHeight` calculation changes from `m.height - 9` to `m.height - 8` to reclaim the freed line
- [ ] The `StatusBar` function in `tui/components/statusbar.go` is NOT deleted — it remains available but unused by the main View
- [ ] The `StatusBarState` struct remains unchanged — it is still used as the data container for playback state
- [ ] Help overlay and stats view remain unchanged (they already don't show the status bar)
- [ ] `CGO_ENABLED=0 go vet ./...` passes
- [ ] `CGO_ENABLED=0 go build ./...` passes

## Functional Requirements

- FR-1: Add an overlay indicator line ("Overlay: on" / "Overlay: off") to `RenderMiniPlayer` in `tui/components/miniplayer.go`
- FR-2: Remove the `components.StatusBar()` call and `statusBar` variable from `View()` in `tui/tui.go`
- FR-3: Update the main layout assembly to omit the status bar line
- FR-4: Update form overlay views (lines ~1522, ~1529, ~1536) to use `RenderMiniPlayer` instead of `statusBar`
- FR-5: Update error view (line ~1497) to use `RenderMiniPlayer` instead of `StatusBar`
- FR-6: Change `colHeight` from `m.height - 9` to `m.height - 8`

## Non-Goals

- Do not delete the `StatusBar` function or `StatusBarState` struct
- Do not change the mini player's bordered card visual style
- Do not change the standalone mini player mode (narrow terminals < 80 cols) — it already works without the status bar
- Do not change help overlay or stats view

## Technical Considerations

- Files to modify: `tui/components/miniplayer.go` (add overlay indicator), `tui/tui.go` (remove status bar from View, update colHeight)
- `StatusBarState` is used throughout the codebase as the playback data container — it must stay
- The `StatusBar` function can remain as dead code for now; it may be useful later or can be cleaned up separately
- Form overlay views currently show `statusBar + "\n" + controlsDisplay + "\n" + formView` — these change to `miniPlayer + "\n" + controlsDisplay + "\n" + formView`

## Success Metrics

- No information is lost — all status bar data visible in the mini player card
- Column area gains 1 extra line of vertical space
- Form and error views still show playback state via the mini player card
- No layout misalignment at any terminal width

## Open Questions

- None
