package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joeyfurness/rdl/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage rdl configuration",
	Long:  `View, edit, or query rdl configuration values.`,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the config file in your editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			editor = "nano"
		}

		path := config.ConfigPath()

		// Ensure the config file exists
		if _, err := config.Load(); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		key := args[0]
		switch key {
		case "download.directory":
			fmt.Println(cfg.Download.Directory)
		case "download.speed_tier":
			fmt.Println(cfg.Download.SpeedTier)
		case "download.max_retries":
			fmt.Println(cfg.Download.MaxRetries)
		case "output.mode":
			fmt.Println(cfg.Output.Mode)
		case "behavior.overwrite":
			fmt.Println(cfg.Behavior.Overwrite)
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file path",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.ConfigPath())
	},
}

func init() {
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
