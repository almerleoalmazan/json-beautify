// Package highlight adds chroma-based JSON syntax highlighting for
// terminal display. There is no Charm library for code tokenizing
// (Glamour is markdown-only, Lip Gloss is styling-only), so chroma is
// the standard choice; its ANSI output is rendered by the Charm
// viewport unchanged.
package highlight

import (
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// DefaultStyle is the chroma style used when none is requested.
const DefaultStyle = "dracula"

// HighlightJSON tokenizes pretty-printed JSON and returns it with ANSI
// colors for the given chroma style. Unknown styles fall back to the
// chroma fallback style; tokenize/format failures return input unchanged.
func HighlightJSON(src, styleName string) string {
	if styleName == "" {
		styleName = DefaultStyle
	}
	lexer := lexers.Get("JSON")
	if lexer == nil {
		return src
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return src
	}
	it, err := lexer.Tokenise(nil, src)
	if err != nil {
		return src
	}
	var sb strings.Builder
	if err := formatter.Format(&sb, style, it); err != nil {
		return src
	}
	return sb.String()
}

// Enabled reports whether colored output should be used: never when
// NO_COLOR is set, otherwise per mode ("always", "never", "auto").
func Enabled(mode string) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	switch strings.ToLower(mode) {
	case "always":
		return true
	case "never":
		return false
	default: // "auto"
		return true
	}
}

// Strip removes ANSI escape sequences, used in tests to verify highlighting
// preserves the underlying text.
func Strip(s string) string {
	var sb strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inEsc {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEsc = false
			}
			continue
		}
		if c == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			inEsc = true
			i++
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}
