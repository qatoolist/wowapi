package sequence_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/sequence"
	"github.com/qatoolist/wowapi/testkit"
)

// fixedIDGen mints the same id every time so a second Allocate collides on the
// sequence_allocations primary key — used to exercise the ledger-insert error
// branch with a real unique-violation from Postgres.
type fixedIDGen struct{ id uuid.UUID }

func (g fixedIDGen) New() uuid.UUID { return g.id }

// TestIntegrationSequenceNewDefaultsIDGen covers New(nil): a nil generator must
// fall back to the production UUIDv7 generator, and the resulting allocator must
// still mint a real ledger id.
func TestIntegrationSequenceNewDefaultsIDGen(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(nil) // exercises the idgen==nil default branch
	ctx := database.WithTenantID(context.Background(), uuid.New())

	var al sequence.Allocation
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var err error
		al, err = a.Allocate(ctx, db, "receipt")
		return err
	}); err != nil {
		t.Fatalf("allocate with default idgen: %v", err)
	}
	if al.Value != 1 {
		t.Fatalf("value = %d, want 1", al.Value)
	}
	if al.ID == uuid.Nil {
		t.Fatalf("default idgen minted a nil id — fallback generator not wired")
	}
}

// TestIntegrationSequenceAllocateEmptyKey covers the seriesKey=="" validation
// branch: it must fail closed with KindValidation and consume nothing.
func TestIntegrationSequenceAllocateEmptyKey(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, aerr := a.Allocate(ctx, db, "")
		return aerr
	})
	if kerr.KindOf(err) != kerr.KindValidation {
		t.Fatalf("empty series key: kind = %v, want KindValidation (err=%v)", kerr.KindOf(err), err)
	}
	e, ok := kerr.As(err)
	if !ok || e.Code != "invalid_series" {
		t.Fatalf("empty series key: code = %v, want invalid_series (err=%v)", e, err)
	}
}

// TestIntegrationSequenceAllocateRecordsActor covers actorOrNil's actor-present
// branch and asserts the ledger row's allocated_by equals the acting user.
func TestIntegrationSequenceAllocateRecordsActor(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	actor := uuid.New()
	ctx := database.WithActorID(database.WithTenantID(context.Background(), uuid.New()), actor)

	var al sequence.Allocation
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var err error
		al, err = a.Allocate(ctx, db, "voucher")
		return err
	}); err != nil {
		t.Fatalf("allocate with actor: %v", err)
	}

	var got uuid.UUID
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT allocated_by FROM sequence_allocations
			  WHERE tenant_id = app_tenant_id() AND series_key = $1 AND value = $2`,
			"voucher", al.Value).Scan(&got)
	}); err != nil {
		t.Fatalf("read allocated_by: %v", err)
	}
	if got != actor {
		t.Fatalf("allocated_by = %v, want %v (actor not recorded)", got, actor)
	}
}

// TestIntegrationSequencePeekNeverAllocated covers Peek's pgx.ErrNoRows branch:
// a series that has never allocated reports 0, not an error.
func TestIntegrationSequencePeekNeverAllocated(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	var peek int64 = -1
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		var err error
		peek, err = a.Peek(ctx, db, "never-touched")
		return err
	}); err != nil {
		t.Fatalf("peek of unallocated series should not error: %v", err)
	}
	if peek != 0 {
		t.Fatalf("Peek of never-allocated series = %d, want 0", peek)
	}
}

// TestIntegrationSequenceAllocateLedgerInsertError covers the ledger-insert error
// branch: a colliding id triggers a real unique-violation, and Allocate wraps it
// under the sequence.Allocate op while the failed tx rolls back gap-free.
func TestIntegrationSequenceAllocateLedgerInsertError(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(fixedIDGen{id: uuid.New()})
	ctx := database.WithTenantID(context.Background(), uuid.New())

	// First allocation commits with the fixed id (value 1).
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, aerr := a.Allocate(ctx, db, "cert")
		return aerr
	}); err != nil {
		t.Fatalf("first allocate: %v", err)
	}

	// Second allocation reuses the same id -> primary-key violation on insert.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, aerr := a.Allocate(ctx, db, "cert")
		return aerr
	})
	if err == nil {
		t.Fatal("expected a unique-violation error on the colliding ledger insert")
	}
	if e, ok := kerr.As(err); !ok || e.Op != "sequence.Allocate" {
		t.Fatalf("error not wrapped under sequence.Allocate: %v", err)
	}

	// The failed tx rolled back, so the counter did not advance: next is still 2.
	// Use a fresh generator so the recovery insert does not collide again.
	a2 := sequence.New(model.UUIDv7())
	var next int64
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		al, aerr := a2.Allocate(ctx, db, "cert")
		next = al.Value
		return aerr
	}); err != nil {
		t.Fatalf("recovery allocate: %v", err)
	}
	if next != 2 {
		t.Fatalf("next value = %d, want 2 (rolled-back allocation must not leave a gap)", next)
	}
}

// TestIntegrationSequenceAllocateQueryError covers Allocate's counter-advance
// error branch: a cancelled context makes the RETURNING query fail, and the
// error is wrapped under sequence.Allocate.
func TestIntegrationSequenceAllocateQueryError(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	base := database.WithTenantID(context.Background(), uuid.New())
	ctx, cancel := context.WithCancel(base)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		cancel() // poison the connection before the counter query runs
		_, aerr := a.Allocate(ctx, db, "receipt")
		return aerr
	})
	if err == nil {
		t.Fatal("expected the cancelled-context query to fail")
	}
	if e, ok := kerr.As(err); !ok || e.Op != "sequence.Allocate" {
		t.Fatalf("error not wrapped under sequence.Allocate: %v", err)
	}
}

// TestIntegrationSequenceVoidExecError covers Void's UPDATE error branch: a
// cancelled context makes the void UPDATE fail, wrapped under sequence.Void.
func TestIntegrationSequenceVoidExecError(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	base := database.WithTenantID(context.Background(), uuid.New())

	// Allocate one number so there is something to (attempt to) void.
	if err := h.TxM.WithTenant(base, func(ctx context.Context, db database.TenantDB) error {
		_, aerr := a.Allocate(ctx, db, "inv")
		return aerr
	}); err != nil {
		t.Fatalf("allocate: %v", err)
	}

	ctx, cancel := context.WithCancel(base)
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		cancel()
		return a.Void(ctx, db, "inv", 1, "boom")
	})
	if err == nil {
		t.Fatal("expected the cancelled-context void UPDATE to fail")
	}
	if e, ok := kerr.As(err); !ok || e.Op != "sequence.Void" {
		t.Fatalf("error not wrapped under sequence.Void: %v", err)
	}
}

// TestIntegrationSequencePeekQueryError covers Peek's non-ErrNoRows error branch:
// a cancelled context makes the read fail and is wrapped under sequence.Peek.
func TestIntegrationSequencePeekQueryError(t *testing.T) {
	h := testkit.NewDB(t)
	a := sequence.New(model.UUIDv7())
	base := database.WithTenantID(context.Background(), uuid.New())
	ctx, cancel := context.WithCancel(base)

	err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		cancel()
		_, perr := a.Peek(ctx, db, "receipt")
		return perr
	})
	if err == nil {
		t.Fatal("expected the cancelled-context peek to fail")
	}
	if e, ok := kerr.As(err); !ok || e.Op != "sequence.Peek" {
		t.Fatalf("error not wrapped under sequence.Peek: %v", err)
	}
}
