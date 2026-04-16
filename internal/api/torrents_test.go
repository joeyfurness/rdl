package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTorrents(t *testing.T) {
	expected := []TorrentInfo{
		{
			ID:       "TORRENT1",
			Filename: "movie.mkv",
			Hash:     "abc123",
			Bytes:    1073741824,
			Host:     "real-debrid.com",
			Status:   "downloaded",
			Progress: 100,
			Links:    []string{"https://real-debrid.com/d/abc"},
		},
		{
			ID:       "TORRENT2",
			Filename: "series.zip",
			Hash:     "def456",
			Bytes:    2147483648,
			Host:     "real-debrid.com",
			Status:   "downloading",
			Progress: 45,
			Links:    []string{},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/torrents" {
			t.Errorf("expected path /torrents, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	client := NewClient("test-token")
	client.BaseURL = srv.URL

	result, err := ListTorrents(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 torrents, got %d", len(result))
	}
	if result[0].Filename != "movie.mkv" {
		t.Errorf("filename = %q, want %q", result[0].Filename, "movie.mkv")
	}
	if result[1].Filename != "series.zip" {
		t.Errorf("filename = %q, want %q", result[1].Filename, "series.zip")
	}
}

func TestAddMagnet(t *testing.T) {
	expected := AddTorrentResponse{
		ID:  "NEW123",
		URI: "https://api.real-debrid.com/rest/1.0/torrents/info/NEW123",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/torrents/addMagnet" {
			t.Errorf("expected path /torrents/addMagnet, got %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		magnet := r.FormValue("magnet")
		if magnet != "magnet:?xt=urn:btih:abc123" {
			t.Errorf("unexpected magnet value: %s", magnet)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	client := NewClient("test-token")
	client.BaseURL = srv.URL

	result, err := AddMagnet(client, "magnet:?xt=urn:btih:abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "NEW123" {
		t.Errorf("ID = %q, want %q", result.ID, "NEW123")
	}
	if result.URI != expected.URI {
		t.Errorf("URI = %q, want %q", result.URI, expected.URI)
	}
}

func TestGetTorrentInfo(t *testing.T) {
	expected := TorrentInfo{
		ID:       "TORRENT1",
		Filename: "movie.mkv",
		Hash:     "abc123",
		Bytes:    1073741824,
		Status:   "downloaded",
		Progress: 100,
		Links:    []string{"https://real-debrid.com/d/abc"},
		Files: []TorrentFile{
			{ID: 1, Path: "movie.mkv", Bytes: 1073741824, Selected: 1},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/torrents/info/TORRENT1" {
			t.Errorf("expected path /torrents/info/TORRENT1, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	client := NewClient("test-token")
	client.BaseURL = srv.URL

	result, err := GetTorrentInfo(client, "TORRENT1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Filename != "movie.mkv" {
		t.Errorf("filename = %q, want %q", result.Filename, "movie.mkv")
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	if result.Files[0].Path != "movie.mkv" {
		t.Errorf("file path = %q, want %q", result.Files[0].Path, "movie.mkv")
	}
}

func TestFormatTorrentStatus(t *testing.T) {
	tests := []struct {
		status   string
		progress int
		want     string
	}{
		{"downloaded", 100, "Downloaded"},
		{"downloading", 45, "Downloading (45%)"},
		{"waiting_files_selection", 0, "Waiting for File Selection"},
		{"magnet_error", 0, "Magnet Error"},
		{"queued", 0, "Queued"},
		{"error", 0, "Error"},
		{"virus", 0, "Virus Detected"},
		{"compressing", 80, "Compressing (80%)"},
		{"uploading", 60, "Uploading (60%)"},
		{"dead", 0, "Dead"},
		{"unknown_status", 0, "unknown_status"},
	}

	for _, tt := range tests {
		got := FormatTorrentStatus(tt.status, tt.progress)
		if got != tt.want {
			t.Errorf("FormatTorrentStatus(%q, %d) = %q, want %q", tt.status, tt.progress, got, tt.want)
		}
	}
}
