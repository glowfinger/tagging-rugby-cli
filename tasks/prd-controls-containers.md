# PRD: Controls in Bordered Group Containers

## Introduction

The controls section in column 1 currently renders as a flat vertical list with emojis, names on one line, and shortcuts on the next. This makes it hard to visually parse related controls. This feature replaces that with bordered containers (boxes) per group, matching the wireframe style in `wireframe/playback.txt`. Each group gets a tab-style header and compact single-line entries. Sub-groups within a container are separated by horizontal dividers (`├───┤`).

## Wireframe Reference

```
 ┌──────────┐
┌┤ Playback ├┐
│└──────────┘└────────────┐
│ Play    [ Space ]       │
│ Back    [ H / ← ]       │
│ Fwd     [ L / → ]       │
├─────────────────────────┤
│ Step -  [ , / < ]       │
│ Step +  [ . / > ]       │
├─────────────────────────┤
│ Frame - [ Ctrl + h ]    │
│ Frame + [ Ctrl + l ]    │
└─────────────────────────┘
```

Source: `wireframe/playback.txt`

## Goals

- Reorganize controls into three logically grouped bordered containers
- Match the wireframe visual style: tab header, box-drawing borders, horizontal dividers between sub-groups
- Use compact single-line format (`Name    [ Shortcut ]`) with no emojis
- Improve scannability of the controls section in column 1

## User Stories

### US-001: Reorganize control groups
**Description:** As a developer, I need to restructure the `GetControlGroups()` data to match the new three-group layout so the rendering logic has the right data.

**Acceptance Criteria:**
- [ ] **Playback** group contains three sub-groups (separated by dividers when rendered):
  - Play/Pause: `Play [Space]`, `Back [H/←]`, `Fwd [L/→]`
  - Step: `Step - [,/<]`, `Step + [./>]`
  - Frame: `Frame - [Ctrl+h]`, `Frame + [Ctrl+l]`
- [ ] **Navigation** group contains: `Prev [J/↑]`, `Next [K/↓]`, `Mute [M]`, `Overlay [O]`
- [ ] **Views** group contains: `Stats [S]`, `Help [?]`, `Quit [Ctrl+C]`
- [ ] The `ControlGroup` struct supports sub-groups (e.g., a `SubGroups [][]Control` field or similar) so the renderer knows where to place horizontal dividers
- [ ] Emoji field can be removed from `Control` struct or left unused
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-002: Render bordered container component
**Description:** As a developer, I need a reusable function that renders a group of controls inside a bordered box with a tab-style header and optional horizontal dividers between sub-groups.

**Acceptance Criteria:**
- [ ] New render function (e.g., `RenderControlBox(group, width)`) in `tui/components/controls.go`
- [ ] Box uses box-drawing characters: `┌ ┐ └ ┘ │ ─` for the outer border
- [ ] Tab-style header: the group name appears as a tab on the top border (e.g., `┌┤ Playback ├┐` extending from the top-left corner)
- [ ] Horizontal dividers (`├─────┤`) separate sub-groups within the box
- [ ] Each control renders as a single line: `Name` left-aligned, `[ Shortcut ]` right-aligned, padded to fill the box width
- [ ] No emojis in the output
- [ ] Border color uses `styles.Purple`, header text uses `styles.Pink` or `styles.Amber`
- [ ] Control name uses `styles.LightLavender`, shortcut uses `styles.Cyan`
- [ ] Box width adapts to the available column width
- [ ] Typecheck/vet passes

### US-003: Update column 1 to use bordered containers
**Description:** As a user, I want to see controls displayed in neat bordered boxes in column 1 so I can quickly find the shortcut I need.

**Acceptance Criteria:**
- [ ] `renderColumn1()` in `tui/tui.go` renders each control group as a bordered container using the new component
- [ ] The Playback box shows three sub-groups separated by horizontal dividers (matching the wireframe)
- [ ] The Navigation box shows all items without internal dividers
- [ ] The Views box shows all items without internal dividers
- [ ] A small gap (1 blank line) separates each bordered container
- [ ] The existing "Playback" status card and "Selected Tag" detail card above the controls remain unchanged
- [ ] Controls fit within the column width without overflow or wrapping
- [ ] Typecheck/vet passes

### US-004: Update horizontal controls display bar
**Description:** As a developer, I need to update the `ControlsDisplay()` (shown during form input) to use the new group structure.

**Acceptance Criteria:**
- [ ] `ControlsDisplay()` continues to render as a single horizontal bar (not bordered boxes)
- [ ] Uses the updated group data from `GetControlGroups()` (new 3-group structure)
- [ ] No emojis in the horizontal bar output (consistent with the new style)
- [ ] Format: `Name [Shortcut]` items space-separated, groups separated by wider spacing
- [ ] Typecheck/vet passes

## Functional Requirements

- FR-1: Controls are organized into three groups: **Playback** (play/step/frame with sub-groups), **Navigation** (prev/next/mute/overlay), **Views** (stats/help/quit)
- FR-2: Each group renders as a bordered box with a tab-style header in column 1
- FR-3: Sub-groups within a box are separated by horizontal dividers (`├───────┤`)
- FR-4: Each control is a single line: name left-aligned, `[ Shortcut ]` right-aligned
- FR-5: No emojis in the controls display
- FR-6: Box width fills the available column 1 width
- FR-7: The horizontal `ControlsDisplay` bar (shown during forms) updates to match the new group structure but remains a flat horizontal layout

## Non-Goals

- No interactive controls (clicking/selecting controls to trigger them)
- No collapsible/expandable groups
- No changes to the actual keybindings themselves
- No bordered containers for the Playback status card or Selected Tag card — only controls

## Design Considerations

- Use raw box-drawing characters (`│`, `─`, `┌`, `┐`, `└`, `┘`, `├`, `┤`) rather than lipgloss `Border()` — this gives precise control over the tab header and horizontal divider rendering that lipgloss borders don't natively support
- Color the border characters with `styles.Purple` to match the existing vertical column separators
- The tab header style (`┌┤ Name ├┐`) is non-standard — it needs manual string construction, not a lipgloss border style
- Keep padding minimal (1 space inside borders) to maximize content width in narrow columns

## Technical Considerations

- The `ControlGroup` struct needs to support sub-groups for the Playback container's internal dividers. Options:
  - Change `Controls []Control` to `SubGroups [][]Control` — each sub-group is a slice of controls
  - Or add a `Divider bool` flag on `Control` to indicate "draw a divider before this control"
- The `renderColumn1()` function currently iterates groups and controls with per-item blank lines and sub-headers — this entire block (lines ~1642-1681 in `tui/tui.go`) will be replaced with calls to the new bordered container renderer
- Column 1 has limited height (`colHeight`) — with 3 bordered containers (Playback ~12 lines, Navigation ~6 lines, Views ~5 lines = ~23 lines + gaps), ensure it fits or scrolls gracefully if the terminal is short

## Success Metrics

- Controls section in column 1 matches the wireframe visual style
- Users can identify the shortcut for any control within 2 seconds of scanning
- No increase in column 1 width requirements

## Open Questions

- Should the Navigation group have a different name (e.g., "General", "Other")? Current spec uses "Navigation" to cover prev/next/mute/overlay.
