// Package outbox is wowapi's transactional outbox: modules write domain events
// into events_outbox in the SAME transaction as their business writes, so an
// event is emitted if and only if the write commits (no lost or phantom
// events). A relay later claims pending events and dispatches them to
// idempotent handlers, deduped via the processed_events inbox. Contract:
// blueprint 07 §3/§7.
package outbox

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// Event is the outbox envelope. Type is "module.resource.verb_past"; Payload is
// the module's event struct (additive within a SchemaVersion). ID and
// OccurredAt are assigned on write when zero.
type Event struct {
	ID            uuid.UUID
	Type          string
	SchemaVersion int
	Resource      resource.Ref
	Actor         json.RawMessage // opaque actor descriptor; never a secret
	Payload       any
	// TenantID is set by the writer from the tx's tenant; callers leave it zero.
	TenantID uuid.UUID
}

// Writer writes events into the outbox within the caller's tenant transaction.
// Stateless: Write takes the tx's TenantDB so the event commits atomically with
// the business write.
type Writer interface {
	Write(ctx context.Context, db database.TenantDB, e Event) error
}

// NewWriter returns the Postgres outbox writer. idgen mints event ids
// (UUIDv7 — time-ordered, so per-aggregate dispatch order is natural).
func NewWriter(idgen model.IDGen) Writer { return pgWriter{idgen: idgen} }

type pgWriter struct{ idgen model.IDGen }

func (w pgWriter) Write(ctx context.Context, db database.TenantDB, e Event) error {
	if e.Type == "" {
		return kerr.E(kerr.KindInternal, "invalid_event", "outbox event requires a Type")
	}
	id := e.ID
	if id == uuid.Nil {
		id = w.idgen.New()
	}
	sv := e.SchemaVersion
	if sv == 0 {
		sv = 1
	}
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		return kerr.E(kerr.KindInternal, "invalid_event", "outbox payload not JSON-encodable")
	}
	actor := e.Actor
	if len(actor) == 0 {
		actor = json.RawMessage("{}")
	}
	var resType any
	var resID any
	if !e.Resource.IsZero() {
		resType = e.Resource.Type
		resID = e.Resource.ID
	}
	// tenant_id from app_tenant_id() so the RLS WITH CHECK holds and the event
	// is bound to the same tenant as the business write.
	_, err = db.Exec(ctx,
		`INSERT INTO events_outbox
             (id, tenant_id, event_type, schema_version, resource_type, resource_id, actor, payload, created_by)
         VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, '00000000-0000-0000-0000-000000000000')`,
		id, e.Type, sv, resType, resID, actor, payload)
	if err != nil {
		return kerr.Wrapf(err, "outbox.Write", "insert event %s", e.Type)
	}
	return nil
}

// Handler processes a dispatched event within a tenant transaction. It must be
// idempotent (the inbox dedups redelivery, but a handler should tolerate it).
type Handler func(ctx context.Context, db database.TenantDB, e DispatchedEvent) error

// DispatchedEvent is what a handler receives: the envelope plus the raw payload
// bytes (the handler unmarshals into its own typed struct).
type DispatchedEvent struct {
	ID            uuid.UUID
	Type          string
	SchemaVersion int
	Resource      resource.Ref
	Actor         json.RawMessage
	Payload       json.RawMessage
	TenantID      uuid.UUID
}

// subscription is one registered handler for an event type.
type subscription struct {
	eventType string
	name      string // handler name, keys the inbox dedup
	fn        Handler
}

// HandlerRegistry collects event subscriptions during module registration.
type HandlerRegistry struct {
	subs []subscription
	seen map[string]bool // eventType+name dedup
	errs []error
}

// NewHandlerRegistry returns an empty registry.
func NewHandlerRegistry() *HandlerRegistry { return &HandlerRegistry{seen: map[string]bool{}} }

// Subscribe registers an idempotent handler for an event type. handlerName must
// be unique per event type and stable across deploys (it keys the inbox).
func (r *HandlerRegistry) Subscribe(eventType, handlerName string, fn Handler) {
	if eventType == "" || handlerName == "" || fn == nil {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_subscription",
			"Subscribe requires eventType, handlerName, and fn"))
		return
	}
	key := eventType + "\x00" + handlerName
	if r.seen[key] {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "duplicate_subscription",
			"handler "+handlerName+" subscribed to "+eventType+" more than once"))
		return
	}
	r.seen[key] = true
	r.subs = append(r.subs, subscription{eventType: eventType, name: handlerName, fn: fn})
}

// handlersFor returns the subscriptions for an event type.
func (r *HandlerRegistry) handlersFor(eventType string) []subscription {
	var out []subscription
	for _, s := range r.subs {
		if s.eventType == eventType {
			out = append(out, s)
		}
	}
	return out
}

// Err returns accumulated subscription errors joined, or nil.
func (r *HandlerRegistry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	msgs := make([]string, len(r.errs))
	for i, e := range r.errs {
		msgs[i] = e.Error()
	}
	joined := msgs[0]
	for i := 1; i < len(msgs); i++ {
		joined += "; " + msgs[i]
	}
	return kerr.E(kerr.KindInternal, "subscription_failed", "event subscription failed: "+joined)
}
