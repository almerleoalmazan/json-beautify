package tui

import (
	"testing"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

func TestGutterWidthConsistent(t *testing.T) {
	g := lineNumberGutter()
	widths := map[int]bool{}
	cases := []viewport.GutterContext{
		{Index: 0, TotalLines: 100},
		{Index: 99, TotalLines: 100},
		{Index: 100, TotalLines: 100}, // past EOF
		{Index: 5, TotalLines: 100, Soft: true},
	}
	for _, c := range cases {
		widths[lipgloss.Width(g(c))] = true
	}
	if len(widths) != 1 {
		t.Fatalf("gutter widths vary: %v", widths)
	}
}

func TestNewDefaults(t *testing.T) {
	m := New(`{"a":1}`, Options{})
	if m.raw != `{"a":1}` {
		t.Fatalf("raw not stored: %q", m.raw)
	}
	if m.display == "" {
		t.Fatal("display empty")
	}
}
