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

// runMigrations creates the schema_migrations table, runs create_tables.sql
// for fresh databases, then applies any unapplied migrations in order.
func runMigrations(db *sql.DB) error {
	// Create schema_migrations table to track applied versions
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY
	)`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	// Run create_tables.sql for fresh databases (all IF NOT EXISTS)
	_, err = db.Exec(CreateTablesSQL)
	if err != nil {
		return fmt.Errorf("running create_tables: %w", err)
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

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("beginning transaction for migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("executing migration %s: %w", m.name, err)
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
