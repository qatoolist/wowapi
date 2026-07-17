package seeds_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/seeds"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// baseBundlePerm returns a minimal, deterministic bundle for apply tests.
func baseBundlePerm() seeds.Bundle {
	return seeds.Bundle{
		Permissions: []seeds.PermissionSeed{
			{Key: "seedcov.doc.read", Description: "read a doc"},
			{Key: "seedcov.doc.write", Description: "write a doc"},
		},
	}
}

// TestApplyIdempotentNoop proves the strongest idempotency guarantee: the
// second Apply against an unchanged manifest short-circuits to outcome "noop"
// and does not touch the existing catalog rows (same xmin).
func TestApplyIdempotentNoop(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	b := baseBundlePerm()

	r1, err := seeds.Apply(ctx, h.Platform, b, seeds.ApplyOptions{Actor: "test"})
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if r1.Outcome != "applied" {
		t.Fatalf("first outcome = %q, want applied", r1.Outcome)
	}

	xminBefore := scanString(t, h.Platform,
		`SELECT xmin::text FROM permissions WHERE key = 'seedcov.doc.read'`)

	r2, err := seeds.Apply(ctx, h.Platform, b, seeds.ApplyOptions{Actor: "test"})
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if r2.Outcome != "noop" {
		t.Fatalf("second outcome = %q, want noop", r2.Outcome)
	}

	xminAfter := scanString(t, h.Platform,
		`SELECT xmin::text FROM permissions WHERE key = 'seedcov.doc.read'`)
	if xminBefore != xminAfter {
		t.Fatalf("second Apply touched catalog rows: xmin %s -> %s", xminBefore, xminAfter)
	}
}

// TestApplyDryRunNoWrites proves a dry-run produces a change plan and writes
// nothing to the database.
func TestApplyDryRunNoWrites(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	b := baseBundlePerm()

	var out bytes.Buffer
	r, err := seeds.Apply(ctx, h.Platform, b, seeds.ApplyOptions{
		DryRun: true,
		Actor:  "test",
		Out:    &out,
	})
	if err != nil {
		t.Fatalf("dry-run apply: %v", err)
	}
	if r.Outcome != "dry_run" {
		t.Fatalf("dry-run outcome = %q, want dry_run", r.Outcome)
	}
	if r.ChangePlan.Permissions.Insert != 2 {
		t.Fatalf("dry-run plan insert = %d, want 2", r.ChangePlan.Permissions.Insert)
	}
	if got := scanInt(t, h.Platform, `SELECT count(*) FROM permissions`); got != 0 {
		t.Fatalf("dry-run wrote %d permissions, want 0", got)
	}
	if !strings.Contains(out.String(), "manifest hash") {
		t.Fatalf("dry-run output missing hash: %q", out.String())
	}
}

// TestApplyRecordsAuditRow proves every successful apply writes a durable
// seed_sync_runs record carrying the hash, actor, outcome, and counts.
func TestApplyRecordsAuditRow(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	b := baseBundlePerm()

	r, err := seeds.Apply(ctx, h.Platform, b, seeds.ApplyOptions{Actor: "migrate-test"})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	var hash, actor, outcome string
	var counts string
	if err := h.Platform.QueryRow(ctx,
		`SELECT manifest_hash, actor, outcome, counts::text FROM seed_sync_runs
		  ORDER BY created_at DESC LIMIT 1`).Scan(&hash, &actor, &outcome, &counts); err != nil {
		t.Fatalf("query audit row: %v", err)
	}
	if hash != r.Hash {
		t.Fatalf("audit hash = %q, want %q", hash, r.Hash)
	}
	if actor != "migrate-test" {
		t.Fatalf("audit actor = %q, want migrate-test", actor)
	}
	if outcome != "applied" {
		t.Fatalf("audit outcome = %q, want applied", outcome)
	}
	if !strings.Contains(counts, `"permissions": 2`) {
		t.Fatalf("audit counts missing permissions: %q", counts)
	}
}

// TestApplyHashStableAcrossOrdering proves the content hash is independent of
// declaration order (canonicalization).
func TestApplyHashStableAcrossOrdering(t *testing.T) {
	b1 := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{
			{Key: "a.b", Description: "first"},
			{Key: "a.c", Description: "second"},
		},
	}
	b2 := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{
			{Key: "a.c", Description: "second"},
			{Key: "a.b", Description: "first"},
		},
	}
	if seeds.Hash(b1) != seeds.Hash(b2) {
		t.Fatalf("hash should be stable across ordering: %q vs %q", seeds.Hash(b1), seeds.Hash(b2))
	}
}

// TestApplyHashExcludesVersionLabel proves a manifest label bump does not
// change the content hash (and therefore a re-apply is a no-op).
func TestApplyHashExcludesVersionLabel(t *testing.T) {
	b1 := seeds.Bundle{Version: "v1", Permissions: []seeds.PermissionSeed{{Key: "a.b", Description: "x"}}}
	b2 := seeds.Bundle{Version: "v2", Permissions: []seeds.PermissionSeed{{Key: "a.b", Description: "x"}}}
	if seeds.Hash(b1) != seeds.Hash(b2) {
		t.Fatalf("hash must exclude version label: %q vs %q", seeds.Hash(b1), seeds.Hash(b2))
	}
}

// TestApplyRLSPostureRespectsPlatformRole proves the sync writes the global
// catalogs under app_platform, does not require superuser/BYPASSRLS, and
// app_rt cannot write seed_sync_runs.
func TestApplyRLSPostureRespectsPlatformRole(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	var bypass bool
	if err := h.Platform.QueryRow(ctx,
		`SELECT rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&bypass); err != nil {
		t.Fatalf("query platform role: %v", err)
	}
	if bypass {
		t.Fatal("platform role must not have BYPASSRLS")
	}

	b := baseBundlePerm()
	if _, err := seeds.Apply(ctx, h.Platform, b, seeds.ApplyOptions{Actor: "test"}); err != nil {
		t.Fatalf("apply as platform: %v", err)
	}

	// app_rt must not be able to append to the global sync audit log.
	if _, err := h.Runtime.Exec(ctx,
		`INSERT INTO seed_sync_runs (manifest_hash, outcome) VALUES ('x', 'failed')`); err == nil {
		t.Fatal("app_rt must not be able to INSERT into seed_sync_runs")
	}
}
