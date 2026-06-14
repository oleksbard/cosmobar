// Package config defines cosmobar's configuration and loads it from TOML.
package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

type Toggle struct {
	Show bool `toml:"show"`
}

type ClockConfig struct {
	Format string `toml:"format"` // 24h | 12h | off
}

type DirConfig struct {
	Style string `toml:"style"` // basename | short-path | full
}

type Config struct {
	Theme           string      `toml:"theme"`
	Order           []string    `toml:"order"`
	Separator       string      `toml:"separator"`
	MaxRows         int         `toml:"max_rows"`
	GaugeWidth      int         `toml:"gauge_width"`
	GaugeThresholds []int       `toml:"gauge_thresholds"`
	Glyphs          string      `toml:"glyphs"` // auto | unicode | ascii
	Clock           ClockConfig `toml:"clock"`
	Dir             DirConfig   `toml:"dir"`
	Context         Toggle      `toml:"context"`
	RateLimits      Toggle      `toml:"rate_limits"`
}

// Default returns the built-in configuration used when no file is present
// and as the base that file values are merged over.
func Default() Config {
	return Config{
		Theme:           "coral",
		Order:           []string{"dir", "git", "model", "context", "cost", "clock"},
		Separator:       " · ",
		MaxRows:         2,
		GaugeWidth:      8,
		GaugeThresholds: []int{70, 90},
		Glyphs:          "auto",
		Clock:           ClockConfig{Format: "24h"},
		Dir:             DirConfig{Style: "basename"},
		Context:         Toggle{Show: true},
		RateLimits:      Toggle{Show: false},
	}
}

// Thresholds returns the (warn, crit) gauge thresholds, falling back to 70/90.
func (c Config) Thresholds() (int, int) {
	if len(c.GaugeThresholds) >= 2 {
		return c.GaugeThresholds[0], c.GaugeThresholds[1]
	}
	return 70, 90
}

// ASCII reports whether bars/symbols should use ASCII-only characters.
func (c Config) ASCII() bool { return c.Glyphs == "ascii" }

// Load returns Default() merged with any values found in the TOML file at
// path. A missing file yields defaults with no error; a malformed file
// returns the partially-applied config and the parse error.
func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
