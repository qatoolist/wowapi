package seeds_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/testkit"
)

// scanString runs a single-column, single-row query on the platform pool and
// returns the string result, failing the test on error.
func scanString(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) string {
	t.Helper()
	var v string
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return v
}

func scanBool(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) bool {
	t.Helper()
	var v bool
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return v
}

func scanInt(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) int {
	t.Helper()
	var v int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&v); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return v
}

// grantsOf returns the sorted permission_key set granted to a role.
func grantsOf(t *testing.T, pool *pgxpool.Pool, roleID string) []string {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT permission_key FROM role_permissions WHERE role_id = $1 ORDER BY permission_key`, roleID)
	if err != nil {
		t.Fatalf("query grants: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			t.Fatalf("scan grant: %v", err)
		}
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate grants: %v", err)
	}
	return out
}

func eqStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// baseBundle is the catalog a "seedcov" module declares. moduleOf must derive
// the module column ("seedcov") from every dotted key.
func baseBundle() seeds.Bundle {
	return seeds.Bundle{
		Permissions: []seeds.PermissionSeed{
			{Key: "seedcov.doc.read", Description: "read a doc", Sensitive: false, GrantedVia: "seedcov.owns"},
			{Key: "seedcov.doc.write", Description: "write a doc", Sensitive: true},
		},
		ResourceTypes: []seeds.ResourceTypeSeed{
			{Key: "seedcov.doc", Description: "a document"},
		},
		RelationshipTypes: []seeds.RelationshipTypeSeed{
			// Empty cardinality must default to "many" (Sync fills the blank).
			{Key: "seedcov.owns", SubjectKind: "party", ObjectKind: "resource", Description: "ownership"},
			{Key: "seedcov.assigned", SubjectKind: "capacity", ObjectKind: "resource", Cardinality: "one", Description: "assignment"},
		},
		Roles: []seeds.RoleSeed{
			{Key: "seedcov.member", Name: "Member", Permissions: []string{"seedcov.doc.read", "seedcov.doc.write"}},
		},
	}
}

// TestSyncLifecycle exercises Sync end to end against a real Postgres: initial
// apply, idempotent re-run, in-place upsert of changed fields, and reconciling
// (pruning) a role's grant set when the seed drops a permission.
func TestSyncLifecycle(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	p := h.Platform

	// --- 1. Initial apply: every catalog table is populated correctly. ---
	if err := seeds.Sync(ctx, p, baseBundle()); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	if got := scanInt(t, p, `SELECT count(*) FROM permissions WHERE module = 'seedcov'`); got != 2 {
		t.Fatalf("permissions count = %d, want 2", got)
	}
	// moduleOf derived the module column from the dotted key.
	if got := scanString(t, p, `SELECT module FROM permissions WHERE key = 'seedcov.doc.read'`); got != "seedcov" {
		t.Fatalf("permission module = %q, want seedcov", got)
	}
	if got := scanString(t, p, `SELECT description FROM permissions WHERE key = 'seedcov.doc.read'`); got != "read a doc" {
		t.Fatalf("permission description = %q, want 'read a doc'", got)
	}
	if scanBool(t, p, `SELECT sensitive FROM permissions WHERE key = 'seedcov.doc.read'`) {
		t.Fatal("seedcov.doc.read should not be sensitive")
	}
	if !scanBool(t, p, `SELECT sensitive FROM permissions WHERE key = 'seedcov.doc.write'`) {
		t.Fatal("seedcov.doc.write should be sensitive")
	}

	if got := scanString(t, p, `SELECT module FROM resource_types WHERE key = 'seedcov.doc'`); got != "seedcov" {
		t.Fatalf("resource_type module = %q, want seedcov", got)
	}

	// Empty cardinality defaulted to "many"; explicit "one" preserved.
	if got := scanString(t, p, `SELECT cardinality FROM relationship_types WHERE key = 'seedcov.owns'`); got != "many" {
		t.Fatalf("owns cardinality = %q, want many (default)", got)
	}
	if got := scanString(t, p, `SELECT cardinality FROM relationship_types WHERE key = 'seedcov.assigned'`); got != "one" {
		t.Fatalf("assigned cardinality = %q, want one", got)
	}
	if got := scanString(t, p, `SELECT subject_kind FROM relationship_types WHERE key = 'seedcov.assigned'`); got != "capacity" {
		t.Fatalf("assigned subject_kind = %q, want capacity", got)
	}

	// Role is a platform template (tenant_id NULL, is_system true).
	roleID := scanString(t, p, `SELECT id FROM roles WHERE key = 'seedcov.member'`)
	if roleID == "" {
		t.Fatal("role seedcov.member not inserted")
	}
	if scanInt(t, p, `SELECT count(*) FROM roles WHERE key='seedcov.member' AND tenant_id IS NULL AND is_system`) != 1 {
		t.Fatal("role should be a NULL-tenant system template")
	}
	if got := scanString(t, p, `SELECT name FROM roles WHERE key = 'seedcov.member'`); got != "Member" {
		t.Fatalf("role name = %q, want Member", got)
	}
	want := []string{"seedcov.doc.read", "seedcov.doc.write"}
	if got := grantsOf(t, p, roleID); !eqStrings(got, want) {
		t.Fatalf("role grants = %v, want %v", got, want)
	}

	// --- 2. Idempotent re-run: no duplicates, same role id, same grants. ---
	if err := seeds.Sync(ctx, p, baseBundle()); err != nil {
		t.Fatalf("idempotent re-sync: %v", err)
	}
	if got := scanInt(t, p, `SELECT count(*) FROM permissions WHERE module='seedcov'`); got != 2 {
		t.Fatalf("permissions count after re-sync = %d, want 2 (no dupes)", got)
	}
	if got := scanInt(t, p, `SELECT count(*) FROM roles WHERE key='seedcov.member'`); got != 1 {
		t.Fatalf("roles count after re-sync = %d, want 1 (no dupes)", got)
	}
	if got := scanString(t, p, `SELECT id FROM roles WHERE key='seedcov.member'`); got != roleID {
		t.Fatalf("role id changed on re-sync: %q -> %q (reseed must update same row)", roleID, got)
	}
	if got := grantsOf(t, p, roleID); !eqStrings(got, want) {
		t.Fatalf("grants after re-sync = %v, want %v", got, want)
	}

	// --- 3. Upsert in place: mutate fields, re-sync, assert DO UPDATE ran. ---
	updated := baseBundle()
	updated.Permissions[0].Description = "read a doc (v2)"
	updated.Permissions[0].Sensitive = true
	updated.ResourceTypes[0].Description = "a document (v2)"
	updated.RelationshipTypes[1].Cardinality = "many" // one -> many
	updated.RelationshipTypes[1].Description = "assignment (v2)"
	updated.Roles[0].Name = "Member (v2)"
	if err := seeds.Sync(ctx, p, updated); err != nil {
		t.Fatalf("upsert re-sync: %v", err)
	}
	if got := scanString(t, p, `SELECT description FROM permissions WHERE key='seedcov.doc.read'`); got != "read a doc (v2)" {
		t.Fatalf("permission description not updated: %q", got)
	}
	if !scanBool(t, p, `SELECT sensitive FROM permissions WHERE key='seedcov.doc.read'`) {
		t.Fatal("permission sensitive flag not updated to true")
	}
	if got := scanString(t, p, `SELECT description FROM resource_types WHERE key='seedcov.doc'`); got != "a document (v2)" {
		t.Fatalf("resource_type description not updated: %q", got)
	}
	if got := scanString(t, p, `SELECT cardinality FROM relationship_types WHERE key='seedcov.assigned'`); got != "many" {
		t.Fatalf("relationship_type cardinality not updated: %q", got)
	}
	if got := scanString(t, p, `SELECT name FROM roles WHERE key='seedcov.member'`); got != "Member (v2)" {
		t.Fatalf("role name not updated: %q", got)
	}
	// Same row (id) throughout — an update, not an insert.
	if got := scanString(t, p, `SELECT id FROM roles WHERE key='seedcov.member'`); got != roleID {
		t.Fatalf("role id changed on upsert: %q -> %q", roleID, got)
	}

	// --- 4. Reconcile: drop a permission from the role, stale grant pruned. ---
	pruned := baseBundle()
	pruned.Roles[0].Permissions = []string{"seedcov.doc.read"} // drop doc.write
	if err := seeds.Sync(ctx, p, pruned); err != nil {
		t.Fatalf("prune re-sync: %v", err)
	}
	if got, wantOne := grantsOf(t, p, roleID), []string{"seedcov.doc.read"}; !eqStrings(got, wantOne) {
		t.Fatalf("stale grant not pruned: grants = %v, want %v", got, wantOne)
	}
	// The dropped permission still exists in the catalog — only the grant edge
	// was removed (least-privilege reconcile, not a catalog delete).
	if scanInt(t, p, `SELECT count(*) FROM permissions WHERE key='seedcov.doc.write'`) != 1 {
		t.Fatal("pruning a grant must not delete the permission itself")
	}
}

// TestSyncPersistsStepUp: Sync writes permissions.step_up on initial insert and
// flips it on re-sync when the seed changes — a permission's step-up
// requirement is a live, updatable catalog attribute, not a one-time insert.
func TestSyncPersistsStepUp(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	p := h.Platform

	b := seeds.Bundle{
		Permissions: []seeds.PermissionSeed{
			{Key: "seedcov.impersonation.assign", Description: "assign impersonation", StepUp: true},
			{Key: "seedcov.doc.read", Description: "read a doc"},
		},
	}
	if err := seeds.Sync(ctx, p, b); err != nil {
		t.Fatalf("initial sync: %v", err)
	}
	if !scanBool(t, p, `SELECT step_up FROM permissions WHERE key='seedcov.impersonation.assign'`) {
		t.Fatal("step_up=true not persisted on initial insert")
	}
	if scanBool(t, p, `SELECT step_up FROM permissions WHERE key='seedcov.doc.read'`) {
		t.Fatal("step_up should default to false when the seed omits it")
	}

	// Flip: the seed drops step_up on re-sync — idempotent updatable, not stuck.
	flipped := b
	flipped.Permissions = []seeds.PermissionSeed{
		{Key: "seedcov.impersonation.assign", Description: "assign impersonation", StepUp: false},
		{Key: "seedcov.doc.read", Description: "read a doc", StepUp: true},
	}
	if err := seeds.Sync(ctx, p, flipped); err != nil {
		t.Fatalf("flip re-sync: %v", err)
	}
	if scanBool(t, p, `SELECT step_up FROM permissions WHERE key='seedcov.impersonation.assign'`) {
		t.Fatal("step_up not flipped to false on re-sync")
	}
	if !scanBool(t, p, `SELECT step_up FROM permissions WHERE key='seedcov.doc.read'`) {
		t.Fatal("step_up not flipped to true on re-sync")
	}
}

// TestSyncEmptyBundle: syncing an empty bundle is a clean no-op that writes no
// catalog rows.
func TestSyncEmptyBundle(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	p := h.Platform

	if err := seeds.Sync(ctx, p, seeds.Bundle{}); err != nil {
		t.Fatalf("empty sync: %v", err)
	}
	if got := scanInt(t, p, `SELECT count(*) FROM permissions`); got != 0 {
		t.Fatalf("empty bundle wrote %d permissions, want 0", got)
	}
	if got := scanInt(t, p, `SELECT count(*) FROM roles`); got != 0 {
		t.Fatalf("empty bundle wrote %d roles, want 0", got)
	}
}

// TestSyncDotlessKeyModule pins moduleOf's fallback: a key with no '.' is its
// own module (the whole key becomes the module column).
func TestSyncDotlessKeyModule(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	p := h.Platform

	b := seeds.Bundle{
		ResourceTypes: []seeds.ResourceTypeSeed{
			{Key: "solo", Description: "a dotless resource type"},
		},
	}
	if err := seeds.Sync(ctx, p, b); err != nil {
		t.Fatalf("sync dotless key: %v", err)
	}
	if got := scanString(t, p, `SELECT module FROM resource_types WHERE key='solo'`); got != "solo" {
		t.Fatalf("dotless key module = %q, want 'solo' (moduleOf fallback)", got)
	}
}

// TestLoadRejectsEmptyKey pins validate's empty-key branch: a blank key is a
// seed error, not a silently-skipped entry.
func TestLoadRejectsEmptyKey(t *testing.T) {
	src := fsys(map[string]string{
		"p.yaml": "permissions:\n  - key: \"\"\n    description: blank\n",
	})
	_, err := seeds.Load(src, "requests")
	if err == nil {
		t.Fatal("an empty key must be rejected by validate")
	}
}

// TestSyncRejectsInvalidRelationshipType: Sync surfaces a DB CHECK violation
// (subject_kind not in the allowed set) as a wrapped error, not a silent skip.
func TestSyncRejectsInvalidRelationshipType(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	b := seeds.Bundle{
		RelationshipTypes: []seeds.RelationshipTypeSeed{
			{Key: "seedcov.bad", SubjectKind: "bogus", ObjectKind: "resource", Description: "invalid"},
		},
	}
	if err := seeds.Sync(ctx, h.Platform, b); err == nil {
		t.Fatal("Sync must fail when a relationship_type violates the subject_kind CHECK")
	}
	if got := scanInt(t, h.Platform, `SELECT count(*) FROM relationship_types WHERE key='seedcov.bad'`); got != 0 {
		t.Fatalf("invalid relationship_type should not have been written, count=%d", got)
	}
}

// TestSyncRejectsUngrantablePermission: a role that grants a permission not in
// the catalog trips the role_permissions FK — Sync returns a wrapped error.
func TestSyncRejectsUngrantablePermission(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	b := seeds.Bundle{
		Roles: []seeds.RoleSeed{
			{Key: "seedcov.admin", Name: "Admin", Permissions: []string{"seedcov.ghost"}},
		},
	}
	if err := seeds.Sync(ctx, h.Platform, b); err == nil {
		t.Fatal("Sync must fail when a role grants a permission absent from the catalog (FK)")
	}
}

// TestSyncViaLoad wires Load into Sync: a bundle parsed from embedded YAML syncs
// cleanly and the parsed granted_via / keys reach the catalog.
func TestSyncViaLoad(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	p := h.Platform

	src := fsys(map[string]string{
		"catalog.yaml": `
permissions:
  - key: loadcov.item.read
    description: read an item
resource_types:
  - key: loadcov.item
    description: an item
relationship_types:
  - key: loadcov.owns
    subject_kind: party
    object_kind: resource
roles:
  - key: loadcov.viewer
    name: Viewer
    permissions: [loadcov.item.read]
`,
	})
	b, err := seeds.Load(src, "loadcov")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := seeds.Sync(ctx, p, b); err != nil {
		t.Fatalf("sync loaded bundle: %v", err)
	}
	roleID := scanString(t, p, `SELECT id FROM roles WHERE key='loadcov.viewer'`)
	if got := grantsOf(t, p, roleID); !eqStrings(got, []string{"loadcov.item.read"}) {
		t.Fatalf("loaded role grants = %v, want [loadcov.item.read]", got)
	}
	if got := scanString(t, p, `SELECT module FROM permissions WHERE key='loadcov.item.read'`); got != "loadcov" {
		t.Fatalf("module = %q, want loadcov", got)
	}
}
