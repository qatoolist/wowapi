package sequence_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/sequence"
	"github.com/qatoolist/wowapi/v2/testkit"
)

func TestIntegrationSequenceSequential(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	for want := int64(1); want <= 3; want++ {
		var got int64
		if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			al, err := a.Allocate(ctx, db, "receipt")
			got = al.Value
			return err
		}); err != nil {
			t.Fatalf("allocate: %v", err)
		}
		if got != want {
			t.Fatalf("allocation = %d, want %d", got, want)
		}
	}
	// Peek reports the last issued value without consuming one.
	var peek int64
	_ = h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var err error
		peek, err = a.Peek(ctx, db, "receipt")
		return err
	})
	if peek != 3 {
		t.Fatalf("Peek = %d, want 3", peek)
	}
}

// TestIntegrationSequenceGapFreeOnRollback is the core E3 guarantee: a number is
// consumed only if the caller's transaction commits. A rollback frees it.
func TestIntegrationSequenceGapFreeOnRollback(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	errRollback := errors.New("force rollback")
	// Allocate #1 but roll the transaction back.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		al, aerr := a.Allocate(ctx, db, "voucher")
		if aerr != nil {
			return aerr
		}
		if al.Value != 1 {
			t.Fatalf("first allocation = %d, want 1", al.Value)
		}
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("expected the forced rollback error, got %v", err)
	}

	// The next committed allocation must reuse 1 — no gap from the rolled-back one.
	var got int64
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		al, aerr := a.Allocate(ctx, db, "voucher")
		got = al.Value
		return aerr
	}); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("after rollback next value = %d, want 1 (gap-free)", got)
	}
}

// TestIntegrationSequenceConcurrentNoGapsNoDupes fires N concurrent allocations,
// each in its own committed transaction, and asserts the values are exactly
// 1..N — no duplicate (race) and no gap.
func TestIntegrationSequenceConcurrentNoGapsNoDupes(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	tenant := uuid.New()

	const n = 12
	var mu sync.Mutex
	seen := make(map[int64]bool, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			ctx := database.WithTenantID(context.Background(), tenant)
			_ = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
				al, err := a.Allocate(ctx, db, "cert")
				if err != nil {
					return err
				}
				mu.Lock()
				seen[al.Value] = true
				mu.Unlock()
				return nil
			})
		}()
	}
	wg.Wait()

	if len(seen) != n {
		t.Fatalf("got %d distinct values, want %d (a duplicate slipped through)", len(seen), n)
	}
	for v := int64(1); v <= n; v++ {
		if !seen[v] {
			t.Fatalf("value %d missing — the series has a gap", v)
		}
	}
}

func TestIntegrationSequenceVoid(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	// Allocate 1 and 2, void 1.
	mustAllocate(t, h, a, ctx, "inv") // 1
	mustAllocate(t, h, a, ctx, "inv") // 2
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return a.Void(ctx, db, "inv", 1, "issued in error")
	}); err != nil {
		t.Fatalf("void: %v", err)
	}
	// Voiding again is a conflict.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return a.Void(ctx, db, "inv", 1, "again")
	}); kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("double void should be KindConflict, got %v", err)
	}
	// Voiding an unallocated number is not-found.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return a.Void(ctx, db, "inv", 999, "nope")
	}); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("void of unallocated value should be KindNotFound, got %v", err)
	}
	// The void does NOT renumber: the next allocation is 3, leaving the gap at 1.
	var got int64
	_ = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		al, err := a.Allocate(ctx, db, "inv")
		got = al.Value
		return err
	})
	if got != 3 {
		t.Fatalf("after void, next value = %d, want 3 (voids leave gaps, never renumber)", got)
	}
}

func TestIntegrationSequenceTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	t1 := database.WithTenantID(context.Background(), uuid.New())
	t2 := database.WithTenantID(context.Background(), uuid.New())

	if v := mustAllocate(t, h, a, t1, "s"); v != 1 {
		t.Fatalf("tenant1 first = %d, want 1", v)
	}
	if v := mustAllocate(t, h, a, t1, "s"); v != 2 {
		t.Fatalf("tenant1 second = %d, want 2", v)
	}
	// A different tenant starts its own series at 1 (RLS-scoped).
	if v := mustAllocate(t, h, a, t2, "s"); v != 1 {
		t.Fatalf("tenant2 first = %d, want 1 (independent series)", v)
	}
}

func mustAllocate(t *testing.T, h *testkit.DBHandle, a *sequence.Allocator, ctx context.Context, series string) int64 {
	t.Helper()
	var v int64
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		al, err := a.Allocate(ctx, db, series)
		v = al.Value
		return err
	}); err != nil {
		t.Fatalf("allocate %q: %v", series, err)
	}
	return v
}
