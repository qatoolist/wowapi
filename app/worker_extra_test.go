package app_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/module"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestStartWorkerRequiresPlatformPool proves the worker/migrate posture guard:
// StartWorker refuses to run against a kernel with no app_platform pool (an
// api-only kernel), since the relay/job-runner/scheduler all do cross-tenant
// work that needs the platform grant.
func TestStartWorkerRequiresPlatformPool(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Tx: h.TxM, // no Platform
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	err = app.StartWorker(context.Background(), booted, app.WorkerConfigOpts{})
	if err == nil || !strings.Contains(err.Error(), "requires the kernel Platform pool") {
		t.Fatalf("StartWorker error = %v, want the no-platform-pool guard", err)
	}
}

// TestStartWorkerDrainTimeout proves the hard shutdown cap (ARCH-57): with a
// 1ns drain budget the select cannot observe the loops finishing before the
// deadline timer fires, so StartWorker releases with work possibly in flight
// rather than hanging the process. A cancelled context drives it straight to
// the drain race.
func TestStartWorkerDrainTimeout(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already done: StartWorker reaches the drain select immediately

	err = app.StartWorker(ctx, booted, app.WorkerConfigOpts{
		RelayPoll:     10 * time.Millisecond,
		JobPoll:       10 * time.Millisecond,
		SchedulerPoll: 10 * time.Millisecond,
		ShutdownDrain: time.Nanosecond, // deadline fires before the loops can drain
	})
	if err == nil || !strings.Contains(err.Error(), "drain deadline exceeded") {
		t.Fatalf("StartWorker error = %v, want the drain-deadline hard cap", err)
	}
}

// TestStartWorkerRunsMaintenanceAndModuleRecurring boots a worker whose module
// registered a per-tenant recurring job, creates an active tenant, then runs the
// worker. The scheduler must fan the recurring job out to the active tenant (in
// that tenant's transaction), proving registerModuleRecurring + activeTenants +
// the kernel-maintenance registration all execute; StartWorker must then return
// cleanly on context cancellation (graceful drain).
func TestStartWorkerRunsMaintenanceAndModuleRecurring(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}

	var ranForTenant atomic.Int64
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.RecurringJob("beat", 120*time.Millisecond, func(_ context.Context, _ database.TenantDB) error {
			ranForTenant.Add(1)
			return nil
		})
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	// An active tenant so activeTenants returns a row and the per-tenant fan-out
	// actually invokes the module callback.
	testkit.CreateTenant(t, h)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{
			RelayPoll:           80 * time.Millisecond,
			JobPoll:             80 * time.Millisecond,
			SchedulerPoll:       40 * time.Millisecond,
			SLAInterval:         120 * time.Millisecond,
			IdempotencyInterval: 120 * time.Millisecond,
			ShutdownDrain:       3 * time.Second,
		})
	}()

	deadline := time.After(12 * time.Second)
	for ranForTenant.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("module recurring job never fanned out to the active tenant")
		case err := <-done:
			t.Fatalf("StartWorker returned early: %v", err)
		case <-time.After(50 * time.Millisecond):
		}
	}

	cancel() // graceful shutdown
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StartWorker returned error on shutdown: %v", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("StartWorker did not drain within the shutdown window")
	}
}
