package jobs_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/jobs"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// seedDiscardedJob inserts a dead-lettered job via the admin pool and returns id.
func seedDiscardedJob(t *testing.T, h *testkit.DBHandle, lastErr string) int64 {
	t.Helper()
	var id int64
	if err := h.Admin.QueryRow(context.Background(),
		`INSERT INTO jobs_queue (kind, tenant_id, payload, status, attempts, max_attempts, last_error, finished_at)
		 VALUES ('test.dlq.job', NULL, '{"n":"x"}', 'discarded', 5, 5, $1, now())
		 RETURNING id`, lastErr).Scan(&id); err != nil {
		t.Fatalf("seed discarded job: %v", err)
	}
	return id
}

func TestIntegrationDLQJobsListReplayDiscard(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	replayID := seedDiscardedJob(t, h, "boom-replay")
	discardID := seedDiscardedJob(t, h, "boom-discard")

	// List includes both.
	entries, err := jobs.ListDead(ctx, h.Platform, 100)
	if err != nil {
		t.Fatalf("ListDead: %v", err)
	}
	seen := map[int64]bool{}
	for _, e := range entries {
		seen[e.ID] = true
	}
	if !seen[replayID] || !seen[discardID] {
		t.Fatalf("ListDead missing seeded jobs: %v", seen)
	}

	// Replay flips status back to available and resets attempts.
	if err := jobs.ReplayDead(ctx, h.Platform, replayID); err != nil {
		t.Fatalf("ReplayDead: %v", err)
	}
	if got := jobStatus(t, h, replayID); got != "available" {
		t.Fatalf("after replay status = %q, want available", got)
	}
	var attempts int
	if err := h.Platform.QueryRow(ctx, `SELECT attempts FROM jobs_queue WHERE id=$1`, replayID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 0 {
		t.Fatalf("after replay attempts = %d, want 0", attempts)
	}

	// Replaying a now-available job is a no-op error (not discarded anymore).
	if err := jobs.ReplayDead(ctx, h.Platform, replayID); errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("replay of non-discarded job should be KindNotFound, got %v", err)
	}

	// Discard permanently removes the row.
	if err := jobs.DiscardDead(ctx, h.Platform, discardID); err != nil {
		t.Fatalf("DiscardDead: %v", err)
	}
	var n int
	if err := h.Platform.QueryRow(ctx, `SELECT count(*) FROM jobs_queue WHERE id=$1`, discardID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("discarded job still present (%d rows)", n)
	}

	// Discard of a missing id is KindNotFound.
	if err := jobs.DiscardDead(ctx, h.Platform, discardID); errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("discard of missing job should be KindNotFound, got %v", err)
	}
}
