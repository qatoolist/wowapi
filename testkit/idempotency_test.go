package testkit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/errors"
)

// TestIntegrationIdempotencyStore exercises the pg-backed IdemStore end to end
// against real Postgres + RLS: first claim is fresh, completion stores the
// response, a retry with the same key+hash replays it, a retry with a
// different hash conflicts, and an in-flight duplicate is rejected.
func TestIntegrationIdempotencyStore(t *testing.T) {
	h := NewDB(t)
	store := database.NewIdemStore()
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)
	const scope, key, hash = "actor-1", "idem-key-1", "hash-abc"

	// First call: fresh claim + complete in one tx.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		rep, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
		if err != nil {
			return err
		}
		if !rep.Fresh {
			t.Fatal("first Begin should be Fresh")
		}
		return store.Complete(ctx, db, scope, key, 201, []byte(`{"id":"x"}`))
	}); err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Retry with same key+hash: replays the stored response.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		rep, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
		if err != nil {
			return err
		}
		if !rep.Found || rep.ResponseStatus != 201 || string(rep.ResponseBody) != `{"id":"x"}` {
			t.Fatalf("retry should replay stored response, got %+v", rep)
		}
		return nil
	}); err != nil {
		t.Fatalf("replay call: %v", err)
	}

	// Retry with a DIFFERENT hash (same key): conflict.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, err := store.Begin(ctx, db, scope, key, "different-hash", time.Hour)
		return err
	})
	if errors.KindOf(err) != errors.KindConflict {
		t.Fatalf("reused key with different request should conflict, got %v", err)
	}

	// A second tenant with the same key is independent (RLS isolation): fresh.
	other := database.WithTenantID(context.Background(), uuid.New())
	if err := h.TxM.WithTenant(other, func(ctx context.Context, db database.TenantDB) error {
		rep, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
		if err != nil {
			return err
		}
		if !rep.Fresh {
			t.Fatal("same key under a different tenant must be Fresh (RLS-scoped)")
		}
		return nil
	}); err != nil {
		t.Fatalf("other tenant: %v", err)
	}
}

// TestIntegrationIdempotencySweepExpired is the S5 regression: SweepExpired
// purges past-expiry keys across ALL tenants in one platform pass, while live
// keys survive. Without a sweep the table grew without bound.
func TestIntegrationIdempotencySweepExpired(t *testing.T) {
	h := NewDB(t)
	store := database.NewIdemStore()
	t1, t2 := uuid.New(), uuid.New()

	seed := func(tenant uuid.UUID, key string, ttl time.Duration) {
		t.Helper()
		ctx := database.WithTenantID(context.Background(), tenant)
		if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			rep, err := store.Begin(ctx, db, "actor-1", key, "h", ttl)
			if err != nil {
				return err
			}
			if !rep.Fresh {
				t.Fatalf("seed %s should be fresh", key)
			}
			return store.Complete(ctx, db, "actor-1", key, 200, []byte("{}"))
		}); err != nil {
			t.Fatalf("seed %s: %v", key, err)
		}
	}

	// A negative TTL seeds an already-expired row; positive stays live.
	seed(t1, "expired-1", -time.Hour)
	seed(t2, "expired-2", -time.Hour) // different tenant — proves cross-tenant sweep
	seed(t1, "live-1", time.Hour)

	n, err := store.SweepExpired(context.Background(), h.PlatformTxM, time.Now())
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if n < 2 {
		t.Fatalf("sweep removed %d rows; want >= 2 (both tenants' expired keys)", n)
	}

	// The live key survives — a fresh Begin replays the stored response.
	if err := h.TxM.WithTenant(database.WithTenantID(context.Background(), t1),
		func(ctx context.Context, db database.TenantDB) error {
			rep, err := store.Begin(ctx, db, "actor-1", "live-1", "h", time.Hour)
			if err != nil {
				return err
			}
			if !rep.Found {
				t.Fatalf("live key must survive the sweep and replay, got %+v", rep)
			}
			return nil
		}); err != nil {
		t.Fatalf("post-sweep live check: %v", err)
	}

	// The other tenant's expired key is gone — a fresh Begin no longer replays.
	if err := h.TxM.WithTenant(database.WithTenantID(context.Background(), t2),
		func(ctx context.Context, db database.TenantDB) error {
			rep, err := store.Begin(ctx, db, "actor-1", "expired-2", "h", time.Hour)
			if err != nil {
				return err
			}
			if rep.Found {
				t.Fatal("a swept expired key must not replay")
			}
			return nil
		}); err != nil {
		t.Fatalf("post-sweep expired check: %v", err)
	}
}

// TestIntegrationIdempotencyReplayAfterExpiryErrors is the S5 acceptance
// regression (roadmap CA-8): a request presenting an idempotency key whose row
// has expired but is still present must receive a DEFINED error
// (KindIdempotencyExpired → 410), NOT a silent re-execution. Before the fix the
// expired branch re-claimed the key as Fresh and the operation ran a second
// time.
func TestIntegrationIdempotencyReplayAfterExpiryErrors(t *testing.T) {
	h := NewDB(t)
	store := database.NewIdemStore()
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)
	const scope, key, hash = "actor-1", "expired-present", "hash-abc"

	// Seed a completed key with a negative TTL: the row is present but already
	// past expiry, and NOT swept.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		rep, err := store.Begin(ctx, db, scope, key, hash, -time.Hour)
		if err != nil {
			return err
		}
		if !rep.Fresh {
			t.Fatal("seed Begin should be Fresh")
		}
		return store.Complete(ctx, db, scope, key, 201, []byte(`{"id":"x"}`))
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// A replay of the same key while the expired row is still present must fail
	// closed with the defined error — never Fresh (which would re-execute) and
	// never Found (the aged-out response is not replayable).
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		rep, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
		if err != nil {
			return err
		}
		t.Fatalf("expired-but-present key must error, got replay %+v", rep)
		return nil
	})
	if errors.KindOf(err) != errors.KindIdempotencyExpired {
		t.Fatalf("replay after expiry should be KindIdempotencyExpired (410), got %v", err)
	}
}

// TestIntegrationIdempotencyInFlight proves a key claimed but not completed is
// reported in-flight to a concurrent request (retry_later / 409).
func TestIntegrationIdempotencyInFlight(t *testing.T) {
	h := NewDB(t)
	store := database.NewIdemStore()
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)
	const scope, key, hash = "actor-1", "inflight-key", "hash-1"

	// Claim in one committed tx but never complete.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		rep, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
		if err != nil {
			return err
		}
		if !rep.Fresh {
			t.Fatal("expected fresh claim")
		}
		return nil // commit an in_progress row
	}); err != nil {
		t.Fatalf("claim: %v", err)
	}

	// A later request with the same key sees it still in progress.
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
		return err
	})
	if errors.KindOf(err) != errors.KindIdempotencyInFlight {
		t.Fatalf("in-flight duplicate should be KindIdempotencyInFlight, got %v", err)
	}
}

// TestIntegrationIdempotencyConcurrent is the SEC-16/ARCH-27 regression: N
// goroutines fire the SAME key concurrently, each running a full claim →
// operation → complete transaction. Exactly ONE must run the operation; every
// other must either replay the stored response or be told to retry — never a
// second execution.
func TestIntegrationIdempotencyConcurrent(t *testing.T) {
	h := NewDB(t)
	store := database.NewIdemStore()
	tenant := uuid.New()
	const scope, key, hash = "actor-1", "race-key", "hash-1"

	const n = 8
	var (
		mu    sync.Mutex
		ran   int // times the operation body actually executed
		fresh int // times Begin reported Fresh
		wg    sync.WaitGroup
	)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			ctx := database.WithTenantID(context.Background(), tenant)
			_ = h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
				rep, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
				if err != nil {
					return err // conflict / retry_later — acceptable loser outcomes
				}
				if rep.Fresh {
					mu.Lock()
					fresh++
					ran++
					mu.Unlock()
					return store.Complete(ctx, db, scope, key, 201, []byte(`{"id":"x"}`))
				}
				return nil // replayed
			})
		}()
	}
	wg.Wait()

	if ran != 1 {
		t.Fatalf("operation executed %d times under concurrency; must be exactly once (SEC-16)", ran)
	}
	if fresh != 1 {
		t.Fatalf("Begin reported Fresh %d times; must be exactly once", fresh)
	}

	// After the dust settles, a fresh request replays the single stored response.
	if err := h.TxM.WithTenant(database.WithTenantID(context.Background(), tenant),
		func(ctx context.Context, db database.TenantDB) error {
			rep, err := store.Begin(ctx, db, scope, key, hash, time.Hour)
			if err != nil {
				return err
			}
			if !rep.Found || string(rep.ResponseBody) != `{"id":"x"}` {
				t.Fatalf("stored response was lost or corrupted: %+v", rep)
			}
			return nil
		}); err != nil {
		t.Fatalf("post-race replay: %v", err)
	}
}
