package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnrestrictLink(t *testing.T) {
	expected := UnrestrictedLink{
		ID:       "ABCDEF123",
		Filename: "movie.mkv",
		Filesize: 1073741824,
		Download: "https://download.example.com/movie.mkv",
		MimeType: "video/x-matroska",
		Host:     "example.com",
		Chunks:   16,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/unrestrict/link" {
			t.Errorf("expected path /unrestrict/link, got %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("link") != "https://hoster.com/file/abc" {
			t.Errorf("unexpected link form value: %s", r.FormValue("link"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	client := NewClient("test-token")
	client.BaseURL = srv.URL

	result, err := UnrestrictLink(client, "https://hoster.com/file/abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Filename != expected.Filename {
		t.Errorf("filename = %q, want %q", result.Filename, expected.Filename)
	}
	if result.Filesize != expected.Filesize {
		t.Errorf("filesize = %d, want %d", result.Filesize, expected.Filesize)
	}
	if result.Download != expected.Download {
		t.Errorf("download = %q, want %q", result.Download, expected.Download)
	}
}

func TestUnrestrictLinkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":      "service_unavailable",
			"error_code": 503,
		})
	}))
	defer srv.Close()

	client := NewClient("test-token")
	client.BaseURL = srv.URL

	_, err := UnrestrictLink(client, "https://hoster.com/file/abc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("HTTP status = %d, want %d", apiErr.HTTPStatus, http.StatusServiceUnavailable)
	}
}

