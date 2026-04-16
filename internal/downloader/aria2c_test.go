package downloader

import (
	"slices"
	"testing"
)

func TestBuildAria2cArgs(t *testing.T) {
	params := ParamsForTier("fast")
	urls := []string{"https://example.com/a.bin", "https://example.com/b.bin"}
	outputDir := "/tmp/downloads"

	args := BuildAria2cArgs(params, urls, outputDir)

	// Flags using "-x value" syntax.
	shortFlags := map[string]string{
		"-j": "2",
		"-d": "/tmp/downloads",
	}
	for flag, value := range shortFlags {
		idx := slices.Index(args, flag)
		if idx == -1 {
			t.Errorf("missing flag %s in args: %v", flag, args)
			continue
		}
		if idx+1 >= len(args) || args[idx+1] != value {
			t.Errorf("flag %s has wrong value: got %q, want %q", flag, args[idx+1], value)
		}
	}

	// Flags using "--key=value" combined syntax.
	longFlags := []string{
		"--file-allocation=falloc",
		"--continue=true",
		"--lowest-speed-limit=100K",
	}
	for _, flag := range longFlags {
		if !slices.Contains(args, flag) {
			t.Errorf("missing flag %s in args: %v", flag, args)
		}
	}

	// Verify URLs are present.
	for _, u := range urls {
		if !slices.Contains(args, u) {
			t.Errorf("missing URL %s in args: %v", u, args)
		}
	}
}

func TestCheckAria2cInstalled(t *testing.T) {
	// Just verify the function doesn't panic; the result depends on the host.
	_ = IsAria2cInstalled()
}
