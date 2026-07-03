// Package relationship is wowapi's ReBAC edge store: the tenant relationships
// graph (subject —rel_type→ object) plus the adapter that answers the authz
// kernel's relationship questions. A permission declared granted_via a
// relationship (authz.Permission.GrantedVia) is allowed on a resource target to
// any actor that stands in that relationship to it (blueprint 01 §3 step 4).
//
// The Checker implements authz.RelationshipChecker: Has runs on the caller's
// TenantDB (the request's tenant tx), so it shares the request's snapshot and
// RLS scoping (review finding ARCH-36) and is stateless.
package relationship

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// Subject/object kinds and the relationship_types.subject_kind / object_kind
// check constraint vocabulary (migration 00005).
const (
	KindParty    = "party"
	KindResource = "resource"
	KindCapacity = "capacity"
)

// Checker answers authz.RelationshipChecker.Has against the relationships table.
// Stateless: Has takes the caller's TenantDB.
type Checker struct{}

// NewChecker builds the ReBAC checker.
func NewChecker() *Checker { return &Checker{} }

var _ authz.RelationshipChecker = (*Checker)(nil)

// Has reports whether subject stands in relation relType to obj at time at,
// querying on the caller's tenant tx (RLS-scoped).
//
// Phase 4 resolves the subject via its acting capacity (subject_kind='capacity',
// subject_id=subject.CapacityID), the identity a human actor carries. Edges
// whose subject is a party (subject_kind='party') are not consulted yet.
// TODO(phase-later): also match subject_kind='party' when the actor resolves to
// a party, so party-level ownership grants ReBAC access.
func (c *Checker) Has(ctx context.Context, db database.TenantDB, subject authz.Actor, relType string, obj resource.Ref, at time.Time) (bool, error) {
	// A system/webhook actor has no capacity, so it can hold no relationship edge.
	if subject.CapacityID == uuid.Nil {
		return false, nil
	}
	const q = `
SELECT EXISTS (
    SELECT 1 FROM relationships
    WHERE rel_type = $1
      AND object_kind = 'resource' AND object_id = $2
      AND subject_kind = 'capacity' AND subject_id = $3
      AND valid_from <= $4 AND (valid_to IS NULL OR valid_to > $4))`
	var has bool
	if err := db.QueryRow(ctx, q, relType, obj.ID, subject.CapacityID, at).Scan(&has); err != nil {
		return false, kerr.Wrapf(err, "relationship.Has", "check %s edge", relType)
	}
	return has, nil
}

// Relate inserts a relationship edge inside the caller's tenant transaction.
//
// SECURITY (SEC-24): relationship edges are authorization inputs — a
// granted_via edge grants a permission on its object. Because every module runs
// as the shared app_rt role, this management seam must NOT be exposed to
// arbitrary module code for security-sensitive edge types; edge creation for
// granted_via relationship types is a kernel/platform capability (an audited
// service running as app_platform, wired with the assignment-management API).
// The DB backstop (migration 00005) removes INSERT/UPDATE on `relationships`
// from app_rt for that reason. This function is used by kernel services and
// tests (which seed via the admin/platform role); tenant_id is set from
// app_tenant_id() so RLS WITH CHECK holds. created_by uses a NIL uuid
// placeholder pending actor attribution.
func Relate(ctx context.Context, db database.TenantDB, idgen model.IDGen, relType, subjectKind string, subjectID uuid.UUID, objectKind string, objectID uuid.UUID) error {
	const q = `
INSERT INTO relationships
    (id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id, valid_from, version, created_at, created_by)
VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, now(), 1, now(), $7)`
	if _, err := db.Exec(ctx, q, idgen.New(), relType, subjectKind, subjectID, objectKind, objectID, uuid.Nil); err != nil {
		return kerr.Wrapf(err, "relationship.Relate", "insert %s edge", relType)
	}
	return nil
}
