# PRD: Improve Data Model

## Introduction

Replace the current flat data model with a normalized, extensible schema. The existing `notes`, `clips`, `tackles`, and `categories` tables are replaced by a central `notes` table with related child tables for videos, clips, timing, tackles, zones, details, and highlights. This eliminates data duplication, provides better structure for querying, and supports richer metadata per annotation.

## Goals

- Replace all existing tables with the new normalized schema
- Maintain a single `notes` table as the root entity for all annotations
- Store video metadata, clip export state, timing, tackles, zones, freeform details, and highlights as separate child tables linked by `note_id`
- Use indexed text columns instead of a separate categories table
- Keep the schema simple with no foreign key to a teams or players table — these remain plain text where needed

## User Stories

### US-047: Create new database schema
**Description:** As a developer, I need the new normalized table structure so that all downstream features can build on it.

**Acceptance Criteria:**
- [ ] Drop existing `notes`, `clips`, `tackles`, and `categories` tables
- [ ] Create `notes` table with columns: `id` (INTEGER PRIMARY KEY), `category` (TEXT, indexed), `created_at` (DATETIME DEFAULT CURRENT_TIMESTAMP)
- [ ] Create `note_videos` table with columns: `id` (INTEGER PRIMARY KEY), `note_id` (INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE), `path` (TEXT), `size` (INTEGER), `duration` (REAL), `format` (TEXT), `stopped_at` (REAL)
- [ ] Create `note_clips` table with columns: `id` (INTEGER PRIMARY KEY), `note_id` (INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE), `name` (TEXT), `duration` (REAL), `started_at` (DATETIME), `finished_at` (DATETIME, nullable), `error_at` (DATETIME, nullable), `error` (TEXT, nullable)
- [ ] Create `note_timing` table with columns: `id` (INTEGER PRIMARY KEY), `note_id` (INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE), `start` (REAL), `end` (REAL)
- [ ] Create `note_tackles` table with columns: `id` (INTEGER PRIMARY KEY), `note_id` (INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE), `player` (TEXT), `attempt` (INTEGER), `outcome` (TEXT)
- [ ] Create `note_zones` table with columns: `id` (INTEGER PRIMARY KEY), `note_id` (INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE), `horizontal` (TEXT), `vertical` (TEXT)
- [ ] Create `note_details` table with columns: `id` (INTEGER PRIMARY KEY), `note_id` (INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE), `type` (TEXT, indexed), `note` (TEXT)
- [ ] Create `note_highlights` table with columns: `id` (INTEGER PRIMARY KEY), `note_id` (INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE), `type` (TEXT, indexed)
- [ ] All `note_id` foreign keys reference `notes(id)`
- [ ] Typecheck/lint passes (`go vet ./...`)

### US-048: Update embedded SQL queries for new schema
**Description:** As a developer, I need the SQL query files updated to work with the new schema so that the application compiles and runs.

**Acceptance Criteria:**
- [ ] Remove SQL files that reference dropped tables/columns (old insert/select for clips, tackles, categories)
- [ ] Create new SQL files for inserting into each child table (`note_videos`, `note_clips`, `note_timing`, `note_tackles`, `note_zones`, `note_details`, `note_highlights`)
- [ ] Create new SQL files for selecting/querying each child table by `note_id`
- [ ] Update `db/queries.go` to embed the new SQL files and remove references to deleted ones
- [ ] All embedded SQL compiles without error
- [ ] Typecheck passes

### US-049: Update Go database layer
**Description:** As a developer, I need the Go code in `db/` updated to use the new schema so that the rest of the application can persist and retrieve data.

**Acceptance Criteria:**
- [ ] Update or replace Go structs to match new table shapes
- [ ] Update all database functions (insert, select, delete) to target new tables
- [ ] Remove functions that operated on the old `clips`, `tackles`, `categories` tables
- [ ] Add functions for inserting/querying each child table
- [ ] Typecheck passes

### US-050: Update CLI commands for new data model
**Description:** As a developer, I need the CLI commands (`cmd/note.go`, `cmd/clip.go`, `cmd/tackle.go`, `cmd/category.go`) updated or removed to reflect the new schema.

**Acceptance Criteria:**
- [ ] Remove `cmd/category.go` (no more categories table)
- [ ] Update `cmd/note.go` to work with the new `notes` table and child tables
- [ ] Update `cmd/clip.go` to work with `note_clips` (or remove if clip commands are folded into note commands)
- [ ] Update `cmd/tackle.go` to work with `note_tackles` (or remove if tackle commands are folded into note commands)
- [ ] Root command no longer registers removed subcommands
- [ ] Application compiles and runs
- [ ] Typecheck passes

### US-051: Update TUI for new data model
**Description:** As a user, I need the TUI to work with the new data model so that I can continue creating and viewing annotations.

**Acceptance Criteria:**
- [ ] TUI note creation writes to `notes` + relevant child tables
- [ ] TUI tackle input writes to `notes` + `note_tackles`
- [ ] Notes list displays data from joined queries across new tables
- [ ] Stats panel queries `note_tackles` instead of old `tackles` table
- [ ] Timeline markers query `note_timing` for timestamp positions
- [ ] Application compiles, runs, and displays correctly
- [ ] Typecheck passes

## Functional Requirements

- FR-1: Drop all existing tables (`notes`, `clips`, `tackles`, `categories`) on schema creation
- FR-2: Create `notes` table with `id`, `category` (indexed), `created_at`
- FR-3: Create `note_videos` table with `id`, `note_id`, `path`, `size`, `duration`, `format`, `stopped_at`
- FR-4: Create `note_clips` table with `id`, `note_id`, `name`, `duration`, `started_at`, `finished_at` (nullable), `error_at` (nullable), `error` (nullable)
- FR-5: Create `note_timing` table with `id`, `note_id`, `start`, `end`
- FR-6: Create `note_tackles` table with `id`, `note_id`, `player`, `attempt`, `outcome`
- FR-7: Create `note_zones` table with `id`, `note_id`, `horizontal`, `vertical`
- FR-8: Create `note_details` table with `id`, `note_id`, `type` (indexed), `note`
- FR-9: Create `note_highlights` table with `id`, `note_id`, `type` (indexed)
- FR-10: All child tables use `note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE` as foreign key
- FR-11: Remove all Go code, SQL files, and CLI commands that depend on the old schema
- FR-12: Update all queries, database functions, CLI commands, and TUI components to use the new schema

## Non-Goals

- No data migration from old schema — this is a clean break
- No separate `teams`, `players`, or `categories` lookup tables
- No changes to mpv integration or video playback controls
- No changes to the TUI layout or styling
- No changes to clip export (ffmpeg) logic beyond updating which table is queried

## Technical Considerations

- SQLite `REFERENCES` constraints require `PRAGMA foreign_keys = ON` to be enforced at runtime — ensure this is set on connection open
- Indexes on `notes.category`, `note_details.type`, and `note_highlights.type` should be created explicitly with `CREATE INDEX` statements
- All SQL files continue to use the `//go:embed` pattern in `db/queries.go`
- The database path remains `~/.local/share/tagging-rugby-cli/data.db`
- Since this is a destructive migration, the simplest approach is to drop all tables and recreate — no migration versioning needed

## Success Metrics

- Application compiles and passes `go vet ./...`
- All TUI workflows (create note, create tackle, view list, view stats) function against new schema
- No references to old table names remain in codebase

## Open Questions

None — all decisions resolved.
