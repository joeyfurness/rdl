package store

import "testing"

func TestQueueAddAndList(t *testing.T) {
	s := openTestStore(t)

	if err := s.QueueAdd("https://example.com/a"); err != nil {
		t.Fatalf("QueueAdd: %v", err)
	}
	if err := s.QueueAdd("https://example.com/b"); err != nil {
		t.Fatalf("QueueAdd: %v", err)
	}

	items, err := s.QueueList()
	if err != nil {
		t.Fatalf("QueueList: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestQueueClear(t *testing.T) {
	s := openTestStore(t)

	s.QueueAdd("https://example.com/a")
	s.QueueAdd("https://example.com/b")

	if err := s.QueueClear(); err != nil {
		t.Fatalf("QueueClear: %v", err)
	}

	items, err := s.QueueList()
	if err != nil {
		t.Fatalf("QueueList: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestQueueDedup(t *testing.T) {
	s := openTestStore(t)

	s.QueueAdd("https://example.com/same")
	s.QueueAdd("https://example.com/same")

	items, err := s.QueueList()
	if err != nil {
		t.Fatalf("QueueList: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item (dedup), got %d", len(items))
	}
}
