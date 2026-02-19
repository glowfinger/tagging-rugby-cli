-- Bootstrap: create the migration tracking table before any migrations run.
-- All application tables are created by versioned migrations in db/sql/migrations/.
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY
);
