# PRD: Mini Player for Small Terminal Windows

## Introduction

When the terminal is resized below 80 columns, the TUI currently shows a static "Terminal too narrow" warning and becomes completely unusable. This feature replaces that dead-end with a functional mini player that shows playback status and accepts basic playback controls. A warning is displayed to inform the user they are in mini player mode with limited functionality.

## Wireframe Reference

```
 ┌──────────┐
┌┤ Playback ├┐
│└──────────┘└────────────┐
│ ⏸ Paused      Step: 30s │
├─────────────────────────┤
│ Time: 1:11:22 / 1:08:11 │
└─────────────────────────┘
```

Source: `wireframe/mini-player.txt` — this represents the mini player at its smallest width.

## Goals

- Replace the "Terminal too narrow" dead screen with a functional mini player when width drops below `minTerminalWidth` (80 cols)
- Provide basic playback controls (play/pause, seek, step size, mute) in mini player mode
- Warn the user that they are in a limited mode so they know to resize for full functionality
- Auto-switch between mini player and full layout based on terminal width

## User Stories

### US-001: Render mini player view
**Description:** As a user, I want to see a compact playback card when my terminal is too narrow for the full layout so I can still monitor and control video playback.

**Acceptance Criteria:**
- [ ] When terminal width is below `minTerminalWidth` (80 cols), the mini player renders instead of the "Terminal too narrow" warning
- [ ] The mini player displays a bordered card matching the wireframe: header "Playback", play/pause state with step size, and time position / duration
- [ ] The card is horizontally centered if terminal width exceeds the card width
- [ ] The play/pause icon updates live (▶ Playing / ⏸ Paused)
- [ ] Step size displays the current value (e.g., "Step: 30s", "Step: 0.5s")
- [ ] Time displays as `MM:SS / MM:SS` (current position / total duration)
- [ ] Mute icon (🔇) is shown on the status line when audio is muted
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-002: Display mini player warning
**Description:** As a user, I want to see a warning message below the playback card so I know I'm in a limited mode and should resize for full features.

**Acceptance Criteria:**
- [ ] A warning line appears below the playback card (e.g., "Mini player mode — resize for full view")
- [ ] The warning uses a dim/subtle style (e.g., `styles.Lavender` italic) so it doesn't dominate the view
- [ ] Typecheck/vet passes

### US-003: Enable playback keybindings in mini player mode
**Description:** As a user, I want basic playback controls to work in mini player mode so I can control the video without resizing.

**Acceptance Criteria:**
- [ ] `space` toggles play/pause
- [ ] `h`/`left` seeks backward by current step size
- [ ] `l`/`right` seeks forward by current step size
- [ ] `<`/`,` decreases step size
- [ ] `>`/`.` increases step size
- [ ] `m` toggles mute
- [ ] `ctrl+h` frame back step, `ctrl+l` frame forward step
- [ ] `ctrl+c` quits the application
- [ ] `?` shows help overlay (existing behavior)
- [ ] All other keybindings (t, n, e, x, j, k, enter, :, s, o) are ignored / non-functional in mini player mode
- [ ] Typecheck/vet passes

### US-004: Auto-switch between mini and full layout
**Description:** As a user, I want the TUI to automatically switch between mini player and full layout when I resize my terminal so I don't have to do anything manually.

**Acceptance Criteria:**
- [ ] Resizing terminal below 80 cols switches to mini player view on the next render
- [ ] Resizing terminal to 80+ cols switches back to the full column layout on the next render
- [ ] State is preserved across transitions — playback position, notes list selection, step size, etc. remain unchanged
- [ ] No flicker or crash during rapid resizing
- [ ] Typecheck/vet passes

## Functional Requirements

- FR-1: Replace the "Terminal too narrow" block in `View()` with a mini player render when `m.width > 0 && m.width < minTerminalWidth`
- FR-2: The mini player card renders using lipgloss borders matching the wireframe layout (tab-style "Playback" header, horizontal divider between status and time lines)
- FR-3: The mini player card content is sourced from the existing `StatusBarState` struct — no new data sources needed
- FR-4: Playback keybindings (space, h/l, </>. m, ctrl+h/ctrl+l, ctrl+c, ?) work in mini player mode
- FR-5: Data keybindings (t, n, e, x, j, k, enter, :, s, o) are suppressed when in mini player mode
- FR-6: A warning line renders below the card indicating limited mode
- FR-7: The tick loop continues running in mini player mode so playback state stays up-to-date

## Non-Goals

- No notes list or data display in mini player mode
- No tackle/note creation or editing in mini player mode
- No manual toggle keybinding — switching is purely based on terminal width
- No command input (`:`) mode in mini player
- No timeline progress bar in mini player
- No stats view access in mini player

## Design Considerations

- Reuse the existing `StatusBarState` struct for all data — no new state needed
- The mini player card should use the existing color palette (`styles.Pink` for header, `styles.LightLavender` for content, `styles.Purple` for borders)
- Use lipgloss `Border()` to draw the card frame — matches the wireframe's box-drawing characters
- The "Playback" header tab can be rendered using lipgloss `JoinHorizontal` with a styled tab element
- Card width should be the minimum of the terminal width and the natural content width (~27 chars from the wireframe)

## Technical Considerations

- The mini player check in `View()` replaces the existing `minTerminalWidth` warning block — same location, different render
- Key suppression can be handled by checking `m.width < minTerminalWidth` early in the `tea.KeyMsg` switch and only allowing playback keys through
- The tick loop already runs unconditionally, so status updates work without changes
- Forms (tackle/note) should not be openable in mini player mode; if a form is active when the terminal shrinks, it may render poorly — consider closing active forms or letting them render at whatever width is available
- The `WindowSizeMsg` handler already stores width/height; no changes needed there

## Success Metrics

- Terminal below 80 cols shows a functional mini player instead of a dead warning screen
- Users can play/pause and seek without resizing
- Resizing to full width restores the complete layout with no state loss

## Open Questions

- Should the mini player have a minimum width below which even it can't render (e.g., below 25 cols)? If so, fall back to a simple text message?
- If a huh form (tackle/note) is active when the user shrinks below 80 cols, should the form be forcibly closed or allowed to render at the narrow width?
