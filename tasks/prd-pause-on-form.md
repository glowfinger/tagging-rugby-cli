# PRD: Pause Video When Form Is Open

## Introduction

When a user opens a form (note, tackle, or confirm discard), the video continues playing in the background. This means the user may miss action while filling in form fields, and the timestamp captured at form open drifts from what they're watching. Automatically pausing mpv when any form opens solves this by freezing playback so the user can focus on data entry without missing anything.

## Goals

- Automatically pause mpv when any form opens (note, tackle, confirm discard)
- Respect the user's existing pause state — if already paused, do nothing
- Never auto-resume — the user manually unpauses after closing the form
- Keep the implementation minimal with no new state beyond tracking whether we triggered the pause

## User Stories

### US-101: Pause mpv when a form opens
**Description:** As a user, I want the video to automatically pause when I open a form so that I don't miss any action while entering data.

**Acceptance Criteria:**
- [ ] When `openNoteInput()` is called and the video is currently playing, mpv is paused via `m.client.Pause()`
- [ ] When `openTackleInput()` is called and the video is currently playing, mpv is paused via `m.client.Pause()`
- [ ] When the confirm discard form is opened and the video is currently playing, mpv is paused via `m.client.Pause()`
- [ ] If the video is already paused when a form opens, no pause command is sent (existing state is preserved)
- [ ] The pause happens after the timestamp is captured (so the recorded timestamp reflects the moment the user pressed the key, not a frame later)
- [ ] Video does NOT auto-resume when the form is closed (submitted, cancelled, or discarded)
- [ ] The client nil/connection check is performed before calling Pause (`m.client != nil && m.client.IsConnected()`)
- [ ] Typecheck/lint passes (`CGO_ENABLED=0 go vet ./...`)

## Functional Requirements

- FR-1: In `openNoteInput()`, after capturing the timestamp, call `m.client.Pause()` if the client is connected and video is currently playing
- FR-2: In `openTackleInput()`, after capturing the timestamp, call `m.client.Pause()` if the client is connected and video is currently playing
- FR-3: When the confirm discard form is opened, call `m.client.Pause()` if the client is connected and video is currently playing
- FR-4: Check `m.client.GetPaused()` before pausing — only pause if currently playing (to avoid unnecessary IPC calls and to preserve the user's intent)
- FR-5: No resume logic is needed — the user manually unpauses with Space

## Non-Goals

- No auto-resume when a form is closed
- No visual indicator that the pause was triggered by a form (the existing pause indicator in the status bar is sufficient)
- No configuration or toggle to disable this behaviour

## Technical Considerations

- The mpv IPC client (`m.client`) already has `GetPaused() (bool, error)` and `Pause() error` methods — no new mpv functionality needed
- The `openNoteInput()` and `openTackleInput()` methods already capture the timestamp before creating the form — the pause call should go after the timestamp capture but before `form.Init()`
- The confirm discard form is created inline in `handleNoteFormUpdate()` and `handleTackleFormUpdate()` — the pause call should be added at those creation points
- Errors from `Pause()` can be silently ignored (consistent with existing space-bar toggle pattern: `_ = m.client.TogglePause()`)

## Success Metrics

- Video is always paused when a form is visible
- User's manual pause state is never overridden (if paused before form, stays paused after form closes)
- No regressions in form submission or timestamp accuracy

## Open Questions

None — scope is well-defined.
