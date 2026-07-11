package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// maxAttempts is the dead-letter ceiling for notification deliveries. A delivery
// that fails this many times transitions to 'dead' and is no longer retried.
const maxAttempts = 3

// claimBatch is the number of queued deliveries claimed per SendPending call.
const claimBatch = 100

// backoffSchedule is the per-attempt cooldown before a failed delivery becomes
// eligible for the next attempt (ARCH-75). Index by newAttempts-1. The last
// entry applies to any further attempts; monotonic non-decreasing so a
// transient outage does not burn all attempts in seconds and dead-letter.
var backoffSchedule = []time.Duration{
	30 * time.Second, // after attempt 1
	2 * time.Minute,  // after attempt 2
	10 * time.Minute, // after attempt 3+ (only reached if maxAttempts raised)
}

// backoff returns the cooldown after the given attempt number (1-based).
func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > len(backoffSchedule) {
		attempt = len(backoffSchedule)
	}
	return backoffSchedule[attempt-1]
}

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
// audit row with provider receipt"). Migration 00011 grants app_platform
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
func (s *Service) RegisterSender(channel Channel, sender ChannelSender) {
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

// SendPending is the async worker step. It runs as app_platform (tenant-bound):
// claims queued notification_deliveries with FOR UPDATE SKIP LOCKED, calls the
// registered ChannelSender for each, and advances status:
//
//   - success → 'sent' (provider_message_id set)
//   - failure → 'failed' (attempts incremented, next_attempt_at = now +
//     backoff(newAttempts)); at maxAttempts → 'dead'
//
// ARCH-75: a 'failed' delivery is re-claimed only once its next_attempt_at has
// elapsed (relative to the passed `now`), so a transient outage does not burn
// all maxAttempts in seconds and permanently dead-letter. The backoff schedule
// is monotonic (see backoff).
//
// NOTE: claim + sender call + status update happen in one transaction. Real
// production deployments should move the network call outside the tx to avoid
// holding locks during I/O; the fake sender in tests is synchronous so this is
// safe for the test suite.
//
// NOTE: ImportanceLegal deliveries additionally write a durable
// "notify.legal_delivery" outbox event carrying the provider's message id as
// receipt, in the same transaction as the 'sent' status update (DATA-08
// W0-T2). Migration 00011 grants app_platform INSERT on events_outbox for
// exactly this use case. Requires the Service to be wired with WithOutbox; if
// no outbox writer is configured, the event is skipped (no error) — the same
// nil-guard convention as kernel/attachment's optional outbox writer.
//
// Returns the number of deliveries successfully sent.
func (s *Service) SendPending(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, now time.Time) (int, error) {
	var sent int
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		// Claim queued and previously-failed (retriable) deliveries whose backoff
		// has elapsed. 'failed' = attempted but below the maxAttempts ceiling;
		// 'dead' = exhausted (never re-claimed). A queued delivery has a NULL
		// next_attempt_at (send immediately); a failed one carries its cooldown
		// deadline (ARCH-75). JOIN notifications for importance (legal audit note).
		rows, err := db.Query(ctx,
			`SELECT d.id, d.notification_id, d.channel, d.destination, d.attempts,
			        n.importance, d.trace_context
			   FROM notification_deliveries d
			   JOIN notifications n ON n.id = d.notification_id
			  WHERE d.status IN ('queued', 'failed')
			    AND (d.next_attempt_at IS NULL OR d.next_attempt_at <= $1)
			  ORDER BY d.created_at
			  FOR UPDATE OF d SKIP LOCKED
			  LIMIT $2`,
			now, claimBatch)
		if err != nil {
			return kerr.Wrapf(err, "notify.SendPending", "claim deliveries")
		}

		type claimed struct {
			id         uuid.UUID
			notifID    uuid.UUID
			channel    string
			dest       string
			attempts   int
			importance string
			trace      *string // W3C traceparent captured at Send (CA-9); nil when absent
		}
		var deliveries []claimed
		for rows.Next() {
			var d claimed
			if err := rows.Scan(&d.id, &d.notifID, &d.channel, &d.dest,
				&d.attempts, &d.importance, &d.trace); err != nil {
				rows.Close()
				return kerr.Wrapf(err, "notify.SendPending", "scan delivery")
			}
			deliveries = append(deliveries, d)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "notify.SendPending", "iterate deliveries")
		}

		for _, d := range deliveries {
			del := Delivery{
				ID:             d.id,
				TenantID:       tenantID,
				NotificationID: d.notifID,
				Channel:        Channel(d.channel),
				Destination:    d.dest,
				Status:         "queued",
				Attempts:       d.attempts,
			}

			// Continue the originating request's trace across the async boundary
			// (roadmap O1/CA-9): extract the traceparent captured at Send, then
			// deliver under a child span. Zero-cost with NoOpTracer.
			sendCtx := ctx
			if d.trace != nil && *d.trace != "" {
				sendCtx = s.tracer.Extract(ctx, *d.trace)
			}
			sendCtx, span := s.tracer.StartSpan(sendCtx, "notify.send "+d.channel)
			span.SetAttr("notify.channel", d.channel)
			span.SetAttr("notify.delivery_id", d.id.String())

			// An unregistered channel is a configuration error, not a transient
			// fault: fail the delivery loudly (and terminally) instead of routing
			// it to a no-op sender that would mark it 'sent' (roadmap CA-15).
			sender, ok := s.senderFor(del.Channel)
			var (
				providerMsgID string
				sendErr       error
			)
			permanent := false
			if !ok {
				sendErr = kerr.E(kerr.KindExternal, "no_channel_sender",
					"no sender registered for channel "+string(del.Channel))
				permanent = true
			} else {
				providerMsgID, sendErr = sender.Send(sendCtx, del)
			}
			if sendErr != nil {
				span.RecordError(sendErr)
			}
			span.End()

			newAttempts := d.attempts + 1
			if sendErr == nil {
				if _, err := db.Exec(ctx,
					`UPDATE notification_deliveries
					    SET status = 'sent', provider_message_id = $2,
					        attempts = $3, updated_at = $4
					  WHERE id = $1`,
					d.id, providerMsgID, newAttempts, now,
				); err != nil {
					return kerr.Wrapf(err, "notify.SendPending", "mark sent %s", d.id)
				}
				sent++
				// Legal delivery audit (DATA-08 W0-T2): a durable outbox event with
				// the provider's message id as receipt, written in the SAME
				// transaction as the 'sent' status update above so the audit trail
				// and the status advance commit or roll back together.
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
						return kerr.Wrapf(err, "notify.SendPending", "write legal delivery audit event %s", d.id)
					}
				}
			} else {
				newStatus := "failed"
				// A dead delivery is never re-claimed, so its next_attempt_at is
				// irrelevant (leave NULL); a failed one gets its cooldown deadline.
				var nextAttempt any
				// A permanent misconfiguration (no sender for the channel) cannot
				// be fixed by retrying, so it goes terminal immediately.
				if permanent || newAttempts >= maxAttempts {
					newStatus = "dead"
				} else {
					nextAttempt = now.Add(backoff(newAttempts))
				}
				if _, err := db.Exec(ctx,
					`UPDATE notification_deliveries
					    SET status = $2, attempts = $3, last_error = $4,
					        next_attempt_at = $5, updated_at = $6
					  WHERE id = $1`,
					d.id, newStatus, newAttempts, sendErr.Error(), nextAttempt, now,
				); err != nil {
					return kerr.Wrapf(err, "notify.SendPending", "mark %s %s", newStatus, d.id)
				}
			}
		}
		return nil
	})
	return sent, err
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
