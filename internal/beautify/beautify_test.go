package beautify

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustExpand(t *testing.T, in string) any {
	t.Helper()
	out, err := Expand([]byte(in), "  ")
	if err != nil {
		t.Fatalf("Expand failed: %v", err)
	}
	var v any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("expanded output is not valid JSON: %v\n%s", err, out)
	}
	return v
}

func TestExpandsStringifiedArrayAndObject(t *testing.T) {
	in := `{"data":{"getAllPromos":{"data":"[{\"id\":1}]","meta":"{\"total\":1}"}}}`
	v := mustExpand(t, in).(map[string]any)
	promos := v["data"].(map[string]any)["getAllPromos"].(map[string]any)
	data, ok := promos["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("data was not expanded to array: %#v", promos["data"])
	}
	meta, ok := promos["meta"].(map[string]any)
	if !ok || meta["total"] == nil {
		t.Fatalf("meta was not expanded to object: %#v", promos["meta"])
	}
}

func TestExpandsTopLevelStringifiedValue(t *testing.T) {
	in := `{"data":{"getAllPromoTypes":"[{\"id\":1}]"}}`
	v := mustExpand(t, in).(map[string]any)
	arr, ok := v["data"].(map[string]any)["getAllPromoTypes"].([]any)
	if !ok || len(arr) != 1 {
		t.Fatalf("value was not expanded: %#v", v)
	}
}

func TestLeavesPlainStringsAlone(t *testing.T) {
	in := `{"description":"BUY 1 AT CP","note":"[not json","brace":"{oops"}`
	out, err := Expand([]byte(in), "  ")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "BUY 1 AT CP") {
		t.Fatalf("plain string lost: %s", out)
	}
}

func TestInvalidJSONReturnsError(t *testing.T) {
	if _, err := Expand([]byte(`{oops`), "  "); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if _, err := Expand([]byte(`   `), "  "); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestDoubleNestedString(t *testing.T) {
	inner := `{"a":"[1,2]"}`
	outer, _ := json.Marshal(map[string]string{"outer": inner})
	v := mustExpand(t, string(outer)).(map[string]any)
	om := v["outer"].(map[string]any)
	if _, ok := om["a"].([]any); !ok {
		t.Fatalf("double nesting not expanded: %#v", v)
	}
}

func TestCustomIndent(t *testing.T) {
	out, err := Expand([]byte(`{"a":1}`), "    ")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "\n    \"a\"") {
		t.Fatalf("custom indent not applied: %s", out)
	}
}
