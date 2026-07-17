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

// ---------- B3 defect 1: unknown "type" must fail closed ----------

// TestRegisterRejectsUnknownType is the B3 headline regression for the
// typeMatches fail-open bug (kernel/rules/schema.go:153 used to `return true`
// for any unrecognized "type" string, so ANY value was accepted). Register
// must now reject a point whose schema names an unknown type, surfaced
// through Registry.Err() — the existing boot-error-accumulation gate
// (app/boot.go calls k.Rules.Err()) turns this into a boot failure.
func TestRegisterRejectsUnknownType(t *testing.T) {
	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:         "core.t.unknowntype",
		ValueSchema: json.RawMessage(`{"type":"weird"}`),
		Default:     json.RawMessage(`"whatever"`),
		Description: "unknown type must fail closed",
	})
	err := r.Err()
	if err == nil {
		t.Fatal("a schema with an unrecognized \"type\" must fail registration")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("registration error kind = %v, want internal", errors.KindOf(err))
	}
}

// ---------- B3 defect 2: unknown keywords must be rejected, not dropped ----------

// TestRegisterRejectsUnknownKeyword is the B3 regression for defect 2: an
// unrecognized JSON Schema keyword (e.g. "multipleOf", which this framework's
// limited grammar deliberately does not implement) used to be silently
// dropped by json.Unmarshal into the unexported schema struct, so it was
// NEVER enforced. Register must now reject it outright rather than pretend
// to honor a constraint it cannot check.
func TestRegisterRejectsUnknownKeyword(t *testing.T) {
	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:         "core.t.unknownkw",
		ValueSchema: json.RawMessage(`{"type":"number","multipleOf":5}`),
		Default:     json.RawMessage(`10`),
		Description: "unknown keyword must be rejected, not dropped",
	})
	err := r.Err()
	if err == nil {
		t.Fatal("a schema with an unrecognized keyword (multipleOf) must fail registration")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("registration error kind = %v, want internal", errors.KindOf(err))
	}
}

// TestSyncDefinitionsRejectsPointWithUnknownKeyword proves the sweep
// requirement: a schema that would newly fail Register-time validation can
// never reach SyncDefinitions (and therefore never reach rule_definitions) —
// Registry.Register already refused to add it to the registry, so
// SyncDefinitions has nothing to sync for that key.
func TestSyncDefinitionsRejectsPointWithUnknownKeyword(t *testing.T) {
	h := testkit.NewDB(t)
	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:         "core.t.syncunknownkw",
		ValueSchema: json.RawMessage(`{"type":"string","multipleOf":5}`),
		Default:     json.RawMessage(`"x"`),
		Description: "must never sync",
	})
	if r.Err() == nil {
		t.Fatal("registration must fail for an unknown keyword before SyncDefinitions is ever reached")
	}

	// Even if a caller ignored Err() and called SyncDefinitions anyway, the
	// rejected point was never added to the registry map, so Keys()/Points()
	// cannot surface it — SyncDefinitions has nothing to upsert for this key.
	if err := rules.SyncDefinitions(context.Background(), h.Platform, r); err != nil {
		t.Fatalf("SyncDefinitions on the (empty-for-this-key) registry must not itself error: %v", err)
	}
	var count int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM rule_definitions WHERE key = $1`, "core.t.syncunknownkw").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rule_definitions has %d rows for a point that failed registration, want 0", count)
	}
}

// ---------- B3 defect 3: defaults must be validated against their own schema ----------

// TestRegisterRejectsDefaultViolatingSchema is the B3 regression for defect
// 3: Registry.Register used to accept a Point whose Default violated its own
// ValueSchema (nothing checked the default against the schema at
// registration time). A registered point whose own default fails its schema
// is a broken point — it should never boot.
func TestRegisterRejectsDefaultViolatingSchema(t *testing.T) {
	cases := []struct {
		name   string
		schema string
		def    string
	}{
		{"wrong_type", `{"type":"integer"}`, `"not-an-int"`},
		{"below_minimum", `{"type":"number","minimum":0}`, `-5`},
		{"above_maximum", `{"type":"number","maximum":10}`, `100`},
		{"not_in_enum", `{"enum":["low","high"]}`, `"medium"`},
		{"too_short", `{"type":"string","minLength":3}`, `"ab"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := rules.NewRegistry()
			r.Register("core", rules.Point{
				Key:         "core.t.baddefault",
				ValueSchema: json.RawMessage(c.schema),
				Default:     json.RawMessage(c.def),
				Description: "default violates its own schema",
			})
			err := r.Err()
			if err == nil {
				t.Fatalf("a default %s violating schema %s must fail registration", c.def, c.schema)
			}
			if errors.KindOf(err) != errors.KindInternal {
				t.Fatalf("registration error kind = %v, want internal", errors.KindOf(err))
			}
		})
	}
}

// TestRegisterAcceptsDefaultConformingToSchema is the accept-path sibling of
// TestRegisterRejectsDefaultViolatingSchema: a well-formed schema whose
// default satisfies it registers cleanly.
func TestRegisterAcceptsDefaultConformingToSchema(t *testing.T) {
	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key:         "core.t.gooddefault",
		ValueSchema: json.RawMessage(`{"type":"number","minimum":0,"maximum":100}`),
		Default:     json.RawMessage(`50`),
		Description: "default conforms to its schema",
	})
	if err := r.Err(); err != nil {
		t.Fatalf("a conforming default must register cleanly: %v", err)
	}
}

// ---------- B3 defect 4: resolve-path validation (defense in depth) ----------

// TestIntegrationResolveRejectsStoredValueViolatingCurrentSchema proves the
// honest contract for Resolver.Resolve: it validates the resolved value
// against the POINT'S CURRENT schema before returning it. This matters when
// a schema is tightened after a value was written (schema evolution) — a
// previously-valid stored value can become invalid for the current schema,
// and Resolve must not silently hand back a value that violates the
// registered contract it is now serving under.
func TestIntegrationResolveRejectsStoredValueViolatingCurrentSchema(t *testing.T) {
	const key = "core.t.driftedschema"
	h := testkit.NewDB(t)

	// Seed rule_definitions + write a version while the schema still allows
	// values up to 100 (mirrors real schema-evolution: the DB has historical
	// rows written under an earlier, looser schema).
	looseSchema := `{"type":"number","minimum":0,"maximum":100}`
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO rule_definitions (key, module, value_schema, default_value, description)
         VALUES ($1,$2,$3,$4,$5) ON CONFLICT (key) DO NOTHING`,
		key, "core", looseSchema, `50`, "drift test"); err != nil {
		t.Fatal(err)
	}
	looseReg := rules.NewRegistry()
	looseReg.Register("core", rules.Point{
		Key: key, ValueSchema: json.RawMessage(looseSchema), Default: json.RawMessage(`50`),
		Description: "drift test",
	})
	if err := looseReg.Err(); err != nil {
		t.Fatal(err)
	}
	store := rules.NewStore(looseReg, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`90`)})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Activate(context.Background(), h.Platform, id, uuid.New()); err != nil {
		t.Fatal(err)
	}

	// Now the point's schema is tightened in-process (module upgrade) to
	// maximum:80 — the stored value 90 no longer conforms. Resolve must
	// surface this rather than silently returning the now-invalid 90.
	tightSchema := `{"type":"number","minimum":0,"maximum":80}`
	tightReg := rules.NewRegistry()
	tightReg.Register("core", rules.Point{
		Key: key, ValueSchema: json.RawMessage(tightSchema), Default: json.RawMessage(`50`),
		Description: "drift test (tightened)",
	})
	if err := tightReg.Err(); err != nil {
		t.Fatal(err)
	}
	resolver := rules.NewResolver(tightReg, nil)

	err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, e := resolver.Resolve(ctx, db, key, uuid.Nil, time.Now())
		return e
	})
	if err == nil {
		t.Fatal("resolving a stored value that violates the CURRENT schema must error, not silently return it")
	}
	if errors.KindOf(err) != errors.KindInternal {
		t.Fatalf("resolve-time schema-drift error kind = %v, want internal", errors.KindOf(err))
	}
}

// TestIntegrationResolveAcceptsConformingStoredValue is the accept-path
// sibling: a stored value that still conforms to the point's current schema
// resolves normally — the resolve-path validation added for B3 must not
// break the ordinary case.
func TestIntegrationResolveAcceptsConformingStoredValue(t *testing.T) {
	const key = "core.t.driftok"
	h := testkit.NewDB(t)
	schema := `{"type":"number","minimum":0,"maximum":100}`
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO rule_definitions (key, module, value_schema, default_value, description)
         VALUES ($1,$2,$3,$4,$5) ON CONFLICT (key) DO NOTHING`,
		key, "core", schema, `50`, "ok"); err != nil {
		t.Fatal(err)
	}
	r := rules.NewRegistry()
	r.Register("core", rules.Point{
		Key: key, ValueSchema: json.RawMessage(schema), Default: json.RawMessage(`50`), Description: "ok",
	})
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	store := rules.NewStore(r, model.UUIDv7())
	resolver := rules.NewResolver(r, nil)
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	var id uuid.UUID
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		id, e = store.Propose(ctx, db, rules.Proposal{Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`42`)})
		return e
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Activate(context.Background(), h.Platform, id, uuid.New()); err != nil {
		t.Fatal(err)
	}

	var res rules.Resolved
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		res, e = resolver.Resolve(ctx, db, key, uuid.Nil, time.Now())
		return e
	}); err != nil {
		t.Fatalf("resolving a conforming stored value must succeed: %v", err)
	}
	var got int
	_ = res.Decode(&got)
	if got != 42 {
		t.Fatalf("resolved value = %d, want 42", got)
	}
}
