package lifecycle

import (
	"fmt"
	"sort"
)

// Violation is one lint finding. Class identifies which failure category
// (a-e, matching the backlog B9 acceptance list) so callers/tests can filter
// or count by class; Provides names the offending descriptor; Message is the
// human-readable detail.
type Violation struct {
	Class    string
	Provides string
	Message  string
}

// Violation classes, matching backlog B9's acceptance list verbatim so a
// reader can cross-reference the backlog item to the check that enforces it.
const (
	ClassScopeLeak       = "scope_leak"        // (a) process-scoped depends on request-scoped
	ClassRawPool         = "raw_pool"          // (b) module receives a raw pool instead of TxManager
	ClassTenantEscape    = "tenant_escape"     // (c) tenant-scoped service escapes its transaction
	ClassMigrateInAPI    = "migrate_in_api"    // (d) migrate-only service wired into API runtime
	ClassMissingProvider = "missing_provider"  // (e) declared Requires has no matching Provides
	ClassCycle           = "cycle"             // (e) dependency cycle
	ClassDuplicate       = "duplicate_provide" // manifest hygiene: two descriptors with the same Provides
	ClassInvalidScope    = "invalid_scope"     // manifest hygiene: Scope is not one of the five recognized values
)

func (v Violation) String() string {
	return fmt.Sprintf("[%s] %s: %s", v.Class, v.Provides, v.Message)
}

// Lint runs every check against m and returns ALL violations found, sorted
// for deterministic output — mirroring the errs-accumulate pattern used by
// kernel/config.Framework.Validate and kernel/rules.Registry.Err (collect
// everything, never stop at the first problem).
func Lint(m Manifest) []Violation {
	var out []Violation
	out = append(out, checkInvalidScope(m)...)
	out = append(out, checkDuplicateProvides(m)...)
	out = append(out, checkMissingProviders(m)...)
	out = append(out, checkCycles(m)...)
	out = append(out, checkNoNarrowerDependency(m)...)
	out = append(out, checkRawPool(m)...)
	out = append(out, checkTenantEscape(m)...)
	out = append(out, checkMigrateInAPIRuntime(m)...)

	sort.Slice(out, func(i, j int) bool {
		if out[i].Provides != out[j].Provides {
			return out[i].Provides < out[j].Provides
		}
		if out[i].Class != out[j].Class {
			return out[i].Class < out[j].Class
		}
		return out[i].Message < out[j].Message
	})
	return out
}

// checkInvalidScope flags any descriptor whose Scope is not one of the five
// recognized values — a typo here would silently disable every scope-aware
// check below, so it is validated first and explicitly.
func checkInvalidScope(m Manifest) []Violation {
	var out []Violation
	for _, d := range m.Descriptors {
		if !d.Scope.valid() {
			out = append(out, Violation{
				Class: ClassInvalidScope, Provides: d.Provides,
				Message: fmt.Sprintf("scope %q is not one of process|request|tenant_tx|job|migrate", string(d.Scope)),
			})
		}
	}
	return out
}

// checkDuplicateProvides flags two descriptors declaring the same Provides
// name — ByProvides() would silently pick one, hiding the other from every
// other check, so this must run (and be visible) independently of it.
func checkDuplicateProvides(m Manifest) []Violation {
	seen := make(map[string]int, len(m.Descriptors))
	var out []Violation
	for _, d := range m.Descriptors {
		seen[d.Provides]++
	}
	for name, n := range seen {
		if n > 1 {
			out = append(out, Violation{
				Class: ClassDuplicate, Provides: name,
				Message: fmt.Sprintf("declared by %d descriptors — Provides names must be unique", n),
			})
		}
	}
	return out
}

// checkMissingProviders is class (e): every declared Requires must match
// another descriptor's Provides.
func checkMissingProviders(m Manifest) []Violation {
	idx := m.ByProvides()
	var out []Violation
	for _, d := range m.Descriptors {
		for _, req := range d.Requires {
			if _, ok := idx[req]; !ok {
				out = append(out, Violation{
					Class: ClassMissingProvider, Provides: d.Provides,
					Message: fmt.Sprintf("requires %q, which no descriptor provides", req),
				})
			}
		}
	}
	return out
}

// checkCycles is class (e): no dependency cycle. DFS with a recursion stack;
// missing providers (already reported by checkMissingProviders) are simply
// skipped here rather than double-reported.
func checkCycles(m Manifest) []Violation {
	idx := m.ByProvides()
	const (
		white = 0 // unvisited
		gray  = 1 // on the current DFS stack
		black = 2 // fully explored
	)
	color := make(map[string]int, len(m.Descriptors))
	var out []Violation
	var stack []string

	var visit func(name string)
	visit = func(name string) {
		if color[name] == black {
			return
		}
		if color[name] == gray {
			// Found a cycle: report it once, anchored at the descriptor that
			// closes the loop, with the cyclic path for readability.
			cyclePath := append(append([]string(nil), stack...), name)
			out = append(out, Violation{
				Class: ClassCycle, Provides: name,
				Message: fmt.Sprintf("dependency cycle: %s", joinCycle(cyclePath)),
			})
			return
		}
		color[name] = gray
		stack = append(stack, name)
		d, ok := idx[name]
		if ok {
			for _, req := range d.Requires {
				if _, ok := idx[req]; ok {
					visit(req)
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[name] = black
	}

	// Deterministic traversal order.
	names := make([]string, 0, len(m.Descriptors))
	for _, d := range m.Descriptors {
		names = append(names, d.Provides)
	}
	sort.Strings(names)
	for _, n := range names {
		if color[n] == white {
			visit(n)
		}
	}
	return out
}

func joinCycle(path []string) string {
	s := ""
	for i, p := range path {
		if i > 0 {
			s += " -> "
		}
		s += p
	}
	return s
}

// checkNoNarrowerDependency is class (a) generalized: a descriptor must not
// Require something with a strictly narrower (shorter-lived) scope than
// itself, per Scope.rank (migrate < process < job < request < tenant_tx). A
// process-scoped service depending on request-scoped state is the specific
// case the backlog calls out; the same reasoning applies to any wider scope
// depending on a narrower one (e.g. process depending on tenant_tx, or job
// depending on request).
func checkNoNarrowerDependency(m Manifest) []Violation {
	idx := m.ByProvides()
	var out []Violation
	for _, d := range m.Descriptors {
		if !d.Scope.valid() {
			continue // already reported by checkInvalidScope
		}
		for _, req := range d.Requires {
			dep, ok := idx[req]
			if !ok || !dep.Scope.valid() {
				continue // missing/invalid already reported elsewhere
			}
			if dep.Scope.rank() > d.Scope.rank() {
				out = append(out, Violation{
					Class: ClassScopeLeak, Provides: d.Provides,
					Message: fmt.Sprintf("scope %q depends on %q (scope %q), which outlives it less than its own scope — a %s-scoped service must not depend on narrower-lived %s state",
						d.Scope, req, dep.Scope, d.Scope, dep.Scope),
				})
			}
		}
	}
	return out
}

// checkRawPool is class (b): no module-runtime-facing descriptor may be a raw
// pool. RawPool descriptors are legitimate ONLY as kernel-internal
// (non-APIRuntime) building blocks that a TxManager wraps — e.g.
// "kernel.Pool" feeding "kernel.Tx" — never as something a module receives
// directly.
func checkRawPool(m Manifest) []Violation {
	var out []Violation
	for _, d := range m.Descriptors {
		if d.RawPool && d.APIRuntime {
			out = append(out, Violation{
				Class: ClassRawPool, Provides: d.Provides,
				Message: "a raw pool is wired into API/module runtime — modules must receive a database.TxManager, never a *pgxpool.Pool",
			})
		}
	}
	return out
}

// checkTenantEscape is class (c): a tenant_tx-scoped descriptor must never be
// Required by something wider-scoped (process/request/job/migrate), because
// that would let the tenant-bound value (a database.TenantDB) be retained
// past the transaction callback that produced it. This is the mirror image of
// checkNoNarrowerDependency's general rule, stated explicitly for tenant_tx
// because "escaping its transaction" is the specific failure class B9 calls
// out and deserves its own violation class/message for operators.
func checkTenantEscape(m Manifest) []Violation {
	idx := m.ByProvides()
	var out []Violation
	for _, d := range m.Descriptors {
		if d.Scope != ScopeTenantTx {
			continue
		}
		for _, other := range m.Descriptors {
			if other.Provides == d.Provides {
				continue
			}
			for _, req := range other.Requires {
				if req != d.Provides {
					continue
				}
				if !other.Scope.valid() {
					continue
				}
				if other.Scope != ScopeTenantTx {
					out = append(out, Violation{
						Class: ClassTenantEscape, Provides: d.Provides,
						Message: fmt.Sprintf("tenant_tx-scoped value is depended on by %q (scope %q) — a tenant-bound value must not escape its transaction callback",
							other.Provides, other.Scope),
					})
				}
			}
		}
	}
	_ = idx // idx kept for symmetry/future use; current check needs only Requires scan
	return out
}

// checkMigrateInAPIRuntime is class (d): a migrate-scoped descriptor must
// never have APIRuntime set (reachable from module.Context / api/worker
// wiring). Migrate-only capabilities (the privileged DDL connection, the
// migrate process's SkipRLSEnforcementCheck opt-out) exist ONLY in the
// migrate binary's boot path.
func checkMigrateInAPIRuntime(m Manifest) []Violation {
	var out []Violation
	for _, d := range m.Descriptors {
		if d.Scope == ScopeMigrate && d.APIRuntime {
			out = append(out, Violation{
				Class: ClassMigrateInAPI, Provides: d.Provides,
				Message: "migrate-scoped descriptor is marked APIRuntime — a migrate-only service must never be wired into the api/worker runtime",
			})
		}
	}
	return out
}
