package store

import "time"

// QueueItem represents a link in the download queue.
type QueueItem struct {
	ID      int64
	Link    string
	AddedAt time.Time
}

// QueueAdd inserts a link into the queue, silently ignoring duplicates thanks
// to the UNIQUE constraint on the link column.
func (s *Store) QueueAdd(link string) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO queue (link) VALUES (?)`, link)
	return err
}

// QueueList returns all items in the queue ordered by addition time.
func (s *Store) QueueList() ([]QueueItem, error) {
	rows, err := s.db.Query(`SELECT id, link, added_at FROM queue ORDER BY added_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []QueueItem
	for rows.Next() {
		var q QueueItem
		if err := rows.Scan(&q.ID, &q.Link, &q.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, q)
	}
	return items, nil
}

// QueueClear removes all items from the queue.
func (s *Store) QueueClear() error {
	_, err := s.db.Exec(`DELETE FROM queue`)
	return err
}

// QueueRemove removes a single item from the queue by ID.
func (s *Store) QueueRemove(id int64) error {
	_, err := s.db.Exec(`DELETE FROM queue WHERE id = ?`, id)
	return err
}
