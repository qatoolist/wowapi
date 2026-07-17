package app_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// countingMetrics records counter increments so the test can observe
// scheduler_task_errors_total without a real metrics adapter.
type countingMetrics struct {
	mu       sync.Mutex
	counters map[string]float64
}

func (m *countingMetrics) IncCounter(name string, v float64, _ map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counters == nil {
		m.counters = map[string]float64{}
	}
	m.counters[name] += v
}
func (m *countingMetrics) SetGauge(string, float64, map[string]string)            {}
func (m *countingMetrics) ObserveRequest(string, string, int, time.Duration, int) {}
func (m *countingMetrics) counter(name string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters[name]
}

// F-09 regression (adversarial-framework-review-2026-07-17): a recurring
// module job that fails for one or all tenants must (a) still attempt every
// tenant, and (b) surface a non-nil error to the scheduler observer so
// scheduler_task_errors_total increments — never report success. Failed
// tenants retry at the task's next interval (the schedule still advances).
func TestIntegrationModuleRecurringTenantFailureIsReported(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	metrics := &countingMetrics{}
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM, Metrics: metrics,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}

	tenantA := testkit.CreateTenant(t, h)
	tenantB := testkit.CreateTenant(t, h)
	_ = tenantB

	var attempts sync.Map // tenant id string -> attempt count
	var ticks atomic.Int64
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.RecurringJob("flaky", 100*time.Millisecond, func(ctx context.Context, _ database.TenantDB) error {
			tid, _ := database.TenantIDFrom(ctx)
			cnt, _ := attempts.LoadOrStore(tid.String(), new(atomic.Int64))
			cnt.(*atomic.Int64).Add(1)
			ticks.Add(1)
			if tid == tenantA.ID {
				return errors.New("boom for tenant A")
			}
			return nil
		})
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{
			RelayPoll:     80 * time.Millisecond,
			JobPoll:       80 * time.Millisecond,
			SchedulerPoll: 40 * time.Millisecond,
			ShutdownDrain: 3 * time.Second,
		})
	}()

	deadline := time.After(15 * time.Second)
	for metrics.counter("scheduler_task_errors_total") == 0 {
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("scheduler_task_errors_total never incremented despite a failing tenant (ticks=%d) — tenant failure reported as success", ticks.Load())
		case <-time.After(50 * time.Millisecond):
		}
	}

	// Every tenant must still have been attempted (failure isolation preserved).
	count := 0
	attempts.Range(func(_, _ any) bool { count++; return true })
	if count < 2 {
		t.Fatalf("only %d tenant(s) attempted; failing tenant must not block others", count)
	}

	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("StartWorker returned unexpected error: %v", err)
	}
}

// Closure-review regression (2026-07-17, F-09): when EVERY tenant fails, the
// run must still attempt every tenant, report failure, and keep scheduling
// (failed tenants retry at the next interval — the schedule advances).
func TestIntegrationModuleRecurringAllTenantsFailingIsReported(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	metrics := &countingMetrics{}
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM, Metrics: metrics,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	testkit.CreateTenant(t, h)
	testkit.CreateTenant(t, h)

	var attempts sync.Map
	var runs atomic.Int64
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.RecurringJob("doomed", 100*time.Millisecond, func(ctx context.Context, _ database.TenantDB) error {
			tid, _ := database.TenantIDFrom(ctx)
			attempts.LoadOrStore(tid.String(), true)
			runs.Add(1)
			return errors.New("boom for every tenant")
		})
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{
			RelayPoll:     80 * time.Millisecond,
			JobPoll:       80 * time.Millisecond,
			SchedulerPoll: 40 * time.Millisecond,
			ShutdownDrain: 3 * time.Second,
		})
	}()

	deadline := time.After(15 * time.Second)
	// Wait until BOTH: the error metric moved AND the job ran more times than
	// there are tenants (i.e. the schedule advanced past an all-fail interval).
	for metrics.counter("scheduler_task_errors_total") == 0 || runs.Load() <= 2 {
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("all-fail run not truthfully reported or schedule stalled (errors=%v runs=%d)",
				metrics.counter("scheduler_task_errors_total"), runs.Load())
		case <-time.After(50 * time.Millisecond):
		}
	}
	count := 0
	attempts.Range(func(_, _ any) bool { count++; return true })
	if count < 2 {
		t.Fatalf("only %d tenant(s) attempted in the all-fail case, want 2", count)
	}
	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("StartWorker: %v", err)
	}
}
