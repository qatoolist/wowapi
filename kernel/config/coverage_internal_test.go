package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// ---------- convertValue / toInt64 (the audited scalar converter) ----------

func convErr(t *testing.T, ptr any, raw any) error {
	t.Helper()
	return convertValue(reflect.ValueOf(ptr).Elem(), raw)
}

func mustConv(t *testing.T, ptr any, raw any) {
	t.Helper()
	if err := convErr(t, ptr, raw); err != nil {
		t.Fatalf("convertValue(%T, %v) unexpected error: %v", ptr, raw, err)
	}
}

func wantConvErr(t *testing.T, ptr any, raw any, sub string) {
	t.Helper()
	err := convErr(t, ptr, raw)
	if err == nil {
		t.Fatalf("convertValue(%T, %v): expected error, got nil", ptr, raw)
	}
	if sub != "" && !strings.Contains(err.Error(), sub) {
		t.Errorf("convertValue(%T, %v) error %q missing %q", ptr, raw, err, sub)
	}
}

// upperText exercises the encoding.TextUnmarshaler fast path in convertValue.
type upperText struct{ s string }

func (u *upperText) UnmarshalText(b []byte) error {
	if string(b) == "boom" {
		return fmt.Errorf("upperText: rejected")
	}
	u.s = strings.ToUpper(string(b))
	return nil
}

func TestConvertValueScalars(t *testing.T) {
	// string: string raw and non-string coercion via fmt.Sprint.
	var s string
	mustConv(t, &s, "hello")
	if s != "hello" {
		t.Errorf("string = %q", s)
	}
	mustConv(t, &s, 42)
	if s != "42" {
		t.Errorf("non-string coercion = %q", s)
	}

	// bool: from bool, from valid string, invalid string, wrong type.
	var b bool
	mustConv(t, &b, true)
	if !b {
		t.Error("bool from bool failed")
	}
	mustConv(t, &b, "false")
	if b {
		t.Error("bool from string failed")
	}
	wantConvErr(t, &b, "notabool", "valid bool")
	wantConvErr(t, &b, 3, "as bool")

	// int: valid, string, string-invalid, overflow.
	var i int
	mustConv(t, &i, 7)
	if i != 7 {
		t.Errorf("int = %d", i)
	}
	mustConv(t, &i, "13")
	if i != 13 {
		t.Errorf("int from string = %d", i)
	}
	wantConvErr(t, &i, "xyz", "valid integer")
	var i8 int8
	wantConvErr(t, &i8, 9999, "overflows")

	// uint: valid, negative, overflow.
	var u uint
	mustConv(t, &u, 5)
	if u != 5 {
		t.Errorf("uint = %d", u)
	}
	wantConvErr(t, &u, -1, "out of range")
	var u8 uint8
	wantConvErr(t, &u8, 9999, "out of range")

	// float: from float64, int, int64, string, string-invalid, wrong type.
	var f float64
	mustConv(t, &f, 2.5)
	if f != 2.5 {
		t.Errorf("float from float64 = %v", f)
	}
	mustConv(t, &f, 3)
	if f != 3 {
		t.Errorf("float from int = %v", f)
	}
	mustConv(t, &f, int64(4))
	if f != 4 {
		t.Errorf("float from int64 = %v", f)
	}
	mustConv(t, &f, "1.25")
	if f != 1.25 {
		t.Errorf("float from string = %v", f)
	}
	wantConvErr(t, &f, "notnum", "valid number")
	wantConvErr(t, &f, true, "as number")
}

func TestConvertValueDurationSliceAndMisc(t *testing.T) {
	// duration: valid string, non-string raw.
	var d time.Duration
	mustConv(t, &d, "5s")
	if d != 5*time.Second {
		t.Errorf("duration = %v", d)
	}
	wantConvErr(t, &d, 5, "duration string")
	wantConvErr(t, &d, "nope", "valid duration")

	// slice: from []any, from CSV string (trimmed), from empty string, wrong
	// type, and per-element conversion error.
	var sl []string
	mustConv(t, &sl, []any{"x", "y"})
	if len(sl) != 2 || sl[0] != "x" {
		t.Errorf("slice from []any = %v", sl)
	}
	mustConv(t, &sl, "a, b ,c")
	if len(sl) != 3 || sl[1] != "b" {
		t.Errorf("slice from CSV = %v", sl)
	}
	var empty []string
	mustConv(t, &empty, "")
	if len(empty) != 0 {
		t.Errorf("empty CSV should yield empty slice, got %v", empty)
	}
	var badList []string
	wantConvErr(t, &badList, 5, "expected a list")
	var ints []int
	wantConvErr(t, &ints, []any{"notanint"}, "[0]")

	// mapping raw into a scalar field is rejected up front.
	var scalar string
	wantConvErr(t, &scalar, map[string]any{"k": 1}, "scalar")

	// unsupported kind (a map field) surfaces a clear error.
	var m map[string]int
	wantConvErr(t, &m, "x", "unsupported")

	// TextUnmarshaler fast path: success and error.
	var u upperText
	mustConv(t, &u, "hello")
	if u.s != "HELLO" {
		t.Errorf("TextUnmarshaler = %q", u.s)
	}
	wantConvErr(t, &u, "boom", "rejected")
}

func TestToInt64(t *testing.T) {
	ok := func(raw any, want int64) {
		t.Helper()
		got, err := toInt64(raw)
		if err != nil {
			t.Fatalf("toInt64(%v): %v", raw, err)
		}
		if got != want {
			t.Errorf("toInt64(%v) = %d, want %d", raw, got, want)
		}
	}
	ok(7, 7)
	ok(int64(9), 9)
	ok(uint64(11), 11)
	ok(6.0, 6)

	bad := func(raw any, sub string) {
		t.Helper()
		_, err := toInt64(raw)
		if err == nil {
			t.Fatalf("toInt64(%v): expected error", raw)
		}
		if !strings.Contains(err.Error(), sub) {
			t.Errorf("toInt64(%v) error %q missing %q", raw, err, sub)
		}
	}
	bad(uint64(1)<<63, "overflows int64")
	bad(6.5, "whole number")
	bad("notanint", "valid integer")
	bad(true, "cannot use")
}

// ---------- tree helpers ----------

func TestDeepCopyTreeNoAliasing(t *testing.T) {
	in := map[string]any{
		"nested": map[string]any{"k": 1},
		"list":   []any{"a", map[string]any{"x": 2}},
		"scalar": "v",
	}
	out := deepCopyTree(in)

	// Mutate the copy deeply; the original must be untouched.
	out["nested"].(map[string]any)["k"] = 99
	out["list"].([]any)[0] = "changed"
	out["list"].([]any)[1].(map[string]any)["x"] = 77

	if in["nested"].(map[string]any)["k"] != 1 {
		t.Error("nested map aliased into copy")
	}
	if in["list"].([]any)[0] != "a" {
		t.Error("slice element aliased into copy")
	}
	if in["list"].([]any)[1].(map[string]any)["x"] != 2 {
		t.Error("map inside slice aliased into copy")
	}
	if out["scalar"] != "v" {
		t.Error("scalar not copied")
	}
}

func TestApplyEnvironSkipsPrefixOnlyAndUnrelated(t *testing.T) {
	dst := map[string]any{}
	prov := Provenance{}
	applyEnviron(dst, "P__", []string{
		"P__=only-prefix",   // rest == "" → skipped
		"P__SECTION__A=val", // maps to section.a
		"UNRELATED=z",       // no prefix → skipped
		"malformed-no-eq",   // no '=' → skipped
	}, prov)

	sec, ok := dst["section"].(map[string]any)
	if !ok || sec["a"] != "val" {
		t.Fatalf("expected section.a=val, got %v", dst)
	}
	if _, ok := dst[""]; ok {
		t.Error("prefix-only env var should not create an empty key")
	}
	if len(dst) != 1 {
		t.Errorf("only one key expected, got %v", dst)
	}
	if prov["section.a"] != LayerEnvVar {
		t.Errorf("provenance = %q", prov["section.a"])
	}
}

func TestParseYAMLFileEmptyAndMissing(t *testing.T) {
	// Empty file yields an empty (non-nil) map, not nil.
	p := filepath.Join(t.TempDir(), "empty.yaml")
	if err := os.WriteFile(p, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := parseYAMLFile(p)
	if err != nil {
		t.Fatalf("empty file: %v", err)
	}
	if m == nil || len(m) != 0 {
		t.Errorf("empty file should yield empty map, got %v", m)
	}

	// Missing file is a read error carrying the path.
	if _, err := parseYAMLFile(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("missing file should error")
	}
}

// ---------- shortHex ----------

func TestShortHex(t *testing.T) {
	if got := shortHex("abc"); got != "abc" {
		t.Errorf("short input passthrough = %q", got)
	}
	if got := shortHex("0123456789abcdef"); got != "0123456789ab" {
		t.Errorf("long input truncation = %q", got)
	}
}

// ---------- Secret.IsZero / GoString ----------

func TestSecretIsZeroAndGoString(t *testing.T) {
	if !(Secret{}).IsZero() {
		t.Error("zero Secret must be IsZero")
	}
	if NewSecret("secretref://env/X", "v").IsZero() {
		t.Error("resolved Secret must not be IsZero")
	}
	// A ref with no value is still non-zero (it carries a reference).
	if NewSecret("secretref://env/X", "").IsZero() {
		t.Error("ref-only Secret must not be IsZero")
	}

	// GoString must redact and never carry the value.
	gs := NewSecret("secretref://env/X", "topsecret").GoString()
	if strings.Contains(gs, "topsecret") {
		t.Errorf("GoString leaked value: %s", gs)
	}
	if !strings.Contains(gs, "redacted") || !strings.Contains(gs, "secretref://env/X") {
		t.Errorf("GoString should redact with ref: %s", gs)
	}
	if zg := (Secret{}).GoString(); !strings.Contains(zg, "[redacted]") {
		t.Errorf("zero GoString = %s", zg)
	}
}

// ---------- FingerprintOf error path ----------

func TestFingerprintOfMarshalError(t *testing.T) {
	// A channel cannot be JSON-marshaled, so FingerprintOf must return an error.
	_, err := FingerprintOf(make(chan int))
	if err == nil {
		t.Fatal("expected a marshal error")
	}
	if !strings.Contains(err.Error(), "fingerprint") {
		t.Errorf("error missing context: %v", err)
	}
}

// ---------- ModuleView.Decode marshal error ----------

func TestMapViewDecodeMarshalError(t *testing.T) {
	// A channel value makes json.Marshal of the namespace fail.
	err := MapView{"bad": make(chan int)}.Decode(&struct{}{})
	if err == nil {
		t.Fatal("expected an encode error")
	}
	if !strings.Contains(err.Error(), "encode module namespace") {
		t.Errorf("error missing context: %v", err)
	}
}

// ---------- Schema edge cases ----------

type schemaEdge struct {
	When    time.Time      `conf:"when" json:"when"`
	Meta    map[string]int `conf:"meta" json:"meta"` // unhandled kind → empty schema
	hidden  int            // unexported → skipped
	Ignored string         `conf:"-" json:"-"` // conf:"-" → skipped
}

func TestSchemaEdgeCases(t *testing.T) {
	// Non-struct target errors.
	if _, err := Schema[int](); err == nil {
		t.Fatal("Schema[int] should error")
	}

	js, err := Schema[schemaEdge]()
	if err != nil {
		t.Fatal(err)
	}
	s := string(js)
	if !strings.Contains(s, "date-time") {
		t.Errorf("time.Time field should render date-time format: %s", s)
	}
	if !strings.Contains(s, `"when"`) {
		t.Errorf("when property missing: %s", s)
	}
	if strings.Contains(s, "ignored") {
		t.Errorf("conf:\"-\" field must not appear: %s", s)
	}
	if strings.Contains(s, "hidden") {
		t.Errorf("unexported field must not appear: %s", s)
	}
	_ = schemaEdge{}.hidden // keep the field referenced
}
