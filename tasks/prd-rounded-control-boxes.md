# PRD: Rounded Control Boxes

## Introduction

The control boxes in column 4 (Playback, Navigation, Views) use a square tab-style border (`┌┤ Playback ├┐` / `└┘`) rendered by `RenderControlBox`, while every other bordered element in the TUI uses the rounded `RenderInfoBox` style (`╭─ Title ─────╮` / `╰╯`). This PRD migrates the control boxes to use `RenderInfoBox` for visual consistency and codebase simplification.

## Goals

- Unify all bordered containers to the rounded `RenderInfoBox` style
- Simplify column 4 rendering by reusing the existing `RenderInfoBox` function
- Mark `RenderControlBox` as deprecated (keep for reference, do not delete)
- Update TUI-ARCHITECTURE.md to reflect the change

## User Stories

### US-001: Render control groups using RenderInfoBox
**Description:** As a developer, I want column 4's control groups to render with the same rounded border style as every other container so the UI is visually consistent.

**Acceptance Criteria:**
- [ ] `renderColumn4` in `tui/columns.go` calls `RenderInfoBox` instead of `RenderControlBox` for each control group
- [ ] Each control group (Playback, Navigation, Views) renders with rounded corners (`╭╮╰╯`) and inline header (`╭─ Playback ─────╮`)
- [ ] Control rows inside each box show `Name  [ Shortcut ]` format (same as before)
- [ ] Sub-groups within a control group (e.g., Playback has 4 sub-groups) are separated by a blank line instead of a horizontal divider
- [ ] Build passes: `CGO_ENABLED=0 go build ./...`
- [ ] Visual check: control boxes match the rounded style of Video, Summary, and InfoBox containers

### US-002: Mark RenderControlBox as deprecated
**Description:** As a developer, I want `RenderControlBox` marked deprecated so it's clear the rounded style should be used for new code, while keeping the old function available for reference.

**Acceptance Criteria:**
- [ ] `RenderControlBox` in `tui/components/controls.go` has a `// Deprecated:` doc comment explaining to use `RenderInfoBox` instead
- [ ] No call sites remain that use `RenderControlBox` (only the deprecated function definition stays)
- [ ] Build passes: `CGO_ENABLED=0 go build ./...`

### US-003: Update TUI-ARCHITECTURE.md
**Description:** As a developer, I want the architecture doc to reflect that column 4 now uses `RenderInfoBox` so documentation stays accurate.

**Acceptance Criteria:**
- [ ] The Controls section in `tui/TUI-ARCHITECTURE.md` notes that `RenderControlBox` is deprecated
- [ ] The Column Rendering table for column 4 mentions `RenderInfoBox` instead of `RenderControlBox`
- [ ] The `controls.go` entry in the Directory Structure section reflects the deprecation

## Functional Requirements

- FR-1: Build a helper function (or inline logic) that converts a `ControlGroup` into `[]string` content lines suitable for `RenderInfoBox`. Each control renders as `Name  [ Shortcut ]` left-aligned. Blank lines separate sub-groups.
- FR-2: `renderColumn4` iterates `GetControlGroups()`, converts each to content lines, and calls `RenderInfoBox(group.Name, contentLines, width, false)` for each group.
- FR-3: Add `// Deprecated: Use RenderInfoBox instead.` comment to `RenderControlBox`.
- FR-4: Update `tui/TUI-ARCHITECTURE.md` to document the change.

## Non-Goals

- No deletion of `RenderControlBox` — it stays as deprecated
- No changes to the control group data structure (`ControlGroup`, `Control`)
- No changes to `RenderInfoBox` itself
- No changes to `ControlsDisplay` (the horizontal bar used in form overlays)
- No changes to keybinding definitions or behavior

## Technical Considerations

- `RenderInfoBox` takes `(title string, contentLines []string, width int, focused bool)` — pass `focused: false` for control boxes since column 4 has no focus state
- Content lines for `RenderInfoBox` must be pre-styled strings. Use the same `nameStyle` (LightLavender) and `shortcutStyle` (Cyan, bold) from the existing `RenderControlBox`
- The helper that builds content lines from a `ControlGroup` can live in `controls.go` as a private function (e.g., `controlGroupLines`)
- Sub-group separation changes from `├───┤` dividers to blank lines — this slightly changes the visual density but improves consistency

## Success Metrics

- All bordered containers in the TUI use the same rounded visual style
- `RenderControlBox` has zero active call sites
- Architecture documentation is accurate

## Open Questions

- None — scope is well defined.
