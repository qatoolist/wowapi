package app

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Second closure-audit regression (2026-07-17, F-10): the runtime view's
// fallback signal is the explicit `set` flag, never member nil-ness — a
// product with ZERO recurring jobs (nil validated slice) must still be immune
// to a caller assigning Booted.Recurring, and the worker/readiness consumers
// (runtimeRecurring / runtimeHealth) must read the validated view.
func TestRuntimeViewIgnoresFieldReplacement(t *testing.T) {
	real := RecurringJob{Name: "widgets.real", Every: time.Minute}
	realCheck := func(context.Context) error { return nil }
	b := &Booted{
		Recurring: []RecurringJob{real},
		Health:    map[string]func(context.Context) error{"real": realCheck},
		runtime: runtimeView{
			set:       true,
			recurring: []RecurringJob{real},
			health:    map[string]func(context.Context) error{"real": realCheck},
		},
	}
	b.Recurring = append(b.Recurring, RecurringJob{Name: "evil.injected", Every: time.Second})
	b.Health = map[string]func(context.Context) error{"evil": realCheck}

	got := b.runtimeRecurring()
	if len(got) != 1 || got[0].Name != "widgets.real" {
		t.Fatalf("runtimeRecurring follows the replaced Recurring field: %+v", got)
	}
	hc := b.runtimeHealth()
	if _, ok := hc["evil"]; ok {
		t.Fatal("runtimeHealth follows the replaced Health field")
	}
	if _, ok := hc["real"]; !ok {
		t.Fatal("runtimeHealth lost the boot-validated check")
	}
}

// A zero-recurring product: the validated view is an EMPTY (nil) slice, and a
// late append to the exported field must still be invisible to the scheduler.
func TestRuntimeViewZeroRecurringIsNotAFallbackHole(t *testing.T) {
	b := &Booted{runtime: runtimeView{set: true}}
	b.Recurring = []RecurringJob{{Name: "evil.injected", Every: time.Second}}
	if got := b.runtimeRecurring(); len(got) != 0 {
		t.Fatalf("zero-recurring product regained %d job(s) through field assignment", len(got))
	}
	b.Health = map[string]func(context.Context) error{"evil": func(context.Context) error { return nil }}
	if got := b.runtimeHealth(); len(got) != 0 {
		t.Fatalf("zero-health product regained %d check(s) through field assignment", len(got))
	}
}

// A hand-constructed Booted (never produced by Boot) must FAIL LOUDLY, never
// silently operate on unvalidated state (third closure audit 2026-07-17):
// there is deliberately no fallback from the runtime view to the exported
// fields.
func TestUnbootedBootedFailsLoudly(t *testing.T) {
	b := &Booted{Recurring: []RecurringJob{{Name: "manual.job", Every: time.Minute}}}

	if err := StartWorker(context.Background(), b, WorkerConfigOpts{}); !errors.Is(err, ErrNotBooted) {
		t.Fatalf("StartWorker on an unbooted value = %v, want ErrNotBooted", err)
	}
	if err := StartWorker(context.Background(), nil, WorkerConfigOpts{}); !errors.Is(err, ErrNotBooted) {
		t.Fatalf("StartWorker on nil = %v, want ErrNotBooted", err)
	}

	for name, fn := range map[string]func(){
		"RuntimeRouter":     func() { b.RuntimeRouter() },
		"RuntimeKernel":     func() { b.RuntimeKernel() },
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
