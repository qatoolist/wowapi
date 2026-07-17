// Package rules is wowapi's rule/configuration engine: modules register rule
// points (a key, a RuleValueSchema'd value, a default, allowed scopes, and
// whether changes require approval); values are stored as versioned rows with
// temporal validity; and resolution picks the most specific active value for a
// (tenant, org, at) — org-ancestry → tenant → platform → code default. Versions
// are immutable (never mutated, only superseded), so any historical `at`
// resolves deterministically. Contract: blueprint 02 §2.
//
// Rule points are the ONLY sanctioned place for values that must change without
// a deploy (feature flags, tenant overrides); framework config holds only their
// platform defaults (12 §6).
//
// # RuleValueSchema
//
// A Point's ValueSchema is NOT JSON Schema — it never was a full
// implementation, and as of B3 the contract is corrected to say so plainly.
// It is RuleValueSchema: a small, closed, framework-specific grammar (ratified
// Decision 2 — a strict limited grammar, no JSON-Schema library dependency)
// recognizing exactly these top-level keywords:
//
//   - "type": one of integer/number/string/boolean/object/array/null (any
//     other value is rejected at Register — B3 defect 1);
//   - "enum": a JSON array of allowed literal values;
//   - "minimum" / "maximum" / "exclusiveMinimum" / "exclusiveMaximum": numeric bounds;
//   - "minLength" / "maxLength" / "pattern" (RE2): string constraints;
//   - "minItems" / "maxItems": array length bounds;
//   - "required": a shallow presence check for object keys (NOT recursive
//     per-property validation — there is no nested "properties" schema).
//
// Any keyword outside this list — "multipleOf", "additionalProperties",
// "items" sub-schemas, "properties", "patternProperties", etc — is REJECTED
// at Register, not silently ignored (B3 defect 2: json.Unmarshal into an
// unexported struct used to drop unrecognized keys, so a schema author could
// write a constraint the framework never enforced without any error). A rule
// point needing per-property typing should declare separate top-level rule
// points instead of one object-shaped point with nested constraints.
//
// Register also validates that a Point's Default conforms to its own
// ValueSchema (B3 defect 3) — a schema that can't even validate its own
// default is broken by construction and must never boot. Resolver.Resolve
// re-validates the winning STORED value against the point's CURRENT schema
// before returning it (B3 defect 4): a value that conformed to an earlier,
// looser schema can drift out of conformance after a module upgrade
// tightens the schema, and Resolve surfaces that as an error rather than
// silently handing back a non-conforming value.
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
	Key    string
	Module string
	// ValueSchema is a RuleValueSchema document (see the package doc above) —
	// a small closed grammar, NOT JSON Schema. Validated at Register (schema
	// well-formedness + Default conformance), at Propose (write time), and at
	// Resolve (defense in depth against post-write schema drift).
	ValueSchema      json.RawMessage
	Default          json.RawMessage // compiled default value; must conform to ValueSchema (checked at Register)
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
	sealed bool
}

// NewRegistry returns an empty rule registry.
func NewRegistry() *Registry { return &Registry{points: map[string]Point{}} }

// Seal freezes the registry once boot validation completes: any later Register
// panics rather than silently adding a rule point the boot gates never saw
// (closure review 2026-07-17, F-10).
func (r *Registry) Seal() { r.sealed = true }

// Register adds a rule point. Malformed keys, a module-prefix mismatch, a
// missing schema/default, a schema that is malformed or names an unknown
// type/keyword outside the RuleValueSchema grammar (B3 defect 1/2), a default
// that violates its own schema (B3 defect 3), or a duplicate are recorded as
// errors surfaced by Err() — the boot-error-accumulation gate (app/boot.go
// calls k.Rules.Err()) turns any of these into a boot failure, so a
// silently-unenforced or self-contradictory rule point can never go live.
func (r *Registry) Register(module string, p Point) {
	if r.sealed {
		panic("rules: rule-point registration after boot: the extension model is sealed")
	}
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
	if err := validateAgainstSchema(p.ValueSchema, p.Default); err != nil {
		r.errf("rule point %s: default violates its own value_schema: %s", p.Key, err.Error())
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

// Points returns a COPY of the registered points keyed by key — a snapshot,
// never the registry's backing map (closure review 2026-07-17, F-10).
func (r *Registry) Points() map[string]Point {
	out := make(map[string]Point, len(r.points))
	for k, v := range r.points {
		out[k] = v
	}
	return out
}

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
