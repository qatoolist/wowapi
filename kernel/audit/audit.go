// Package audit is the durable, append-only, field-level audit trail (roadmap
// E1): a standardized record of who changed what — entity, field, before/after,
// actor, capacity, impersonator, request id — written INSIDE the business
// transaction so an audit row commits iff the change does. Append-only is
// enforced by the grants (app_rt has no UPDATE/DELETE on audit_logs); this
// package never offers a mutate path. Cryptographic tamper-evidence
// (hash-chaining, S6) layers on top of this table later.
package audit

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
)

// Entry is a change to record. Action is required (e.g. "document.download",
// "receipt.void"); the rest are optional. For a field-level change set Field +
// OldValue + NewValue; for a whole-entity action leave Field empty. Values are
// passed through the Writer's redactor before persistence.
type Entry struct {
	Action         string
	EntityType     string
	EntityID       uuid.UUID // uuid.Nil → NULL
	Field          string
	OldValue       string
	NewValue       string
	Reason         string
	ActorKind      string    // user | system | webhook (optional)
	ImpersonatorID uuid.UUID // support impersonation (optional)
	Metadata       map[string]any
}

// Log is a persisted audit row returned by Query.
type Log struct {
	ID             uuid.UUID
	OccurredAt     time.Time
	ActorID        *uuid.UUID
	ActorKind      string
	ImpersonatorID *uuid.UUID
	RequestID      string
	Action         string
	EntityType     string
	EntityID       *uuid.UUID
	Field          string
	OldValue       string
	NewValue       string
	Reason         string
}

// Redactor may mutate an Entry before it is written — e.g. mask the values of
// known-sensitive fields so they never land in the audit table. It is the
// module's per-record redaction hook (blueprint 07 §1 "per-module redaction").
type Redactor func(*Entry)

// Writer appends and queries audit rows. It is stateless beyond its id generator
// and optional redactor.
type Writer struct {
	idgen  model.IDGen
	redact Redactor
}

// New builds a Writer. redact may be nil (no redaction).
func New(idgen model.IDGen, redact Redactor) *Writer {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &Writer{idgen: idgen, redact: redact}
}

// Record appends one audit row in db's transaction (so it commits with the
// business write). The acting actor id and the request id are read from ctx; the
// caller supplies the semantic fields via e. Action is required.
func (w *Writer) Record(ctx context.Context, db database.TenantDB, e Entry) error {
	if e.Action == "" {
		return kerr.E(kerr.KindValidation, "invalid_audit", "audit action is required")
	}
	if w.redact != nil {
		w.redact(&e)
	}
	meta := e.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return kerr.Wrapf(err, "audit.Record", "marshal metadata")
	}
	var actorID any
	if id, ok := database.ActorIDFrom(ctx); ok {
		actorID = id
	}
	_, err = db.Exec(ctx,
		`INSERT INTO audit_logs
		    (id, tenant_id, actor_id, actor_kind, impersonator_id, request_id,
		     action, entity_type, entity_id, field, old_value, new_value, reason, metadata)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		w.idgen.New(), actorID, nullStr(e.ActorKind), nullUUID(e.ImpersonatorID),
		nullStr(httpx.RequestIDFrom(ctx)), e.Action, nullStr(e.EntityType), nullUUID(e.EntityID),
		nullStr(e.Field), nullStr(e.OldValue), nullStr(e.NewValue), nullStr(e.Reason), metaJSON)
	if err != nil {
		return kerr.Wrapf(err, "audit.Record", "insert audit row")
	}
	return nil
}

// Filter narrows a Query. Zero-valued fields are ignored; Limit defaults to 100.
type Filter struct {
	EntityType string
	EntityID   uuid.UUID
	ActorID    uuid.UUID
	Action     string
	Limit      int
}

// Query returns audit rows matching the filter, newest first, in the caller's
// tenant transaction (RLS-scoped). All filter values are bound as parameters.
func (w *Writer) Query(ctx context.Context, db database.TenantDB, f Filter) ([]Log, error) {
	conds := []string{"true"}
	args := []any{}
	add := func(clause string, val any) {
		args = append(args, val)
		conds = append(conds, clause+" $"+strconv.Itoa(len(args)))
	}
	if f.EntityType != "" {
		add("entity_type =", f.EntityType)
	}
	if f.EntityID != uuid.Nil {
		add("entity_id =", f.EntityID)
	}
	if f.ActorID != uuid.Nil {
		add("actor_id =", f.ActorID)
	}
	if f.Action != "" {
		add("action =", f.Action)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	args = append(args, limit)
	sql := `SELECT id, occurred_at, actor_id, actor_kind, impersonator_id, request_id,
	               action, entity_type, entity_id, field, old_value, new_value, reason
	          FROM audit_logs
	         WHERE ` + strings.Join(conds, " AND ") +
		// id (UUIDv7) is a creation-ordered tiebreaker so rows written in the same
		// transaction (identical occurred_at) still sort newest-first.
		` ORDER BY occurred_at DESC, id DESC LIMIT $` + strconv.Itoa(len(args))

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, kerr.Wrapf(err, "audit.Query", "query audit logs")
	}
	defer rows.Close()
	var out []Log
	for rows.Next() {
		var l Log
		var actorKind, requestID, entityType, field, oldV, newV, reason *string
		if err := rows.Scan(&l.ID, &l.OccurredAt, &l.ActorID, &actorKind, &l.ImpersonatorID,
			&requestID, &l.Action, &entityType, &l.EntityID, &field, &oldV, &newV, &reason); err != nil {
			return nil, kerr.Wrapf(err, "audit.Query", "scan audit row")
		}
		l.ActorKind = deref(actorKind)
		l.RequestID = deref(requestID)
		l.EntityType = deref(entityType)
		l.Field = deref(field)
		l.OldValue = deref(oldV)
		l.NewValue = deref(newV)
		l.Reason = deref(reason)
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "audit.Query", "iterate audit logs")
	}
	return out, nil
}

// --- helpers: NULL-safe binding + scanning ---

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
