-- Migration 001: Create videos table and add video_id to notes

CREATE TABLE IF NOT EXISTS videos (
    id INTEGER PRIMARY KEY,
    path TEXT,
    filename TEXT,
    extension TEXT,
    format TEXT,
    filesize INTEGER,
    stop_time REAL
);

-- Add video_id to notes for existing databases.
-- On fresh DBs where create_tables.sql already has video_id, the duplicate
-- column error is caught and ignored by the migration runner.
ALTER TABLE notes ADD COLUMN video_id INTEGER DEFAULT 0;
