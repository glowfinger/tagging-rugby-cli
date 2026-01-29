CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY,
    video_path TEXT,
    timestamp_seconds REAL,
    text TEXT,
    category TEXT,
    player TEXT,
    team TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS clips (
    id INTEGER PRIMARY KEY,
    video_path TEXT,
    start_seconds REAL,
    end_seconds REAL,
    description TEXT,
    category TEXT,
    player TEXT,
    team TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tackles (
    id INTEGER PRIMARY KEY,
    video_path TEXT,
    timestamp_seconds REAL,
    player TEXT,
    team TEXT,
    attempt INTEGER,
    outcome TEXT CHECK(outcome IN ('missed', 'completed', 'possible', 'other')),
    followed TEXT,
    star INTEGER DEFAULT 0,
    notes TEXT,
    zone TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
