package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/settings"
)

// cmdUninstall is the inverse of init: it removes the statusLine block from
// settings.json. With --purge it also removes the config file and the binary.
func cmdUninstall(args []string) int {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	purge := fs.Bool("purge", false, "also remove the config file and the cosmobar binary")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	sp := settings.Path()
	removed, err := settings.UnwireStatusLine(sp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: failed to update settings:", err)
		return 1
	}
	if removed {
		fmt.Println("cosmobar: removed statusLine from", sp, "(backup:", sp+".bak)")
	} else {
		fmt.Println("cosmobar: no statusLine entry found in", sp, "(nothing to unwire)")
	}

	if *purge {
		cp := config.DefaultPath()
		if err := os.Remove(cp); err == nil {
			fmt.Println("cosmobar: removed config", cp)
			os.Remove(filepath.Dir(cp)) // succeeds only if the dir is now empty
		}
		if bin, err := os.Executable(); err == nil {
			if resolved, err := filepath.EvalSymlinks(bin); err == nil {
				bin = resolved
			}
			if err := os.Remove(bin); err == nil {
				fmt.Println("cosmobar: removed binary", bin)
			} else {
				fmt.Fprintln(os.Stderr, "cosmobar: could not remove binary", bin+":", err)
			}
		}
	}

	fmt.Println("cosmobar: done. Restart Claude Code (or send a message) to clear the bar.")
	return 0
}
