package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed assets/skill/SKILL.md
var setupSkill string

// installSkill writes the setup skill into ~/.claude/skills so it is
// available as the /cosmobar guided installer inside Claude Code.
func installSkill() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".claude", "skills", "cosmobar")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(p, []byte(setupSkill), 0o644); err != nil {
		return "", err
	}
	return p, nil
}

func cmdInstallSkill(_ []string) int {
	p, err := installSkill()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: failed to install skill:", err)
		return 1
	}
	fmt.Println("cosmobar: installed setup skill →", p)
	fmt.Println("cosmobar: in Claude Code, run /cosmobar (or just ask Claude to set up cosmobar).")
	return 0
}
