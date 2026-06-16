package cli

import (
	"fmt"
	"os"
)

// version holds the running binary's version. It is injected at the root via
// ldflags (-X main.version) and threaded in through Run, so the upgrade command
// and the --version flag can report it.
var version = "dev"

// Run dispatches the first argument to a command and returns the process exit
// code. The default command (no args, or a flag-prefixed first arg) is print.
func Run(v string, args []string) int {
	version = v
	// version flags anywhere as the first arg
	if len(args) > 0 {
		switch args[0] {
		case "-v", "--version", "version":
			fmt.Println("cosmobar " + version)
			return 0
		}
	}
	cmd := "print"
	rest := args
	if len(args) > 0 && args[0][0] != '-' {
		cmd = args[0]
		rest = args[1:]
	}
	switch cmd {
	case "print":
		return cmdPrint(rest)
	case "preview":
		return cmdPreview(rest)
	case "init":
		return cmdInit(rest)
	case "uninstall":
		return cmdUninstall(rest)
	case "segments":
		return cmdSegments(rest)
	case "install-skill":
		return cmdInstallSkill(rest)
	case "doctor":
		return cmdDoctor(rest)
	case "themes":
		return cmdThemes(rest)
	case "upgrade":
		return cmdUpgrade(rest)
	default:
		fmt.Fprintln(os.Stderr, "cosmobar: unknown command:", cmd)
		return 2
	}
}
