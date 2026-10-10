# json-beautify

Beautifies AppSync-style JSON where values are themselves JSON-encoded strings,
then shows the result in a scrollable terminal viewer with one-key copy to clipboard.

Example input (note the escaped strings):

```json
{"data": {"getAllPromos": {"data": "[{\"id\":1}]", "meta": "{\"total\":1}"}}}
```

becomes real nested JSON:

```json
{
  "data": {
    "getAllPromos": {
      "data": [{ "id": 1 }],
      "meta": { "total": 1 }
    }
  }
}
```

## TUI stack ([charm.land](https://charm.land/))

- **Bubble Tea v2** — Elm-style TUI runtime (alt-screen + mouse set per-view)
- **Bubbles/viewport v2** — scrollable pager (arrows, pgup/pgdn, horizontal scroll, mouse wheel) with native `LeftGutterFunc` line numbers
- **Lip Gloss v2** — header/footer/gutter styling
- **chroma v2** (non-Charm) — JSON syntax highlighting; Charm has no code tokenizer (Glamour is markdown-only). Viewport renders chroma's ANSI output.
- **atotto/clipboard** — cross-platform clipboard (Charm has no clipboard lib)

## Usage

```sh
pbpaste | json-beautify
json-beautify response.json
cat response.json | json-beautify --no-tui
json-beautify --copy response.json   # print + copy, skip viewer
json-beautify --indent "    " response.json
json-beautify --theme monokai --color always --no-tui response.json
json-beautify --line-numbers=false response.json
```

Keys in viewer: `↑/↓` scroll, `←/→` horizontal, `pgup/pgdn`, `l` line numbers, `s` highlighting, `y`/`c` copy raw JSON, `q`/`esc` quit.
Copy always puts plain JSON (no ANSI colors, no line numbers) on the clipboard.

## Paste from clipboard

Easiest — same command on every OS (reads the system clipboard directly):

```sh
json-beautify --paste
```

Native pipe equivalents, alongside macOS `pbpaste`:

```sh
# macOS
pbpaste | json-beautify

# Linux (X11) — needs xclip or xsel installed
xclip -o -selection clipboard | json-beautify
xsel --clipboard --output | json-beautify

# Linux (Wayland) — needs the wl-clipboard package
wl-paste | json-beautify

# Windows (PowerShell)
Get-Clipboard | json-beautify
# from cmd.exe:
powershell -Command Get-Clipboard | json-beautify
```

Notes:

- `--paste`, `--copy`, and the viewer's `y` key all use the system clipboard on macOS, Linux, and Windows.
- On Linux, clipboard access needs `xclip` or `xsel` installed (`sudo apt install xclip`, `sudo dnf install xclip`, etc.).
- On Wayland without XWayland, `--paste` may not see the clipboard — use the `wl-paste |` pipe form instead.

## Install

Requires Go 1.27 or newer.

```sh
go build -o json-beautify .
```

## Layout

- `main.go` — stdin/file input, flags, TUI-or-print dispatch
- `internal/beautify/` — recursive string-JSON expansion + pretty print
- `internal/highlight/` — chroma JSON highlighting + color-mode helper
- `internal/tui/` — Bubble Tea viewport model, gutter, toggles, clipboard
- `testdata/` — sample AppSync payloads
