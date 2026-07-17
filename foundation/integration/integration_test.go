package integration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/foundation/integration"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/secrets"
	"github.com/qatoolist/wowapi/v2/testkit"
)

type fakeProvider struct {
	key, kind string
	health    error
	gotCred   string
}

func (f *fakeProvider) Key() string  { return f.key }
func (f *fakeProvider) Kind() string { return f.kind }
func (f *fakeProvider) HealthCheck(_ context.Context, cfg integration.Config) error {
	f.gotCred = cfg.Credential.Reveal()
	return f.health
}

type fakeSecrets map[string]string

func (f fakeSecrets) Resolve(_ context.Context, ref secrets.Ref) (string, error) {
	v, ok := f[ref.String()]
	if !ok {
		return "", fmt.Errorf("secret not found: %s", ref)
	}
	return v, nil
}

func reg(t *testing.T, p integration.Provider) *integration.Registry {
	t.Helper()
	r := integration.NewRegistry()
	r.Register("core", p)
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRegistryValidation(t *testing.T) {
	r := integration.NewRegistry()
	r.Register("core", &fakeProvider{key: "core.pay", kind: "payment"})
	if err := r.Err(); err != nil {
		t.Fatalf("valid provider rejected: %v", err)
	}
	// foreign module
	r2 := integration.NewRegistry()
	r2.Register("core", &fakeProvider{key: "other.pay", kind: "payment"})
	if r2.Err() == nil {
		t.Fatal("foreign-module key must fail")
	}
	// bad kind
	r3 := integration.NewRegistry()
	r3.Register("core", &fakeProvider{key: "core.x", kind: "bogus"})
	if r3.Err() == nil {
		t.Fatal("invalid kind must fail")
	}
	// duplicate
	r4 := integration.NewRegistry()
	r4.Register("core", &fakeProvider{key: "core.pay", kind: "payment"})
	r4.Register("core", &fakeProvider{key: "core.pay", kind: "payment"})
	if r4.Err() == nil {
		t.Fatal("duplicate must fail")
	}
}

func TestIntegrationResolvePlatformAndOverride(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	// Platform-wide provider row (tenant_id NULL), written on the unbound platform pool.
	if _, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", Settings: map[string]any{"mode": "platform"},
	}); err != nil {
		t.Fatalf("platform upsert: %v", err)
	}

	resolve := func() integration.Config {
		t.Helper()
		var cfg integration.Config
		if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
			var e error
			cfg, e = store.Resolve(ctx, db, "core.pay")
			return e
		}); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		return cfg
	}
	if cfg := resolve(); !cfg.IsPlatform || cfg.Settings["mode"] != "platform" {
		t.Fatalf("platform resolution wrong: %+v", cfg)
	}

	// A tenant override wins (written on the tenant-bound platform manager).
	if err := h.PlatformTxM.WithTenant(database.WithTenantID(context.Background(), tn.ID),
		func(ctx context.Context, db database.TenantDB) error {
			_, e := store.Upsert(ctx, db, integration.UpsertIn{
				TenantID: tn.ID, Key: "core.pay", Kind: "payment", Settings: map[string]any{"mode": "tenant"},
			})
			return e
		}); err != nil {
		t.Fatalf("tenant upsert: %v", err)
	}
	if cfg := resolve(); cfg.IsPlatform || cfg.Settings["mode"] != "tenant" {
		t.Fatalf("tenant override should win: %+v", cfg)
	}
}

func TestIntegrationCredentialResolved(t *testing.T) {
	h := testkit.NewDB(t)
	sec := fakeSecrets{"secretref://env/PAY_KEY": "sk_live_xyz"}
	prov := &fakeProvider{key: "core.pay", kind: "payment"}
	store := integration.NewStore(reg(t, prov), sec, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	if _, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", CredentialRef: "secretref://env/PAY_KEY",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var cfg integration.Config
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		var e error
		cfg, e = store.Resolve(ctx, db, "core.pay")
		return e
	}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if cfg.Credential.Reveal() != "sk_live_xyz" {
		t.Fatalf("credential not resolved from secret ref: %q", cfg.Credential.Reveal())
	}
	// S4/CA-14: the credential is a config.Secret, so it must be redacted in any
	// string/format rendering — the plaintext must never leak into logs or dumps.
	if rendered := fmt.Sprintf("%v %s %#v", cfg.Credential, cfg.Credential, cfg.Credential); strings.Contains(rendered, "sk_live_xyz") {
		t.Fatalf("credential leaked in formatted output: %q", rendered)
	}
}

func TestIntegrationUpsertRejectsPlaintextCredential(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())
	_, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{
		Key: "core.pay", Kind: "payment", CredentialRef: "sk_live_plaintext",
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("plaintext credential must be rejected, got %v", err)
	}
}

func TestIntegrationHealthChecks(t *testing.T) {
	h := testkit.NewDB(t)
	prov := &fakeProvider{key: "core.pay", kind: "payment"}
	store := integration.NewStore(reg(t, prov), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	// Not configured → HealthChecks skips it (no entry).
	var res map[string]error
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		res = store.HealthChecks(ctx, db)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, present := res["core.pay"]; present {
		t.Fatal("an unconfigured provider must be skipped by HealthChecks")
	}

	// Configure it → HealthChecks probes it (healthy).
	if _, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{Key: "core.pay", Kind: "payment"}); err != nil {
		t.Fatal(err)
	}
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		res = store.HealthChecks(ctx, db)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err, present := res["core.pay"]; !present || err != nil {
		t.Fatalf("configured healthy provider should report nil, got present=%v err=%v", present, err)
	}
}

// TestIntegrationUpsertReturnsPersistedID is the ARCH-71 regression: the second
// (conflict → DO UPDATE) upsert must return the id of the row that actually
// exists, not the freshly-generated, discarded one.
func TestIntegrationUpsertReturnsPersistedID(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())

	id1, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{Key: "core.pay", Kind: "payment"})
	if err != nil {
		t.Fatal(err)
	}
	id2, err := store.Upsert(context.Background(), h.Platform, integration.UpsertIn{Key: "core.pay", Kind: "payment", Settings: map[string]any{"v": 2}})
	if err != nil {
		t.Fatal(err)
	}
	if id2 != id1 {
		t.Fatalf("conflict upsert must return the existing id %s, got %s", id1, id2)
	}
	// The returned id must actually exist.
	var exists bool
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM integration_providers WHERE id = $1)`, id2).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("upsert returned a non-existent id %s", id2)
	}
}

func TestIntegrationResolveNotConfigured(t *testing.T) {
	h := testkit.NewDB(t)
	store := integration.NewStore(reg(t, &fakeProvider{key: "core.pay", kind: "payment"}), nil, model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	err := h.TxM.WithTenantRO(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, e := store.Resolve(ctx, db, "core.pay")
		return e
	})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("unconfigured provider must resolve to NotFound, got %v", err)
	}
}
