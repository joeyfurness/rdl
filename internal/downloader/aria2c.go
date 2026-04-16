package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// IsAria2cInstalled reports whether the aria2c binary is available on PATH.
func IsAria2cInstalled() bool {
	_, err := exec.LookPath("aria2c")
	return err == nil
}

// BuildAria2cArgs constructs the base argument list for an aria2c invocation.
// URLs are not included — they should be passed via an input file (-i flag).
func BuildAria2cArgs(params DownloadParams, outputDir string) []string {
	args := []string{
		"-d", outputDir,
		"-j", strconv.Itoa(params.ConcurrentFiles),
		"-x", strconv.Itoa(params.ConnectionsPerFile),
		"-s", strconv.Itoa(params.ConnectionsPerFile),
		"-k", params.PieceSize,
		"--file-allocation=" + params.FileAllocation,
		"--continue=true",
		"--auto-file-renaming=false",
		"--max-tries=5",
		"--retry-wait=15",
		"--max-connection-per-server=" + strconv.Itoa(params.ConnectionsPerFile),
		"--connect-timeout=30",
		"--timeout=600",
		"--lowest-speed-limit=100K",
		"--max-overall-upload-limit=0",
		"--check-certificate=true",
		"--console-log-level=notice",
		"--summary-interval=1",
	}
	return args
}

// RunAria2cRaw executes aria2c as a subprocess with stdin, stdout, and stderr
// connected directly to the parent process. URLs are passed via an input file
// (-i flag) so each URL is treated as a separate download.
func RunAria2cRaw(params DownloadParams, urls []string, outputDir string) error {
	inputFile, err := WriteInputFile(urls)
	if err != nil {
		return err
	}
	defer os.Remove(inputFile)

	args := BuildAria2cArgs(params, outputDir)
	args = append(args, "-i", inputFile)

	cmd := exec.Command("aria2c", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WriteInputFile writes the given URLs to a temporary file suitable for use
// with aria2c's -i flag. The caller is responsible for removing the file.
func WriteInputFile(urls []string) (string, error) {
	f, err := os.CreateTemp("", "rdl-aria2c-input-*.txt")
	if err != nil {
		return "", fmt.Errorf("creating aria2c input file: %w", err)
	}
	defer f.Close()

	for _, u := range urls {
		if _, err := fmt.Fprintln(f, u); err != nil {
			os.Remove(f.Name())
			return "", fmt.Errorf("writing aria2c input file: %w", err)
		}
	}
	return f.Name(), nil
}
