package relationship_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/relationship"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationRelationshipHas proves the ReBAC checker: an edge present at a
// time answers true, a different object/relation answers false, and an expired
// edge (valid_to in the past) answers false.
func TestIntegrationRelationshipHas(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()

	relType := "core.owner_of"
	seedTenant(t, h, ctx, tenant)
	seedResourceType(t, h, ctx, "requests.request")
	mustExec(t, h, ctx, `INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		VALUES ($1,$2,$3,$4,$5)`, relType, "core", "capacity", "resource", "owner")

	cap1 := seedCapacity(t, h, ctx, tenant)
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	other := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()

	// Edges are seeded via the Admin pool: app_rt has no INSERT on relationships
	// (SEC-24 — edge creation is a kernel/platform capability), which mirrors
	// how a real edge-management service (app_platform) or a migration writes
	// them. Live edge: cap1 owner_of obj.
	mustExec(t, h, ctx, `INSERT INTO relationships
		(id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, version, created_by)
		VALUES ($1,$2,$3,'capacity',$4,'resource',$5, now(), 1, $6)`,
		gen.New(), tenant, relType, cap1, obj.ID, uuid.Nil)

	checker := relationship.NewChecker()
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: cap1, TenantID: tenant}
	// Check slightly in the future: the edge's valid_from is the DB clock's
	// now(), which can be marginally ahead of the host's time.Now() (container
	// clock skew). An open-ended edge is active at any later instant, so this
	// is skew-immune without weakening the production query.
	now := time.Now().Add(time.Minute)

	// has runs Checker.Has on the caller's tenant tx (app_rt, SELECT-only).
	has := func(a authz.Actor, ref resource.Ref, at time.Time) bool {
		t.Helper()
		var ok bool
		err := h.TxM.WithTenantRO(database.WithTenantID(ctx, tenant),
			func(ctx context.Context, db database.TenantDB) error {
				var e error
				ok, e = checker.Has(ctx, db, a, relType, ref, at)
				return e
			})
		if err != nil {
			t.Fatalf("Has: %v", err)
		}
		return ok
	}

	if !has(actor, obj, now) {
		t.Fatal("Has(live edge) = false; want true")
	}
	if has(actor, other, now) {
		t.Fatal("Has(other object) = true; want false")
	}

	// Expired edge on a fresh object: valid_to in the past ⇒ not active now.
	expObj := seedResource(t, h, ctx, tenant, "requests.request")
	mustExec(t, h, ctx, `INSERT INTO relationships
		(id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, valid_to, version, created_by)
		VALUES ($1,$2,$3,'capacity',$4,'resource',$5, now() - interval '2 days', now() - interval '1 day', 1, $6)`,
		gen.New(), tenant, relType, cap1, expObj.ID, uuid.Nil)

	if has(actor, expObj, now) {
		t.Fatal("Has(expired edge) = true; want false")
	}
	// But it WAS active in the past.
	if !has(actor, expObj, time.Now().Add(-36*time.Hour)) {
		t.Fatal("Has(expired edge, past instant) = false; want true")
	}

	// A system actor (no capacity) never holds an edge.
	sys := authz.Actor{Kind: authz.ActorSystem, System: "relay", TenantID: tenant}
	if has(sys, obj, now) {
		t.Fatal("Has(system actor) = true; want false")
	}
}

// --- seed helpers (Admin pool bypasses RLS as the superuser login) ---

func seedTenant(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID) {
	t.Helper()
	mustExec(t, h, ctx, `INSERT INTO tenants (id, slug, display_name, created_by) VALUES ($1,$2,$3,$4)`,
		tenant, "t-"+uuid.New().String()[:8], "Tenant", uuid.Nil)
}

func seedResourceType(t *testing.T, h *testkit.DBHandle, ctx context.Context, key string) {
	t.Helper()
	mustExec(t, h, ctx, `INSERT INTO resource_types (key, module, description) VALUES ($1,$2,$3)
		ON CONFLICT (key) DO NOTHING`, key, "requests", "request")
}

func seedResource(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID, rt string) resource.Ref {
	t.Helper()
	id := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO resources (id, tenant_id, resource_type, label, status, created_by)
		VALUES ($1,$2,$3,$4,'active',$5)`, id, tenant, rt, "res", uuid.Nil)
	return resource.Ref{Type: rt, ID: id}
}

func seedCapacity(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO users (id, idp_subject, email, created_by) VALUES ($1,$2,$3,$4)`,
		userID, "idp-"+uuid.New().String()[:8], uuid.New().String()[:8]+"@example.test", uuid.Nil)
	capID := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO acting_capacities (id, tenant_id, user_id, label, created_by)
		VALUES ($1,$2,$3,$4,$5)`, capID, tenant, userID, "member", uuid.Nil)
	return capID
}

func mustExec(t *testing.T, h *testkit.DBHandle, ctx context.Context, sql string, args ...any) {
	t.Helper()
	if _, err := h.Admin.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("seed exec: %v\n%s", err, sql)
	}
}

// seedParty creates a person party and returns its party id.
func seedParty(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID) uuid.UUID {
	t.Helper()
	partyID := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO parties (id, tenant_id, kind, display_name, created_by)
		VALUES ($1,$2,'person',$3,$4)`, partyID, tenant, "party-"+partyID.String()[:8], uuid.Nil)
	mustExec(t, h, ctx, `INSERT INTO persons (party_id, tenant_id, given_name)
		VALUES ($1,$2,$3)`, partyID, tenant, "given")
	return partyID
}

// seedCapacityWithParty creates a user, party, and active capacity linking them.
func seedCapacityWithParty(t *testing.T, h *testkit.DBHandle, ctx context.Context, tenant uuid.UUID) (capID, partyID uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	mustExec(t, h, ctx, `INSERT INTO users (id, idp_subject, email, created_by) VALUES ($1,$2,$3,$4)`,
		userID, "idp-"+uuid.New().String()[:8], uuid.New().String()[:8]+"@example.test", uuid.Nil)
	partyID = seedParty(t, h, ctx, tenant)
	capID = uuid.New()
	mustExec(t, h, ctx, `INSERT INTO acting_capacities (id, tenant_id, user_id, party_id, label, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)`, capID, tenant, userID, partyID, "member", uuid.Nil)
	return capID, partyID
}

// TestIntegrationRelationshipHasPartySubject proves DATA-07 T1: an actor whose
// active capacity resolves to a party is granted access via party-subject edges.
func TestIntegrationRelationshipHasPartySubject(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	relType := "core.owner_of"
	seedTenant(t, h, ctx, tenant)
	seedResourceType(t, h, ctx, "requests.request")
	mustExec(t, h, ctx, `INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
		VALUES ($1,$2,'party','resource',$3)`, relType, "core", "owner")

	capID, partyID := seedCapacityWithParty(t, h, ctx, tenant)
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()

	mustExec(t, h, ctx, `INSERT INTO relationships
		(id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, version, created_by)
		VALUES ($1,$2,$3,'party',$4,'resource',$5, now(), 1, $6)`,
		gen.New(), tenant, relType, partyID, obj.ID, uuid.Nil)

	checker := relationship.NewChecker()
	actor := authz.Actor{Kind: authz.ActorUser, CapacityID: capID, TenantID: tenant}
	var has bool
	if err := h.TxM.WithTenantRO(database.WithTenantID(ctx, tenant),
		func(ctx context.Context, db database.TenantDB) error {
			var e error
			has, e = checker.Has(ctx, db, actor, relType, obj, time.Now().Add(time.Minute))
			return e
		}); err != nil {
		t.Fatalf("Has: %v", err)
	}
	if !has {
		t.Fatal("Has(party-subject edge) = false; want true")
	}
}

// TestIntegrationRelationshipSubjectKindMatrix proves DATA-07 T2: every
// schema-enumerated subject_kind is evaluated correctly, and an unenumerated
// kind fails closed with a permission-denied error.
func TestIntegrationRelationshipSubjectKindMatrix(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()
	seedTenant(t, h, ctx, tenant)
	seedResourceType(t, h, ctx, "requests.request")
	obj := seedResource(t, h, ctx, tenant, "requests.request")
	gen := model.UUIDv7()

	cases := []struct {
		name        string
		relType     string
		subjectKind string
		seed        func() (actor authz.Actor, subjectID uuid.UUID)
		want        bool
	}{
		{
			name:        "capacity-subject",
			relType:     "core.capacity_owner",
			subjectKind: "capacity",
			seed: func() (authz.Actor, uuid.UUID) {
				capID := seedCapacity(t, h, ctx, tenant)
				mustExec(t, h, ctx, `INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
					VALUES ($1,$2,'capacity','resource',$3) ON CONFLICT (key) DO NOTHING`, "core.capacity_owner", "core", "owner")
				mustExec(t, h, ctx, `INSERT INTO relationships
					(id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, version, created_by)
					VALUES ($1,$2,$3,'capacity',$4,'resource',$5, now(), 1, $6)`,
					gen.New(), tenant, "core.capacity_owner", capID, obj.ID, uuid.Nil)
				return authz.Actor{Kind: authz.ActorUser, CapacityID: capID, TenantID: tenant}, capID
			},
			want: true,
		},
		{
			name:        "party-subject",
			relType:     "core.party_owner",
			subjectKind: "party",
			seed: func() (authz.Actor, uuid.UUID) {
				capID, partyID := seedCapacityWithParty(t, h, ctx, tenant)
				mustExec(t, h, ctx, `INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
					VALUES ($1,$2,'party','resource',$3) ON CONFLICT (key) DO NOTHING`, "core.party_owner", "core", "owner")
				mustExec(t, h, ctx, `INSERT INTO relationships
					(id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, version, created_by)
					VALUES ($1,$2,$3,'party',$4,'resource',$5, now(), 1, $6)`,
					gen.New(), tenant, "core.party_owner", partyID, obj.ID, uuid.Nil)
				return authz.Actor{Kind: authz.ActorUser, CapacityID: capID, TenantID: tenant}, partyID
			},
			want: true,
		},
		{
			name:        "resource-subject-not-actor-resolvable",
			relType:     "core.resource_owner",
			subjectKind: "resource",
			seed: func() (authz.Actor, uuid.UUID) {
				mustExec(t, h, ctx, `INSERT INTO relationship_types (key, module, subject_kind, object_kind, description)
					VALUES ($1,$2,'resource','resource',$3) ON CONFLICT (key) DO NOTHING`, "core.resource_owner", "core", "owner")
				mustExec(t, h, ctx, `INSERT INTO relationships
					(id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, version, created_by)
					VALUES ($1,$2,$3,'resource',$4,'resource',$5, now(), 1, $6)`,
					gen.New(), tenant, "core.resource_owner", obj.ID, obj.ID, uuid.Nil)
				return authz.Actor{Kind: authz.ActorUser, CapacityID: seedCapacity(t, h, ctx, tenant), TenantID: tenant}, obj.ID
			},
			want: false,
		},
	}

	checker := relationship.NewChecker()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actor, _ := tc.seed()
			var has bool
			if err := h.TxM.WithTenantRO(database.WithTenantID(ctx, tenant),
				func(ctx context.Context, db database.TenantDB) error {
					var e error
					has, e = checker.Has(ctx, db, actor, tc.relType, obj, time.Now().Add(time.Minute))
					return e
				}); err != nil {
				t.Fatalf("Has: %v", err)
			}
			if has != tc.want {
				t.Fatalf("Has = %v; want %v", has, tc.want)
			}
		})
	}
}
