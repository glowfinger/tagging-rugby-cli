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
