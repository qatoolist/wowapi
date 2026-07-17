package rules

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
)

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

// Proposal is a requested rule value change at a scope.
type Proposal struct {
	Key           string
	Scope         ScopeKind
	ScopeID       uuid.UUID // org id for org scope; zero otherwise
	Value         json.RawMessage
	EffectiveFrom time.Time // zero → now
}

// Store persists rule versions. Writes run on the caller's TenantDB (draft
// proposals as app_rt) or a platform connection (activation as app_platform).
type Store struct {
	reg   *Registry
	idgen model.IDGen
}

// NewStore builds the version store over the rule registry.
func NewStore(reg *Registry, idgen model.IDGen) *Store { return &Store{reg: reg, idgen: idgen} }

// Propose inserts a DRAFT rule version in the caller's tenant tx (app_rt may
// INSERT). A draft never resolves — it must be Activate'd (a platform/kernel
// operation, app_platform) to take effect. This keeps rule ACTIVATION — which
// changes runtime behavior — off the module-facing app_rt role, consistent with
// the config-write posture (SEC-13). The RequiresApproval flag governs whether a
// human/workflow approval must precede Activate; the store mechanics are uniform.
func (s *Store) Propose(ctx context.Context, db database.TenantDB, p Proposal) (uuid.UUID, error) {
	point, ok := s.reg.Get(p.Key)
	if !ok {
		return uuid.Nil, kerr.E(kerr.KindInternal, "unregistered_rule", "proposed an unregistered rule point: "+p.Key)
	}
	if !point.allowsScope(p.Scope) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "rule_violation",
			"rule point does not allow the requested scope")
	}
	// Defense in depth: the value must satisfy the point's schema at WRITE time,
	// not just fail on the read-path Decode (review finding SEC-40).
	if err := validateAgainstSchema(point.ValueSchema, p.Value); err != nil {
		return uuid.Nil, err
	}
	from := p.EffectiveFrom
	if from.IsZero() {
		from = time.Now()
	}
	id := s.idgen.New()
	tenantExpr := "app_tenant_id()"
	if p.Scope == ScopePlatform {
		tenantExpr = "NULL"
	}
	// created_by is the proposing actor (audit integrity — ARCH-62); falls back
	// to nil UUID only when no actor is in context.
	createdBy := uuid.Nil
	if a, ok := database.ActorIDFrom(ctx); ok {
		createdBy = a
	}
	_, err := db.Exec(ctx,
		`INSERT INTO rule_versions
             (id, rule_key, tenant_id, scope_kind, scope_id, value, effective_from, status, created_by)
         VALUES ($1, $2, `+tenantExpr+`, $3, $4, $5, $6, 'draft', $7)`,
		id, p.Key, string(p.Scope), scopeIDArg(p), p.Value, from, createdBy)
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "rules.Propose", "insert draft version for %s", p.Key)
	}
	return id, nil
}

// Activate approves a draft version: it supersedes any active version at the
// same scope and marks the draft active, recording the approver — all in one
// tx. Runs with platform privilege (rule activation is a kernel/platform
// concern). Returns an error if the version is not in draft/pending.
func (s *Store) Activate(ctx context.Context, db database.DBTX, versionID, approvedBy uuid.UUID) error {
	// Load the draft to know its scope for supersede.
	var (
		key      string
		scope    string
		scopeID  *uuid.UUID
		from     time.Time
		status   string
		tenantID *uuid.UUID
	)
	err := db.QueryRow(ctx,
		`SELECT rule_key, scope_kind, scope_id, effective_from, status, tenant_id
           FROM rule_versions WHERE id = $1`, versionID).Scan(&key, &scope, &scopeID, &from, &status, &tenantID)
	if isNoRows(err) {
		return kerr.E(kerr.KindNotFound, "not_found", "rule version not found")
	}
	if err != nil {
		return kerr.Wrapf(err, "rules.Activate", "load version")
	}
	if status != "draft" && status != "pending_approval" {
		return kerr.E(kerr.KindConflict, "invalid_transition", "only a draft/pending version can be activated")
	}
	// Supersede any currently-active version at the same (key, scope, scope_id, tenant).
	if _, err := db.Exec(ctx,
		`UPDATE rule_versions SET status = 'superseded', effective_to = $5
          WHERE rule_key = $1 AND scope_kind = $2
            AND COALESCE(scope_id,'00000000-0000-0000-0000-000000000000'::uuid) = COALESCE($3::uuid,'00000000-0000-0000-0000-000000000000'::uuid)
            AND COALESCE(tenant_id,'00000000-0000-0000-0000-000000000000'::uuid) = COALESCE($4::uuid,'00000000-0000-0000-0000-000000000000'::uuid)
            AND status = 'active'`,
		key, scope, scopeID, tenantID, from); err != nil {
		return kerr.Wrapf(err, "rules.Activate", "supersede active")
	}
	if _, err := db.Exec(ctx,
		`UPDATE rule_versions SET status = 'active', approved_by = $2 WHERE id = $1`,
		versionID, approvedBy); err != nil {
		return kerr.Wrapf(err, "rules.Activate", "activate")
	}
	return nil
}

func scopeIDArg(p Proposal) any {
	if p.ScopeID != uuid.Nil {
		return p.ScopeID
	}
	return nil
}
