package database_test

import (
	"testing"

	"lite-rag/internal/database"
)

func TestOpen_InMemory(t *testing.T) {
	// Empty path opens an in-memory DuckDB instance.
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
}

func TestMigrate_TablesExist(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Both tables must be queryable after migration.
	tables := []string{"documents", "chunks"}
	for _, tbl := range tables {
		if err := db.Ping(); err != nil {
			t.Fatalf("Ping() error = %v", err)
		}
		rows, err := db.QueryRaw("SELECT COUNT(*) FROM " + tbl)
		if err != nil {
			t.Errorf("table %q not found after migration: %v", tbl, err)
			continue
		}
		rows.Close()
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	// Opening (and migrating) twice against the same file must not error.
	dir := t.TempDir()
	path := dir + "/test.db"

	db1, err := database.Open(path)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	db1.Close()

	db2, err := database.Open(path)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	db2.Close()
}

func TestOpen_DirectoryPathFails(t *testing.T) {
	// DuckDB cannot open a directory as a database file; Open must return an error.
	dir := t.TempDir()
	_, err := database.Open(dir)
	if err == nil {
		t.Error("expected error when opening a directory as DB path, got nil")
	}
}

func TestClose_Idempotent(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
}
