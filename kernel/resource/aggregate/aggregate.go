// Package aggregate is the framework's mandatory aggregate write contract
// (DATA-06): one Writer.Write call performs a module aggregate's business-row
// write, the kernel resources-mirror upsert, an audit row, and an outbox
// event in the SAME tenant transaction. A fault at any stage rolls the whole
// transaction back — a module using this path structurally cannot produce a
// business row without its mirror, audit, and outbox companions.
//
// The write is attributed to a real actor resolved from context
// (ResolveActor): a user-initiated write with no resolvable actor fails fast;
// system-initiated paths (jobs, relays) are attributed to a deterministic
// system-actor id. The low-level resource.Registrar Upsert API remains
// available for not-yet-migrated callers; this package is the preferred —
// and, for new modules, the expected — write path.
//
// AR-03 compatibility note (RISK-W02-E04-001): the helper is deliberately a
// thin transactional composition over the existing registrar, audit, and
// outbox surfaces — it introduces no new persistent state or projection
// model, so a future authoritative-declaration/projection design (W05-E03)
// can replace its internals without changing the module-facing contract
// (business write in, mirror+audit+outbox guaranteed).
package aggregate

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// mirrorUpserter is the mirror stage seam. resource.PgRegistrar is the
// production implementation; the fault-injection suite substitutes failing
// decorators around it.
type mirrorUpserter interface {
	UpsertAs(ctx context.Context, db database.TenantDB, actorID uuid.UUID,
		ref resource.Ref, orgID *uuid.UUID, label, status string) error
}

// auditRecorder is the audit stage seam; audit.Writer is the production
// implementation.
type auditRecorder interface {
	Record(ctx context.Context, db database.TenantDB, e audit.Entry) error
}

// Writer performs aggregate writes under the mandatory-mirror contract. Wire
// one with New (a module builds it from its module.Context accessors: Tx,
// Audit, Outbox) and call Write per unit of work.
type Writer struct {
	tx     database.TxManager
	mirror mirrorUpserter
	audit  auditRecorder
	outbox outbox.Writer
}

// New wires the production aggregate writer from the framework's concrete
// components.
func New(tx database.TxManager, reg *resource.PgRegistrar, aud *audit.Writer, ob outbox.Writer) *Writer {
	return &Writer{tx: tx, mirror: reg, audit: aud, outbox: ob}
}

// Write describes one aggregate write. Apply is the module's business-row
// write — the only module-supplied stage; the framework owns the mirror,
// audit, and outbox stages. Audit.Action and Event.Type are required;
// EntityType/EntityID, Event.Resource, Event.Actor, and Audit.ActorKind
// default from Resource and the resolved actor.
type Write struct {
	Resource resource.Ref
	OrgID    *uuid.UUID
	Label    string
	Status   string
	Audit    audit.Entry
	Event    outbox.Event
	Apply    func(ctx context.Context, db database.TenantDB, actorID uuid.UUID) error
}

// Write runs the four-stage aggregate write contract in one tenant
// transaction: (1) the module's business write, (2) the resources-mirror
// upsert, (3) the audit row, (4) the outbox event. Any stage failing rolls
// back all of them. The resolved actor id is passed to Apply and bound into
// ctx before the transaction opens, so the business row, the mirror row, the
// audit row, and the app.actor_id GUC all carry the same attribution.
func (w *Writer) Write(ctx context.Context, in Write) error {
	switch {
	case in.Apply == nil:
		return kerr.E(kerr.KindInternal, "invalid_aggregate_write", "aggregate write requires Apply")
	case in.Resource.IsZero():
		return kerr.E(kerr.KindInternal, "invalid_aggregate_write", "aggregate write requires a resource ref")
	case in.Audit.Action == "":
		return kerr.E(kerr.KindInternal, "invalid_aggregate_write", "aggregate write requires an audit action")
	case in.Event.Type == "":
		return kerr.E(kerr.KindInternal, "invalid_aggregate_write", "aggregate write requires an outbox event type")
	}

	actorID, actorKind, err := ResolveActor(ctx)
	if err != nil {
		return err
	}
	// Rebind BEFORE the tx opens: TxManager mirrors the binding into the
	// app.actor_id GUC at BEGIN, and audit.Record reads the same key.
	ctx = database.WithActorID(ctx, actorID)

	if err := w.tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		// Stage 1 — business write (module-supplied).
		if err := in.Apply(ctx, db, actorID); err != nil {
			return kerr.Wrapf(err, "aggregate.Write", "business write for %s", in.Resource.Type)
		}
		// Stage 2 — resources-mirror upsert.
		if err := w.mirror.UpsertAs(ctx, db, actorID, in.Resource, in.OrgID, in.Label, in.Status); err != nil {
			return kerr.Wrapf(err, "aggregate.Write", "mirror upsert for %s", in.Resource.Type)
		}
		// Stage 3 — audit row.
		entry := in.Audit
		if entry.EntityType == "" {
			entry.EntityType = in.Resource.Type
		}
		if entry.EntityID == uuid.Nil {
			entry.EntityID = in.Resource.ID
		}
		if entry.ActorKind == "" {
			entry.ActorKind = actorKind
		}
		if err := w.audit.Record(ctx, db, entry); err != nil {
			return kerr.Wrapf(err, "aggregate.Write", "audit row for %s", in.Resource.Type)
		}
		// Stage 4 — outbox event.
		event := in.Event
		if event.Resource.IsZero() {
			event.Resource = in.Resource
		}
		if len(event.Actor) == 0 {
			event.Actor = actorDescriptor(actorID, actorKind)
		}
		if err := w.outbox.Write(ctx, db, event); err != nil {
			return kerr.Wrapf(err, "aggregate.Write", "outbox event for %s", in.Resource.Type)
		}
		return nil
	}); err != nil {
		return kerr.Wrapf(err, "aggregate.Write", "aggregate write for %s", in.Resource.Type)
	}
	return nil
}

// actorDescriptor renders the opaque outbox actor field for the resolved
// principal. Marshaling a flat two-string struct cannot fail.
func actorDescriptor(actorID uuid.UUID, kind string) json.RawMessage {
	raw, err := json.Marshal(struct {
		ID   uuid.UUID `json:"id"`
		Kind string    `json:"kind"`
	}{ID: actorID, Kind: kind})
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}
