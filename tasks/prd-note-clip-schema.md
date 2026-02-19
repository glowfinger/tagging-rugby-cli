# PRD: note_clips Schema Redesign

## Introduction

The `note_clips` table currently stores clip metadata using `name`, `duration`, and `error` fields that don't reflect how clips are actually produced. This PRD replaces that schema with fields that model the clip as a file on disk (`folder`, `filename`, `extension`, `format`, `filesize`) plus a lifecycle (`status`, `log`) alongside the existing timestamps. All existing data is test data and can be wiped via migration.

## Goals

- Replace the `note_clips` schema to match real clip file metadata.
- Provide a `status` plain-text field to track clip lifecycle (e.g. `"pending"`, `"processing"`, `"done"`, `"error"`).
- Provide a `log` plain-text field to carry a single error or info message.
- Keep all Go types, SQL queries, and db functions consistent with the new schema.
- Leave no dead code or orphaned references to the old columns (`name`, `duration`, `error`).

## User Stories

### US-001: Update migration to new note_clips schema
**Description:** As a developer, I need the canonical schema in `001_create_videos_table.sql` to reflect the new `note_clips` columns so that fresh database initialisation produces the correct table.

**Acceptance Criteria:**
- [ ] `note_clips` definition in `001_create_videos_table.sql` has exactly these columns: `id`, `note_id`, `folder`, `filename`, `extension`, `format`, `filesize`, `status`, `started_at`, `finished_at`, `error_at`, `log`
- [ ] Old columns `name`, `duration`, `error` are removed
- [ ] `note_id` still carries `NOT NULL REFERENCES notes(id) ON DELETE CASCADE`
- [ ] All other tables in the file are unchanged
- [ ] `CGO_ENABLED=0 go vet ./...` passes

### US-002: Update NoteClip Go struct
**Description:** As a developer, I need the `NoteClip` struct in `db/models.go` to reflect the new columns so that Go code that reads or writes clips compiles and is type-safe.

**Acceptance Criteria:**
- [ ] `NoteClip` struct has fields: `ID int64`, `NoteID int64`, `Folder string`, `Filename string`, `Extension string`, `Format string`, `Filesize int64`, `Status string`, `StartedAt *time.Time`, `FinishedAt *time.Time`, `ErrorAt *time.Time`, `Log string`
- [ ] Fields `Name string`, `Duration float64`, `Error string` are removed
- [ ] `CGO_ENABLED=0 go vet ./...` passes

### US-003: Update insert_note_clip.sql and select_note_clips_by_note.sql
**Description:** As a developer, I need the existing SQL files for inserting and selecting note clips to use the new column set so that db functions produce correct SQL at runtime.

**Acceptance Criteria:**
- [ ] `db/sql/insert_note_clip.sql` inserts `(note_id, folder, filename, extension, format, filesize, status, started_at, finished_at, error_at, log)`
- [ ] `db/sql/select_note_clips_by_note.sql` selects those same columns for a given `note_id`
- [ ] Old columns `name`, `duration`, `error` are absent from both files
- [ ] `CGO_ENABLED=0 go vet ./...` passes

### US-004: Add update_note_clip.sql and select_note_clip_by_id.sql
**Description:** As a developer, I need SQL files to update a clip's status/timestamps/log and to fetch a single clip by ID so that the export process can track progress over the clip lifecycle.

**Acceptance Criteria:**
- [ ] `db/sql/update_note_clip.sql` updates `status`, `started_at`, `finished_at`, `error_at`, `log` by `id` (e.g. `UPDATE note_clips SET status=?, started_at=?, finished_at=?, error_at=?, log=? WHERE id=?`)
- [ ] `db/sql/select_note_clip_by_id.sql` selects all columns for a single clip by `id`
- [ ] Both files are embedded in `db/queries.go` as `UpdateNoteClipSQL` and `SelectNoteClipByIDSQL`
- [ ] `CGO_ENABLED=0 go vet ./...` passes

### US-005: Update db/functions.go for new schema
**Description:** As a developer, I need all db functions that touch `note_clips` to be updated so that they compile against the new struct and SQL, and two new helper functions (`UpdateNoteClip`, `SelectNoteClipByID`) are available for the export pipeline.

**Acceptance Criteria:**
- [ ] `InsertNoteClip(db, noteID, folder, filename, extension, format string, filesize int64, status string, startedAt, finishedAt, errorAt interface{}, log string) error` — new signature, executes `InsertNoteClipSQL`
- [ ] The clip loop inside `InsertNoteWithChildren` passes the new struct fields to `InsertNoteClipSQL` (matching the new column order)
- [ ] `SelectNoteClipsByNote` scans into the updated `NoteClip` fields (no `Name`, `Duration`, `Error`)
- [ ] `SelectNoteClipByID(db *sql.DB, id int64) (*NoteClip, error)` — new function, executes `SelectNoteClipByIDSQL`
- [ ] `UpdateNoteClip(db *sql.DB, id int64, status string, startedAt, finishedAt, errorAt interface{}, log string) error` — new function, executes `UpdateNoteClipSQL`
- [ ] `CGO_ENABLED=0 go vet ./...` passes

## Functional Requirements

- FR-1: The `note_clips` table must have exactly these columns: `id` (INTEGER PK), `note_id` (INTEGER NOT NULL FK → notes), `folder` (TEXT), `filename` (TEXT), `extension` (TEXT), `format` (TEXT), `filesize` (INTEGER), `status` (TEXT), `started_at` (DATETIME), `finished_at` (DATETIME), `error_at` (DATETIME), `log` (TEXT).
- FR-2: `status` is plain TEXT with no database-level constraint; valid values are managed at the application layer.
- FR-3: `log` stores a single plain-text message (error reason or info string); it replaces the old `error` column.
- FR-4: `folder` stores the absolute filesystem path to the directory containing the clip file.
- FR-5: A clip can be updated in place (status, timestamps, log) via `UpdateNoteClip` without deleting and re-inserting it.
- FR-6: A clip can be fetched individually by its `id` via `SelectNoteClipByID`.

## Non-Goals

- No TUI changes — clip display in the UI is out of scope for this PRD.
- No data migration script — existing test data will be dropped when the database is re-initialised from the updated migration.
- No validation of `status` values at the database or Go layer.
- No removal of the `note_clips` cascade delete behaviour — it must stay.

## Technical Considerations

- The migration file `001_create_videos_table.sql` is the single source of truth for the schema. It is applied via `db/migrations.go` on startup.
- All SQL is embedded with `//go:embed` in `db/queries.go`; any new `.sql` file needs a corresponding embed var there.
- Build requirement: `CGO_ENABLED=0` (modernc.org/sqlite).
- Go binary path: `/home/node/go/bin/go`; ensure `PATH` includes it when running vet/build.
- `NoteChildren.Clips []NoteClip` is used in `InsertNoteWithChildren` — the struct fields must stay consistent with the insert SQL parameter order.

## Success Metrics

- `CGO_ENABLED=0 go vet ./...` exits 0 after all changes.
- A fresh database initialised from `001_create_videos_table.sql` has the correct `note_clips` schema (verifiable with `.schema note_clips` in sqlite3).
- No references to `Name`, `Duration`, or `Error` (old fields) remain anywhere in the `db/` package.

## Open Questions

- Should `filesize` be nullable (`*int64`) to handle clips not yet written to disk? Currently specified as non-pointer `int64` (zero when unknown). Confirm before implementing US-002/US-005.
- Should `status` default to `"pending"` in the SQL schema (`DEFAULT 'pending'`)? Keeping it explicit makes the insert intent clear without relying on defaults.
