package privileged

import (
	"context"
	"time"

	"github.com/google/uuid"

	kaudit "github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/relationship"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
)

// Relationships is the scoped privileged service for ReBAC relationship edges.
// It lets a module GRANT and REVOKE edges of a relationship type it owns,
// running with app_platform write privilege but tenant-bound so RLS still
// isolates. It absorbs, framework-side, every check a product's SECURITY DEFINER
// grant/revoke bridge functions would otherwise perform.
type Relationships struct {
	tx    database.TxManager // the app_platform, tenant-bindable manager
	audit *kaudit.Writer
	idgen model.IDGen
	own   *ownership
}

// GrantSpec describes an edge to create. RelType must be a relationship type the
// module owns (prefix or allow-list). Subject is the edge's subject side
// (kind + id, e.g. a capacity); Object is the resource the edge points at. The
// optional temporal window [ValidFrom, ValidTo) matches the framework edge
// shape; a zero ValidFrom means "now", a nil ValidTo means "open-ended". Actor
// is recorded as created_by and in the audit trail.
type GrantSpec struct {
	RelType     string
	SubjectKind string
	SubjectID   uuid.UUID
	Object      resource.Ref
	ValidFrom   time.Time
	ValidTo     *time.Time
	Actor       uuid.UUID
}

// Grant creates the relationship edge and writes an audit row, atomically, in a
// tenant-bound app_platform transaction. It enforces (in order): a bound tenant;
// module ownership of RelType; a valid temporal window; existence of the subject
// (an active acting capacity, when SubjectKind is capacity) and of the object
// resource — both in the caller's bound tenant. Returns the new edge id.
//
// Concurrency: two concurrent grants each insert their own edge (edges are
// many-cardinality by default); tenant isolation and the resource-existence
// checks are re-evaluated inside the transaction under the same snapshot as the
// insert, so a resource deleted concurrently cannot slip a dangling edge past
// the FK/RLS. This reproduces the bridge's guarantee without a product function.
func (r *Relationships) Grant(ctx context.Context, spec GrantSpec) (uuid.UUID, error) {
	if err := requireTenant(ctx); err != nil {
		return uuid.Nil, err
	}
	if !r.own.ownsRelType(spec.RelType) {
		return uuid.Nil, r.own.denyRelType(spec.RelType)
	}
	if spec.SubjectID == uuid.Nil {
		return uuid.Nil, kerr.E(kerr.KindValidation, "invalid_subject", "relationship subject id is required")
	}
	if spec.Object.ID == uuid.Nil || spec.Object.Type == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "invalid_object", "relationship object resource is required")
	}
	from := spec.ValidFrom
	if from.IsZero() {
		from = time.Now()
	}
	if spec.ValidTo != nil && !spec.ValidTo.After(from) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "invalid_window",
			"relationship valid_to must be after valid_from")
	}

	id := r.idgen.New()
	err := r.tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		// Subject existence: when the subject is a capacity, it must be a real,
		// currently-active acting capacity in the bound tenant (the bridge's
		// capacity-invalid check). RLS scopes the SELECT to the
		// Kind integrity: the edge's subject/object kinds must match what the
		// relationship TYPE declares, so a caller cannot point a party-subject type
		// at a capacity id (or vice versa). relationship_types is the global registry
		// (app_platform holds SELECT); a missing type is a validation error, not a raw
		// FK violation. This also confirms the type exists before we touch tenant rows.
		var declSubjectKind, declObjectKind string
		if err := db.QueryRow(ctx,
			`SELECT subject_kind, object_kind FROM relationship_types WHERE key = $1`, spec.RelType).
			Scan(&declSubjectKind, &declObjectKind); err != nil {
			if isNoRows(err) {
				return kerr.E(kerr.KindValidation, "unknown_rel_type",
					"relationship type "+spec.RelType+" is not registered")
			}
			return kerr.Wrapf(err, "privileged.Relationships.Grant", "load relationship type")
		}
		if subjectKind(spec.SubjectKind) != declSubjectKind {
			return kerr.E(kerr.KindValidation, "subject_kind_mismatch",
				"relationship type "+spec.RelType+" requires subject kind "+declSubjectKind)
		}
		// The framework grant path targets a resource object (a kernel Ref); refuse a
		// type whose declared object side is not a resource.
		if declObjectKind != relationship.KindResource {
			return kerr.E(kerr.KindValidation, "object_kind_unsupported",
				"relationship type "+spec.RelType+" object side is "+declObjectKind+", not resource")
		}

		// bound tenant, so a foreign-tenant capacity is invisible here.
		if subjectKind(spec.SubjectKind) == relationship.KindCapacity {
			ok, err := existsActiveCapacity(ctx, db, spec.SubjectID)
			if err != nil {
				return err
			}
			if !ok {
				return kerr.E(kerr.KindNotFound, "subject_not_found",
					"relationship subject capacity does not exist or is not active in this tenant")
			}
		}
		// Object existence: the resource must be registered in this tenant with the
		// declared type (the bridge's resource-invalid check).
		ok, err := existsResource(ctx, db, spec.Object)
		if err != nil {
			return err
		}
		if !ok {
			return kerr.E(kerr.KindNotFound, "object_not_found",
				"relationship object resource does not exist in this tenant")
		}

		// Insert the edge. tenant_id = app_tenant_id() so the RLS WITH CHECK on
		// relationships holds; valid_to nullable for open-ended edges.
		if _, err := db.Exec(ctx,
			`INSERT INTO relationships
			    (id, tenant_id, rel_type, subject_kind, subject_id, object_kind, object_id,
			     valid_from, valid_to, version, created_at, created_by)
			 VALUES ($1, app_tenant_id(), $2, $3, $4, 'resource', $5, $6, $7, 1, now(), $8)`,
			id, spec.RelType, subjectKind(spec.SubjectKind), spec.SubjectID, spec.Object.ID,
			from, spec.ValidTo, spec.Actor); err != nil {
			return kerr.Wrapf(err, "privileged.Relationships.Grant", "insert %s edge", spec.RelType)
		}

		return r.audit.Record(ctx, db, kaudit.Entry{
			Action:     "relationship.grant",
			EntityType: "relationship",
			EntityID:   id,
			ActorKind:  "system",
			Metadata: map[string]any{
				"rel_type":     spec.RelType,
				"subject_kind": subjectKind(spec.SubjectKind),
				"subject_id":   spec.SubjectID.String(),
				"object_type":  spec.Object.Type,
				"object_id":    spec.Object.ID.String(),
				"module":       r.own.module,
			},
		})
	})
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// Revoke soft-revokes an owned edge: it sets valid_to = now() (never deletes, so
// the historical grant survives — audit-friendly, matching a revoke bridge's
// soft-revoke semantics), bumps version, and writes an audit row. It
// enforces a bound tenant, that the edge's rel_type is owned by this module, and
// tenant scope (RLS + an explicit re-check under FOR UPDATE). A missing or
// already-revoked edge is reported; a double-revoke is a no-op conflict rather
// than a silent success.
func (r *Relationships) Revoke(ctx context.Context, id, actor uuid.UUID) error {
	if err := requireTenant(ctx); err != nil {
		return err
	}
	if id == uuid.Nil {
		return kerr.E(kerr.KindValidation, "invalid_id", "relationship id is required")
	}
	return r.tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		// Lock the row for the read-modify-write so a concurrent revoke cannot
		// race us to a double close (bridge used SELECT … FOR UPDATE). RLS scopes
		// the row to the bound tenant, so a foreign-tenant id is simply NotFound.
		var relType string
		err := db.QueryRow(ctx,
			`SELECT rel_type FROM relationships WHERE id = $1 FOR UPDATE`, id).Scan(&relType)
		if isNoRows(err) {
			return kerr.E(kerr.KindNotFound, "not_found", "relationship edge not found")
		}
		if err != nil {
			return kerr.Wrapf(err, "privileged.Relationships.Revoke", "load edge")
		}
		// Ownership is checked on the STORED rel_type: a module may not revoke an
		// edge of a type it does not own even if it knows the id.
		if !r.own.ownsRelType(relType) {
			return r.own.denyRelType(relType)
		}
		// Close the edge only if it is still open at the DB's clock — the guard is
		// evaluated with the same now() as the write, so a double-revoke can never
		// slip through on host/DB clock skew. Zero rows affected ⇒ already revoked.
		tag, err := db.Exec(ctx,
			`UPDATE relationships
			    SET valid_to = now(), updated_by = $2, updated_at = now(), version = version + 1
			  WHERE id = $1 AND (valid_to IS NULL OR valid_to > now())`, id, nullUUID(actor))
		if err != nil {
			return kerr.Wrapf(err, "privileged.Relationships.Revoke", "revoke edge")
		}
		if tag.RowsAffected() == 0 {
			return kerr.E(kerr.KindConflict, "already_revoked", "relationship edge is already revoked")
		}
		return r.audit.Record(ctx, db, kaudit.Entry{
			Action:     "relationship.revoke",
			EntityType: "relationship",
			EntityID:   id,
			ActorKind:  "system",
			Metadata:   map[string]any{"rel_type": relType, "module": r.own.module},
		})
	})
}

// subjectKind defaults an unset subject kind to capacity — the identity a human
// actor carries and the only kind the ReBAC checker consults today.
func subjectKind(k string) string {
	if k == "" {
		return relationship.KindCapacity
	}
	return k
}

func existsActiveCapacity(ctx context.Context, db database.TenantDB, id uuid.UUID) (bool, error) {
	var ok bool
	if err := db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM acting_capacities WHERE id = $1 AND valid_to IS NULL)`, id).
		Scan(&ok); err != nil {
		return false, kerr.Wrapf(err, "privileged.Relationships.Grant", "check subject capacity")
	}
	return ok, nil
}

func existsResource(ctx context.Context, db database.TenantDB, ref resource.Ref) (bool, error) {
	var ok bool
	if err := db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM resources WHERE id = $1 AND resource_type = $2)`, ref.ID, ref.Type).
		Scan(&ok); err != nil {
		return false, kerr.Wrapf(err, "privileged.Relationships.Grant", "check object resource")
	}
	return ok, nil
}

func nullUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
