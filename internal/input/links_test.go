package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com/file", true},
		{"http://example.com/file", true},
		{"magnet:?xt=urn:btih:abc123", true},
		{"not-a-url", false},
		{"", false},
		{"ftp://example.com/file", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsURL(tt.input)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseLinksFromArgs(t *testing.T) {
	args := []string{"https://rapidgator.net/file/abc", "https://mega.nz/file/xyz", "not-a-url"}
	got := ParseArgs(args)
	if len(got) != 2 {
		t.Fatalf("ParseArgs returned %d links, want 2", len(got))
	}
	if got[0] != "https://rapidgator.net/file/abc" {
		t.Errorf("got[0] = %q, want %q", got[0], "https://rapidgator.net/file/abc")
	}
	if got[1] != "https://mega.nz/file/xyz" {
		t.Errorf("got[1] = %q, want %q", got[1], "https://mega.nz/file/xyz")
	}
}

func TestParseLinksFromFile(t *testing.T) {
	content := `# This is a comment
https://example.com/file1

https://example.com/file2
# Another comment

https://example.com/file3
not-a-url
`
	dir := t.TempDir()
	path := filepath.Join(dir, "links.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("ParseFile returned %d links, want 3", len(got))
	}
}

func TestParseLinksFromString(t *testing.T) {
	input := `# comment
https://example.com/a
some random text
http://example.com/b

magnet:?xt=urn:btih:abc
not-a-url
ftp://nope.com
https://example.com/c
`
	got := ParseString(input)
	if len(got) != 4 {
		t.Fatalf("ParseString returned %d links, want 4", len(got))
	}
}

func TestDeduplicateLinks(t *testing.T) {
	input := []string{
		"https://example.com/a",
		"https://example.com/b",
		"https://example.com/a",
		"https://example.com/b",
	}
	got := Deduplicate(input)
	if len(got) != 2 {
		t.Fatalf("Deduplicate returned %d links, want 2", len(got))
	}
	if got[0] != "https://example.com/a" {
		t.Errorf("got[0] = %q, want %q", got[0], "https://example.com/a")
	}
	if got[1] != "https://example.com/b" {
		t.Errorf("got[1] = %q, want %q", got[1], "https://example.com/b")
	}
}
