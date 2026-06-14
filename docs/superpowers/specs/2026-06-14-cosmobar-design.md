# cosmobar — Design Spec

**Date:** 2026-06-14
**Status:** Approved for planning
**One-liner:** A fast, dependency-light, starship-inspired statusline for Claude Code, written in Go.

---

## 1. Summary

cosmobar is a single statically-linked Go binary that Claude Code runs as its
`statusLine` command. It reads session JSON from stdin, reads terminal size from
the environment, makes at most one cached `git` refresh, and prints a styled
status line (one or two rows) to stdout.

It takes coralline's *purpose* (a rich Claude Code statusline: directory, git,
model, context gauge, cost, etc.) and applies starship's *philosophy*: one fast
artifact, trivial install, minimal dependencies, a clean TOML config, and a
release flow that lets the developer ship updates with a single git tag.

## 2. Goals

- **Easy to install** — prebuilt binary via a `curl | sh` script; no language
  runtime, no `jq`, no Nerd Font required.
- **Minimal dependencies** — Go stdlib for everything except one TOML parser.
- **Fast** — pure JSON parsing plus at most one cached `git` refresh per render;
  target well under ~10ms, no network in the render path.
- **Easy to maintain (developer ergonomics)** — instant local preview without
  Claude Code, a segment-registry pattern where a new segment is ~15 lines, and
  a one-command release (`git tag` → CI builds and publishes everything).
- **Easy for users to update** — `cosmobar upgrade` self-update, plus re-running
  the installer.

## 3. Non-goals (v1)

- Pill / powerline rendering and Nerd Font glyphs (lean style only in v1).
- Starship-level per-segment format strings and conditional expressions
  (curated config only; deeper customization is a possible future direction).
- Use as a general shell prompt (this is a Claude Code statusline).
- Homebrew tap (deferred; GitHub Releases + curl installer only).
- First-class Windows support (binary may run under Git Bash, but v1 targets
  macOS and Linux; Windows is best-effort / future).

## 4. Data contract (Claude Code → statusline)

Claude Code pipes session JSON on **stdin** and exposes `COLUMNS`/`LINES` as
**environment variables** (Claude Code ≥ 2.1.153). The render command should
always exit `0` and print output; a non-zero exit or empty output blanks the bar.

Fields cosmobar consumes (all others ignored):

| Segment / use | JSON field(s) | Notes |
|---|---|---|
| dir | `workspace.current_dir` (fallback `cwd`) | basename / short-path / full |
| model | `model.display_name` | |
| context gauge | `context_window.used_percentage` | may be `null` early / after `/compact` → segment hidden until first value |
| cost | `cost.total_cost_usd` | |
| duration | `cost.total_duration_ms` | |
| lines ± | `cost.total_lines_added`, `cost.total_lines_removed` | |
| rate limits | `rate_limits.five_hour.{used_percentage,resets_at}`, `rate_limits.seven_day.{…}` | **Pro/Max only**, may be absent → hide |
| output style | `output_style.name` | |
| effort | `effort.level` | absent on models without effort → hide |
| caching key | `session_id` | stable per session; used for git cache filename |
| git context | `workspace.repo.{host,owner,name}`, `workspace.git_worktree` | identity from JSON; working-tree state from `git` |

Git working-tree state (branch, ahead/behind, dirty counts, stash count) is **not**
in the JSON and is the only data requiring a subprocess.

Refresh model: runs after each assistant message, after `/compact`, on permission
mode change, and on vim toggle; debounced 300ms; in-flight runs are cancelled if a
newer update arrives. An optional `refreshInterval` (≥ 1s) re-runs on a timer for
time-based segments (the clock) and to refresh git while idle.

## 5. CLI surface

- **`cosmobar`** (alias `cosmobar print`) — default. Read stdin JSON, render, print.
  This is the command wired into `settings.json`.
- **`cosmobar init`** — idempotently wire `statusLine` into `~/.claude/settings.json`
  (backing up the existing file first) using the absolute path of the running
  binary, and create a default config if none exists. Re-running is safe.
- **`cosmobar preview [--cols N] [--theme NAME] [--config PATH]`** — render against
  bundled mock session JSON. No Claude Code, no stdin required. Primary dev loop.
- **`cosmobar doctor`** — offline diagnostics: is `statusLine` wired, is the binary
  on PATH, is `git` available, does the terminal support color, is the config valid.
  **No network calls.**
- **`cosmobar themes`** — list built-in themes.
- **`cosmobar upgrade [--check]`** — the *only* networked command. Query the latest
  GitHub Release, compare to the embedded build version, download the matching
  OS/arch asset, verify its SHA-256 against the release `checksums.txt`, and
  atomically replace the running binary (temp file + rename). `--check` reports
  current vs latest and exits without changing anything.
- **`cosmobar --version`** — print the embedded version.

## 6. Configuration

Location: `$XDG_CONFIG_HOME/cosmobar/config.toml` (default
`~/.config/cosmobar/config.toml`). Overridable via `--config PATH` or
`COSMOBAR_CONFIG`. Missing config → built-in defaults (no file required to run).

```toml
theme            = "coral"            # coral | catppuccin | nord | gruvbox
order            = ["dir","git","model","context","cost","clock"]
separator        = " · "
max_rows         = 2                  # responsive wrap budget (1 = single line, truncate only)
gauge_width      = 8                  # cells for context / rate-limit bars
gauge_thresholds = [70, 90]           # green → yellow → red
glyphs           = "auto"             # auto | unicode | ascii  (bar/sep characters)

[clock]        format = "24h"         # 24h | 12h | off
[dir]          style  = "basename"    # basename | short-path | full
[context]      show   = true
[rate_limits]  show   = false         # off by default (Pro/Max only)
# Any built-in segment may be toggled and reordered via `order`; per-segment
# tables are optional and fall back to sensible defaults when omitted.
```

Default-on layout (when `order` is omitted): `dir`, `git`, `model`, `context`,
`cost`, `clock`. The remaining segments are opt-in by adding them to `order`.

Colors honor `NO_COLOR`; truecolor vs 256-color is detected via `COLORTERM`.
Individual theme colors can be overridden in config (future-friendly, kept simple
in v1).

## 7. Segments (built-in library, ~11)

Each is independently toggleable. **Default-on** marked ★.

| Segment | Source | Renders |
|---|---|---|
| `dir` ★ | JSON `workspace.current_dir` | basename / short / full path |
| `git` ★ | `git` (cached) | branch, ahead/behind, dirty counts (staged/modified/untracked) |
| `model` ★ | JSON `model.display_name` | model name |
| `context` ★ | JSON `context_window.used_percentage` | color gauge + `%` |
| `cost` ★ | JSON `cost.total_cost_usd` | `$0.12` |
| `clock` ★ | system time | `14:32` (24h/12h/off) |
| `rate_limits` | JSON `rate_limits.*` | `5h 23% · 7d 41%` (+ reset countdown); hidden if absent |
| `duration` | JSON `cost.total_duration_ms` | `12m 03s` |
| `lines` | JSON `cost.total_lines_{added,removed}` | `+156 -23` |
| `output_style` | JSON `output_style.name` | style name |
| `git_stash` | `git` (cached) | stash count |
| `effort` | JSON `effort.level` | reasoning effort; hidden if absent |

**Segment interface (registry pattern):** each segment lives in its own file and
implements roughly:

```go
type Segment struct {
    Text  string      // already-formatted content, no color
    State GaugeState  // ok | warn | crit (for threshold coloring), or none
    Prio  int         // higher = kept longer when width is tight
}

type Renderer interface {
    Name() string
    Render(ctx *RenderContext) (Segment, bool) // ok=false → hidden this render
}
```

A central registry maps names → `Renderer`. Adding a segment later = one new file
+ one registry entry (documented in README for future maintenance).

## 8. Architecture & data flow

```
stdin JSON ─▶ session.Parse ─┐
env COLUMNS/LINES ───────────┤
git (cached refresh, /tmp ───┼─▶ RenderContext ─▶ ordered, enabled segments
keyed by session_id, ~1s TTL)┘                       │ each → Segment{Text,State,Prio}
                                                      ▼
                          render: join with separator · apply theme colors ·
                          build gauges · width-fit (wrap rows / truncate dir) ─▶ stdout
```

**Packages (`internal/`):**

- `session` — JSON schema structs + tolerant stdin parsing (absent/null safe).
- `git` — cached working-tree state. On stale cache (file older than ~1s or
  missing), gathers branch + ahead/behind + dirty counts via
  `git status --porcelain=v2 --branch` and stash count via `git stash list`,
  writes to `/tmp/cosmobar-git-<session_id>`. Not a repo / git error → empty
  result (segments self-hide). Two git calls max per refresh, both cached.
- `segments` — one file per segment + the registry.
- `render` — lean styling, separators, Unicode/ASCII gauges, ANSI color
  application, color-profile detection, and the responsive width-fitter.
- `theme` — named palettes (accent + ok/warn/crit state colors); ~4 built in.
- `config` — TOML load, defaults, validation.
- `settings` — read/write `~/.claude/settings.json` for `init`.

`main.go` wires subcommand dispatch; version is injected at build time via
`-ldflags -X main.version=…`.

## 9. Rendering, layout & theming (lean)

- **Lean style:** flat colored text joined by `separator`. No powerline glyphs,
  no Nerd Font.
- **Gauges:** Unicode block run `▓░` (universally supported, not Nerd-Font-specific),
  color-shifting ok→warn→crit at `gauge_thresholds`. `glyphs = "ascii"` swaps to
  `#`/`-` for the most minimal terminals.
- **Responsive width-fit:** if the joined line fits `COLUMNS`, print one row. If
  not, wrap at segment boundaries into at most `max_rows` rows. If a row still
  overflows, shrink lowest-priority content first and middle-ellipsis the `dir`
  segment. `max_rows = 1` means single line with truncation only.
- **Theming:** a theme is a named palette. Ship `coral` (default), `catppuccin`,
  `nord`, `gruvbox`. Users may override individual colors in config.

## 10. Dev loop & updates

**Local development (no install, no Claude Code):**

```bash
go run . preview --cols 80 --theme nord --config ./my.toml   # bundled mock JSON
cat testdata/heavy.json | go run . print                     # real stdin path
go test ./...                                                # golden tests
```

**Live inside Claude Code without a global install** — build to a path and point a
*project-scoped* settings file at it (never touches `~/.claude`):

```bash
go build -o ./bin/cosmobar .
```
```jsonc
// <test-repo>/.claude/settings.json
{ "statusLine": { "type": "command", "command": "/abs/path/bin/cosmobar" } }
```

Rebuild → the next message in that session picks up the new binary. A `make dev`
target builds and wires a throwaway test settings file in one step.

> Caveat (documented in README): never set the command to `go run .` — Go
> recompiles on each invocation (hundreds of ms) and the statusline runs on every
> message. Always point at a *built* binary.

**How users receive updates** (both backed by the same GitHub Releases):

1. **Re-run the installer:** `curl -sS .../install.sh | sh` always fetches the
   latest release. The install script is the updater.
2. **`cosmobar upgrade`:** self-update in one command (download latest matching
   asset, verify checksum, atomic swap). `cosmobar upgrade --check` reports
   current vs latest only.

No background update checks ever run in the render path (keeps it network-free
and fast); version checks are explicit and user-initiated.

## 11. Error handling (never break the bar)

- Always exit `0` and print something.
- Missing/`null` JSON fields → the affected segment self-hides (e.g. no
  `rate_limits` for non-Pro/Max users; `context` null after `/compact`).
- Not a git repo or git error → git/stash segments self-hide; never blocks.
- Config parse error → fall back to built-in defaults and write a one-line note to
  stderr (visible only under `claude --debug`); the bar still renders.
- Caching keeps git cost low even with `refreshInterval` enabled.

## 12. Testing

- **Golden-file tests:** `testdata/*.json` fixtures (`minimal`, `heavy`,
  `pro-max` with rate limits, `no-git`, `compacted` with null context) → expected
  exact stdout including ANSI. The render layer is pure (no I/O) given a
  `RenderContext`, so it is fully golden-testable.
- **Per-segment table tests:** each segment's formatting and hide/show logic.
- **Width/responsive tests:** render `heavy.json` at `COLUMNS` 120 / 80 / 40 and
  assert row count and `dir` truncation behavior.
- **git module:** tested against a temp repo fixture; cache TTL behavior unit-tested.
- `cosmobar preview` doubles as the manual visual check.

## 13. Repo layout

```
cosmobar/
  main.go
  internal/
    session/   git/   segments/   render/   theme/   config/   settings/
  testdata/                 # mock session JSON + golden outputs
  install.sh                # curl | sh installer (detect OS/arch, drop in ~/.local/bin)
  .goreleaser.yaml          # cross-compile matrix + release artifacts + checksums
  .github/workflows/release.yml   # tag → GoReleaser
  Makefile                  # dev, build, test
  README.md
  docs/superpowers/specs/2026-06-14-cosmobar-design.md
```

## 14. Build & release

- `git tag vX.Y.Z && git push --tags` → GitHub Actions runs GoReleaser.
- GoReleaser cross-compiles darwin/linux × amd64/arm64, sets the version via
  ldflags, and publishes a GitHub Release with binaries + `checksums.txt`.
- `install.sh` detects OS/arch, downloads the matching asset from the latest
  release, verifies the checksum, and installs to `~/.local/bin/cosmobar`.
- After install: `cosmobar init` wires it into `~/.claude/settings.json`.

`settings.json` written by `init`:

```jsonc
{
  "statusLine": {
    "type": "command",
    "command": "/Users/you/.local/bin/cosmobar",
    "padding": 0,
    "refreshInterval": 10
  }
}
```

(`refreshInterval` defaults to 10s to keep the clock and idle git state fresh;
configurable, and only meaningful when a time-based segment is enabled.)

## 15. Implementation milestones (for the plan)

1. **Skeleton & test harness** — `main` dispatch, `session` parse, `config` load
   with defaults, `print` with a few segments, golden-test scaffolding.
2. **Segment library + git module** — all ~11 segments and the cached `git` package.
3. **Render & layout** — lean styling, themes, gauges, responsive width-fitting.
4. **CLI commands** — `init`, `preview`, `doctor`, `themes`, `upgrade`.
5. **Release pipeline** — `.goreleaser.yaml`, `install.sh`, CI workflow, README.

## 16. Defaults chosen (previously open)

- **Config location:** `~/.config/cosmobar/config.toml` (XDG, starship convention).
- **`max_rows`:** `2`.
- **Default-on segments:** `dir`, `git`, `model`, `context`, `cost`, `clock`.

## 17. Possible future work (out of scope for v1)

- Pill/powerline style + Nerd Font glyphs as an opt-in.
- Deeper, starship-style per-segment format strings.
- Homebrew tap.
- Native Windows support.
- Subagent statusline (`subagentStatusLine`) rendering.
