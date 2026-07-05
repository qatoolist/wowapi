package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------- audit: argument validation (no DB needed) ----------

func TestAuditNoArgsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runAudit(nil, &out, &errb); code != 2 {
		t.Fatalf("no args should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi audit verify") {
		t.Fatalf("expected usage on stderr, got %q", errb.String())
	}
}

func TestAuditUnknownSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runAudit([]string{"bogus"}, &out, &errb); code != 2 {
		t.Fatalf("unknown subcommand should exit 2, got %d", code)
	}
}

func TestAuditBadTenant(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runAudit([]string{"verify", "--tenant", "not-a-uuid"}, &out, &errb); code != 2 {
		t.Fatalf("bad tenant should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "--tenant must be a uuid") {
		t.Fatalf("expected tenant error, got %q", errb.String())
	}
}

func TestAuditMissingDSN(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	var out, errb bytes.Buffer
	if code := runAudit([]string{"verify", "--tenant", uuid.New().String()}, &out, &errb); code != 1 {
		t.Fatalf("missing DATABASE_URL should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "DATABASE_URL is not set") {
		t.Fatalf("expected DSN error, got %q", errb.String())
	}
}

// ---------- audit: DB-backed verification ----------

func TestAuditVerifyEmptyTenantDB(t *testing.T) {
	requireDSN(t)
	tenant := uuid.New() // no audit rows for this tenant
	var out, errb bytes.Buffer
	if code := runAudit([]string{"verify", "--tenant", tenant.String()}, &out, &errb); code != 0 {
		t.Fatalf("verify of empty tenant should exit 0, got %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "OK: audit chain intact") {
		t.Fatalf("expected intact message, got %q", out.String())
	}
}

func TestAuditVerifyIntactAfterWriteDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	tenant := uuid.New()
	t.Cleanup(func() { cleanupTenant(t, pool, tenant) })

	// Issuing a key writes one audit row + chain head for the tenant.
	var out, errb bytes.Buffer
	if code := runApikey([]string{"issue", "--tenant", tenant.String(), "--name", "audit-src"}, &out, &errb); code != 0 {
		t.Fatalf("issue exit %d: %s", code, errb.String())
	}

	out.Reset()
	errb.Reset()
	if code := runAudit([]string{"verify", "--tenant", tenant.String()}, &out, &errb); code != 0 {
		t.Fatalf("verify should be intact (exit 0), got %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "OK: audit chain intact") {
		t.Fatalf("expected intact message, got %q", out.String())
	}
}

func TestAuditVerifyDetectsTamperDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	tenant := uuid.New()
	t.Cleanup(func() { cleanupTenant(t, pool, tenant) })

	var out, errb bytes.Buffer
	if code := runApikey([]string{"issue", "--tenant", tenant.String(), "--name", "tamper-src"}, &out, &errb); code != 0 {
		t.Fatalf("issue exit %d: %s", code, errb.String())
	}

	// Mutate a hashed field of the audit row so its row_hash no longer matches.
	// (Superuser bypasses FORCE RLS, so this reaches the row directly.)
	execAdmin(t, pool, `UPDATE audit_logs SET action = action || '_tampered' WHERE tenant_id = $1`, tenant)

	out.Reset()
	errb.Reset()
	code := runAudit([]string{"verify", "--tenant", tenant.String()}, &out, &errb)
	if code != 1 {
		t.Fatalf("tampered chain should exit 1, got %d (stdout=%q)", code, out.String())
	}
	if !strings.Contains(errb.String(), "TAMPER DETECTED") {
		t.Fatalf("expected tamper message on stderr, got %q", errb.String())
	}
}
