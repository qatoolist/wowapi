package comment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
)

// Comment is a persisted comment row.
type Comment struct {
	ID               uuid.UUID
	Resource         resource.Ref
	ParentID         *uuid.UUID
	AuthorCapacityID uuid.UUID
	Body             string
	Status           string
	Version          int
	CreatedAt        time.Time
	CreatedBy        uuid.UUID
	UpdatedAt        *time.Time
	UpdatedBy        *uuid.UUID
}

// CreateIn is the input to Create.
type CreateIn struct {
	Resource         resource.Ref
	ParentID         *uuid.UUID
	AuthorCapacityID uuid.UUID
	Body             string
}

// Service manages comments inside a tenant transaction.
type Service struct {
	idgen model.IDGen
	ob    outbox.Writer // optional; nil-guarded
}

// New wires the service. ob may be nil.
func New(idgen model.IDGen, ob outbox.Writer) *Service {
	return &Service{idgen: idgen, ob: ob}
}

// Create inserts a new comment. Validates inputs; checks parent thread consistency.
func (s *Service) Create(ctx context.Context, db database.TenantDB, in CreateIn) (uuid.UUID, error) {
	if in.Resource.IsZero() {
		return uuid.Nil, kerr.E(kerr.KindValidation, "comment_resource_required", "resource ref is required")
	}
	if strings.TrimSpace(in.Body) == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "comment_body_required", "comment body must not be empty")
	}
	if in.AuthorCapacityID == uuid.Nil {
		return uuid.Nil, kerr.E(kerr.KindValidation, "comment_author_required", "author capacity id is required")
	}
	if in.ParentID != nil {
		// Verify parent exists in this tenant and anchors the same resource.
		var pResType string
		var pResID uuid.UUID
		err := db.QueryRow(ctx,
			`SELECT resource_type, resource_id FROM comments WHERE id = $1`, *in.ParentID).
			Scan(&pResType, &pResID)
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, kerr.E(kerr.KindNotFound, "comment_parent_not_found", "parent comment not found")
		}
		if err != nil {
			return uuid.Nil, kerr.Wrapf(err, "comment.Create", "load parent")
		}
		if pResType != in.Resource.Type || pResID != in.Resource.ID {
			return uuid.Nil, kerr.E(kerr.KindValidation, "comment_parent_resource_mismatch",
				"parent comment belongs to a different resource")
		}
	}

	id := s.idgen.New()
	_, err := db.Exec(ctx,
		`INSERT INTO comments
		    (id, tenant_id, resource_type, resource_id, parent_comment_id, author_capacity_id, body, status, version, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, 'active', 1, $7)`,
		id, in.Resource.Type, in.Resource.ID, uuidPtrArg(in.ParentID), in.AuthorCapacityID, in.Body, actorFrom(ctx))
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "comment.Create", "insert comment")
	}
	if s.ob != nil {
		_ = s.ob.Write(ctx, db, outbox.Event{
			Type:     "comment.created",
			Resource: in.Resource,
			Payload:  map[string]any{"comment_id": id.String()},
		})
	}
	return id, nil
}

// Edit updates a comment body with optimistic locking. Voided comments cannot be edited.
// When ob is wired, emits "comment.edited" with previous and new body for audit history.
func (s *Service) Edit(ctx context.Context, db database.TenantDB, id uuid.UUID, expectedVersion int, newBody string) error {
	if strings.TrimSpace(newBody) == "" {
		return kerr.E(kerr.KindValidation, "comment_body_required", "comment body must not be empty")
	}
	// Load to check status, capture prior body for history, and authorize.
	var curStatus, prevBody, resType string
	var resID, author, createdBy uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT status, body, resource_type, resource_id, author_capacity_id, created_by FROM comments WHERE id = $1`, id).
		Scan(&curStatus, &prevBody, &resType, &resID, &author, &createdBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return kerr.E(kerr.KindNotFound, "comment_not_found", "comment not found")
	}
	if err != nil {
		return kerr.Wrapf(err, "comment.Edit", "load comment")
	}
	if err := ensureAuthor(ctx, author, createdBy); err != nil {
		return err
	}
	if curStatus == "voided" {
		return kerr.E(kerr.KindConflict, "comment_voided", "voided comments cannot be edited")
	}
	tag, err := db.Exec(ctx,
		`UPDATE comments
		    SET body = $3, status = 'edited', version = version + 1,
		        updated_at = now(), updated_by = $4
		  WHERE id = $1 AND version = $2`,
		id, expectedVersion, newBody, actorNullArg(ctx))
	if err != nil {
		return kerr.Wrapf(err, "comment.Edit", "update comment")
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindVersionConflict, "version_conflict", "optimistic lock conflict on "+id.String())
	}
	if s.ob != nil {
		_ = s.ob.Write(ctx, db, outbox.Event{
			Type:     "comment.edited",
			Resource: resource.Ref{Type: resType, ID: resID},
			Payload:  map[string]any{"comment_id": id.String(), "previous_body": prevBody, "new_body": newBody},
		})
	}
	return nil
}

// Void sets a comment's status to 'voided' with optimistic locking. Only the
// comment's author may void it (SEC-45).
func (s *Service) Void(ctx context.Context, db database.TenantDB, id uuid.UUID, expectedVersion int) error {
	var author, createdBy uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT author_capacity_id, created_by FROM comments WHERE id = $1`, id).Scan(&author, &createdBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return kerr.E(kerr.KindNotFound, "comment_not_found", "comment not found")
	}
	if err != nil {
		return kerr.Wrapf(err, "comment.Void", "load comment")
	}
	if err := ensureAuthor(ctx, author, createdBy); err != nil {
		return err
	}
	tag, err := db.Exec(ctx,
		`UPDATE comments
		    SET status = 'voided', version = version + 1,
		        updated_at = now(), updated_by = $3
		  WHERE id = $1 AND version = $2`,
		id, expectedVersion, actorNullArg(ctx))
	if err != nil {
		return kerr.Wrapf(err, "comment.Void", "void comment")
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindVersionConflict, "version_conflict", "optimistic lock conflict on "+id.String())
	}
	return nil
}

// ensureAuthor authorizes an edit/void: the context actor must be the comment's
// author (its authoring capacity or its creating actor). Fails closed when no
// actor is bound (SEC-45).
func ensureAuthor(ctx context.Context, author, createdBy uuid.UUID) error {
	a, ok := database.ActorIDFrom(ctx)
	if ok && a != uuid.Nil && (a == author || a == createdBy) {
		return nil
	}
	return kerr.E(kerr.KindForbidden, "comment_forbidden", "only the comment author may modify it")
}

// List returns all comments for the given resource, ordered by created_at ASC.
// Includes all statuses (active, edited, voided).
func (s *Service) List(ctx context.Context, db database.TenantDB, ref resource.Ref) ([]Comment, error) {
	rows, err := db.Query(ctx,
		`SELECT id, resource_type, resource_id, parent_comment_id, author_capacity_id,
		        body, status, version, created_at, created_by, updated_at, updated_by
		   FROM comments
		  WHERE resource_type = $1 AND resource_id = $2
		  ORDER BY created_at ASC`, ref.Type, ref.ID)
	if err != nil {
		return nil, kerr.Wrapf(err, "comment.List", "query comments")
	}
	defer rows.Close()
	var out []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(
			&c.ID, &c.Resource.Type, &c.Resource.ID,
			&c.ParentID, &c.AuthorCapacityID,
			&c.Body, &c.Status, &c.Version,
			&c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy,
		); err != nil {
			return nil, kerr.Wrapf(err, "comment.List", "scan comment")
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "comment.List", "iterate comments")
	}
	return out, nil
}

// actorFrom returns the actor id from ctx, falling back to uuid.Nil (for NOT NULL columns).
func actorFrom(ctx context.Context) uuid.UUID {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return uuid.Nil
}

// actorNullArg returns the actor id as any, returning nil (SQL NULL) when absent.
func actorNullArg(ctx context.Context) any {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return nil
}

func uuidPtrArg(p *uuid.UUID) any {
	if p == nil {
		return nil
	}
	return *p
}
