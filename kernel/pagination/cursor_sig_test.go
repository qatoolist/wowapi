package pagination_test

import (
	"encoding/base64"
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

func TestCursorSigEmptyRejected(t *testing.T) {
	if _, err := pagination.EncodeCursorWithSig("", map[string]any{"id": int64(1)}); err == nil {
		t.Fatal("empty cursor signature must be rejected")
	}
}

func TestFlatCursorRejected(t *testing.T) {
	flat := base64.RawURLEncoding.EncodeToString([]byte(`{"id":9}`))
	if _, err := pagination.DecodeCursor(flat); err == nil {
		t.Fatal("flat unsigned cursor must be rejected")
	}
}
