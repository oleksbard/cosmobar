# cosmobar

A fast, dependency-light, starship-inspired status line for [Claude Code](https://code.claude.com).

- Single static Go binary — no runtime, no `jq`, no Nerd Font required
- Lean, themed output with a context gauge, git status, cost, model, and more
- One cached `git` call per refresh; everything else is parsed from stdin JSON
- TOML config; instant local preview; one-command self-update

## Install

```sh
curl -sS https://raw.githubusercontent.com/oleksbard/cosmobar/main/install.sh | sh
cosmobar init
```

`init` wires `cosmobar` into `~/.claude/settings.json`, writes a default
config to `~/.config/cosmobar/config.toml`, and installs the guided-setup
skill. Restart Claude Code (or send a message) to see the status line.

## Guided setup inside Claude Code

`init` (or `cosmobar install-skill`) drops a `/cosmobar` skill into
`~/.claude/skills/`. In Claude Code, run **`/cosmobar`** (or just ask
"set up cosmobar") and Claude will:

1. discover the available segments and themes **dynamically** (`cosmobar
   segments --json`, `cosmobar themes --json`) — new segments appear
   automatically, nothing is hardcoded;
2. ask you which segments to show, plus theme, clock, and glyph style;
3. apply everything with one command:
   `cosmobar init --force --theme <t> --order <a,b,c> --clock <fmt> --glyphs <g>`;
4. show you the result with `cosmobar preview`.

You can re-run it anytime to reconfigure.

## Configure

Edit `~/.config/cosmobar/config.toml`:

```toml
theme            = "coral"            # coral | catppuccin | nord | gruvbox
order            = ["dir", "git", "model", "context", "cost", "clock"]
separator        = " · "
max_rows         = 2
gauge_width      = 8
gauge_thresholds = [70, 90]
glyphs           = "auto"             # auto | unicode | ascii

[clock]
format = "24h"

[dir]
style = "basename"                    # basename | short-path | full

[context]
show = true

[rate_limits]
show = false                          # Pro/Max only
```

Available segments: `dir`, `git`, `model`, `context`, `cost`, `clock`,
`rate_limits`, `duration`, `lines`, `output_style`, `git_stash`, `effort`.
Add or reorder them in `order`.

Preview changes without launching Claude Code:

```sh
cosmobar preview --cols 80 --theme nord
```

## Commands

| Command | What it does |
|---|---|
| `cosmobar` | Render the status line (reads JSON from stdin). |
| `cosmobar init` | Wire into `settings.json`, write config, install the setup skill. Flags: `--theme --order --clock --glyphs --force --no-skill`. |
| `cosmobar install-skill` | Install the `/cosmobar` guided-setup skill into `~/.claude/skills/`. |
| `cosmobar segments [--json]` | List all available segments (the dynamic catalog). |
| `cosmobar uninstall [--purge]` | Remove the `statusLine` block from `settings.json` (inverse of `init`). `--purge` also deletes the config file and the binary. |
| `cosmobar preview` | Render the bundled mock session (`--cols`, `--theme`, `--config`). |
| `cosmobar doctor` | Offline diagnostics. |
| `cosmobar themes` | List built-in themes. |
| `cosmobar upgrade [--check]` | Self-update from the latest GitHub Release. |

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

`uninstall` preserves your other `settings.json` keys and writes a `settings.json.bak`
backup first. You can also revert manually with
`mv ~/.claude/settings.json.bak ~/.claude/settings.json`.

## Development

```sh
make test                   # go test ./...
cosmobar preview --cols 80  # fast visual loop (use `go run . preview` too)
make dev                    # build + wire ./testsettings/.claude/settings.json
```

> Don't set the Claude Code `statusLine` command to `go run .` — it recompiles
> on every invocation. Always point at a built binary (`make build`).

### Adding a segment

1. Create `internal/segments/<name>.go` implementing the `Renderer` interface
   (`Name()` + `Render(ctx) (Segment, bool)`), and call `register(...)` in `init()`.
2. Add a table test in `internal/segments/`.
3. Add the name to `order` in your config to enable it.

## Releasing

```sh
git tag vX.Y.Z && git push --tags
```

GitHub Actions runs GoReleaser, which cross-compiles darwin/linux × amd64/arm64
and publishes a Release with binaries + `checksums.txt`.

## License

MIT
