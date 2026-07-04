package retention

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
)

// DSR is the Data Subject Request ledger (export / erasure). It tracks the
// request lifecycle and the statutory-override reason when an erasure is refused
// because a retention obligation (or a legal hold) forbids it. The actual data
// export/erasure is performed by product-registered callbacks per record class;
// this ledger is the auditable request record.
type DSR struct {
	idgen model.IDGen
}

// NewDSR builds the DSR ledger service.
func NewDSR(idgen model.IDGen) *DSR {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &DSR{idgen: idgen}
}

// Kind is the DSR type.
type Kind string

const (
	KindExport  Kind = "export"
	KindErasure Kind = "erasure"
)

// Request is a DSR ledger row.
type Request struct {
	ID             uuid.UUID
	SubjectRef     string
	Kind           Kind
	Status         string // pending | completed | rejected
	OverrideReason string
}

// Open records a new DSR for a subject. subjectRef is the product's subject
// identifier. Runs in the caller's tenant transaction.
func (d *DSR) Open(ctx context.Context, db database.TenantDB, subjectRef string, kind Kind) (uuid.UUID, error) {
	if subjectRef == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "invalid_dsr", "subject ref is required")
	}
	if kind != KindExport && kind != KindErasure {
		return uuid.Nil, kerr.E(kerr.KindValidation, "invalid_dsr", "kind must be export or erasure")
	}
	id := d.idgen.New()
	if _, err := db.Exec(ctx,
		`INSERT INTO dsr_requests (id, tenant_id, subject_ref, kind, requested_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4)`,
		id, subjectRef, string(kind), actorOrNil(ctx)); err != nil {
		return uuid.Nil, kerr.Wrapf(err, "dsr.Open", "insert request")
	}
	return id, nil
}

// Complete marks a pending request fulfilled (after the product has performed the
// export/erasure). KindConflict if the request is not pending.
func (d *DSR) Complete(ctx context.Context, db database.TenantDB, id uuid.UUID) error {
	return d.transition(ctx, db, id, "completed", "")
}

// Reject marks a pending request refused, recording the statutory-override reason
// (e.g. "retained under §X for 7 years"). KindConflict if not pending.
func (d *DSR) Reject(ctx context.Context, db database.TenantDB, id uuid.UUID, overrideReason string) error {
	if overrideReason == "" {
		return kerr.E(kerr.KindValidation, "invalid_dsr", "a reason is required to reject a DSR")
	}
	return d.transition(ctx, db, id, "rejected", overrideReason)
}

func (d *DSR) transition(ctx context.Context, db database.TenantDB, id uuid.UUID, status, reason string) error {
	tag, err := db.Exec(ctx,
		`UPDATE dsr_requests
		    SET status = $2, override_reason = NULLIF($3, ''), completed_at = now()
		  WHERE id = $1 AND status = 'pending'`, id, status, reason)
	if err != nil {
		return kerr.Wrapf(err, "dsr.transition", "set status %s", status)
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindConflict, "not_pending", "DSR is not pending (already completed or rejected)")
	}
	return nil
}

// Get reads a DSR request.
func (d *DSR) Get(ctx context.Context, db database.TenantDB, id uuid.UUID) (Request, error) {
	var r Request
	var kind string
	var override *string
	if err := db.QueryRow(ctx,
		`SELECT id, subject_ref, kind, status, override_reason FROM dsr_requests WHERE id = $1`, id).
		Scan(&r.ID, &r.SubjectRef, &kind, &r.Status, &override); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Request{}, kerr.E(kerr.KindNotFound, "not_found", "no such DSR request")
		}
		return Request{}, kerr.Wrapf(err, "dsr.Get", "read request")
	}
	r.Kind = Kind(kind)
	if override != nil {
		r.OverrideReason = *override
	}
	return r, nil
}
