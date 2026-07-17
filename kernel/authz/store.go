package authz

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// Assignment is one active role grant at a scope, with the role's permission
// keys pre-joined so the evaluator needs no second query. Loaded by Store.
type Assignment struct {
	ID        uuid.UUID
	RoleKey   string
	ScopeKind ScopeKind
	ScopeID   uuid.UUID // org id (org scope) or resource id (resource scope); zero for tenant
	ScopeType string    // resource_type key when ScopeKind == ScopeResourceType
	Perms     []string  // permission keys the role grants
}

// grants reports whether this assignment's role includes perm.
func (a Assignment) grants(perm string) bool {
	for _, p := range a.Perms {
		if p == perm {
			return true
		}
	}
	return false
}

// PolicyEffect is allow or deny.
type PolicyEffect string

const (
	EffectAllow PolicyEffect = "allow"
	EffectDeny  PolicyEffect = "deny"
)

// Condition is one ABAC predicate over the attribute bag.
type Condition struct {
	Attribute string          // "resource.status", "actor.relationship", "env.time_of_day"
	Op        string          // eq|neq|in|not_in|contains|within|gte|lte
	Value     json.RawMessage // comparison operand
}

// Policy is an active ABAC policy applicable to a permission/resource type.
type Policy struct {
	ID         uuid.UUID
	Key        string
	Effect     PolicyEffect
	Priority   int // lower first
	Conditions []Condition
}

// Store loads the authorization facts for a decision. It is the only DB seam;
// the evaluator is pure over it, so unit tests use an in-memory fake. Every
// method runs on the caller's TenantDB — the request's own tenant transaction —
// so all reads share one snapshot and see the request's uncommitted writes, and
// no extra connections are opened on the hot path (ARCH-36). RLS scopes reads;
// global rows (platform roles/policies, tenant_id IS NULL) are admitted by
// their read policy.
type Store interface {
	// ActiveAssignments returns the actor's active assignments (role perms
	// joined) at time at.
	ActiveAssignments(ctx context.Context, db database.TenantDB, a Actor, at time.Time) ([]Assignment, error)
	// OrgAncestors returns orgID and all its ancestor org ids (self-first), so
	// the evaluator can decide whether an org-scoped assignment (a grant at an
	// ancestor org) covers a target in orgID. Empty/zero orgID → empty.
	OrgAncestors(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error)
	// OrgSubtree returns orgID and all descendant org ids, for building list
	// filters from org-scoped assignments.
	OrgSubtree(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error)
	// Policies returns active policies applicable to perm on resource type rt,
	// ordered by priority ascending (evaluator applies deny-first anyway).
	Policies(ctx context.Context, db database.TenantDB, a Actor, perm, rt string) ([]Policy, error)
	// ResourceOrg returns the org id owning a resource (for org-scope checks on
	// a resource target), or zero if none/unknown.
	ResourceOrg(ctx context.Context, db database.TenantDB, ref resource.Ref) (uuid.UUID, error)
}

// AuditSink records authorization denials that must be audited (sensitive
// permission denials and explicit policy denies — 01 §3 step 7). The durable
// audit_logs writer lands in Phase 6; Phase 4 wires this port and a capturing
// test fake so the "denials audited" guarantee is testable now.
type AuditSink interface {
	AuthzDenial(ctx context.Context, a Actor, perm string, t Target, reason string)
}
