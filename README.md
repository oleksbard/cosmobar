# cosmobar

<p align="center">
  <img src="assets/cosmobar.png" alt="cosmobar mascot" width="160" height="160">
</p>

A fast, dependency-light, starship-inspired status line for [Claude Code](https://code.claude.com).

- Single static Go binary — no runtime, no `jq`, no Nerd Font required
- Themed segments: context gauge, git status, cost, model, clock, and more
- TOML config, instant local preview, one-command self-update

## Install

```sh
curl -sS https://raw.githubusercontent.com/oleksbard/cosmobar/main/install.sh | sh
cosmobar init
```

`init` wires `cosmobar` into `~/.claude/settings.json`, writes a default config
to `~/.config/cosmobar/config.toml`, and installs the guided-setup skill.
Restart Claude Code (or send a message) to see the status line.

### From inside Claude Code

cosmobar also ships as a Claude Code plugin (macOS/Linux):

```sh
/plugin marketplace add oleksbard/cosmobar
/plugin install cosmobar@cosmobar
```

Then say **"install cosmobar"** (or run `/cosmobar:install`) — Claude downloads
the binary, wires it in, and walks you through setup, no shell command to copy.

## Configure

Run **`/cosmobar`** in Claude Code (or just ask "set up cosmobar") for guided
setup, or edit `~/.config/cosmobar/config.toml` directly:

```toml
theme            = "coral"            # coral | catppuccin | nord | gruvbox
order            = ["dir", "git", "model", "context", "cost", "clock"]
separator        = " · "
max_rows         = 2
gauge_width      = 8
gauge_thresholds = [70, 90]
glyphs           = "auto"             # auto | unicode | ascii
style            = "lean"             # lean | tick | blocks
block_caps       = "soft"             # soft | square  (blocks style only)

[clock]
format = "24h"                        # 24h | 12h | off

[dir]
style = "basename"                    # basename | short-path | full

[context]
show = true

[rate_limits]
show   = false                        # Pro/Max only
window = "both"                       # both | 5h | 7d

[animation]
enabled     = true                    # briefly scramble a value when it changes
duration_ms = 700
variants    = ["glitch"]              # glitch | decode | scatter (list to mix)
```

Available segments: `dir`, `git`, `model`, `context`, `cost`, `clock`,
`rate_limits`, `duration`, `lines`, `output_style`, `git_stash`, `effort`.
Add or reorder them in `order`.

Preview any look without launching Claude Code (every flag is optional and
overrides just that field):

```sh
cosmobar preview --theme nord --style blocks --caps soft --order git,model,context,lines
# --cols --theme --style --caps --glyphs --clock --rate-window --order --config
# add --animate to watch value changes scramble
```

## Commands

| Command | What it does |
|---|---|
| `cosmobar` | Render the status line (reads JSON from stdin). |
| `cosmobar init` | Wire into `settings.json`, write config, install the setup skill. Flags: `--theme --order --clock --glyphs --style --caps --rate-window --animate --force --no-skill`. |
| `cosmobar install-skill` | Install the `/cosmobar` guided-setup skill into `~/.claude/skills/`. |
| `cosmobar segments [--json]` | List all available segments. |
| `cosmobar preview` | Render a mock session locally. Flags: `--cols --theme --style --caps --glyphs --clock --rate-window --order --config --animate`. |
| `cosmobar themes` | List built-in themes. |
| `cosmobar doctor` | Offline diagnostics. |
| `cosmobar upgrade [--check]` | Self-update from the latest GitHub Release. |
| `cosmobar uninstall [--purge]` | Remove the `statusLine` block from `settings.json`. `--purge` also deletes the config file and the binary. |

## Updating

```sh
cosmobar upgrade            # download + verify + replace
cosmobar upgrade --check    # just report current vs latest
```

Or re-run the installer — it always fetches the latest release.

## Uninstall

```sh
cosmobar uninstall          # remove the statusLine block from ~/.claude/settings.json
cosmobar uninstall --purge  # also delete ~/.config/cosmobar/ and the binary
```

`uninstall` preserves your other `settings.json` keys and writes a
`settings.json.bak` backup first. You can also revert manually with
`mv ~/.claude/settings.json.bak ~/.claude/settings.json`.

## License

MIT
