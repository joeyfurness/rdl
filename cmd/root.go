package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rdl [links...]",
	Short: "Real-Debrid download CLI",
	Long: `rdl automates downloading files through Real-Debrid.

Provide one or more links as arguments, from a file, or from the clipboard.
Links are unrestricted via the Real-Debrid API and downloaded with aria2c.`,
	RunE: runDownload,
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
	fmt.Println("rdl: download command not yet implemented")
	return nil
}
