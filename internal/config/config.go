package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the top-level configuration for rdl.
type Config struct {
	Download DownloadConfig `toml:"download"`
	Output   OutputConfig   `toml:"output"`
	Behavior BehaviorConfig `toml:"behavior"`
}

// DownloadConfig holds download-related settings.
type DownloadConfig struct {
	Directory  string `toml:"directory"`
	SpeedTier  string `toml:"speed_tier"`
	MaxRetries int    `toml:"max_retries"`
}

// OutputConfig holds output display settings.
type OutputConfig struct {
	Mode string `toml:"mode"`
}

// BehaviorConfig holds behavioral settings.
type BehaviorConfig struct {
	Overwrite string `toml:"overwrite"`
}

// DefaultConfig returns a Config populated with default values.
func DefaultConfig() Config {
	return Config{
		Download: DownloadConfig{
			Directory:  "~/Downloads",
			SpeedTier:  "auto",
			MaxRetries: 3,
		},
		Output: OutputConfig{
			Mode: "auto",
		},
		Behavior: BehaviorConfig{
			Overwrite: "resume",
		},
	}
}

// ConfigDir returns the configuration directory path.
// It respects the RDL_CONFIG_DIR environment variable.
func ConfigDir() string {
	if dir := os.Getenv("RDL_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "rdl")
}

// ConfigPath returns the path to the config.toml file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// AuthPath returns the path to the auth.json file.
func AuthPath() string {
	return filepath.Join(ConfigDir(), "auth.json")
}

// Load reads the configuration from disk. If the config file does not exist,
// it creates one with default values.
func Load() (*Config, error) {
	path := ConfigPath()

	// If file doesn't exist, create with defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(&cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	cfg := DefaultConfig()
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the configuration to disk as TOML.
func Save(cfg *Config) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

// ExpandPath expands a leading ~ to the user's home directory.
func ExpandPath(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
