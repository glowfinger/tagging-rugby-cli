# PRD: Column 2 Form and Overlay System

## Introduction

Currently, all forms (note input, tackle wizard, confirm discard) and overlays (help, stats view) take over the entire terminal screen via early-return rendering in `View()`. This replaces the whole TUI layout and loses context the user might need while filling out a form.

This feature changes all forms and overlays to render **inside Column 2's space** while keeping Column 1 (video status, export indicator) and Column 4 (keybinding reference) visible. Column 3 is hidden when any form or overlay is active to free up horizontal space for Column 2. This keeps the user oriented and gives them access to the control reference (Column 4) while entering data.

Esc is unified as the single key to dismiss any active form or overlay.

---

## Goals

- All forms and overlays render inside Column 2 instead of as full-screen takeovers
- Column 1 and Column 4 remain visible while a form/overlay is active; Column 3 is hidden
- Forms cannot be opened when Column 2 is not visible (terminal < 61 cols)
- Esc is the single, consistent way to exit any form or overlay
- The stats view renders in Column 2 (accepting the narrower width trade-off)
- TUI-ARCHITECTURE.md is updated to reflect the new rendering model

---

## User Stories

### US-001: Guard form/overlay activation in narrow terminal
**Description:** As a user on a narrow terminal, I want the TUI to prevent me from opening forms that cannot be displayed, rather than opening a broken or invisible form.

**Acceptance Criteria:**
- [ ] When terminal width < 61 (Column 2 is hidden by responsive layout), pressing `N`, `T`, `?`, or `S` does nothing
- [ ] No error or panic occurs in narrow mode when those keys are pressed
- [ ] All other keybindings continue to work normally in narrow mode
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

---

### US-002: Render huh forms in Column 2
**Description:** As a user, I want to fill in the note form, tackle form, or confirm-discard dialog inside Column 2 so I can still see the video status and keybinding reference while entering data.

**Acceptance Criteria:**
- [ ] When the note form is active, `renderColumn2` renders `m.form.View()` wrapped in a `layout.Container{Width: col2Width, Height: col2Height}` instead of the notes list
- [ ] When the tackle form is active, `renderColumn2` does the same
- [ ] When the confirm-discard dialog is active, `renderColumn2` renders it inside Column 2
- [ ] The huh form is constrained to Column 2's exact width and height via `layout.Container`
- [ ] Column 1 and Column 4 render normally; Column 3 is hidden (see US-004)
- [ ] Form submit and cancel (StateCompleted / StateAborted) continue to work correctly
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

---

### US-003: Render overlays (help, stats view) in Column 2
**Description:** As a user, I want the help overlay and stats view to appear inside Column 2 so the layout stays consistent and I can reference Column 4 controls while viewing them.

**Acceptance Criteria:**
- [ ] When the help overlay is active (`m.showHelp == true`), `renderColumn2` renders `HelpOverlay(col2Width, col2Height)` instead of the notes list
- [ ] When the stats view is active (`m.statsView.Active == true`), `renderColumn2` renders `StatsView(state, col2Width, col2Height)`
- [ ] `HelpOverlay` and `StatsView` are called with Column 2's dimensions (not full terminal dimensions)
- [ ] The help overlay and stats view are contained within `layout.Container{Width: col2Width, Height: col2Height}` — the Container's `↓ More...` scroll indicator appears if content overflows
- [ ] Column 1 and Column 4 render normally; Column 3 is hidden (see US-004)
- [ ] The early-return rendering blocks for help overlay and stats view are removed from `View()`
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

---

### US-004: Column layout adapts when form or overlay is active
**Description:** As a developer, I need the responsive column layout to hide Column 3 and preserve Column 4 whenever any form or overlay is active, so Column 2 has enough width to be usable.

**Acceptance Criteria:**
- [ ] A helper function or flag `overlayActive bool` is added to `ComputeColumnWidths` (or a new `ComputeColumnWidthsWithOverlay` function is introduced)
- [ ] When `overlayActive == true`:
  - `showCol3` is always `false`, regardless of terminal width
  - `showCol4` follows the normal threshold (terminal >= `Col4ShowThreshold` = 170)
  - `col2Width` is computed as: `termWidth - col1Width - (col4Width if showCol4 else 0)`
- [ ] When `overlayActive == false`, behaviour is identical to the current implementation (no regression)
- [ ] `overlayActive` is derived in `View()` by checking: `m.form != nil || m.showHelp || m.statsView.Active`
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

---

### US-005: Unify Esc as the form/overlay dismiss key
**Description:** As a user, I want to press Esc to exit any form or overlay with a single consistent key, and have Esc behave sensibly when nothing is active.

**Acceptance Criteria:**
- [ ] Esc closes the active huh form (huh already handles this via `StateAborted`) — existing cancel handling continues to work
- [ ] Esc closes the help overlay (`m.showHelp = false`) when it is active
- [ ] Esc closes the stats view (`m.statsView.Active = false`) when it is active
- [ ] When no form or overlay is active:
  - If focus is `FocusSearch`, Esc clears the search input and returns focus to `FocusNotes` (existing behaviour preserved)
  - Otherwise, Esc is a no-op
- [ ] `?` and `S` keys open help/stats only when no other form/overlay is already active
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

---

### US-006: Update TUI-ARCHITECTURE.md
**Description:** As a developer, I want the architecture document to accurately describe the new form/overlay rendering system so future contributors understand it.

**Acceptance Criteria:**
- [ ] **Rendering Pipeline** section updated: remove "Form overlays" early-return step; add note that forms and overlays render via `renderColumn2` when active
- [ ] **Column Rendering** table updated: Column 2 row describes conditional render — notes list normally, or active form/overlay
- [ ] **Layout System** section: document the `overlayActive` parameter (or new function) in `ComputeColumnWidths`
- [ ] **Keybindings** section: add Esc under a **Global Keys** subsection with the full Esc behaviour rules (form active → dismiss; FocusSearch → clear search; otherwise no-op)
- [ ] **Forms Integration** section: update Integration Pattern step 3 to note that forms now render in Column 2 at `col2Width` via `layout.Container`
- [ ] New section **Overlay-in-Column-2 Pattern** added describing: which items use it, the `overlayActive` flag, the narrow-mode guard, and the column visibility rules

---

## Functional Requirements

- **FR-1:** `ComputeColumnWidths` (or a new variant) accepts an `overlayActive bool`; when true, `showCol3 = false` always, `col2Width` expands to fill the space freed by hiding Column 3.
- **FR-2:** `renderColumn2(width, height int)` checks `m.form != nil`, `m.showHelp`, and `m.statsView.Active` at the top of the function and renders the relevant form/overlay content via `layout.Container{Width: width, Height: height}.Render(content)` instead of the normal notes list.
- **FR-3:** The early-return blocks in `View()` for forms, help overlay, and stats view are removed; `View()` always renders the full column layout.
- **FR-4:** `View()` computes `overlayActive := m.form != nil || m.showHelp || m.statsView.Active` and passes it to `ComputeColumnWidths`.
- **FR-5:** In `Update()`, Esc (`tea.KeyEsc`) is handled in a single guarded block before focus-specific handlers: if `m.showHelp`, close help; else if `m.statsView.Active`, close stats; else if `m.form != nil`, the huh form handles it via delegation; else fall through to per-focus Esc behaviour.
- **FR-6:** Before opening any form or overlay (N, T, ?, S), check that `col2Visible` (i.e. `termWidth >= 61`). If column 2 is hidden, return the model unchanged with no side effects.
- **FR-7:** `HelpOverlay(width, height int)` and `StatsView(state, width, height int)` must accept Column 2's dimensions correctly; verify their signatures match and adjust if needed.

---

## Non-Goals

- No animation or transition between normal layout and form-in-column-2
- No resizing or reflowing of huh form content within column 2 (Container truncation is acceptable)
- No scroll controls added to help overlay or stats view beyond what already exists
- No changes to Column 4 content (keybinding reference stays as-is)
- No changes to Column 1 content
- No mobile or very-narrow (< 61 col) fallback for forms — they are simply blocked

---

## Design Considerations

- **Column 4 as contextual help:** Keeping Column 4 visible while a form is open is intentional — it gives the user a keybinding reference without pressing `?`.
- **Container truncation:** Both `HelpOverlay` and `StatsView` may have more content than Column 2's height. The `layout.Container` `↓ More...` indicator handles this gracefully. No additional scrolling UX is needed for this PRD.
- **huh form width:** huh forms render at the terminal width by default. Wrapping `m.form.View()` in `layout.Container{Width: col2Width, Height: col2Height}` will constrain the output. If huh exposes a `WithWidth(int)` method, use it when creating the form to avoid huh rendering wide and being truncated.

---

## Technical Considerations

- **`ComputeColumnWidths` signature change** is the most impactful change — all call sites in `tui.go` and `columns.go` must be updated.
- **No circular imports:** `overlayActive` logic lives in `tui.go` / `columns.go`; `layout/columns.go` only needs the bool parameter.
- **huh delegation unchanged:** `Update()` still forwards all messages to `m.form` when it is non-nil. The only change is where `m.form.View()` output is placed.
- **Stats view scroll state:** `StatsViewState` already tracks `SelectedRow` and `ScrollOffset`. When rendered at column 2 width/height, these continue to function correctly.
- **Existing Esc in huh:** huh handles Esc internally and sets `form.State = huh.StateAborted`. The existing cancel-handling code in `Update()` (which sets `m.form = nil`) already covers form dismissal via Esc. US-005 ensures that help and stats also close on Esc using the same key.

---

## Success Metrics

- All five form/overlay types (note form, tackle form, confirm discard, help overlay, stats view) appear in Column 2 without layout breaks
- Column 1 and Column 4 remain fully visible and interactive while any form/overlay is open
- Esc consistently dismisses the active form/overlay from a single key press
- No regressions in narrow-mode layout (< 61 cols)
- `go vet ./...` passes with `CGO_ENABLED=0` after all changes

---

## Open Questions

- Does `huh.Form` expose a `WithWidth(int)` method? If yes, call it when constructing forms to avoid wide-render-then-truncate. If no, Container truncation is the fallback.
- Should `StatsView` keybindings (sort column, scroll) remain active when rendered in Column 2, or should it become read-only? (Current assumption: fully interactive, same as today.)
- Should the confirm-discard dialog be visually centred within Column 2, or left-aligned like other forms?
