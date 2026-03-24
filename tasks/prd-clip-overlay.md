# PRD: Clip Overlay Update

## Introduction

Replace the current 3-line drawtext overlay on exported clips with a 5-line overlay showing richer metadata. The overlay appears in the bottom-left of the frame for the first 2 seconds of the clip (reduced from 3). Text is white with a `#131211` 2-pixel border/outline. All required data is already present in the `PendingClip` struct — this is a single-file change to `clip/processor.go`.

## Goals

- Show 5 lines of metadata on each clip for the first 2 seconds
- Apply a dark `#131211` 2-pixel text border to all drawtext lines for legibility on any background
- Keep the overlay positioned at bottom-left (x=10), stacked upward

## User Stories

### US-001: Update clip overlay drawtext filter
**Description:** As a user, I want the exported clip to show a richer 5-line metadata overlay in the bottom-left for the first 2 seconds so I can immediately identify the clip without watching it.

**Acceptance Criteria:**
- [ ] The drawtext filter in `clip/processor.go` `processClip()` is updated to produce exactly 5 lines
- [ ] Line content (from bottom to top of frame):
  - Line 1 (bottom): `Note ID: {c.NoteID}` — `y=h-th`
  - Line 2: `Filename: {c.Filename}` — `y=h-th-36`
  - Line 3: `Outcome: {c.Outcome} {c.Attempt}` — `y=h-th-72`
  - Line 4: `Player: {c.Player}` — `y=h-th-108`
  - Line 5 (top): `Time: {startFormatted}/{endFormatted}` — `y=h-th-144`
- [ ] `{startFormatted}` and `{endFormatted}` are both formatted as `HH\:MM\:SS` using the same colon-escaping logic already applied to `timestamp` (i.e. `%02d\\\\:%02d\\\\:%02d`)
- [ ] Every drawtext segment includes `bordercolor=#131211:borderw=2` (text outline, not box)
- [ ] The `enable` expression on every drawtext segment is `lt(t\\,2)` (2 seconds, reduced from 3)
- [ ] `fontsize=28`, `fontcolor=white`, `x=10` unchanged on all lines
- [ ] No changes to any other file — only `clip/processor.go`
- [ ] `CGO_ENABLED=0 go vet ./...` passes

## Functional Requirements

- FR-1: Overlay duration is 2 seconds — `enable=lt(t\\,2)` on every drawtext segment
- FR-2: Text border is `bordercolor=#131211:borderw=2` on every drawtext segment (ffmpeg drawtext outline, not `box`)
- FR-3: Five drawtext segments are chained with `,` in a single `-vf` value
- FR-4: Line 5 (Time) uses both `c.Start` and `c.End`, each formatted as `HH\\:MM\\:SS` with colons double-escaped for the filtergraph
- FR-5: Line 3 (Outcome) formats as `Outcome: {outcome} {attempt}` — outcome string followed by a space and the integer attempt number
- FR-6: Line spacing is 36px between each drawtext baseline (`y` decrements by 36 per line from bottom)
- FR-7: `c.Filename` and `c.NoteID` are already in `PendingClip` — no DB or struct changes required

## Non-Goals

- No changes to overlay font, font size, or x-position
- No changes to clip duration, folder/filename structure, or any other ffmpeg arguments
- No UI changes in the TUI
- No DB schema changes
- No changes to the `clip export` CLI command in `cmd/clip.go` (separate code path)

## Technical Considerations

- The current `timestamp` variable is built as `fmt.Sprintf("%02d\\\\:%02d\\\\:%02d", hours, minutes, seconds)` using `c.Start` — replicate the same logic for `c.End` to produce `endTimestamp`
- ffmpeg `drawtext` color format: `#RRGGBB` hex is accepted directly in the filter string (e.g. `bordercolor=#131211`)
- The `borderw` option draws an outline around each glyph — it does not add a background box (`box=1` would be needed for a box, which is not wanted here)
- The existing colon-escaping comment in `processClip` explains the `\\\\:` requirement — both `startTimestamp` and `endTimestamp` need the same treatment
- Any free-text fields placed in drawtext (e.g. `c.Filename`, `c.Player`, `c.Outcome`) should have colons replaced with `\\\\:` via `strings.ReplaceAll` to prevent filtergraph parse errors — the existing `outcome` variable already does this, apply same to `filename` and `player`

## Success Metrics

- Exported clips show all 5 metadata lines at the bottom-left for the first 2 seconds
- Text is readable on both light and dark backgrounds due to the `#131211` border
- `go vet` passes with no errors after the change

## Open Questions

- None — all requirements are explicit and all data is available in `PendingClip`
