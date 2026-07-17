package testkit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/qatoolist/wowapi/kernel/database"
)

// DATA-01 / MATRIX CS-18 adversarial matrix: a seeded cross-tenant parent/child
// insert must FAIL on every confirmed tenant-FK edge. RLS proves
// the child row's own tenant; only a composite FK on (tenant_id, <ref>) pointing
// at the parent's UNIQUE (tenant_id, id) proves parent and child AGREE — FK
// lookups bypass RLS by design, so before migrations 00034/00035/00036 a child
// in tenant B could reference a parent in tenant A (CS-18: "platform-role seeded
// cross-tenant parent/child insert succeeds today; fails after").
//
// Three probes per edge, deliberately graded by privilege:
//
//   - admin (BYPASSRLS superuser — the "platform provisioner" posture CS-18
//     names): RLS never runs, so the ONLY thing that can block the mismatch is
//     the composite FK itself. Post-migration this must fail with SQLSTATE
//     23503 (foreign_key_violation) on all 9 edges.
//   - app_rt: the production runtime role. The cross-tenant row passes the
//     tenant-isolation WITH CHECK (its own tenant_id matches the binding), so
//     only the composite FK blocks it → 23503. Exception: the
//     document_access_grants edge carries a RESTRICTIVE owner-write policy
//     whose sub-SELECT on documents is itself RLS-bound, so that edge is
//     blocked at the policy layer (42501) before the FK is consulted.
//   - app_platform: holds INSERT on document_versions only (least-privilege
//     grants). On that edge the FK result is asserted directly — PLAN T7's
//     "confirm platform role doesn't bypass FK constraints — don't assume".
//     On the other 7 the failure is a grant denial (42501), and the FK
//     backstop for a hypothetical future grant widening is what the admin
//     probe proves (BYPASSRLS does not bypass referential integrity).
//
// TestIntegrationTenantFKEdgeCensus keeps this hand-written matrix honest
// against the live catalog: every strict-tenant→strict-tenant FK edge that
// carries a composite tenant FK must be probed here, and vice versa.

// tenantFKEdge is one confirmed DATA-01 edge of the adversarial matrix.
type tenantFKEdge struct {
	child      string // child table
	fkCol      string // referencing column on child
	parent     string // parent table
	constraint string // composite FK name added by migration 00035

	// restrictive marks the document_access_grants edge: its RESTRICTIVE
	// owner-write policy (RLS-bound sub-SELECT on documents) blocks the
	// cross-tenant reference under app_rt/app_platform BEFORE the FK runs,
	// so those probes accept 42501 as well as 23503. The admin probe still
	// pins this edge to a pure 23503.
	restrictive bool

	// seedParent inserts one parent row under tenant via the Admin pool
	// (RLS-bypassing, like a provisioner) and returns its id.
	seedParent func(t *testing.T, h *DBHandle, tenant uuid.UUID) uuid.UUID
	// childRow returns the child's non-tenant columns, referencing parentID
	// through fkCol. tenant is the tenant the row will be written under.
	childRow func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any
}

// tenantFKEdges is the DATA-01 matrix (PLAN's original evidence set plus each
// subsequently introduced tenant FK), re-confirmed mechanically by
// internal/tools/tenantfk and the edge census test below.
func tenantFKEdges() []tenantFKEdge {
	return []tenantFKEdge{
		{
			child: "persons", fkCol: "party_id", parent: "parties",
			constraint: "persons_party_id_tenant_fkey",
			seedParent: seedParty,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{"party_id": parentID, "given_name": "Given"}
			},
		},
		{
			child: "legal_entities", fkCol: "party_id", parent: "parties",
			constraint: "legal_entities_party_id_tenant_fkey",
			seedParent: seedParty,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{"party_id": parentID, "legal_name": "Acme " + randHex(4)}
			},
		},
		{
			child: "party_contacts", fkCol: "party_id", parent: "parties",
			constraint: "party_contacts_party_id_tenant_fkey",
			seedParent: seedParty,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{
					"id": uuid.New(), "party_id": parentID, "kind": "email",
					"value": randHex(8) + "@example.test", "created_by": uuid.Nil,
				}
			},
		},
		{
			child: "acting_capacities", fkCol: "party_id", parent: "parties",
			constraint: "acting_capacities_party_id_tenant_fkey",
			seedParent: seedParty,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{
					"id": uuid.New(), "user_id": CreateUser(t, h), "party_id": parentID,
					"label": "member", "created_by": uuid.Nil,
				}
			},
		},
		{
			child: "resources", fkCol: "org_id", parent: "organizations",
			constraint: "resources_org_id_tenant_fkey",
			seedParent: func(t *testing.T, h *DBHandle, tenant uuid.UUID) uuid.UUID {
				return CreateOrg(t, h, tenant, nil, "Org "+randHex(6))
			},
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{
					"id": uuid.New(), "resource_type": seedResourceType(t, h),
					"org_id": parentID, "label": "r", "created_by": uuid.Nil,
				}
			},
		},
		{
			child: "document_versions", fkCol: "document_id", parent: "documents",
			constraint: "document_versions_document_id_tenant_fkey",
			seedParent: seedDocument,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{
					"id": uuid.New(), "document_id": parentID, "version_no": 1,
					"storage_key": "s/" + randHex(8), "mime_type": "text/plain",
					"size_bytes": int64(1), "checksum_sha256": randHex(16), "uploaded_by": uuid.Nil,
				}
			},
		},
		{
			child: "document_access_grants", fkCol: "document_id", parent: "documents",
			constraint:  "document_access_grants_document_id_tenant_fkey",
			restrictive: true,
			seedParent:  seedDocument,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{
					"id": uuid.New(), "document_id": parentID, "grantee_kind": "role",
					"grantee_ref": "r-" + randHex(4), "access": "read", "created_by": uuid.Nil,
				}
			},
		},
		{
			child: "attachments", fkCol: "document_version_id", parent: "document_versions",
			constraint: "attachments_document_version_id_tenant_fkey",
			seedParent: seedDocumentVersion,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{
					"id": uuid.New(), "resource_type": "document", "resource_id": uuid.New(),
					"document_version_id": parentID, "created_by": uuid.Nil,
				}
			},
		},
		{
			child: "webhook_failed_signature_audit", fkCol: "endpoint_id", parent: "webhook_endpoints",
			constraint: "webhook_failed_signature_audit_tenant_id_endpoint_id_fkey",
			seedParent: seedWebhookEndpoint,
			childRow: func(t *testing.T, h *DBHandle, tenant, parentID uuid.UUID) map[string]any {
				return map[string]any{
					"id": uuid.New(), "endpoint_id": parentID,
					"event_type": "test.event", "failure_reason": "invalid signature",
				}
			},
		},
	}
}

// fkViolation is SQLSTATE 23503 (foreign_key_violation); rlsDenied is 42501
// (insufficient_privilege — both a missing grant and a "new row violates
// row-level security policy" rejection surface as 42501).
const (
	fkViolation = "23503"
	rlsDenied   = "42501"
)

// sqlState extracts the PgError SQLSTATE, or "" for a nil/non-PG error.
func sqlState(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// adminInsert inserts cols (INCLUDING tenant_id) into table via the Admin pool,
// mirroring insertRow's parameterized construction. Admin is a BYPASSRLS
// superuser login: RLS and grants never run, so the composite FK is the only
// integrity layer left — exactly the CS-18 "platform-role seeded" posture.
func adminInsert(ctx context.Context, h *DBHandle, table string, cols map[string]any) error {
	keys := sortedKeys(cols)
	names := make([]string, len(keys))
	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, k := range keys {
		if !identRE.MatchString(k) {
			return fmt.Errorf("testkit: invalid column name %q", k)
		}
		names[i] = quoteIdent(k)
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = cols[k]
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdent(table), strings.Join(names, ", "), strings.Join(placeholders, ", "))
	_, err := h.Admin.Exec(ctx, sql, args...)
	return err
}

// TestIntegrationTenantFKCrossTenantInsertBlocked is the DATA-01 T7 / CS-18
// adversarial matrix: for each edge, seed the parent in tenant A and
// attempt a child insert in tenant B that references it, under admin
// (BYPASSRLS), app_rt, and app_platform. Every attempt must fail; the admin
// probe must fail with a pure foreign_key_violation, proving the database's own
// referential-integrity machinery — not RLS convention — enforces tenant
// agreement. Before migrations 00034/00035/00036 these inserts succeed (the
// captured fail-first state).
func TestIntegrationTenantFKCrossTenantInsertBlocked(t *testing.T) {
	h := NewDB(t)

	for _, e := range tenantFKEdges() {
		t.Run(e.child+"_"+e.fkCol, func(t *testing.T) {
			tenantA := CreateTenant(t, h).ID
			tenantB := CreateTenant(t, h).ID

			// Each probe seeds its OWN parent in tenant A: two edges' child PK or
			// unique key is derived from the parent id (persons/legal_entities PK,
			// document_versions (document_id, version_no)), so reusing one parent
			// across probes would fail on 23505 instead of exercising the FK.

			// Probe 1 — admin (BYPASSRLS provisioner). Only the composite FK
			// can block; require exactly foreign_key_violation.
			parentA := e.seedParent(t, h, tenantA)
			cols := withTenant(e.childRow(t, h, tenantB, parentA), tenantB)
			err := adminInsert(context.Background(), h, e.child, cols)
			switch {
			case err == nil:
				t.Errorf("[%s→%s] admin (BYPASSRLS): cross-tenant insert SUCCEEDED — "+
					"nothing proves parent and child agree on tenant (DATA-01 gap open)",
					e.child, e.parent)
			case sqlState(err) != fkViolation:
				t.Errorf("[%s→%s] admin (BYPASSRLS): want SQLSTATE %s (foreign_key_violation), got %s: %v",
					e.child, e.parent, fkViolation, sqlState(err), err)
			}

			// Probe 2 — app_rt (production runtime role, tenant B bound).
			ctxB := database.WithActorID(database.WithTenantID(context.Background(), tenantB), uuid.New())
			parentA = e.seedParent(t, h, tenantA)
			err = h.TxM.WithTenant(ctxB, func(ctx context.Context, db database.TenantDB) error {
				return insertRow(ctx, db, e.child, withTenant(e.childRow(t, h, tenantB, parentA), tenantB))
			})
			assertCrossTenantBlocked(t, "app_rt", e, err)

			// Probe 3 — app_platform (tenant B bound). Holds INSERT only on
			// document_versions; elsewhere the grant denial (42501) is itself the
			// defense, with the FK backstop proven by probe 1.
			parentA = e.seedParent(t, h, tenantA)
			err = h.PlatformTxM.WithTenant(ctxB, func(ctx context.Context, db database.TenantDB) error {
				return insertRow(ctx, db, e.child, withTenant(e.childRow(t, h, tenantB, parentA), tenantB))
			})
			assertCrossTenantBlocked(t, "app_platform", e, err)
		})
	}
}

// assertCrossTenantBlocked requires the cross-tenant insert to have failed.
// Non-restrictive edges under a role holding INSERT must fail on the FK itself
// (23503); a 42501 is also accepted where the role lacks the grant or a
// RESTRICTIVE policy fires first — both are failures of the adversarial insert,
// and the pure-FK proof is pinned by the admin probe.
func assertCrossTenantBlocked(t *testing.T, role string, e tenantFKEdge, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("[%s→%s] %s: cross-tenant insert SUCCEEDED, want FK violation (DATA-01 gap open)",
			e.child, e.parent, role)
		return
	}
	if code := sqlState(err); code != fkViolation && code != rlsDenied {
		t.Errorf("[%s→%s] %s: want SQLSTATE %s or %s, got %s: %v",
			e.child, e.parent, role, fkViolation, rlsDenied, code, err)
	}
}

// TestIntegrationTenantFKEdgeCensus keeps the hand-written edge matrix honest
// against the live catalog (the same self-maintenance posture as
// TestIntegrationRLSCensusComplete): the set of composite tenant FKs actually
// present in the schema must be exactly the set probed above — an edge gaining
// a composite FK without an adversarial probe here, or a probed constraint
// disappearing from the schema, fails the suite.
func TestIntegrationTenantFKEdgeCensus(t *testing.T) {
	h := NewDB(t)

	rows, err := h.Admin.Query(context.Background(), `
		SELECT con.conname
		  FROM pg_constraint con
		  JOIN pg_class child ON child.oid = con.conrelid
		  JOIN pg_namespace n ON n.oid = child.relnamespace
		 WHERE n.nspname = 'public' AND con.contype = 'f'
		   AND array_length(con.conkey, 1) > 1
		   AND EXISTS (
		       SELECT 1 FROM unnest(con.conkey) k
		         JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k
		        WHERE a.attname = 'tenant_id')`)
	if err != nil {
		t.Fatalf("query composite tenant FKs: %v", err)
	}
	defer rows.Close()

	live := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan constraint name: %v", err)
		}
		live[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate composite tenant FKs: %v", err)
	}

	probed := map[string]bool{}
	for _, e := range tenantFKEdges() {
		probed[e.constraint] = true
		if !live[e.constraint] {
			t.Errorf("edge %s.%s→%s: composite FK %q not present in the live schema",
				e.child, e.fkCol, e.parent, e.constraint)
		}
	}
	for name := range live {
		if !probed[name] {
			t.Errorf("composite tenant FK %q exists in the schema but has no adversarial probe in tenantFKEdges()", name)
		}
	}
}
