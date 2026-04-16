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

// BuildAria2cArgs constructs the full argument list for an aria2c invocation.
func BuildAria2cArgs(params DownloadParams, urls []string, outputDir string) []string {
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
		"--lowest-speed-limit=1M",
		"--max-overall-upload-limit=0",
		"--check-certificate=true",
		"--console-log-level=notice",
		"--summary-interval=1",
	}
	args = append(args, urls...)
	return args
}

// RunAria2cRaw executes aria2c as a subprocess with stdin, stdout, and stderr
// connected directly to the parent process.
func RunAria2cRaw(params DownloadParams, urls []string, outputDir string) error {
	args := BuildAria2cArgs(params, urls, outputDir)
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
