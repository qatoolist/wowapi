package authz

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// PolicyEngine evaluates an ABAC policy's conditions against the attribute bag.
// Implemented by kernel/policy; injected so the evaluator stays free of the
// condition-matching detail.
type PolicyEngine interface {
	// Matches reports whether every condition holds for the given attributes.
	Matches(conds []Condition, attrs map[string]any) (bool, error)
}

// engine is the concrete deny-by-default evaluator.
type engine struct {
	store    Store
	rels     RelationshipChecker
	registry *Registry
	policies PolicyEngine
	audit    AuditSink
	now      func() time.Time
}

// Options configures New. Store, Registry, and PolicyEngine are required;
// RelationshipChecker and AuditSink may be nil (ReBAC/denial-audit disabled).
type Options struct {
	Store         Store
	Relationships RelationshipChecker
	Registry      *Registry
	Policies      PolicyEngine
	Audit         AuditSink
	Now           func() time.Time
}

// New builds an Evaluator. It panics on missing required collaborators — that
// is a wiring error at composition, not a runtime condition.
func New(o Options) Evaluator {
	if o.Store == nil || o.Registry == nil || o.Policies == nil {
		panic("authz.New: Store, Registry, and Policies are required")
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}
	return &engine{
		store: o.Store, rels: o.Relationships, registry: o.Registry,
		policies: o.Policies, audit: o.Audit, now: now,
	}
}

// Evaluate implements the layered algorithm (01 §3): deny-by-default, then RBAC,
// then ReBAC, then ABAC deny-first (a matching deny is absolute; allow policies
// may add a narrow grant). A permission absent from the registry is a
// programming error, never a silent allow.
func (e *engine) Evaluate(ctx context.Context, db database.TenantDB, a Actor, perm string, t Target) (Decision, error) {
	pdef, ok := e.registry.Get(perm)
	if !ok {
		return Decision{}, kerr.E(kerr.KindInternal, "unregistered_permission",
			"authorization checked an unregistered permission: "+perm)
	}

	now := e.now()
	decision := Decision{Allowed: false, Reason: "default_deny"}

	// --- RBAC: any active assignment whose role grants perm and whose scope
	// covers the target is a candidate ALLOW. ---
	assignments, err := e.store.ActiveAssignments(ctx, db, a, now)
	if err != nil {
		return Decision{}, kerr.Wrapf(err, "authz.Evaluate", "load assignments")
	}
	// Resolve the target's org ancestry once for org-scope coverage.
	var targetOrgAncestors []uuid.UUID
	orgID, err := e.targetOrg(ctx, db, t)
	if err != nil {
		return Decision{}, err
	}
	if orgID != uuid.Nil {
		targetOrgAncestors, err = e.store.OrgAncestors(ctx, db, orgID)
		if err != nil {
			return Decision{}, kerr.Wrapf(err, "authz.Evaluate", "resolve org ancestry")
		}
	}
	for _, asg := range assignments {
		if asg.grants(perm) && covers(asg, t, targetOrgAncestors) {
			decision.Allowed = true
			decision.Reason = "role:" + asg.RoleKey
			break
		}
	}

	// --- Machine scope: a service principal / API key (ActorSystem with an
	// explicit scope set) is authorized by its scopes, which act like an RBAC
	// grant — still subject to the ABAC deny pass below (a deny policy overrides a
	// scope). Internal system actors carry no scopes and are unaffected, so this
	// never widens their authority. Deny-by-default holds: allow only on an
	// explicit scope match. ---
	if !decision.Allowed && a.Kind == ActorSystem && slices.Contains(a.Scopes, perm) {
		decision.Allowed = true
		decision.Reason = "machine_scope"
	}

	// --- ReBAC: if RBAC did not allow and the permission declares a
	// granted_via relationship, check it against the resource target. ---
	if !decision.Allowed && pdef.GrantedVia != "" && e.rels != nil && !t.Resource.IsZero() {
		has, err := e.rels.Has(ctx, db, a, pdef.GrantedVia, t.Resource, now)
		if err != nil {
			return Decision{}, kerr.Wrapf(err, "authz.Evaluate", "relationship check")
		}
		if has {
			decision.Allowed = true
			decision.Reason = "rel:" + pdef.GrantedVia
		}
	}

	// --- ABAC: policies, deny-first. A matching deny is absolute; an allow
	// policy can grant when nothing else did. ---
	pols, err := e.store.Policies(ctx, db, a, perm, t.Resource.Type)
	if err != nil {
		return Decision{}, kerr.Wrapf(err, "authz.Evaluate", "load policies")
	}
	if len(pols) > 0 {
		attrs := e.attributes(a, t, now)
		// Deny policies first (highest authority), by priority.
		ordered := slices.Clone(pols)
		slices.SortStableFunc(ordered, func(x, y Policy) int { return x.Priority - y.Priority })
		for _, p := range ordered {
			if p.Effect != EffectDeny {
				continue
			}
			// A deny policy that references an attribute the evaluator cannot
			// resolve must FAIL CLOSED (deny), never silently not-match and let
			// a prior allow stand (review finding SEC-25). We check the
			// attribute bag has every referenced key before matching.
			if missing := unresolved(p.Conditions, attrs); missing != "" {
				decision.Allowed = false
				decision.Reason = "policy:" + p.Key + " (unresolved:" + missing + ")"
				decision.PolicyIDs = append(decision.PolicyIDs, p.ID)
				e.maybeAudit(ctx, a, perm, t, pdef, decision)
				return decision, nil
			}
			match, err := e.policies.Matches(p.Conditions, attrs)
			if err != nil {
				return Decision{}, kerr.Wrapf(err, "authz.Evaluate", "policy eval")
			}
			if match {
				decision.Allowed = false
				decision.Reason = "policy:" + p.Key
				decision.PolicyIDs = append(decision.PolicyIDs, p.ID)
				e.maybeAudit(ctx, a, perm, t, pdef, decision)
				return decision, nil
			}
		}
		if !decision.Allowed {
			for _, p := range ordered {
				if p.Effect != EffectAllow {
					continue
				}
				match, err := e.policies.Matches(p.Conditions, attrs)
				if err != nil {
					return Decision{}, kerr.Wrapf(err, "authz.Evaluate", "policy eval")
				}
				if match {
					decision.Allowed = true
					decision.Reason = "policy:" + p.Key
					decision.PolicyIDs = append(decision.PolicyIDs, p.ID)
					break
				}
			}
		}
	}

	e.maybeAudit(ctx, a, perm, t, pdef, decision)
	return decision, nil
}

// maybeAudit fires the audit sink for a denial of a sensitive permission or any
// explicit policy deny, and for every break-glass/impersonated decision.
func (e *engine) maybeAudit(ctx context.Context, a Actor, perm string, t Target, pdef Permission, d Decision) {
	if e.audit == nil {
		return
	}
	explicitDeny := !d.Allowed && len(d.PolicyIDs) > 0
	if (!d.Allowed && pdef.Sensitive) || explicitDeny || a.BreakGlass || a.ImpersonatorUserID != uuid.Nil {
		e.audit.AuthzDenial(ctx, a, perm, t, d.Reason)
	}
}

// targetOrg resolves the org id relevant to the target: the explicit OrgID for
// org scope, or the owning org of a resource target.
func (e *engine) targetOrg(ctx context.Context, db database.TenantDB, t Target) (uuid.UUID, error) {
	if t.OrgID != uuid.Nil {
		return t.OrgID, nil
	}
	if !t.Resource.IsZero() {
		org, err := e.store.ResourceOrg(ctx, db, t.Resource)
		if err != nil {
			return uuid.Nil, kerr.Wrapf(err, "authz.Evaluate", "resolve resource org")
		}
		return org, nil
	}
	return uuid.Nil, nil
}

// unresolved returns the first condition attribute absent from the bag, or ""
// if all are present. Used to fail deny policies closed when the evaluator
// cannot supply an attribute a deny gates on (SEC-25).
func unresolved(conds []Condition, attrs map[string]any) string {
	for _, c := range conds {
		if _, ok := attrs[c.Attribute]; !ok {
			return c.Attribute
		}
	}
	return ""
}

// covers reports whether an assignment's scope covers the target.
func covers(a Assignment, t Target, targetOrgAncestors []uuid.UUID) bool {
	switch a.ScopeKind {
	case ScopeTenant:
		// Tenant-wide grant covers everything in the tenant.
		return true
	case ScopeOrg:
		// Grant at org O covers O and its descendants: O covers the target iff
		// O is an ancestor-or-self of the target's org.
		return slices.Contains(targetOrgAncestors, a.ScopeID)
	case ScopeResourceType:
		// Covers any target of the matching resource type. A NULL scope_type
		// (a.ScopeType == "") must never match — otherwise it collides with a
		// typeless target ("" == "") and over-grants (review finding SEC-26).
		return a.ScopeType != "" && t.Resource.Type == a.ScopeType
	case ScopeResource:
		// Covers only the exact, non-nil resource. A NULL scope_id must not
		// match a typeless/nil target (review finding SEC-29).
		return a.ScopeID != uuid.Nil && t.Resource.ID == a.ScopeID
	}
	return false
}

// attributes builds the ABAC bag. Resource-attribute enrichment (status, etc.)
// is added by the store-backed evaluator in later phases; Phase 4 exposes the
// actor/env/target identity attributes policies commonly gate on.
func (e *engine) attributes(a Actor, t Target, now time.Time) map[string]any {
	return map[string]any{
		"actor.user_id":       a.UserID.String(),
		"actor.capacity_id":   a.CapacityID.String(),
		"actor.kind":          string(a.Kind),
		"actor.impersonating": a.ImpersonatorUserID != uuid.Nil,
		"actor.break_glass":   a.BreakGlass,
		"env.time":            now.Format(time.RFC3339),
		"env.hour":            now.Hour(),
		"resource.type":       t.Resource.Type,
		"resource.id":         t.Resource.ID.String(),
		"target.scope":        string(t.Scope),
	}
}

// Filter returns the record-level constraint for listing resources of type rt.
// Deny-by-default: with no covering assignment the filter denies all rows.
//
// SCOPE (Phase 4): Filter covers RBAC scopes only. It does NOT apply ABAC deny
// policies (SEC-28) or relationship-derived visibility (ARCH-37). Consequences,
// closed when list endpoints ship in Phase 5:
//   - a per-row deny policy is NOT reflected in the returned filter, so a list
//     handler MUST still run per-row Evaluate for permissions that have active
//     deny policies (or Filter must gain a deny-aware seam);
//   - a resource visible only via a granted_via relationship is NOT in the
//     filter, so relationship-only users get an empty list until the
//     Store.RelationshipResourceIDs seam lands.
//
// Until then Filter is safe (it under-grants, never over-grants) but incomplete.
func (e *engine) Filter(ctx context.Context, db database.TenantDB, a Actor, perm, rt string) (ListFilter, error) {
	if !e.registry.Has(perm) {
		return ListFilter{}, kerr.E(kerr.KindInternal, "unregistered_permission",
			"authorization filtered on an unregistered permission: "+perm)
	}
	now := e.now()
	assignments, err := e.store.ActiveAssignments(ctx, db, a, now)
	if err != nil {
		return ListFilter{}, kerr.Wrapf(err, "authz.Filter", "load assignments")
	}

	var orgIDs []uuid.UUID
	var resourceIDs []uuid.UUID
	for _, asg := range assignments {
		if !asg.grants(perm) {
			continue
		}
		switch asg.ScopeKind {
		case ScopeTenant:
			return ListFilter{All: true}, nil // unrestricted within the tenant
		case ScopeResourceType:
			if asg.ScopeType == rt {
				return ListFilter{All: true}, nil
			}
		case ScopeOrg:
			subtree, err := e.store.OrgSubtree(ctx, db, asg.ScopeID)
			if err != nil {
				return ListFilter{}, kerr.Wrapf(err, "authz.Filter", "org subtree")
			}
			orgIDs = append(orgIDs, subtree...)
		case ScopeResource:
			resourceIDs = append(resourceIDs, asg.ScopeID)
		}
	}

	// ReBAC visibility: relationship-derived resource ids are added by the
	// store-backed evaluator when the permission is granted_via a relationship
	// (a dedicated Store method lands with the pg store). Phase 4's pure
	// evaluator surfaces RBAC-derived constraints; relationship expansion is
	// wired in the pg Store.Filter path.
	if len(orgIDs) == 0 && len(resourceIDs) == 0 {
		return ListFilter{All: false}, nil // deny all: nothing visible
	}
	return ListFilter{OrgIDs: dedupe(orgIDs), ResourceIDs: dedupe(resourceIDs)}, nil
}

func dedupe(in []uuid.UUID) []uuid.UUID {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(in))
	out := in[:0]
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
