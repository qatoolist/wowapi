package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------- apikey: argument validation (no DB needed) ----------

func TestApikeyNoArgsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runApikey(nil, &out, &errb); code != 2 {
		t.Fatalf("no args should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi apikey") {
		t.Fatalf("expected usage on stderr, got %q", errb.String())
	}
}

func TestApikeyUnknownAction(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runApikey([]string{"frobnicate"}, &out, &errb); code != 2 {
		t.Fatalf("unknown action should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi apikey") {
		t.Fatalf("expected usage on stderr, got %q", errb.String())
	}
}

func TestApikeyBadTenant(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runApikey([]string{"list", "--tenant", "not-a-uuid"}, &out, &errb); code != 2 {
		t.Fatalf("bad tenant should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "--tenant must be a uuid") {
		t.Fatalf("expected tenant error, got %q", errb.String())
	}
}

func TestApikeyMissingDSN(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	var out, errb bytes.Buffer
	code := runApikey([]string{"list", "--tenant", uuid.New().String()}, &out, &errb)
	if code != 1 {
		t.Fatalf("missing DATABASE_URL should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "DATABASE_URL is not set") {
		t.Fatalf("expected DSN error, got %q", errb.String())
	}
}

// ---------- apikey: DB-backed lifecycle ----------

func TestApikeyIssueListRotateRevokeDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	tenant := uuid.New()
	t.Cleanup(func() { cleanupTenant(t, pool, tenant) })

	// issue
	var out, errb bytes.Buffer
	code := runApikey([]string{
		"issue", "--tenant", tenant.String(), "--name", "ci-key",
		"--scopes", "widgets.widget.read,widgets.widget.write", "--expires", "720h",
	}, &out, &errb)
	if code != 0 {
		t.Fatalf("issue exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "issued key") || !strings.Contains(out.String(), "token (shown once)") {
		t.Fatalf("issue output missing expected lines: %q", out.String())
	}

	// The key must actually exist in the tenant's rows.
	var id uuid.UUID
	if err := pool.QueryRow(context.Background(),
		`SELECT id FROM api_keys WHERE tenant_id = $1 AND name = 'ci-key'`, tenant).Scan(&id); err != nil {
		t.Fatalf("issued key not found in api_keys: %v", err)
	}

	// list shows the active key with its scopes.
	out.Reset()
	errb.Reset()
	if code := runApikey([]string{"list", "--tenant", tenant.String()}, &out, &errb); code != 0 {
		t.Fatalf("list exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "ci-key") || !strings.Contains(out.String(), "active") {
		t.Fatalf("list output missing key/status: %q", out.String())
	}
	if !strings.Contains(out.String(), "widgets.widget.read") {
		t.Fatalf("list output missing scopes: %q", out.String())
	}

	// rotate mints a new secret for the same logical key.
	out.Reset()
	errb.Reset()
	if code := runApikey([]string{"rotate", "--tenant", tenant.String(), "--id", id.String()}, &out, &errb); code != 0 {
		t.Fatalf("rotate exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "rotated key") || !strings.Contains(out.String(), "token (shown once)") {
		t.Fatalf("rotate output unexpected: %q", out.String())
	}

	// revoke the original key.
	out.Reset()
	errb.Reset()
	if code := runApikey([]string{"revoke", "--tenant", tenant.String(), "--id", id.String()}, &out, &errb); code != 0 {
		t.Fatalf("revoke exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "revoked key") {
		t.Fatalf("revoke output unexpected: %q", out.String())
	}
	// The revoked key must now report status "revoked" in list.
	out.Reset()
	errb.Reset()
	_ = runApikey([]string{"list", "--tenant", tenant.String()}, &out, &errb)
	if !strings.Contains(out.String(), "revoked") {
		t.Fatalf("expected revoked status after revoke, got %q", out.String())
	}
}

func TestApikeyListEmptyDB(t *testing.T) {
	requireDSN(t)
	tenant := uuid.New() // fresh tenant with no keys
	var out, errb bytes.Buffer
	if code := runApikey([]string{"list", "--tenant", tenant.String()}, &out, &errb); code != 0 {
		t.Fatalf("list exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "no keys") {
		t.Fatalf("empty tenant should print 'no keys', got %q", out.String())
	}
}

func TestApikeyIssueRequiresName(t *testing.T) {
	requireDSN(t)
	var out, errb bytes.Buffer
	code := runApikey([]string{"issue", "--tenant", uuid.New().String()}, &out, &errb)
	if code != 2 {
		t.Fatalf("issue without --name should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "--name is required") {
		t.Fatalf("expected name error, got %q", errb.String())
	}
}

func TestApikeyRotateBadID(t *testing.T) {
	requireDSN(t)
	var out, errb bytes.Buffer
	if code := runApikey([]string{"rotate", "--tenant", uuid.New().String(), "--id", "nope"}, &out, &errb); code != 2 {
		t.Fatalf("rotate with bad id should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "--id must be a uuid") {
		t.Fatalf("expected id error, got %q", errb.String())
	}
}

func TestApikeyRevokeBadID(t *testing.T) {
	requireDSN(t)
	var out, errb bytes.Buffer
	if code := runApikey([]string{"revoke", "--tenant", uuid.New().String(), "--id", "nope"}, &out, &errb); code != 2 {
		t.Fatalf("revoke with bad id should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "--id must be a uuid") {
		t.Fatalf("expected id error, got %q", errb.String())
	}
}
