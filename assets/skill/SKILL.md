---
name: cosmobar
description: Interactively configure the cosmobar Claude Code statusline — discovers the available segments dynamically, asks the user which to show plus theme and clock, then writes the config and wires it into settings.json. Use when the user wants to set up, install, configure, customize, or reconfigure cosmobar (the statusline).
---

# cosmobar guided setup

Configure cosmobar by interviewing the user, then apply the result with one command. **Do not hardcode the list of segments or themes — read them live from the binary so this stays correct as cosmobar gains new segments.**

## Steps

1. **Confirm the binary is installed.** Run `cosmobar --version`.
   - If you get "command not found", tell the user to install it first:
     `curl -sS https://raw.githubusercontent.com/oleksbard/cosmobar/main/install.sh | sh`
     then stop and let them install.

2. **Discover what's available (dynamic — always run these, never assume):**
   - `cosmobar segments --json` → array of segments, each with `name`, `description`, `default_on`, `requires_git`, `pro_max_only`.
   - `cosmobar themes --json` → array of theme names.

3. **Interview the user** with the AskUserQuestion tool. Keep it short — group related choices:
   - **Segments:** present *every* segment returned by the JSON, using its `description`. Pre-select the ones where `default_on` is true. Mention that `pro_max_only` segments only show data on Pro/Max plans, and `requires_git` segments only appear inside a git repo. Let the user pick the set to show. The order they end up with is the left-to-right display order.
   - **Theme:** offer the names from the themes JSON (default `coral`).
   - **Clock:** `24h`, `12h`, or `off`.
   - **Glyphs (optional):** `auto` (default) or `ascii` — pick `ascii` only if their terminal can't render block characters like `▓░`.

4. **Apply everything with one command.** This writes `~/.config/cosmobar/config.toml` AND wires `~/.claude/settings.json` (backing up the previous file):
   ```
   cosmobar init --force --theme <THEME> --order <comma,separated,enabled,segments> --clock <24h|12h|off> --glyphs <auto|ascii>
   ```
   - `--order` is the comma-separated list of the segments the user enabled, in their chosen order.
   - `--force` overwrites any existing config (this is a reconfigure).

5. **Show the result.** Run `cosmobar preview` and show the user the rendered status line. Tell them it will appear at the bottom of Claude Code on their next message, that they can re-run `/cosmobar` anytime, or hand-edit `~/.config/cosmobar/config.toml` and preview again with `cosmobar preview`.

## Notes
- Everything is reversible: `cosmobar uninstall` removes the statusline wiring (add `--purge` to also delete the config and the binary).
- `cosmobar doctor` reports whether the binary, config, color support, and wiring are all healthy.
