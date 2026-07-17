package testkit

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
)

// TestDiscardWriter covers the contract logger's throwaway sink.
func TestDiscardWriter(t *testing.T) {
	payload := []byte("some log line")
	n, err := discard{}.Write(payload)
	if err != nil {
		t.Fatalf("discard.Write err = %v", err)
	}
	if n != len(payload) {
		t.Fatalf("discard.Write n = %d, want %d", n, len(payload))
	}
}

// TestWithTenantHelper covers withTenant on both the nil-map and populated-map
// paths: it must always force tenant_id to the given id.
func TestWithTenantHelper(t *testing.T) {
	id := uuid.New()

	got := withTenant(nil, id)
	if got["tenant_id"] != id {
		t.Fatalf("withTenant(nil) tenant_id = %v, want %v", got["tenant_id"], id)
	}

	other := uuid.New()
	got = withTenant(map[string]any{"tenant_id": other, "note": "x"}, id)
	if got["tenant_id"] != id {
		t.Fatalf("withTenant overwrite tenant_id = %v, want %v", got["tenant_id"], id)
	}
	if got["note"] != "x" {
		t.Fatal("withTenant dropped an existing column")
	}
}

// TestAssertTablesRLSEmpty covers the no-tables early return of assertTablesRLS
// (a module with no migrations creates no tables to RLS-check, which is allowed).
func TestAssertTablesRLSEmpty(t *testing.T) {
	// declaredMigrations=false + created=nil → the permitted early return.
	assertTablesRLS(t, nil, "cov.module", nil, false)
}

// TestInsertRowRejectsInvalidColumn covers insertRow's column-name validation:
// a column that is not a valid identifier must be refused before any SQL runs.
func TestIntegrationInsertRowRejectsInvalidColumn(t *testing.T) {
	h := NewDB(t)
	table := CreateProbeTable(t, h)
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return insertRow(ctx, db, table, map[string]any{
			"id":          uuid.New(),
			"tenant_id":   tenant,
			"bad-column!": "x", // not a valid identifier
		})
	})
	if err == nil || !strings.Contains(err.Error(), "invalid column name") {
		t.Fatalf("insertRow with invalid column = %v, want an invalid-column error", err)
	}
}

// TestRequireDB covers the exported RequireDB policy flag. It sets the env var
// itself (never clearing the suite-wide value) so it is deterministic whether or
// not the ambient WOWAPI_REQUIRE_DB is set — a plain `go test ./...` must pass.
func TestRequireDB(t *testing.T) {
	t.Setenv("WOWAPI_REQUIRE_DB", "1")
	if !RequireDB() {
		t.Fatal("RequireDB() = false when WOWAPI_REQUIRE_DB is set")
	}
}

// TestTestDBNameTruncation drives testDBName from a deliberately long subtest
// name so the >63-char truncation branch is exercised, and asserts the result
// stays within Postgres's identifier limit and prefix/suffix invariants.
func TestTestDBNameTruncation(t *testing.T) {
	longName := strings.Repeat("abcXYZ_", 20) // ~140 chars, well over the budget
	t.Run(longName, func(st *testing.T) {
		name := testDBName(st)
		if len(name) > 63 {
			st.Fatalf("testDBName len = %d, want <= 63", len(name))
		}
		if !strings.HasPrefix(name, "wt_") {
			st.Fatalf("testDBName = %q, want wt_ prefix", name)
		}
	})
}
