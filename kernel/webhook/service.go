package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/outbox"
)

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

	if verr := v.Verify(secret, in.RawBody, in.Headers); verr != nil {
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

	// Timestamp window ±5 m.
	delta := s.now().Sub(in.Timestamp)
	if delta < 0 {
		delta = -delta
	}
	if delta > TimestampWindow {
		return kerr.E(kerr.KindValidation, "timestamp_out_of_window", "webhook timestamp is outside the ±5 m replay window")
	}

	payload := json.RawMessage(`{}`)
	if len(in.RawBody) > 0 && json.Valid(in.RawBody) {
		payload = json.RawMessage(in.RawBody)
	}

	// Replay-dedup id: the provider's external id when present, else a stable
	// synthetic id derived from the body (SEC-49/ARCH-74), so an id-less event is
	// still replay-protected under the partial unique index.
	sigTrue := true
	dedup := dedupExtID(in)
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
				msg := truncate(herr.Error(), 500)
				if next >= MaxAttempts {
					if _, uerr := db.Exec(ctx,
						`UPDATE webhook_events
						    SET delivery_status = 'dead', attempts = $2, last_error = $3
						  WHERE id = $1`,
						r.id, next, msg); uerr != nil {
						return kerr.Wrapf(uerr, "webhook.ProcessInbound", "dead-letter event")
					}
				} else {
					nextAt := now.Add(backoff(next))
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
// tenantID whose subscribed_events contain ev.Type. For each matching endpoint
// it upserts a delivery row, checks the circuit breaker, signs the body, POSTs
// via Sender, then records the outcome. Open-circuit endpoints are skipped
// silently (their delivery rows stay pending). Runs as app_platform.
func (s *Service) DispatchOutbound(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, ev outbox.Event, now time.Time) error {
	return plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		eps, err := s.loadOutboundEndpoints(ctx, db, ev.Type)
		if err != nil {
			return err
		}
		body, err := marshalOutboundBody(ev, tenantID)
		if err != nil {
			return kerr.Wrapf(err, "webhook.DispatchOutbound", "marshal body")
		}
		for _, ep := range eps {
			// Ignore per-endpoint delivery errors: one failure must not block others.
			_ = s.deliverToEndpoint(ctx, db, ep, ev, body, now)
		}
		return nil
	})
}

// RetryOutbound re-delivers previously-failed outbound webhook events for
// tenantID whose backoff has elapsed. Runs as app_platform (plat is a TxManager
// over the platform pool), mirroring ProcessInbound. Without this, DispatchOutbound
// would leave a 'failed' row untouched and the outbox relay — having marked its
// source event dispatched — would never re-drive delivery (ARCH-70). It claims
// failed rows FOR UPDATE SKIP LOCKED, re-runs deliverToEndpoint for each, and
// advances status to delivered / failed+backoff / dead at the ceiling.
func (s *Service) RetryOutbound(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, now time.Time) error {
	return plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx,
			`SELECT id, endpoint_id, external_event_id, event_type, payload
			   FROM webhook_events
			  WHERE direction = 'outbound'
			    AND delivery_status = 'failed'
			    AND (next_attempt_at IS NULL OR next_attempt_at <= $1)
			  ORDER BY received_at
			  LIMIT 10
			    FOR UPDATE SKIP LOCKED`,
			now)
		if err != nil {
			return kerr.Wrapf(err, "webhook.RetryOutbound", "claim events")
		}
		type claimedRow struct {
			id         uuid.UUID
			endpointID uuid.UUID
			extID      string
			eventType  string
			payload    json.RawMessage
		}
		var claimed []claimedRow
		for rows.Next() {
			var r claimedRow
			if err := rows.Scan(&r.id, &r.endpointID, &r.extID, &r.eventType, &r.payload); err != nil {
				rows.Close()
				return kerr.Wrapf(err, "webhook.RetryOutbound", "scan event")
			}
			claimed = append(claimed, r)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "webhook.RetryOutbound", "iterate events")
		}

		for _, r := range claimed {
			ep, err := s.loadEndpoint(ctx, db, r.endpointID)
			if err != nil {
				return err
			}
			evID, perr := uuid.Parse(r.extID)
			if perr != nil {
				// A synthetic/non-UUID id should never appear on an outbound row
				// (they key on the outbox event UUID); skip rather than crash.
				continue
			}
			// Reconstruct the minimal outbox.Event deliverToEndpoint needs (ID +
			// Type); the signing/POST body is the persisted payload column.
			ev := outbox.Event{ID: evID, Type: r.eventType}
			// One failure must not block the rest of the batch.
			_ = s.deliverToEndpoint(ctx, db, ep, ev, r.payload, now)
		}
		return nil
	})
}

// --- internal ---

func (s *Service) runInboundHandler(ctx context.Context, db database.TenantDB, ev Event) error {
	h, ok := s.handlers[ev.EventType]
	if !ok {
		return nil // no handler registered — treat as processed (no-op)
	}
	return h(ctx, db, ev)
}

func (s *Service) deliverToEndpoint(
	ctx context.Context,
	db database.TenantDB,
	ep Endpoint,
	ev outbox.Event,
	body []byte,
	now time.Time,
) error {
	extID := ev.ID.String() // outbox event id is the outbound idempotency key
	extIDPtr := &extID

	// Upsert a pending delivery row (idempotent on the outbox event id).
	_, err := s.insertEventOnConflictIgnore(ctx, db, insertEventParams{
		endpointID:      ep.ID,
		direction:       DirectionOutbound,
		externalEventID: extIDPtr,
		eventType:       ev.Type,
		payload:         json.RawMessage(body),
		signatureOk:     nil,
		deliveryStatus:  StatusPending,
		receivedAt:      now,
	})
	if err != nil {
		return kerr.Wrapf(err, "webhook.deliverToEndpoint", "upsert delivery row")
	}

	// Load the current row state.
	var (
		rowID         uuid.UUID
		attempts      int
		status        string
		nextAttemptAt *time.Time
	)
	if err := db.QueryRow(ctx,
		`SELECT id, attempts, delivery_status, next_attempt_at
		   FROM webhook_events
		  WHERE endpoint_id = $1 AND external_event_id = $2`,
		ep.ID, extID).Scan(&rowID, &attempts, &status, &nextAttemptAt); err != nil {
		return kerr.Wrapf(err, "webhook.deliverToEndpoint", "load delivery row")
	}

	// Skip terminal or not-yet-due rows.
	if status == StatusDelivered || status == StatusDead {
		return nil
	}
	if status == StatusFailed && nextAttemptAt != nil && nextAttemptAt.After(now) {
		return nil
	}

	// Circuit breaker.
	br := s.breaker.get(ep.ID)
	if !br.allow(s.now()) {
		return nil // open — leave row pending
	}

	secret, err := s.secrets.Resolve(ctx, ep.SecretRef)
	if err != nil {
		return kerr.Wrapf(err, "webhook.deliverToEndpoint", "resolve secret")
	}

	ts := fmt.Sprintf("%d", now.Unix())
	// SEC-52: sign "<timestamp>.<body>" (Stripe/GitHub style) so X-Timestamp is
	// covered by the MAC and cannot be replayed with a forged timestamp.
	sig := signPayload(secret, ts, body)
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Signature":  "sha256=" + sig,
		"X-Timestamp":  ts,
		"X-Event-Id":   ev.ID.String(),
	}

	url := ""
	if ep.URL != nil {
		url = *ep.URL
	}
	dctx, cancel := context.WithTimeout(ctx, OutboundTimeout)
	defer cancel()
	statusCode, postErr := s.sender.Post(dctx, url, body, headers)

	next := attempts + 1
	sigTrue := true

	if postErr == nil && statusCode >= 200 && statusCode < 300 {
		br.recordSuccess()
		// ARCH-72: a recovered endpoint must return to 'active' — otherwise it
		// stays 'degraded' forever after the breaker closes.
		if _, cerr := db.Exec(ctx,
			`UPDATE webhook_endpoints
			    SET status = 'active', updated_at = $2, updated_by = $3
			  WHERE id = $1 AND status = 'degraded'`,
			ep.ID, s.now(), uuid.Nil); cerr != nil {
			return kerr.Wrapf(cerr, "webhook.deliverToEndpoint", "clear degraded status")
		}
		_, uerr := db.Exec(ctx,
			`UPDATE webhook_events
			    SET delivery_status = 'delivered', attempts = $2,
			        signature_ok = $3, last_error = NULL
			  WHERE id = $1`,
			rowID, next, sigTrue)
		return kerr.Wrapf(uerr, "webhook.deliverToEndpoint", "mark delivered")
	}

	br.recordFailure(s.now())

	// Persist endpoint status='degraded' when the breaker just opened.
	if br.isOpen(s.now()) {
		_, _ = db.Exec(ctx,
			`UPDATE webhook_endpoints
			    SET status = 'degraded', updated_at = $2, updated_by = $3
			  WHERE id = $1`,
			ep.ID, s.now(), uuid.Nil)
	}

	var errMsg string
	if postErr != nil {
		errMsg = truncate(postErr.Error(), 500)
	} else {
		errMsg = fmt.Sprintf("non-2xx status: %d", statusCode)
	}

	if next >= MaxAttempts {
		_, uerr := db.Exec(ctx,
			`UPDATE webhook_events
			    SET delivery_status = 'dead', attempts = $2, last_error = $3
			  WHERE id = $1`,
			rowID, next, errMsg)
		return kerr.Wrapf(uerr, "webhook.deliverToEndpoint", "dead-letter delivery")
	}

	nextAt := now.Add(backoff(next))
	_, uerr := db.Exec(ctx,
		`UPDATE webhook_events
		    SET delivery_status = 'failed', attempts = $2,
		        next_attempt_at = $3, last_error = $4
		  WHERE id = $1`,
		rowID, next, nextAt, errMsg)
	return kerr.Wrapf(uerr, "webhook.deliverToEndpoint", "mark failed delivery")
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
		&ep.SecretRef, &ep.SignatureScheme, &ep.SubscribedEvents, &ep.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Endpoint{}, kerr.E(kerr.KindNotFound, "not_found", "webhook endpoint not found")
	}
	if err != nil {
		return Endpoint{}, kerr.Wrapf(err, "webhook.loadEndpoint", "load endpoint")
	}
	return ep, nil
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

func backoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return time.Second
	case 2:
		return 5 * time.Second
	case 3:
		return 30 * time.Second
	case 4:
		return 2 * time.Minute
	default:
		return 5 * time.Minute
	}
}

// --- misc ---

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// dedupExtID returns the replay-dedup id for an inbound event: the provider's
// external id when present, else a stable synthetic id derived from the raw body
// ("sha256:<hex>"). Synthesizing an id for id-less events keeps EVERY inbound
// event replay-protected under the partial unique index (SEC-49/ARCH-74) — a
// NULL id would be treated as distinct and defeat dedup.
func dedupExtID(in InboundIn) string {
	if in.ExternalEventID != "" {
		return in.ExternalEventID
	}
	sum := sha256.Sum256(in.RawBody)
	return "sha256:" + hex.EncodeToString(sum[:])
}
