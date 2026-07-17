package testkit_test

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/jobs"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/module"
	"github.com/qatoolist/wowapi/v2/testkit"
)

type pingJob struct{}

func (pingJob) Kind() string { return "test.ping" }

// inlineModule is a minimal in-test module that wires event + job handlers.
type inlineModule struct {
	register func(mc module.Context) error
}

func (inlineModule) Name() string                       { return "test" }
func (inlineModule) DependsOn() []string                { return nil }
func (m inlineModule) Register(mc module.Context) error { return m.register(mc) }

type discardW struct{}

func (discardW) Write(p []byte) (int, error) { return len(p), nil }

// TestIntegrationWorkerEndToEnd boots an app whose module subscribes to an event
// and registers a job kind, runs StartWorker, and asserts the emitted event is
// dispatched and the enqueued job executed, then that StartWorker returns
// cleanly on context cancellation (graceful shutdown).
func TestIntegrationWorkerEndToEnd(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(discardW{}, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatal(err)
	}

	var eventSeen, jobRan atomic.Int64
	a := app.New()
	a.Register(inlineModule{register: func(mc module.Context) error {
		mc.Events().Subscribe("test.thing.happened", "test.handler",
			func(ctx context.Context, db database.TenantDB, e outbox.DispatchedEvent) error {
				eventSeen.Add(1)
				return nil
			})
		mc.Jobs().RegisterKindWithIdempotency("test.ping", func(ctx context.Context, db database.TenantDB, payload []byte) error {
			jobRan.Add(1)
			return nil
		}, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("boot: %v", err)
	}

	tn := testkit.CreateTenant(t, h)
	w := outbox.NewWriter(model.UUIDv7())
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		if e := w.Write(ctx, db, outbox.Event{Type: "test.thing.happened"}); e != nil {
			return e
		}
		return jobs.Enqueue(ctx, db, pingJob{})
	}); err != nil {
		t.Fatalf("emit+enqueue: %v", err)
	}

	// Generous ceilings: a normal run finishes in ~1s, but coverage
	// instrumentation and a loaded CI box slow the relay→job pipeline, so keep
	// the timeouts well above the happy path to avoid a flaky deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.StartWorker(ctx, booted, app.WorkerConfigOpts{RelayPoll: 100 * time.Millisecond, JobPoll: 100 * time.Millisecond})
	}()

	deadline := time.After(30 * time.Second)
	for eventSeen.Load() == 0 || jobRan.Load() == 0 {
		select {
		case <-deadline:
			t.Fatalf("worker did not process in time: eventSeen=%d jobRan=%d", eventSeen.Load(), jobRan.Load())
		case <-time.After(50 * time.Millisecond):
		}
	}
	cancel() // trigger graceful shutdown
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StartWorker returned error: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("StartWorker did not shut down within the drain window")
	}
}
