# PRD: Edit and Delete Tackles

## Introduction

Add the ability to edit and delete existing tackle entries from the TUI notes list. Currently, tackles can only be created — once saved, there is no way to fix mistakes or remove incorrect entries without using the CLI `note delete` command. This feature adds `e` (edit) and `x` (delete) keybindings to the notes list so users can manage tackles directly in the TUI.

## Goals

- Allow users to delete a tackle entry instantly with a single keypress (`x`) from the notes list
- Allow users to edit all fields of a tackle entry via a pre-filled huh form (`e`) from the notes list
- Support editing the start timestamp and setting an end timestamp (with a default of +2 seconds)
- Reuse existing form and database patterns to keep the implementation consistent

## User Stories

### US-001: Add update SQL and DB functions for note children
**Description:** As a developer, I need database functions to update existing tackle records and their related child tables so the edit feature has a persistence layer.

**Acceptance Criteria:**
- [ ] Add `UpdateNoteWithChildren(db, noteID, children)` function that updates all child tables in a transaction
- [ ] The function deletes existing child rows (details, zones, highlights, tackles) and re-inserts from the provided `NoteChildren` struct — this is simpler than diffing individual rows since each note has at most one of each child type
- [ ] Add `UpdateNoteTiming(db, noteID, start, end)` function to update the timing record for a note
- [ ] Add corresponding SQL files: `update_note_timing.sql`, etc.
- [ ] Typecheck/vet passes: `CGO_ENABLED=0 go vet ./...`

### US-002: Delete tackle from TUI notes list
**Description:** As a user, I want to press `x` on a selected tackle in the notes list to immediately delete it so I can quickly remove incorrect entries.

**Acceptance Criteria:**
- [ ] Pressing `x` when a tackle or note is selected in the notes list deletes it immediately (no confirmation dialog)
- [ ] Deletion calls existing `db.DeleteNote()` which cascades to all child tables
- [ ] After deletion, the notes list reloads and stats panel refreshes
- [ ] A confirmation message appears in the command bar (e.g., "Deleted tackle 42")
- [ ] If the list is empty after deletion, selection resets gracefully (no index out of bounds)
- [ ] Typecheck/vet passes

### US-003: Add edit tackle form variant
**Description:** As a developer, I need a tackle form that can be pre-populated with existing data for editing, including a timestamp/end field.

**Acceptance Criteria:**
- [ ] Add `NewEditTackleForm(timestamp, endSeconds, result)` to `tui/forms/tackleform.go` that creates a form pre-filled with values from `TackleFormResult`
- [ ] The form header shows "Edit Tackle @ MM:SS" instead of "Add Tackle @ MM:SS"
- [ ] Add a "Timestamp" field (editable, float64 displayed as MM:SS.s) to step 1 of the form
- [ ] Add an "End (seconds)" field to step 1 — this sets how many seconds after the start timestamp the end should be. Default value is `2` (representing the existing `end` column in `note_timing`)
- [ ] Validation: end seconds must be a positive number
- [ ] All existing fields (player, attempt, outcome, followed, notes, zone, star) remain and are pre-populated
- [ ] Typecheck/vet passes

### US-004: Load existing tackle data for editing
**Description:** As a developer, I need to fetch all child records for a note so the edit form can be pre-populated.

**Acceptance Criteria:**
- [ ] Add a `LoadNoteForEdit(db, noteID)` function (or similar) that returns `TackleFormResult` + timestamp + end seconds by querying `note_tackles`, `note_timing`, `note_details`, `note_zones`, `note_highlights` for the given note ID
- [ ] The end seconds value is calculated as `timing.End - timing.Start` (falls back to 2.0 if no timing or if start == end)
- [ ] Handles missing optional fields gracefully (empty strings for followed/notes/zone, false for star)
- [ ] Typecheck/vet passes

### US-005: Wire up edit flow in TUI
**Description:** As a user, I want to press `e` on a selected tackle in the notes list to open an edit form pre-filled with the existing data so I can fix mistakes.

**Acceptance Criteria:**
- [ ] Pressing `e` when a tackle is selected opens the edit tackle form
- [ ] The form is pre-populated with all existing data (player, attempt, outcome, followed, notes, zone, star, timestamp, end seconds)
- [ ] On submit, the existing note's child records are updated via `UpdateNoteWithChildren` (the parent note row itself doesn't change)
- [ ] The timing record is updated with the new start timestamp and calculated end (`start + endSeconds`)
- [ ] After save, the notes list reloads and stats panel refreshes
- [ ] A confirmation message appears (e.g., "Updated tackle 42")
- [ ] Pressing Esc on the edit form with changes triggers the existing confirm-discard dialog; without changes closes immediately
- [ ] Pressing `e` on a **note** (non-tackle) item does nothing or shows "Edit not supported for notes" message
- [ ] Typecheck/vet passes

### US-006: Update help overlay with new keybindings
**Description:** As a user, I want to see the new `e` and `x` keybindings in the help overlay so I know they exist.

**Acceptance Criteria:**
- [ ] Help overlay (`?`) shows `x` — Delete selected item
- [ ] Help overlay (`?`) shows `e` — Edit selected tackle
- [ ] Typecheck/vet passes

## Functional Requirements

- FR-1: Pressing `x` in normal mode with a selected list item deletes that note (and all children via cascade) immediately, with no confirmation dialog
- FR-2: Pressing `e` in normal mode with a selected tackle opens the edit form pre-populated with existing data
- FR-3: The edit form includes all fields from the create form plus an editable start timestamp and an "end seconds" field (default 2)
- FR-4: The "end seconds" field controls how many seconds after the start the `note_timing.end` value is set to (i.e., `end = start + endSeconds`)
- FR-5: On edit form submit, all child tables are updated in a single transaction (delete old children, insert new ones)
- FR-6: The timing record is updated separately with new start and calculated end values
- FR-7: After any edit or delete, the notes list and stats panel refresh
- FR-8: The edit form reuses the existing confirm-discard flow when Esc is pressed with unsaved changes

## Non-Goals

- No bulk delete (select multiple items and delete at once)
- No undo/redo for deletions
- No editing of note-type items (only tackles)
- No editing the video association or note category
- No CLI commands for edit (TUI only for now)

## Technical Considerations

- The update strategy for child tables is "delete all + re-insert" within a transaction — this is simpler than diffing and the data volumes are tiny (1 row per child table per note)
- The existing `DeleteNote` function with CASCADE already handles deletion correctly
- The `ListItem` struct in `noteslist.go` already carries the note `ID` and `Type`, which is sufficient to determine if edit is applicable and to load data
- The existing `tackleFormResult` / `tackleFormTimestamp` fields on Model can be reused; an additional field like `editingNoteID int64` (0 = create mode, >0 = edit mode) can distinguish create vs edit
- The end seconds field is new — currently `saveTackleFromForm` sets `End: timestamp` (same as start). The edit form introduces the concept of a duration offset

## Success Metrics

- Users can delete a tackle in 1 keypress
- Users can edit a tackle and save in under 10 seconds
- No data loss — edited tackles retain their note ID and video association

## Open Questions

- Should `x` also work for note-type items, or only tackles? (Current spec: works for all note types since `DeleteNote` handles any category)
- Should the timestamp field in the edit form accept raw seconds or MM:SS format? (Current spec: display as MM:SS.s but store as float64)
