# PRD: Video Table

## Introduction

Currently, video metadata is stored in the `note_videos` table, which duplicates video data across every note that references the same video file. This PRD introduces a dedicated `videos` table as a first-class entity, replaces the `note_videos` junction table with a direct `video_id` foreign key on the `notes` table, and migrates existing data. This deduplicates video records and simplifies the one-video-to-many-notes relationship.

## Goals

- Create a `videos` table with columns: `id`, `path`, `filename`, `extension`, `format`, `filesize`, `stop_time`
- Add a `video_id` (NOT NULL) foreign key column to the `notes` table
- Remove the `note_videos` junction table
- Migrate existing `note_videos` data into the new `videos` table, deduplicating by file path
- Update all Go structs, database functions, and SQL queries to use the new schema

## User Stories

### US-001: Create videos table schema
**Description:** As a developer, I need a `videos` table so video metadata is stored once per unique video file.

**Acceptance Criteria:**
- [ ] New `videos` table created with columns: `id` (INTEGER PRIMARY KEY), `path` (TEXT), `filename` (TEXT), `extension` (TEXT), `format` (TEXT), `filesize` (INTEGER), `stop_time` (REAL)
- [ ] `notes` table gains a `video_id` (INTEGER NOT NULL) column with a foreign key reference to `videos(id)`
- [ ] `note_videos` table is dropped
- [ ] Schema defined in `db/sql/create_tables.sql`
- [ ] Typecheck/vet passes (`CGO_ENABLED=0 go vet ./...`)

### US-002: Write data migration
**Description:** As a developer, I need to migrate existing `note_videos` rows into the new `videos` table so no data is lost.

**Acceptance Criteria:**
- [ ] Migration logic extracts unique videos from `note_videos` (deduplicated by `path`) and inserts into `videos`
- [ ] `filename` and `extension` are derived from the existing `path` column during migration
- [ ] Each note's `video_id` is set to the corresponding `videos.id` based on its `note_videos.path`
- [ ] `note_videos.stopped_at` maps to `videos.stop_time`; `note_videos.size` maps to `videos.filesize`
- [ ] `note_videos.duration` is dropped (not carried to the new schema)
- [ ] Migration runs automatically on app startup (or as part of DB init)
- [ ] Existing database with data migrates without errors
- [ ] Typecheck/vet passes

### US-003: Update Go model structs
**Description:** As a developer, I need Go structs that match the new schema so the DB layer compiles and works.

**Acceptance Criteria:**
- [ ] New `Video` struct in `db/models.go` with fields: `ID int64`, `Path string`, `Filename string`, `Extension string`, `Format string`, `Filesize int64`, `StopTime float64`
- [ ] `NoteVideo` struct is removed from `db/models.go`
- [ ] `Note` struct gains a `VideoID int64` field
- [ ] `NoteChildren` struct no longer has a `Videos` field
- [ ] Typecheck/vet passes

### US-004: Update database functions and SQL queries
**Description:** As a developer, I need CRUD functions for the `videos` table and updated note queries that use `video_id`.

**Acceptance Criteria:**
- [ ] New SQL file: `insert_video.sql` — inserts a video row and returns its ID
- [ ] New SQL file: `select_video_by_id.sql` — selects a video by ID
- [ ] New SQL file: `select_video_by_path.sql` — selects a video by file path (for dedup/upsert logic)
- [ ] New Go function: `InsertVideo(db, path, filename, extension, format, filesize, stopTime) (int64, error)`
- [ ] New Go function: `SelectVideoByID(db, id) (*Video, error)`
- [ ] New Go function: `SelectVideoByPath(db, path) (*Video, error)`
- [ ] New Go function: `UpdateVideoStopTime(db, id, stopTime) error`
- [ ] `InsertNote` updated to accept `videoID int64` parameter
- [ ] `InsertNoteWithChildren` updated — no longer inserts into `note_videos`; expects `videoID` to be set on the note
- [ ] `select_notes_with_video.sql` rewritten to JOIN `notes` → `videos` via `notes.video_id` instead of through `note_videos`
- [ ] `SelectNoteVideosByNote` function removed (replaced by selecting the video via `notes.video_id`)
- [ ] Old SQL files removed: `insert_note_video.sql`, `select_note_videos_by_note.sql`
- [ ] Embedded SQL variables updated in `db/sql.go` (or wherever `//go:embed` directives live)
- [ ] Typecheck/vet passes

### US-005: Update callers of removed/changed functions
**Description:** As a developer, I need to update all code that previously used `NoteVideo`, `InsertNoteVideo`, or `SelectNoteVideosByNote` to use the new `Video` functions and `notes.video_id`.

**Acceptance Criteria:**
- [ ] All references to `NoteVideo` struct replaced with `Video` struct usage
- [ ] All calls to `InsertNoteVideo` replaced with `InsertVideo` + setting `video_id` on the note
- [ ] All calls to `SelectNoteVideosByNote` replaced with `SelectVideoByID` using the note's `video_id`
- [ ] `NoteChildren.Videos` field usage removed from TUI and CLI code
- [ ] Full build succeeds: `CGO_ENABLED=0 go build ./...`
- [ ] Typecheck/vet passes

## Functional Requirements

- FR-1: The `videos` table must have columns: `id` (INTEGER PRIMARY KEY), `path` (TEXT), `filename` (TEXT), `extension` (TEXT), `format` (TEXT), `filesize` (INTEGER), `stop_time` (REAL)
- FR-2: The `notes` table must have a `video_id` column (INTEGER NOT NULL) with a foreign key to `videos(id)`
- FR-3: The `note_videos` table must be dropped
- FR-4: `InsertVideo` must return the new video's ID so it can be set as `video_id` on notes
- FR-5: `SelectVideoByPath` must enable deduplication — before inserting a new video, check if one already exists at that path
- FR-6: `UpdateVideoStopTime` must allow updating the playback position for a video
- FR-7: Migration must deduplicate `note_videos` rows by `path`, keeping one canonical video record per unique path
- FR-8: Migration must populate `filename` and `extension` by parsing the `path` value (e.g., path `/foo/bar/game.mp4` → filename `game.mp4`, extension `.mp4`)

## Non-Goals

- No TUI/UI changes (this PRD covers schema + DB layer only)
- No video file management (upload, delete, transcode)
- No video duration tracking (dropped from schema)
- No changes to `note_clips` table (it retains its own `note_id` foreign key)

## Technical Considerations

- SQLite does not support `ALTER TABLE ... ADD COLUMN ... REFERENCES` with foreign key enforcement. The migration may need to recreate the `notes` table to add the `video_id` NOT NULL column with a proper foreign key constraint.
- The app uses `modernc.org/sqlite` (pure Go, requires `CGO_ENABLED=0`).
- All SQL is embedded via `//go:embed` directives — new `.sql` files must be added to the embed block.
- Foreign keys are enabled via `PRAGMA foreign_keys = ON` at connection time.
- Migration should be idempotent — running it on an already-migrated database should be a no-op.

## Success Metrics

- `CGO_ENABLED=0 go build ./...` succeeds with zero errors
- `CGO_ENABLED=0 go vet ./...` passes cleanly
- Existing databases migrate without data loss
- Each unique video path results in exactly one row in the `videos` table after migration

## Open Questions

- Should `videos.path` have a UNIQUE constraint to enforce deduplication at the DB level?
- Should the migration back up the database file before running?
- How should notes without a matching `note_videos` row be handled during migration (orphan notes)? Options: fail migration, assign a placeholder video, or make `video_id` temporarily nullable during migration.
