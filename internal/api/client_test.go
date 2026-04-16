package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestClientGet(t *testing.T) {
	type resp struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp{Name: "alice", ID: 42})
	}))
	defer server.Close()

	c := NewClient("test-token")
	c.BaseURL = server.URL

	var got resp
	err := c.Get("/user", &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "alice" || got.ID != 42 {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestClientPost(t *testing.T) {
	type resp struct {
		Status string `json:"status"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer post-token" {
			t.Errorf("expected Bearer post-token, got %s", auth)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content type, got %s", ct)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.PostFormValue("link") != "https://example.com/file.zip" {
			t.Errorf("unexpected link value: %s", r.PostFormValue("link"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp{Status: "ok"})
	}))
	defer server.Close()

	c := NewClient("post-token")
	c.BaseURL = server.URL

	form := url.Values{}
	form.Set("link", "https://example.com/file.zip")

	var got resp
	err := c.Post("/unrestrict/link", form, &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestClientErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":      "permission_denied",
			"error_code": 9,
		})
	}))
	defer server.Close()

	c := NewClient("bad-token")
	c.BaseURL = server.URL

	var result map[string]interface{}
	err := c.Get("/protected", &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "permission_denied" {
		t.Errorf("expected code permission_denied, got %s", apiErr.Code)
	}
	if apiErr.ErrorCode != 9 {
		t.Errorf("expected error_code 9, got %d", apiErr.ErrorCode)
	}
	if apiErr.HTTPStatus != http.StatusForbidden {
		t.Errorf("expected HTTP 403, got %d", apiErr.HTTPStatus)
	}
}

func TestClientRateLimitRetry(t *testing.T) {
	var attempts int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt64(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	c := NewClient("token")
	c.BaseURL = server.URL

	var got map[string]string
	err := c.Get("/torrents", &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["result"] != "success" {
		t.Errorf("unexpected result: %+v", got)
	}
	if atomic.LoadInt64(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", atomic.LoadInt64(&attempts))
	}
}
