// Package database provides DuckDB connectivity and schema management for lite-rag.
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb/v2"
)

// DB wraps a DuckDB connection.
type DB struct {
	conn *sql.DB
}

// Open opens (or creates) the DuckDB database at path and applies migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb %s: %w", path, err)
	}
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return db, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Ping verifies the database connection is alive.
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// QueryRaw executes a raw SQL query and returns the resulting rows.
// Intended for tests and diagnostic use only.
func (db *DB) QueryRaw(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// migrate creates all required tables if they do not already exist and applies
// incremental schema changes for pre-existing databases.
func (db *DB) migrate() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS documents (
    id              TEXT      PRIMARY KEY,
    file_path       TEXT      NOT NULL,
    file_hash       TEXT      NOT NULL,
    total_chunks    INTEGER   NOT NULL,
    indexed_at      TIMESTAMP NOT NULL,
    embedding_model TEXT      NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS chunks (
    id           TEXT     PRIMARY KEY,
    document_id  TEXT     NOT NULL,  -- logical FK to documents(id); enforced by app
    chunk_index  INTEGER  NOT NULL,
    heading_path TEXT,
    content      TEXT     NOT NULL,
    embedding    FLOAT[]            -- []float32 stored as DuckDB list; requires go-duckdb v2
);
`
	if _, err := db.conn.Exec(ddl); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	// Add embedding_model to pre-existing databases that lack the column.
	const alterDDL = `ALTER TABLE documents ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT '';`
	if _, err := db.conn.Exec(alterDDL); err != nil {
		return fmt.Errorf("migrate alter: %w", err)
	}
	return nil
}
