package ui

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDetectMode(t *testing.T) {
	tests := []struct {
		name      string
		jsonFlag  bool
		quietFlag bool
		envMode   string
		cfgMode   string
		isTTY     bool
		want      Mode
	}{
		{
			name:     "json flag true",
			jsonFlag: true,
			want:     ModeJSON,
		},
		{
			name:      "quiet flag true",
			quietFlag: true,
			want:      ModeQuiet,
		},
		{
			name:    "env json",
			envMode: "json",
			want:    ModeJSON,
		},
		{
			name:    "config quiet",
			cfgMode: "quiet",
			want:    ModeQuiet,
		},
		{
			name:  "isTTY true, auto",
			isTTY: true,
			want:  ModeInteractive,
		},
		{
			name:  "isTTY false, auto",
			isTTY: false,
			want:  ModeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectMode(tt.jsonFlag, tt.quietFlag, tt.envMode, tt.cfgMode, tt.isTTY)
			if got != tt.want {
				t.Errorf("DetectMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONEmitter(t *testing.T) {
	var buf bytes.Buffer
	em := NewJSONEmitter(&buf)

	em.Emit(Event{
		Type:     "download",
		Filename: "file1.zip",
		Size:     1024,
	})
	em.Emit(Event{
		Type:    "error",
		Message: "not found",
		Code:    404,
	})

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 NDJSON lines, got %d", len(lines))
	}

	var first map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("failed to parse first line: %v", err)
	}

	if first["event"] != "download" {
		t.Errorf("first event type = %v, want %q", first["event"], "download")
	}
	if first["filename"] != "file1.zip" {
		t.Errorf("first filename = %v, want %q", first["filename"], "file1.zip")
	}
	if first["size"] != float64(1024) {
		t.Errorf("first size = %v, want %v", first["size"], 1024)
	}
}

func TestQuietEmitter_OnlyErrors(t *testing.T) {
	var buf bytes.Buffer
	em := NewQuietEmitter(&buf)

	em.Emit(Event{Type: "download", Filename: "file.zip"})
	em.Emit(Event{Type: "error", Message: "something failed"})

	output := buf.String()
	if !strings.Contains(output, "something failed") {
		t.Error("expected error message in output")
	}
	if strings.Contains(output, "file.zip") {
		t.Error("non-error event should not appear in quiet output")
	}
}
