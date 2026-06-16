package main

import (
	"os"

	"github.com/oleksbard/cosmobar/internal/cli"
)

// version is injected at build time via ldflags (-X main.version) and handed to
// the CLI dispatcher. It defaults to "dev" for local builds.
var version = "dev"

func main() {
	os.Exit(cli.Run(version, os.Args[1:]))
}
