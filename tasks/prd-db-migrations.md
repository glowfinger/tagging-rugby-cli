# PRD: Database Migration System

## Introduction

The app currently initializes the database by running `CREATE TABLE IF NOT EXISTS` statements. This works for fresh databases but cannot handle schema changes to existing databases (e.g., adding columns, moving data between tables, dropping tables). The recent `note_videos` → `videos` migration failed because there was no mechanism to alter existing tables.

This PRD introduces a versioned migration system: a `schema_migrations` table tracks which migrations have run, and numbered SQL files execute in order on startup. The first two migrations will handle the `note_videos` → `videos` data conversion.

## Goals

- Add a `schema_migrations` table to track applied migrations by version number
- Run pending migrations automatically on app startup, in order
- Store migrations as embedded SQL files (`db/sql/migrations/NNN_name.sql`)
- Implement migration 001: create `videos` table and add `video_id` column to `notes`
- Implement migration 002: migrate `note_videos` data into `videos`, set `video_id` on notes, drop `note_videos`
- Ensure fresh databases (no existing data) work without errors
- Ensure existing databases (with `note_videos` data) migrate cleanly

## User Stories

### US-001: Create migration runner infrastructure
**Description:** As a developer, I need a migration runner so that schema changes are applied automatically and tracked by version.

**Acceptance Criteria:**
- [ ] `schema_migrations` table created with column: `version INTEGER PRIMARY KEY`
- [ ] Migration files live in `db/sql/migrations/` as `NNN_name.sql` (zero-padded 3-digit version)
- [ ] All migration SQL files are embedded via `go:embed` in `db/queries.go` (or a new `db/migrations.go` embed file)
- [ ] New function `runMigrations(db)` in `db/db.go` replaces the current `migrate()` function
- [ ] `runMigrations` reads `schema_migrations` to find the highest applied version
- [ ] `runMigrations` executes each unapplied migration SQL file in version order
- [ ] Each migration runs inside a transaction — on failure, that migration is rolled back and the error is returned
- [ ] After successful execution, the migration version is inserted into `schema_migrations`
- [ ] `Open()` calls `runMigrations` instead of `migrate`
- [ ] `create_tables.sql` is still executed (for fresh DBs) before migrations run — or migrations handle all table creation
- [ ] CGO_ENABLED=0 go vet ./... passes
- [ ] CGO_ENABLED=0 go build ./... passes

### US-002: Migration 001 — Create videos table and add video_id to notes
**Description:** As a developer, I need the first migration to create the `videos` table and add `video_id` to the `notes` table so the new schema is in place for data migration.

**Acceptance Criteria:**
- [ ] File: `db/sql/migrations/001_create_videos_table.sql`
- [ ] Creates `videos` table if it doesn't exist (same schema as current `create_tables.sql`)
- [ ] Adds `video_id` column to `notes` table if it doesn't exist — must handle SQLite's limited ALTER TABLE (no NOT NULL without default on existing rows; use default 0 initially)
- [ ] On a fresh database (tables already have video_id from create_tables.sql), migration is a no-op or handles gracefully
- [ ] On an existing database (notes table without video_id), the column is added
- [ ] CGO_ENABLED=0 go vet ./... passes
- [ ] CGO_ENABLED=0 go build ./... passes

### US-003: Migration 002 — Migrate note_videos data and drop table
**Description:** As a developer, I need the second migration to move data from `note_videos` into `videos`, link notes to their videos, and drop `note_videos`.

**Acceptance Criteria:**
- [ ] File: `db/sql/migrations/002_migrate_note_videos.sql`
- [ ] Inserts unique videos from `note_videos` into `videos` (deduplicated by path)
- [ ] `filename` derived from `path` using SQLite string functions (e.g., `SUBSTR` + `INSTR` or `REPLACE`)
- [ ] `extension` derived from `path` using SQLite string functions
- [ ] `note_videos.size` maps to `videos.filesize`, `note_videos.stopped_at` maps to `videos.stop_time`
- [ ] `note_videos.duration` is not migrated (dropped)
- [ ] Updates `notes.video_id` to the corresponding `videos.id` based on matching path from `note_videos`
- [ ] Drops `note_videos` table after data migration
- [ ] On a fresh database (no `note_videos` table), migration handles gracefully (no error)
- [ ] On an existing database with data, all notes get correct `video_id` values
- [ ] CGO_ENABLED=0 go vet ./... passes
- [ ] CGO_ENABLED=0 go build ./... passes

## Functional Requirements

- FR-1: The `schema_migrations` table must be created automatically (not via a migration itself) before any migrations run
- FR-2: Migrations must execute in version order (001, 002, 003, ...)
- FR-3: Each migration must run in a transaction — all-or-nothing
- FR-4: Already-applied migrations must be skipped (idempotent startup)
- FR-5: `create_tables.sql` still runs with `IF NOT EXISTS` for fresh databases — migrations handle the delta for existing databases
- FR-6: Migration 001 must use SQLite-compatible ALTER TABLE (no ADD COLUMN ... NOT NULL without a default value on a table with existing rows)
- FR-7: Migration 002 must handle the case where `note_videos` doesn't exist (fresh DB) without erroring
- FR-8: SQLite `SUBSTR` / string functions must be used for filename/extension extraction since migrations are pure SQL

## Non-Goals

- No rollback / down migrations — forward-only
- No CLI command to run migrations manually (they run on startup)
- No migration generation tooling
- No changes to TUI or CLI commands

## Technical Considerations

- SQLite does not support `ALTER TABLE ... ADD COLUMN ... NOT NULL` on a table with existing rows unless a default is provided. Migration 001 should add `video_id INTEGER DEFAULT 0` first, then migration 002 sets the real values.
- SQLite's `SUBSTR` and `INSTR` (or `REPLACE`) can extract filename and extension from a path. Example: `SUBSTR(path, LENGTH(path) - LENGTH(REPLACE(path, '/', '')) + 1)` extracts the filename after the last `/`.
- Embedded SQL files via `go:embed` can use `//go:embed sql/migrations/*` to embed the entire directory, or individual file embeds.
- The `note_videos` table may not exist on fresh databases — migration 002 SQL should guard with a check (e.g., query `sqlite_master` for table existence).
- Transactions in SQLite cannot contain DDL that implicitly commits (like `DROP TABLE`). May need to split the transaction or use DDL outside the transaction boundary.

## Success Metrics

- Fresh database initializes without errors (all migrations run as no-ops)
- Existing database with `note_videos` data migrates successfully — notes get correct `video_id` values
- `schema_migrations` table shows versions 1 and 2 after startup
- Restarting the app does not re-run already-applied migrations
- CGO_ENABLED=0 go build ./... succeeds

## Open Questions

- Should SQLite's inability to run DDL (DROP TABLE) inside a transaction affect how we wrap migrations? Option: run DDL statements outside the transaction, or accept that SQLite auto-commits DDL.
- Should there be a `--migrate` CLI flag for explicit control, or is auto-migration on startup sufficient?
