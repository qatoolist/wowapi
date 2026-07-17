// Package artifact is the snapshot/artifact pipeline (roadmap E4): it turns a
// product-rendered dataset into an IMMUTABLE, versioned artifact — content plus
// its sha256, a structured sidecar, and the template version/effective date it
// was produced under. The framework owns immutability (append-only grants),
// per-(tenant,kind) versioning, hashing, and tamper-verification; the product
// supplies the rendered bytes (e.g. a PDF/A from its own renderer), so no
// document-format library enters the kernel.
package artifact

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
)

// Pipeline generates and reads artifacts.
type Pipeline struct {
	idgen model.IDGen
}

// New builds the pipeline.
func New(idgen model.IDGen) *Pipeline {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &Pipeline{idgen: idgen}
}

// Input describes one artifact to generate. Content is the product-rendered bytes
// (the framework hashes and stores them verbatim — it never re-encodes them).
type Input struct {
	Kind            string // 'receipt', 'certificate', …
	Content         []byte
	ContentType     string // default application/pdf
	Sidecar         map[string]any
	TemplateVersion string
	EffectiveDate   time.Time // zero → NULL
}

// Artifact is a stored artifact's metadata (Content is included only by Get).
type Artifact struct {
	ID              uuid.UUID
	Kind            string
	Version         int
	ContentHash     string
	ContentType     string
	Sidecar         map[string]any
	TemplateVersion string
	Content         []byte
}

// Generate hashes and stores an immutable artifact in the caller's tenant tx,
// assigning the next per-(tenant,kind) version. Content is required. A concurrent
// generate of the same kind that collides on version returns KindConflict (retry).
func (p *Pipeline) Generate(ctx context.Context, db database.TenantDB, in Input) (Artifact, error) {
	if in.Kind == "" || len(in.Content) == 0 {
		return Artifact{}, kerr.E(kerr.KindValidation, "invalid_artifact", "kind and content are required")
	}
	ct := in.ContentType
	if ct == "" {
		ct = "application/pdf"
	}
	sidecar := in.Sidecar
	if sidecar == nil {
		sidecar = map[string]any{}
	}
	sidecarJSON, err := json.Marshal(sidecar)
	if err != nil {
		return Artifact{}, kerr.Wrapf(err, "artifact.Generate", "marshal sidecar")
	}
	sum := sha256.Sum256(in.Content)
	hash := hex.EncodeToString(sum[:])
	id := p.idgen.New()

	var version int
	var effective any
	if !in.EffectiveDate.IsZero() {
		effective = in.EffectiveDate
	}
	if err := db.QueryRow(ctx,
		`INSERT INTO version_counters (tenant_id, scope, value)
		 VALUES (app_tenant_id(), $1, 1)
		 ON CONFLICT (tenant_id, scope) DO UPDATE SET value = version_counters.value + 1
		 RETURNING value`,
		"artifact:"+in.Kind).Scan(&version); err != nil {
		return Artifact{}, kerr.Wrapf(err, "artifact.Generate", "allocate version")
	}
	err = db.QueryRow(ctx,
		`INSERT INTO artifacts
		    (id, tenant_id, kind, version, content_hash, content, content_type, sidecar, template_version, effective_date, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING version`,
		id, in.Kind, version, hash, in.Content, ct, sidecarJSON, nullStr(in.TemplateVersion), effective, actorOrNil(ctx)).
		Scan(&version)
	if err != nil {
		if isUniqueViolation(err) {
			return Artifact{}, kerr.E(kerr.KindConflict, "version_conflict", "concurrent artifact generation; retry")
		}
		return Artifact{}, kerr.Wrapf(err, "artifact.Generate", "insert artifact")
	}
	return Artifact{ID: id, Kind: in.Kind, Version: version, ContentHash: hash, ContentType: ct, Sidecar: sidecar}, nil
}

// Get returns an artifact including its content.
func (p *Pipeline) Get(ctx context.Context, db database.TenantDB, id uuid.UUID) (Artifact, error) {
	var a Artifact
	var sidecar []byte
	var tmpl *string
	if err := db.QueryRow(ctx,
		`SELECT id, kind, version, content_hash, content_type, content, sidecar, template_version
		   FROM artifacts WHERE id = $1`, id).
		Scan(&a.ID, &a.Kind, &a.Version, &a.ContentHash, &a.ContentType, &a.Content, &sidecar, &tmpl); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Artifact{}, kerr.E(kerr.KindNotFound, "not_found", "no such artifact")
		}
		return Artifact{}, kerr.Wrapf(err, "artifact.Get", "read artifact")
	}
	_ = json.Unmarshal(sidecar, &a.Sidecar)
	if tmpl != nil {
		a.TemplateVersion = *tmpl
	}
	return a, nil
}

// Verify re-hashes the stored content and reports whether it still matches the
// recorded content_hash — detecting any out-of-band mutation of an artifact
// (app_rt cannot UPDATE/DELETE, but this catches a DBA/owner tamper).
func (p *Pipeline) Verify(ctx context.Context, db database.TenantDB, id uuid.UUID) (bool, error) {
	a, err := p.Get(ctx, db, id)
	if err != nil {
		return false, err
	}
	sum := sha256.Sum256(a.Content)
	return hex.EncodeToString(sum[:]) == a.ContentHash, nil
}

// List returns a kind's artifacts newest-version first (metadata only, no content).
func (p *Pipeline) List(ctx context.Context, db database.TenantDB, kind string) ([]Artifact, error) {
	rows, err := db.Query(ctx,
		`SELECT id, kind, version, content_hash, content_type, template_version
		   FROM artifacts WHERE kind = $1 ORDER BY version DESC`, kind)
	if err != nil {
		return nil, kerr.Wrapf(err, "artifact.List", "query artifacts")
	}
	defer rows.Close()
	var out []Artifact
	for rows.Next() {
		var a Artifact
		var tmpl *string
		if err := rows.Scan(&a.ID, &a.Kind, &a.Version, &a.ContentHash, &a.ContentType, &tmpl); err != nil {
			return nil, kerr.Wrapf(err, "artifact.List", "scan artifact")
		}
		if tmpl != nil {
			a.TemplateVersion = *tmpl
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func actorOrNil(ctx context.Context) any {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
