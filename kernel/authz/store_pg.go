package authz

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// PgStore is the Postgres-backed authz.Store. Every method runs on the caller's
// TenantDB — the request's own tenant transaction — so all authz reads share
// one MVCC snapshot with the request's writes and open no extra connections
// (review finding ARCH-36). PgStore is therefore stateless. RLS scopes reads to
// the tenant; global rows (platform roles/policies with tenant_id IS NULL) are
// admitted by their read policy.
type PgStore struct{}

// NewStore builds the authz store.
func NewStore() *PgStore { return &PgStore{} }

var _ Store = (*PgStore)(nil)

// ActiveAssignments loads the actor's active role assignments with each role's
// permission keys aggregated. The actor is matched by capacity_id (human) or
// system_actor (non-human); temporal validity is [valid_from, valid_to) at at.
func (s *PgStore) ActiveAssignments(ctx context.Context, db database.TenantDB, a Actor, at time.Time) ([]Assignment, error) {
	var capArg, sysArg any
	if a.CapacityID != uuid.Nil {
		capArg = a.CapacityID
	}
	if a.System != "" {
		sysArg = a.System
	}

	const q = `
SELECT aa.id, r.key, aa.scope_kind, aa.scope_id, aa.scope_type,
       COALESCE(array_agg(rp.permission_key) FILTER (WHERE rp.permission_key IS NOT NULL), '{}') AS perms
FROM actor_assignments aa
JOIN roles r ON r.id = aa.role_id
LEFT JOIN role_permissions rp ON rp.role_id = r.id
WHERE aa.valid_from <= $1 AND (aa.valid_to IS NULL OR aa.valid_to > $1)
  AND (($2::uuid IS NOT NULL AND aa.capacity_id = $2)
       OR ($3::text IS NOT NULL AND aa.system_actor = $3))
GROUP BY aa.id, r.key, aa.scope_kind, aa.scope_id, aa.scope_type`

	rows, err := db.Query(ctx, q, at, capArg, sysArg)
	if err != nil {
		return nil, kerr.Wrapf(err, "authz.ActiveAssignments", "load assignments")
	}
	defer rows.Close()
	var out []Assignment
	for rows.Next() {
		var (
			asg       Assignment
			scopeKind string
			scopeID   *uuid.UUID
			scopeType *string
		)
		if err := rows.Scan(&asg.ID, &asg.RoleKey, &scopeKind, &scopeID, &scopeType, &asg.Perms); err != nil {
			return nil, kerr.Wrapf(err, "authz.ActiveAssignments", "scan")
		}
		asg.ScopeKind = ScopeKind(scopeKind)
		if scopeID != nil {
			asg.ScopeID = *scopeID
		}
		if scopeType != nil {
			asg.ScopeType = *scopeType
		}
		out = append(out, asg)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "authz.ActiveAssignments", "iterate")
	}
	return out, nil
}

// OrgAncestors returns orgID then each ancestor walking parent_org_id upward
// (self-first). The recursive CTE is cycle-guarded (SEC-30): UNION dedups and a
// depth cap stops a malicious/broken org cycle from running away.
func (s *PgStore) OrgAncestors(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
	if orgID == uuid.Nil {
		return nil, nil
	}
	const q = `
WITH RECURSIVE anc AS (
    SELECT id, parent_org_id, 0 AS depth FROM organizations WHERE id = $1
    UNION
    SELECT o.id, o.parent_org_id, anc.depth + 1
    FROM organizations o JOIN anc ON o.id = anc.parent_org_id
    WHERE anc.depth < 64)
SELECT id FROM anc ORDER BY depth`
	return orgIDs(ctx, db, "authz.OrgAncestors", q, orgID)
}

// OrgSubtree returns orgID then all descendant org ids (walking children
// downward), cycle-guarded like OrgAncestors.
func (s *PgStore) OrgSubtree(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error) {
	if orgID == uuid.Nil {
		return nil, nil
	}
	const q = `
WITH RECURSIVE sub AS (
    SELECT id, 0 AS depth FROM organizations WHERE id = $1
    UNION
    SELECT o.id, sub.depth + 1
    FROM organizations o JOIN sub ON o.parent_org_id = sub.id
    WHERE sub.depth < 64)
SELECT id FROM sub ORDER BY depth`
	return orgIDs(ctx, db, "authz.OrgSubtree", q, orgID)
}

func orgIDs(ctx context.Context, db database.TenantDB, op, q string, arg uuid.UUID) ([]uuid.UUID, error) {
	rows, err := db.Query(ctx, q, arg)
	if err != nil {
		return nil, kerr.Wrapf(err, op, "walk org tree")
	}
	defer rows.Close()
	var out []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, kerr.Wrapf(err, op, "scan")
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, op, "iterate")
	}
	return out, nil
}

// Policies returns active policies applicable to perm on resource type rt, with
// their conditions, ordered by priority ascending. A policy applies when its
// applies_to_permission matches (or is NULL = any permission) and its
// applies_to_resource_type matches (or is NULL = any type). When rt is empty
// (a check with no resource type) ONLY type-agnostic policies apply — a policy
// bound to a specific resource type must never leak into a typeless check
// (review finding SEC-27).
func (s *PgStore) Policies(ctx context.Context, db database.TenantDB, a Actor, perm, rt string) ([]Policy, error) {
	const qPolicies = `
SELECT id, key, effect, priority
FROM policies
WHERE status = 'active'
  AND (applies_to_permission = $1 OR applies_to_permission IS NULL)
  AND (applies_to_resource_type IS NULL
       OR ($2 <> '' AND applies_to_resource_type = $2))
ORDER BY priority`
	const qConds = `
SELECT policy_id, attribute, op, value
FROM policy_conditions
WHERE policy_id = ANY($1)`

	rows, err := db.Query(ctx, qPolicies, perm, rt)
	if err != nil {
		return nil, kerr.Wrapf(err, "authz.Policies", "load policies")
	}
	byID := map[uuid.UUID]int{}
	var ids []uuid.UUID
	var out []Policy
	for rows.Next() {
		var (
			p      Policy
			effect string
		)
		if err := rows.Scan(&p.ID, &p.Key, &effect, &p.Priority); err != nil {
			rows.Close()
			return nil, kerr.Wrapf(err, "authz.Policies", "scan")
		}
		p.Effect = PolicyEffect(effect)
		byID[p.ID] = len(out)
		ids = append(ids, p.ID)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, kerr.Wrapf(err, "authz.Policies", "iterate")
	}
	rows.Close()
	if len(ids) == 0 {
		return nil, nil
	}

	crows, err := db.Query(ctx, qConds, ids)
	if err != nil {
		return nil, kerr.Wrapf(err, "authz.Policies", "load conditions")
	}
	defer crows.Close()
	for crows.Next() {
		var (
			pid uuid.UUID
			c   Condition
			val []byte
		)
		if err := crows.Scan(&pid, &c.Attribute, &c.Op, &val); err != nil {
			return nil, kerr.Wrapf(err, "authz.Policies", "scan condition")
		}
		c.Value = json.RawMessage(val)
		if idx, ok := byID[pid]; ok {
			out[idx].Conditions = append(out[idx].Conditions, c)
		}
	}
	if err := crows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "authz.Policies", "iterate conditions")
	}
	return out, nil
}

// ResourceOrg returns the org id owning a resource (RLS-scoped), or the zero
// uuid when the resource is unknown or has no org. The resource type is matched
// too, so a mismatched {Type, ID} does not silently resolve a wrong org
// (review finding ARCH-46).
func (s *PgStore) ResourceOrg(ctx context.Context, db database.TenantDB, ref resource.Ref) (uuid.UUID, error) {
	var orgID *uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT org_id FROM resources WHERE id = $1 AND resource_type = $2`, ref.ID, ref.Type).Scan(&orgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "authz.ResourceOrg", "resolve resource org")
	}
	if orgID != nil {
		return *orgID, nil
	}
	return uuid.Nil, nil
}
