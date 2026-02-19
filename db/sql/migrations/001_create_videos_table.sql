-- Migration 001: Create all base tables.
-- This is the canonical schema for all databases.

CREATE TABLE IF NOT EXISTS videos (
    id INTEGER PRIMARY KEY,
    path TEXT,
    filename TEXT,
    extension TEXT,
    format TEXT,
    filesize INTEGER
);

CREATE TABLE IF NOT EXISTS video_timings (
    id INTEGER PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    stopped REAL,
    length REAL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_video_timings_video_id ON video_timings(video_id);

CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY,
    video_id INTEGER DEFAULT 0 REFERENCES videos(id),
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS note_clips (
    id INTEGER PRIMARY KEY,
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    folder TEXT,
    filename TEXT,
    extension TEXT,
    format TEXT,
    filesize INTEGER,
    status TEXT,
    started_at DATETIME,
    finished_at DATETIME,
    error_at DATETIME,
    log TEXT
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
