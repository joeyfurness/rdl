package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeyfurness/rdl/internal/api"
	"github.com/joeyfurness/rdl/internal/config"
	"github.com/joeyfurness/rdl/internal/downloader"
	"github.com/joeyfurness/rdl/internal/input"
	"github.com/joeyfurness/rdl/internal/store"
	"github.com/joeyfurness/rdl/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rdl [links...]",
	Short: "Real-Debrid download CLI",
	Long: `rdl automates downloading files through Real-Debrid.

Provide one or more links as arguments, from a file, or from the clipboard.
Links are unrestricted via the Real-Debrid API and downloaded with aria2c.`,
	Args:                  cobra.ArbitraryArgs,
	DisableFlagParsing:    false,
	RunE:                  runDownload,
	TraverseChildren:      true,
	SilenceUsage:          true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().String("to", "", "download destination directory")
	rootCmd.PersistentFlags().Bool("json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "show what would be done without downloading")
	rootCmd.PersistentFlags().Bool("fast", false, "skip slow operations for speed")
	rootCmd.PersistentFlags().Bool("slow", false, "prefer reliability over speed")
	rootCmd.PersistentFlags().CountP("verbose", "v", "increase verbosity (repeatable)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")

	// Local flags
	rootCmd.Flags().StringP("file", "f", "", "read links from file (- for stdin)")
	rootCmd.Flags().Bool("clip", false, "read links from clipboard")
	rootCmd.Flags().Bool("retry-failed", false, "retry previously failed downloads")
}

func runDownload(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Determine output mode
	jsonFlag, _ := cmd.Flags().GetBool("json")
	quietFlag, _ := cmd.Flags().GetBool("quiet")
	isTTY := input.IsTerminal()
	envMode := os.Getenv("RDL_MODE")
	mode := ui.DetectMode(jsonFlag, quietFlag, envMode, cfg.Output.Mode, isTTY)

	var emitter ui.Emitter
	switch mode {
	case ui.ModeJSON:
		emitter = ui.NewJSONEmitter(os.Stdout)
	case ui.ModeQuiet:
		emitter = ui.NewQuietEmitter(os.Stderr)
	default:
		emitter = ui.NewJSONEmitter(os.Stdout) // fallback until TUI is integrated
	}

	// Open history store (warn if fails, but non-fatal)
	historyPath := filepath.Join(config.ConfigDir(), "history.db")
	db, dbErr := store.Open(historyPath)
	if dbErr != nil {
		fmt.Fprintf(os.Stderr, "rdl: warning: could not open history database: %v\n", dbErr)
	}
	if db != nil {
		defer db.Close()
	}

	var links []string

	retryFailed, _ := cmd.Flags().GetBool("retry-failed")
	if retryFailed {
		if db == nil {
			return fmt.Errorf("cannot retry failed downloads: history database unavailable")
		}
		failed, err := db.ListFailed()
		if err != nil {
			return fmt.Errorf("listing failed downloads: %w", err)
		}
		for _, entry := range failed {
			links = append(links, entry.OriginalLink)
		}
	}

	// Positional args
	links = append(links, input.ParseArgs(args)...)

	// --file flag
	fileFlag, _ := cmd.Flags().GetString("file")
	if fileFlag != "" {
		if fileFlag == "-" {
			stdinLinks, err := input.ParseStdin()
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			links = append(links, stdinLinks...)
		} else {
			fileLinks, err := input.ParseFile(fileFlag)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			links = append(links, fileLinks...)
		}
	}

	// --clip flag
	clipFlag, _ := cmd.Flags().GetBool("clip")
	if clipFlag {
		clipLinks, err := input.ParseClipboard()
		if err != nil {
			return fmt.Errorf("reading clipboard: %w", err)
		}
		links = append(links, clipLinks...)
	}

	// stdin pipe (if not TTY and no other input sources provided)
	if !isTTY && len(links) == 0 && fileFlag == "" {
		stdinLinks, err := input.ParseStdin()
		if err == nil {
			links = append(links, stdinLinks...)
		}
	}

	// Deduplicate all links
	links = input.Deduplicate(links)

	// If no links, print usage hint
	if len(links) == 0 {
		fmt.Fprintln(os.Stderr, "No links provided. Usage:")
		fmt.Fprintln(os.Stderr, "  rdl <link> [link...]")
		fmt.Fprintln(os.Stderr, "  rdl -f links.txt")
		fmt.Fprintln(os.Stderr, "  rdl --clip")
		fmt.Fprintln(os.Stderr, "  echo 'https://...' | rdl")
		return nil
	}

	// Check aria2c early to avoid wasting API calls
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if !dryRun && !downloader.IsAria2cInstalled() {
		return fmt.Errorf("aria2c is not installed. Install it with: brew install aria2")
	}

	client, err := GetAuthenticatedClient()
	if err != nil {
		return fmt.Errorf("authentication: %w", err)
	}

	// Emit start event
	emitter.Emit(ui.Event{
		Type:       "start",
		TotalLinks: len(links),
	})

	// Unrestrict each link
	var downloadURLs []string
	failCount := 0

	for _, link := range links {
		result, err := api.UnrestrictLink(client, link)
		if err != nil {
			emitter.Emit(ui.Event{
				Type:    "error",
				Link:    link,
				Message: err.Error(),
			})
			if db != nil {
				db.AddHistory(store.HistoryEntry{
					OriginalLink: link,
					Status:       store.StatusFailed,
					ErrorMsg:     err.Error(),
				})
			}
			failCount++
			continue
		}

		emitter.Emit(ui.Event{
			Type:     "unrestricted",
			Link:     link,
			Filename: result.Filename,
			Size:     result.Filesize,
		})

		if db != nil {
			db.AddHistory(store.HistoryEntry{
				OriginalLink: link,
				Filename:     result.Filename,
				Filesize:     result.Filesize,
				DownloadURL:  result.Download,
				Status:       store.StatusDownloading,
			})
		}

		downloadURLs = append(downloadURLs, result.Download)
	}

	if dryRun {
		fmt.Fprintf(os.Stderr, "Dry run: would download %d file(s)\n", len(downloadURLs))
		for _, u := range downloadURLs {
			fmt.Println(u)
		}
		return nil
	}

	if len(downloadURLs) == 0 {
		return fmt.Errorf("all links failed to unrestrict")
	}

	// Determine output directory (flag --to > env RDL_OUTPUT_DIR > config)
	outputDir, _ := cmd.Flags().GetString("to")
	if outputDir == "" {
		outputDir = os.Getenv("RDL_OUTPUT_DIR")
	}
	if outputDir == "" {
		outputDir = cfg.Download.Directory
	}
	outputDir = config.ExpandPath(outputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Determine speed tier (--fast/--slow flags > config)
	fastFlag, _ := cmd.Flags().GetBool("fast")
	slowFlag, _ := cmd.Flags().GetBool("slow")

	tier := cfg.Download.SpeedTier
	if fastFlag {
		tier = "fast"
	} else if slowFlag {
		tier = "standard"
	}
	params := downloader.ParamsForTier(tier)

	// Run aria2c
	dlErr := downloader.RunAria2cRaw(params, downloadURLs, outputDir)

	// Emit summary
	succeeded := len(downloadURLs)
	if dlErr != nil {
		// If aria2c failed, we count all as failed
		failCount += succeeded
		succeeded = 0
	}
	emitter.Emit(ui.Event{
		Type:      "summary",
		Total:     len(links),
		Succeeded: succeeded,
		Failed:    failCount,
	})

	// Print failure hint
	if failCount > 0 {
		fmt.Fprintf(os.Stderr, "\n%d download(s) failed. Retry with: rdl --retry-failed\n", failCount)
	}

	if dlErr != nil {
		return fmt.Errorf("aria2c: %w", dlErr)
	}
	if failCount > 0 {
		return fmt.Errorf("%d download(s) failed", failCount)
	}

	return nil
}
