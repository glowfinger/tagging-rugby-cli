# PRD: Playback Speed Control

## Introduction

Add the ability to control video playback speed from the TUI using `[`/`{` to decrease speed, `]`/`}` to increase speed, and `\` to reset to 1.0x. The current speed is always visible in the status bar. Speed cycles through a predefined list of values, following the same pattern as step size control.

## Goals

- Allow users to change playback speed without leaving the TUI
- Provide clear visual feedback of the current speed in the status bar
- Use a predefined list of speed values for predictable, useful increments
- Support quick reset to normal speed with a single keypress

## User Stories

### US-001: Add speed state and display to status bar
**Description:** As a user, I want to see the current playback speed in the status bar so I always know what speed the video is playing at.

**Acceptance Criteria:**
- [ ] Add `Speed float64` field to `StatusBarState` struct in `tui/components/statusbar.go`
- [ ] Status bar displays speed (e.g., "Speed: 1.0x", "Speed: 1.5x") on the right side alongside step size
- [ ] Speed is always shown, even at 1.0x
- [ ] Format: `Speed: Nx` where N omits trailing zeros (e.g., `1x`, `1.5x`, `0.75x`)
- [ ] The `updateStatusFromMpv()` tick handler in `tui/tui.go` polls `client.GetSpeed()` and updates `statusBar.Speed`
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-002: Add speed control keybindings
**Description:** As a user, I want to press `[`/`{` to slow down and `]`/`}` to speed up playback, and `\` to reset to normal speed.

**Acceptance Criteria:**
- [ ] Define a predefined speed list: `[]float64{0.5, 0.75, 1.0, 1.25, 1.5, 2.0}`
- [ ] `[` and `{` decrease speed — cycle to previous value in the list; do nothing if already at minimum (0.5x)
- [ ] `]` and `}` increase speed — cycle to next value in the list; do nothing if already at maximum (2.0x)
- [ ] `\` resets speed to 1.0x
- [ ] All keys call `client.SetSpeed(multiplier)` to apply the change
- [ ] Keys only work in normal mode (not during command input, forms, or stats view)
- [ ] Keys work in mini player mode (playback control)
- [ ] Typecheck/vet passes

### US-003: Update controls display and help overlay
**Description:** As a user, I want to see the speed control keybindings in the controls panel and help overlay.

**Acceptance Criteria:**
- [ ] Add speed controls to the Playback group as a new sub-group (separated by divider): `Speed - [ [ / { ]`, `Speed + [ ] / } ]`, `Speed 1x [ \ ]`
- [ ] Help overlay updated with the three speed bindings
- [ ] Typecheck/vet passes

### US-004: Update mini player to show speed
**Description:** As a user, I want the mini player to show the current playback speed so I know what speed I'm at even in compact mode.

**Acceptance Criteria:**
- [ ] The mini player status line includes speed (e.g., "⏸ Paused  Speed: 1.5x  Step: 30s")
- [ ] Speed display uses the same format as the full status bar
- [ ] Typecheck/vet passes

## Functional Requirements

- FR-1: Playback speed cycles through a predefined list: `0.5, 0.75, 1.0, 1.25, 1.5, 2.0`
- FR-2: `[` and `{` decrease speed to the previous value in the list; clamp at 0.5x
- FR-3: `]` and `}` increase speed to the next value in the list; clamp at 2.0x
- FR-4: `\` resets speed to 1.0x regardless of current speed
- FR-5: Current speed is always displayed in the status bar as "Speed: Nx"
- FR-6: Speed state is polled from mpv on each tick via `GetSpeed()` and stored in `StatusBarState`
- FR-7: Speed controls work in both full layout and mini player mode

## Non-Goals

- No arbitrary speed input (e.g., typing "1.3x" in command mode)
- No per-section or per-clip speed memory
- No audio pitch correction toggle (mpv handles this internally)
- No speed display in the notes list or stats view

## Technical Considerations

- `mpv.Client` already has `SetSpeed(float64)` and `GetSpeed()` methods — no mpv client changes needed
- The speed list pattern mirrors the existing `stepSizes` variable and `increaseStepSize()`/`decreaseStepSize()` helpers in `tui/tui.go`
- Speed format: use `%g` to avoid trailing zeros (e.g., `1` not `1.0`, `1.5` not `1.50`, `0.75` stays `0.75`)
- The `\` key in Bubble Tea's `msg.String()` returns `"\\"` — verify the exact string before wiring up
- Status bar right side currently shows `Step: Xs [mute] [overlay]` — speed should be added before or after step size

## Success Metrics

- Users can change speed in 1 keypress
- Current speed is always visible in the status bar
- Reset to 1x is a single keypress away

## Open Questions

- Should the speed list be configurable or is the hardcoded list sufficient?
