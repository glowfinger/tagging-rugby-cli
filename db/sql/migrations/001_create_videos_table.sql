-- Migration 001: Create all base tables.
-- This is the canonical schema for fresh databases.
-- note_videos is intentionally omitted — it was the old schema that migration 002 removes.

CREATE TABLE IF NOT EXISTS videos (
    id INTEGER PRIMARY KEY,
    path TEXT,
    filename TEXT,
    extension TEXT,
    format TEXT,
    filesize INTEGER,
    stop_time REAL
);

CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY,
    category TEXT,
    video_id INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS note_clips (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    name TEXT,
    duration REAL,
    started_at DATETIME,
    finished_at DATETIME,
    error_at DATETIME,
    error TEXT
);

CREATE TABLE IF NOT EXISTS note_timing (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    start REAL,
    end REAL
);

CREATE TABLE IF NOT EXISTS note_tackles (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    player TEXT,
    attempt INTEGER,
    outcome TEXT
);

CREATE TABLE IF NOT EXISTS note_zones (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    horizontal TEXT,
    vertical TEXT
);

CREATE TABLE IF NOT EXISTS note_details (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    type TEXT,
    note TEXT
);

CREATE TABLE IF NOT EXISTS note_highlights (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    type TEXT
);

CREATE INDEX IF NOT EXISTS idx_notes_category ON notes(category);
CREATE INDEX IF NOT EXISTS idx_note_details_type ON note_details(type);
CREATE INDEX IF NOT EXISTS idx_note_highlights_type ON note_highlights(type);

-- Add video_id to notes for existing databases that predate this migration.
-- On fresh databases the column already exists from the CREATE TABLE above,
-- and the migration runner ignores the resulting "duplicate column" error.
ALTER TABLE notes ADD COLUMN video_id INTEGER DEFAULT 0;
