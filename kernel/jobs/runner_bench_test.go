package jobs_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/testkit"
)

type benchmarkJob struct {
	Sequence int `json:"sequence"`
}

func (benchmarkJob) Kind() string { return "benchmark.claim-finalize" }

// BenchmarkJobClaimFinalize measures the real PostgreSQL claim statement,
// tenant worker transaction, and fenced success finalization for one job.
// Enqueue setup is excluded from the timed region.
func BenchmarkJobClaimFinalize(b *testing.B) {
	h := testkit.NewDB(b)
	tenant := testkit.CreateTenant(b, h)
	ctx := testkit.TenantCtx(tenant.ID)
	registry := jobs.NewRegistry()
	registry.RegisterKind(benchmarkJob{}.Kind(),
		func(context.Context, database.TenantDB, []byte) error { return nil },
		jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
	if err := registry.Err(); err != nil {
		b.Fatal(err)
	}
	runner := jobs.NewRunner(h.Platform, h.TxM, registry, jobs.WithPoolSize(1))

	b.ReportAllocs()
	b.ResetTimer()
	sequence := 0
	for b.Loop() {
		b.StopTimer()
		sequence++
		err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			return jobs.Enqueue(ctx, db, benchmarkJob{Sequence: sequence})
		})
		if err != nil {
			b.Fatalf("enqueue fixture: %v", err)
		}
		b.StartTimer()
		claimed, err := runner.ClaimOnce(ctx)
		if err != nil || claimed != 1 {
			b.Fatalf("claim/finalize: claimed=%d err=%v", claimed, err)
		}
	}
}
