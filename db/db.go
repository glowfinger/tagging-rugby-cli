package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open opens or creates the SQLite database at the default location.
// The database file is created at ~/.local/share/tagging-rugby-cli/data.db.
// Parent directories are created if they don't exist.
func Open() (*sql.DB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open the database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Verify connection works
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Enable WAL journal mode: allows concurrent readers + one writer,
	// greatly reducing SQLITE_BUSY errors from the background clip processor.
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, err
	}
	// Wait up to 5 seconds on lock contention rather than failing immediately.
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	// Ensure the UNIQUE INDEX on note_clips(note_id) exists. This index is
	// required for the ON CONFLICT(note_id) upsert in UpsertNoteClipPending.
	// Existing databases that were migrated before this index was added to
	// the migration file won't have it, so we create it here idempotently.
	if _, err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_note_clips_note_id ON note_clips(note_id)"); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// getDBPath returns the path to the database file.
func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".local", "share", "tagging-rugby-cli", "data.db"), nil
}
