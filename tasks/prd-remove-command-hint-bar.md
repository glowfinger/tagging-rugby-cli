# PRD: Remove Command Mode Hint Bar

## Introduction

The bottom bar of the TUI displays a persistent hint message — "Press : to enter command mode, ? for help, q to quit" — when not in command mode and no result message is showing. This hint takes up a full line of vertical space and is unnecessary once the user has learned the keybindings (which are already shown in the controls panel and help overlay). This change removes the hint, rendering an empty styled bar in its default state instead.

## Goals

- Reclaim the bottom bar's default state as a clean empty line
- Keep command mode (`:` prompt + input) and result messages (success/error) working exactly as they do today

## User Stories

### US-001: Remove default hint text from command input bar
**Description:** As a user, I want the bottom bar to be empty when not in use so the UI is cleaner and I get more visual breathing room.

**Acceptance Criteria:**
- [ ] When command mode is not active and no result message is showing, the bottom bar renders as an empty styled line (dark purple background, full width) with no text
- [ ] When command mode is active, the `:` prompt and input with cursor still render exactly as before
- [ ] When a result message is showing (success or error), it still renders exactly as before
- [ ] The `CommandInput()` function in `tui/components/commandinput.go` no longer outputs "Press : to enter command mode, ? for help, q to quit"
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

## Functional Requirements

- FR-1: The default (idle) state of the command input bar is an empty line with `DarkPurple` background at full width
- FR-2: Command mode active state (`:` prompt, input, cursor) is unchanged
- FR-3: Result message display (success in cyan, error in pink) is unchanged

## Non-Goals

- No removal of the command input bar itself — it still occupies 1 line at the bottom
- No changes to command mode functionality or keybindings
- No changes to result message timing or display

## Technical Considerations

- The change is a single block in `CommandInput()` in `tui/components/commandinput.go` (lines 79-88): replace the hint text render with an empty styled line
- The `hintStyle` variable and the string literal on line 88 can be removed entirely
- The empty bar should still use `lineStyle` with `Background(styles.DarkPurple).Width(width)` for visual consistency

## Success Metrics

- Bottom bar is visually clean when idle
- No loss of command mode or result message functionality

## Open Questions

- None.
