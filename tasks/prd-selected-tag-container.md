# PRD: Bordered Container for Selected Tag

## Introduction

The "Selected Tag" detail card in column 1 currently renders as plain unstyled text lines (ID, type, timestamp, category, player, team, text) with just a pink header. The playback card and control groups already use bordered containers with tab-style headers. This change wraps the selected tag detail in the same bordered box style for visual consistency across column 1.

## Goals

- Wrap the selected tag detail section in a bordered container with a "Selected Tag" tab header
- Match the visual style of the existing control boxes (`RenderControlBox`) and mini player (`RenderMiniPlayer`)
- Keep all the same detail fields currently shown

## User Stories

### US-001: Render selected tag in bordered container
**Description:** As a user, I want the selected tag detail to appear in a bordered box matching the rest of column 1 so the UI looks consistent.

**Acceptance Criteria:**
- [ ] The selected tag section in `renderColumn1()` renders inside a bordered box with tab-style "Selected Tag" header (same box-drawing style as `RenderControlBox`: `┌┤ Selected Tag ├┐` with `│` sides and `└───┘` bottom)
- [ ] All existing detail fields are shown inside the box: `#ID Type ★`, `@ H:MM:SS`, `[category]`, `Player: name`, `Team: name`, text
- [ ] Optional fields (category, player, team, text) still only appear when present — the box height is dynamic
- [ ] When no item is selected, the box is either hidden entirely (current behavior: nothing renders) or shows an empty box with a "No selection" message
- [ ] Border color uses `styles.Purple`, header text uses `styles.Pink`, detail text uses `styles.LightLavender`, dim fields use `styles.Lavender` — consistent with existing styling
- [ ] The box fills the column 1 width
- [ ] Text field truncation still works correctly within the box inner width
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

## Functional Requirements

- FR-1: The selected tag section renders as a bordered box with a tab-style "Selected Tag" header
- FR-2: The box uses the same box-drawing characters and color scheme as the control boxes
- FR-3: Content rows are left-aligned with 1-space padding inside borders
- FR-4: The box height is dynamic — only rows with data are included
- FR-5: When no item is selected, the entire box is hidden (matches current behavior)

## Non-Goals

- No new data fields — the content inside the box is identical to what's currently shown
- No interactivity on the selected tag box (no click, no expand/collapse)
- No changes to how item selection works (j/k navigation, enter to jump)

## Design Considerations

- Reuse the existing `RenderControlBox` pattern or create a generic `RenderInfoBox(title string, lines []string, width int)` helper that both the selected tag and other future info cards can use
- The box-drawing pattern (tab header 3 lines + content rows + bottom border) is already implemented in `RenderControlBox` and `RenderMiniPlayer` — extracting a shared helper avoids further duplication
- The selected tag box sits between the mini player card (above) and the control boxes (below) in column 1

## Technical Considerations

- The change is in `renderColumn1()` in `tui/tui.go` (lines ~1793-1835) — replace the inline styled lines with a bordered card render
- The content lines (ID/type, timestamp, category, player, team, text) can be built as a `[]string` slice and passed to a box renderer
- Text truncation (lines 1823-1831) needs to account for the inner width being 2 chars narrower (left border + padding, padding + right border) than the column width
- A generic box renderer could accept `title string`, `contentLines []string`, and `width int` — this would be reusable for both the selected tag and any future info cards

## Success Metrics

- The selected tag section visually matches the control boxes and mini player card
- All detail fields remain visible and correctly formatted
- Column 1 has a consistent bordered-box visual style from top to bottom

## Open Questions

- Should a generic `RenderInfoBox(title, lines, width)` be extracted, or should this use a one-off render like `RenderMiniPlayer` does?
