package store

import (
	"path/filepath"
	"testing"
)

func TestOpenCreatesDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// Verify both tables exist via sqlite_master.
	for _, table := range []string{"history", "queue"} {
		var name string
		err := s.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %q not found: %v", table, err)
		}
		if name != table {
			t.Fatalf("expected table %q, got %q", table, name)
		}
	}
}
