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

type RateLimitsConfig struct {
	Show          bool   `toml:"show"`
	Window        string `toml:"window"` // both | 5h | 7d
	ShowBlockCost bool   `toml:"show_block_cost"`
}

type CostConfig struct {
	Rollups []string `toml:"rollups"` // any of: today | week | month
}

type AnimationConfig struct {
	Enabled    bool     `toml:"enabled"`
	DurationMs int      `toml:"duration_ms"`
	Variants   []string `toml:"variants"` // pool to pick randomly from: decode | glitch | scatter
}

type Config struct {
	Theme           string           `toml:"theme"`
	Order           []string         `toml:"order"`
	Separator       string           `toml:"separator"`
	MaxRows         int              `toml:"max_rows"`
	GaugeWidth      int              `toml:"gauge_width"`
	GaugeThresholds []int            `toml:"gauge_thresholds"`
	Glyphs          string           `toml:"glyphs"`     // auto | unicode | ascii
	Style           string           `toml:"style"`      // lean | tick | blocks
	BlockCaps       string           `toml:"block_caps"` // soft | square
	Clock           ClockConfig      `toml:"clock"`
	Dir             DirConfig        `toml:"dir"`
	Context         Toggle           `toml:"context"`
	RateLimits      RateLimitsConfig `toml:"rate_limits"`
	Cost            CostConfig       `toml:"cost"`
	Animation       AnimationConfig  `toml:"animation"`
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
		Style:           "lean",
		BlockCaps:       "soft",
		Clock:           ClockConfig{Format: "24h"},
		Dir:             DirConfig{Style: "basename"},
		Context:         Toggle{Show: true},
		RateLimits:      RateLimitsConfig{Show: false, Window: "both", ShowBlockCost: true},
		Cost:            CostConfig{Rollups: []string{"today"}},
		Animation:       AnimationConfig{Enabled: true, DurationMs: 700, Variants: []string{"glitch"}},
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

// BackgroundStyle reports whether the active style paints pill backgrounds.
func (c Config) BackgroundStyle() bool {
	return c.Style == "blocks"
}

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
