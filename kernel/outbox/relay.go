package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// Relay claims pending outbox events and dispatches them to registered handlers.
// It reads across tenants on a platform-privileged pool (app_platform — the
// relay RLS policy admits all rows, D-0048), then RE-ENTERS a tenant transaction
// bound to each event's tenant to run handlers under normal tenant RLS and the
// processed_events inbox.
//
// Ordering is per-aggregate (blueprint 07 §7): the claim only picks the earliest
// still-undispatched event for each (tenant, resource), so a later event never
// overtakes an earlier pending/failed one, and a transaction-scoped advisory
// lock keyed on the aggregate serializes concurrent relays. Handlers get an
// exactly-once DB EFFECT via the inbox (dedup + effect share the tenant tx);
// external side effects in a handler are still at-least-once. A poison event
// dead-letters ('dead') after max_attempts rather than retrying forever.
type Relay struct {
	pool     *pgxpool.Pool      // app_platform pool for cross-tenant claim
	txm      database.TxManager // tenant tx manager for handler dispatch
	registry *HandlerRegistry
	batch    int
}

// NewRelay builds the relay. pool must authenticate as the relay role
// (app_platform); txm runs handler transactions per tenant.
func NewRelay(pool *pgxpool.Pool, txm database.TxManager, registry *HandlerRegistry, batchSize int) *Relay {
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Relay{pool: pool, txm: txm, registry: registry, batch: batchSize}
}

// row is a claimed outbox event.
type row struct {
	id       uuid.UUID
	tenant   uuid.UUID
	evType   string
	schemaV  int
	resType  *string
	resID    *uuid.UUID
	actor    []byte
	payload  []byte
	attempts int
}

// DispatchOnce claims up to batch pending events (FOR UPDATE SKIP LOCKED, so
// concurrent relays never double-claim), dispatches each to its handlers, and
// marks it dispatched (or failed, to retry later). It returns the number of
// events processed. A relay loop calls this until it returns 0, then sleeps.
func (r *Relay) DispatchOnce(ctx context.Context) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "begin claim tx")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Per-aggregate ordering (blueprint 07 §7): only claim an event if it is the
	// EARLIEST still-undispatched event for its aggregate — a later event for the
	// same (tenant, resource) is never dispatched while an earlier one is pending
	// or failed, so a transient failure on the older event cannot let the newer
	// one overtake it. Events with no resource (aggregate-less) are unordered.
	// FOR UPDATE SKIP LOCKED keeps concurrent relays from double-claiming; the
	// dispatch tx additionally takes a per-aggregate advisory lock so two relays
	// never process the same aggregate concurrently.
	rows, err := tx.Query(ctx,
		`SELECT e.id, e.tenant_id, e.event_type, e.schema_version, e.resource_type, e.resource_id, e.actor, e.payload, e.attempts
           FROM events_outbox e
          WHERE e.dispatch_status = 'pending'
            AND NOT EXISTS (
                SELECT 1 FROM events_outbox p
                 WHERE p.tenant_id = e.tenant_id
                   AND p.resource_type IS NOT DISTINCT FROM e.resource_type
                   AND p.resource_id IS NOT DISTINCT FROM e.resource_id
                   AND e.resource_id IS NOT NULL
                   AND p.dispatch_status IN ('pending','failed')
                   AND (p.occurred_at, p.id) < (e.occurred_at, e.id))
          ORDER BY e.occurred_at, e.id
          FOR UPDATE SKIP LOCKED
          LIMIT $1`, r.batch)
	if err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "claim events")
	}
	var claimed []row
	for rows.Next() {
		var rw row
		if err := rows.Scan(&rw.id, &rw.tenant, &rw.evType, &rw.schemaV, &rw.resType, &rw.resID, &rw.actor, &rw.payload, &rw.attempts); err != nil {
			rows.Close()
			return 0, kerr.Wrapf(err, "relay.DispatchOnce", "scan event")
		}
		claimed = append(claimed, rw)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "iterate events")
	}

	for _, rw := range claimed {
		if derr := r.dispatch(ctx, rw); derr != nil {
			// Fail with attempt increment. On reaching max_attempts the event
			// is dead-lettered ('dead') instead of retrying forever (poison
			// ceiling, review finding ARCH-54); failed_at drives the cooldown
			// (ARCH-55). last_error is a bounded, non-secret message.
			if _, mErr := tx.Exec(ctx,
				`UPDATE events_outbox
                    SET attempts = attempts + 1,
                        failed_at = now(),
                        last_error = left($2, 500),
                        dispatch_status = CASE WHEN attempts + 1 >= max_attempts THEN 'dead' ELSE 'failed' END
                  WHERE id = $1`,
				rw.id, derr.Error()); mErr != nil {
				return 0, kerr.Wrapf(mErr, "relay.DispatchOnce", "mark failed")
			}
			continue
		}
		if _, err := tx.Exec(ctx,
			`UPDATE events_outbox SET dispatch_status = 'dispatched', dispatched_at = now() WHERE id = $1`,
			rw.id); err != nil {
			return 0, kerr.Wrapf(err, "relay.DispatchOnce", "mark dispatched")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "commit")
	}
	return len(claimed), nil
}

// dispatch runs every handler for an event within a tenant transaction bound to
// the event's tenant, deduped by the inbox. All handlers for one event share the
// tenant tx so a handler failure retries the whole event (at-least-once with
// idempotent handlers = effectively-once).
func (r *Relay) dispatch(ctx context.Context, rw row) error {
	subs := r.registry.handlersFor(rw.evType)
	if len(subs) == 0 {
		return nil // no subscribers — dispatched is a no-op
	}
	de := DispatchedEvent{
		ID: rw.id, Type: rw.evType, SchemaVersion: rw.schemaV,
		Actor: json.RawMessage(rw.actor), Payload: json.RawMessage(rw.payload), TenantID: rw.tenant,
	}
	if rw.resType != nil && rw.resID != nil {
		de.Resource = resource.Ref{Type: *rw.resType, ID: *rw.resID}
	}

	tctx := database.WithTenantID(ctx, rw.tenant)
	return r.txm.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		// Serialize dispatch per aggregate across concurrent relays: a
		// transaction-scoped advisory lock keyed on (tenant, resource) means two
		// relays never run handlers for the same aggregate at once, preserving
		// per-aggregate order under horizontal scale-out (07 §3, ARCH-53).
		if !de.Resource.IsZero() {
			// Key on tenant|type|id with a '/' separator (never a NUL — Postgres
			// text rejects it). hashtextextended → a stable bigint lock key.
			if _, err := db.Exec(ctx,
				`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
				rw.tenant.String()+"/"+de.Resource.Type+"/"+de.Resource.ID.String()); err != nil {
				return kerr.Wrapf(err, "relay.dispatch", "aggregate lock")
			}
		}
		for _, s := range subs {
			// Inbox dedup: claim (handler, event_id); if already present this
			// handler already ran for this event — skip.
			var inserted bool
			err := db.QueryRow(ctx,
				`INSERT INTO processed_events (handler, event_id, tenant_id)
                      VALUES ($1, $2, app_tenant_id())
                 ON CONFLICT (handler, event_id) DO NOTHING
                 RETURNING true`, s.name, rw.id).Scan(&inserted)
			if errors.Is(err, pgx.ErrNoRows) {
				continue // already processed by this handler
			}
			if err != nil {
				return kerr.Wrapf(err, "relay.dispatch", "inbox claim for %s", s.name)
			}
			if err := s.fn(ctx, db, de); err != nil {
				return kerr.Wrapf(err, "relay.dispatch", "handler %s", s.name)
			}
		}
		return nil
	})
}

// Run drives the relay until ctx is cancelled: dispatch batches back-to-back
// while there is work, then poll on the interval. Failed events (marked
// 'failed') are re-claimed by resetting them to pending after a cooldown — see
// RequeueFailed. Run returns nil on clean cancellation.
func (r *Relay) Run(ctx context.Context, poll time.Duration) error {
	if poll <= 0 {
		poll = time.Second
	}
	t := time.NewTicker(poll)
	defer t.Stop()
	for {
		n, err := r.DispatchOnce(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if n > 0 {
			continue // drain
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			_ = r.RequeueFailed(ctx, 30*time.Second)
		}
	}
}

// RequeueFailed resets 'failed' events (not 'dead' — those are terminal) back to
// 'pending' once the last failure is older than cooldown. The cooldown is keyed
// on failed_at (the actual failure time), not occurred_at (write time), so a
// just-failed event actually waits (review finding ARCH-55). Dead-lettered
// events are left for the admin requeue path.
func (r *Relay) RequeueFailed(ctx context.Context, cooldown time.Duration) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE events_outbox SET dispatch_status = 'pending'
          WHERE dispatch_status = 'failed'
            AND failed_at IS NOT NULL
            AND failed_at < now() - make_interval(secs => $1)`,
		cooldown.Seconds())
	if err != nil {
		return kerr.Wrapf(err, "relay.RequeueFailed", "requeue")
	}
	return nil
}
