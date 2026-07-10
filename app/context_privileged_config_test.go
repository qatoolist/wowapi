package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/privileged"
	"github.com/qatoolist/wowapi/kernel/relationship"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationPrivilegedConfigAllowListWiredThroughContext is the backlog
// B10 wiring proof: a product config allow-list (config.Framework.Privileged)
// for module "committee" reaches module.Context.Privileged() through Boot, and
// lets committee manage a NON-prefixed, kernel-owned relationship type
// ("core.owner_of") it would otherwise be denied. A sibling module ("other")
// with NO config entry keeps EXACTLY today's prefix-only behavior, and is
// still denied both the kernel type and committee's own type — proving
// committee's allow-list does not leak to other modules (cross-module
// isolation).
func TestIntegrationPrivilegedConfigAllowListWiredThroughContext(t *testing.T) {
	h := testkit.NewDB(t)

	cfg := config.Defaults()
	cfg.Privileged = config.Privileged{
		"committee": config.PrivilegedGrant{AllowRelTypes: []string{"core.owner_of"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid explicit allow-list must pass boot validation: %v", err)
	}

	k, err := kernel.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}

	var committeeCtx, otherCtx module.Context
	a := app.New()
	a.Register(
		funcModule{name: "committee", reg: func(mc module.Context) error { committeeCtx = mc; return nil }},
		funcModule{name: "other", reg: func(mc module.Context) error { otherCtx = mc; return nil }},
	)
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("Boot: %v", err)
	}

	tenant := testkit.CreateTenant(t, h).ID
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('core.owner_of','core','capacity','resource','owner') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed core rel type: %v", err)
	}
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('committee.seat_of','committee','capacity','resource','seat') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed committee rel type: %v", err)
	}
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")
	ctx := testkit.TenantCtx(tenant)

	// 1. WITH config allow-list: committee manages the kernel "core.owner_of"
	// type through the privileged service — tenant-bound, still audited.
	id, err := committeeCtx.Privileged().Relationships().Grant(ctx, privileged.GrantSpec{
		RelType: "core.owner_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1,
		Object: resource.Ref{Type: obj.Type, ID: obj.ID}, Actor: user,
	})
	if err != nil {
		t.Fatalf("committee (allow-listed) Grant of core.owner_of: %v", err)
	}
	if err := committeeCtx.Privileged().Relationships().Revoke(ctx, id, user); err != nil {
		t.Fatalf("committee Revoke: %v", err)
	}

	// committee still owns its OWN prefix — allow-list widens, does not replace.
	if _, err := committeeCtx.Privileged().Relationships().Grant(ctx, privileged.GrantSpec{
		RelType: "committee.seat_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1,
		Object: resource.Ref{Type: obj.Type, ID: obj.ID}, Actor: user,
	}); err != nil {
		t.Fatalf("committee prefix-owned Grant must still work: %v", err)
	}

	// 2. WITHOUT config (module "other" has no Privileged entry): unchanged —
	// only prefix-owned keys, still denied the kernel type AND committee's type.
	if _, err := otherCtx.Privileged().Relationships().Grant(ctx, privileged.GrantSpec{
		RelType: "core.owner_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1,
		Object: resource.Ref{Type: obj.Type, ID: obj.ID}, Actor: user,
	}); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("other (no config) must be denied core.owner_of, got %v", err)
	}

	// 3. Cross-module isolation: committee's allow-list must not leak to other.
	if _, err := otherCtx.Privileged().Relationships().Grant(ctx, privileged.GrantSpec{
		RelType: "committee.seat_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1,
		Object: resource.Ref{Type: obj.Type, ID: obj.ID}, Actor: user,
	}); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("other must still be denied committee.seat_of (no cross-module leak), got %v", err)
	}
}

// TestIntegrationPrivilegedConfigNoEntryUnchanged proves a module absent from
// config.Framework.Privileged entirely gets the zero value — prefix-ownership
// only, byte-for-byte the pre-B10 behavior (no allow-list at all is wired).
func TestIntegrationPrivilegedConfigNoEntryUnchanged(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h) // config.Defaults(): zero-value Privileged

	var mc module.Context
	a := app.New()
	a.Register(funcModule{name: "committee", reg: func(m module.Context) error { mc = m; return nil }})
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("Boot: %v", err)
	}

	tenant := testkit.CreateTenant(t, h).ID
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('core.owner_of','core','capacity','resource','owner') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed core rel type: %v", err)
	}
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")
	ctx := testkit.TenantCtx(tenant)

	if _, err := mc.Privileged().Relationships().Grant(ctx, privileged.GrantSpec{
		RelType: "core.owner_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1,
		Object: resource.Ref{Type: obj.Type, ID: obj.ID}, Actor: user,
	}); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("module with no Privileged config entry must stay prefix-only, got %v", err)
	}
}

// TestPrivilegedConfigBootRejectsWildcard proves the fail-closed SEC
// requirement at the boot-validation seam actually used by product Load():
// Framework.Validate rejects a wildcard entry in Privileged before Boot ever
// runs, so a widened-but-unsafe config never reaches module wiring.
func TestPrivilegedConfigBootRejectsWildcard(t *testing.T) {
	cfg := config.Defaults()
	cfg.Privileged = config.Privileged{
		"committee": config.PrivilegedGrant{AllowRelTypes: []string{"*"}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("wildcard allow-list entry must fail Framework.Validate (fail closed)")
	}
}
