package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Replay is the outcome of IdemStore.Begin: either this is the first time the
// key is seen (Fresh), or a completed response is available to replay (Found),
// or the same key is still being processed by a concurrent request (InFlight).
type Replay struct {
	Fresh          bool   // no prior record — proceed with the operation
	Found          bool   // a completed response exists — replay it
	ResponseStatus int    // valid when Found
	ResponseBody   []byte // valid when Found
}

// IdemStore persists idempotency keys and their stored responses, scoped to the
// current tenant via RLS (the table is tenant-scoped). All methods run inside
// the caller's tenant transaction so the key row and the business writes commit
// atomically (blueprint 05 §1–2).
type IdemStore interface {
	// Begin claims the key. It returns Found with the stored response when the
	// key already completed with a MATCHING request hash; a KindConflict error
	// when the hash differs (same key, different request); a
	// KindIdempotencyInFlight error when another request holds the key
	// unfinished; otherwise Fresh (the caller claimed the key and should
	// perform the operation, then call Complete in the same tx).
	Begin(ctx context.Context, db TenantDB, actorScope, key, requestHash string, ttl time.Duration) (Replay, error)
	// Complete records the final response for a key claimed by Begin.
	Complete(ctx context.Context, db TenantDB, actorScope, key string, status int, body []byte) error
	// Discard removes a claim without storing a response — used when the
	// operation did not succeed and should remain retryable (not idempotent).
	Discard(ctx context.Context, db TenantDB, actorScope, key string) error
}

// PgIdemStore is the Postgres-backed IdemStore over idempotency_keys.
type PgIdemStore struct {
	now func() time.Time // injectable clock; defaults to time.Now
}

// NewIdemStore builds a store using the wall clock.
func NewIdemStore() *PgIdemStore { return &PgIdemStore{now: time.Now} }

// NewIdemStoreWithClock builds a store with an injected clock (tests).
func NewIdemStoreWithClock(now func() time.Time) *PgIdemStore { return &PgIdemStore{now: now} }

func (s *PgIdemStore) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *PgIdemStore) Begin(ctx context.Context, db TenantDB, actorScope, key, requestHash string, ttl time.Duration) (Replay, error) {
	now := s.clock()

	// Atomic claim: INSERT … ON CONFLICT DO NOTHING. A returned row means WE
	// inserted it and are the sole owner — this is the only path that yields
	// Fresh from an insert, so concurrent same-key requests cannot both claim
	// (review findings SEC-16/ARCH-27; the previous "SELECT FOR UPDATE then
	// upsert" raced because FOR UPDATE cannot lock a not-yet-existing row).
	var claimed bool
	err := db.QueryRow(ctx,
		`INSERT INTO idempotency_keys (tenant_id, actor_scope, idem_key, request_hash, status, expires_at)
              VALUES (app_tenant_id(), $1, $2, $3, 'in_progress', $4)
         ON CONFLICT (tenant_id, actor_scope, idem_key) DO NOTHING
         RETURNING true`,
		actorScope, key, requestHash, now.Add(ttl)).Scan(&claimed)
	if err == nil && claimed {
		return Replay{Fresh: true}, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Replay{}, kerr.Wrapf(err, "IdemStore.Begin", "claim idempotency key")
	}

	// A row already exists. Lock it and decide. The lock also blocks until any
	// concurrent claiming transaction commits or rolls back, so we observe its
	// final state (completed → replay; rolled back → the row is gone and we
	// re-claim below).
	var (
		existingHash string
		status       string
		respStatus   *int
		respBody     []byte
		expiresAt    time.Time
	)
	err = db.QueryRow(ctx,
		`SELECT request_hash, status, response_status, response_body, expires_at
           FROM idempotency_keys
          WHERE actor_scope = $1 AND idem_key = $2
          FOR UPDATE`,
		actorScope, key).Scan(&existingHash, &status, &respStatus, &respBody, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// The conflicting claimant rolled back (or the row was swept) between
		// our insert and this read. Ask the caller to retry rather than racing.
		return Replay{}, kerr.E(kerr.KindIdempotencyInFlight, "retry_later",
			"idempotency key contended; retry")
	}
	if err != nil {
		return Replay{}, kerr.Wrapf(err, "IdemStore.Begin", "read idempotency key")
	}

	// Expired row → fail closed with a defined error rather than silently
	// re-executing the operation (roadmap S5). The stored response has aged out
	// past its TTL, so we can neither replay it nor safely assume the caller
	// wants the side effects run a second time under the same key. The client
	// must mint a fresh idempotency key to perform the operation again.
	//
	// Note: this guarantee holds while the expired row is still present. Once
	// SweepExpired removes it, an identical key is indistinguishable from a new
	// one and yields Fresh; size the sweep horizon to exceed the client retry
	// window so the defined-error window is meaningful.
	if !expiresAt.After(now) {
		return Replay{}, kerr.E(kerr.KindIdempotencyExpired, "idempotency_key_expired",
			"idempotency key has expired; retry with a new key")
	}

	if existingHash != requestHash {
		return Replay{}, kerr.E(kerr.KindConflict, "conflict",
			"idempotency key reused with a different request")
	}
	if status == "completed" {
		r := Replay{Found: true, ResponseBody: respBody}
		if respStatus != nil {
			r.ResponseStatus = *respStatus
		}
		return r, nil
	}
	// A live, un-expired, still-in_progress row owned by a concurrent request.
	return Replay{}, kerr.E(kerr.KindIdempotencyInFlight, "retry_later",
		"a request with this idempotency key is still being processed")
}

// SweepExpired deletes every idempotency key whose expires_at has passed, across
// ALL tenants, in a single platform transaction (roadmap S5). It runs as
// app_platform via TxManager.Platform — the tenant-scoped app_rt lifecycle is
// unchanged; only this cross-tenant maintenance path may purge other tenants'
// rows (migration 00012). Returns the number of rows removed. Safe alongside
// request traffic: DELETE takes row locks, so a key still held by a live claim
// blocks until that request commits rather than vanishing under it. It is
// scheduled on the leader-safe recurring scheduler at boot
// (app/maintenance.go registers "kernel.idempotency.sweep").
func (s *PgIdemStore) SweepExpired(ctx context.Context, plat TxManager, before time.Time) (int64, error) {
	var n int64
	err := plat.Platform(ctx, func(ctx context.Context, db DB) error {
		tag, err := db.Exec(ctx,
			`DELETE FROM idempotency_keys WHERE expires_at <= $1`, before)
		if err != nil {
			return kerr.Wrapf(err, "IdemStore.SweepExpired", "delete expired keys")
		}
		n = tag.RowsAffected()
		return nil
	})
	return n, err
}

// Discard removes an in_progress claim so the operation stays retryable.
func (s *PgIdemStore) Discard(ctx context.Context, db TenantDB, actorScope, key string) error {
	if _, err := db.Exec(ctx,
		`DELETE FROM idempotency_keys WHERE actor_scope = $1 AND idem_key = $2 AND status = 'in_progress'`,
		actorScope, key); err != nil {
		return kerr.Wrapf(err, "IdemStore.Discard", "discard idempotency claim")
	}
	return nil
}

func (s *PgIdemStore) Complete(ctx context.Context, db TenantDB, actorScope, key string, status int, body []byte) error {
	tag, err := db.Exec(ctx,
		`UPDATE idempotency_keys
            SET status = 'completed', response_status = $1, response_body = $2
          WHERE actor_scope = $3 AND idem_key = $4`,
		status, body, actorScope, key)
	if err != nil {
		return kerr.Wrapf(err, "IdemStore.Complete", "store idempotent response")
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindInternal, "internal", "idempotency key vanished before completion")
	}
	return nil
}
