// Package theme defines named color palettes for cosmobar.
package theme

import "sort"

type RGB struct{ R, G, B uint8 }

// Palette holds the colors a theme applies to segment text and gauges.
type Palette struct {
	Accent RGB // primary segment text
	Muted  RGB // separators / secondary text
	OK     RGB // gauge low
	Warn   RGB // gauge mid
	Crit   RGB // gauge high
}

var palettes = map[string]Palette{
	"coral": {
		Accent: RGB{0xff, 0x7e, 0x6b}, Muted: RGB{0x8a, 0x8a, 0x8a},
		OK: RGB{0x8e, 0xc0, 0x7c}, Warn: RGB{0xe5, 0xc0, 0x7b}, Crit: RGB{0xe0, 0x6c, 0x75},
	},
	"nord": {
		Accent: RGB{0x88, 0xc0, 0xd0}, Muted: RGB{0x4c, 0x56, 0x6a},
		OK: RGB{0xa3, 0xbe, 0x8c}, Warn: RGB{0xeb, 0xcb, 0x8b}, Crit: RGB{0xbf, 0x61, 0x6a},
	},
	"catppuccin": {
		Accent: RGB{0xf5, 0xc2, 0xe7}, Muted: RGB{0x6c, 0x70, 0x86},
		OK: RGB{0xa6, 0xe3, 0xa1}, Warn: RGB{0xf9, 0xe2, 0xaf}, Crit: RGB{0xf3, 0x8b, 0xa8},
	},
	"gruvbox": {
		Accent: RGB{0xfe, 0x80, 0x19}, Muted: RGB{0x92, 0x83, 0x74},
		OK: RGB{0xb8, 0xbb, 0x26}, Warn: RGB{0xfa, 0xbd, 0x2f}, Crit: RGB{0xfb, 0x49, 0x34},
	},
}

// Get returns the named palette and whether it exists.
func Get(name string) (Palette, bool) {
	p, ok := palettes[name]
	return p, ok
}

// Names returns the sorted list of built-in theme names.
func Names() []string {
	out := make([]string, 0, len(palettes))
	for k := range palettes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
