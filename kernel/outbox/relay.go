package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/lease"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
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
const defaultRelayLeaseTTL = 30 * time.Second

var (
	outboxRelayMetricLabels = map[string]string{"worker": "outbox_relay"}
	// ErrLeaseMismatch means a stale relay attempted to finalize an event after
	// another worker reclaimed its expired W04 DATA-02 lease epoch.
	ErrLeaseMismatch = kerr.E(kerr.KindConflict, "lease_mismatch", "stale outbox finalize rejected")
)

type Relay struct {
	pool     *pgxpool.Pool      // app_platform pool for cross-tenant claim/finalize
	txm      database.TxManager // tenant tx manager for one event's handlers
	registry *HandlerRegistry
	batch    int
	leaseTTL time.Duration
	tracer   observability.Tracer
	metrics  observability.Metrics
	// hooks overrides RequeueFailed/DispatchOnce in tests (fault injection for
	// the F-07 recovery-observability regression); nil means the real
	// implementations. A pointer (not func fields) so Relay stays comparable —
	// the Go API compatibility gate guards that property of the v1 surface.
	hooks *relayTestHooks
}

// relayTestHooks carries test-only fault-injection seams for Relay.Run.
type relayTestHooks struct {
	requeue  func(ctx context.Context, cooldown time.Duration) error
	dispatch func(ctx context.Context) (int, error)
}

// RelayOption customizes the relay.
type RelayOption func(*Relay)

// WithRelayTracer wires a tracer so the relay continues the originating request's
// trace when it dispatches an event (roadmap O1/CA-9): it extracts the event's
// stored traceparent and runs the handler under a child span. Default: NoOpTracer.
func WithRelayTracer(tr observability.Tracer) RelayOption {
	return func(r *Relay) {
		if tr != nil {
			r.tracer = tr
		}
	}
}

// WithRelayLeaseTTL changes the W04 shared-primitive lease lifetime. It is
// primarily useful for deterministic crash/reclaim tests.
func WithRelayLeaseTTL(ttl time.Duration) RelayOption {
	return func(r *Relay) {
		if ttl > 0 {
			r.leaseTTL = ttl
		}
	}
}

// WithRelayMetrics wires bounded-cardinality queue lag and batch duration
// gauges. A nil sink leaves the safe NoOp default.
func WithRelayMetrics(metrics observability.Metrics) RelayOption {
	return func(r *Relay) {
		if metrics != nil {
			r.metrics = metrics
		}
	}
}

// NewRelay builds the relay. pool must authenticate as the relay role
// (app_platform); txm runs handler transactions per tenant.
func NewRelay(pool *pgxpool.Pool, txm database.TxManager, registry *HandlerRegistry, batchSize int, opts ...RelayOption) *Relay {
	if batchSize <= 0 {
		batchSize = 100
	}
	r := &Relay{
		pool: pool, txm: txm, registry: registry, batch: batchSize,
		leaseTTL: defaultRelayLeaseTTL, tracer: observability.NoOpTracer, metrics: observability.NoOp,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// row is a claimed outbox event.
type row struct {
	id         uuid.UUID
	tenant     uuid.UUID
	evType     string
	schemaV    int
	resType    *string
	resID      *uuid.UUID
	actor      []byte
	payload    []byte
	attempts   int
	occurredAt time.Time
	trace      *string // W3C traceparent captured at write time (CA-9); nil when absent
	lease      lease.Lease
}

// DispatchOnce claims up to batch pending events (FOR UPDATE SKIP LOCKED, so
// concurrent relays never double-claim), dispatches each to its handlers, and
// marks it dispatched (or failed, to retry later). It returns the number of
// events processed. A relay loop calls this until it returns 0, then sleeps.
func (r *Relay) DispatchOnce(ctx context.Context) (int, error) {
	started := time.Now()
	maxLag := time.Duration(0)
	defer func() {
		r.metrics.SetGauge("worker_queue_lag_seconds", maxLag.Seconds(), outboxRelayMetricLabels)
		observability.ObserveHistogram(r.metrics, "worker_batch_duration_seconds", time.Since(started).Seconds(), outboxRelayMetricLabels)
	}()

	// Stage 1 — claim and commit. The W04 DATA-02 Lease value is persisted as
	// token+generation+expiry. No transaction from this stage survives into a
	// tenant handler or other remote/consumer work.
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "begin claim tx")
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx,
		`SELECT e.id, e.tenant_id, e.event_type, e.schema_version,
		        e.resource_type, e.resource_id, e.actor, e.payload, e.attempts,
		        e.occurred_at, e.trace_context, e.lease_generation
		   FROM events_outbox e
		  WHERE e.dispatch_status = 'pending'
		    AND (e.lease_token IS NULL OR e.lease_expires_at <= now())
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
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "select claimable events")
	}
	claimed := make([]row, 0, r.batch)
	for rows.Next() {
		var rw row
		var generation int64
		if err := rows.Scan(
			&rw.id, &rw.tenant, &rw.evType, &rw.schemaV, &rw.resType, &rw.resID,
			&rw.actor, &rw.payload, &rw.attempts, &rw.occurredAt, &rw.trace, &generation,
		); err != nil {
			rows.Close()
			return 0, kerr.Wrapf(err, "relay.DispatchOnce", "scan claimable event")
		}
		if generation == 0 {
			rw.lease = lease.New(r.leaseTTL)
		} else {
			// NextEpoch is the accepted W04 primitive's reclaim operation: a
			// fresh opaque token and strictly increasing generation.
			rw.lease = (lease.Lease{Generation: generation}).NextEpoch(r.leaseTTL)
		}
		if lag := started.Sub(rw.occurredAt); lag > maxLag {
			maxLag = lag
		}
		claimed = append(claimed, rw)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "iterate claimable events")
	}
	if len(claimed) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return 0, kerr.Wrapf(err, "relay.DispatchOnce", "commit empty claim")
		}
		return 0, nil
	}
	ids := make([]uuid.UUID, len(claimed))
	tokens := make([]string, len(claimed))
	generations := make([]int64, len(claimed))
	expiries := make([]time.Time, len(claimed))
	for i := range claimed {
		ids[i] = claimed[i].id
		tokens[i] = claimed[i].lease.Token
		generations[i] = claimed[i].lease.Generation
		expiries[i] = claimed[i].lease.ExpiresAt
	}
	tag, err := tx.Exec(ctx, `UPDATE events_outbox AS event
		SET lease_token = claimed.token,
		    lease_generation = claimed.generation,
		    lease_expires_at = claimed.expires_at
		FROM unnest($1::uuid[], $2::text[], $3::bigint[], $4::timestamptz[])
		     AS claimed(id, token, generation, expires_at)
		WHERE event.id = claimed.id
		  AND event.dispatch_status = 'pending'
		  AND (event.lease_token IS NULL OR event.lease_expires_at <= now())`,
		ids, tokens, generations, expiries)
	if err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "persist leases")
	}
	if tag.RowsAffected() != int64(len(claimed)) {
		return 0, kerr.E(kerr.KindConflict, "outbox_claim_changed", "outbox claim changed while locked")
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, kerr.Wrapf(err, "relay.DispatchOnce", "commit claims")
	}

	// Stage 2 — tenant handler transaction(s), with no claim transaction open.
	// Stage 3 — a short fenced platform finalize. A crash between stages leaves
	// a reclaimable expired lease; processed_events dedups a repeated handler.
	for _, rw := range claimed {
		if dispatchErr := r.dispatch(ctx, rw); dispatchErr != nil {
			if err := r.finalizeFailure(ctx, rw, dispatchErr); err != nil {
				return len(claimed), err
			}
			continue
		}
		if err := r.finalizeSuccess(ctx, rw); err != nil {
			return len(claimed), err
		}
	}
	return len(claimed), nil
}

func (r *Relay) finalizeSuccess(ctx context.Context, rw row) error {
	tag, err := r.pool.Exec(ctx, `UPDATE events_outbox
		SET dispatch_status = 'dispatched',
		    dispatched_at = now(),
		    lease_token = NULL,
		    lease_expires_at = NULL
		WHERE id = $1
		  AND dispatch_status = 'pending'
		  AND lease_token = $2
		  AND lease_generation = $3
		  AND lease_expires_at > now()`,
		rw.id, rw.lease.Token, rw.lease.Generation)
	if err != nil {
		return kerr.Wrapf(err, "relay.DispatchOnce", "finalize dispatched")
	}
	if tag.RowsAffected() == 0 {
		return ErrLeaseMismatch
	}
	return nil
}

func (r *Relay) finalizeFailure(ctx context.Context, rw row, dispatchErr error) error {
	tag, err := r.pool.Exec(ctx, `UPDATE events_outbox
		SET attempts = attempts + 1,
		    failed_at = now(),
		    last_error = left($2, 500),
		    dispatch_status = CASE WHEN attempts + 1 >= max_attempts THEN 'dead' ELSE 'failed' END,
		    lease_token = NULL,
		    lease_expires_at = NULL
		WHERE id = $1
		  AND dispatch_status = 'pending'
		  AND lease_token = $3
		  AND lease_generation = $4
		  AND lease_expires_at > now()`,
		rw.id, dispatchErr.Error(), rw.lease.Token, rw.lease.Generation)
	if err != nil {
		return kerr.Wrapf(err, "relay.DispatchOnce", "finalize failed")
	}
	if tag.RowsAffected() == 0 {
		return ErrLeaseMismatch
	}
	return nil
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
	// Continue the originating request's trace across the async boundary
	// (roadmap O1/CA-9): extract the traceparent stored at write time, then run
	// the dispatch under a child span. Zero-cost with NoOpTracer.
	if rw.trace != nil && *rw.trace != "" {
		ctx = r.tracer.Extract(ctx, *rw.trace)
	}
	ctx, span := r.tracer.StartSpan(ctx, "outbox.dispatch "+rw.evType)
	span.SetAttr("event.type", rw.evType)
	span.SetAttr("event.id", rw.id.String())
	defer span.End()
	de := DispatchedEvent{
		ID: rw.id, Type: rw.evType, SchemaVersion: rw.schemaV,
		Actor: json.RawMessage(rw.actor), Payload: json.RawMessage(rw.payload), TenantID: rw.tenant,
	}
	if rw.resType != nil && rw.resID != nil {
		de.Resource = resource.Ref{Type: *rw.resType, ID: *rw.resID}
	}

	tctx := database.WithTenantID(ctx, rw.tenant)
	derr := r.txm.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
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
	if derr != nil {
		span.RecordError(derr)
	}
	return derr
}

// relayRequeueMaxConsecutiveFailures bounds silent retry of failed-event
// maintenance: each failure increments outbox_requeue_errors_total, and after
// this many CONSECUTIVE failures Run returns the error so the process
// supervisor restarts the relay instead of hiding a persistent grant/schema
// problem forever (adversarial review 2026-07-17, F-07).
const relayRequeueMaxConsecutiveFailures = 5

// Run drives the relay until ctx is cancelled: dispatch batches back-to-back
// while there is work, then poll on the interval. Failed events (marked
// 'failed') are re-claimed by resetting them to pending after a cooldown — see
// RequeueFailed. Recovery runs on its OWN due schedule, even while the drain
// loop is busy: under sustained pending traffic the idle branch may never
// execute, and a failed predecessor would otherwise starve behind unrelated
// events (F-07). A requeue error is counted on outbox_requeue_errors_total and,
// after relayRequeueMaxConsecutiveFailures consecutive failures, returned. Run
// returns nil on clean cancellation.
func (r *Relay) Run(ctx context.Context, poll time.Duration) error {
	if poll <= 0 {
		poll = time.Second
	}
	requeue, dispatchOnce := r.RequeueFailed, r.DispatchOnce
	if r.hooks != nil {
		if r.hooks.requeue != nil {
			requeue = r.hooks.requeue
		}
		if r.hooks.dispatch != nil {
			dispatchOnce = r.hooks.dispatch
		}
	}
	var (
		nextRequeue     time.Time // zero: due immediately
		requeueFailures int
	)
	maybeRequeue := func() error {
		if time.Now().Before(nextRequeue) {
			return nil
		}
		nextRequeue = time.Now().Add(poll)
		if err := requeue(ctx, 30*time.Second); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			requeueFailures++
			r.metrics.IncCounter("outbox_requeue_errors_total", 1, outboxRelayMetricLabels)
			if requeueFailures >= relayRequeueMaxConsecutiveFailures {
				return kerr.Wrapf(err, "relay.Run", "failed-event requeue failed %d consecutive times", requeueFailures)
			}
			return nil
		}
		requeueFailures = 0
		return nil
	}
	t := time.NewTicker(poll)
	defer t.Stop()
	for {
		if err := maybeRequeue(); err != nil {
			return err
		}
		n, err := dispatchOnce(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if n > 0 {
			continue // drain; maybeRequeue above keeps recovery on schedule
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
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
