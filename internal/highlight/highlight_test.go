package highlight

import (
	"strings"
	"testing"
)

func TestHighlightPreservesText(t *testing.T) {
	src := "{\n  \"id\": 1,\n  \"name\": \"x\",\n  \"ok\": true,\n  \"n\": null\n}"
	out := HighlightJSON(src, "dracula")
	if !strings.Contains(out, "\x1b[") {
		t.Fatal("expected ANSI codes in highlighted output")
	}
	if Strip(out) != src {
		t.Fatalf("highlight changed text:\n%q\nvs\n%q", Strip(out), src)
	}
}

func TestUnknownStyleFallsBack(t *testing.T) {
	src := `{"a": 1}`
	out := HighlightJSON(src, "does-not-exist")
	if Strip(out) != src {
		t.Fatalf("fallback changed text: %q", out)
	}
}

func TestEnabledRespectsMode(t *testing.T) {
	if !Enabled("always") {
		t.Fatal("always should enable")
	}
	if Enabled("never") {
		t.Fatal("never should disable")
	}
	t.Setenv("NO_COLOR", "1")
	if Enabled("always") {
		t.Fatal("NO_COLOR should disable even when always")
	}
}
