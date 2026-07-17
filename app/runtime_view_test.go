package app

import (
	"context"
	"errors"
	"testing"
)

// V2 (fifth closure audit 2026-07-17): Booted has NO informational mirror
// fields — the former field-replacement tests are structurally obsolete; only
// the fail-loud contract for unbooted values remains meaningful here.

// A hand-constructed Booted (never produced by Boot) must FAIL LOUDLY, never
// silently operate on unvalidated state (third closure audit 2026-07-17):
// there is deliberately no fallback from the runtime view to the exported
// fields.
func TestUnbootedBootedFailsLoudly(t *testing.T) {
	b := &Booted{}

	if err := StartWorker(context.Background(), b, WorkerConfigOpts{}); !errors.Is(err, ErrNotBooted) {
		t.Fatalf("StartWorker on an unbooted value = %v, want ErrNotBooted", err)
	}
	if err := StartWorker(context.Background(), nil, WorkerConfigOpts{}); !errors.Is(err, ErrNotBooted) {
		t.Fatalf("StartWorker on nil = %v, want ErrNotBooted", err)
	}

	for name, fn := range map[string]func(){
		"RuntimeRouter":     func() { b.RuntimeRouter() },
		"RuntimeAuthz":      func() { b.RuntimeAuthz() },
		"RuntimeTx":         func() { b.RuntimeTx() },
		"RuntimeEvents":     func() { b.RuntimeEvents() },
		"RuntimeJobs":       func() { b.RuntimeJobs() },
		"RuntimeMigrations": func() { _ = b.RuntimeMigrations() },
		"RuntimeSeeds":      func() { _ = b.RuntimeSeeds() },
		"RuntimeI18n":       func() { b.RuntimeI18n() },
		"runtimeRecurring":  func() { b.runtimeRecurring() },
		"runtimeHealth":     func() { b.runtimeHealth() },
		"runtimeSeeds":      func() { _ = b.runtimeSeeds() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s on an unbooted value did not panic — unvalidated state would run", name)
				}
			}()
			fn()
		})
	}
}
