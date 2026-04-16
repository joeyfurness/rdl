package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store wraps an SQLite database for download history and queue persistence.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) an SQLite database at the given path, enables WAL
// mode, and runs migrations.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates the required tables if they do not already exist.
func (s *Store) migrate() error {
	const historyDDL = `CREATE TABLE IF NOT EXISTS history (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		original_link TEXT    NOT NULL,
		filename      TEXT    NOT NULL DEFAULT '',
		filesize      INTEGER NOT NULL DEFAULT 0,
		download_url  TEXT    NOT NULL DEFAULT '',
		local_path    TEXT    NOT NULL DEFAULT '',
		status        TEXT    NOT NULL DEFAULT 'pending',
		error_msg     TEXT    NOT NULL DEFAULT '',
		created_at    DATETIME NOT NULL DEFAULT (datetime('now')),
		updated_at    DATETIME NOT NULL DEFAULT (datetime('now'))
	)`

	const queueDDL = `CREATE TABLE IF NOT EXISTS queue (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		link     TEXT    NOT NULL UNIQUE,
		added_at DATETIME NOT NULL DEFAULT (datetime('now'))
	)`

	if _, err := s.db.Exec(historyDDL); err != nil {
		return fmt.Errorf("create history table: %w", err)
	}
	if _, err := s.db.Exec(queueDDL); err != nil {
		return fmt.Errorf("create queue table: %w", err)
	}
	return nil
}
