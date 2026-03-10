# PRD: Add Techniques Field to Tackle

## Introduction

Add an optional `techniques` string field to the `note_tackles` table to capture tackle technique quality (e.g. "ok", "good", "poor"). The field is displayed and editable on Step 2 of the tackle form, positioned above the existing Notes field, defaulting to `"ok"` in the add form. Since all data is test data, the migration modifies the canonical `001_create_videos_table.sql` rather than adding a separate migration file.

## Goals

- Store tackle technique alongside player, attempt, outcome, and height in `note_tackles`
- Surface techniques as an optional input on Step 2 of both the add and edit tackle forms, above Notes
- Default to `"ok"` in the add form (same pattern as height)

## User Stories

### US-001: Add techniques column to note_tackles table
**Description:** As a developer, I need `techniques TEXT` added to `note_tackles` in the canonical migration so the schema is correct for fresh databases.

**Acceptance Criteria:**
- [ ] `001_create_videos_table.sql` — `note_tackles` table definition includes `techniques TEXT` (nullable, no default) after the `height` column
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

### US-002: Update Go model and DB functions to include techniques
**Description:** As a developer, I need all Go DB layer code to read and write the `techniques` field so it is persisted correctly.

**Acceptance Criteria:**
- [ ] `db/models.go` — `NoteTackle` struct gains `Techniques string` field
- [ ] `db/sql/insert_note_tackle.sql` — INSERT includes `techniques` as 6th parameter: `INSERT INTO note_tackles (note_id, player, attempt, outcome, height, techniques) VALUES (?, ?, ?, ?, ?, ?);`
- [ ] `db/sql/select_note_tackles_by_note.sql` — SELECT includes `techniques` column
- [ ] `db/functions.go` — `InsertNoteTackle` signature gains `techniques string` parameter and passes it to Exec
- [ ] `db/functions.go` — `SelectNoteTacklesByNote` Scan includes `&t.Techniques`
- [ ] `db/functions.go` — `InsertNoteWithChildren` loop passes `t.Techniques` to `InsertNoteTackleSQL` exec call
- [ ] `db/functions.go` — `UpdateNoteWithChildren` loop passes `t.Techniques` to `InsertNoteTackleSQL` exec call
- [ ] `db/functions.go` — `EditTackleData` struct gains `Techniques string` field
- [ ] `db/functions.go` — `LoadNoteForEdit` populates `data.Techniques` from `tackles[0].Techniques`
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

### US-003: Add techniques field to TackleFormResult and form
**Description:** As a developer, I need the form result struct and both form constructors to carry the techniques value.

**Acceptance Criteria:**
- [ ] `tui/forms/tackleform.go` — `TackleFormResult` gains `Techniques string` field
- [ ] `tui/forms/tackleform.go` — `NewTackleForm` sets `result.Techniques = "ok"` if empty before building the form (same pattern as Height)
- [ ] `tui/forms/tackleform.go` — `NewTackleForm` Step 2 group includes a `huh.NewInput()` for `Techniques`, positioned above the existing `Notes` input, with title "Techniques" and description "Optional - tackle technique (e.g. ok, good, poor)"
- [ ] `tui/forms/tackleform.go` — `NewEditTackleForm` Step 2 group includes the same Techniques input pre-bound to `result.Techniques`, positioned above Notes
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

### US-004: Wire techniques through TUI submit handlers
**Description:** As a developer, I need all call sites that submit tackle data to pass the `techniques` value so it is saved to the database.

**Acceptance Criteria:**
- [ ] `saveTackleFromForm` in `tui/tui.go` — `db.NoteTackle{}` literal includes `Techniques: result.Techniques`
- [ ] `saveEditTackleFromForm` in `tui/tui.go` — `db.NoteTackle{}` literal includes `Techniques: result.Techniques`
- [ ] Edit form pre-fill in `tui/tui.go` — `TackleFormResult` initialisation includes `Techniques: data.Techniques`
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

## Functional Requirements

- FR-1: `note_tackles` table has a nullable `techniques TEXT` column with no default, after `height`
- FR-2: All INSERT statements for `note_tackles` bind the techniques value (empty string = empty in DB)
- FR-3: All SELECT statements for `note_tackles` return `techniques`
- FR-4: Step 2 of the add tackle form shows a "Techniques" input above "Notes", pre-filled with `"ok"`
- FR-5: Step 2 of the edit tackle form shows a "Techniques" input above "Notes", pre-filled from the DB value
- FR-6: Techniques is optional — form submits successfully when left blank

## Non-Goals

- No validation or enumeration of techniques values (free-text string)
- No display of techniques in the notes list or column overlays
- No data migration of existing rows (test data only)
- No filtering or sorting by techniques

## Technical Considerations

- Modify `001_create_videos_table.sql` in-place (test data, no separate migration needed)
- Add `techniques` after `height` in the column order for consistency
- `InsertNoteTackle` arity changes from 5 params to 6 — all call sites must be updated together
- Use `CGO_ENABLED=0` for all build/vet commands (modernc.org/sqlite requirement)
- The `addTackle` legacy command path (`tui/tui.go` and `cmd/tackle.go`) has no techniques source — Height zero value is already accepted there, same applies to Techniques

## Success Metrics

- Fresh database from `001_create_videos_table.sql` has the `techniques` column in `note_tackles`
- A tackle saved with a techniques value round-trips correctly through insert → select → edit form pre-fill
- A tackle saved without techniques (blank) works without error
- Add form opens with "ok" pre-filled in the Techniques field

## Open Questions

- None — requirements are fully specified.
