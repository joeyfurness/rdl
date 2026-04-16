package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "tokens.json")

	original := &TokenData{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		ObtainedAt:   1700000000,
	}

	if err := SaveTokens(path, original); err != nil {
		t.Fatalf("SaveTokens failed: %v", err)
	}

	// Verify file permissions are 0600
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}

	// Load back and verify
	loaded, err := LoadTokens(path)
	if err != nil {
		t.Fatalf("LoadTokens failed: %v", err)
	}

	if loaded.AccessToken != original.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, original.AccessToken)
	}
	if loaded.RefreshToken != original.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %q, want %q", loaded.RefreshToken, original.RefreshToken)
	}
	if loaded.ExpiresIn != original.ExpiresIn {
		t.Errorf("ExpiresIn mismatch: got %d, want %d", loaded.ExpiresIn, original.ExpiresIn)
	}
	if loaded.TokenType != original.TokenType {
		t.Errorf("TokenType mismatch: got %q, want %q", loaded.TokenType, original.TokenType)
	}
	if loaded.ClientID != original.ClientID {
		t.Errorf("ClientID mismatch: got %q, want %q", loaded.ClientID, original.ClientID)
	}
	if loaded.ClientSecret != original.ClientSecret {
		t.Errorf("ClientSecret mismatch: got %q, want %q", loaded.ClientSecret, original.ClientSecret)
	}
	if loaded.ObtainedAt != original.ObtainedAt {
		t.Errorf("ObtainedAt mismatch: got %d, want %d", loaded.ObtainedAt, original.ObtainedAt)
	}
}

func TestLoadTokensMissing(t *testing.T) {
	_, err := LoadTokens("/nonexistent/path/tokens.json")
	if err == nil {
		t.Fatal("expected error loading from nonexistent path, got nil")
	}
}

func TestDeviceCodeRequest(t *testing.T) {
	expected := DeviceCodeResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "ABCD1234",
		VerificationURL: "https://real-debrid.com/device",
		ExpiresIn:       600,
		Interval:        5,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/code" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		clientID := r.URL.Query().Get("client_id")
		if clientID != DefaultClientID {
			t.Errorf("unexpected client_id: %s", clientID)
		}

		newCreds := r.URL.Query().Get("new_credentials")
		if newCreds != "yes" {
			t.Errorf("expected new_credentials=yes, got %s", newCreds)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	resp, err := RequestDeviceCode(server.URL, DefaultClientID)
	if err != nil {
		t.Fatalf("RequestDeviceCode failed: %v", err)
	}

	if resp.UserCode != expected.UserCode {
		t.Errorf("UserCode mismatch: got %q, want %q", resp.UserCode, expected.UserCode)
	}
	if resp.DeviceCode != expected.DeviceCode {
		t.Errorf("DeviceCode mismatch: got %q, want %q", resp.DeviceCode, expected.DeviceCode)
	}
	if resp.VerificationURL != expected.VerificationURL {
		t.Errorf("VerificationURL mismatch: got %q, want %q", resp.VerificationURL, expected.VerificationURL)
	}
	if resp.ExpiresIn != expected.ExpiresIn {
		t.Errorf("ExpiresIn mismatch: got %d, want %d", resp.ExpiresIn, expected.ExpiresIn)
	}
	if resp.Interval != expected.Interval {
		t.Errorf("Interval mismatch: got %d, want %d", resp.Interval, expected.Interval)
	}
}
