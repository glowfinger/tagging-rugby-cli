# PRD: Notes List Clip Status Indicator

## Introduction

Prepend a one-letter clip export status indicator to the **Text** column of each row in the notes list. The indicator shows the current clip processing state (`[p]`, `[w]`, `[e]`, `[f]`) so the user can see at a glance which clips are queued, being processed, have errored, or have recently finished — without leaving the notes list.

The indicator is **transient for finished clips**: the `[f]` badge disappears 5 seconds after the clip's `finished_at` timestamp, keeping the list clean once all exports are done.

---

## Goals

- Show a colour-coded clip status indicator prepended to the Text column for each note that has a clip record.
- Suppress the indicator once a completed clip has been finished for more than 5 seconds.
- Never show an indicator when no clip record exists (e.g. notes that are not tackles, or tackles not yet queued).
- Keep the TUI-ARCHITECTURE.md up to date.

---

## User Stories

### US-001: Extend ListItem with clip status fields

**Description:** As a developer, I need the `ListItem` struct to carry clip status data so the renderer can decide whether and what to display.

**Acceptance Criteria:**
- [ ] Add `ClipStatus string` field to `components.ListItem` (values: `""`, `"pending"`, `"processing"`, `"completed"`, `"error"`)
- [ ] Add `ClipFinishedAt *time.Time` field to `components.ListItem`
- [ ] `loadNotesAndTackles()` in `tui/tui.go` extends its main SQL query with a `LEFT JOIN note_clips nc ON nc.note_id = n.id` and selects `COALESCE(nc.status, '')` and `nc.finished_at`
- [ ] The two new fields are populated for every item in the built `items` slice
- [ ] `CGO_ENABLED=0` build passes (`go build ./...`)
- [ ] `go vet ./...` passes

### US-002: Render clip status indicator in the Text column

**Description:** As a user, I want to see a one-letter clip status badge at the start of the Text column so I know which clips are pending, processing, errored, or recently finished.

**Indicator format:** `[X] ` (letter, wrapped in square brackets, followed by a space before the note text)

| Clip status | Letter | Colour |
|------------|--------|--------|
| `pending` | `p` | `styles.Lavender` |
| `processing` | `w` | `styles.Amber` |
| `error` | `e` | `styles.Red` |
| `completed` (≤ 5 s ago) | `f` | `styles.Green` |
| `completed` (> 5 s ago) or no clip | *(none)* | — |

**Rendering rules:**
- The indicator is computed in `renderTableRow` in `tui/components/noteslist.go`
- `renderTableRow` receives `clipStatus string` and `clipFinishedAt *time.Time` (passed through from `ListItem`)
- The `[X] ` prefix is prepended to the `text` variable before width truncation so the full text field including the badge is bounded by `textWidth`
- The badge letters are individually styled with their colour (inline ANSI); the surrounding `[`, `]`, and space use the row's `baseStyle`
- `time.Now()` is called once per `NotesList` render and passed down to `renderTableRow` to evaluate the 5-second window

**Acceptance Criteria:**
- [ ] `[p] ` prefix rendered in Lavender for `pending` clips
- [ ] `[w] ` prefix rendered in Amber for `processing` clips
- [ ] `[e] ` prefix rendered in Red for `error` clips
- [ ] `[f] ` prefix rendered in Green for `completed` clips whose `finished_at` is ≤ 5 seconds before `time.Now()`
- [ ] No indicator prefix when `finished_at` is > 5 seconds ago or `ClipStatus == ""`
- [ ] Text field (including badge) is still bounded to `textWidth` (truncated with `...` if needed)
- [ ] Highlight/search substring matching still works on the text portion after the badge prefix
- [ ] `CGO_ENABLED=0` build passes
- [ ] `go vet ./...` passes

### US-003: Update TUI-ARCHITECTURE.md

**Description:** As a developer, I need the architecture document to reflect the new `ListItem` fields and indicator rendering behaviour so future contributors understand the design.

**Acceptance Criteria:**
- [ ] `NotesList` section updated to document `ClipStatus` and `ClipFinishedAt` fields on `ListItem`
- [ ] Documents the four indicator letters, their colours, and the 5-second hide rule for `[f]`
- [ ] Notes that `loadNotesAndTackles()` populates these fields via LEFT JOIN on `note_clips`

---

## Functional Requirements

- **FR-1:** `components.ListItem` gains two fields: `ClipStatus string` and `ClipFinishedAt *time.Time`.
- **FR-2:** `loadNotesAndTackles()` extends its inline SQL to include `LEFT JOIN note_clips nc ON nc.note_id = n.id` and scans `COALESCE(nc.status, '')` and `nc.finished_at` into the new fields.
- **FR-3:** `NotesList()` passes `time.Now()` into `renderTableRow` (or computes it once at the top of the function and passes it to each row render call).
- **FR-4:** `renderTableRow` builds an optional badge prefix (`[X] `) based on `ClipStatus` and `ClipFinishedAt`, rendering each badge letter in its designated colour using an inline `lipgloss.NewStyle().Foreground(...)` style.
- **FR-5:** The badge prefix is prepended to the raw `text` string before the existing `textWidth` truncation step, so the combined string (badge + text) is always bounded.
- **FR-6:** No indicator is shown when `ClipStatus == ""` (note has no clip record) or when `ClipStatus == "completed"` and `time.Now().Sub(*ClipFinishedAt) > 5*time.Second`.

---

## Non-Goals

- No indicator for non-tackle notes that will never have a clip (indicator simply not shown — no special label).
- No click/keyboard interaction on the badge itself.
- No polling rate change — the existing ~250 ms tick via `refreshExportProgress()` is sufficient; `loadNotesAndTackles()` is already called after state changes.
- No change to the ExportIndicator box in Column 1 — that remains as-is.
- No persistent storage of the "shown for 5 s" state — purely time-based at render time.

---

## Technical Considerations

- The `loadNotesAndTackles()` inline SQL already uses `LEFT JOIN note_timing`. Adding `LEFT JOIN note_clips` to the same query is straightforward. Note: there is at most one `note_clips` row per note (enforced by `upsert_note_clip_pending.sql`), so no row multiplication occurs.
- `clipFinishedAt` must be scanned as `*time.Time` (nullable column). Use `var finishedAt sql.NullTime` then set `item.ClipFinishedAt` from `finishedAt.Time` if `finishedAt.Valid`.
- `renderTableRow` signature changes: add `clipStatus string, clipFinishedAt *time.Time, now time.Time` parameters. All call sites are inside `NotesList()` — no external callers.
- The `NotesList` function signature itself does **not** change — `time.Now()` is called inside the function, not passed from the caller.
- Colour references live in `tui/styles/styles.go`: `Lavender`, `Amber`, `Red`, `Green`.

---

## Success Metrics

- A pending/processing/errored clip shows its indicator immediately in the notes list without any extra keypress.
- The `[f]` badge disappears within one render cycle after the 5-second window elapses (~250 ms resolution from the tick handler).
- No layout shift or column misalignment — the text field is correctly truncated to `textWidth`.

---

## Open Questions

- Should the indicator bracket characters (`[`, `]`) also be coloured, or use the row's default foreground? (Suggested: use `baseStyle` for brackets, coloured only for the letter.)
- If a note has multiple clip records (edge case not expected by current schema), should the worst-status rule apply (`error > processing > pending > completed`)? Current assumption: single clip per note.
