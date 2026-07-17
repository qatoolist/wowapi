package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/lease"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/retry"
	"github.com/qatoolist/wowapi/kernel/safety"
)

// maxAttempts is the dead-letter ceiling for notification deliveries. A delivery
// that fails this many times transitions to 'dead' and is no longer retried.
const maxAttempts = 3

// claimBatch is the number of queued deliveries claimed per SendPending call.
const claimBatch = 100

// leaseTTL is the duration a claimed notification delivery row is fenced from
// concurrent re-claim. It must cover the effect stage (sender.Send) plus the
// finalize round-trip.
const leaseTTL = 5 * time.Minute

// notifyBackoff is the per-attempt cooldown before a failed delivery becomes
// eligible for the next attempt (ARCH-75). It is backed by
// cenkalti/backoff/v5 via the shared kernel/retry package. The schedule is
// monotonic non-decreasing so a transient outage does not burn all attempts in
// seconds and dead-letter.
var notifyBackoff = retry.NewSchedule(retry.NewSequenceBackOff(
	30*time.Second, // after attempt 1
	2*time.Minute,  // after attempt 2
	10*time.Minute, // after attempt 3+ (only reached if maxAttempts raised)
))

// ChannelDest is a channel + destination pair in a Send request.
//
// Channel resolution simplification (SEC decision): party_contacts.kind
// ('email','phone','address','other') does not map cleanly to notification
// channels ('inapp','email','sms','whatsapp','push'). Rather than invent a
// domain contact schema, Message accepts explicit ChannelDest pairs. The
// in-app channel needs no external destination — it defaults to the recipient
// party ID string when Destination is empty.
type ChannelDest struct {
	Channel     Channel
	Destination string // empty = auto-set to partyID for inapp; required for others
}

// Message is the input to Service.Send.
type Message struct {
	TemplateKey      string
	RecipientPartyID uuid.UUID
	// Variables is the variable map rendered into the template body by
	// async senders. All keys must be in the registered TemplateSpec.Vars
	// allowlist; unknown keys are rejected with KindValidation.
	Variables  map[string]any
	Channels   []ChannelDest
	Importance Importance
	Resource   resource.Ref
	// Locale is the desired locale for template lookup; empty defaults to "en".
	// Template resolution follows the fallback chain: e.g. "hi-IN" → "hi" → "en".
	Locale string
}

// Notification is a persisted notification row returned by ListForParty.
type Notification struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	TemplateKey      string
	RecipientPartyID uuid.UUID
	Variables        map[string]any
	ResourceType     *string
	ResourceID       *uuid.UUID
	Importance       Importance
	Status           string
	CreatedAt        time.Time
	CreatedBy        uuid.UUID
}

// Delivery is a persisted notification_deliveries row, passed to ChannelSender
// on each send attempt. It carries routing information (channel, destination)
// only; real adapters must load and render the template body themselves via
// RenderBody if needed.
type Delivery struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	NotificationID uuid.UUID
	Channel        Channel
	Destination    string
	Status         string
	Attempts       int
	ProviderMsgID  string
	LastError      string
}

// Service is the notification framework. Module-facing operations (Send,
// ListForParty) run inside the caller's tenant transaction (app_rt). Platform
// operations (SendPending) run under a tenant-bound TxManager (app_platform) to
// advance append-only delivery status — the same split as kernel/document.
type Service struct {
	reg     *Registry
	idgen   model.IDGen
	now     func() time.Time
	senders map[string]ChannelSender
	tracer  observability.Tracer
	ob      outbox.Writer // optional; when set, ImportanceLegal deliveries write a delivery-audit event
}

// Option customizes the notify Service.
type Option func(*Service)

// WithTracer wires a tracer so each queued delivery captures the current
// request's W3C traceparent (roadmap O1/CA-9) into the delivery envelope; the
// async sender (SendPending) continues that trace when it delivers. Default:
// NoOpTracer (empty trace context — no behavior change). Mirrors the outbox
// writer/relay tracer seam.
func WithTracer(tr observability.Tracer) Option {
	return func(s *Service) {
		if tr != nil {
			s.tracer = tr
		}
	}
}

// WithOutbox wires an outbox.Writer so ImportanceLegal deliveries write a
// durable "notify.legal_delivery" audit event (with the provider's message id
// as receipt) in the SAME transaction as the 'sent' status update (DATA-08
// W0-T2, blueprint 07 §5: "importance=legal deliveries additionally write an
// audit row with provider receipt"). The clean baseline grants app_platform
// INSERT on events_outbox specifically for this use case. Default: nil (no
// legal-delivery audit event written) — matches the previous deferred
// behavior for callers that do not opt in.
func WithOutbox(ob outbox.Writer) Option {
	return func(s *Service) {
		if ob != nil {
			s.ob = ob
		}
	}
}

// New wires the service. reg and idgen are required.
func New(reg *Registry, idgen model.IDGen, opts ...Option) *Service {
	if reg == nil || idgen == nil {
		panic("notify.New: reg and idgen are required")
	}
	s := &Service{
		reg:     reg,
		idgen:   idgen,
		now:     time.Now,
		senders: map[string]ChannelSender{},
		tracer:  observability.NoOpTracer,
	}
	// Built-in in-app sender: "delivering" inapp is a no-op (row already written).
	s.senders[string(ChannelInApp)] = inAppSender{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// RegisterSender registers a ChannelSender for the given channel. Call this
// at wiring time for each transport (email, sms, whatsapp, push). An in-app
// sender is registered by default.
//
// Adapters must implement safety.Declarer and declare their duplicate-safety
// mechanism; registration panics otherwise.
func (s *Service) RegisterSender(channel Channel, sender ChannelSender) {
	if _, ok := sender.(safety.Declarer); !ok {
		panic(fmt.Sprintf("notify.RegisterSender: sender %T for channel %q must implement safety.Declarer", sender, channel))
	}
	s.senders[string(channel)] = sender
}

// --- module-facing operations (run on the caller's app_rt tenant tx) ---

// Send writes one notifications row and one notification_deliveries row per
// resolved channel, all within the caller's tenant transaction. Returns the
// notification id. Errors if the template key is not registered, any variable
// key is outside the spec's allowlist, or no template exists in the DB for any
// requested channel.
func (s *Service) Send(ctx context.Context, db database.TenantDB, msg Message) (uuid.UUID, error) {
	// 1. Registry gate: template key must be registered.
	spec, ok := s.reg.Get(msg.TemplateKey)
	if !ok {
		return uuid.Nil, kerr.E(kerr.KindValidation, "unknown_template_key",
			"notify: unknown template key: "+msg.TemplateKey)
	}

	// 2. Variable allowlist: no caller-supplied var may be outside the spec.
	for k := range msg.Variables {
		if !spec.allowsVar(k) {
			return uuid.Nil, kerr.E(kerr.KindValidation, "template_var_not_in_allowlist",
				fmt.Sprintf("notify: variable %q is not allowlisted for key %s", k, msg.TemplateKey))
		}
	}

	// 3. Defaults.
	locale := msg.Locale
	if locale == "" {
		locale = "en"
	}
	importance := msg.Importance
	if importance == "" {
		importance = ImportanceNormal
	}
	if msg.RecipientPartyID == uuid.Nil {
		return uuid.Nil, kerr.E(kerr.KindValidation, "missing_recipient",
			"notify: RecipientPartyID is required")
	}
	if len(msg.Channels) == 0 {
		return uuid.Nil, kerr.E(kerr.KindValidation, "no_channels",
			"notify: at least one channel is required")
	}

	// Supplied variables, used for the ARCH-77 dry-run render below.
	renderVars := msg.Variables
	if renderVars == nil {
		renderVars = map[string]any{}
	}

	// 4. Resolve channels: keep only those with a template in the DB, and
	// dry-run the render of each resolved template body against the supplied
	// variables so a template that REFERENCES a variable the caller did not
	// supply fails SYNCHRONOUSLY (KindValidation) before any row is written
	// (ARCH-77) — rather than committing rows that fail at delivery time, or
	// are silently wrong for the never-rendered in-app channel.
	type resolved struct {
		channel     Channel
		destination string
	}
	var channels []resolved
	optedOut := false
	for _, cd := range msg.Channels {
		if cd.Channel == "" {
			continue
		}
		// Skip a channel the recipient has opted out of (R5 channel preferences).
		off, err := s.channelDisabled(ctx, db, msg.RecipientPartyID, cd.Channel)
		if err != nil {
			return uuid.Nil, err
		}
		if off {
			optedOut = true
			continue
		}
		dest := cd.Destination
		if cd.Channel == ChannelInApp && dest == "" {
			dest = msg.RecipientPartyID.String()
		}
		body, found, err := s.lookupTemplate(ctx, db, msg.TemplateKey, string(cd.Channel), locale)
		if err != nil {
			return uuid.Nil, err
		}
		if !found {
			continue
		}
		// Dry-run render: missingkey=error makes an unsupplied referenced var
		// fail here. Escaping context matches the channel (SEC-51).
		if _, rerr := renderBody(spec, cd.Channel, body, renderVars); rerr != nil {
			return uuid.Nil, kerr.Wrapf(rerr, "notify.Send", "template for channel %s cannot render with the supplied variables", cd.Channel)
		}
		channels = append(channels, resolved{channel: cd.Channel, destination: dest})
	}
	if len(channels) == 0 {
		if optedOut {
			return uuid.Nil, kerr.E(kerr.KindValidation, "all_channels_opted_out",
				"notify: the recipient has opted out of every requested channel")
		}
		return uuid.Nil, kerr.E(kerr.KindValidation, "no_template_found",
			"notify: no template found for key "+msg.TemplateKey+" on any requested channel")
	}

	// 5. Serialize variables for storage.
	varBytes, err := json.Marshal(renderVars)
	if err != nil {
		return uuid.Nil, kerr.E(kerr.KindInternal, "marshal_vars",
			"notify: failed to marshal variables: "+err.Error())
	}

	// 6. Optional resource anchor.
	var resType any
	var resID any
	if !msg.Resource.IsZero() {
		resType = msg.Resource.Type
		resID = msg.Resource.ID
	}

	actor := actorFromCtx(ctx)
	notifID := s.idgen.New()

	// 7. INSERT notifications row.
	_, err = db.Exec(ctx,
		`INSERT INTO notifications
		    (id, tenant_id, template_key, recipient_party_id, variables,
		     resource_type, resource_id, importance, status, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, 'pending', $8)`,
		notifID, msg.TemplateKey, msg.RecipientPartyID, varBytes,
		resType, resID, string(importance), actor)
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "notify.Send", "insert notification")
	}

	// Capture the current distributed-trace context (W3C traceparent) so the async
	// sender can continue the SAME trace when it delivers (roadmap O1/CA-9). Empty
	// (NoOp / no active span) → NULL trace_context.
	var traceCtx any
	if tc := s.tracer.Inject(ctx); tc != "" {
		traceCtx = tc
	}

	// 8. INSERT one notification_deliveries row per resolved channel.
	for _, rc := range channels {
		delID := s.idgen.New()
		_, err = db.Exec(ctx,
			`INSERT INTO notification_deliveries
			    (id, tenant_id, notification_id, channel, destination, status, trace_context)
			 VALUES ($1, app_tenant_id(), $2, $3, $4, 'queued', $5)`,
			delID, notifID, string(rc.channel), rc.destination, traceCtx)
		if err != nil {
			return uuid.Nil, kerr.Wrapf(err, "notify.Send", "insert delivery for channel %s", rc.channel)
		}
	}

	return notifID, nil
}

// ListForParty returns the notifications for a party (the in-app inbox),
// newest first. Runs on the caller's read-only or read-write tenant tx.
func (s *Service) ListForParty(ctx context.Context, db database.TenantDB, partyID uuid.UUID) ([]Notification, error) {
	rows, err := db.Query(ctx,
		`SELECT id, tenant_id, template_key, recipient_party_id, variables,
		        resource_type, resource_id, importance, status, created_at, created_by
		   FROM notifications
		  WHERE recipient_party_id = $1
		  ORDER BY created_at DESC`,
		partyID)
	if err != nil {
		return nil, kerr.Wrapf(err, "notify.ListForParty", "query notifications")
	}
	defer rows.Close()

	var out []Notification
	for rows.Next() {
		var n Notification
		var varBytes []byte
		if err := rows.Scan(
			&n.ID, &n.TenantID, &n.TemplateKey, &n.RecipientPartyID,
			&varBytes, &n.ResourceType, &n.ResourceID,
			&n.Importance, &n.Status, &n.CreatedAt, &n.CreatedBy,
		); err != nil {
			return nil, kerr.Wrapf(err, "notify.ListForParty", "scan notification")
		}
		if len(varBytes) > 0 {
			if err := json.Unmarshal(varBytes, &n.Variables); err != nil {
				return nil, kerr.Wrapf(err, "notify.ListForParty", "unmarshal variables")
			}
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "notify.ListForParty", "iterate notifications")
	}
	return out, nil
}

// DeliveryReceipt is the per-channel delivery record for a notification: its
// status, attempt count, the provider's message id (receipt), and the last
// error. It answers "did this notification actually go out, on which channels,
// and what did the provider say" (roadmap R5).
type DeliveryReceipt struct {
	ID            uuid.UUID
	Channel       Channel
	Destination   string
	Status        string // queued | sent | delivered | failed | dead
	Attempts      int
	ProviderMsgID string
	LastError     string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// Deliveries returns the delivery receipts for a notification, one per channel
// fan-out, newest state first-written order. Runs in the caller's tenant tx
// (RLS-scoped), so a caller only ever sees its own tenant's receipts.
func (s *Service) Deliveries(ctx context.Context, db database.TenantDB, notificationID uuid.UUID) ([]DeliveryReceipt, error) {
	rows, err := db.Query(ctx,
		`SELECT id, channel, destination, status, attempts,
		        COALESCE(provider_message_id,''), COALESCE(last_error,''), created_at, updated_at
		   FROM notification_deliveries
		  WHERE notification_id = $1
		  ORDER BY created_at`, notificationID)
	if err != nil {
		return nil, kerr.Wrapf(err, "notify.Deliveries", "query deliveries")
	}
	defer rows.Close()

	var out []DeliveryReceipt
	for rows.Next() {
		var r DeliveryReceipt
		var channel string
		if err := rows.Scan(&r.ID, &channel, &r.Destination, &r.Status, &r.Attempts,
			&r.ProviderMsgID, &r.LastError, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, kerr.Wrapf(err, "notify.Deliveries", "scan delivery")
		}
		r.Channel = Channel(channel)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "notify.Deliveries", "iterate deliveries")
	}
	return out, nil
}

// SetChannelPref records a recipient's opt-in/opt-out for a channel (R5). Absence
// of a preference means enabled, so this is only needed to opt OUT (or to re-enable
// after opting out). Runs in the caller's tenant tx.
func (s *Service) SetChannelPref(ctx context.Context, db database.TenantDB, partyID uuid.UUID, channel Channel, enabled bool) error {
	if _, err := db.Exec(ctx,
		`INSERT INTO notification_channel_prefs (tenant_id, party_id, channel, enabled)
		 VALUES (app_tenant_id(), $1, $2, $3)
		 ON CONFLICT (tenant_id, party_id, channel)
		 DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = now()`,
		partyID, string(channel), enabled); err != nil {
		return kerr.Wrapf(err, "notify.SetChannelPref", "upsert preference")
	}
	return nil
}

// channelDisabled reports whether a party has an explicit opt-out for a channel.
func (s *Service) channelDisabled(ctx context.Context, db database.TenantDB, partyID uuid.UUID, channel Channel) (bool, error) {
	var disabled bool
	if err := db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM notification_channel_prefs
		                 WHERE party_id = $1 AND channel = $2 AND enabled = false)`,
		partyID, string(channel)).Scan(&disabled); err != nil {
		return false, kerr.Wrapf(err, "notify.channelDisabled", "read preference")
	}
	return disabled, nil
}

// --- platform-privileged operations (run on a tenant-bound app_platform tx) ---

// SendPending is the async worker step. It runs as app_platform (tenant-bound)
// using a three-stage claim/effect/finalize protocol built on the shared
// kernel/lease primitive:
//
//  1. Claim-tx: selects eligible notification_deliveries rows with
//     FOR UPDATE SKIP LOCKED, assigns a fresh lease, and commits.
//  2. Effect stage: calls the registered ChannelSender for each claimed
//     delivery entirely outside any database transaction. The delivery ID is
//     the idempotency key passed to the adapter (via Delivery.ID).
//  3. Finalize-tx: per delivery, updates status only if the lease token and
//     generation still match and the lease has not expired. A mismatch means
//     the row was reclaimed by another worker; the effect result is discarded.
//
// Outcomes:
//   - success → 'sent' (provider_message_id set)
//   - failure → 'failed' (attempts incremented, next_attempt_at = now +
//     backoff(newAttempts)); at maxAttempts → 'dead'
//
// ARCH-75: a 'failed' delivery is re-claimed only once both its backoff and
// its previous lease have elapsed (relative to the passed `now`), so a
// transient outage does not burn all maxAttempts in seconds and permanently
// dead-letter.
//
// NOTE: ImportanceLegal deliveries additionally write a durable
// "notify.legal_delivery" outbox event carrying the provider's message id as
// receipt, in the SAME finalize transaction as the 'sent' status update
// (DATA-08 W0-T2). The clean baseline grants app_platform INSERT on events_outbox
// for exactly this use case. Requires the Service to be wired with WithOutbox;
// if no outbox writer is configured, the event is skipped (no error).
//
// Returns the number of deliveries successfully finalized as sent.
func (s *Service) SendPending(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, now time.Time) (int, error) {
	claimed, err := s.claimPending(ctx, plat, tenantID, now)
	if err != nil {
		return 0, err
	}

	slog.InfoContext(ctx, "notify.claimed", "tenant", tenantID, "count", len(claimed))

	var sent int
	for _, d := range claimed {
		providerMsgID, sendErr, permanent := s.effectSend(ctx, d)

		applied, err := s.finalizeDelivery(ctx, plat, tenantID, d, providerMsgID, sendErr, permanent, now)
		if err != nil {
			// A finalize transaction error is fatal to the batch only in the
			// sense that we cannot trust the outcome of this delivery. Keep
			// processing the remaining claimed rows so one bad finalize does
			// not stall the whole batch.
			slog.ErrorContext(ctx, "notify.finalize_error", "delivery_id", d.id, "err", err)
			continue
		}
		if applied && sendErr == nil {
			sent++
		} else if !applied {
			slog.InfoContext(ctx, "notify.finalize_discard", "delivery_id", d.id, "reason", "lease_mismatch_or_expired")
		}
	}

	return sent, nil
}

// claimedDelivery is a notification_deliveries row that has been assigned a
// lease in the claim stage and is ready for the effect/finalize stages.
type claimedDelivery struct {
	id         uuid.UUID
	notifID    uuid.UUID
	channel    string
	dest       string
	attempts   int
	importance string
	trace      *string
	lease      lease.Lease
}

// claimPending claims up to claimBatch eligible notification deliveries and
// assigns each a fresh lease from the shared kernel/lease primitive.
func (s *Service) claimPending(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, now time.Time) ([]claimedDelivery, error) {
	var out []claimedDelivery
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx,
			`SELECT d.id, d.notification_id, d.channel, d.destination, d.attempts,
			        n.importance, d.trace_context,
			        d.lease_token, d.lease_generation, d.lease_expires_at
			   FROM notification_deliveries d
			   JOIN notifications n ON n.id = d.notification_id
			  WHERE d.status IN ('queued', 'failed')
			    AND (d.next_attempt_at IS NULL OR d.next_attempt_at <= $1)
			    AND (d.lease_expires_at IS NULL OR d.lease_expires_at <= $1)
			  ORDER BY d.created_at
			  FOR UPDATE OF d SKIP LOCKED
			  LIMIT $2`,
			now, claimBatch)
		if err != nil {
			return kerr.Wrapf(err, "notify.claimPending", "claim deliveries")
		}

		var scanned []claimedDelivery
		for rows.Next() {
			var d claimedDelivery
			var leaseToken *string
			var leaseGeneration *int64
			var leaseExpiresAt *time.Time
			if err := rows.Scan(&d.id, &d.notifID, &d.channel, &d.dest,
				&d.attempts, &d.importance, &d.trace,
				&leaseToken, &leaseGeneration, &leaseExpiresAt); err != nil {
				rows.Close()
				return kerr.Wrapf(err, "notify.claimPending", "scan delivery")
			}

			var existing lease.Lease
			if leaseToken != nil {
				existing.Token = *leaseToken
				existing.Generation = *leaseGeneration
				existing.ExpiresAt = *leaseExpiresAt
			}
			if existing.Zero() {
				d.lease = lease.New(leaseTTL)
			} else {
				d.lease = existing.NextEpoch(leaseTTL)
			}
			// Use the orchestrated clock so tests with fake time fence leases correctly.
			d.lease.ExpiresAt = now.Add(leaseTTL)

			scanned = append(scanned, d)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return kerr.Wrapf(err, "notify.claimPending", "iterate deliveries")
		}
		rows.Close()

		// Assign leases after closing the cursor so we don't try to run UPDATE
		// while the SELECT cursor is still active on the same connection.
		for _, d := range scanned {
			if _, err := db.Exec(ctx,
				`UPDATE notification_deliveries
				    SET lease_token = $2, lease_generation = $3, lease_expires_at = $4
				  WHERE id = $1`,
				d.id, d.lease.Token, d.lease.Generation, d.lease.ExpiresAt); err != nil {
				return kerr.Wrapf(err, "notify.claimPending", "assign lease %s", d.id)
			}
			out = append(out, d)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// effectSend performs the remote ChannelSender.Send call for a claimed
// delivery. It MUST NOT run inside a database transaction.
func (s *Service) effectSend(ctx context.Context, d claimedDelivery) (providerMsgID string, sendErr error, permanent bool) {
	del := Delivery{
		ID:             d.id,
		TenantID:       uuid.Nil, // not used by adapters
		NotificationID: d.notifID,
		Channel:        Channel(d.channel),
		Destination:    d.dest,
		Status:         "queued",
		Attempts:       d.attempts,
	}

	sendCtx := ctx
	if d.trace != nil && *d.trace != "" {
		sendCtx = s.tracer.Extract(ctx, *d.trace)
	}
	sendCtx, span := s.tracer.StartSpan(sendCtx, "notify.send "+d.channel)
	span.SetAttr("notify.channel", d.channel)
	span.SetAttr("notify.delivery_id", d.id.String())
	defer span.End()

	sender, ok := s.senderFor(del.Channel)
	if !ok {
		return "", kerr.E(kerr.KindExternal, "no_channel_sender",
			"no sender registered for channel "+string(del.Channel)), true
	}

	providerMsgID, sendErr = sender.Send(sendCtx, del)
	if sendErr != nil {
		span.RecordError(sendErr)
	}
	slog.InfoContext(sendCtx, "notify.effect", "delivery_id", d.id, "channel", d.channel, "ok", sendErr == nil)
	return providerMsgID, sendErr, false
}

// finalizeDelivery writes the outcome of the effect stage in a short,
// lease-fenced transaction. It returns (true, nil) when the update was applied,
// (false, nil) when the lease was stale/expired and the result was discarded,
// and (false, err) on a database error.
func (s *Service) finalizeDelivery(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, d claimedDelivery, providerMsgID string, sendErr error, permanent bool, now time.Time) (bool, error) {
	var applied bool
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		newAttempts := d.attempts + 1
		if sendErr == nil {
			ct, err := db.Exec(ctx,
				`UPDATE notification_deliveries
				    SET status = 'sent', provider_message_id = $2,
				        attempts = $3, updated_at = $4
				  WHERE id = $1
				    AND lease_token = $5
				    AND lease_generation = $6
				    AND lease_expires_at > $7`,
				d.id, providerMsgID, newAttempts, now,
				d.lease.Token, d.lease.Generation, now)
			if err != nil {
				return kerr.Wrapf(err, "notify.finalizeDelivery", "mark sent %s", d.id)
			}
			if ct.RowsAffected() == 0 {
				return nil // lease stale/expired — discard silently
			}
			applied = true

			if s.ob != nil && d.importance == string(ImportanceLegal) {
				if err := s.ob.Write(ctx, db, outbox.Event{
					Type:     "notify.legal_delivery",
					Resource: resource.Ref{Type: "notify.delivery", ID: d.id},
					Payload: map[string]any{
						"delivery_id":     d.id.String(),
						"notification_id": d.notifID.String(),
						"channel":         d.channel,
						"provider_msg_id": providerMsgID,
						"sent_at":         now,
					},
				}); err != nil {
					return kerr.Wrapf(err, "notify.finalizeDelivery", "write legal delivery audit event %s", d.id)
				}
			}
			return nil
		}

		newStatus := "failed"
		var nextAttempt any
		if permanent || newAttempts >= maxAttempts {
			newStatus = "dead"
		} else {
			nextAttempt = now.Add(notifyBackoff.Next(newAttempts))
		}

		ct, err := db.Exec(ctx,
			`UPDATE notification_deliveries
			    SET status = $2, attempts = $3, last_error = $4,
			        next_attempt_at = $5, updated_at = $6
			  WHERE id = $1
			    AND lease_token = $7
			    AND lease_generation = $8
			    AND lease_expires_at > $9`,
			d.id, newStatus, newAttempts, sendErr.Error(), nextAttempt, now,
			d.lease.Token, d.lease.Generation, now)
		if err != nil {
			return kerr.Wrapf(err, "notify.finalizeDelivery", "mark %s %s", newStatus, d.id)
		}
		applied = ct.RowsAffected() > 0
		return nil
	})
	return applied, err
}

// --- helpers ---

// lookupTemplate finds a template body in notification_templates for (key,
// channel, locale) with locale-fallback chain. tenant rows win over platform
// defaults via ORDER BY tenant_id NULLS LAST. The hybrid RLS policy
// (app_tenant_id_or_null) makes both tenant-specific and platform rows visible
// within the current tenant tx.
func (s *Service) lookupTemplate(ctx context.Context, db database.TenantDB, key, channel, locale string) (body string, found bool, err error) {
	for _, loc := range localeFallback(locale) {
		var b string
		qerr := db.QueryRow(ctx,
			`SELECT body FROM notification_templates
			  WHERE key = $1 AND channel = $2 AND locale = $3 AND status = 'active'
			  ORDER BY tenant_id NULLS LAST
			  LIMIT 1`,
			key, channel, loc).Scan(&b)
		if errors.Is(qerr, pgx.ErrNoRows) {
			continue
		}
		if qerr != nil {
			return "", false, kerr.Wrapf(qerr, "notify.lookupTemplate", "lookup %s/%s/%s", key, channel, loc)
		}
		return b, true, nil
	}
	return "", false, nil
}

// localeFallback returns the locale lookup chain: e.g. "hi-IN" → ["hi-IN",
// "hi", "en"]; "en" → ["en"]; "" → ["en"].
func localeFallback(locale string) []string {
	if locale == "" || locale == "en" {
		return []string{"en"}
	}
	chain := []string{locale}
	if idx := strings.Index(locale, "-"); idx > 0 {
		base := locale[:idx]
		if base != locale && base != "en" {
			chain = append(chain, base)
		}
	}
	// Always end with "en" as the final fallback.
	chain = append(chain, "en")
	return chain
}

// senderFor returns the registered sender for ch. The bool is false when no
// sender is wired for the channel — the caller must treat that as a delivery
// failure, NOT a silent success (roadmap CA-15).
func (s *Service) senderFor(ch Channel) (ChannelSender, bool) {
	sndr, ok := s.senders[string(ch)]
	return sndr, ok
}

func actorFromCtx(ctx context.Context) uuid.UUID {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return uuid.Nil
}
