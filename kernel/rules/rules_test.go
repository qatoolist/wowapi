package rules_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/testkit"
)

func reg(t *testing.T, requiresApproval bool) *rules.Registry {
	t.Helper()
	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:              "core.retention.audit_days",
		ValueSchema:      json.RawMessage(`{"type":"integer"}`),
		Default:          json.RawMessage(`30`),
		RequiresApproval: requiresApproval,
		Description:      "audit retention days",
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	return r
}

// seedRuleDef mirrors the point into rule_definitions (FK for versions is not
// required, but the resolver reads versions only).
func seedRuleDef(t *testing.T, h *testkit.DBHandle, key string) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO rule_definitions (key, module, value_schema, default_value, description)
         VALUES ($1,$2,$3,$4,$5) ON CONFLICT (key) DO NOTHING`,
		key, "core", `{"type":"integer"}`, `30`, "audit retention"); err != nil {
		t.Fatal(err)
	}
}

// proposeActivate drafts a version as app_rt then activates it as app_platform
// (rule activation is platform-gated, SEC-13).
func proposeActivate(t *testing.T, h *testkit.DBHandle, ctx context.Context, store *rules.Store, p rules.Proposal) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = store.Propose(ctx, db, p)
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Activate(context.Background(), h.Platform, id, uuid.New()); err != nil {
		t.Fatalf("activate: %v", err)
	}
	return id
}

// ---------- unit ----------

func TestRegistryValidatesKeys(t *testing.T) {
	r := rules.NewRegistry()
	r.Register("core", rules.Point{Key: "core.x", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`1`)})
	if r.Err() == nil {
		t.Fatal("malformed key must fail registration")
	}
	r2 := rules.NewRegistry()
	r2.Register("core", rules.Point{Key: "other.a.b", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`1`)})
	if r2.Err() == nil {
		t.Fatal("foreign-module key must fail registration")
	}
}

// ---------- integration ----------

func TestIntegrationRuleResolutionPrecedence(t *testing.T) {
	h := testkit.NewDB(t)
	seedRuleDef(t, h, "core.retention.audit_days")
	r := reg(t, false)
	store := rules.NewStore(r, model.UUIDv7())
	resolver := rules.NewResolver(r, nil) // no org ancestry for this test
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	// No versions → code default (30).
	var got int
	must := func() rules.Resolved {
		t.Helper()
		var res rules.Resolved
		if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			res, e = resolver.Resolve(ctx, db, "core.retention.audit_days", uuid.Nil, time.Now())
			return e
		}); err != nil {
			t.Fatal(err)
		}
		return res
	}
	res := must()
	if !res.IsDefault {
		t.Fatal("no versions should resolve to the code default")
	}
	_ = res.Decode(&got)
	if got != 30 {
		t.Fatalf("default = %d, want 30", got)
	}

	// Platform version = 90.
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: "core.retention.audit_days", Scope: rules.ScopePlatform, Value: json.RawMessage(`90`)})
	res = must()
	_ = res.Decode(&got)
	if got != 90 || res.Scope != rules.ScopePlatform {
		t.Fatalf("platform version should win: got %d scope %s", got, res.Scope)
	}

	// Tenant version = 7 overrides platform.
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: "core.retention.audit_days", Scope: rules.ScopeTenant, Value: json.RawMessage(`7`)})
	res = must()
	_ = res.Decode(&got)
	if got != 7 || res.Scope != rules.ScopeTenant {
		t.Fatalf("tenant version should override platform: got %d scope %s", got, res.Scope)
	}
}

func TestIntegrationRuleHistoricalResolution(t *testing.T) {
	h := testkit.NewDB(t)
	seedRuleDef(t, h, "core.retention.audit_days")
	r := reg(t, false)
	store := rules.NewStore(r, model.UUIDv7())
	resolver := rules.NewResolver(r, nil)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	past := time.Now().Add(-48 * time.Hour)
	// A tenant version effective from 24h ago = 5.
	from := time.Now().Add(-24 * time.Hour)
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: "core.retention.audit_days", Scope: rules.ScopeTenant, Value: json.RawMessage(`5`), EffectiveFrom: from})

	resolve := func(at time.Time) rules.Resolved {
		t.Helper()
		var res rules.Resolved
		if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			res, e = resolver.Resolve(ctx, db, "core.retention.audit_days", uuid.Nil, at)
			return e
		}); err != nil {
			t.Fatal(err)
		}
		return res
	}
	// At `past` (before the version's effective_from) → code default.
	if r := resolve(past); !r.IsDefault {
		t.Fatalf("historical resolution before effective_from must be the default, got %+v", r)
	}
	// Now → the version.
	var got int
	nowRes := resolve(time.Now())
	_ = nowRes.Decode(&got)
	if got != 5 {
		t.Fatalf("current resolution = %d, want 5", got)
	}
}

// TestIntegrationRuleHistoricalSupersededWindow is the ARCH-60 regression: a
// value that was active in the past and later superseded must still resolve for
// an `at` INSIDE its old window — not fall through to the default.
func TestIntegrationRuleHistoricalSupersededWindow(t *testing.T) {
	h := testkit.NewDB(t)
	seedRuleDef(t, h, "core.retention.audit_days")
	r := reg(t, false)
	store := rules.NewStore(r, model.UUIDv7())
	resolver := rules.NewResolver(r, nil)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	// v1 = 5 effective 10 days ago; v2 = 9 effective 2 days ago (supersedes v1).
	proposeActivate(t, h, ctx, store, rules.Proposal{
		Key: "core.retention.audit_days", Scope: rules.ScopeTenant,
		Value: json.RawMessage(`5`), EffectiveFrom: time.Now().Add(-10 * 24 * time.Hour),
	})
	proposeActivate(t, h, ctx, store, rules.Proposal{
		Key: "core.retention.audit_days", Scope: rules.ScopeTenant,
		Value: json.RawMessage(`9`), EffectiveFrom: time.Now().Add(-2 * 24 * time.Hour),
	})

	resolve := func(at time.Time) int {
		t.Helper()
		var res rules.Resolved
		if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			res, e = resolver.Resolve(ctx, db, "core.retention.audit_days", uuid.Nil, at)
			return e
		}); err != nil {
			t.Fatal(err)
		}
		var n int
		_ = res.Decode(&n)
		return n
	}
	if got := resolve(time.Now().Add(-5 * 24 * time.Hour)); got != 5 {
		t.Fatalf("resolution in a superseded window = %d, want 5 (ARCH-60)", got)
	}
	if got := resolve(time.Now()); got != 9 {
		t.Fatalf("current resolution = %d, want 9", got)
	}
}

// TestIntegrationRuleSchemaValidationAtWrite is the SEC-40 regression: a value
// that violates the point's schema is rejected at Propose time, not read time.
func TestIntegrationRuleSchemaValidationAtWrite(t *testing.T) {
	h := testkit.NewDB(t)
	seedRuleDef(t, h, "core.retention.audit_days")
	r := reg(t, false) // integer schema
	store := rules.NewStore(r, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := store.Propose(ctx, db, rules.Proposal{
			Key: "core.retention.audit_days", Scope: rules.ScopeTenant, Value: json.RawMessage(`"not-an-int"`),
		})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("a string value for an integer point must be rejected at write: %v", err)
	}
}

func TestIntegrationRuleApprovalGating(t *testing.T) {
	h := testkit.NewDB(t)
	seedRuleDef(t, h, "core.retention.audit_days")
	r := reg(t, true) // requires approval
	store := rules.NewStore(r, model.UUIDv7())
	resolver := rules.NewResolver(r, nil)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	var versionID uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		versionID, e = store.Propose(ctx, db, rules.Proposal{Key: "core.retention.audit_days", Scope: rules.ScopeTenant, Value: json.RawMessage(`3`)})
		return e
	}); err != nil {
		t.Fatal(err)
	}

	resolve := func() rules.Resolved {
		t.Helper()
		var res rules.Resolved
		_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			res, e = resolver.Resolve(ctx, db, "core.retention.audit_days", uuid.Nil, time.Now())
			return e
		})
		return res
	}
	// Before approval: the draft does NOT resolve — still the default.
	if r := resolve(); !r.IsDefault {
		t.Fatalf("an unapproved draft must not resolve, got %+v", r)
	}
	// Activate (platform privilege) and re-resolve → the value.
	if err := store.Activate(context.Background(), h.Platform, versionID, uuid.New()); err != nil {
		t.Fatalf("activate: %v", err)
	}
	var got int
	_ = resolve().Decode(&got)
	if got != 3 {
		t.Fatalf("after approval the value should resolve: got %d, want 3", got)
	}
}
