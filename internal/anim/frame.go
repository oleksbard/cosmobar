// Package anim renders value-change "scramble" animations for statusline
// segments and tracks per-session animation state across invocations.
package anim

import (
	"encoding/binary"
	"hash/fnv"

	"github.com/oleksbard/cosmobar/internal/render"
)

// steps is how many times the glitch characters re-roll across a full
// animation. Higher = busier flicker.
const steps = 7

// variantOrder is the canonical pool used when config supplies none.
var variantOrder = []string{"decode", "glitch", "scatter"}

// palette holds the unicode and ascii glitch character sets per variant.
type palette struct{ unicode, ascii string }

var palettes = map[string]palette{
	"decode":  {unicode: "#%&$*+=?/<>~", ascii: "#%&$*+=?/<>~"},
	"glitch":  {unicode: "░▒▓█▞▚▙▟", ascii: "#@%&*+=:"},
	"scatter": {unicode: `@#%&*!?+=/\`, ascii: `@#%&*!?+=/\`},
}

// Frame renders target at the given progress in [0,1] using the named variant.
// progress>=1 returns target exactly. seed makes the random-looking glitch
// characters and (for "glitch") the lock order deterministic. ascii forces an
// ASCII-only glitch palette. Width is always preserved: only single-width,
// non-space cells scramble; spaces and wide runes stay locked to their final
// glyph.
func Frame(target string, progress float64, variant string, seed uint64, ascii bool) string {
	if progress >= 1 {
		return target
	}
	if progress < 0 {
		progress = 0
	}
	runes := []rune(target)
	out := make([]rune, len(runes))
	copy(out, runes) // default: final glyphs (covers spaces, wide runes, locked cells)

	eligible := eligibleCells(runes)
	pal := paletteRunes(variant, ascii)
	bucket := frameBucket(progress)

	if variant == "scatter" {
		for _, i := range eligible {
			out[i] = pal[glitchIndex(seed, i, bucket, len(pal))]
		}
		return string(out)
	}

	// decode / glitch: progressively lock cells (final), glitch the rest.
	order := make([]int, len(eligible))
	copy(order, eligible)
	if variant == "glitch" {
		shuffle(order, seed)
	}
	// Round to nearest: exactly 0 locked at progress 0, exactly len(order) at
	// progress 1, so cells reveal smoothly across the sweep.
	locked := int(float64(len(order))*progress + 0.5)
	isLocked := make([]bool, len(runes)) // dense rune indices → plain bool slice
	for k := 0; k < locked && k < len(order); k++ {
		isLocked[order[k]] = true
	}
	for _, i := range eligible {
		if !isLocked[i] {
			out[i] = pal[glitchIndex(seed, i, bucket, len(pal))]
		}
	}
	return string(out)
}

// eligibleCells returns indices that may scramble: non-space cells of display
// width 1, so total width never changes. cosmobar's render.Width counts runes
// (it assumes single-width runes throughout), so today this excludes only
// spaces; the width!=1 guard also keeps wide runes locked should render.Width
// ever become display-width aware.
func eligibleCells(runes []rune) []int {
	var idx []int
	for i, r := range runes {
		if r != ' ' && render.Width(string(r)) == 1 {
			idx = append(idx, i)
		}
	}
	return idx
}

func paletteRunes(variant string, ascii bool) []rune {
	p, ok := palettes[variant]
	if !ok {
		p = palettes["decode"]
	}
	if ascii {
		return []rune(p.ascii)
	}
	return []rune(p.unicode)
}

func frameBucket(progress float64) int {
	b := int(progress * float64(steps))
	if b >= steps {
		b = steps - 1
	}
	return b
}

// glitchIndex deterministically chooses a palette index for (cell, bucket).
func glitchIndex(seed uint64, cell, bucket, n int) int {
	if n <= 0 {
		return 0
	}
	return int(mix(seed, uint64(cell)*0x9e3779b1+uint64(bucket)*0x85ebca77) % uint64(n))
}

// mix hashes two values into one (a small avalanche function).
func mix(a, b uint64) uint64 {
	x := a ^ (b + 0x9e3779b97f4a7c15 + (a << 6) + (a >> 2))
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	return x
}

// shuffle permutes order deterministically using a seeded LCG (Fisher–Yates).
func shuffle(order []int, seed uint64) {
	s := seed | 1
	for i := len(order) - 1; i > 0; i-- {
		s = s*6364136223846793005 + 1442695040888963407
		j := int((s >> 33) % uint64(i+1))
		order[i], order[j] = order[j], order[i]
	}
}

// seedFor derives a stable per-animation seed.
func seedFor(name, target string, startMs int64) uint64 {
	h := fnv.New64a()
	h.Write([]byte(name))
	h.Write([]byte{'|'})
	h.Write([]byte(target))
	h.Write([]byte{'|'})
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(startMs))
	h.Write(buf[:])
	return h.Sum64()
}

// pickVariant chooses a variant from the pool using seed.
func pickVariant(pool []string, seed uint64) string {
	if len(pool) == 0 {
		pool = variantOrder
	}
	return pool[seed%uint64(len(pool))]
}
