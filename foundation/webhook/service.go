package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/lease"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/retry"
)

var webhookBackoff = retry.NewSchedule(retry.NewSequenceBackOff(
	time.Second,
	5*time.Second,
	30*time.Second,
	2*time.Minute,
	5*time.Minute,
))

var webhookRetryMetricLabels = map[string]string{"worker": "webhook_retry"}

// leaseTTL is the duration a claimed outbound webhook delivery row is fenced
// from concurrent re-claim (C-1 / W04-E02-S001 AC-02/03, DATA-03 T1). It must
// cover the effect stage (secret resolve + Sender.Post, bounded by
// OutboundTimeout) plus the finalize round-trip; mirrors foundation/notify's
// leaseTTL.
const leaseTTL = 5 * time.Minute

// HandleInbound verifies, replay-checks, and persists an inbound webhook event.
// Runs inside the caller's app_rt tenant transaction (db). On success returns
// nil (ack fast); actual processing is async via ProcessInbound.
//
//  1. Look up the Verifier for in.ProviderKey; resolve the endpoint secret.
//  2. Verify signature — on failure insert a signature_ok=false audit row
//     (best-effort in the same tx) and return KindUnauthenticated.
//  3. Reject timestamps outside ±5 m of now → KindValidation.
//  4. Insert a pending row; UNIQUE(endpoint_id, external_event_id) detects
//     replays → KindConflict (idempotent ack).
func (s *Service) HandleInbound(ctx context.Context, db database.TenantDB, in InboundIn) error {
	ep, err := s.loadEndpoint(ctx, db, in.EndpointID)
	if err != nil {
		return err
	}
	if ep.Direction != DirectionInbound {
		return kerr.E(kerr.KindValidation, "endpoint_direction", "endpoint is not inbound")
	}
	if ep.Status != "active" {
		return kerr.E(kerr.KindConflict, "endpoint_inactive", "webhook endpoint is not active")
	}

	secret, err := s.secrets.Resolve(ctx, ep.SecretRef)
	if err != nil {
		return kerr.Wrapf(err, "webhook.HandleInbound", "resolve secret")
	}

	v, ok := s.verifiers[in.ProviderKey]
	if !ok {
		return kerr.E(kerr.KindValidation, "no_verifier", "no verifier registered for provider: "+in.ProviderKey)
	}

	env, verr := v.Verify(secret, in.RawBody, in.Headers)
	if verr != nil {
		// Audit row: signature_ok=false, empty payload (NO body logging). This
		// is BEST-EFFORT: it persists only if the HTTP layer commits the tenant
		// tx before returning 401 (ARCH-76). The 401 response and metric are the
		// primary signals (blueprint 07 §6).
		//
		// external_event_id is forced NULL here (SEC-50): a spoofed unsigned
		// request must never pre-claim a legitimate event's dedup slot and block
		// it. With the partial unique index (WHERE external_event_id IS NOT
		// NULL), a NULL audit row never participates in dedup.
		sigFalse := false
		_ = s.insertEvent(ctx, db, insertEventParams{
			endpointID:      in.EndpointID,
			direction:       DirectionInbound,
			externalEventID: nil,
			eventType:       in.EventType,
			payload:         json.RawMessage(`{}`),
			signatureOk:     &sigFalse,
			deliveryStatus:  StatusDead,
			receivedAt:      s.now(),
		})
		return kerr.E(kerr.KindUnauthenticated, "signature_invalid", "webhook signature verification failed")
	}

	// Timestamp window ±5 m, using the verifier-attested occurred time.
	delta := s.now().Sub(env.OccurredAt)
	if delta < 0 {
		delta = -delta
	}
	if delta > TimestampWindow {
		return kerr.E(kerr.KindValidation, "timestamp_out_of_window", "webhook timestamp is outside the ±5 m replay window")
	}

	payload := json.RawMessage(`{}`)
	if len(env.CanonicalBody) > 0 && json.Valid(env.CanonicalBody) {
		payload = json.RawMessage(env.CanonicalBody)
	}

	// Replay-dedup id: the verifier-derived event id when present, else a stable
	// synthetic id derived from the authenticated body (SEC-49/ARCH-74), so an
	// id-less event is still replay-protected under the partial unique index.
	sigTrue := true
	dedup := dedupExtIDEnv(env)
	inserted, err := s.insertEventOnConflictIgnore(ctx, db, insertEventParams{
		endpointID:      in.EndpointID,
		direction:       DirectionInbound,
		externalEventID: &dedup,
		eventType:       in.EventType,
		payload:         payload,
		signatureOk:     &sigTrue,
		deliveryStatus:  StatusPending,
		receivedAt:      s.now(),
	})
	if err != nil {
		return kerr.Wrapf(err, "webhook.HandleInbound", "persist event")
	}
	if !inserted {
		return kerr.E(kerr.KindConflict, "event_duplicate", "webhook event already received")
	}
	return nil
}

// ProcessInbound claims pending inbound events for tenantID and runs registered
// handlers. Runs as app_platform (plat is a TxManager over the platform pool).
// Advances delivery_status: pending/failed → processed on success, or increments
// attempts toward dead on handler error.
func (s *Service) ProcessInbound(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, now time.Time) error {
	return plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx,
			`SELECT id, endpoint_id, event_type, payload, attempts
			   FROM webhook_events
			  WHERE direction = 'inbound'
			    AND delivery_status IN ('pending','failed')
			    AND (next_attempt_at IS NULL OR next_attempt_at <= $1)
			  ORDER BY received_at
			  LIMIT 10
			    FOR UPDATE SKIP LOCKED`,
			now)
		if err != nil {
			return kerr.Wrapf(err, "webhook.ProcessInbound", "claim events")
		}
		type claimedRow struct {
			id         uuid.UUID
			endpointID uuid.UUID
			eventType  string
			payload    json.RawMessage
			attempts   int
		}
		var claimed []claimedRow
		for rows.Next() {
			var r claimedRow
			if err := rows.Scan(&r.id, &r.endpointID, &r.eventType, &r.payload, &r.attempts); err != nil {
				rows.Close()
				return kerr.Wrapf(err, "webhook.ProcessInbound", "scan event")
			}
			claimed = append(claimed, r)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "webhook.ProcessInbound", "iterate events")
		}

		for _, r := range claimed {
			ev := Event{
				ID:         r.id,
				TenantID:   tenantID,
				EndpointID: r.endpointID,
				Direction:  DirectionInbound,
				EventType:  r.eventType,
				Payload:    r.payload,
				Attempts:   r.attempts,
			}
			herr := s.runInboundHandler(ctx, db, ev)
			next := r.attempts + 1
			if herr != nil {
				msg := truncate(herr.Error())
				if next >= MaxAttempts {
					if _, uerr := db.Exec(ctx,
						`UPDATE webhook_events
						    SET delivery_status = 'dead', attempts = $2, last_error = $3
						  WHERE id = $1`,
						r.id, next, msg); uerr != nil {
						return kerr.Wrapf(uerr, "webhook.ProcessInbound", "dead-letter event")
					}
				} else {
					nextAt := now.Add(webhookBackoff.Next(next))
					if _, uerr := db.Exec(ctx,
						`UPDATE webhook_events
						    SET delivery_status = 'failed', attempts = $2,
						        next_attempt_at = $3, last_error = $4
						  WHERE id = $1`,
						r.id, next, nextAt, msg); uerr != nil {
						return kerr.Wrapf(uerr, "webhook.ProcessInbound", "mark failed")
					}
				}
				continue
			}
			if _, uerr := db.Exec(ctx,
				`UPDATE webhook_events
				    SET delivery_status = 'processed', attempts = $2, last_error = NULL
				  WHERE id = $1`,
				r.id, next); uerr != nil {
				return kerr.Wrapf(uerr, "webhook.ProcessInbound", "mark processed")
			}
		}
		return nil
	})
}

// DispatchOutbound fans ev to all active (or degraded) outbound endpoints for
// the event's tenant whose subscribed_events contain ev.Type. It runs a
// three-stage claim/effect/finalize protocol (C-1 / W04-E02-S001 AC-02/03,
// mirroring foundation/notify's SendPending):
//
//  1. Claim-tx: upserts a pending delivery row per matching endpoint, checks
//     the circuit breaker and terminal/backoff state, and assigns a fresh
//     lease from kernel/lease to every row eligible for delivery this cycle.
//     Commits.
//  2. Effect stage: for each claimed row, resolves the endpoint secret and
//     signs + POSTs the body via Sender — entirely OUTSIDE any database
//     transaction.
//  3. Finalize-tx: per delivery, updates status only if the lease token and
//     generation still match and the lease has not expired (a mismatch means
//     the row was reclaimed by another worker; the effect result is
//     discarded).
//
// Open-circuit endpoints are skipped silently (their delivery rows stay
// pending, no lease assigned). One endpoint's delivery/finalize failure does
// not block the others — each is finalized independently. Runs as
// app_platform.
//
// SEC/H2: the delivery tenant is authoritative from ev.TenantID, never the
// decoupled tenantID param. Without this, a caller passing B's id with A's event
// would look up B's endpoints and sign A's payload with B's secret — a
// cross-tenant leak. When ev carries a tenant (relay/writer sets it), a
// disagreeing tenantID is rejected fail-closed (KindValidation); the event's
// tenant then binds the whole dispatch. Both identities are mandatory and must
// match; there is no inference or zero-value fallback.
func (s *Service) DispatchOutbound(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, ev outbox.Event, now time.Time) error {
	if tenantID == uuid.Nil || ev.TenantID == uuid.Nil || tenantID != ev.TenantID {
		return kerr.E(kerr.KindValidation, "tenant_mismatch",
			"webhook dispatch requires matching nonzero scope and event tenants")
	}

	body, err := marshalOutboundBody(ev, tenantID)
	if err != nil {
		return kerr.Wrapf(err, "webhook.DispatchOutbound", "marshal body")
	}

	claimed, err := s.claimDispatch(ctx, plat, tenantID, ev, body, now)
	if err != nil {
		return err
	}
	s.deliverClaimed(ctx, plat, tenantID, claimed, now)
	return nil
}

// claimDispatch is DispatchOutbound's claim stage: it loads the matching
// outbound endpoints and, for each, claims a delivery row (see
// claimDeliveryRow) inside a single tenant transaction. A per-endpoint claim
// error is logged and skipped rather than aborting the whole batch, matching
// the pre-existing per-endpoint isolation contract.
func (s *Service) claimDispatch(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, ev outbox.Event, body []byte, now time.Time) ([]claimedOutboundDelivery, error) {
	var out []claimedOutboundDelivery
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		eps, err := s.loadOutboundEndpoints(ctx, db, ev.Type)
		if err != nil {
			return err
		}
		for _, ep := range eps {
			cd, cerr := s.claimDeliveryRow(ctx, db, ep, ev, body, now)
			if cerr != nil {
				slog.ErrorContext(ctx, "webhook.claim_error", "endpoint_id", ep.ID, "err", cerr)
				continue
			}
			if cd != nil {
				out = append(out, *cd)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RetryOutbound re-delivers previously-failed outbound webhook events for
// tenantID whose backoff has elapsed. Runs as app_platform (plat is a
// TxManager over the platform pool), mirroring ProcessInbound and sharing
// DispatchOutbound's claim/effect/finalize protocol (C-1 / W04-E02-S001
// AC-02/03). Without this, DispatchOutbound would leave a 'failed' row
// untouched and the outbox relay — having marked its source event dispatched
// — would never re-drive delivery (ARCH-70). It claims failed, due,
// unleased rows FOR UPDATE SKIP LOCKED, assigns each a fresh lease, then
// delivers/finalizes exactly like DispatchOutbound.
func (s *Service) RetryOutbound(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, now time.Time) error {
	started := time.Now()
	defer func() {
		observability.ObserveHistogram(s.metrics, "worker_batch_duration_seconds", time.Since(started).Seconds(), webhookRetryMetricLabels)
	}()

	claimed, err := s.claimRetry(ctx, plat, tenantID, now)
	if err != nil {
		return err
	}
	s.deliverClaimed(ctx, plat, tenantID, claimed, now)
	return nil
}

// claimRetry is RetryOutbound's claim stage.
func (s *Service) claimRetry(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, now time.Time) ([]claimedOutboundDelivery, error) {
	var out []claimedOutboundDelivery
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx,
			`SELECT id, endpoint_id, external_event_id, event_type, payload, attempts, received_at,
			        lease_token, lease_generation, lease_expires_at
			   FROM webhook_events
			  WHERE direction = 'outbound'
			    AND delivery_status = 'failed'
			    AND (next_attempt_at IS NULL OR next_attempt_at <= $1)
			    AND (lease_expires_at IS NULL OR lease_expires_at <= $1)
			  ORDER BY received_at
			  LIMIT 10
			    FOR UPDATE SKIP LOCKED`,
			now)
		if err != nil {
			return kerr.Wrapf(err, "webhook.claimRetry", "claim events")
		}
		type scannedRow struct {
			id              uuid.UUID
			endpointID      uuid.UUID
			extID           string
			eventType       string
			payload         json.RawMessage
			attempts        int
			receivedAt      time.Time
			leaseToken      *string
			leaseGeneration *int64
			leaseExpiresAt  *time.Time
		}
		var scanned []scannedRow
		endpointIDs := make([]uuid.UUID, 0, 10)
		seenEndpoints := make(map[uuid.UUID]struct{}, 10)
		maxLag := time.Duration(0)
		for rows.Next() {
			var r scannedRow
			if err := rows.Scan(&r.id, &r.endpointID, &r.extID, &r.eventType, &r.payload,
				&r.attempts, &r.receivedAt,
				&r.leaseToken, &r.leaseGeneration, &r.leaseExpiresAt); err != nil {
				rows.Close()
				return kerr.Wrapf(err, "webhook.claimRetry", "scan event")
			}
			scanned = append(scanned, r)
			if _, seen := seenEndpoints[r.endpointID]; !seen {
				seenEndpoints[r.endpointID] = struct{}{}
				endpointIDs = append(endpointIDs, r.endpointID)
			}
			if lag := now.Sub(r.receivedAt); lag > maxLag {
				maxLag = lag
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "webhook.claimRetry", "iterate events")
		}
		s.metrics.SetGauge("worker_queue_lag_seconds", maxLag.Seconds(), webhookRetryMetricLabels)

		endpoints, err := s.loadEndpoints(ctx, db, endpointIDs)
		if err != nil {
			return err
		}

		// Assign leases after closing the cursor so we don't try to run UPDATE
		// while the SELECT cursor is still active on the same connection
		// (mirrors notify.claimPending).
		for _, r := range scanned {
			ep, ok := endpoints[r.endpointID]
			if !ok {
				return kerr.E(kerr.KindNotFound, "not_found", "webhook endpoint not found")
			}
			evID, parseErr := uuid.Parse(r.extID)
			if parseErr != nil {
				// A synthetic/non-UUID id should never appear on an outbound row
				// (they key on the outbox event UUID); skip rather than crash.
				continue
			}

			br := s.breaker.get(ep.ID)
			if !br.allow(s.now()) {
				continue // open — leave row failed, no lease assigned
			}

			lse := nextLease(r.leaseToken, r.leaseGeneration, r.leaseExpiresAt, now)
			if _, err := db.Exec(ctx,
				`UPDATE webhook_events
				    SET lease_token = $2, lease_generation = $3, lease_expires_at = $4
				  WHERE id = $1`,
				r.id, lse.Token, lse.Generation, lse.ExpiresAt); err != nil {
				return kerr.Wrapf(err, "webhook.claimRetry", "assign lease %s", r.id)
			}

			out = append(out, claimedOutboundDelivery{
				rowID:    r.id,
				ep:       ep,
				ev:       outbox.Event{ID: evID, Type: r.eventType},
				body:     r.payload,
				attempts: r.attempts,
				lease:    lse,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// --- internal ---

func (s *Service) runInboundHandler(ctx context.Context, db database.TenantDB, ev Event) error {
	h, ok := s.handlers[ev.EventType]
	if !ok {
		return nil // no handler registered — treat as processed (no-op)
	}
	return h(ctx, db, ev)
}

// claimedOutboundDelivery is a webhook_events outbound row that has been
// assigned a lease in the claim stage and is ready for the effect/finalize
// stages (C-1 / W04-E02-S001 AC-02/03).
type claimedOutboundDelivery struct {
	rowID    uuid.UUID
	ep       Endpoint
	ev       outbox.Event
	body     []byte
	attempts int
	lease    lease.Lease
}

// deliveryResult is the outcome of the effect stage for one claimed delivery.
// secretErr is set when the endpoint secret itself could not be resolved —
// distinct from ok/err/statusCode, which describe the outcome of an actual
// POST attempt. No POST is attempted when secretErr is set.
type deliveryResult struct {
	ok         bool
	statusCode int
	err        error
	secretErr  error
}

// nextLease assigns a fresh lease for a row, reusing the row's current lease
// generation (bumping it) when one already exists, or starting at generation
// zero otherwise. now backs ExpiresAt so tests with a fake clock fence leases
// correctly (mirrors notify.claimPending).
func nextLease(token *string, generation *int64, expiresAt *time.Time, now time.Time) lease.Lease {
	var existing lease.Lease
	if token != nil {
		existing.Token = *token
		existing.Generation = *generation
		existing.ExpiresAt = *expiresAt
	}
	var l lease.Lease
	if existing.Zero() {
		l = lease.New(leaseTTL)
	} else {
		l = existing.NextEpoch(leaseTTL)
	}
	l.ExpiresAt = now.Add(leaseTTL)
	return l
}

// claimDeliveryRow upserts the (idempotent) pending delivery row for ep/ev,
// evaluates eligibility — terminal status, backoff window, an already-live
// lease held by another worker, and the circuit breaker — and, when
// eligible, assigns a fresh lease inside the same transaction. Returns
// (nil, nil) when the row is not eligible for delivery this cycle: it stays
// pending/failed in the DB with no lease assigned, no secret resolved, and
// no network call made. MUST be called from inside an open tenant
// transaction; MUST NOT itself perform any remote I/O (C-1 / W04-E02-S001
// AC-02/03).
func (s *Service) claimDeliveryRow(ctx context.Context, db database.TenantDB, ep Endpoint, ev outbox.Event, body []byte, now time.Time) (*claimedOutboundDelivery, error) {
	extID := ev.ID.String() // outbox event id is the outbound idempotency key
	extIDPtr := &extID

	// Upsert a pending delivery row (idempotent on the outbox event id).
	if _, err := s.insertEventOnConflictIgnore(ctx, db, insertEventParams{
		endpointID:      ep.ID,
		direction:       DirectionOutbound,
		externalEventID: extIDPtr,
		eventType:       ev.Type,
		payload:         json.RawMessage(body),
		signatureOk:     nil,
		deliveryStatus:  StatusPending,
		receivedAt:      now,
	}); err != nil {
		return nil, kerr.Wrapf(err, "webhook.claimDeliveryRow", "upsert delivery row")
	}

	// Load the current row state, locking it against a concurrent claimer.
	var (
		rowID           uuid.UUID
		attempts        int
		status          string
		nextAttemptAt   *time.Time
		leaseToken      *string
		leaseGeneration *int64
		leaseExpiresAt  *time.Time
	)
	err := db.QueryRow(ctx,
		`SELECT id, attempts, delivery_status, next_attempt_at,
		        lease_token, lease_generation, lease_expires_at
		   FROM webhook_events
		  WHERE endpoint_id = $1 AND external_event_id = $2
		    FOR UPDATE SKIP LOCKED`,
		ep.ID, extID).Scan(&rowID, &attempts, &status, &nextAttemptAt,
		&leaseToken, &leaseGeneration, &leaseExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Locked by a concurrent claimer (SKIP LOCKED) — leave it for that
		// worker; not an error.
		return nil, nil
	}
	if err != nil {
		return nil, kerr.Wrapf(err, "webhook.claimDeliveryRow", "load delivery row")
	}

	// Skip terminal or not-yet-due rows.
	if status == StatusDelivered || status == StatusDead {
		return nil, nil
	}
	if status == StatusFailed && nextAttemptAt != nil && nextAttemptAt.After(now) {
		return nil, nil
	}
	// Already leased by a live worker — the row was committed by a prior
	// claim-tx and is mid effect/finalize.
	if leaseExpiresAt != nil && leaseExpiresAt.After(now) {
		return nil, nil
	}

	// Circuit breaker.
	br := s.breaker.get(ep.ID)
	if !br.allow(s.now()) {
		return nil, nil // open — leave row pending, no lease assigned
	}

	lse := nextLease(leaseToken, leaseGeneration, leaseExpiresAt, now)
	if _, err := db.Exec(ctx,
		`UPDATE webhook_events
		    SET lease_token = $2, lease_generation = $3, lease_expires_at = $4
		  WHERE id = $1`,
		rowID, lse.Token, lse.Generation, lse.ExpiresAt); err != nil {
		return nil, kerr.Wrapf(err, "webhook.claimDeliveryRow", "assign lease")
	}

	return &claimedOutboundDelivery{
		rowID:    rowID,
		ep:       ep,
		ev:       ev,
		body:     body,
		attempts: attempts,
		lease:    lse,
	}, nil
}

// deliverClaimed runs the effect stage (secret resolve + signed POST) for
// each claimed delivery entirely OUTSIDE any database transaction, then
// finalizes the outcome in its own short, lease-fenced transaction. A
// finalize error for one delivery is logged and does not stop the batch —
// one endpoint's failure must not block delivery to the others (C-1 /
// W04-E02-S001 AC-02/03).
func (s *Service) deliverClaimed(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, claimed []claimedOutboundDelivery, now time.Time) {
	for _, cd := range claimed {
		res := s.effectDeliver(ctx, cd, now)
		if _, err := s.finalizeOutboundDelivery(ctx, plat, tenantID, cd, res, now); err != nil {
			slog.ErrorContext(ctx, "webhook.finalize_error", "delivery_id", cd.rowID, "endpoint_id", cd.ep.ID, "err", err)
		}
	}
}

// effectDeliver resolves the endpoint secret and performs the signed HTTP
// POST for a claimed delivery. It MUST NOT run inside a database transaction
// (C-1 / W04-E02-S001 AC-02/03).
func (s *Service) effectDeliver(ctx context.Context, cd claimedOutboundDelivery, now time.Time) deliveryResult {
	secret, err := s.secrets.Resolve(ctx, cd.ep.SecretRef)
	if err != nil {
		return deliveryResult{secretErr: kerr.Wrapf(err, "webhook.effectDeliver", "resolve secret")}
	}

	ts := fmt.Sprintf("%d", now.Unix())
	// SEC-52: sign "<timestamp>.<body>" (Stripe/GitHub style) so X-Timestamp is
	// covered by the MAC and cannot be replayed with a forged timestamp.
	sig := signPayload(secret, ts, cd.body)
	headers := map[string]string{
		"Content-Type":    "application/json",
		"X-Signature":     "sha256=" + sig,
		"X-Timestamp":     ts,
		"X-Event-Id":      cd.ev.ID.String(),
		"Idempotency-Key": cd.ev.ID.String(),
	}

	url := ""
	if cd.ep.URL != nil {
		url = *cd.ep.URL
	}
	dctx, cancel := context.WithTimeout(ctx, OutboundTimeout)
	defer cancel()
	statusCode, postErr := s.sender.Post(dctx, url, cd.body, headers)
	return deliveryResult{
		ok:         postErr == nil && statusCode >= 200 && statusCode < 300,
		statusCode: statusCode,
		err:        postErr,
	}
}

// finalizeOutboundDelivery writes the outcome of the effect stage in a short,
// lease-fenced transaction: circuit-breaker bookkeeping, endpoint
// active/degraded transitions, and the webhook_events status update. It
// returns (true, nil) when the update was applied, (false, nil) when the
// lease was stale/expired and the result was discarded, and (false, err) on
// a database error.
func (s *Service) finalizeOutboundDelivery(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, cd claimedOutboundDelivery, res deliveryResult, now time.Time) (bool, error) {
	br := s.breaker.get(cd.ep.ID)
	next := cd.attempts + 1
	var applied bool
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		if res.secretErr != nil {
			// Matches the pre-staging behavior: a secret-resolution failure never
			// reached the POST or any webhook_events/breaker mutation, so the row
			// stays exactly as claimed (status/attempts untouched). Release the
			// lease so the row is immediately reclaimable rather than fenced for
			// leaseTTL.
			ct, uerr := db.Exec(ctx,
				`UPDATE webhook_events
				    SET lease_token = NULL, lease_generation = 0, lease_expires_at = NULL
				  WHERE id = $1
				    AND lease_token = $2 AND lease_generation = $3 AND lease_expires_at > $4`,
				cd.rowID, cd.lease.Token, cd.lease.Generation, now)
			if uerr != nil {
				return kerr.Wrapf(uerr, "webhook.finalizeOutboundDelivery", "release lease after secret error")
			}
			applied = ct.RowsAffected() > 0
			return nil
		}

		if res.ok {
			br.recordSuccess()
			s.emitBreakerState(cd.ep.ID, br)
			// ARCH-72: a recovered endpoint must return to 'active' — otherwise it
			// stays 'degraded' forever after the breaker closes.
			if _, cerr := db.Exec(ctx,
				`UPDATE webhook_endpoints
				    SET status = 'active', updated_at = $2, updated_by = $3
				  WHERE id = $1 AND status = 'degraded'`,
				cd.ep.ID, s.now(), uuid.Nil); cerr != nil {
				return kerr.Wrapf(cerr, "webhook.finalizeOutboundDelivery", "clear degraded status")
			}
			sigTrue := true
			ct, uerr := db.Exec(ctx,
				`UPDATE webhook_events
				    SET delivery_status = 'delivered', attempts = $2,
				        signature_ok = $3, last_error = NULL
				  WHERE id = $1
				    AND lease_token = $4 AND lease_generation = $5 AND lease_expires_at > $6`,
				cd.rowID, next, sigTrue, cd.lease.Token, cd.lease.Generation, now)
			if uerr != nil {
				return kerr.Wrapf(uerr, "webhook.finalizeOutboundDelivery", "mark delivered")
			}
			applied = ct.RowsAffected() > 0
			return nil
		}

		br.recordFailure(s.now())
		s.emitBreakerState(cd.ep.ID, br)

		// Persist endpoint status='degraded' when the breaker just opened.
		if br.isOpen(s.now()) {
			if _, derr := db.Exec(ctx,
				`UPDATE webhook_endpoints
				    SET status = 'degraded', updated_at = $2, updated_by = $3
				  WHERE id = $1`,
				cd.ep.ID, s.now(), uuid.Nil); derr != nil {
				return kerr.Wrapf(derr, "webhook.finalizeOutboundDelivery", "mark endpoint degraded")
			}
		}

		var errMsg string
		if res.err != nil {
			errMsg = truncate(res.err.Error())
		} else {
			errMsg = fmt.Sprintf("non-2xx status: %d", res.statusCode)
		}

		if next >= MaxAttempts {
			ct, uerr := db.Exec(ctx,
				`UPDATE webhook_events
				    SET delivery_status = 'dead', attempts = $2, last_error = $3
				  WHERE id = $1
				    AND lease_token = $4 AND lease_generation = $5 AND lease_expires_at > $6`,
				cd.rowID, next, errMsg, cd.lease.Token, cd.lease.Generation, now)
			if uerr != nil {
				return kerr.Wrapf(uerr, "webhook.finalizeOutboundDelivery", "dead-letter delivery")
			}
			applied = ct.RowsAffected() > 0
			return nil
		}

		nextAt := now.Add(webhookBackoff.Next(next))
		ct, uerr := db.Exec(ctx,
			`UPDATE webhook_events
			    SET delivery_status = 'failed', attempts = $2,
			        next_attempt_at = $3, last_error = $4
			  WHERE id = $1
			    AND lease_token = $5 AND lease_generation = $6 AND lease_expires_at > $7`,
			cd.rowID, next, nextAt, errMsg, cd.lease.Token, cd.lease.Generation, now)
		if uerr != nil {
			return kerr.Wrapf(uerr, "webhook.finalizeOutboundDelivery", "mark failed delivery")
		}
		applied = ct.RowsAffected() > 0
		return nil
	})
	return applied, err
}

// --- database helpers ---

type insertEventParams struct {
	endpointID      uuid.UUID
	direction       string
	externalEventID *string // nullable
	eventType       string
	payload         json.RawMessage
	signatureOk     *bool
	deliveryStatus  string
	receivedAt      time.Time
}

func (s *Service) insertEvent(ctx context.Context, db database.TenantDB, p insertEventParams) error {
	id := s.idgen.New()
	_, err := db.Exec(ctx,
		`INSERT INTO webhook_events
		    (id, tenant_id, endpoint_id, direction, external_event_id, event_type,
		     payload, signature_ok, received_at, delivery_status)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, p.endpointID, p.direction, p.externalEventID,
		p.eventType, p.payload, p.signatureOk, p.receivedAt, p.deliveryStatus)
	return err
}

// insertEventOnConflictIgnore inserts a webhook_events row and returns
// (true, nil) when inserted, (false, nil) on UNIQUE(endpoint_id,
// external_event_id) conflict, or (false, err) on any other error.
func (s *Service) insertEventOnConflictIgnore(ctx context.Context, db database.TenantDB, p insertEventParams) (bool, error) {
	id := s.idgen.New()
	var inserted bool
	err := db.QueryRow(ctx,
		// The partial unique index webhook_events_dedup covers only rows with a
		// non-NULL external_event_id, so ON CONFLICT must name the same predicate
		// for the arbiter to match. Every dedup-path caller supplies a non-NULL id.
		`INSERT INTO webhook_events
		    (id, tenant_id, endpoint_id, direction, external_event_id, event_type,
		     payload, signature_ok, received_at, delivery_status)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (endpoint_id, external_event_id) WHERE external_event_id IS NOT NULL DO NOTHING
		 RETURNING true`,
		id, p.endpointID, p.direction, p.externalEventID,
		p.eventType, p.payload, p.signatureOk, p.receivedAt, p.deliveryStatus).Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) loadEndpoint(ctx context.Context, db database.TenantDB, id uuid.UUID) (Endpoint, error) {
	var ep Endpoint
	err := db.QueryRow(ctx,
		`SELECT id, tenant_id, direction, provider_id, url, secret_ref,
		        signature_scheme, subscribed_events, status
		   FROM webhook_endpoints WHERE id = $1`,
		id).Scan(
		&ep.ID, &ep.TenantID, &ep.Direction, &ep.ProviderID, &ep.URL,
		&ep.SecretRef, &ep.SignatureScheme, &ep.SubscribedEvents, &ep.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Endpoint{}, kerr.E(kerr.KindNotFound, "not_found", "webhook endpoint not found")
	}
	if err != nil {
		return Endpoint{}, kerr.Wrapf(err, "webhook.loadEndpoint", "load endpoint")
	}
	return ep, nil
}

func (s *Service) loadEndpoints(ctx context.Context, db database.TenantDB, ids []uuid.UUID) (map[uuid.UUID]Endpoint, error) {
	endpoints := make(map[uuid.UUID]Endpoint, len(ids))
	if len(ids) == 0 {
		return endpoints, nil
	}
	rows, err := db.Query(ctx,
		`SELECT id, tenant_id, direction, provider_id, url, secret_ref,
		        signature_scheme, subscribed_events, status
		   FROM webhook_endpoints WHERE id = ANY($1::uuid[])`,
		ids)
	if err != nil {
		return nil, kerr.Wrapf(err, "webhook.loadEndpoints", "batch load endpoints")
	}
	defer rows.Close()
	for rows.Next() {
		var ep Endpoint
		if err := rows.Scan(
			&ep.ID, &ep.TenantID, &ep.Direction, &ep.ProviderID, &ep.URL,
			&ep.SecretRef, &ep.SignatureScheme, &ep.SubscribedEvents, &ep.Status,
		); err != nil {
			return nil, kerr.Wrapf(err, "webhook.loadEndpoints", "scan endpoint")
		}
		endpoints[ep.ID] = ep
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "webhook.loadEndpoints", "iterate endpoints")
	}
	return endpoints, nil
}

func (s *Service) loadOutboundEndpoints(ctx context.Context, db database.TenantDB, eventType string) ([]Endpoint, error) {
	rows, err := db.Query(ctx,
		`SELECT id, tenant_id, direction, provider_id, url, secret_ref,
		        signature_scheme, subscribed_events, status
		   FROM webhook_endpoints
		  WHERE direction = 'outbound'
		    AND status IN ('active','degraded')
		    AND $1 = ANY(subscribed_events)`,
		eventType)
	if err != nil {
		return nil, kerr.Wrapf(err, "webhook.loadOutboundEndpoints", "query endpoints")
	}
	defer rows.Close()
	var eps []Endpoint
	for rows.Next() {
		var ep Endpoint
		if err := rows.Scan(&ep.ID, &ep.TenantID, &ep.Direction, &ep.ProviderID,
			&ep.URL, &ep.SecretRef, &ep.SignatureScheme, &ep.SubscribedEvents, &ep.Status); err != nil {
			return nil, kerr.Wrapf(err, "webhook.loadOutboundEndpoints", "scan endpoint")
		}
		eps = append(eps, ep)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "webhook.loadOutboundEndpoints", "iterate endpoints")
	}
	return eps, nil
}

// --- signing ---

// signPayload computes the outbound X-Signature: HMAC-SHA256 over
// "<timestamp>.<body>" so the timestamp is authenticated alongside the body
// (SEC-52). Returned as lowercase hex (the caller prefixes "sha256=").
func signPayload(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// --- outbound body ---

func marshalOutboundBody(ev outbox.Event, tenantID uuid.UUID) ([]byte, error) {
	env := map[string]any{
		"id":        ev.ID,
		"type":      ev.Type,
		"tenant_id": tenantID,
		"payload":   ev.Payload,
	}
	return json.Marshal(env)
}

// --- retry backoff (exponential: 1 s, 5 s, 30 s, 2 m, 5 m) ---

// --- misc ---

// truncate caps a stored error message at maxErrLen bytes (every call site uses
// the same cap, so it is a constant rather than a parameter). It backs the cut up
// to a UTF-8 rune boundary so the stored string is never invalid UTF-8 — which a
// Postgres text column would reject.
func truncate(s string) string {
	const maxErrLen = 500
	if len(s) <= maxErrLen {
		return s
	}
	n := maxErrLen
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// dedupExtIDEnv returns the replay-dedup id for an inbound event: the
// verifier-derived event id when present, else a stable synthetic id derived
// from the authenticated canonical body ("sha256:<hex>"). Synthesizing an id
// for id-less events keeps EVERY inbound event replay-protected under the
// partial unique index (SEC-49/ARCH-74) — a NULL id would be treated as distinct
// and defeat dedup.
func dedupExtIDEnv(env Envelope) string {
	if env.EventID != "" {
		return env.EventID
	}
	sum := sha256.Sum256(env.CanonicalBody)
	return "sha256:" + hex.EncodeToString(sum[:])
}
