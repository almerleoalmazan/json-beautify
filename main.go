// Command json-beautify expands AppSync-style JSON (stringified values
// decoded recursively) and shows it in a scrollable terminal viewer with
// syntax highlighting, line numbers, and one-key copy to clipboard.
//
// Usage:
//
//	pbpaste | json-beautify
//	json-beautify --paste        # read input from system clipboard (macOS/Linux/Windows)
//	json-beautify response.json
//	cat response.json | json-beautify --no-tui
//	json-beautify --copy response.json   # copy result, skip TUI
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/atotto/clipboard"
	"json-beautify/internal/beautify"
	"json-beautify/internal/highlight"
	"json-beautify/internal/tui"
)

type config struct {
	noTUI       bool
	copyIt      bool
	paste       bool
	indent      string
	theme       string
	color       string
	lineNumbers bool
}

func main() {
	var cfg config
	flag.BoolVar(&cfg.noTUI, "no-tui", false, "print beautified JSON to stdout instead of opening the viewer")
	flag.BoolVar(&cfg.copyIt, "copy", false, "copy beautified JSON to clipboard and exit (output still prints)")
	flag.BoolVar(&cfg.paste, "paste", false, "read input from system clipboard instead of stdin/file")
	flag.StringVar(&cfg.indent, "indent", "  ", "indentation string (e.g. \"  \" or \"    \")")
	flag.StringVar(&cfg.theme, "theme", highlight.DefaultStyle, "chroma style for highlighting (e.g. dracula, monokai, swapoff)")
	flag.StringVar(&cfg.color, "color", "auto", "colored output: auto, always, or never")
	flag.BoolVar(&cfg.lineNumbers, "line-numbers", true, "show line numbers in the viewer (toggle with l)")
	flag.BoolVar(&cfg.paste, "p", false, "shorthand for --paste")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  pbpaste | json-beautify\n  json-beautify --paste\n  json-beautify [flags] [file.json]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(cfg, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "json-beautify:", err)
		os.Exit(1)
	}
}

func run(cfg config, args []string) error {
	input, err := readInput(cfg, args)
	if err != nil {
		return err
	}
	pretty, err := beautify.Expand(input, cfg.indent)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.copyIt {
		if err := clipboard.WriteAll(string(pretty)); err != nil {
			return fmt.Errorf("copy to clipboard: %w", err)
		}
		fmt.Println(string(pretty))
		fmt.Fprintln(os.Stderr, "✓ copied to clipboard")
		return nil
	}

	if cfg.noTUI || !isTTY() {
		fmt.Println(coloredOutput(string(pretty), cfg))
		return nil
	}

	return tui.Run(string(pretty), tui.Options{
		Theme:       cfg.theme,
		Color:       highlight.Enabled(cfg.color),
		LineNumbers: cfg.lineNumbers,
	})
}

// coloredOutput applies highlighting for --no-tui pipe output only when
// explicitly wanted: --color always, or auto with a TTY and no NO_COLOR.
func coloredOutput(pretty string, cfg config) string {
	mode := cfg.color
	if mode == "auto" && !isTTY() {
		mode = "never"
	}
	if !highlight.Enabled(mode) {
		return pretty
	}
	return highlight.HighlightJSON(pretty, cfg.theme)
}

func readInput(cfg config, args []string) ([]byte, error) {
	if cfg.paste {
		if len(args) > 0 {
			return nil, fmt.Errorf("cannot combine --paste with a file argument")
		}
		text, err := clipboard.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("read clipboard: %w (Linux needs xclip or xsel installed)", err)
		}
		if len(text) == 0 {
			return nil, fmt.Errorf("clipboard is empty: copy JSON first, then run with --paste")
		}
		return []byte(text), nil
	}
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	st, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if (st.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no input: pipe JSON via stdin or pass a file argument")
	}
	return io.ReadAll(os.Stdin)
}

func isTTY() bool {
	st, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (st.Mode() & os.ModeCharDevice) != 0
}
