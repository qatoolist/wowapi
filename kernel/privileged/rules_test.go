package privileged_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	kaudit "github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/privileged"
	"github.com/qatoolist/wowapi/v2/kernel/rules"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// These integration tests pin, framework-side, the behaviour the product
// SECURITY DEFINER bridge policy_activate_rule_version provided (GAP-006):
// tenant binding, rule-key ownership, tenant-scope restriction, the draft→active
// transition, an atomic product gate, and one-active-per-instant arbitration.

const polRuleKey = "policy.retention.days"

// ruleReg builds a registry owning the policy rule key and its store.
func ruleStore(t *testing.T, h *testkit.DBHandle) *rules.Store {
	t.Helper()
	reg := rules.NewRegistry()
	reg.Register("policy", rules.Point{
		Key: polRuleKey, ValueSchema: json.RawMessage(`{"type":"integer"}`),
		Default: json.RawMessage(`30`), AllowedScopes: []rules.ScopeKind{rules.ScopeTenant},
	})
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}
	// Mirror the definition into rule_definitions (FK target for versions).
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO rule_definitions (key, module, value_schema, default_value, allowed_scopes, description)
		 VALUES ($1,'policy','{"type":"integer"}','30','{tenant}','days') ON CONFLICT (key) DO NOTHING`,
		polRuleKey); err != nil {
		t.Fatalf("seed rule def: %v", err)
	}
	return rules.NewStore(reg, model.UUIDv7())
}

func newRuleSvc(h *testkit.DBHandle, store *rules.Store) *privileged.Rules {
	svc := privileged.New("policy", h.PlatformTxM, store, kaudit.New(model.UUIDv7(), nil), model.UUIDv7(), privileged.Config{})
	return svc.Rules()
}

// proposeTenantDraft inserts a tenant-scope draft version as app_rt and returns its id.
func proposeTenantDraft(t *testing.T, h *testkit.DBHandle, store *rules.Store, tenant uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := testkit.TenantCtx(tenant)
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = store.Propose(ctx, db, rules.Proposal{
			Key: polRuleKey, Scope: rules.ScopeTenant, Value: json.RawMessage(`90`),
		})
		return e
	}); err != nil {
		t.Fatalf("propose: %v", err)
	}
	return id
}

func statusOf(t *testing.T, h *testkit.DBHandle, id uuid.UUID) string {
	t.Helper()
	var s string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT status FROM rule_versions WHERE id = $1`, id).Scan(&s); err != nil {
		t.Fatalf("status: %v", err)
	}
	return s
}

func TestIntegrationActivateTenantHappyPath(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	id := proposeTenantDraft(t, h, store, tenant)

	svc := newRuleSvc(h, store)
	if err := svc.ActivateTenant(testkit.TenantCtx(tenant), id, uuid.New(), privileged.ActivateOptions{}); err != nil {
		t.Fatalf("ActivateTenant: %v", err)
	}
	if got := statusOf(t, h, id); got != "active" {
		t.Fatalf("want active, got %q", got)
	}
}

func TestIntegrationActivateOwnershipPrivilegeDenied(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	id := proposeTenantDraft(t, h, store, tenant)

	// Service bound to a DIFFERENT module ("other").
	other := privileged.New("other", h.PlatformTxM, store, kaudit.New(model.UUIDv7(), nil), model.UUIDv7(), privileged.Config{})
	err := other.Rules().ActivateTenant(testkit.TenantCtx(tenant), id, uuid.New(), privileged.ActivateOptions{})
	if kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("want ownership_denied (Forbidden), got %v", err)
	}
	if got := statusOf(t, h, id); got != "draft" {
		t.Fatalf("denied activation must leave draft, got %q", got)
	}
}

func TestIntegrationActivateForeignTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tenantA := testkit.CreateTenant(t, h).ID
	tenantB := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	id := proposeTenantDraft(t, h, store, tenantA) // draft belongs to A

	svc := newRuleSvc(h, store)
	// Tenant B tries to activate A's version.
	err := svc.ActivateTenant(testkit.TenantCtx(tenantB), id, uuid.New(), privileged.ActivateOptions{})
	if kerr.KindOf(err) != kerr.KindTenantIsolation {
		t.Fatalf("cross-tenant activation must be TenantIsolation, got %v", err)
	}
	if got := statusOf(t, h, id); got != "draft" {
		t.Fatalf("denied activation must leave draft, got %q", got)
	}
}

func TestIntegrationActivatePlatformScopeIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	// A platform-scope draft (tenant_id NULL) seeded directly via the platform pool.
	id := uuid.New()
	if _, err := h.Platform.Exec(context.Background(),
		`INSERT INTO rule_versions (id, rule_key, tenant_id, scope_kind, value, effective_from, status, created_by)
		 VALUES ($1,$2,NULL,'platform','90', now(), 'draft', $3)`,
		id, polRuleKey, uuid.Nil); err != nil {
		t.Fatalf("seed platform draft: %v", err)
	}

	svc := newRuleSvc(h, store)
	err := svc.ActivateTenant(testkit.TenantCtx(tenant), id, uuid.New(), privileged.ActivateOptions{})
	if kerr.KindOf(err) != kerr.KindTenantIsolation {
		t.Fatalf("platform-scope activation via tenant API must be refused, got %v", err)
	}
}

func TestIntegrationActivateNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	svc := newRuleSvc(h, store)
	err := svc.ActivateTenant(testkit.TenantCtx(tenant), uuid.New(), uuid.New(), privileged.ActivateOptions{})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("want NotFound, got %v", err)
	}
}

func TestIntegrationActivateNoTenantFailsClosed(t *testing.T) {
	h := testkit.NewDB(t)
	store := ruleStore(t, h)
	svc := newRuleSvc(h, store)
	err := svc.ActivateTenant(context.Background(), uuid.New(), uuid.New(), privileged.ActivateOptions{})
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("want Unauthenticated, got %v", err)
	}
}

func TestIntegrationActivateGateAborts(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	id := proposeTenantDraft(t, h, store, tenant)

	svc := newRuleSvc(h, store)
	gateErr := kerr.E(kerr.KindRuleViolation, "no_citation", "no verified citation")
	err := svc.ActivateTenant(testkit.TenantCtx(tenant), id, uuid.New(), privileged.ActivateOptions{
		Gate: func(ctx context.Context, db database.TenantDB) error { return gateErr },
	})
	if !errors.Is(err, gateErr) && kerr.KindOf(err) != kerr.KindRuleViolation {
		t.Fatalf("gate must abort activation, got %v", err)
	}
	if got := statusOf(t, h, id); got != "draft" {
		t.Fatalf("gated activation must roll back, got %q", got)
	}
}

func TestIntegrationActivateDoubleActivateConflict(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	id := proposeTenantDraft(t, h, store, tenant)

	svc := newRuleSvc(h, store)
	ctx := testkit.TenantCtx(tenant)
	if err := svc.ActivateTenant(ctx, id, uuid.New(), privileged.ActivateOptions{}); err != nil {
		t.Fatalf("first activate: %v", err)
	}
	// Re-activating the now-active version is an invalid transition (conflict).
	err := svc.ActivateTenant(ctx, id, uuid.New(), privileged.ActivateOptions{})
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("re-activation must conflict, got %v", err)
	}
}

func TestIntegrationActivateSupersedesPrior(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h)
	svc := newRuleSvc(h, store)
	ctx := testkit.TenantCtx(tenant)

	id1 := proposeTenantDraft(t, h, store, tenant)
	if err := svc.ActivateTenant(ctx, id1, uuid.New(), privileged.ActivateOptions{}); err != nil {
		t.Fatalf("activate v1: %v", err)
	}
	id2 := proposeTenantDraft(t, h, store, tenant)
	if err := svc.ActivateTenant(ctx, id2, uuid.New(), privileged.ActivateOptions{}); err != nil {
		t.Fatalf("activate v2: %v", err)
	}
	if got := statusOf(t, h, id1); got != "superseded" {
		t.Fatalf("v1 must be superseded, got %q", got)
	}
	if got := statusOf(t, h, id2); got != "active" {
		t.Fatalf("v2 must be active, got %q", got)
	}
}
