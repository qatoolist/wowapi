package outbox_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationDLQListDefaultLimit covers ListDeadEvents' limit<=0 default
// (→50): a non-positive limit still lists dead events rather than returning
// none.
func TestIntegrationDLQListDefaultLimit(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := uuid.New()
	id := seedDeadEvent(t, h, tenant, "boom-default-limit")

	entries, err := outbox.ListDeadEvents(context.Background(), h.Platform, 0)
	if err != nil {
		t.Fatalf("ListDeadEvents(limit=0): %v", err)
	}
	found := false
	for _, e := range entries {
		if e.ID == id {
			found = true
		}
	}
	if !found {
		t.Fatalf("ListDeadEvents(limit=0) did not return seeded dead event %s", id)
	}
}

// TestIntegrationDLQDiscardNotFound covers DiscardDeadEvent's zero-rows arm: an
// unknown (non-dead) id yields KindNotFound.
func TestIntegrationDLQDiscardNotFound(t *testing.T) {
	h := testkit.NewDB(t)
	err := outbox.DiscardDeadEvent(context.Background(), h.Platform, uuid.New())
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("DiscardDeadEvent(unknown) = %v, want KindNotFound", err)
	}
}

// TestDLQPoolFailures covers the DB-failure branches of the three DLQ helpers
// via a closed pool: query/exec errors are wrapped and surfaced.
func TestDLQPoolFailures(t *testing.T) {
	pool := closedPool(t)
	ctx := context.Background()

	if _, err := outbox.ListDeadEvents(ctx, pool, 10); err == nil {
		t.Fatal("ListDeadEvents on closed pool = nil error, want query failure")
	}
	if err := outbox.ReplayDeadEvent(ctx, pool, uuid.New()); err == nil {
		t.Fatal("ReplayDeadEvent on closed pool = nil error, want exec failure")
	}
	if err := outbox.DiscardDeadEvent(ctx, pool, uuid.New()); err == nil {
		t.Fatal("DiscardDeadEvent on closed pool = nil error, want exec failure")
	}
}
