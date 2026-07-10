package rules_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/testkit"
)

// moduleOfKey returns the first dotted segment of a rule key.
func moduleOfKey(key string) string {
	if i := strings.IndexByte(key, '.'); i > 0 {
		return key[:i]
	}
	return key
}

// regPoint builds a one-point registry, failing on registration error.
func regPoint(t *testing.T, key, schema, def string, scopes []rules.ScopeKind, requiresApproval bool) *rules.Registry {
	t.Helper()
	r := rules.NewRegistry()
	r.Register(moduleOfKey(key), rules.Point{
		Key:              key,
		ValueSchema:      json.RawMessage(schema),
		Default:          json.RawMessage(def),
		AllowedScopes:    scopes,
		RequiresApproval: requiresApproval,
		Description:      "cov",
	})
	if err := r.Err(); err != nil {
		t.Fatalf("register %s: %v", key, err)
	}
	return r
}

// seedDefFull mirrors a point into rule_definitions (FK target for versions).
func seedDefFull(t *testing.T, h *testkit.DBHandle, key, schema, def string) {
	t.Helper()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO rule_definitions (key, module, value_schema, default_value, description)
         VALUES ($1,$2,$3,$4,$5) ON CONFLICT (key) DO NOTHING`,
		key, moduleOfKey(key), schema, def, "cov"); err != nil {
		t.Fatal(err)
	}
}

func jsonEqual(t *testing.T, a, b json.RawMessage) bool {
	t.Helper()
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		t.Fatalf("unmarshal %s: %v", a, err)
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		t.Fatalf("unmarshal %s: %v", b, err)
	}
	return reflect.DeepEqual(av, bv)
}

// ---------- unit: registry surface ----------

func TestRegistryKeysAndPoints(t *testing.T) {
	r := rules.NewRegistry()
	r.Register("core", rules.Point{Key: "core.b.two", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`2`)})
	r.Register("core", rules.Point{Key: "core.a.one", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`1`)})
	if err := r.Err(); err != nil {
		t.Fatalf("valid registration failed: %v", err)
	}
	keys := r.Keys()
	want := []string{"core.a.one", "core.b.two"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("Keys() = %v, want sorted %v", keys, want)
	}
	pts := r.Points()
	if len(pts) != 2 {
		t.Fatalf("Points() len = %d, want 2", len(pts))
	}
	if p, ok := pts["core.a.one"]; !ok || p.Module != "core" {
		t.Fatalf("Points() missing core.a.one or wrong module: %+v", p)
	}
}

func TestRegistryErrJoinsMultiple(t *testing.T) {
	r := rules.NewRegistry()
	r.Register("core", rules.Point{Key: "bad key", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`1`)})   // malformed key
	r.Register("core", rules.Point{Key: "other.a.b", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`1`)}) // foreign module
	r.Register("core", rules.Point{Key: "core.a.b"})                                                                     // missing schema/default
	err := r.Err()
	if err == nil {
		t.Fatal("three bad registrations must produce an error")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("registration error kind = %v, want internal", errors.KindOf(err))
	}
	if !strings.Contains(err.Error(), "; ") {
		t.Fatalf("multiple errors must be joined with '; ': %q", err.Error())
	}
}

func TestRegisterMissingSchemaAndDuplicate(t *testing.T) {
	// Missing schema.
	r := rules.NewRegistry()
	r.Register("core", rules.Point{Key: "core.a.b", Default: json.RawMessage(`1`)})
	if r.Err() == nil {
		t.Fatal("missing value_schema must fail registration")
	}
	// Missing default.
	r2 := rules.NewRegistry()
	r2.Register("core", rules.Point{Key: "core.a.b", ValueSchema: json.RawMessage(`{}`)})
	if r2.Err() == nil {
		t.Fatal("missing default must fail registration")
	}
	// Duplicate.
	r3 := rules.NewRegistry()
	r3.Register("core", rules.Point{Key: "core.a.b", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`1`)})
	r3.Register("core", rules.Point{Key: "core.a.b", ValueSchema: json.RawMessage(`{}`), Default: json.RawMessage(`2`)})
	err := r3.Err()
	if err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("duplicate registration must fail with 'more than once': %v", err)
	}
}

// ---------- unit: resolved decode ----------

func TestResolvedDecodeError(t *testing.T) {
	res := rules.Resolved{Key: "k", Value: json.RawMessage(`"a string"`)}
	var n int
	if err := res.Decode(&n); err == nil {
		t.Fatal("decoding a string into an int must error")
	} else if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("decode error kind = %v, want internal", errors.KindOf(err))
	}
	// Sanity: a matching decode succeeds.
	res2 := rules.Resolved{Key: "k", Value: json.RawMessage(`5`)}
	if err := res2.Decode(&n); err != nil || n != 5 {
		t.Fatalf("valid decode failed: n=%d err=%v", n, err)
	}
}

// ---------- unit: resolve on unregistered key ----------

func TestResolveUnregisteredKey(t *testing.T) {
	r := regPoint(t, "core.retention.audit_days", `{"type":"integer"}`, `30`, nil, false)
	resolver := rules.NewResolver(r, nil)
	// db is never touched: Get fails first.
	_, err := resolver.Resolve(context.Background(), nil, "core.absent.key", uuid.Nil, time.Now())
	if err == nil || errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("resolving an unregistered key must be an internal error: %v", err)
	}
}

// ---------- unit: propose guard rails (no DB touched) ----------

func TestProposeUnregisteredAndDisallowedScope(t *testing.T) {
	// Unregistered key.
	r := regPoint(t, "core.retention.audit_days", `{"type":"integer"}`, `30`, nil, false)
	store := rules.NewStore(r, model.UUIDv7())
	if _, err := store.Propose(context.Background(), nil, rules.Proposal{
		Key: "core.absent.key", Scope: rules.ScopeTenant, Value: json.RawMessage(`1`),
	}); err == nil || errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("proposing an unregistered key must be an internal error: %v", err)
	}

	// Scope not allowed by the point.
	rp := regPoint(t, "core.retention.audit_days", `{"type":"integer"}`, `30`,
		[]rules.ScopeKind{rules.ScopePlatform}, false)
	sp := rules.NewStore(rp, model.UUIDv7())
	if _, err := sp.Propose(context.Background(), nil, rules.Proposal{
		Key: "core.retention.audit_days", Scope: rules.ScopeTenant, Value: json.RawMessage(`1`),
	}); err == nil || errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("proposing at a disallowed scope must be a validation error: %v", err)
	}
}

// ---------- unit: schema validation rejections (no DB touched) ----------

// TestSchemaValidationRejections drives validateAgainstSchema/typeMatches/compact
// through Propose's write-time check. Every case here is REJECTED before any DB
// call, so a nil db is safe and the assertions are on the returned error kind.
//
// B3: Register now validates a point's DEFAULT against its own schema (defect
// 3), so each case must supply a `def` that actually conforms to its schema —
// a single hard-coded `0` default (the pre-B3 shape of this table) would now
// fail most of these schemas at registration, before Propose is ever reached.
// The malformed-schema cases ("schema_malformed", "pattern_malformed") reject
// at Register regardless of default (a broken schema can't validate anything,
// including its own default), so those two are asserted directly against
// Registry.Err() rather than routed through Propose.
func TestSchemaValidationRejections(t *testing.T) {
	cases := []struct {
		name   string
		key    string
		schema string
		def    string
		value  string
		kind   errors.Kind
	}{
		{"integer_mismatch", "core.t.intv", `{"type":"integer"}`, `0`, `"x"`, errors.KindValidation},
		{"integer_non_whole", "core.t.intw", `{"type":"integer"}`, `0`, `1.5`, errors.KindValidation},
		{"number_mismatch", "core.t.numv", `{"type":"number"}`, `0`, `"x"`, errors.KindValidation},
		{"string_mismatch", "core.t.strv", `{"type":"string"}`, `""`, `5`, errors.KindValidation},
		{"boolean_mismatch", "core.t.boolv", `{"type":"boolean"}`, `false`, `5`, errors.KindValidation},
		{"object_mismatch", "core.t.objv", `{"type":"object"}`, `{}`, `5`, errors.KindValidation},
		{"array_mismatch", "core.t.arrv", `{"type":"array"}`, `[]`, `5`, errors.KindValidation},
		{"null_mismatch", "core.t.nullv", `{"type":"null"}`, `null`, `5`, errors.KindValidation},
		{"enum_miss", "core.t.enumv", `{"enum":["low","high"]}`, `"low"`, `"nope"`, errors.KindValidation},
		{"enum_bad_json", "core.t.enumbad", `{"enum":["low"]}`, `"low"`, `{bad`, errors.KindValidation},
		{"value_not_json", "core.t.badval", `{"type":"integer"}`, `0`, `{bad`, errors.KindValidation},

		// GAP-007: expanded RuleValueSchema keyword validation.
		{"minimum_violation", "core.t.minv", `{"type":"number","minimum":0}`, `0`, `-1`, errors.KindValidation},
		{"maximum_violation", "core.t.maxv", `{"type":"number","maximum":100}`, `0`, `101`, errors.KindValidation},
		{"exclusive_minimum_violation_eq", "core.t.exminv", `{"type":"number","exclusiveMinimum":0}`, `1`, `0`, errors.KindValidation},
		{"exclusive_minimum_violation_lt", "core.t.exminv2", `{"type":"number","exclusiveMinimum":0}`, `1`, `-1`, errors.KindValidation},
		{"exclusive_maximum_violation_eq", "core.t.exmaxv", `{"type":"number","exclusiveMaximum":100}`, `0`, `100`, errors.KindValidation},
		{"exclusive_maximum_violation_gt", "core.t.exmaxv2", `{"type":"number","exclusiveMaximum":100}`, `0`, `101`, errors.KindValidation},
		{"min_length_violation", "core.t.minlenv", `{"type":"string","minLength":3}`, `"abc"`, `"ab"`, errors.KindValidation},
		{"max_length_violation", "core.t.maxlenv", `{"type":"string","maxLength":3}`, `"abc"`, `"abcd"`, errors.KindValidation},
		{"pattern_violation", "core.t.patv", `{"type":"string","pattern":"^[a-z]+$"}`, `"abc"`, `"ABC"`, errors.KindValidation},
		{"min_items_violation", "core.t.minitemsv", `{"type":"array","minItems":2}`, `[1,2]`, `[1]`, errors.KindValidation},
		{"max_items_violation", "core.t.maxitemsv", `{"type":"array","maxItems":2}`, `[1]`, `[1,2,3]`, errors.KindValidation},
		{"required_missing", "core.t.reqv", `{"type":"object","required":["name"]}`, `{"name":"x"}`, `{"other":1}`, errors.KindValidation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := regPoint(t, c.key, c.schema, c.def, nil, false)
			store := rules.NewStore(r, model.UUIDv7())
			_, err := store.Propose(context.Background(), nil, rules.Proposal{
				Key: c.key, Scope: rules.ScopeTenant, Value: json.RawMessage(c.value),
			})
			if err == nil {
				t.Fatalf("value %s against %s must be rejected", c.value, c.schema)
			}
			if errors.KindOf(err) != c.kind {
				t.Fatalf("error kind = %v, want %v (%v)", errors.KindOf(err), c.kind, err)
			}
		})
	}
}

// TestSchemaValidationRejectionsAtRegister covers the two malformed-schema
// cases carried over from the pre-B3 TestSchemaValidationRejections table
// ("schema_malformed", "pattern_malformed"): a schema that cannot even
// type-check its OWN default now fails at Register (B3 defect 3 built on top
// of the pre-existing malformed-schema/malformed-pattern checks), so these
// are asserted directly against Registry.Err() rather than routed through
// Propose (which they can no longer reach).
func TestSchemaValidationRejectionsAtRegister(t *testing.T) {
	cases := []struct {
		name   string
		key    string
		schema string
		def    string
	}{
		{"schema_malformed", "core.t.badschema", `{bad`, `1`},
		{"pattern_malformed", "core.t.patbadv", `{"type":"string","pattern":"("}`, `"x"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := rules.NewRegistry()
			r.Register(moduleOfKey(c.key), rules.Point{
				Key: c.key, ValueSchema: json.RawMessage(c.schema), Default: json.RawMessage(c.def),
				Description: "cov",
			})
			err := r.Err()
			if err == nil {
				t.Fatalf("schema %s must fail registration", c.schema)
			}
			if errors.KindOf(err) != errors.KindInternal {
				t.Fatalf("error kind = %v, want internal (%v)", errors.KindOf(err), err)
			}
		})
	}
}

// ---------- integration: numeric bounds enforced at Propose (GAP-007) ----------

// TestIntegrationRulePropose_RejectsOutOfBoundsNumeric is the GAP-007
// headline acceptance criterion: Store.Propose rejects an out-of-bounds
// numeric value when the point's schema declares minimum/maximum, against a
// real database — this is what let wowsociety delete its product-side bounds
// check (rulepoints.go checkValue) once the framework schema enforces it.
func TestIntegrationRulePropose_RejectsOutOfBoundsNumeric(t *testing.T) {
	const key = "policy.mh.interest_rate_cap_pct"
	schema := `{"type":"number","minimum":0,"maximum":21}`
	h := testkit.NewDB(t)
	seedDefFull(t, h, key, schema, `21`)
	r := rules.NewRegistry()
	r.Register("policy", rules.Point{
		Key: key, ValueSchema: json.RawMessage(schema), Default: json.RawMessage(`21`),
		AllowedScopes: []rules.ScopeKind{rules.ScopePlatform, rules.ScopeTenant},
		Description:   "interest rate cap",
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	store := rules.NewStore(r, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	// Below minimum.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`-1`)})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("a value below minimum must be rejected at write: %v", err)
	}

	// Above maximum.
	err = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`22`)})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("a value above maximum must be rejected at write: %v", err)
	}

	// Within bounds: accepted.
	err = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`18`)})
		return e
	})
	if err != nil {
		t.Fatalf("a value within bounds must be accepted: %v", err)
	}
}

// ---------- integration: value types round-trip (accept paths) ----------

// TestIntegrationRuleValueTypesRoundTrip proposes SCHEMA-VALID values of every
// supported type, activates them, and asserts the resolver returns them
// verbatim — exercising typeMatches' accept branches, the no-type escape,
// enum membership, and compact() on the enum success path.
//
// B3: an unknown "type" keyword used to be silently accepted by typeMatches'
// default `true` branch (fail-open); that case has moved to
// TestRegisterRejectsUnknownType in the registration-rejection suite, which
// asserts the point is now REJECTED at Register — it can no longer reach
// Propose/Resolve at all.
func TestIntegrationRuleValueTypesRoundTrip(t *testing.T) {
	cases := []struct {
		key    string
		schema string
		def    string
		value  string
	}{
		{"core.rt.numv", `{"type":"number"}`, `0`, `1.5`},
		{"core.rt.strv", `{"type":"string"}`, `""`, `"hello"`},
		{"core.rt.boolv", `{"type":"boolean"}`, `false`, `true`},
		{"core.rt.arrv", `{"type":"array"}`, `[]`, `[1,2,3]`},
		{"core.rt.objv", `{"type":"object"}`, `{}`, `{"a":1}`},
		{"core.rt.nullv", `{"type":"null"}`, `null`, `null`},
		{"core.rt.anyv", `{}`, `0`, `42`},
		{"core.rt.enumv", `{"enum":["low","high"]}`, `"low"`, `"high"`},
	}

	h := testkit.NewDB(t)
	r := rules.NewRegistry()
	for _, c := range cases {
		seedDefFull(t, h, c.key, c.schema, c.def)
		r.Register("core", rules.Point{
			Key: c.key, ValueSchema: json.RawMessage(c.schema),
			Default: json.RawMessage(c.def), Description: "rt",
		})
	}
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	store := rules.NewStore(r, model.UUIDv7())
	resolver := rules.NewResolver(r, nil)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			proposeActivate(t, h, ctx, store, rules.Proposal{
				Key: c.key, Scope: rules.ScopeTenant, Value: json.RawMessage(c.value),
			})
			var res rules.Resolved
			if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
				var e error
				res, e = resolver.Resolve(ctx, db, c.key, uuid.Nil, time.Now())
				return e
			}); err != nil {
				t.Fatal(err)
			}
			if res.IsDefault {
				t.Fatalf("%s resolved to default; want the activated value", c.key)
			}
			if !jsonEqual(t, res.Value, json.RawMessage(c.value)) {
				t.Fatalf("%s resolved = %s, want %s", c.key, res.Value, c.value)
			}
		})
	}
}

// ---------- integration: feature flag as a feature.* rule point ----------

func TestIntegrationFeatureFlagRollout(t *testing.T) {
	type flag struct {
		Enabled bool `json:"enabled"`
		Rollout int  `json:"rollout"`
	}
	const key = "feature.newui.rollout"
	schema := `{"type":"object"}`
	def := `{"enabled":false,"rollout":0}`

	h := testkit.NewDB(t)
	seedDefFull(t, h, key, schema, def)
	r := rules.NewRegistry()
	r.Register("feature", rules.Point{
		Key: key, ValueSchema: json.RawMessage(schema), Default: json.RawMessage(def),
		Description: "new ui rollout flag",
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	store := rules.NewStore(r, model.UUIDv7())
	resolver := rules.NewResolver(r, nil)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	resolve := func() flag {
		t.Helper()
		var res rules.Resolved
		if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			res, e = resolver.Resolve(ctx, db, key, uuid.Nil, time.Now())
			return e
		}); err != nil {
			t.Fatal(err)
		}
		var f flag
		if err := res.Decode(&f); err != nil {
			t.Fatal(err)
		}
		return f
	}

	// Default: disabled, 0% rollout.
	if f := resolve(); f.Enabled || f.Rollout != 0 {
		t.Fatalf("default flag = %+v, want disabled/0", f)
	}
	// Roll out to 25% for the tenant.
	proposeActivate(t, h, ctx, store, rules.Proposal{
		Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`{"enabled":true,"rollout":25}`),
	})
	if f := resolve(); !f.Enabled || f.Rollout != 25 {
		t.Fatalf("rolled-out flag = %+v, want enabled/25", f)
	}
	// A schema-invalid flag value (array, not object) is rejected at write.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := store.Propose(ctx, db, rules.Proposal{
			Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`[1,2]`),
		})
		return e
	})
	if errors.KindOf(err) != errors.KindValidation {
		t.Fatalf("a non-object flag value must be rejected at write: %v", err)
	}
}

// ---------- integration: org-scope resolution precedence ----------

func TestIntegrationRuleOrgScopePrecedence(t *testing.T) {
	const key = "core.retention.audit_days"
	h := testkit.NewDB(t)
	seedRuleDef(t, h, key)
	r := reg(t, false)
	store := rules.NewStore(r, model.UUIDv7())

	// Genuine self-first ancestry walk over organizations.parent_org_id.
	ancestry := func(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
		chain := []uuid.UUID{}
		cur := orgID
		for cur != uuid.Nil {
			chain = append(chain, cur)
			var parent *uuid.UUID
			if err := db.QueryRow(ctx, `SELECT parent_org_id FROM organizations WHERE id = $1`, cur).Scan(&parent); err != nil {
				return nil, err
			}
			if parent == nil {
				break
			}
			cur = *parent
		}
		return chain, nil
	}
	resolver := rules.NewResolver(r, ancestry)

	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	parentOrg := testkit.CreateOrg(t, h, tn.ID, nil, "parent")
	childOrg := testkit.CreateOrg(t, h, tn.ID, &parentOrg, "child")

	// Platform=90, Tenant=20, ParentOrg=50.
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: key, Scope: rules.ScopePlatform, Value: json.RawMessage(`90`)})
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`20`)})
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: key, Scope: rules.ScopeOrg, ScopeID: parentOrg, Value: json.RawMessage(`50`)})

	resolveFor := func(org uuid.UUID) rules.Resolved {
		t.Helper()
		var res rules.Resolved
		if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
			var e error
			res, e = resolver.Resolve(ctx, db, key, org, time.Now())
			return e
		}); err != nil {
			t.Fatal(err)
		}
		return res
	}

	// Child has no org version → nearest ancestor with one (parent) wins over tenant.
	var got int
	res := resolveFor(childOrg)
	_ = res.Decode(&got)
	if got != 50 || res.Scope != rules.ScopeOrg {
		t.Fatalf("child should inherit parent org value: got %d scope %s", got, res.Scope)
	}

	// Give the child its own org version = 11 → nearest (self) org wins.
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: key, Scope: rules.ScopeOrg, ScopeID: childOrg, Value: json.RawMessage(`11`)})
	res = resolveFor(childOrg)
	_ = res.Decode(&got)
	if got != 11 || res.Scope != rules.ScopeOrg {
		t.Fatalf("child's own org value should win: got %d scope %s", got, res.Scope)
	}

	// An org with no version anywhere in its (empty) chain falls to tenant=20.
	orphan := testkit.CreateOrg(t, h, tn.ID, nil, "orphan")
	res = resolveFor(orphan)
	_ = res.Decode(&got)
	if got != 20 || res.Scope != rules.ScopeTenant {
		t.Fatalf("orphan org should fall through to tenant: got %d scope %s", got, res.Scope)
	}
}

// ---------- integration: ancestry error propagates ----------

func TestIntegrationRuleAncestryError(t *testing.T) {
	const key = "core.retention.audit_days"
	h := testkit.NewDB(t)
	seedRuleDef(t, h, key)
	r := reg(t, false)
	boom := errors.E(errors.KindInternal, "ancestry_boom", "ancestry lookup failed")
	ancestry := func(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
		return nil, boom
	}
	resolver := rules.NewResolver(r, ancestry)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)
	org := testkit.CreateOrg(t, h, tn.ID, nil, "o")

	err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := resolver.Resolve(ctx, db, key, org, time.Now())
		return e
	})
	if err == nil {
		t.Fatal("an ancestry failure must surface from Resolve")
	}
}

// ---------- integration: activation error branches ----------

func TestIntegrationRuleActivateErrors(t *testing.T) {
	const key = "core.retention.audit_days"
	h := testkit.NewDB(t)
	seedRuleDef(t, h, key)
	r := reg(t, false)
	store := rules.NewStore(r, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	// Activating a non-existent version → not found.
	if err := store.Activate(context.Background(), h.Platform, uuid.New(), uuid.New()); errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("activating a missing version must be not-found: %v", err)
	}

	// Propose + activate once → active.
	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`4`)})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Activate(context.Background(), h.Platform, id, uuid.New()); err != nil {
		t.Fatalf("first activate: %v", err)
	}
	// Re-activating an already-active version → conflict (invalid transition).
	if err := store.Activate(context.Background(), h.Platform, id, uuid.New()); errors.KindOf(err) != errors.KindConflict {
		t.Fatalf("re-activating an active version must be a conflict: %v", err)
	}
}

// ---------- integration: propose records the acting user ----------

func TestIntegrationRuleProposeRecordsActor(t *testing.T) {
	const key = "core.retention.audit_days"
	h := testkit.NewDB(t)
	seedRuleDef(t, h, key)
	r := reg(t, false)
	store := rules.NewStore(r, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	actor := uuid.New()
	ctx := database.WithActorID(testkit.TenantCtx(tn.ID), actor)

	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`8`)})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	var createdBy uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT created_by FROM rule_versions WHERE id = $1`, id).Scan(&createdBy); err != nil {
		t.Fatal(err)
	}
	if createdBy != actor {
		t.Fatalf("created_by = %s, want the acting user %s", createdBy, actor)
	}
}

// ---------- integration: at-most-one-active exclusion constraint ----------

// TestIntegrationRuleNoOverlapExclusion asserts the DB invariant the resolver
// relies on: two ACTIVE versions for the same (rule, scope) may not overlap in
// time. The first activation succeeds; a second, overlapping ACTIVE row inserted
// directly (bypassing Activate's supersede) is rejected by the exclusion GiST.
func TestIntegrationRuleNoOverlapExclusion(t *testing.T) {
	const key = "core.retention.audit_days"
	h := testkit.NewDB(t)
	seedRuleDef(t, h, key)
	r := reg(t, false)
	store := rules.NewStore(r, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	from := time.Now().Add(-time.Hour)
	proposeActivate(t, h, ctx, store, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`5`), EffectiveFrom: from})

	// A second open-ended ACTIVE tenant version overlapping the first must fail.
	_, err := h.Platform.Exec(context.Background(),
		`INSERT INTO rule_versions
             (id, rule_key, tenant_id, scope_kind, scope_id, value, effective_from, status, created_by)
         VALUES ($1,$2,$3,'tenant',NULL,$4,$5,'active',$6)`,
		uuid.New(), key, tn.ID, json.RawMessage(`6`), from.Add(time.Minute), uuid.Nil)
	if err == nil {
		t.Fatal("a second overlapping active version must be rejected by the exclusion constraint")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "exclu") && !strings.Contains(strings.ToLower(err.Error()), "conflict") {
		t.Fatalf("expected an exclusion-constraint violation, got: %v", err)
	}
}
