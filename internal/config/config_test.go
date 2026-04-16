package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Download.Directory != "~/Downloads" {
		t.Errorf("expected directory ~/Downloads, got %s", cfg.Download.Directory)
	}
	if cfg.Download.SpeedTier != "auto" {
		t.Errorf("expected speed_tier auto, got %s", cfg.Download.SpeedTier)
	}
	if cfg.Download.MaxRetries != 3 {
		t.Errorf("expected max_retries 3, got %d", cfg.Download.MaxRetries)
	}
	if cfg.Output.Mode != "auto" {
		t.Errorf("expected output.mode auto, got %s", cfg.Output.Mode)
	}
	if cfg.Behavior.Overwrite != "resume" {
		t.Errorf("expected behavior.overwrite resume, got %s", cfg.Behavior.Overwrite)
	}
}

func TestConfigDir(t *testing.T) {
	// Test default config dir
	t.Run("default", func(t *testing.T) {
		os.Unsetenv("RDL_CONFIG_DIR")
		dir := ConfigDir()
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", "rdl")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})

	// Test RDL_CONFIG_DIR override
	t.Run("env_override", func(t *testing.T) {
		override := t.TempDir()
		t.Setenv("RDL_CONFIG_DIR", override)
		dir := ConfigDir()
		if dir != override {
			t.Errorf("expected %s, got %s", override, dir)
		}
	})
}

func TestLoadConfigCreatesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("RDL_CONFIG_DIR", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Verify defaults
	if cfg.Download.Directory != "~/Downloads" {
		t.Errorf("expected directory ~/Downloads, got %s", cfg.Download.Directory)
	}

	// Verify file was created
	configFile := filepath.Join(tmpDir, "config.toml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("config.toml was not created")
	}

	// Verify file contains TOML content
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "~/Downloads") {
		t.Error("config.toml does not contain expected default directory")
	}
}

func TestExpandDir(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/Downloads", filepath.Join(home, "Downloads")},
		{"/absolute/path", "/absolute/path"},
		{"~/", home},
		{"~", home},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
