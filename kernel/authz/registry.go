package authz

import (
	"regexp"
	"sort"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// permKeyRE constrains permission keys to "module.resource.action".
var permKeyRE = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// verbs is the closed action set (01 §3). A permission's action segment must be
// one of these, so the permission surface stays a small, auditable vocabulary.
var verbs = map[string]bool{
	"create": true, "read": true, "list": true, "update": true,
	"deactivate": true, "restore": true, "approve": true, "reject": true,
	"assign": true, "export": true, "admin": true, "ingest": true, "activate": true,
}

// Permission is a registered permission. GrantedVia, when set, declares the
// ReBAC rule "this permission is granted on a resource target to any actor that
// has the named relationship to it" (01 §3 step 4).
type Permission struct {
	Key        string
	Sensitive  bool
	GrantedVia string // relationship type key, or "" for none
	// StepUp requires the actor to have satisfied an elevated authentication
	// factor (MFA) for this permission: an otherwise-allowed decision becomes a
	// step-up challenge when the actor's AMR carries no strong factor (roadmap
	// S3). MFA itself is the IdP's job; this gates on the surfaced amr claim.
	// This is the persisted shorthand (permissions.step_up) — "require ANY
	// factor from the deployment's configured default strong-factor set".
	StepUp bool
	// StepUpPolicy, when non-nil, REPLACES the default-set behavior of StepUp
	// with a permission-specific requirement (e.g. "require hwk specifically").
	// It is declared by a seed's richer step_up form (kernel/seeds) and lives
	// only in this in-memory, boot-populated registry — it is NOT persisted
	// (permissions.step_up remains a plain bool; see kernel/seeds doc comment
	// on PermissionSeed.StepUpAMR for the rationale). A permission with
	// StepUpPolicy set is treated as StepUp-gated regardless of the StepUp bool.
	StepUpPolicy *StepUpPolicy
}

// StepUpPolicy is a permission-specific step-up requirement: the actor must
// present at least one AMR value from RequiredAMR (any-of — the usual
// step-up semantic: any single elevated factor satisfies the gate, factors
// are not required in combination). Challenge names the factor/hint the HTTP
// gate advertises in WWW-Authenticate (e.g. "hwk", "mfa").
//
// Scope (Decision 4, framework-engineering-backlog B8, archived to the wowapi2 doc archive): this is AMR-only. The
// production IdP's ability to reliably emit `auth_time` could not be confirmed
// from the codebase, so no MaxAge/freshness field exists here. The struct is
// shaped so a MaxAge *time.Duration could be added later as an additive field
// without breaking existing callers — but that is explicitly out of scope now.
type StepUpPolicy struct {
	// RequiredAMR is the set of AMR values that satisfy this permission's
	// step-up gate; the actor needs ANY ONE of them. Empty means "fall back to
	// the deployment's configured default strong-factor set" (the StepUp bool
	// shorthand's behavior).
	RequiredAMR []string
	// Challenge is the factor/hint advertised in the step-up challenge's
	// WWW-Authenticate header (e.g. `step_up="hwk"`). Empty falls back to the
	// deployment's default challenge hint.
	Challenge string
}

// Registry is the boot-time permission catalog. Evaluating a permission absent
// from the registry is a programming error, so registration is validated and
// its Err() must gate boot — an unknown permission can never silently allow.
type Registry struct {
	perms map[string]Permission
	errs  []error
}

// NewRegistry returns an empty permission registry.
func NewRegistry() *Registry { return &Registry{perms: map[string]Permission{}} }

// Register adds a permission. Malformed keys, unknown action verbs, and
// duplicates are recorded as errors surfaced by Err().
func (r *Registry) Register(p Permission) {
	if !permKeyRE.MatchString(p.Key) {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_permission",
			"permission key must be module.resource.action: "+p.Key))
		return
	}
	action := p.Key[lastDot(p.Key)+1:]
	if !verbs[action] {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_permission",
			"permission action %q is not in the closed verb set: "+p.Key))
		return
	}
	if _, dup := r.perms[p.Key]; dup {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "duplicate_permission",
			"permission registered more than once: "+p.Key))
		return
	}
	r.perms[p.Key] = p
}

// Has reports whether key is registered.
func (r *Registry) Has(key string) bool { _, ok := r.perms[key]; return ok }

// Get returns the permission definition.
func (r *Registry) Get(key string) (Permission, bool) { p, ok := r.perms[key]; return p, ok }

// Keys returns all registered permission keys, sorted (for seed sync + tests).
func (r *Registry) Keys() []string {
	out := make([]string, 0, len(r.perms))
	for k := range r.perms {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Err returns accumulated registration errors joined, or nil.
func (r *Registry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	msgs := make([]string, len(r.errs))
	for i, e := range r.errs {
		msgs[i] = e.Error()
	}
	joined := msgs[0]
	for i := 1; i < len(msgs); i++ {
		joined += "; " + msgs[i]
	}
	return kerr.E(kerr.KindInternal, "permission_registration_failed", "permission registration failed: "+joined)
}

func lastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}
