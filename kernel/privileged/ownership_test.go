package privileged_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	kaudit "github.com/qatoolist/wowapi/v2/kernel/audit"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/privileged"
	"github.com/qatoolist/wowapi/v2/kernel/relationship"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationGrantAllowListRelType proves a module can manage a NON-prefixed
// relationship type it has been explicitly allow-listed for (Config.AllowRelTypes)
// — e.g. a "core." kernel type a product module is sanctioned to grant — while
// prefix ownership still covers its own types.
func TestIntegrationGrantAllowListRelType(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('core.owner_of','core','capacity','resource','owner') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	user := testkit.CreateUser(t, h)
	cap1 := testkit.CreateCapacity(t, h, tenant, user)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")

	// "committee" module allow-listed for the kernel "core.owner_of" type.
	svc := privileged.New("committee", h.PlatformTxM, nil,
		kaudit.New(model.UUIDv7(), nil), model.UUIDv7(),
		privileged.Config{AllowRelTypes: []string{"core.owner_of"}})
	id, err := svc.Relationships().Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: "core.owner_of", SubjectKind: relationship.KindCapacity, SubjectID: cap1, Object: obj, Actor: user,
	})
	if err != nil {
		t.Fatalf("allow-listed Grant: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("nil id")
	}
}

// TestIntegrationGrantPartySubjectSkipsCapacityCheck proves a non-capacity
// subject kind is accepted without the acting-capacity existence check (that
// check applies only when the subject is a capacity), exercising the
// subject-kind passthrough.
func TestIntegrationGrantPartySubjectSkipsCapacityCheck(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	// A committee type whose subject side is a party.
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('committee.party_seat','committee','party','resource','seat') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")
	rel := newRelSvc(h)
	// A party subject id that is not an acting capacity: accepted, because the
	// capacity existence check only runs for capacity subjects.
	id, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: "committee.party_seat", SubjectKind: relationship.KindParty, SubjectID: uuid.New(), Object: obj,
	})
	if err != nil {
		t.Fatalf("party-subject Grant: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("nil id")
	}
}

// TestIntegrationGrantSubjectKindMismatch proves the edge's subject kind must
// match what the relationship type declares — a capacity id cannot be granted
// against a party-subject type.
func TestIntegrationGrantSubjectKindMismatch(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	// seatRelType declares subject_kind='capacity'.
	seedSeatRelType(t, h)
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")
	rel := newRelSvc(h)
	_, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: seatRelType, SubjectKind: relationship.KindParty, SubjectID: uuid.New(), Object: obj,
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("subject-kind mismatch must be Validation, got %v", err)
	}
}

// TestIntegrationGrantUnknownRelTypeAfterOwnership proves an owned-but-unregistered
// relationship type is a clean validation error (not a raw FK violation).
func TestIntegrationGrantUnknownRelTypeAfterOwnership(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")
	rel := newRelSvc(h) // "committee" owns the "committee." prefix
	_, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: "committee.unregistered", SubjectKind: relationship.KindCapacity,
		SubjectID: uuid.New(), Object: obj,
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("unknown rel type must be Validation, got %v", err)
	}
}

// TestIntegrationGrantNonResourceObjectRejected proves the framework grant path
// refuses a relationship type whose declared object side is not a resource.
func TestIntegrationGrantNonResourceObjectRejected(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		 VALUES ('committee.member_of','committee','capacity','party','member') ON CONFLICT (key) DO NOTHING`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	obj := testkit.CreateResourceTypeAndResource(t, h, tenant, "committee.committee")
	rel := newRelSvc(h)
	_, err := rel.Grant(testkit.TenantCtx(tenant), privileged.GrantSpec{
		RelType: "committee.member_of", SubjectKind: relationship.KindCapacity,
		SubjectID: uuid.New(), Object: obj,
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("non-resource object type must be Validation, got %v", err)
	}
}

func TestGrantValidationErrors(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	seedSeatRelType(t, h)
	rel := newRelSvc(h)
	ctx := testkit.TenantCtx(tenant)

	// Empty subject id.
	if _, err := rel.Grant(ctx, privileged.GrantSpec{
		RelType: seatRelType, Object: resource.Ref{Type: "committee.committee", ID: uuid.New()},
	}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("empty subject must be Validation, got %v", err)
	}
	// Empty object.
	if _, err := rel.Grant(ctx, privileged.GrantSpec{
		RelType: seatRelType, SubjectID: uuid.New(),
	}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("empty object must be Validation, got %v", err)
	}
	// Empty revoke id.
	if err := rel.Revoke(ctx, uuid.Nil, uuid.New()); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("empty revoke id must be Validation, got %v", err)
	}
}

// TestIntegrationActivateAllowListRuleKey proves a module allow-listed for a
// non-prefixed rule key may activate that key's tenant versions.
func TestIntegrationActivateAllowListRuleKey(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	store := ruleStore(t, h) // owns "policy.retention.days"
	id := proposeTenantDraft(t, h, store, tenant)

	// A module named "audit" allow-listed for the policy rule key.
	svc := privileged.New("audit", h.PlatformTxM, store,
		kaudit.New(model.UUIDv7(), nil), model.UUIDv7(),
		privileged.Config{AllowRuleKeys: []string{polRuleKey}})
	if err := svc.Rules().ActivateTenant(testkit.TenantCtx(tenant), id, uuid.New(), privileged.ActivateOptions{}); err != nil {
		t.Fatalf("allow-listed ActivateTenant: %v", err)
	}
	if got := statusOf(t, h, id); got != "active" {
		t.Fatalf("want active, got %q", got)
	}
	// Validation: empty version id.
	if err := svc.Rules().ActivateTenant(testkit.TenantCtx(tenant), uuid.Nil, uuid.New(), privileged.ActivateOptions{}); kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("empty version id must be Validation, got %v", err)
	}
}
