// Package authz is wowapi's authorization kernel: a deny-by-default evaluator
// that layers RBAC (role→permission assignments), ReBAC (relationship-derived
// grants), and ABAC (attribute policies, deny-first) exactly as specified in
// blueprint 01 §3. The evaluator is pure over a Store port so it is unit-
// testable without a database; the Postgres-backed Store lives alongside.
//
// Two invariants are structural, not configurable:
//   - deny by default: no matching grant → denied;
//   - permission must be registered: evaluating an unregistered permission is a
//     programming error (surfaced at boot when routes/permissions register),
//     never a silent runtime allow.
package authz

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// ActorKind enumerates who is acting.
type ActorKind string

const (
	ActorUser    ActorKind = "user"
	ActorSystem  ActorKind = "system"
	ActorWebhook ActorKind = "webhook"
)

// Actor is the authenticated principal for an authorization decision. For a
// human it carries the user and their active capacity in the tenant; for a
// non-human it carries a system identifier.
type Actor struct {
	Kind       ActorKind
	UserID     uuid.UUID
	CapacityID uuid.UUID // zero for system/webhook actors
	System     string    // "outbox-relay", "webhook:payments"
	TenantID   uuid.UUID
	// ImpersonatorUserID is set when a support actor impersonates a user; both
	// identities are audited and impersonation is policy-restricted (01 §3).
	ImpersonatorUserID uuid.UUID
	// BreakGlass marks an actor operating under an activated break-glass grant;
	// every decision it produces is audited and bannered.
	BreakGlass bool
	// Scopes is the explicit permission set of a machine principal (API key /
	// service principal). It is meaningful only for ActorSystem actors: a scope
	// authorizes like an RBAC grant but remains subject to ABAC deny policies.
	// Human and internal-system actors leave it empty and are unaffected.
	Scopes []string
	// AMR is the authentication-methods-references set surfaced from the IdP token
	// (e.g. ["pwd","mfa","otp"]). It drives step-up enforcement and the env.mfa
	// ABAC attribute (roadmap S3).
	AMR []string
}

// ScopeKind is the granularity of an authorization target.
type ScopeKind string

const (
	ScopeTenant       ScopeKind = "tenant"
	ScopeOrg          ScopeKind = "org"
	ScopeResourceType ScopeKind = "resource_type"
	ScopeResource     ScopeKind = "resource"
)

// Target is what an actor wants to act upon.
type Target struct {
	Scope    ScopeKind
	OrgID    uuid.UUID
	Resource resource.Ref
}

// Decision is the outcome of Evaluate. Reason names the matched grant/policy
// for audit ("role:requests.org.approver", "rel:core.owner_of",
// "policy:deny_locked"); it is safe to log.
type Decision struct {
	Allowed   bool
	Reason    string
	PolicyIDs []uuid.UUID
	// StepUpRequired is set when the actor would be allowed but the permission
	// demands an elevated auth factor the actor has not satisfied (roadmap S3).
	// The HTTP gate turns this into a re-authentication challenge rather than a
	// flat 403.
	StepUpRequired bool
}

// ListFilter is the structured constraint Filter returns so list queries embed
// authorization in SQL instead of loading-then-filtering. An empty filter with
// All=true means unrestricted (a tenant-wide grant); All=false with no
// constraints means "deny all" (no rows visible).
type ListFilter struct {
	All bool // true → no record-level restriction (tenant RLS still applies)
	// OrgIDs, when non-nil, restricts to resources in these orgs.
	OrgIDs []uuid.UUID
	// ResourceIDs, when non-nil, restricts to these specific resource ids
	// (e.g. relationship-derived visibility).
	ResourceIDs []uuid.UUID
}

// Evaluator is the authorization decision port modules receive. Both methods
// take the caller's tenant TenantDB so every authorization read runs in the
// SAME transaction (and MVCC snapshot) as the request's business writes — an
// authz check right after a mirror-row write must see that write, and the hot
// path must not open extra connections (review finding ARCH-36).
type Evaluator interface {
	// Evaluate returns whether actor a may exercise permission perm on target t.
	Evaluate(ctx context.Context, db database.TenantDB, a Actor, perm string, t Target) (Decision, error)
	// Filter returns the record-level constraint for listing resources of type
	// rt that a may exercise perm on.
	Filter(ctx context.Context, db database.TenantDB, a Actor, perm string, rt string) (ListFilter, error)
}

// RelationshipChecker answers ReBAC questions: does subject stand in relation
// relType to obj at time at. Implemented by kernel/relationship; runs on the
// caller's tenant tx.
type RelationshipChecker interface {
	Has(ctx context.Context, db database.TenantDB, subject Actor, relType string, obj resource.Ref, at time.Time) (bool, error)
}
