package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/privileged"
	"github.com/qatoolist/wowapi/kernel/relationship"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationModuleContextPrivilegedWired is the GAP-006 wiring regression:
// it proves the scoped privileged services reach a module through the REAL
// module.Context the app builds during Boot (deps threaded from the kernel and
// bound to the module name) — not merely in the kernel/privileged unit tests. A
// module named "committee" grants and revokes an edge of its OWN type through
// mc.Privileged().Relationships(), and a foreign-module context is denied the
// same type — proving ownership binds to the module the context belongs to.
func TestIntegrationModuleContextPrivilegedWired(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	// Capture each module's real Context during Register.
	var committeeCtx, otherCtx module.Context
	a := app.New()
	a.Register(
		funcModule{name: "committee", reg: func(mc module.Context) error { committeeCtx = mc; return nil }},
		funcModule{name: "other", reg: func(mc module.Context) error { otherCtx = mc; return nil }},
	)
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if committeeCtx == nil || otherCtx == nil {
		t.Fatal("module contexts were not captured")
	}

	// Seed a committee-owned relationship type, a capacity, and an object resource.
	tenant := testkit.CreateTenant(t, h).ID
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('committee.seat_of','committee','capacity','resource','seat') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed rel type: %v", err)
	}
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	ctx := testkit.TenantCtx(tenant)
	rel := committeeCtx.Privileged().Relationships()

	// Grant an owned edge through the module context and confirm the checker sees it.
	spec := privileged.GrantSpec{
		RelType: "committee.seat_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1,
		Object: resource.Ref{Type: obj.Type, ID: obj.ID}, Actor: user,
	}
	id, err := rel.Grant(ctx, spec)
	if err != nil {
		t.Fatalf("Grant via module.Context: %v", err)
	}
	checker := relationship.NewChecker()
	var has bool
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var e error
		has, e = checker.Has(ctx, db,
			authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant},
			"committee.seat_of", obj, time.Now().Add(time.Minute))
		return e
	}); err != nil {
		t.Fatalf("Has: %v", err)
	}
	if !has {
		t.Fatal("edge granted through module.Context is not visible to the checker")
	}

	// Revoke it back through the same context.
	if err := rel.Revoke(ctx, id, user); err != nil {
		t.Fatalf("Revoke via module.Context: %v", err)
	}

	// The "other" module's context must NOT be able to grant a committee-owned type.
	if _, err := otherCtx.Privileged().Relationships().Grant(ctx, spec); kerr.KindOf(err) != kerr.KindForbidden {
		t.Fatalf("foreign module must be denied a committee type, got %v", err)
	}
}
