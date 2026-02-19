package db

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed all:sql/migrations
var migrationsFS embed.FS

// runMigrations bootstraps schema_migrations (via create_tables.sql) then
// applies any unapplied versioned migrations in order.
func runMigrations(db *sql.DB) error {
	// Bootstrap: create schema_migrations table before anything else.
	// This is the only table created outside of versioned migrations.
	if _, err := db.Exec(BootstrapSQL); err != nil {
		return fmt.Errorf("bootstrap schema_migrations: %w", err)
	}

	// Read all migration files
	entries, err := migrationsFS.ReadDir("sql/migrations")
	if err != nil {
		return fmt.Errorf("reading migrations dir: %w", err)
	}

	// Parse and sort migration files by version number
	type migration struct {
		version int
		name    string
	}
	var migrations []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		// Parse version from NNN_name.sql format
		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		migrations = append(migrations, migration{version: v, name: e.Name()})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Get already-applied versions
	applied := make(map[int]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("querying schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scanning migration version: %w", err)
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating migration versions: %w", err)
	}

	// Apply unapplied migrations in order
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}

		sqlBytes, err := migrationsFS.ReadFile(filepath.Join("sql/migrations", m.name))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", m.name, err)
		}
		sqlContent := string(sqlBytes)

		// Check -- requires-table: directives. If a required table does not exist,
		// skip the SQL body but still mark the migration as applied.
		// This handles migrations that reference tables which may have already been
		// removed (e.g. note_videos on a fresh database).
		skip, err := shouldSkip(db, sqlContent)
		if err != nil {
			return fmt.Errorf("checking preconditions for migration %s: %w", m.name, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("beginning transaction for migration %d: %w", m.version, err)
		}

		if !skip {
			// Execute each statement in the migration separately.
			// This allows graceful handling of idempotent DDL like
			// ALTER TABLE ADD COLUMN on columns that already exist.
			stmts := splitStatements(sqlContent)
			for _, stmt := range stmts {
				if _, err := tx.Exec(stmt); err != nil {
					// Ignore "duplicate column" errors for ADD COLUMN idempotency
					if strings.Contains(err.Error(), "duplicate column") {
						continue
					}
					tx.Rollback()
					return fmt.Errorf("executing migration %s: %w", m.name, err)
				}
			}
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", m.version, err)
		}
	}

	return nil
}

// shouldSkip checks for -- requires-table: directives in the migration SQL.
// If any required table is absent from the database, it returns true so the
// migration body is skipped (but the version is still recorded as applied).
func shouldSkip(db *sql.DB, sqlContent string) (bool, error) {
	for _, line := range strings.Split(sqlContent, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "-- requires-table:") {
			continue
		}
		table := strings.TrimSpace(strings.TrimPrefix(line, "-- requires-table:"))
		if table == "" {
			continue
		}
		var count int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&count)
		if err != nil {
			return false, err
		}
		if count == 0 {
			return true, nil
		}
	}
	return false, nil
}

// splitStatements splits a SQL string into individual statements on semicolons.
// Empty statements and comment-only blocks are skipped.
func splitStatements(sql string) []string {
	raw := strings.Split(sql, ";")
	var stmts []string
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// Skip comment-only blocks
		lines := strings.Split(s, "\n")
		hasCode := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "--") {
				hasCode = true
				break
			}
		}
		if hasCode {
			stmts = append(stmts, s)
		}
	}
	return stmts
}
