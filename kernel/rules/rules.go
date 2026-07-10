// Package rules is wowapi's rule/configuration engine: modules register rule
// points (a key, a JSON-Schema'd value, a default, allowed scopes, and whether
// changes require approval); values are stored as versioned rows with temporal
// validity; and resolution picks the most specific active value for a
// (tenant, org, at) — org-ancestry → tenant → platform → code default. Versions
// are immutable (never mutated, only superseded), so any historical `at`
// resolves deterministically. Contract: blueprint 02 §2.
//
// Rule points are the ONLY sanctioned place for values that must change without
// a deploy (feature flags, tenant overrides); framework config holds only their
// platform defaults (12 §6).
package rules

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// ScopeKind is the level a rule value applies at.
type ScopeKind string

const (
	ScopePlatform ScopeKind = "platform"
	ScopeTenant   ScopeKind = "tenant"
	ScopeOrg      ScopeKind = "org"
)

// keyRE constrains rule keys to module.area.name.
var keyRE = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// Point is a registered rule point: the schema + default + policy for a key.
type Point struct {
	Key              string
	Module           string
	ValueSchema      json.RawMessage // JSON Schema (validated at write + resolve)
	Default          json.RawMessage // compiled default value
	AllowedScopes    []ScopeKind
	RequiresApproval bool
	Description      string
}

// allowsScope reports whether a value may be set at the given scope.
func (p Point) allowsScope(s ScopeKind) bool {
	for _, a := range p.AllowedScopes {
		if a == s {
			return true
		}
	}
	return false
}

// Registry collects rule points during module registration; SyncDefinitions
// persists it to rule_definitions (the generated migrate main calls it right
// after seed sync, GAP-007), and it is consulted by the resolver for defaults
// + policy.
type Registry struct {
	points map[string]Point
	errs   []error
}

// NewRegistry returns an empty rule registry.
func NewRegistry() *Registry { return &Registry{points: map[string]Point{}} }

// Register adds a rule point. Malformed keys, a module-prefix mismatch, a
// missing schema/default, or a duplicate are recorded as errors surfaced by
// Err().
func (r *Registry) Register(module string, p Point) {
	if !keyRE.MatchString(p.Key) {
		r.errf("rule key must be module.area.name: %s", p.Key)
		return
	}
	if prefix := module + "."; len(p.Key) <= len(prefix) || p.Key[:len(prefix)] != prefix {
		r.errf("module %s may not register rule point %s", module, p.Key)
		return
	}
	if len(p.ValueSchema) == 0 || len(p.Default) == 0 {
		r.errf("rule point %s requires a value_schema and a default", p.Key)
		return
	}
	if _, dup := r.points[p.Key]; dup {
		r.errf("rule point registered more than once: %s", p.Key)
		return
	}
	p.Module = module
	if len(p.AllowedScopes) == 0 {
		p.AllowedScopes = []ScopeKind{ScopePlatform, ScopeTenant, ScopeOrg}
	}
	r.points[p.Key] = p
}

// Get returns the registered point.
func (r *Registry) Get(key string) (Point, bool) { p, ok := r.points[key]; return p, ok }

// Keys returns registered keys, sorted.
func (r *Registry) Keys() []string {
	out := make([]string, 0, len(r.points))
	for k := range r.points {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Points returns the registered points keyed by key.
func (r *Registry) Points() map[string]Point { return r.points }

func (r *Registry) errf(format string, args ...any) {
	r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_rule", fmt.Sprintf(format, args...)))
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
	return kerr.E(kerr.KindInternal, "rule_registration_failed", "rule point registration failed: "+joined)
}
