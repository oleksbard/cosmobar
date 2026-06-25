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
   - **Style:** `lean` (default), `tick`, or `blocks` (background pills). All styles are font-free — there is no Nerd Font / powerline option.
   - **Caps — ask ONLY when the chosen style is `blocks`:** `soft` (default) or `square`. For `lean` or `tick`, do **not** ask about caps at all. `AskUserQuestion` can't branch mid-batch, so put **style** in the first batch and send the **caps** question separately afterwards, only if style came back `blocks`.
   - **Animation:** `on` (default) or `off`. When on, a segment's value briefly scrambles through symbols and "decodes" into its new value whenever it changes — purely visual, width-stable. Offer `off` for users who want a completely static bar.
   - **Rate-limit window** (only if `rate_limits` is enabled): `both` (default), `5h`, or `7d`.
   - **Cost rollup** (only if `cost` is enabled): show today's cross-session spend (`· $5.30 today`) on the cost segment — `today` (default) to show, empty to hide.
   - **Block cost** (only if `rate_limits` is enabled): `on` (default) or `off`. When on, the 5-hour window shows its spend, e.g. `5h 31% $4.20 (2h30m left)`.

4. **Apply immediately — no preview step, no confirmation prompt.** As soon as the interview answers are in, write the config and wire it. Don't render a preview and don't put up a "shall I apply this?" question — both are just friction, and applying is reversible. This one command writes `~/.config/cosmobar/config.toml` AND wires `~/.claude/settings.json` (backing up the previous file):
   ```
   cosmobar init --force --theme <THEME> --order <comma,separated,enabled,segments> --clock <24h|12h|off> --glyphs <auto|ascii> --style <lean|tick|blocks> [--caps <soft|square>] --rate-window <both|5h|7d> [--animate <on|off>] [--cost-rollups today] [--block-cost <on|off>]
   ```
   - `--order` is the comma-separated list of the segments the user enabled, in their chosen order.
   - `--force` overwrites any existing config (this is a reconfigure).
   - `--caps` is only relevant when `--style blocks` is used — omit it otherwise.
   - `--rate-window` is only relevant when `rate_limits` is included in `--order`.
   - `--animate on|off` toggles the value-change scramble animation (default on).
   - `--cost-rollups` is only relevant when `cost` is in `--order` — pass `today` to show the today rollup (the default), or omit to keep it.
   - `--block-cost on|off` is only relevant when `rate_limits` is in `--order` — shows the 5h block's spend (default on).

5. **Confirm in one line.** Just tell the user it's applied — it appears at the bottom of Claude Code on their next message, they can re-run `/cosmobar` anytime, or hand-edit `~/.config/cosmobar/config.toml` and check it with `cosmobar preview`. Don't ask follow-up questions.

## Notes
- Everything is reversible: `cosmobar uninstall` removes the statusline wiring (add `--purge` to also delete the config and the binary).
- `cosmobar doctor` reports whether the binary, config, color support, and wiring are all healthy.
