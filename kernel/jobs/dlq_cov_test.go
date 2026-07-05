package jobs_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationListDeadDefaultLimit proves ListDead applies the default limit
// (50) when called with limit <= 0 and still returns the seeded discarded job.
func TestIntegrationListDeadDefaultLimit(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	id := seedDiscardedJob(t, h, "boom-default-limit")

	entries, err := jobs.ListDead(ctx, h.Platform, 0) // limit <= 0 → default 50
	if err != nil {
		t.Fatalf("ListDead: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.ID == id {
			found = true
			if e.LastError != "boom-default-limit" {
				t.Fatalf("LastError = %q, want boom-default-limit", e.LastError)
			}
		}
	}
	if !found {
		t.Fatalf("ListDead(limit=0) did not return seeded job %d", id)
	}
}
