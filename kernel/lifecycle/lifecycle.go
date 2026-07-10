// Package lifecycle is wowapi's STATIC provider/lifecycle manifest (backlog
// B9). It is deliberately NOT a runtime DI container: there is no reflection,
// no codegen, and no service locator. It is a small, hand-maintained
// descriptor model — ProviderDescriptor{Provides, Requires, Scope} — that
// captures the REAL provider graph wired by kernel.New (kernel/kernel.go),
// app.Boot/moduleContext (app/context.go, app/boot.go), and module.Context
// (module/module.go), plus a pure-function lint over that manifest that
// catches the wiring-mistake classes the framework-competitive-architecture-
// benchmark's "DI / IoC: Static Lifecycle Graph For Go" section calls out.
//
// The manifest is maintained by hand alongside the wiring it describes (there
// is no generator reading kernel.go via reflection/AST) — Manifest() below IS
// the source of truth the lint runs over, and CurrentManifest's own test
// (TestCurrentManifestLintsClean) is the regression net: if a future change to
// kernel.New/app.Boot/module.Context adds a wiring mistake, either that test
// or a forgotten manifest update will need updating, keeping the manifest
// honest by review.
package lifecycle

import (
	"fmt"
	"sort"
	"strings"
)

// Scope is the lifetime a provided value is valid for. Mirrors the benchmark
// doc's recommended Go-native shape (framework-competitive-architecture-
// benchmark.md, "DI / IoC: Static Lifecycle Graph For Go").
type Scope string

const (
	// ScopeProcess: constructed once in kernel.New, lives for the process
	// lifetime (pools, registries, evaluators, kernel services).
	ScopeProcess Scope = "process"
	// ScopeRequest: constructed per inbound HTTP request (module.Context
	// accessors handed to Module.Register are process-scoped services, but
	// request-handling code that consumes them operates per-request).
	ScopeRequest Scope = "request"
	// ScopeTenantTx: valid only for the lifetime of one tenant-bound
	// transaction opened via database.TxManager.WithTenant/WithTenantRO
	// (database.TenantDB). Must never escape the callback that receives it.
	ScopeTenantTx Scope = "tenant_tx"
	// ScopeJob: constructed for one worker job execution.
	ScopeJob Scope = "job"
	// ScopeMigrate: valid only in the migrate process (privileged DDL
	// connection, app.SkipRLSEnforcementCheck). Must never be wired into a
	// process that serves API/worker runtime traffic.
	ScopeMigrate Scope = "migrate"
)

// valid reports whether s is one of the five recognized scopes.
func (s Scope) valid() bool {
	switch s {
	case ScopeProcess, ScopeRequest, ScopeTenantTx, ScopeJob, ScopeMigrate:
		return true
	default:
		return false
	}
}

// rank orders scopes from widest/longest-lived to narrowest/shortest-lived
// lifetime. A provider must not depend on something narrower-lived than
// itself (e.g. a process-scoped service depending on request-scoped state) —
// checkNoNarrowerDependency uses this ordering.
func (s Scope) rank() int {
	switch s {
	case ScopeMigrate:
		return 0
	case ScopeProcess:
		return 1
	case ScopeJob:
		return 2
	case ScopeRequest:
		return 3
	case ScopeTenantTx:
		return 4
	default:
		return -1
	}
}

// ProviderDescriptor describes one provided value/capability in the wiring
// graph: what it provides, what it requires (by Provides name of another
// descriptor), and the scope it is valid at.
type ProviderDescriptor struct {
	// Provides is the unique name of the capability this descriptor wires,
	// e.g. "kernel.Pool" or "module.Context.Tx". Conventionally
	// "<owner>.<FieldOrMethod>" matching the real Go identifier so the
	// manifest stays traceable to kernel.go/context.go/module.go.
	Provides string
	// Requires lists the Provides names this descriptor's construction
	// depends on. Every entry must match another descriptor's Provides (lint
	// class (e): missing provider) and must not form a cycle.
	Requires []string
	// Scope is this descriptor's lifetime.
	Scope Scope
	// RawPool marks a descriptor that hands out an unwrapped *pgxpool.Pool
	// (or equivalent raw connection) rather than a TxManager/TenantDB. Lint
	// class (b) flags any module-facing descriptor with this set.
	RawPool bool
	// TenantScoped marks a descriptor whose value must not outlive the
	// tenant transaction that produced it (a database.TenantDB or something
	// derived from one). Lint class (c) flags such a descriptor with a
	// Requires edge FROM a wider-scoped (process/request) descriptor that
	// would let the value escape its transaction.
	TenantScoped bool
	// APIRuntime marks a descriptor that is wired into the api/worker
	// runtime module.Context surface (i.e. reachable from a module's
	// Register). Lint class (d) flags a migrate-scoped descriptor with
	// APIRuntime set — a migrate-only service must never be wired into a
	// process that serves API/worker traffic.
	APIRuntime bool
}

// Manifest is an ordered set of descriptors. Order is preserved for stable
// printing; lint functions are order-independent (they index by Provides).
type Manifest struct {
	Descriptors []ProviderDescriptor
}

// ByProvides indexes the manifest's descriptors by their Provides name.
// Behavior is undefined (last-wins) if two descriptors share a Provides name
// — checkDuplicateProvides reports that as a violation.
func (m Manifest) ByProvides() map[string]ProviderDescriptor {
	idx := make(map[string]ProviderDescriptor, len(m.Descriptors))
	for _, d := range m.Descriptors {
		idx[d.Provides] = d
	}
	return idx
}

// Sorted returns a copy of the descriptors sorted by Provides, for
// deterministic printing/diffing.
func (m Manifest) Sorted() []ProviderDescriptor {
	out := make([]ProviderDescriptor, len(m.Descriptors))
	copy(out, m.Descriptors)
	sort.Slice(out, func(i, j int) bool { return out[i].Provides < out[j].Provides })
	return out
}

// Print renders the manifest as a stable, human-readable table (used by
// `wowapi lint lifecycle` and tests).
func (m Manifest) Print() string {
	var b strings.Builder
	for _, d := range m.Sorted() {
		fmt.Fprintf(&b, "%-32s scope=%-9s requires=%s\n", d.Provides, d.Scope, strings.Join(d.Requires, ","))
	}
	return b.String()
}
