package store

import (
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestHistoryAddAndList(t *testing.T) {
	s := openTestStore(t)

	id, err := s.AddHistory(HistoryEntry{
		OriginalLink: "https://example.com/file",
		Filename:     "file.zip",
		Filesize:     1024,
		DownloadURL:  "https://cdn.example.com/file.zip",
		LocalPath:    "/tmp/file.zip",
		Status:       StatusCompleted,
	})
	if err != nil {
		t.Fatalf("AddHistory: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	entries, err := s.ListHistory(0, 10)
	if err != nil {
		t.Fatalf("ListHistory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.OriginalLink != "https://example.com/file" {
		t.Errorf("unexpected original_link: %s", e.OriginalLink)
	}
	if e.Filename != "file.zip" {
		t.Errorf("unexpected filename: %s", e.Filename)
	}
	if e.Filesize != 1024 {
		t.Errorf("unexpected filesize: %d", e.Filesize)
	}
	if e.Status != StatusCompleted {
		t.Errorf("unexpected status: %s", e.Status)
	}
}

func TestHistoryListFailed(t *testing.T) {
	s := openTestStore(t)

	s.AddHistory(HistoryEntry{OriginalLink: "link1", Status: StatusCompleted})
	s.AddHistory(HistoryEntry{OriginalLink: "link2", Status: StatusFailed, ErrorMsg: "timeout"})
	s.AddHistory(HistoryEntry{OriginalLink: "link3", Status: StatusFailed, ErrorMsg: "404"})

	failed, err := s.ListFailed()
	if err != nil {
		t.Fatalf("ListFailed: %v", err)
	}
	if len(failed) != 2 {
		t.Fatalf("expected 2 failed entries, got %d", len(failed))
	}
}

func TestHistoryUpdateStatus(t *testing.T) {
	s := openTestStore(t)

	id, err := s.AddHistory(HistoryEntry{
		OriginalLink: "link1",
		Status:       StatusDownloading,
	})
	if err != nil {
		t.Fatalf("AddHistory: %v", err)
	}

	if err := s.UpdateHistoryStatus(id, StatusCompleted, ""); err != nil {
		t.Fatalf("UpdateHistoryStatus: %v", err)
	}

	entries, err := s.ListHistory(0, 10)
	if err != nil {
		t.Fatalf("ListHistory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Status != StatusCompleted {
		t.Errorf("expected status %q, got %q", StatusCompleted, entries[0].Status)
	}
}
