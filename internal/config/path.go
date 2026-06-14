package config

import (
	"os"
	"path/filepath"
)

// DefaultPath resolves the config file location, honoring COSMOBAR_CONFIG and
// XDG_CONFIG_HOME, defaulting to ~/.config/cosmobar/config.toml.
func DefaultPath() string {
	if p := os.Getenv("COSMOBAR_CONFIG"); p != "" {
		return p
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "cosmobar.toml"
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "cosmobar", "config.toml")
}
