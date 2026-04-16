package store

import "time"

// Status constants for history entries.
const (
	StatusPending       = "pending"
	StatusUnrestricting = "unrestricting"
	StatusDownloading   = "downloading"
	StatusCompleted     = "completed"
	StatusFailed        = "failed"
)

// HistoryEntry represents a single download in the history table.
type HistoryEntry struct {
	ID           int64
	OriginalLink string
	Filename     string
	Filesize     int64
	DownloadURL  string
	LocalPath    string
	Status       string
	ErrorMsg     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AddHistory inserts a new history entry and returns the assigned ID.
func (s *Store) AddHistory(e HistoryEntry) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO history (original_link, filename, filesize, download_url, local_path, status, error_msg)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.OriginalLink, e.Filename, e.Filesize, e.DownloadURL, e.LocalPath, e.Status, e.ErrorMsg,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateHistoryStatus updates the status and error message for a history entry.
func (s *Store) UpdateHistoryStatus(id int64, status string, errorMsg string) error {
	_, err := s.db.Exec(
		`UPDATE history SET status = ?, error_msg = ?, updated_at = datetime('now') WHERE id = ?`,
		status, errorMsg, id,
	)
	return err
}

// ListHistory returns history entries ordered by creation time (newest first),
// with pagination via offset and limit.
func (s *Store) ListHistory(offset, limit int) ([]HistoryEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, original_link, filename, filesize, download_url, local_path, status, error_msg, created_at, updated_at
		 FROM history ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

// ListFailed returns all history entries with a failed status.
func (s *Store) ListFailed() ([]HistoryEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, original_link, filename, filesize, download_url, local_path, status, error_msg, created_at, updated_at
		 FROM history WHERE status = ?`,
		StatusFailed,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

func scanHistoryRows(rows interface {
	Next() bool
	Scan(dest ...any) error
}) ([]HistoryEntry, error) {
	var entries []HistoryEntry
	for rows.Next() {
		var e HistoryEntry
		if err := rows.Scan(
			&e.ID, &e.OriginalLink, &e.Filename, &e.Filesize,
			&e.DownloadURL, &e.LocalPath, &e.Status, &e.ErrorMsg,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
