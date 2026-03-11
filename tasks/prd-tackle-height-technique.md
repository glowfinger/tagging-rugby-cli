# PRD: Add Height and Technique to Tackle

## Introduction

Add two optional fields — `height` and `technique` — to the `note_tackles` database table and expose them in the tackle wizard form. Both fields are optional at every layer (DB, form, and save logic). They appear on Step 2 of both the "Add Tackle" and "Edit Tackle" huh forms. The database migration modifies `001_create_videos_table.sql` directly since all data is currently test data.

## Goals

- Add `height TEXT` and `technique TEXT` (nullable) columns to `note_tackles` in the canonical schema migration
- Surface both fields on Step 2 ("Optional Details") of the tackle wizard (both add and edit flows)
- Propagate values through db models, SQL queries, db functions, form result types, and tui.go save functions
- Update `TUI-ARCHITECTURE.md` to reflect the new fields
- Zero breaking changes to existing note data — both fields default to NULL / empty string

## User Stories

### US-001: Update canonical DB schema
**Description:** As a developer, I need the `note_tackles` table to include `height` and `technique` columns so the data can be persisted.

**Acceptance Criteria:**
- [ ] `db/sql/migrations/001_create_videos_table.sql` — `note_tackles` table has `height TEXT` and `technique TEXT` columns (both without NOT NULL, no default value)
- [ ] `db/sql/insert_note_tackle.sql` — INSERT includes `height` and `technique` placeholders: `INSERT INTO note_tackles (note_id, player, attempt, outcome, height, technique) VALUES (?, ?, ?, ?, ?, ?);`
- [ ] `db/sql/select_note_tackles_by_note.sql` — SELECT includes `height, technique` in the column list
- [ ] `db/models.go` — `NoteTackle` struct has `Height string` and `Technique string` fields
- [ ] `db/models.go` — `EditTackleData` struct has `Height string` and `Technique string` fields
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

### US-002: Update db functions
**Description:** As a developer, I need all database functions that read or write `note_tackles` to handle the new columns.

**Acceptance Criteria:**
- [ ] `db/functions.go` — `InsertNoteTackle(db, noteID, player, attempt, outcome, height, technique string)` signature extended with `height, technique string` parameters
- [ ] `db/functions.go` — `SelectNoteTacklesByNote()` scan includes `&t.Height, &t.Technique`
- [ ] `db/functions.go` — `InsertNoteWithChildren()` tackle loop passes `t.Height, t.Technique` to `InsertNoteTackleSQL`
- [ ] `db/functions.go` — `UpdateNoteWithChildren()` tackle loop passes `t.Height, t.Technique` to `InsertNoteTackleSQL`
- [ ] `db/functions.go` — `LoadNoteForEdit()` populates `data.Height` and `data.Technique` from the first loaded tackle
- [ ] Typecheck/vet passes

### US-003: Add Height and Technique to tackle form types and UI
**Description:** As a user, I want to optionally record the height and technique of a tackle when adding or editing one, so that analysis can be more detailed.

**Acceptance Criteria:**
- [ ] `tui/forms/tackleform.go` — `TackleFormResult` struct has `Height string` and `Technique string` fields
- [ ] `tui/forms/tackleform.go` — `NewTackleForm()` Step 2 group includes:
  - `huh.NewSelect[string]()` for `Height` with options: `""` (blank/unset), `"high"`, `"mid"`, `"low"` — title "Height", description "Optional"
  - `huh.NewInput()` for `Technique` — title "Technique", description "Optional - e.g. shoulder, choke, ankle"
- [ ] `tui/forms/tackleform.go` — `NewEditTackleForm()` Step 2 group includes the same Height select and Technique input, bound to `result.Height` and `result.Technique`
- [ ] Height select is pre-selected to `""` (blank) by default so the form passes without input
- [ ] Typecheck/vet passes

### US-004: Wire tui.go save functions
**Description:** As a developer, I need the TUI save functions to pass the new form values through to the database.

**Acceptance Criteria:**
- [ ] `tui/tui.go` — `saveTackleFromForm()`: `db.NoteTackle` literal includes `Height: result.Height, Technique: result.Technique`
- [ ] `tui/tui.go` — `saveEditTackleFromForm()`: same — `db.NoteTackle` literal includes `Height: result.Height, Technique: result.Technique`
- [ ] `tui/tui.go` — `editTackleFormResult` mapping (where `db.EditTackleData` is mapped to `forms.EditTackleFormResult`) includes `Height: data.Height, Technique: data.Technique`
- [ ] Typecheck/vet passes

### US-005: Update TUI-ARCHITECTURE.md
**Description:** As a developer, I need the architecture documentation to reflect the new tackle fields so the next developer understands what is stored.

**Acceptance Criteria:**
- [ ] `tui/TUI-ARCHITECTURE.md` — `TackleFormResult` table / description updated to include `Height` and `Technique` fields with their descriptions
- [ ] `tui/TUI-ARCHITECTURE.md` — `note_tackles` schema reference (if present) updated, or a note added under the Forms Integration section documenting the new fields

## Functional Requirements

- FR-1: `note_tackles` table MUST have `height TEXT` and `technique TEXT` columns, both nullable with no default
- FR-2: Both columns MUST be included in the INSERT and SELECT SQL for `note_tackles`
- FR-3: `NoteTackle` and `EditTackleData` Go structs MUST expose `Height` and `Technique` as `string` fields (empty string = not set)
- FR-4: `InsertNoteTackle()` MUST accept height and technique and pass them to the SQL
- FR-5: `SelectNoteTacklesByNote()` MUST scan height and technique into the struct
- FR-6: `LoadNoteForEdit()` MUST populate height and technique from the DB into `EditTackleData`
- FR-7: Both fields MUST appear on Step 2 of `NewTackleForm()` and `NewEditTackleForm()`, after the existing Followed/Notes/Zone inputs and before Star
- FR-8: Height MUST be a select with options: `""` (none), `"high"`, `"mid"`, `"low"`
- FR-9: Technique MUST be a free-text input
- FR-10: Neither field may block form submission when empty (no validation required)

## Non-Goals

- No display of height/technique in the notes list (Column 2 table) at this time
- No filtering or sorting by height/technique
- No stats breakdown by height/technique
- No clip filename changes based on height/technique
- No database migration file (001 is modified directly since data is test-only)

## Technical Considerations

- The `InsertNoteTackleSQL` embed and `InsertNoteTackle()` function are used in two places in `InsertNoteWithChildren` and `UpdateNoteWithChildren` — both must be updated
- The `EditTackleFormResult` embeds `TackleFormResult`, so adding fields to `TackleFormResult` automatically adds them to `EditTackleFormResult` — no separate struct change needed beyond the `TackleFormResult` addition
- Height select needs a blank option (`huh.NewOption("None", "")`) as the first option so it is valid/safe to submit without selection
- When scanning `SelectNoteTacklesByNote`, SQLite NULL will scan into an empty Go string via `database/sql` — no pointer needed

## Success Metrics

- Both fields are stored in the DB when a tackle is saved with them populated
- Both fields are blank/empty by default — existing workflow is unchanged
- Edit form pre-fills height and technique from the DB correctly
- `CGO_ENABLED=0 go vet ./...` passes with no errors

## Open Questions

- Should Height and Technique be visible in the selected note detail view in Column 1? (deferred — not in scope here)
- Should the Height select have a "None" label vs a blank label? (implementation detail — either works)
