package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/oleksbard/cosmobar/internal/segments"
)

func segmentsList() string {
	var b strings.Builder
	b.WriteString("Available segments (★ = on by default):\n")
	for _, m := range segments.Catalog() {
		star := " "
		if m.DefaultOn {
			star = "★"
		}
		var tags []string
		if m.RequiresGit {
			tags = append(tags, "needs-git")
		}
		if m.ProMaxOnly {
			tags = append(tags, "pro/max")
		}
		tag := ""
		if len(tags) > 0 {
			tag = "  [" + strings.Join(tags, ", ") + "]"
		}
		fmt.Fprintf(&b, "  %s %-13s %s%s\n", star, m.Name, m.Description, tag)
	}
	return b.String()
}

// segmentsJSON returns the segment catalog as indented JSON.
func segmentsJSON() string {
	out, err := json.MarshalIndent(segments.Catalog(), "", "  ")
	if err != nil {
		return "[]"
	}
	return string(out)
}

func cmdSegments(args []string) int {
	fs := flag.NewFlagSet("segments", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "output the segment catalog as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *asJSON {
		fmt.Println(segmentsJSON())
		return 0
	}
	fmt.Print(segmentsList())
	return 0
}
