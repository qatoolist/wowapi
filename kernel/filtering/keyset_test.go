package filtering_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/filtering"
	"github.com/qatoolist/wowapi/kernel/pagination"
)

func TestKeysetClauseEmpty(t *testing.T) {
	sql, args, next, err := filtering.KeysetClause(filtering.Sort{}, pagination.Cursor{}, 1)
	if err != nil || sql != "" || len(args) != 0 || next != 1 {
		t.Fatalf("empty sort/cursor should be a no-op: sql=%q args=%v next=%d err=%v", sql, args, next, err)
	}
}

func TestKeysetClauseLexicographic(t *testing.T) {
	allow := filtering.SortAllowlist{
		"created_at": {Col: "created_at"},
		"id":         {Col: "id"},
	}
	sort, err := filtering.ParseSort("created_at:desc,id:asc", allow)
	if err != nil {
		t.Fatal(err)
	}
	cur, err := pagination.DecodeCursor(mustEncode(t, sort.Signature(), map[string]any{"created_at": "2026-01-01T00:00:00Z", "id": "abc"}))
	if err != nil {
		t.Fatal(err)
	}
	sql, args, next, err := filtering.KeysetClause(sort, cur, 3)
	if err != nil {
		t.Fatal(err)
	}
	// desc → "<", asc → ">"; placeholders start at $3.
	want := "((created_at < $3) OR (created_at = $3 AND id > $4))"
	if sql != want {
		t.Errorf("sql = %q, want %q", sql, want)
	}
	if len(args) != 2 || next != 5 {
		t.Errorf("args=%v next=%d", args, next)
	}
}

func TestKeysetClauseRejectsMismatchedCursor(t *testing.T) {
	allow := filtering.SortAllowlist{"id": {Col: "id"}}
	sort, _ := filtering.ParseSort("id:asc", allow)
	// Cursor carries a DIFFERENT/extra key than the sort — a forged or stale
	// cursor. Must be rejected, never silently used (SEC-22).
	cur, _ := pagination.DecodeCursor(mustEncode(t, sort.Signature(), map[string]any{"evil_col": "x", "id": "y"}))
	_, _, _, err := filtering.KeysetClause(sort, cur, 1)
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("mismatched cursor should be KindValidation, got %v", err)
	}
}

func TestKeysetClauseColumnsAreAllowlisted(t *testing.T) {
	// Even with an attacker-controlled cursor value, only the allowlisted
	// column name and $N placeholders appear in the SQL.
	allow := filtering.SortAllowlist{"id": {Col: "id"}}
	sort, _ := filtering.ParseSort("id:asc", allow)
	payload := "x'; DROP TABLE users;--"
	cur, _ := pagination.DecodeCursor(mustEncode(t, sort.Signature(), map[string]any{"id": payload}))
	sql, args, _, err := filtering.KeysetClause(sort, cur, 1)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sql, "DROP") || strings.Contains(sql, payload) {
		t.Errorf("payload leaked into SQL: %q", sql)
	}
	if len(args) != 1 || args[0] != payload {
		t.Errorf("payload must be a bound arg: %v", args)
	}
}

func TestSortSignatureStableAndOrderSensitive(t *testing.T) {
	allow := filtering.SortAllowlist{"created_at": {Col: "created_at"}, "id": {Col: "id"}}
	a, _ := filtering.ParseSort("created_at:asc,id:asc", allow)
	b, _ := filtering.ParseSort("created_at:asc,id:asc", allow)
	if a.Signature() != b.Signature() || a.Signature() == "" {
		t.Fatalf("same sort must share a non-empty signature: %q vs %q", a.Signature(), b.Signature())
	}
	// Direction flip and column reorder must each change the signature — these are
	// exactly the cases the column-set check cannot distinguish.
	flip, _ := filtering.ParseSort("created_at:desc,id:asc", allow)
	reorder, _ := filtering.ParseSort("id:asc,created_at:asc", allow)
	if a.Signature() == flip.Signature() {
		t.Error("direction flip must change the signature")
	}
	if a.Signature() == reorder.Signature() {
		t.Error("column reorder must change the signature")
	}
}

func TestKeysetClauseRejectsSortSpecChange(t *testing.T) {
	allow := filtering.SortAllowlist{"created_at": {Col: "created_at"}, "id": {Col: "id"}}
	// Mint a cursor under the ORIGINAL sort (asc, asc) using the sig-aware helper.
	orig, _ := filtering.ParseSort("created_at:asc,id:asc", allow)
	enc, err := filtering.NextCursor(orig, map[string]any{"created_at": "2026-01-01T00:00:00Z", "id": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	cur, err := pagination.DecodeCursor(enc)
	if err != nil {
		t.Fatal(err)
	}
	// Same column SET, but the direction changed — previously a silent wrong-page.
	changed, _ := filtering.ParseSort("created_at:desc,id:asc", allow)
	if _, _, _, err := filtering.KeysetClause(changed, cur, 1); errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("a sort-order change must fail loudly with KindValidation, got %v", err)
	}
	// The SAME sort still works — round-trip is intact.
	if _, _, _, err := filtering.KeysetClause(orig, cur, 1); err != nil {
		t.Fatalf("cursor under its own sort must still decode: %v", err)
	}
}

func mustEncode(t *testing.T, sig string, m map[string]any) string {
	t.Helper()
	s, err := pagination.EncodeCursorWithSig(sig, m)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
