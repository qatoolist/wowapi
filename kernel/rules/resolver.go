package rules

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// Resolved is a rule value plus provenance.
type Resolved struct {
	Key       string
	Value     json.RawMessage
	Scope     ScopeKind // scope the winning version was set at (or "" for the code default)
	VersionID uuid.UUID // zero when the code default won
	IsDefault bool
}

// Decode unmarshals the resolved value into out.
func (r Resolved) Decode(out any) error {
	if err := json.Unmarshal(r.Value, out); err != nil {
		return kerr.E(kerr.KindInternal, "invalid_rule_value", "rule value does not match the target type")
	}
	return nil
}

// OrgAncestry resolves an org's ancestor chain (self-first) so the resolver can
// walk org scope upward. Implemented against the DB by the caller-provided func
// so kernel/rules need not import the org store.
type OrgAncestry func(ctx context.Context, db database.TenantDB, orgID uuid.UUID) ([]uuid.UUID, error)

// Resolver resolves rule values. It runs on the caller's TenantDB (one
// snapshot), reads active versions, and falls back to the registered default.
type Resolver struct {
	reg      *Registry
	ancestry OrgAncestry
}

// NewResolver builds a resolver over the rule registry. ancestry may be nil
// (org-scope resolution then falls back to tenant/platform/default).
func NewResolver(reg *Registry, ancestry OrgAncestry) *Resolver {
	return &Resolver{reg: reg, ancestry: ancestry}
}

// Resolve returns the effective value of key for (tenant, org, at):
// the most specific active version wins — org-ancestry (nearest first) → tenant
// → platform → code default. Versions are immutable, so any historical `at`
// resolves deterministically (blueprint 02 §2.2). An unregistered key is a
// programming error.
// Resolve validates the winning value against the point's CURRENT
// value_schema before returning it (B3 defect 4): the value was validated
// against whatever schema was live at Propose time, but a point's schema can
// be tightened later (module upgrade) — a stored value that conformed to an
// earlier, looser schema can drift out of conformance with the schema the
// point is registered under now. Re-checking here is cheap (pure in-memory,
// no extra I/O — the point's schema is already loaded from the registry) and
// turns silent schema drift into a loud KindInternal error naming the rule
// key, rather than handing a caller a value that violates the very contract
// it is being served under. The code default itself is never re-checked here
// — it was already validated against this same schema at Register (B3 defect
// 3), so it is trusted by construction.
func (r *Resolver) Resolve(ctx context.Context, db database.TenantDB, key string, org uuid.UUID, at time.Time) (Resolved, error) {
	point, ok := r.reg.Get(key)
	if !ok {
		return Resolved{}, kerr.E(kerr.KindInternal, "unregistered_rule",
			"resolved an unregistered rule point: "+key)
	}

	var ancestors []uuid.UUID
	if org != uuid.Nil && r.ancestry != nil {
		var err error
		ancestors, err = r.ancestry(ctx, db, org)
		if err != nil {
			return Resolved{}, kerr.Wrapf(err, "rules.Resolve", "org ancestry")
		}
	}

	var (
		id    uuid.UUID
		val   []byte
		scope string
	)
	err := db.QueryRow(ctx, resolveVersionSQL, key, ancestors, at).Scan(&id, &val, &scope)
	if err != nil {
		if isNoRows(err) {
			// Code default: already validated against this schema at Register.
			return Resolved{Key: key, Value: point.Default, IsDefault: true}, nil
		}
		return Resolved{}, kerr.Wrapf(err, "rules.Resolve", "lookup %s", key)
	}
	return validateResolved(point, Resolved{
		Key: key, Value: json.RawMessage(val), Scope: ScopeKind(scope), VersionID: id,
	})
}

// resolveVersionSQL evaluates every eligible org, tenant, and platform version
// in one statement. WITH ORDINALITY preserves the ancestry callback's
// nearest-first order; tenant and platform are assigned priorities after every
// org. The final effective_from ordering retains the historical lookup rule
// within one scope. RLS continues to constrain tenant/platform visibility.
const resolveVersionSQL = `
WITH ancestry(scope_id, precedence) AS (
    SELECT scope_id, precedence
    FROM unnest($2::uuid[]) WITH ORDINALITY AS a(scope_id, precedence)
),
candidates AS (
    SELECT org_version.id, org_version.value, org_version.scope_kind,
           org_version.effective_from, a.precedence
    FROM ancestry a
    CROSS JOIN LATERAL (
        SELECT rv.id, rv.value, rv.scope_kind, rv.effective_from
        FROM rule_versions rv
        WHERE rv.rule_key = $1
          AND rv.scope_kind = 'org'
          AND rv.scope_id = a.scope_id
          AND rv.status IN ('active','superseded')
          AND rv.effective_from <= $3
          AND (rv.effective_to IS NULL OR rv.effective_to > $3)
        ORDER BY rv.effective_from DESC
        LIMIT 1
    ) org_version

    UNION ALL

    SELECT tenant_version.id, tenant_version.value, tenant_version.scope_kind,
           tenant_version.effective_from,
           COALESCE(cardinality($2::uuid[]), 0) + 1
    FROM LATERAL (
        SELECT rv.id, rv.value, rv.scope_kind, rv.effective_from
        FROM rule_versions rv
        WHERE rv.rule_key = $1
          AND rv.scope_kind = 'tenant'
          AND rv.scope_id IS NULL
          AND rv.status IN ('active','superseded')
          AND rv.effective_from <= $3
          AND (rv.effective_to IS NULL OR rv.effective_to > $3)
        ORDER BY rv.effective_from DESC
        LIMIT 1
    ) tenant_version

    UNION ALL

    SELECT platform_version.id, platform_version.value, platform_version.scope_kind,
           platform_version.effective_from,
           COALESCE(cardinality($2::uuid[]), 0) + 2
    FROM LATERAL (
        SELECT rv.id, rv.value, rv.scope_kind, rv.effective_from
        FROM rule_versions rv
        WHERE rv.rule_key = $1
          AND rv.scope_kind = 'platform'
          AND rv.scope_id IS NULL
          AND rv.status IN ('active','superseded')
          AND rv.effective_from <= $3
          AND (rv.effective_to IS NULL OR rv.effective_to > $3)
        ORDER BY rv.effective_from DESC
        LIMIT 1
    ) platform_version
)
SELECT id, value, scope_kind
FROM candidates
ORDER BY precedence, effective_from DESC
LIMIT 1`

// validateResolved re-checks a resolved (non-default) value against the
// point's current schema, surfacing schema drift as a loud KindInternal
// error instead of returning a value that no longer conforms.
func validateResolved(point Point, res Resolved) (Resolved, error) {
	if err := validateAgainstSchema(point.ValueSchema, res.Value); err != nil {
		return Resolved{}, kerr.E(kerr.KindInternal, "rule_schema_drift",
			"stored rule value for "+point.Key+" no longer conforms to its current value_schema: "+err.Error())
	}
	return res, nil
}
