package pagination_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/pagination"
)

func TestCursorRoundTrip(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	ts := time.Date(2026, 7, 3, 12, 30, 0, 0, time.UTC)

	in := map[string]any{
		"id":         id,
		"created_at": ts,
		"seq":        int64(9_223_372_036_854_775_000), // near int64 max: must not lose precision
		"score":      3.5,
		"name":       "alice",
		"active":     true,
	}
	s, err := pagination.EncodeCursor(in)
	if err != nil {
		t.Fatalf("EncodeCursor: %v", err)
	}
	if s == "" {
		t.Fatal("expected non-empty cursor")
	}

	cur, err := pagination.DecodeCursor(s)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	got := cur.Values()

	if got["id"] != id.String() {
		t.Errorf("id: got %v (%T), want %q", got["id"], got["id"], id.String())
	}
	if got["created_at"] != ts.Format(time.RFC3339Nano) {
		t.Errorf("created_at: got %v, want %q", got["created_at"], ts.Format(time.RFC3339Nano))
	}
	if got["seq"] != int64(9_223_372_036_854_775_000) {
		t.Errorf("seq: got %v (%T), want int64 preserved", got["seq"], got["seq"])
	}
	if got["score"] != 3.5 {
		t.Errorf("score: got %v (%T), want 3.5 float64", got["score"], got["score"])
	}
	if got["name"] != "alice" {
		t.Errorf("name: got %v", got["name"])
	}
	if got["active"] != true {
		t.Errorf("active: got %v", got["active"])
	}
}

func TestEncodeEmptyCursorIsZero(t *testing.T) {
	s, err := pagination.EncodeCursor(nil)
	if err != nil {
		t.Fatalf("EncodeCursor(nil): %v", err)
	}
	if s != "" {
		t.Errorf("empty map should encode to \"\", got %q", s)
	}
	cur, err := pagination.DecodeCursor("")
	if err != nil {
		t.Fatalf("DecodeCursor(\"\"): %v", err)
	}
	if !cur.IsZero() {
		t.Error("decoded empty cursor should be zero")
	}
	if cur.Values() != nil {
		t.Error("zero cursor Values() should be nil")
	}
}

func TestDecodeMalformedCursor(t *testing.T) {
	cases := map[string]string{
		"not base64":     "!!!not-base64!!!",
		"garbage base64": "Zm9vYmFy",                    // base64 of "foobar", not JSON
		"json array":     mustB64(t, `[1,2,3]`),         // not an object
		"json scalar":    mustB64(t, `42`),              // not an object
		"trailing data":  mustB64(t, `{"a":1} {"b":2}`), // trailing content
		"truncated":      mustB64(t, `{"a":1`),          // truncated object
		"oversized":      strings.Repeat("A", 4097),     // exceeds maxCursorLen
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := pagination.DecodeCursor(in)
			if err == nil {
				t.Fatalf("expected error for %q", in)
			}
			if k := errors.KindOf(err); k != errors.KindValidation {
				t.Errorf("expected KindValidation, got %v", k)
			}
		})
	}
}

func TestParsePerPageClamping(t *testing.T) {
	def := pagination.Defaults{PerPage: 20, MaxPerPage: 100}
	cases := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{"", 20, false},     // empty → default
		{"0", 20, false},    // zero → default
		{"50", 50, false},   // in range → as-is
		{"500", 100, false}, // over max → max
		{"-1", 0, true},     // negative → error
		{"abc", 0, true},    // non-integer → error
	}
	for _, c := range cases {
		req, err := pagination.Parse(c.raw, "", def)
		if c.wantErr {
			if err == nil {
				t.Errorf("per_page=%q: expected error", c.raw)
				continue
			}
			if k := errors.KindOf(err); k != errors.KindValidation {
				t.Errorf("per_page=%q: expected KindValidation, got %v", c.raw, k)
			}
			continue
		}
		if err != nil {
			t.Errorf("per_page=%q: unexpected error: %v", c.raw, err)
			continue
		}
		if req.Limit != c.want {
			t.Errorf("per_page=%q: got Limit %d, want %d", c.raw, req.Limit, c.want)
		}
	}
}

func TestParseCarriesCursor(t *testing.T) {
	s, err := pagination.EncodeCursor(map[string]any{"id": int64(7)})
	if err != nil {
		t.Fatal(err)
	}
	req, err := pagination.Parse("10", s, pagination.Defaults{PerPage: 20, MaxPerPage: 100})
	if err != nil {
		t.Fatal(err)
	}
	if req.Limit != 10 {
		t.Errorf("Limit: got %d, want 10", req.Limit)
	}
	if req.Cursor.IsZero() {
		t.Error("expected non-zero cursor")
	}
	if req.Cursor.Values()["id"] != int64(7) {
		t.Errorf("cursor id: got %v", req.Cursor.Values()["id"])
	}
}

func TestParseRejectsBadCursor(t *testing.T) {
	_, err := pagination.Parse("10", "!!!bad!!!", pagination.Defaults{PerPage: 20, MaxPerPage: 100})
	if err == nil {
		t.Fatal("expected error for bad cursor")
	}
	if k := errors.KindOf(err); k != errors.KindValidation {
		t.Errorf("expected KindValidation, got %v", k)
	}
}

func TestEncodeUnsupportedType(t *testing.T) {
	_, err := pagination.EncodeCursor(map[string]any{"bad": struct{ X int }{1}})
	if err == nil {
		t.Fatal("expected error for unsupported cursor value type")
	}
}

func mustB64(t *testing.T, s string) string {
	t.Helper()
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
