package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeyfurness/rdl/internal/config"
	"github.com/joeyfurness/rdl/internal/store"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show download history",
	Long:  `Display recent download history. Use --failed to show only failed downloads.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		failed, _ := cmd.Flags().GetBool("failed")

		dir := config.ConfigDir()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}
		dbPath := filepath.Join(dir, "queue.db")
		db, err := store.Open(dbPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		var entries []store.HistoryEntry
		if failed {
			entries, err = db.ListFailed()
		} else {
			entries, err = db.ListHistory(0, 50)
		}
		if err != nil {
			return fmt.Errorf("listing history: %w", err)
		}

		if len(entries) == 0 {
			if failed {
				fmt.Println("no failed downloads")
			} else {
				fmt.Println("no download history")
			}
			return nil
		}

		for _, e := range entries {
			name := e.Filename
			if name == "" {
				name = e.OriginalLink
			}
			size := formatBytes(e.Filesize)
			line := fmt.Sprintf("[%s] %s (%s)", e.Status, name, size)
			if e.ErrorMsg != "" {
				line += " - " + e.ErrorMsg
			}
			fmt.Println(line)
		}
		return nil
	},
}

// formatBytes returns a human-readable byte string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), suffixes[exp])
}

func init() {
	historyCmd.Flags().Bool("failed", false, "show only failed downloads")
	rootCmd.AddCommand(historyCmd)
}
