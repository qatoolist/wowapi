package lifecycle

import (
	"strings"
	"testing"
)

// TestCurrentManifestLintsClean is the CI regression net (backlog B9): it
// asserts the REAL, current wowapi wiring graph (CurrentManifest, hand-built
// from kernel.New / app.Boot / module.Context) has zero lint violations. If a
// future change wires a migrate-only service into api, hands a module a raw
// pool, lets a tenant-scoped value escape its transaction, or introduces a
// missing-provider/cycle, this test fails — provided the manifest is kept in
// sync with the real wiring (CurrentManifest's doc comment states that
// obligation).
func TestCurrentManifestLintsClean(t *testing.T) {
	violations := Lint(CurrentManifest())
	if len(violations) != 0 {
		var b strings.Builder
		for _, v := range violations {
			b.WriteString(v.String())
			b.WriteString("\n")
		}
		t.Fatalf("current manifest has %d lint violation(s):\n%s", len(violations), b.String())
	}
}

// TestManifestPrintDeterministic guards the CLI's `wowapi lint lifecycle`
// output: Print must be alphabetically stable across calls (CI diffing / log
// comparison depends on this).
func TestManifestPrintDeterministic(t *testing.T) {
	m := CurrentManifest()
	first := m.Print()
	second := m.Print()
	if first != second {
		t.Fatalf("Print is not deterministic across calls")
	}
	// Spot check a known entry renders with its scope.
	if !strings.Contains(first, "kernel.Tx") || !strings.Contains(first, "scope=process") {
		t.Fatalf("Print output missing expected kernel.Tx/process entry:\n%s", first)
	}
}

func TestScopeRankInvalid(t *testing.T) {
	if got := Scope("bogus").rank(); got != -1 {
		t.Fatalf("Scope(%q).rank() = %d, want -1", "bogus", got)
	}
}

func TestLint_TenantEscape_SkipsInvalidScopeDependent(t *testing.T) {
	// A dependent with an invalid scope must not be reported as a tenant
	// escape (that dependent's own scope is already flagged by
	// checkInvalidScope) — checkTenantEscape's !other.Scope.valid() guard.
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "kernel.Tx", Scope: ScopeProcess},
		{Provides: "database.TenantDB", Requires: []string{"kernel.Tx"}, Scope: ScopeTenantTx},
		{Provides: "weird.dependent", Requires: []string{"database.TenantDB"}, Scope: Scope("bogus")},
	}}
	violations := Lint(m)
	if hasClass(violations, ClassTenantEscape, "database.TenantDB") {
		t.Fatalf("did not expect a %s violation when the only dependent has an invalid scope: %v", ClassTenantEscape, violations)
	}
	if !hasClass(violations, ClassInvalidScope, "weird.dependent") {
		t.Fatalf("expected the invalid scope on weird.dependent to still be reported: %v", violations)
	}
}

func TestLint_ViolationSortTieBreakByMessage(t *testing.T) {
	// Two violations sharing the same Provides but different classes exercise
	// the Class/Message tie-break in Lint's final sort.
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "a", Requires: []string{"a", "missing.x"}, Scope: Scope("bogus")},
	}}
	violations := Lint(m)
	if len(violations) < 2 {
		t.Fatalf("expected at least 2 violations on %q to exercise the sort tie-break, got %v", "a", violations)
	}
	for i := 1; i < len(violations); i++ {
		prev, cur := violations[i-1], violations[i]
		if prev.Provides > cur.Provides {
			t.Fatalf("violations not sorted by Provides: %v", violations)
		}
	}
}

func TestScopeValidAndRank(t *testing.T) {
	valid := []Scope{ScopeProcess, ScopeRequest, ScopeTenantTx, ScopeJob, ScopeMigrate}
	for _, s := range valid {
		if !s.valid() {
			t.Errorf("Scope(%q).valid() = false, want true", s)
		}
	}
	if Scope("bogus").valid() {
		t.Errorf("Scope(\"bogus\").valid() = true, want false")
	}
	// Widest-to-narrowest ordering: migrate < process < job < request < tenant_tx.
	inOrder := ScopeMigrate.rank() < ScopeProcess.rank() &&
		ScopeProcess.rank() < ScopeJob.rank() &&
		ScopeJob.rank() < ScopeRequest.rank() &&
		ScopeRequest.rank() < ScopeTenantTx.rank()
	if !inOrder {
		t.Fatalf("scope rank ordering violated: migrate=%d process=%d job=%d request=%d tenant_tx=%d",
			ScopeMigrate.rank(), ScopeProcess.rank(), ScopeJob.rank(), ScopeRequest.rank(), ScopeTenantTx.rank())
	}
}

// --- Negative tests: one deliberately-broken manifest per violation class. ---

func TestLint_ScopeLeak_ProcessDependsOnRequest(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "req.thing", Scope: ScopeRequest},
		{Provides: "proc.thing", Requires: []string{"req.thing"}, Scope: ScopeProcess},
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassScopeLeak, "proc.thing") {
		t.Fatalf("expected a %s violation on proc.thing, got: %v", ClassScopeLeak, violations)
	}
}

func TestLint_RawPoolIntoAPIRuntime(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "module.Context.RawDB", Scope: ScopeProcess, RawPool: true, APIRuntime: true},
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassRawPool, "module.Context.RawDB") {
		t.Fatalf("expected a %s violation, got: %v", ClassRawPool, violations)
	}
}

func TestLint_RawPoolInternalOnlyIsFine(t *testing.T) {
	// A raw pool that is NOT APIRuntime (e.g. kernel.Pool feeding kernel.Tx,
	// never handed to a module) must not be flagged.
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "kernel.Pool", Scope: ScopeProcess, RawPool: true},
		{Provides: "kernel.Tx", Requires: []string{"kernel.Pool"}, Scope: ScopeProcess},
	}}
	violations := Lint(m)
	if hasClass(violations, ClassRawPool, "kernel.Pool") {
		t.Fatalf("did not expect a %s violation for a kernel-internal raw pool: %v", ClassRawPool, violations)
	}
}

func TestLint_TenantEscape(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "kernel.Tx", Scope: ScopeProcess},
		{Provides: "database.TenantDB", Requires: []string{"kernel.Tx"}, Scope: ScopeTenantTx},
		// A process-scoped cache illegitimately retaining a TenantDB handle.
		{Provides: "kernel.LeakyCache", Requires: []string{"database.TenantDB"}, Scope: ScopeProcess},
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassTenantEscape, "database.TenantDB") {
		t.Fatalf("expected a %s violation, got: %v", ClassTenantEscape, violations)
	}
}

func TestLint_MigrateOnlyWiredIntoAPIRuntime(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "migrate.DDLConn", Scope: ScopeMigrate, APIRuntime: true},
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassMigrateInAPI, "migrate.DDLConn") {
		t.Fatalf("expected a %s violation, got: %v", ClassMigrateInAPI, violations)
	}
}

func TestLint_MigrateOnlyConfinedToMigrateIsFine(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "migrate.DDLConn", Scope: ScopeMigrate, APIRuntime: false},
	}}
	violations := Lint(m)
	if hasClass(violations, ClassMigrateInAPI, "migrate.DDLConn") {
		t.Fatalf("did not expect a %s violation for a migrate-confined descriptor: %v", ClassMigrateInAPI, violations)
	}
}

func TestLint_MissingProvider(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "module.Context.Tx", Requires: []string{"kernel.Tx"}, Scope: ScopeProcess},
		// kernel.Tx is never declared.
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassMissingProvider, "module.Context.Tx") {
		t.Fatalf("expected a %s violation, got: %v", ClassMissingProvider, violations)
	}
}

func TestLint_Cycle(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "a", Requires: []string{"b"}, Scope: ScopeProcess},
		{Provides: "b", Requires: []string{"c"}, Scope: ScopeProcess},
		{Provides: "c", Requires: []string{"a"}, Scope: ScopeProcess},
	}}
	violations := Lint(m)
	found := false
	for _, v := range violations {
		if v.Class == ClassCycle {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a %s violation, got: %v", ClassCycle, violations)
	}
}

func TestLint_SelfCycle(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "a", Requires: []string{"a"}, Scope: ScopeProcess},
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassCycle, "a") {
		t.Fatalf("expected a %s violation for a self-dependency, got: %v", ClassCycle, violations)
	}
}

func TestLint_DuplicateProvides(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "dup.thing", Scope: ScopeProcess},
		{Provides: "dup.thing", Scope: ScopeProcess},
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassDuplicate, "dup.thing") {
		t.Fatalf("expected a %s violation, got: %v", ClassDuplicate, violations)
	}
}

func TestLint_InvalidScope(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "weird.thing", Scope: Scope("not-a-real-scope")},
	}}
	violations := Lint(m)
	if !hasClass(violations, ClassInvalidScope, "weird.thing") {
		t.Fatalf("expected a %s violation, got: %v", ClassInvalidScope, violations)
	}
}

func TestLint_CleanManifestHasNoViolations(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "kernel.Pool", Scope: ScopeProcess, RawPool: true},
		{Provides: "kernel.Tx", Requires: []string{"kernel.Pool"}, Scope: ScopeProcess},
		{Provides: "module.Context.Tx", Requires: []string{"kernel.Tx"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "database.TenantDB", Requires: []string{"kernel.Tx"}, Scope: ScopeTenantTx},
	}}
	violations := Lint(m)
	if len(violations) != 0 {
		t.Fatalf("expected no violations for a clean manifest, got: %v", violations)
	}
}

func TestViolationString(t *testing.T) {
	v := Violation{Class: ClassCycle, Provides: "a", Message: "dependency cycle: a -> b -> a"}
	got := v.String()
	if !strings.Contains(got, "[cycle]") || !strings.Contains(got, "a:") || !strings.Contains(got, "dependency cycle") {
		t.Fatalf("String() = %q, missing expected components", got)
	}
}

func TestManifest_ByProvidesAndSorted(t *testing.T) {
	m := Manifest{Descriptors: []ProviderDescriptor{
		{Provides: "z.thing", Scope: ScopeProcess},
		{Provides: "a.thing", Scope: ScopeProcess},
	}}
	idx := m.ByProvides()
	if len(idx) != 2 {
		t.Fatalf("ByProvides: got %d entries, want 2", len(idx))
	}
	sorted := m.Sorted()
	if sorted[0].Provides != "a.thing" || sorted[1].Provides != "z.thing" {
		t.Fatalf("Sorted() = %v, want a.thing before z.thing", sorted)
	}
}

// hasClass reports whether violations contains an entry with the given class
// and Provides name.
func hasClass(violations []Violation, class, provides string) bool {
	for _, v := range violations {
		if v.Class == class && v.Provides == provides {
			return true
		}
	}
	return false
}
