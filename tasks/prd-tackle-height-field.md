# PRD: Add Height Field to Tackle

## Introduction

Add an optional `height` string field to the `note_tackles` table to capture tackle height (e.g., "high", "mid", "low"). The field is displayed and editable on Step 2 of the tackle form, positioned above the existing Notes field. Since all data is test data, the migration modifies the canonical `001_create_videos_table.sql` rather than adding a separate migration file.

## Goals

- Store tackle height alongside player, attempt, and outcome in `note_tackles`
- Surface height as an optional input on Step 2 of both the add and edit tackle forms
- Keep all downstream read paths (select queries, Go structs, display) consistent with the new column

## User Stories

### US-001: Add height column to note_tackles table
**Description:** As a developer, I need `height TEXT` added to `note_tackles` in the canonical migration so the schema is correct for fresh databases.

**Acceptance Criteria:**
- [ ] `001_create_videos_table.sql` — `note_tackles` table definition includes `height TEXT` (nullable, no default)
- [ ] The column appears after `outcome` in the CREATE TABLE statement
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

### US-002: Update Go model and DB functions to include height
**Description:** As a developer, I need all Go DB layer code to read and write the `height` field so it is persisted correctly.

**Acceptance Criteria:**
- [ ] `db/models.go` — `NoteTackle` struct gains `Height string` field
- [ ] `db/sql/insert_note_tackle.sql` — INSERT includes `height` as 5th parameter: `INSERT INTO note_tackles (note_id, player, attempt, outcome, height) VALUES (?, ?, ?, ?, ?);`
- [ ] `db/sql/select_note_tackles_by_note.sql` — SELECT includes `height` column
- [ ] `db/functions.go` — `InsertNoteTackle` signature gains `height string` parameter and passes it to Exec
- [ ] `db/functions.go` — `SelectNoteTacklesByNote` Scan includes `&t.Height`
- [ ] `db/functions.go` — `InsertNoteWithChildren` loop passes `t.Height` to `InsertNoteTackleSQL` exec call
- [ ] `db/functions.go` — `UpdateNoteWithChildren` loop passes `t.Height` to `InsertNoteTackleSQL` exec call
- [ ] `db/functions.go` — `EditTackleData` struct gains `Height string` field
- [ ] `db/functions.go` — `LoadNoteForEdit` populates `data.Height` from `tackles[0].Height`
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

### US-003: Add height field to TackleFormResult and form structs
**Description:** As a developer, I need the form result structs to carry the height value so the TUI can pass it through to the DB layer.

**Acceptance Criteria:**
- [ ] `tui/forms/tackleform.go` — `TackleFormResult` gains `Height string` field
- [ ] `tui/forms/tackleform.go` — `NewTackleForm` Step 2 group includes a `huh.NewInput()` for `Height`, positioned above the existing `Notes` input, with title "Height" and description "Optional - tackle height (e.g. high, mid, low)"
- [ ] `tui/forms/tackleform.go` — `NewEditTackleForm` Step 2 group includes the same Height input pre-bound to `result.Height`, positioned above Notes
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

### US-004: Wire height through TUI submit handlers
**Description:** As a developer, I need all call sites that submit tackle data to pass the `height` value so it is saved to the database.

**Acceptance Criteria:**
- [ ] Every call to `db.InsertNoteTackle(...)` in `tui/` passes `result.Height` as the height argument
- [ ] Every `db.NoteTackle{...}` literal in `tui/` that populates `NoteChildren.Tackles` includes `Height: result.Height`
- [ ] Edit flow: `EditTackleData.Height` is used to pre-fill `result.Height` before the edit form is opened (mirrors existing Followed/Notes/Zone pre-fill pattern)
- [ ] `go vet ./...` passes with `CGO_ENABLED=0`

## Functional Requirements

- FR-1: `note_tackles` table has a nullable `height TEXT` column with no default
- FR-2: All INSERT statements for `note_tackles` bind the height value (empty string maps to empty string in DB, not NULL)
- FR-3: All SELECT statements for `note_tackles` return `height`
- FR-4: Step 2 of the tackle form (both add and edit) shows a "Height" input above the "Notes" input
- FR-5: Height is optional — form submits successfully when height is left blank

## Non-Goals

- No validation or enumeration of height values (free-text string)
- No display of height in the notes list or column 2 overlay (display changes are out of scope)
- No data migration of existing rows (test data only — fresh DB from migration is sufficient)
- No filtering or sorting by height

## Technical Considerations

- Because this is test data, modify `001_create_videos_table.sql` in-place rather than adding `002_*.sql`
- The `InsertNoteTackleSQL` is embedded from `db/sql/insert_note_tackle.sql` — updating the `.sql` file is sufficient; no Go embed changes needed
- All Go DB call sites for tackles must be updated together to avoid compile errors (arity mismatch on `InsertNoteTackle`)
- Use `CGO_ENABLED=0` for all build/vet commands (modernc.org/sqlite requirement)

## Success Metrics

- Fresh database created from `001_create_videos_table.sql` has the `height` column in `note_tackles`
- A tackle saved with a height value round-trips correctly through insert → select → edit form pre-fill
- A tackle saved without a height value (blank) works without error

## Open Questions

- None — requirements are fully specified.
