// Package beautify expands AppSync-style JSON where values are
// themselves JSON-encoded strings (e.g. `"data": "[{\"id\":1}]"`),
// then pretty-prints the result.
package beautify

import (
	"bytes"
	"encoding/json"
	"strings"
)

const maxDepth = 20

// Expand parses input as JSON, recursively decodes any string values that
// themselves contain JSON, and returns indented JSON using the given indent.
// An empty indent defaults to two spaces.
func Expand(input []byte, indent string) ([]byte, error) {
	if indent == "" {
		indent = "  "
	}
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return nil, errEmpty()
	}
	var v any
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	v = normalize(v, 0)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", indent)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func normalize(v any, depth int) any {
	if depth > maxDepth {
		return v
	}
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			t[k] = normalize(val, depth+1)
		}
		return t
	case []any:
		for i, val := range t {
			t[i] = normalize(val, depth+1)
		}
		return t
	case string:
		s := strings.TrimSpace(t)
		if len(s) < 2 {
			return v
		}
		c := s[0]
		last := s[len(s)-1]
		// Quick reject: must look like an object or array.
		if (c == '{' && last != '}') || (c == '[' && last != ']') {
			return v
		}
		if c != '{' && c != '[' {
			return v
		}
		var inner any
		dec := json.NewDecoder(strings.NewReader(s))
		dec.UseNumber()
		if err := dec.Decode(&inner); err != nil {
			return v
		}
		// Reject trailing garbage beyond whitespace.
		if rest := strings.TrimSpace(s[dec.InputOffset():]); rest != "" {
			return v
		}
		return normalize(inner, depth+1)
	default:
		return v
	}
}

type emptyError struct{}

func (emptyError) Error() string { return "empty input: expected JSON" }

func errEmpty() error { return emptyError{} }
