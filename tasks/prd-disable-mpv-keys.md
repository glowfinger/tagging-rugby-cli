# PRD: Disable mpv Default Keybindings

## Introduction

Disable mpv's built-in keyboard shortcuts to prevent users from accidentally triggering mpv actions (fullscreen toggle, subtitle cycling, screenshot, audio track switching, etc.) that disrupt the tagging workflow. Only three mpv keys are kept: space (toggle pause), q (quit), and f (fullscreen toggle). All other playback control is handled exclusively through the TUI via IPC.

## Goals

- Prevent accidental mpv key presses from disrupting the tagging workflow
- Keep only space (pause), q (quit), and f (fullscreen) active in mpv
- Embed a minimal `input.conf` in the Go binary so no external config files are needed
- Write the embedded config to a temp file at launch and pass it to mpv via `--input-conf`

## User Stories

### US-099: Embed minimal mpv input.conf in Go binary
**Description:** As a developer, I need a minimal mpv input configuration embedded in the binary so it can be written to disk at launch time without requiring users to manage config files.

**Acceptance Criteria:**
- [ ] Create an `input.conf` file in the `mpv/` package directory with only three bindings: `space cycle pause`, `q quit`, `f cycle fullscreen`
- [ ] Embed the file using `go:embed` in the `mpv` package (e.g. in `launch.go` or a new `config.go`)
- [ ] The embedded content is accessible as a `[]byte` or `string` variable
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

### US-100: Write embedded input.conf to temp file and pass to mpv at launch
**Description:** As a user, I want mpv to launch with only the essential keybindings so I don't accidentally trigger disruptive actions like subtitle cycling or screenshots.

**Acceptance Criteria:**
- [ ] `LaunchMpv()` writes the embedded `input.conf` content to a temporary file (e.g. `/tmp/tagging-rugby-input.conf` or `os.CreateTemp`)
- [ ] mpv is launched with `--no-input-default-bindings` flag to strip all built-in defaults
- [ ] mpv is launched with `--input-conf=<temp-file-path>` pointing to the written config
- [ ] The temp file is cleaned up when the mpv process exits or the app shuts down (best effort — use `defer os.Remove()`)
- [ ] mpv only responds to space (pause), q (quit), and f (fullscreen) — all other keys are ignored by mpv
- [ ] The TUI's IPC-based controls (seek, speed, mute, frame step, overlay, etc.) continue to work exactly as before
- [ ] Typecheck passes (`CGO_ENABLED=0 go vet ./...`)

## Functional Requirements

- FR-1: A minimal `input.conf` file is embedded in the `mpv` package containing exactly three bindings: `space cycle pause`, `q quit`, `f cycle fullscreen`
- FR-2: `LaunchMpv()` writes the embedded config to a temp file before starting the mpv process
- FR-3: mpv is started with `--no-input-default-bindings` and `--input-conf=<temp-path>` flags in addition to the existing `--input-ipc-server` flag
- FR-4: The temp config file is removed after mpv exits (best effort cleanup)
- FR-5: All existing IPC-based commands (seek, speed, mute, frame step, overlay, A-B loop, etc.) are unaffected — `--no-input-default-bindings` only affects keyboard input, not IPC

## Non-Goals

- No changes to the TUI keybindings or controls
- No changes to mpv exit/crash handling (current "not connected" behavior is kept)
- No user-configurable mpv keybindings (the embedded config is fixed)
- No changes to the mpv IPC client or socket communication

## Technical Considerations

- `go:embed` can only access files within the same package directory — the `input.conf` file must live in `mpv/`
- `--no-input-default-bindings` disables mpv's ~100+ default key bindings but does NOT affect IPC commands — all `SetProperty`, `Seek`, etc. calls continue to work
- `--input-conf` takes precedence and provides the only active keybindings when combined with `--no-input-default-bindings`
- Temp file should use a predictable path (e.g. `/tmp/tagging-rugby-input.conf`) or `os.CreateTemp` — either approach works since only one instance runs at a time
- The `LaunchMpv` function signature may need to return the temp file path for cleanup, or cleanup can be handled via `defer` in the caller

## Success Metrics

- Pressing any key other than space, q, or f in the mpv window has no effect
- All TUI controls (seek, speed, mute, frame step, overlay, clip loop) work exactly as before via IPC
- No leftover temp files after normal app shutdown

## Open Questions

- None
