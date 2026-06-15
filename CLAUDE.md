# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

cosmobar is a single static Go binary that renders the Claude Code status line. Claude Code invokes it once per refresh, piping a session JSON blob to stdin; cosmobar prints one-or-more terminal rows to stdout. No runtime deps, no `jq`, no Nerd Font — the only module dependency is `BurntSushi/toml`.

## Commands

```sh
make build                   # go build -o ./bin/cosmobar .
make test                    # go test ./...
make fmt                     # go fmt ./...
go test ./internal/render    # test one package
go test ./internal/segments -run TestCatalogMatchesRegistry  # one test
go vet ./...                 # must pass; CI runs it
go run . preview --cols 80   # fast visual loop without building
make dev                     # build + wire ./testsettings/.claude/settings.json for live testing
```

Never point the Claude Code `statusLine` command at `go run .` — it recompiles every refresh. Always use a built binary.

## Architecture

**Render pipeline (the spine).** `cmd_print.go:renderFromJSON` is the testable core: it parses stdin (`session.Parse`), loads TOML config (`config.Load`/`config.DefaultPath`), detects the terminal `render.Profile`, does one `git` call (cached per session ~1s in the OS temp dir via `git.Lookup`; everything else comes from the stdin JSON), and hands everything to `statusline.Render` as a fully-injected `statusline.Input` (Git, Cols, Profile, Now are all passed in, so the render is deterministic and unit-testable). `cosmobar preview` (`cmd_preview.go`) reuses this **exact same pipeline** with mocked session/git data — preview output is guaranteed to match the live line.

**Segments are a self-registering plugin registry.** Each segment lives in `internal/segments/<name>.go`, implements `Renderer` (`Name()` + `Render(ctx *Context) (Segment, bool)` where `ok=false` hides it), and calls `register(...)` in an `init()`. `statusline.Render` walks `config.Order`, looks each name up via `segments.Get`, and collects the shown segments. To add a segment: create the file, register it, add a matching entry to the `catalog` in `catalog.go`, and add a table test.

**The catalog is the source of truth for discovery.** `internal/segments/catalog.go` holds `[]Meta` (description, `DefaultOn`, `RequiresGit`, `ProMaxOnly`, `Role`). The `/cosmobar` guided-setup skill reads it dynamically (`cosmobar segments --json`, plus `cosmobar themes --json` for themes) — nothing is hardcoded in the UI — then applies the chosen config with a single `cosmobar init --force --theme … --order … --clock … --glyphs … --style … --caps … --rate-window …` call. `TestCatalogMatchesRegistry` fails the build if a registered segment lacks a catalog entry or vice versa, so the two must stay in sync.

**Segment formatting quirks live in the segment files.** A few non-obvious behaviors worth knowing before touching them: branch names are capped at `maxBranchWidth` (28 cols) with a middle ellipsis (`render.Truncate`); model names are compacted by `shortenModel` (e.g. `Opus 4.8 (1M context)` → `Opus 4.8(1M)`); the `lines` segment shows working-tree changes vs the last commit (`git diff HEAD --numstat`, untracked files excluded) and hides itself again after a commit, rendering its `+N`/`-N` as one flush two-tone pill in `blocks` style.

**Layout is width-aware fitting, not truncation.** `render.Fit(widths, prios, sepWidth, cols, maxRows)` packs segments into rows by terminal width, dropping the lowest-`Prio` segments first when space is tight. `render.Profile` (detected from env) decides color depth and glyph capability; `config.Glyphs`/`ASCII()` force unicode vs ascii. `internal/render/width.go` measures display width (handles wide runes / ellipsis) — use `render.Width`/`render.Truncate`, never `len()`, for on-screen sizing.

**Styles and themes are orthogonal.** `internal/theme` provides palettes (coral/catppuccin/nord/gruvbox); `internal/statusline/style.go` provides the layout flavor (lean/tick/blocks). A segment's `Role` (identity/vcs/gauge/metric/ambient) maps to a palette color; a `Segment` can emit multi-color `Part`s (e.g. `+12`/`-3` in `lines`) via `EffectiveParts`.

**Animation is a visual overlay, stateful across refreshes.** `internal/anim` persists per-session segment values to disk (`anim.Load`/`Save`). When a segment's signature changes between refreshes, `animateSegment` replaces its `Part` texts with a scramble frame (`anim.Frame`) for the transition — running for `config.Animation.DurationMs`, with the flavor picked from `Animation.Variants` (`glitch`/`decode`/`scatter`) — preserving each part's width and color. It only animates while Claude Code actively refreshes; otherwise it settles to the final value. `cosmobar preview --animate` drives it standalone.

**`config` vs `settings` are different files.** `internal/config` is cosmobar's own TOML (`~/.config/cosmobar/config.toml`, honoring `COSMOBAR_CONFIG`/`XDG_CONFIG_HOME`). `internal/settings` reads/writes Claude Code's `~/.claude/settings.json` to wire/unwire the `statusLine` block (always writing a `.bak` first). `cmd_init.go`/`cmd_uninstall.go` are the wire/unwire entry points.

## Commands map (`main.go:run`)

`main.go` dispatches the first arg to a `cmd*` function; default (no/`-`-prefixed arg) is `print`. Each `cmd_<x>.go` is a thin CLI wrapper around testable core logic (e.g. `renderFromJSON`, `installSkill`, `assetURL`/`extractBinary`). `cosmobar upgrade` self-updates from the latest GitHub Release: download tarball + `checksums.txt`, verify sha256, extract `cosmobar`, atomically replace the running binary.

## Releasing

Releases are tag-driven: `git tag vX.Y.Z && git push --tags` triggers `.github/workflows/release.yml`, which runs GoReleaser (`.goreleaser.yaml`) to cross-compile darwin/linux × amd64/arm64 and publish binaries + `checksums.txt`. The `internal/release` package parses release JSON / version comparison and is shared by the upgrade command.

## Testing conventions

Tests are table-driven and colocated (`*_test.go`). Because the render pipeline injects all external state, prefer testing core functions (`renderFromJSON`, `statusline.Render`, individual segment `Render`) with constructed inputs rather than shelling out. `internal/statusline/testdata/*.json` holds sample session blobs. `plugin_manifest_test.go` validates the `.claude-plugin` manifests.
