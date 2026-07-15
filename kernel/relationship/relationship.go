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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
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
// DATA-07 T1/T2: the evaluator resolves the actor to the subject_kind declared
// by relationship_types for relType. Supported actor-resolvable kinds are
// 'capacity' (subject.CapacityID) and 'party' (party_id of the actor's active
// capacity). 'resource' subject_kind is not actor-resolvable and fails closed.
// Any unenumerated subject_kind returns a distinct permission-denied error so
// it cannot be mistaken for an infrastructure fault.
func (c *Checker) Has(ctx context.Context, db database.TenantDB, subject authz.Actor, relType string, obj resource.Ref, at time.Time) (bool, error) {
	var subjectKind string
	if err := db.QueryRow(ctx,
		`SELECT subject_kind FROM relationship_types WHERE key = $1`, relType).Scan(&subjectKind); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, kerr.E(kerr.KindForbidden, "unsupported_rel_type",
				fmt.Sprintf("relationship type %q is not registered", relType), kerr.Op("relationship.Has"))
		}
		return false, kerr.Wrapf(err, "relationship.Has", "lookup subject kind for %s", relType)
	}

	subjectID, err := c.resolveSubject(ctx, db, subject, subjectKind)
	if err != nil {
		return false, err
	}
	if subjectID == uuid.Nil {
		return false, nil
	}

	const q = `
SELECT EXISTS (
    SELECT 1 FROM relationships
    WHERE rel_type = $1
      AND object_kind = 'resource' AND object_id = $2
      AND subject_kind = $3 AND subject_id = $4
      AND valid_from <= $5 AND (valid_to IS NULL OR valid_to > $5))`
	var has bool
	if err := db.QueryRow(ctx, q, relType, obj.ID, subjectKind, subjectID, at).Scan(&has); err != nil {
		return false, kerr.Wrapf(err, "relationship.Has", "check %s edge", relType)
	}
	return has, nil
}

// resolveSubject maps the actor to a subject_id for the requested subject_kind.
// It returns uuid.Nil when the actor cannot resolve to that kind (a safe
// "no edge" result for valid, enumerated kinds that do not apply to actors).
// An unenumerated kind returns a permission-denied error (fail closed).
func (c *Checker) resolveSubject(ctx context.Context, db database.DBTX, subject authz.Actor, kind string) (uuid.UUID, error) {
	switch kind {
	case KindCapacity:
		return subject.CapacityID, nil
	case KindParty:
		if subject.CapacityID == uuid.Nil {
			return uuid.Nil, nil
		}
		var partyID uuid.UUID
		if err := db.QueryRow(ctx,
			`SELECT party_id FROM acting_capacities WHERE id = $1 AND tenant_id = app_tenant_id()`,
			subject.CapacityID).Scan(&partyID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return uuid.Nil, nil
			}
			return uuid.Nil, kerr.Wrapf(err, "relationship.resolveSubject", "lookup party for capacity %s", subject.CapacityID)
		}
		return partyID, nil
	case KindResource:
		// Actors do not stand in relationships as resources.
		return uuid.Nil, nil
	default:
		return uuid.Nil, kerr.E(kerr.KindForbidden, "unsupported_subject_kind",
			fmt.Sprintf("subject_kind %q is not supported", kind), kerr.Op("relationship.resolveSubject"))
	}
}

// Relate upserts a relationship edge inside the caller's tenant transaction.
//
// DATA-07 T3/T4 governance: an actor must be bound in ctx (ownership/attribution
// fail-closed); created_by/updated_by are sourced from that actor via
// database.ActorIDFrom (DATA-06 T2). If an active edge already exists it is
// refreshed (valid_to cleared if set) and its version bumped; otherwise a new
// edge with version 1 is inserted. Every mutation writes an audit row in the
// same transaction.
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
// app_tenant_id() so RLS WITH CHECK holds.
func Relate(ctx context.Context, db database.TenantDB, idgen model.IDGen, relType, subjectKind string, subjectID uuid.UUID, objectKind string, objectID uuid.UUID) error {
	actorID, ok := database.ActorIDFrom(ctx)
	if !ok || actorID == uuid.Nil {
		return kerr.E(kerr.KindForbidden, "permission_denied",
			"relationship edge mutation requires a bound actor", kerr.Op("relationship.Relate"))
	}

	const q = `
WITH upsert AS (
    UPDATE relationships
    SET version    = version + 1,
        updated_at = now(),
        updated_by = $7,
        valid_to   = NULL
    WHERE tenant_id = app_tenant_id()
      AND rel_type = $2 AND subject_kind = $3 AND subject_id = $4
      AND object_kind = $5 AND object_id = $6
      AND valid_to IS NULL
    RETURNING id, version
),
inserted AS (
    INSERT INTO relationships
        (id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id,
         valid_from, version, created_at, created_by)
    SELECT $1, app_tenant_id(), $2, $3, $4, $5, $6, now(), 1, now(), $7
    WHERE NOT EXISTS (SELECT 1 FROM upsert)
    RETURNING id, version
)
SELECT id, version FROM upsert
UNION ALL
SELECT id, version FROM inserted`

	var edgeID uuid.UUID
	var version int
	if err := db.QueryRow(ctx, q, idgen.New(), relType, subjectKind, subjectID, objectKind, objectID, actorID).Scan(&edgeID, &version); err != nil {
		return kerr.Wrapf(err, "relationship.Relate", "upsert %s edge", relType)
	}

	writer := audit.New(idgen, nil)
	if err := writer.Record(ctx, db, audit.Entry{
		Action:     "relationship.relate",
		EntityType: "relationship",
		EntityID:   edgeID,
		NewValue:   fmt.Sprintf("%s %s:%s -> %s:%s v=%d", relType, subjectKind, subjectID, objectKind, objectID, version),
	}); err != nil {
		return kerr.Wrapf(err, "relationship.Relate", "audit %s edge", relType)
	}
	return nil
}
