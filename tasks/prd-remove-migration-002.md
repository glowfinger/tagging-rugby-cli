# PRD: Remove Migration 002 and Canonicalise Migration 001

## Introduction

The database has two migrations:
- `001_create_videos_table.sql` — creates the full schema for fresh databases, but contains a backward-compat `ALTER TABLE notes ADD COLUMN video_id` hack at the bottom for old databases that predate the `video_id` column.
- `002_migrate_note_videos.sql` — moves data from the old `note_videos` table into the new `videos` table, then drops `note_videos`.

Since all data is test data, there are no real databases to migrate. We can delete migration 002 entirely and clean up migration 001 so it is a clean, canonical schema with no backward-compat hacks. A bonus cleanup: `db/sql/insert_note_video.sql` is a dead file that still targets the defunct `note_videos` table and should be deleted.

## Goals

- Remove the now-unnecessary `002_migrate_note_videos.sql` migration file.
- Make `001_create_videos_table.sql` the single, clean, canonical schema — no backward-compat workarounds.
- Delete the orphaned `insert_note_video.sql` file that references the defunct `note_videos` table.
- Verify the Go project builds and passes vet with no errors.

## User Stories

### US-001: Remove migration 002 file
**Description:** As a developer, I want migration 002 deleted so the migration system only contains a single, authoritative schema.

**Acceptance Criteria:**
- [ ] `db/sql/migrations/002_migrate_note_videos.sql` is deleted.
- [ ] No other file references `002_migrate_note_videos` (check with grep).
- [ ] `go build ./...` passes (`CGO_ENABLED=0`).

### US-002: Clean up migration 001 to be the canonical final schema
**Description:** As a developer, I want migration 001 to be a clean schema definition with no backward-compat hacks so a fresh database is set up correctly with a single, readable migration.

**Acceptance Criteria:**
- [ ] The backward-compat block is removed from `001_create_videos_table.sql`:
  ```sql
  -- Add video_id to notes for existing databases that predate this migration.
  -- On fresh databases the column already exists from the CREATE TABLE above,
  -- and the migration runner ignores the resulting "duplicate column" error.
  ALTER TABLE notes ADD COLUMN video_id INTEGER DEFAULT 0;
  ```
- [ ] The comment on line 3 (`-- note_videos is intentionally omitted — it was the old schema that migration 002 removes.`) is removed or updated to simply say this is the canonical schema.
- [ ] `notes.video_id` has a foreign key reference to `videos(id)`:
  ```sql
  video_id INTEGER DEFAULT 0 REFERENCES videos(id)
  ```
- [ ] The remaining SQL (CREATE TABLE statements + indexes) is otherwise unchanged.
- [ ] `go build ./...` passes (`CGO_ENABLED=0`).

### US-003: Delete orphaned insert_note_video.sql
**Description:** As a developer, I want the dead `insert_note_video.sql` file removed so there is no SQL that references the defunct `note_videos` table.

**Acceptance Criteria:**
- [ ] `db/sql/insert_note_video.sql` is deleted.
- [ ] Grep confirms no Go file embeds or references `insert_note_video.sql`.
- [ ] `go build ./...` passes (`CGO_ENABLED=0`).

### US-004: Verify full build and vet passes
**Description:** As a developer, I want to confirm the project builds cleanly after all deletions so nothing is accidentally broken.

**Acceptance Criteria:**
- [ ] `CGO_ENABLED=0 go build ./...` exits 0.
- [ ] `CGO_ENABLED=0 go vet ./...` exits 0.

## Functional Requirements

- FR-1: `db/sql/migrations/002_migrate_note_videos.sql` must not exist after this change.
- FR-2: `db/sql/migrations/001_create_videos_table.sql` must not contain any `ALTER TABLE` statements.
- FR-3: `db/sql/migrations/001_create_videos_table.sql` must contain all the correct `CREATE TABLE IF NOT EXISTS` statements for the full schema (videos, notes, note_clips, note_timing, note_tackles, note_zones, note_details, note_highlights) and the three indexes. The `notes.video_id` column must declare `REFERENCES videos(id)`.
- FR-4: `db/sql/insert_note_video.sql` must not exist after this change.
- FR-5: No Go source file may embed or reference `insert_note_video.sql`.

## Non-Goals

- Renaming or removing the `NoteVideo` Go struct — it is still used as a DTO throughout `tui.go` and `cmd/*.go` and is a separate concern.
- Changing any queries, functions, or TUI logic beyond the files listed above.
- Renaming `select_note_videos_by_note.sql` — it already correctly queries the `videos` table via the `notes.video_id` join.

## Technical Considerations

- The migration runner in `db/migrations.go` reads migration files dynamically from the embedded directory; deleting a file is sufficient — no version number hard-coding needs updating.
- `CGO_ENABLED=0` is required for all Go commands because the project uses `modernc.org/sqlite`.
- Go binary is at `/home/node/go/bin/go` — export `PATH` before running commands.

## Success Metrics

- Migration directory contains exactly one file: `001_create_videos_table.sql`.
- `001_create_videos_table.sql` contains no `ALTER TABLE` statements.
- Project builds and vets cleanly.

## Open Questions

- None. Scope is well-defined and data loss is acceptable (test data only).
