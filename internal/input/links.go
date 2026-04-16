package input

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsURL checks whether a string is a supported URL (http, https, or magnet).
func IsURL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	return strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "magnet:")
}

// ParseArgs filters CLI arguments to only valid URLs.
func ParseArgs(args []string) []string {
	var links []string
	for _, a := range args {
		a = strings.TrimSpace(a)
		if IsURL(a) {
			links = append(links, a)
		}
	}
	return links
}

// ParseFile reads a file and parses its contents for valid URLs.
func ParseFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading link file: %w", err)
	}
	return ParseString(string(data)), nil
}

// ParseString extracts valid URLs from a string, line by line.
// Blank lines and lines starting with # are skipped.
func ParseString(input string) []string {
	var links []string
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if IsURL(line) {
			links = append(links, line)
		}
	}
	return links
}

// ParseClipboard reads the system clipboard via pbpaste and parses for URLs.
func ParseClipboard() ([]string, error) {
	out, err := exec.Command("pbpaste").Output()
	if err != nil {
		return nil, fmt.Errorf("reading clipboard: %w", err)
	}
	return ParseString(string(out)), nil
}

// ParseStdin reads all of stdin and parses for URLs.
func ParseStdin() ([]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var sb strings.Builder
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	return ParseString(sb.String()), nil
}

// Deduplicate removes duplicate links while preserving order.
func Deduplicate(links []string) []string {
	seen := make(map[string]struct{}, len(links))
	var result []string
	for _, l := range links {
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		result = append(result, l)
	}
	return result
}

// IsTerminal returns true if stdin is connected to a terminal (TTY).
func IsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
