package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joeyfurness/rdl/internal/api"
	"github.com/spf13/cobra"
)

var torrentCmd = &cobra.Command{
	Use:   "torrent",
	Short: "Manage Real-Debrid torrents",
	Long:  `Add, list, select files, and download torrents via Real-Debrid.`,
}

var torrentAddCmd = &cobra.Command{
	Use:   "add <magnet>",
	Short: "Add a magnet link to Real-Debrid",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		magnet := args[0]
		if !strings.HasPrefix(magnet, "magnet:") {
			return fmt.Errorf("invalid magnet link: must start with 'magnet:'")
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		resp, err := api.AddMagnet(client, magnet)
		if err != nil {
			return fmt.Errorf("adding magnet: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Torrent added: %s\n", resp.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "\nNext steps:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Select files:  rdl torrent select %s all\n", resp.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "  Check status:  rdl torrent list\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Download:      rdl torrent download %s\n", resp.ID)
		return nil
	},
}

var torrentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your Real-Debrid torrents",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}

		torrents, err := api.ListTorrents(client)
		if err != nil {
			return fmt.Errorf("listing torrents: %w", err)
		}

		if len(torrents) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No torrents found.")
			return nil
		}

		for _, t := range torrents {
			status := api.FormatTorrentStatus(t.Status, t.Progress)
			fmt.Fprintf(cmd.OutOrStdout(), "%-12s  %-25s  %s\n", t.ID, status, t.Filename)
		}
		return nil
	},
}

var torrentSelectCmd = &cobra.Command{
	Use:   "select <id> [file_ids]",
	Short: "Select files to download from a torrent",
	Long: `Select which files to download from a torrent.
Provide a comma-separated list of file IDs, or "all" to select all files.
If no file IDs are given, defaults to "all".`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		fileIDs := "all"
		if len(args) > 1 {
			fileIDs = args[1]
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := api.SelectTorrentFiles(client, id, fileIDs); err != nil {
			return fmt.Errorf("selecting files: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Files selected for torrent %s\n", id)
		return nil
	},
}

var torrentDownloadCmd = &cobra.Command{
	Use:   "download <id>",
	Short: "Download a completed torrent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		client, err := getClient()
		if err != nil {
			return err
		}

		info, err := api.GetTorrentInfo(client, id)
		if err != nil {
			return fmt.Errorf("getting torrent info: %w", err)
		}

		if info.Status != "downloaded" {
			status := api.FormatTorrentStatus(info.Status, info.Progress)
			return fmt.Errorf("torrent is not ready for download (status: %s)", status)
		}

		if len(info.Links) == 0 {
			return fmt.Errorf("torrent has no download links")
		}

		return runDownload(cmd, info.Links)
	},
}

func init() {
	rootCmd.AddCommand(torrentCmd)
	torrentCmd.AddCommand(torrentAddCmd)
	torrentCmd.AddCommand(torrentListCmd)
	torrentCmd.AddCommand(torrentSelectCmd)
	torrentCmd.AddCommand(torrentDownloadCmd)
}

// getClient creates an authenticated API client.
// This is a temporary helper that will be replaced during integration.
func getClient() (*api.Client, error) {
	token := os.Getenv("RDL_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("not authenticated. Run: rdl auth login")
	}
	return api.NewClient(token), nil
}
