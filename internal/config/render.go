package config

import (
	"fmt"
	"strings"
)

// RenderTOML serializes a Config into a commented TOML document suitable for
// writing to disk. It is deterministic, so `cosmobar init` can generate a
// config from chosen options without hand-formatting TOML.
func RenderTOML(c Config) string {
	q := func(s string) string { return "\"" + s + "\"" }

	items := make([]string, len(c.Order))
	for i, s := range c.Order {
		items[i] = q(s)
	}
	th := c.GaugeThresholds
	if len(th) < 2 {
		th = []int{70, 90}
	}

	var b strings.Builder
	b.WriteString("# cosmobar config — https://github.com/oleksbard/cosmobar\n")
	fmt.Fprintf(&b, "theme            = %s\n", q(c.Theme))
	fmt.Fprintf(&b, "order            = [%s]\n", strings.Join(items, ", "))
	fmt.Fprintf(&b, "separator        = %s\n", q(c.Separator))
	fmt.Fprintf(&b, "max_rows         = %d\n", c.MaxRows)
	fmt.Fprintf(&b, "gauge_width      = %d\n", c.GaugeWidth)
	fmt.Fprintf(&b, "gauge_thresholds = [%d, %d]\n", th[0], th[1])
	fmt.Fprintf(&b, "glyphs           = %s\n", q(c.Glyphs))
	fmt.Fprintf(&b, "style            = %s\n", q(c.Style))
	fmt.Fprintf(&b, "block_caps       = %s\n\n", q(c.BlockCaps))
	fmt.Fprintf(&b, "[clock]\nformat = %s\n\n", q(c.Clock.Format))
	fmt.Fprintf(&b, "[dir]\nstyle = %s\n\n", q(c.Dir.Style))
	fmt.Fprintf(&b, "[context]\nshow = %t\n\n", c.Context.Show)
	fmt.Fprintf(&b, "[rate_limits]\nshow = %t\nwindow = %s\n", c.RateLimits.Show, q(c.RateLimits.Window))
	return b.String()
}
