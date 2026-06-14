---
name: install
description: Install the cosmobar status line for Claude Code — downloads the binary and runs guided setup. Use when the user wants to install or set up cosmobar for the first time (when it is not yet installed).
---

# Install cosmobar

Install the cosmobar status line and walk the user through setup. cosmobar is a single static Go binary (macOS/Linux). Follow these steps in order.

## Steps

1. **Check whether it's already installed.** Run `command -v cosmobar`.
   - If found: tell the user cosmobar is already installed, run `cosmobar upgrade --check` to report whether a newer release exists, and skip to step 5 (offer to configure). Do **not** re-download.

2. **Install the binary.** Run the official installer:
   ```
   curl -sS https://raw.githubusercontent.com/oleksbard/cosmobar/main/install.sh | sh
   ```
   The user will see the normal Bash permission prompt for this command — that is expected and intentional. Show the installer's output. If the download fails (network error, non-zero exit), report the failure and give the user the same `curl … | sh` command to run themselves, then stop.

3. **Verify the install.** Run `cosmobar --version`.
   - If it prints a version, continue.
   - If it reports "command not found", the binary was installed to `~/.local/bin`, which is not on the user's `PATH`. Tell the user to add it (`export PATH="$HOME/.local/bin:$PATH"` in their shell profile) and, for the next step, use the full path `~/.local/bin/cosmobar`.

4. **Initialize.** Run `cosmobar init` (or `~/.local/bin/cosmobar init` if it wasn't on `PATH`). This wires `~/.claude/settings.json`, writes a default config to `~/.config/cosmobar/config.toml`, and installs the `/cosmobar` guided-setup skill.

5. **Hand off to guided setup.** The `/cosmobar` skill is now available. Continue directly into the cosmobar guided configuration (discover segments and themes, interview the user, apply) so install and configuration happen in one flow.

## Notes
- macOS/Linux only.
- Everything is reversible: `cosmobar uninstall` removes the wiring (add `--purge` to also delete the config and binary).
