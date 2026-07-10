package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestSeedSyncDB is the GAP-003 core proof: `wowapi seed sync` connects to a
// real (migrated) database via DATABASE_URL and upserts a module's seed
// bundle into the platform catalog tables — the production lifecycle path
// that was previously only exercised by testkit / a hand-written product
// migrate main (see docs/upstream/02-pf-9-no-production-seed-sync-path.md in
// the wowsociety product repo this was upstreamed from).
func TestSeedSyncDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)

	dir := t.TempDir()
	writeFile(t, dir, "permissions.yaml",
		"permissions:\n  - key: widgets.widget.create\n    description: create a widget\n"+
			"resource_types:\n  - key: widgets.widget\n    description: a widget\n")

	var out, errb bytes.Buffer
	code := runSeed([]string{"sync", "--module", "widgets=" + dir}, &out, &errb)
	if code != 0 {
		t.Fatalf("seed sync exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "OK") || !strings.Contains(out.String(), "1 permissions") {
		t.Fatalf("unexpected seed sync output: %q", out.String())
	}

	var permCount, rtCount int
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM permissions WHERE key = 'widgets.widget.create'`).Scan(&permCount); err != nil {
		t.Fatalf("count permissions: %v", err)
	}
	if permCount != 1 {
		t.Fatalf("permission not synced, count = %d", permCount)
	}
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM resource_types WHERE key = 'widgets.widget'`).Scan(&rtCount); err != nil {
		t.Fatalf("count resource_types: %v", err)
	}
	if rtCount != 1 {
		t.Fatalf("resource_type not synced, count = %d", rtCount)
	}
}

// TestSeedSyncDBIdempotent is the GAP-003 idempotency acceptance criterion:
// running seed sync twice must converge with no error and no duplicate rows —
// seeds.Sync's ON CONFLICT upserts, exercised end-to-end through the CLI path
// a real deploy uses (migrate re-runs on every deploy).
func TestSeedSyncDBIdempotent(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)

	dir := t.TempDir()
	writeFile(t, dir, "seeds.yaml",
		"permissions:\n  - key: widgets.widget.create\n    description: create a widget\n"+
			"resource_types:\n  - key: widgets.widget\n    description: a widget\n"+
			"relationship_types:\n  - key: widgets.owner_of\n    subject_kind: party\n    object_kind: resource\n    description: owns a widget\n"+
			"roles:\n  - key: widgets.editor\n    name: Widget Editor\n    permissions: [widgets.widget.create]\n")

	for i := 0; i < 2; i++ {
		var out, errb bytes.Buffer
		code := runSeed([]string{"sync", "--module", "widgets=" + dir}, &out, &errb)
		if code != 0 {
			t.Fatalf("run %d: seed sync exit %d: %s", i, code, errb.String())
		}
	}

	assertCount := func(query string, want int) {
		t.Helper()
		var n int
		if err := pool.QueryRow(t.Context(), query).Scan(&n); err != nil {
			t.Fatalf("query %q: %v", query, err)
		}
		if n != want {
			t.Fatalf("query %q = %d, want %d (duplicate rows after re-sync)", query, n, want)
		}
	}
	assertCount(`SELECT count(*) FROM permissions WHERE key = 'widgets.widget.create'`, 1)
	assertCount(`SELECT count(*) FROM resource_types WHERE key = 'widgets.widget'`, 1)
	assertCount(`SELECT count(*) FROM relationship_types WHERE key = 'widgets.owner_of'`, 1)
	assertCount(`SELECT count(*) FROM roles WHERE key = 'widgets.editor'`, 1)
	assertCount(`SELECT count(*) FROM role_permissions rp JOIN roles r ON r.id = rp.role_id WHERE r.key = 'widgets.editor'`, 1)
}

// TestSeedSyncDBMultiModule proves the merged-bundle path: two modules' seed
// directories are loaded and synced together in one invocation (the shape a
// real multi-module product needs — a single privileged connection applying
// every module's catalog, matching what app.Boot merges in-memory at runtime).
func TestSeedSyncDBMultiModule(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)

	dirA := t.TempDir()
	writeFile(t, dirA, "permissions.yaml", "permissions:\n  - key: widgets.widget.create\n    description: c\n")
	dirB := t.TempDir()
	writeFile(t, dirB, "permissions.yaml", "permissions:\n  - key: gadgets.gadget.create\n    description: c\n")

	var out, errb bytes.Buffer
	code := runSeed([]string{"sync", "--module", "widgets=" + dirA, "--module", "gadgets=" + dirB}, &out, &errb)
	if code != 0 {
		t.Fatalf("seed sync exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "2 permissions") {
		t.Fatalf("expected merged bundle of 2 permissions, got %q", out.String())
	}

	var n int
	if err := pool.QueryRow(t.Context(),
		`SELECT count(*) FROM permissions WHERE key IN ('widgets.widget.create','gadgets.gadget.create')`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected both modules' permissions synced, count = %d", n)
	}
}
