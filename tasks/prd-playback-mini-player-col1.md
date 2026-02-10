# PRD: Replace Playback Indicator with Mini Player in Column 1

## Introduction

Column 1 currently has a plain-text "Playback" status card showing play/pause state, time, step size, mute, and overlay status as unstyled lines. The mini player component (`RenderMiniPlayer`) already renders a bordered card with the same information in a polished box-drawing style matching `wireframe/mini-player.txt`. This change replaces the plain playback card in column 1 with the mini player component, giving a consistent bordered visual style throughout the UI.

## Goals

- Replace the plain-text playback status card in `renderColumn1()` with the bordered mini player card
- Reuse the existing `RenderMiniPlayer()` component (or a column-width variant of it)
- Maintain all the same playback information currently shown

## User Stories

### US-001: Replace playback status card with mini player in column 1
**Description:** As a user, I want the playback status section in column 1 to use the same bordered card style as the mini player so the UI looks consistent and polished.

**Acceptance Criteria:**
- [ ] The plain-text "Playback" section at the top of `renderColumn1()` (play state, time, step, mute, overlay lines) is replaced with `RenderMiniPlayer()` or a similar bordered card render
- [ ] The bordered card fills the column width (no centering — left-aligned to fill col1)
- [ ] The card shows: play/pause icon, step size, time position / duration, mute icon when muted
- [ ] The overlay status indicator is included (currently shown in the plain card but not in `RenderMiniPlayer` — add it if missing)
- [ ] The "Selected Tag" detail card below the playback card remains unchanged
- [ ] The control boxes below the detail card remain unchanged
- [ ] The card renders correctly at column 1 width (30 chars after the fixed-width change)
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

## Functional Requirements

- FR-1: The playback section in column 1 renders as a bordered card with tab-style "Playback" header, matching the `wireframe/mini-player.txt` style
- FR-2: The card width matches the column 1 width (not centered, fills available space)
- FR-3: All current playback info is preserved: play/pause, time, step size, mute, overlay
- FR-4: The rest of column 1 (selected tag detail, controls) is unaffected

## Non-Goals

- No changes to the standalone mini player (used for narrow terminals < 80 cols) — that stays centered with its warning message
- No changes to the Selected Tag detail card
- No changes to the bordered control boxes
- No new data or state — all info comes from existing `StatusBarState`

## Technical Considerations

- `RenderMiniPlayer()` currently auto-sizes its width from content and centers in `termWidth`. For column 1, the card should fill the column width instead. Options:
  - Add a width parameter to `RenderMiniPlayer` (or create a `RenderMiniPlayerFixed(state, width)` variant) that uses the given width instead of auto-sizing
  - Or call the existing function with `termWidth = col1Width` and it will left-align since `col1Width` will be close to the natural card width
- The existing `RenderMiniPlayer()` includes a warning line ("Mini player mode — resize for full view") — the column 1 version should NOT include this warning
- The existing `RenderMiniPlayer()` does not show overlay status — it may need adding to match the current plain-text card
- The plain-text playback card in `renderColumn1()` spans lines ~1767-1791 in `tui/tui.go` — this block gets replaced with the mini player card output split into lines and appended
- Since `RenderMiniPlayer` returns a multi-line string, split on `\n` and append each line to the `lines` slice

## Success Metrics

- Column 1 playback section renders as a bordered card matching the wireframe style
- No information lost compared to the current plain-text card
- Visual consistency with the control boxes below

## Open Questions

- Should `RenderMiniPlayer` be refactored to accept a `width` param and a `showWarning` bool, or should a separate function be created for the column 1 variant?
