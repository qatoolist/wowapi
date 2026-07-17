package pagination_test

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/pagination"
)

func TestCursorSigRoundTrip(t *testing.T) {
	enc, err := pagination.EncodeCursorWithSig("created_at:asc,id:asc", map[string]any{"id": int64(7), "created_at": "2026-07-04T00:00:00Z"})
	if err != nil {
		t.Fatalf("EncodeCursorWithSig: %v", err)
	}
	cur, err := pagination.DecodeCursor(enc)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if cur.Sig() != "created_at:asc,id:asc" {
		t.Errorf("Sig = %q, want the embedded signature", cur.Sig())
	}
	vals := cur.Values()
	if vals["id"] != int64(7) {
		t.Errorf("values id = %v, want int64(7) — envelope must not corrupt the tuple", vals["id"])
	}
	if vals["created_at"] != "2026-07-04T00:00:00Z" {
		t.Errorf("values created_at = %v", vals["created_at"])
	}
}

func TestCursorSigEmptyFallsBackToFlat(t *testing.T) {
	// An empty signature must produce the legacy flat encoding (no sig).
	enc, err := pagination.EncodeCursorWithSig("", map[string]any{"id": int64(1)})
	if err != nil {
		t.Fatal(err)
	}
	cur, err := pagination.DecodeCursor(enc)
	if err != nil {
		t.Fatal(err)
	}
	if cur.Sig() != "" {
		t.Errorf("Sig = %q, want empty for a flat cursor", cur.Sig())
	}
	if cur.Values()["id"] != int64(1) {
		t.Errorf("flat values corrupted: %v", cur.Values())
	}
}

func TestFlatCursorHasNoSig(t *testing.T) {
	enc, err := pagination.EncodeCursor(map[string]any{"id": int64(9)})
	if err != nil {
		t.Fatal(err)
	}
	cur, err := pagination.DecodeCursor(enc)
	if err != nil {
		t.Fatal(err)
	}
	if cur.Sig() != "" {
		t.Errorf("legacy flat cursor reported Sig %q, want empty", cur.Sig())
	}
}
