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
