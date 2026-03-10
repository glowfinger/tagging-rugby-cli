# PRD: Column 2 Content Replacement Pattern

## Introduction

Currently, all forms (Note, Tackle, Confirm Discard) and views (Help, Stats) are
rendered as **full-screen takeovers** that replace the entire TUI. This hides Column 1
(video status, export progress) and Column 4 (keybinding reference) while the user
interacts with a form — losing important context.

This feature replaces Column 2's normal content (search bar + notes list) with a
full-width, full-height container that renders the active form or view. Forms and views
are **not** floating overlays — they are placed statically in the Column 2 slot, sized
to fill it exactly via `layout.Container{Width, Height}.Render(...)`. Columns 1 and 4
stay visible at all times. When Column 2 is in replacement mode, Column 3 (stats) hides
to give Column 2 the extra space.

The TUI-ARCHITECTURE.md already documents this as the target design. This PRD implements
what the architecture describes.

---

## Goals

- All five items (Note form, Tackle form, Confirm Discard, Help, Stats) replace Column 2's
  search bar + notes list as full-width, full-height containers
- The replacement content is sized exactly to the Column 2 slot via `layout.Container`
- Column 3 hides whenever Column 2 is in replacement mode (`overlayActive` flag)
- Column 2 expands to use the space freed by hiding Column 3
- `ComputeColumnWidths` accepts an `overlayActive bool` parameter and applies the
  overlay layout table from the architecture doc
- `Esc` is the single, unified key to exit any active replacement content
- Forms cannot be opened when the terminal is too narrow (width < 61)
- TUI-ARCHITECTURE.md is updated to reflect any implementation details that diverged
  from the original spec during development

---

## User Stories

### US-001: Add `overlayActive` to `ComputeColumnWidths`

**Description:** As a developer, I need `ComputeColumnWidths` to accept an `overlayActive`
parameter so Column 3 hides and Column 2 expands whenever a form or overlay is open.

**Acceptance Criteria:**
- [ ] `ComputeColumnWidths(termWidth int, overlayActive bool)` — new signature with second param
- [ ] When `overlayActive == false`, behaviour is identical to current (no regression)
- [ ] When `overlayActive == true`, Column 3 is always hidden regardless of terminal width
- [ ] Overlay layout widths match the architecture table:
  - `>= 170`: Col1=30, Col4=30, Col2=termWidth−60; Col3 hidden
  - `61–169`: Col1=30, Col2=termWidth−30; Col3 and Col4 hidden
  - `<= 60`: Col1=30 only; Col2, Col3, Col4 hidden
- [ ] All existing call-sites of `ComputeColumnWidths` updated to pass the new param
- [ ] `CGO_ENABLED=0 go vet ./...` passes

---

### US-002: Add `overlayActive` computation in `View()`

**Description:** As a developer, I need `View()` to compute a single `overlayActive` flag
and pass it everywhere that needs it, so the layout reacts consistently to any active form
or overlay.

**Acceptance Criteria:**
- [ ] `View()` computes `overlayActive := m.noteForm != nil || m.tackleForm != nil || m.confirmDiscardForm != nil || m.showHelp || m.statsView.Active`
- [ ] `overlayActive` is passed to `ComputeColumnWidths`
- [ ] `overlayActive` is passed to `renderColumn3` (to render empty / skip when true)
- [ ] `CGO_ENABLED=0 go vet ./...` passes

---

### US-003: Replace Column 2 with forms (Note, Tackle, Confirm Discard)

**Description:** As a user, I want the Note form, Tackle form, and Confirm Discard dialog
to replace Column 2's search bar and notes list so that Column 1 (video status, export)
and Column 4 (keybindings) remain visible while I fill in a form.

**Acceptance Criteria:**
- [ ] `renderColumn2(width, height int)` checks form state at the top, in this order:
  1. `m.confirmDiscardForm != nil` → return `layout.Container{Width:width, Height:height}.Render(m.confirmDiscardForm.View())`
  2. `m.noteForm != nil` → return `layout.Container{Width:width, Height:height}.Render(m.noteForm.View())`
  3. `m.tackleForm != nil` → return `layout.Container{Width:width, Height:height}.Render(m.tackleForm.View())`
  4. Otherwise → existing search input + notes list (unchanged)
- [ ] The full-screen early-return code in `View()` for all three forms is removed
- [ ] The search bar and notes list are not rendered while any form is active
- [ ] Column 1 and Column 4 are visible while any form is open (manual verification)
- [ ] Form content fills exactly the Column 2 width × height via Container
- [ ] `CGO_ENABLED=0 go vet ./...` passes

---

### US-004: Replace Column 2 with Help content

**Description:** As a user, I want pressing `?` to replace Column 2's search bar and
notes list with the Help content so I can read keybinding reference while Column 1 and
Column 4 remain visible.

**Acceptance Criteria:**
- [ ] `renderColumn2` returns `layout.Container{Width:width, Height:height}.Render(components.HelpOverlay(width, height))` when `m.showHelp == true`
- [ ] The early-return `if m.showHelp { return components.HelpOverlay(...) }` in `View()` is removed
- [ ] The search bar and notes list are not rendered while help is active
- [ ] Column 1 and Column 4 remain visible when help is active (manual verification)
- [ ] `CGO_ENABLED=0 go vet ./...` passes

---

### US-005: Replace Column 2 with Stats View

**Description:** As a user, I want pressing `S` to replace Column 2's search bar and
notes list with the Stats view so I can browse stats while Column 1 and Column 4 remain
visible.

**Acceptance Criteria:**
- [ ] `renderColumn2` returns `layout.Container{Width:width, Height:height}.Render(components.StatsView(m.statsView, width, height))` when `m.statsView.Active == true`
- [ ] The early-return `if m.statsView.Active { return ... }` in `View()` is removed
- [ ] The search bar and notes list are not rendered while stats view is active
- [ ] Column 1 and Column 4 remain visible when stats view is active (manual verification)
- [ ] `CGO_ENABLED=0 go vet ./...` passes

---

### US-006: Unified Esc handler to dismiss any form or overlay

**Description:** As a user, I want to press `Esc` once to exit any active form or overlay,
without needing to know which one is open.

**Acceptance Criteria:**
- [ ] A single, early Esc handler in `Update()` runs before focus-specific handlers
- [ ] Esc handler priority (checked in order):
  1. `m.confirmDiscardForm != nil` → close it and re-open the parent form (existing behaviour — preserve)
  2. `m.noteForm != nil` → trigger huh abort (existing note Esc flow — preserve discard guard)
  3. `m.tackleForm != nil` → trigger huh abort (existing tackle Esc flow — preserve discard guard)
  4. `m.showHelp == true` → set `m.showHelp = false`
  5. `m.statsView.Active == true` → set `m.statsView.Active = false`
  6. `m.focus == FocusSearch` → clear search input and return to FocusNotes (existing)
  7. Otherwise → no-op
- [ ] No other Esc handlers in the codebase handle help or stats dismissal (deduplicated)
- [ ] Pressing Esc while in FocusSearch with no form/overlay still clears search (unchanged)
- [ ] `CGO_ENABLED=0 go vet ./...` passes

---

### US-007: Activation guard for narrow terminals

**Description:** As a developer, I want form and overlay activation to be silently blocked
when the terminal is too narrow (< 61 columns) so that forms are never rendered into a
zero-width or negative-width column.

**Acceptance Criteria:**
- [ ] All form/overlay open functions check `m.width >= 61` before activating
- [ ] Affected functions: note form open, tackle form open, `m.showHelp = true`, `m.statsView.Active = true`
- [ ] If terminal is too narrow, the key press is silently ignored (no error message)
- [ ] Confirm discard guard is exempt (it can only appear if a parent form is already open, so the check was already passed)
- [ ] `CGO_ENABLED=0 go vet ./...` passes

---

### US-008: Update TUI-ARCHITECTURE.md

**Description:** As a developer, I want TUI-ARCHITECTURE.md to accurately reflect the
implemented design so it can be used as a reliable reference.

**Acceptance Criteria:**
- [ ] `ComputeColumnWidths` signature in the doc updated to include `overlayActive bool`
- [ ] "Overlay-in-Column-2 Pattern" section verified against the implemented code and corrected where needed
- [ ] `renderColumn2` conditional rendering order in the doc matches the actual implementation order
- [ ] Any references to full-screen overlay return paths removed from the doc
- [ ] Esc handler section updated to reflect the unified handler priority list
- [ ] Doc reviewed and no stale references remain

---

## Functional Requirements

- **FR-1:** `ComputeColumnWidths(termWidth int, overlayActive bool)` — second param added; when `overlayActive == true`, Column 3 is always hidden and Column 2 expands as per the architecture table
- **FR-2:** `View()` computes `overlayActive` as the union of all form and overlay active states and passes it to `ComputeColumnWidths`
- **FR-3:** `renderColumn2` uses a priority waterfall: confirmDiscardForm → noteForm → tackleForm → showHelp → statsView.Active → normal search+notes
- **FR-4:** All form/overlay content is wrapped in `layout.Container{Width: col2Width, Height: col2Height}.Render(...)` for exact bounding box
- **FR-5:** The full-screen early-return paths for forms, help, and stats in `View()` are removed
- **FR-6:** A single early Esc handler in `Update()` covers all dismiss cases before focus-specific handlers
- **FR-7:** All form and overlay open paths check `m.width >= 61` and silently ignore the key if below threshold
- **FR-8:** Existing discard-confirmation flow (Esc on a huh form with unsaved data → Confirm Discard dialog) is preserved

---

## Non-Goals

- No visual changes to the forms or overlays themselves (content unchanged)
- No new keybindings beyond the existing `?`, `S`, `N`, `T`, `Esc` keys
- No changes to Column 1 or Column 4 rendering
- No changes to the huh form constructors or their result types
- No animation or transition effects when opening/closing forms

---

## Design Considerations

- Confirm Discard must be the highest-priority check in `renderColumn2` because it
  appears on top of an existing form (note or tackle). Both `m.confirmDiscardForm` and
  `m.noteForm` (or `m.tackleForm`) may be non-nil simultaneously during the Esc flow —
  Confirm Discard wins the display slot.
- The `overlayActive` flag drives Column 3 visibility. When true, Column 3 produces an
  empty string (or is skipped in `JoinColumns`) and Column 2 gets the freed space.
- Esc on the Confirm Discard dialog should re-open the parent form (not dismiss it
  entirely), matching the existing UX intent. This is preserved in the unified handler.

---

## Technical Considerations

- `ComputeColumnWidths` is called in `View()` only; update its call-site there and update
  the function signature in `tui/layout/columns.go`
- The existing `overlayEnabled` field on `Model` is unrelated — it controls the mpv video
  overlay (OSD), not TUI form overlays. Do not confuse the two.
- `renderColumn3` should return `layout.Container{Width: width, Height: height}.Render("")`
  when `overlayActive` is true, so `JoinColumns` still receives the correct number of
  columns (avoids index-out-of-bounds in column joining logic)
- `truncateViewToWidth` helper used in the old full-screen form paths can be removed or
  left in place if still used elsewhere; audit before deleting

---

## Success Metrics

- All five form/overlay types (Note, Tackle, Confirm Discard, Help, Stats) render inside
  Column 2 with Column 1 and Column 4 visible
- `Esc` dismisses any form or overlay in a single key press
- No regression in the discard-confirmation flow
- `CGO_ENABLED=0 go vet ./...` passes clean after all changes
- `CGO_ENABLED=0 go build ./...` succeeds

---

## Open Questions

- Should `renderColumn3` return an empty container or should `JoinColumns` be told to skip
  it entirely when `overlayActive` is true? (Empty container is safer — avoids changing
  `JoinColumns` logic.)
- Is `truncateViewToWidth` used anywhere outside the old form full-screen paths? If not,
  it can be removed as part of this cleanup.
