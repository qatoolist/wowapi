package jobs_test

import (
	"context"
	"testing"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/testkit"
)

// enqueue_global_test.go — QA G12 (data path): EnqueueGlobal inserts a
// TENANT-LESS (tenant_id NULL) job — a distinct SQL path from the tenant Enqueue
// (used for cross-tenant/global work like sweepers). It had no test.

type globalJob struct {
	Task string `json:"task"`
}

func (globalJob) Kind() string { return "core.global.sweep" }

func TestIntegrationEnqueueGlobalInsertsTenantlessJob(t *testing.T) {
	h := testkit.NewDB(t)
	r := jobs.NewRunner(h.Platform, h.TxM, jobs.NewRegistry())

	if err := r.EnqueueGlobal(context.Background(), globalJob{Task: "nightly"}); err != nil {
		t.Fatalf("EnqueueGlobal: %v", err)
	}

	// Exactly one queued row of this kind, with a NULL tenant.
	var count, nullTenant int
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT count(*), count(*) FILTER (WHERE tenant_id IS NULL)
		   FROM jobs_queue WHERE kind = 'core.global.sweep'`).Scan(&count, &nullTenant); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("global enqueue produced %d rows, want 1", count)
	}
	if nullTenant != 1 {
		t.Fatalf("global job must have tenant_id NULL, got %d null rows", nullTenant)
	}
}

func TestEnqueueGlobalRejectsInvalidJob(t *testing.T) {
	h := testkit.NewDB(t)
	r := jobs.NewRunner(h.Platform, h.TxM, jobs.NewRegistry())
	ctx := context.Background()

	if err := r.EnqueueGlobal(ctx, nil); kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("nil job must be rejected, got %v", err)
	}
	if err := r.EnqueueGlobal(ctx, emptyKindJob{}); kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("empty-kind job must be rejected, got %v", err)
	}
}

type emptyKindJob struct{}

func (emptyKindJob) Kind() string { return "" }
