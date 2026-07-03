package attachment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// Attachment is a persisted attachment row.
type Attachment struct {
	ID                uuid.UUID
	Resource          resource.Ref
	DocumentVersionID uuid.UUID
	CommentID         *uuid.UUID
	WorkflowTaskID    *uuid.UUID
	Status            string
	Version           int
	CreatedAt         time.Time
	CreatedBy         uuid.UUID
}

// AttachIn is the input to Attach.
type AttachIn struct {
	Resource          resource.Ref
	DocumentVersionID uuid.UUID
	CommentID         *uuid.UUID
	WorkflowTaskID    *uuid.UUID
}

// Service manages attachments inside a tenant transaction.
type Service struct {
	idgen model.IDGen
	ob    outbox.Writer // optional; nil-guarded
}

// New wires the service. ob may be nil.
func New(idgen model.IDGen, ob outbox.Writer) *Service {
	return &Service{idgen: idgen, ob: ob}
}

// Attach inserts a new attachment for a resource. Pre-checks document version
// existence via SELECT for a clean error rather than relying on the FK violation.
func (s *Service) Attach(ctx context.Context, db database.TenantDB, in AttachIn) (uuid.UUID, error) {
	if in.Resource.IsZero() {
		return uuid.Nil, kerr.E(kerr.KindValidation, "attachment_resource_required", "resource ref is required")
	}
	if in.DocumentVersionID == uuid.Nil {
		return uuid.Nil, kerr.E(kerr.KindValidation, "attachment_document_version_required", "document version id is required")
	}
	// Pre-check: verify the document version exists in this tenant (RLS-scoped)
	// AND is still active — attaching a voided (retention-tombstoned) version would
	// link a resource to a destroyed blob (ARCH-68).
	var exists bool
	if err := db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM document_versions WHERE id = $1 AND status = 'active')`, in.DocumentVersionID).
		Scan(&exists); err != nil {
		return uuid.Nil, kerr.Wrapf(err, "attachment.Attach", "check document version")
	}
	if !exists {
		return uuid.Nil, kerr.E(kerr.KindNotFound, "document_version_not_found", "active document version not found")
	}

	id := s.idgen.New()
	_, err := db.Exec(ctx,
		`INSERT INTO attachments
		    (id, tenant_id, resource_type, resource_id, document_version_id, comment_id, workflow_task_id, status, version, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, 'active', 1, $7)`,
		id, in.Resource.Type, in.Resource.ID, in.DocumentVersionID,
		uuidPtrArg(in.CommentID), uuidPtrArg(in.WorkflowTaskID), actorFrom(ctx))
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "attachment.Attach", "insert attachment")
	}
	if s.ob != nil {
		_ = s.ob.Write(ctx, db, outbox.Event{
			Type:     "attachment.created",
			Resource: in.Resource,
			Payload:  map[string]any{"attachment_id": id.String()},
		})
	}
	return id, nil
}

// Detach voids an attachment with optimistic locking. Void != delete.
func (s *Service) Detach(ctx context.Context, db database.TenantDB, id uuid.UUID, expectedVersion int) error {
	// Only the actor who created the attachment may detach it (SEC-46); fails
	// closed when no actor is bound.
	var createdBy uuid.UUID
	err := db.QueryRow(ctx, `SELECT created_by FROM attachments WHERE id = $1`, id).Scan(&createdBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return kerr.E(kerr.KindNotFound, "attachment_not_found", "attachment not found")
	}
	if err != nil {
		return kerr.Wrapf(err, "attachment.Detach", "load attachment")
	}
	a, ok := database.ActorIDFrom(ctx)
	if !ok || a == uuid.Nil || a != createdBy {
		return kerr.E(kerr.KindForbidden, "attachment_forbidden", "only the actor who attached it may detach it")
	}
	tag, err := db.Exec(ctx,
		`UPDATE attachments
		    SET status = 'voided', version = version + 1
		  WHERE id = $1 AND version = $2`,
		id, expectedVersion)
	if err != nil {
		return kerr.Wrapf(err, "attachment.Detach", "void attachment")
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindVersionConflict, "version_conflict", "optimistic lock conflict on "+id.String())
	}
	return nil
}

// List returns all attachments (active and voided) for the given resource,
// ordered by created_at ASC.
func (s *Service) List(ctx context.Context, db database.TenantDB, ref resource.Ref) ([]Attachment, error) {
	rows, err := db.Query(ctx,
		`SELECT id, resource_type, resource_id, document_version_id, comment_id, workflow_task_id,
		        status, version, created_at, created_by
		   FROM attachments
		  WHERE resource_type = $1 AND resource_id = $2
		  ORDER BY created_at ASC`, ref.Type, ref.ID)
	if err != nil {
		return nil, kerr.Wrapf(err, "attachment.List", "query attachments")
	}
	defer rows.Close()
	var out []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(
			&a.ID, &a.Resource.Type, &a.Resource.ID,
			&a.DocumentVersionID, &a.CommentID, &a.WorkflowTaskID,
			&a.Status, &a.Version, &a.CreatedAt, &a.CreatedBy,
		); err != nil {
			return nil, kerr.Wrapf(err, "attachment.List", "scan attachment")
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "attachment.List", "iterate attachments")
	}
	return out, nil
}

// actorFrom returns the actor id from ctx, falling back to uuid.Nil (for NOT NULL created_by).
func actorFrom(ctx context.Context) uuid.UUID {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return uuid.Nil
}

func uuidPtrArg(p *uuid.UUID) any {
	if p == nil {
		return nil
	}
	return *p
}
