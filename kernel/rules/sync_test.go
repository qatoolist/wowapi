package rules_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/testkit"
)

// syncedDef is one row read back from rule_definitions for assertions.
type syncedDef struct {
	module           string
	schema           string
	defaultValue     string
	allowedScopes    []string
	requiresApproval bool
	description      string
}

func readDef(t *testing.T, h *testkit.DBHandle, key string) (syncedDef, bool) {
	t.Helper()
	var d syncedDef
	err := h.Admin.QueryRow(context.Background(),
		`SELECT module, value_schema, default_value, allowed_scopes, requires_approval, description
		   FROM rule_definitions WHERE key = $1`, key).
		Scan(&d.module, &d.schema, &d.defaultValue, &d.allowedScopes, &d.requiresApproval, &d.description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return syncedDef{}, false
		}
		t.Fatalf("query rule_definitions: %v", err)
	}
	return d, true
}

// TestSyncDefinitionsCreatesRow is the GAP-007 core proof: registering a rule
// point in Go and calling SyncDefinitions creates the rule_definitions mirror
// row — the FK rule_versions.rule_key depends on without any product SQL.
func TestSyncDefinitionsCreatesRow(t *testing.T) {
	h := testkit.NewDB(t)
	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:              "core.retention.audit_days",
		ValueSchema:      json.RawMessage(`{"type":"integer","minimum":1,"maximum":3650}`),
		Default:          json.RawMessage(`30`),
		AllowedScopes:    []rules.ScopeKind{rules.ScopePlatform, rules.ScopeTenant},
		RequiresApproval: true,
		Description:      "audit retention days",
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	if err := rules.SyncDefinitions(context.Background(), h.Platform, r); err != nil {
		t.Fatalf("SyncDefinitions: %v", err)
	}

	d, ok := readDef(t, h, "core.retention.audit_days")
	if !ok {
		t.Fatal("rule_definitions row was not created")
	}
	if d.module != "core" {
		t.Fatalf("module = %q, want core", d.module)
	}
	if d.description != "audit retention days" {
		t.Fatalf("description = %q", d.description)
	}
	if !d.requiresApproval {
		t.Fatal("requires_approval should be true")
	}
	wantScopes := []string{"platform", "tenant"}
	if len(d.allowedScopes) != len(wantScopes) || d.allowedScopes[0] != wantScopes[0] || d.allowedScopes[1] != wantScopes[1] {
		t.Fatalf("allowed_scopes = %v, want %v", d.allowedScopes, wantScopes)
	}
}

// TestSyncDefinitionsUpdatesOnFieldChange is a GAP-007 acceptance criterion:
// changing a registered point's schema/default/scopes/approval/description and
// re-syncing must update the existing row in place (no duplicate, converges).
func TestSyncDefinitionsUpdatesOnFieldChange(t *testing.T) {
	h := testkit.NewDB(t)
	const key = "core.retention.audit_days"

	r1 := rules.NewRegistry()
	r1.Register("core", rules.Point{
		Key:              key,
		ValueSchema:      json.RawMessage(`{"type":"integer"}`),
		Default:          json.RawMessage(`30`),
		AllowedScopes:    []rules.ScopeKind{rules.ScopePlatform},
		RequiresApproval: false,
		Description:      "v1 description",
	})
	if err := r1.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rules.SyncDefinitions(context.Background(), h.Platform, r1); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	r2 := rules.NewRegistry()
	r2.Register("core", rules.Point{
		Key:              key,
		ValueSchema:      json.RawMessage(`{"type":"integer","minimum":1,"maximum":365}`),
		Default:          json.RawMessage(`90`),
		AllowedScopes:    []rules.ScopeKind{rules.ScopePlatform, rules.ScopeTenant, rules.ScopeOrg},
		RequiresApproval: true,
		Description:      "v2 description",
	})
	if err := r2.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rules.SyncDefinitions(context.Background(), h.Platform, r2); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	d, ok := readDef(t, h, key)
	if !ok {
		t.Fatal("row missing after update")
	}
	if d.description != "v2 description" {
		t.Fatalf("description = %q, want updated v2 description", d.description)
	}
	if !d.requiresApproval {
		t.Fatal("requires_approval should have converged to true")
	}
	if d.defaultValue != "90" {
		t.Fatalf("default_value = %q, want 90", d.defaultValue)
	}
	wantScopes := []string{"platform", "tenant", "org"}
	if len(d.allowedScopes) != len(wantScopes) {
		t.Fatalf("allowed_scopes = %v, want %v", d.allowedScopes, wantScopes)
	}

	var count int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM rule_definitions WHERE key = $1`, key).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rule_definitions has %d rows for %s, want 1 (no duplicate)", count, key)
	}
}

// TestSyncDefinitionsIdempotentDoubleSync is the GAP-007 idempotency
// acceptance criterion: syncing the SAME registry twice converges with no
// error and no duplicate rows.
func TestSyncDefinitionsIdempotentDoubleSync(t *testing.T) {
	h := testkit.NewDB(t)
	const key = "core.retention.audit_days"

	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:         key,
		ValueSchema: json.RawMessage(`{"type":"integer"}`),
		Default:     json.RawMessage(`30`),
		Description: "audit retention days",
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := rules.SyncDefinitions(context.Background(), h.Platform, r); err != nil {
			t.Fatalf("sync run %d: %v", i, err)
		}
	}

	var count int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM rule_definitions WHERE key = $1`, key).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rule_definitions has %d rows after double sync, want 1", count)
	}
}

// TestSyncDefinitionsNoDriftFromRegistry_MultiKeyPack is the framework-level
// test that subsumes wowsociety's product drift guard
// (internal/modules/policy/rulemirror_test.go
// TestRuleDefinitionsMirror_MatchesRegistry): that test asserted the SQL
// mirror a product hand-wrote in a migration never drifted from its Go
// registry declarations, across every field (schema, default, allowed_scopes,
// requires_approval, description) AND that the count of mirrored keys
// matched the registry exactly (no missing/extra rows). With SyncDefinitions,
// there is no hand-written SQL mirror to drift — the same assertions hold
// because the DB rows ARE derived from the registry on every sync. This test
// registers a small multi-point "pack" (mirroring the shape of wowsociety's
// MH pack: several points, mixed scopes, mixed approval, one enum-typed
// point) and proves every field converges with no missing/extra rows.
func TestSyncDefinitionsNoDriftFromRegistry_MultiKeyPack(t *testing.T) {
	h := testkit.NewDB(t)
	type decl struct {
		key              string
		schema           string
		def              string
		scopes           []rules.ScopeKind
		requiresApproval bool
		description      string
	}
	pack := []decl{
		{
			"policy.mh.interest_rate_cap_pct", `{"type":"number","minimum":0,"maximum":21}`, `21`,
			[]rules.ScopeKind{rules.ScopePlatform, rules.ScopeTenant},
			true,
			"Maximum simple interest rate (% p.a.) a society may charge on arrears",
		},
		{
			"policy.mh.noc_max_pct_of_service_charges", `{"type":"number","minimum":0,"maximum":10}`, `10`,
			[]rules.ScopeKind{rules.ScopePlatform, rules.ScopeTenant},
			true,
			"Maximum non-occupancy charges as % of service charges",
		},
		{
			"policy.mh.gst_registration_threshold_rs", `{"type":"number","minimum":0}`, `2000000`,
			[]rules.ScopeKind{rules.ScopePlatform},
			false,
			"GST registration threshold (Rs aggregate annual turnover)",
		},
		{
			"policy.mh.gst_interpretation_mode", `{"type":"string","enum":["entire","excess"]}`, `"entire"`,
			[]rules.ScopeKind{rules.ScopePlatform, rules.ScopeTenant},
			true,
			"GST interpretation once charges exceed the exemption",
		},
	}

	r := rules.NewRegistry()
	for _, d := range pack {
		r.Register("policy", rules.Point{
			Key: d.key, ValueSchema: json.RawMessage(d.schema), Default: json.RawMessage(d.def),
			AllowedScopes: d.scopes, RequiresApproval: d.requiresApproval, Description: d.description,
		})
	}
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rules.SyncDefinitions(context.Background(), h.Platform, r); err != nil {
		t.Fatalf("SyncDefinitions: %v", err)
	}

	// Count: no missing/extra rows for this module (the drift guard's headline check).
	var count int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM rule_definitions WHERE module = 'policy'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != len(pack) {
		t.Fatalf("rule_definitions mirrors %d policy keys, registry declares %d", count, len(pack))
	}

	// Per-field convergence for every declared point.
	for _, d := range pack {
		got, ok := readDef(t, h, d.key)
		if !ok {
			t.Fatalf("%s: declared in registry but missing from rule_definitions", d.key)
		}
		var wantSchema, gotSchema any
		if err := json.Unmarshal([]byte(d.schema), &wantSchema); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(got.schema), &gotSchema); err != nil {
			t.Fatal(err)
		}
		if !jsonDeepEqual(wantSchema, gotSchema) {
			t.Errorf("%s: value_schema mirror %s != registry schema %s", d.key, got.schema, d.schema)
		}
		if !jsonRawEqual(got.defaultValue, d.def) {
			t.Errorf("%s: default_value mirror %s != registry default %s", d.key, got.defaultValue, d.def)
		}
		wantScopes := make([]string, len(d.scopes))
		for i, s := range d.scopes {
			wantScopes[i] = string(s)
		}
		if len(got.allowedScopes) != len(wantScopes) {
			t.Errorf("%s: allowed_scopes mirror %v != registry scopes %v", d.key, got.allowedScopes, wantScopes)
		} else {
			for i := range wantScopes {
				if got.allowedScopes[i] != wantScopes[i] {
					t.Errorf("%s: allowed_scopes mirror %v != registry scopes %v", d.key, got.allowedScopes, wantScopes)
					break
				}
			}
		}
		if got.requiresApproval != d.requiresApproval {
			t.Errorf("%s: requires_approval mirror %v != registry %v", d.key, got.requiresApproval, d.requiresApproval)
		}
		if got.description != d.description {
			t.Errorf("%s: description mirror %q != registry %q", d.key, got.description, d.description)
		}
	}
}

func jsonDeepEqual(a, b any) bool {
	ab, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(ab) == string(bb)
}

func jsonRawEqual(a, b string) bool {
	var av, bv any
	if json.Unmarshal([]byte(a), &av) != nil || json.Unmarshal([]byte(b), &bv) != nil {
		return false
	}
	return jsonDeepEqual(av, bv)
}

// TestSyncDefinitionsSatisfiesRuleVersionsFK proves the headline GAP-007
// acceptance criterion end-to-end: with ONLY SyncDefinitions run (no product
// SQL mirror, no manual rule_definitions insert), Propose+Activate succeed
// because the rule_versions.rule_key FK is satisfied purely by the framework
// lifecycle sync.
func TestSyncDefinitionsSatisfiesRuleVersionsFK(t *testing.T) {
	h := testkit.NewDB(t)
	const key = "core.retention.audit_days"

	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:         key,
		ValueSchema: json.RawMessage(`{"type":"integer"}`),
		Default:     json.RawMessage(`30`),
		Description: "audit retention days",
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rules.SyncDefinitions(context.Background(), h.Platform, r); err != nil {
		t.Fatalf("SyncDefinitions: %v", err)
	}

	store := rules.NewStore(r, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`45`)})
		return e
	}); err != nil {
		t.Fatalf("propose failed — rule_definitions FK not satisfied by SyncDefinitions alone: %v", err)
	}
	if err := store.Activate(context.Background(), h.Platform, id, uuid.New()); err != nil {
		t.Fatalf("activate: %v", err)
	}
}
