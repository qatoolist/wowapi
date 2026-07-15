package integration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/integration"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

// TestRegistryMalformedKeyAndMultiError covers the malformed-key branch of
// Register (a key that is not module.name) and the multi-error join path of
// Err (two accumulated errors must be joined with "; ").
func TestRegistryMalformedKeyAndMultiError(t *testing.T) {
	r := integration.NewRegistry()
	// "core" has no ".name" segment → fails keyRE before the module-prefix check.
	r.Register("core", &fakeProvider{key: "core", kind: "payment"})
	if r.Err() == nil {
		t.Fatal("a malformed provider key (not module.name) must fail registration")
	}

	// Two distinct invalid registrations → two accumulated errors → Err joins them.
	r2 := integration.NewRegistry()
	r2.Register("core", &fakeProvider{key: "BADKEY", kind: "payment"}) // malformed key
	r2.Register("core", &fakeProvider{key: "core.x", kind: "bogus"})   // invalid kind
	err := r2.Err()
	if err == nil {
		t.Fatal("two invalid registrations must surface an error")
	}
	if !strings.Contains(err.Error(), "; ") {
		t.Fatalf("multiple registration errors must be joined with %q, got: %v", "; ", err)
	}
}

// TestUpsertValidationBranches covers the three pre-DB validation rejections of
// Upsert: invalid key, invalid kind, and non-JSON-encodable settings. Each
// returns before touching the DB, so a nil DBTX exercises only the guard.
func TestUpsertValidationBranches(t *testing.T) {
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())
	ctx := context.Background()

	// Invalid key (not module.name).
	if _, err := store.Upsert(ctx, nil, integration.UpsertIn{Key: "notakey", Kind: "payment"}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("invalid key must be a validation error, got %v", err)
	}
	// Invalid kind (outside the closed set).
	if _, err := store.Upsert(ctx, nil, integration.UpsertIn{Key: "core.pay", Kind: "bogus"}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("invalid kind must be a validation error, got %v", err)
	}
	// Settings not JSON-encodable (a channel cannot be marshaled).
	if _, err := store.Upsert(ctx, nil, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", Settings: map[string]any{"bad": make(chan int)},
	}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("non-JSON-encodable settings must be a validation error, got %v", err)
	}
}

// TestUpsertDBError covers the Upsert DB-error wrap path: a canceled context
// makes the INSERT fail, and the store must wrap (not swallow) the error.
func TestUpsertDBError(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // poison the query before it runs
	if _, err := store.Upsert(ctx, h.Platform, integration.UpsertIn{Key: "core.pay", Kind: "payment"}); err == nil {
		t.Fatal("upsert with a canceled context must return an error")
	}
}

// TestUpsertRecordsActor covers the actorFrom(ctx) success branch: when the
// context carries an actor id, Upsert persists it as created_by.
func TestUpsertRecordsActor(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())

	actor := uuid.New()
	ctx := database.WithActorID(context.Background(), actor)
	id, err := store.Upsert(ctx, h.Platform, integration.UpsertIn{Key: "core.pay", Kind: "payment"})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var createdBy uuid.UUID
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT created_by FROM integration_providers WHERE id = $1`, id).Scan(&createdBy); err != nil {
		t.Fatal(err)
	}
	if createdBy != actor {
		t.Fatalf("created_by must equal the context actor: want %s, got %s", actor, createdBy)
	}
}

// TestResolveDBError covers the Resolve non-ErrNoRows DB-error path: a canceled
// context makes the SELECT fail with something other than ErrNoRows.
func TestResolveDBError(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	var resErr error
	_ = h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		_, resErr = store.Resolve(cctx, db, "core.pay")
		return nil
	})
	if resErr == nil {
		t.Fatal("resolve with a canceled context must return an error")
	}
	if kerr.KindOf(resErr) == kerr.KindNotFound {
		t.Fatalf("a DB failure must not masquerade as NotFound, got %v", resErr)
	}
}

// TestResolveMalformedConfigJSON covers the config json.Unmarshal error branch:
// a row whose config is a JSON array (not an object) fails to decode into the
// settings map.
func TestResolveMalformedConfigJSON(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	// Insert a platform row (tenant_id NULL, visible to the tenant via RLS) whose
	// config is a JSON array — Resolve unmarshals into map[string]any and must err.
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO integration_providers (id, tenant_id, key, kind, config, status, created_by)
		 VALUES ($1, NULL, $2, 'payment', '[1,2,3]'::jsonb, 'active', $3)`,
		model.UUIDv7().New(), "core.pay", uuid.Nil); err != nil {
		t.Fatal(err)
	}

	var resErr error
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, resErr = store.Resolve(ctx, db, "core.pay")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if resErr == nil {
		t.Fatal("a non-object config JSON must fail to decode into the settings map")
	}
}

// TestResolveNoSecretsProvider covers the branch where a row carries a
// credential_ref but no secrets provider is wired.
func TestResolveNoSecretsProvider(t *testing.T) {
	h := testkit.NewDB(t)
	// secrets provider is nil.
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	if _, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", CredentialRef: "secretref://env/PAY_KEY",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	var resErr error
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, resErr = store.Resolve(ctx, db, "core.pay")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if kerr.KindOf(resErr) != kerr.KindInternal {
		t.Fatalf("a credential_ref with no secrets provider must be an internal error, got %v", resErr)
	}
}

// TestResolveMalformedCredentialRef covers the ParseRef error branch: a ref that
// passes IsRef (correct scheme prefix) but is structurally malformed (no path)
// must fail when Resolve parses it.
func TestResolveMalformedCredentialRef(t *testing.T) {
	h := testkit.NewDB(t)
	// Non-nil secrets so we get past the nil-provider guard to ParseRef.
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), fakeSecrets{}, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	// "secretref://foo" has the scheme prefix (IsRef true) but no "/path" segment,
	// so Upsert accepts it while Resolve's ParseRef rejects it.
	if _, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", CredentialRef: "secretref://foo",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	var resErr error
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, resErr = store.Resolve(ctx, db, "core.pay")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if kerr.KindOf(resErr) != kerr.KindInternal {
		t.Fatalf("a malformed credential_ref must be an internal error, got %v", resErr)
	}
}

// TestResolveSecretResolutionError covers the branch where the secrets provider
// returns an error for a well-formed ref (the secret does not exist).
func TestResolveSecretResolutionError(t *testing.T) {
	h := testkit.NewDB(t)
	// Empty map → any lookup fails.
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), fakeSecrets{}, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	if _, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", CredentialRef: "secretref://env/MISSING",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	var resErr error
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, resErr = store.Resolve(ctx, db, "core.pay")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if resErr == nil {
		t.Fatal("a ref whose secret cannot be resolved must return an error")
	}
}

// TestHealthChecksReportsResolveError covers the HealthChecks branch where a
// configured provider fails to resolve (non-NotFound): the error must be
// reported under the provider key rather than skipped.
func TestHealthChecksReportsResolveError(t *testing.T) {
	h := testkit.NewDB(t)
	prov := &fakeProvider{key: "core.pay", kind: "payment"}
	// Empty secrets → credential resolution fails for the configured row.
	store := integration.NewStore(reg(t, prov), fakeSecrets{}, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	if _, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", CredentialRef: "secretref://env/MISSING",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	var res map[string]error
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		res = store.HealthChecks(ctx, db)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	err, present := res["core.pay"]
	if !present {
		t.Fatal("a configured provider that fails to resolve must be reported, not skipped")
	}
	if err == nil {
		t.Fatal("the reported health entry must carry the resolve error")
	}
}
