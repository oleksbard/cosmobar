package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/oleksbard/cosmobar/internal/theme"
)

func themesList() string {
	var b strings.Builder
	b.WriteString("Available themes:\n")
	for _, name := range theme.Names() {
		fmt.Fprintf(&b, "  %s\n", name)
	}
	b.WriteString("\nPreview one with: cosmobar preview --theme <name>\n")
	return b.String()
}

func cmdThemes(args []string) int {
	fs := flag.NewFlagSet("themes", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "output theme names as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *asJSON {
		out, err := json.MarshalIndent(theme.Names(), "", "  ")
		if err != nil {
			return 1
		}
		fmt.Println(string(out))
		return 0
	}
	fmt.Print(themesList())
	return 0
}
