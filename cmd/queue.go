package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeyfurness/rdl/internal/config"
	"github.com/joeyfurness/rdl/internal/input"
	"github.com/joeyfurness/rdl/internal/store"
	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Manage the download queue",
	Long:  `Add, list, clear, or run queued download links.`,
}

var queueAddCmd = &cobra.Command{
	Use:   "add [links...]",
	Short: "Add links to the download queue",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openQueueDB()
		if err != nil {
			return err
		}
		defer db.Close()

		links := input.ParseArgs(args)
		if len(links) == 0 {
			return fmt.Errorf("no valid URLs provided")
		}

		for _, link := range links {
			if err := db.QueueAdd(link); err != nil {
				return fmt.Errorf("adding to queue: %w", err)
			}
			fmt.Printf("queued: %s\n", link)
		}
		return nil
	},
}

var queueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all queued links",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openQueueDB()
		if err != nil {
			return err
		}
		defer db.Close()

		items, err := db.QueueList()
		if err != nil {
			return fmt.Errorf("listing queue: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("queue is empty")
			return nil
		}

		for _, item := range items {
			fmt.Printf("%s  %s\n", item.AddedAt.Format("2006-01-02 15:04"), item.Link)
		}
		return nil
	},
}

var queueClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all links from the queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openQueueDB()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.QueueClear(); err != nil {
			return fmt.Errorf("clearing queue: %w", err)
		}
		fmt.Println("queue cleared")
		return nil
	},
}

var queueRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Download all queued links and clear the queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openQueueDB()
		if err != nil {
			return err
		}
		defer db.Close()

		items, err := db.QueueList()
		if err != nil {
			return fmt.Errorf("listing queue: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("queue is empty, nothing to run")
			return nil
		}

		var links []string
		for _, item := range items {
			links = append(links, item.Link)
		}

		if err := db.QueueClear(); err != nil {
			return fmt.Errorf("clearing queue: %w", err)
		}

		return runDownload(cmd, links)
	},
}

func openQueueDB() (*store.Store, error) {
	dir := config.ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating config directory: %w", err)
	}
	dbPath := filepath.Join(dir, "queue.db")
	db, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening queue database: %w", err)
	}
	return db, nil
}

func init() {
	queueCmd.AddCommand(queueAddCmd)
	queueCmd.AddCommand(queueListCmd)
	queueCmd.AddCommand(queueClearCmd)
	queueCmd.AddCommand(queueRunCmd)
	rootCmd.AddCommand(queueCmd)
}
