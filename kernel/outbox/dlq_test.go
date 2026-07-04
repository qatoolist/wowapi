package outbox_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/testkit"
)

// seedDeadEvent inserts a dead-lettered event via the admin pool and returns id.
func seedDeadEvent(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, lastErr string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO events_outbox (id, tenant_id, event_type, payload, dispatch_status, attempts, max_attempts, last_error, failed_at, created_by)
		 VALUES ($1, $2, 'test.dead.event', '{"n":"x"}', 'dead', 10, 10, $3, now(), $4)`,
		id, tenant, lastErr, uuid.New()); err != nil {
		t.Fatalf("seed dead event: %v", err)
	}
	return id
}

func TestIntegrationDLQEventsListReplayDiscard(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	tenant := uuid.New()

	replayID := seedDeadEvent(t, h, tenant, "boom-replay")
	discardID := seedDeadEvent(t, h, tenant, "boom-discard")

	entries, err := outbox.ListDeadEvents(ctx, h.Platform, 100)
	if err != nil {
		t.Fatalf("ListDeadEvents: %v", err)
	}
	seen := map[uuid.UUID]bool{}
	for _, e := range entries {
		seen[e.ID] = true
	}
	if !seen[replayID] || !seen[discardID] {
		t.Fatalf("ListDeadEvents missing seeded events: %v", seen)
	}

	// Replay flips the event back to pending for re-dispatch.
	if err := outbox.ReplayDeadEvent(ctx, h.Platform, replayID); err != nil {
		t.Fatalf("ReplayDeadEvent: %v", err)
	}
	var status string
	var attempts int
	if err := h.Platform.QueryRow(ctx,
		`SELECT dispatch_status, attempts FROM events_outbox WHERE id=$1`, replayID).Scan(&status, &attempts); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || attempts != 0 {
		t.Fatalf("after replay status=%q attempts=%d, want pending/0", status, attempts)
	}

	// Replaying a non-dead event is KindNotFound.
	if err := outbox.ReplayDeadEvent(ctx, h.Platform, replayID); errors.KindOf(err) != errors.KindNotFound {
		t.Fatalf("replay of non-dead event should be KindNotFound, got %v", err)
	}

	// Discard permanently removes the dead event.
	if err := outbox.DiscardDeadEvent(ctx, h.Platform, discardID); err != nil {
		t.Fatalf("DiscardDeadEvent: %v", err)
	}
	var n int
	if err := h.Platform.QueryRow(ctx, `SELECT count(*) FROM events_outbox WHERE id=$1`, discardID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("discarded event still present (%d rows)", n)
	}
}
