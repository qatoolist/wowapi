package rules

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
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
func (r *Resolver) Resolve(ctx context.Context, db database.TenantDB, key string, org uuid.UUID, at time.Time) (Resolved, error) {
	point, ok := r.reg.Get(key)
	if !ok {
		return Resolved{}, kerr.E(kerr.KindInternal, "unregistered_rule",
			"resolved an unregistered rule point: "+key)
	}

	// Org scope: walk ancestry nearest-first, taking the first org that has an
	// active version at `at`.
	if org != uuid.Nil && r.ancestry != nil {
		ancestors, err := r.ancestry(ctx, db, org)
		if err != nil {
			return Resolved{}, kerr.Wrapf(err, "rules.Resolve", "org ancestry")
		}
		for _, oid := range ancestors {
			if res, found, err := r.lookup(ctx, db, key, ScopeOrg, oid, at); err != nil {
				return Resolved{}, err
			} else if found {
				return res, nil
			}
		}
	}

	// Tenant scope.
	if res, found, err := r.lookup(ctx, db, key, ScopeTenant, uuid.Nil, at); err != nil {
		return Resolved{}, err
	} else if found {
		return res, nil
	}
	// Platform scope.
	if res, found, err := r.lookup(ctx, db, key, ScopePlatform, uuid.Nil, at); err != nil {
		return Resolved{}, err
	} else if found {
		return res, nil
	}
	// Code default.
	return Resolved{Key: key, Value: point.Default, IsDefault: true}, nil
}

// lookup finds the version of key that was IN EFFECT at `at` for a scope+id.
// It considers 'active' AND 'superseded' versions (a superseded version was the
// effective value during its [effective_from, effective_to) window — excluding
// it would make any historical `at` inside a closed window fall through to the
// default, review finding ARCH-60). Draft/pending/rejected are excluded so
// approval gating holds. The exclusion constraint guarantees one active per
// instant; the supersession chain partitions the past into non-overlapping
// windows, so ORDER BY effective_from DESC picks the one covering `at`.
// RLS scopes the read to the tenant (+ platform rows).
func (r *Resolver) lookup(ctx context.Context, db database.TenantDB, key string, scope ScopeKind, scopeID uuid.UUID, at time.Time) (Resolved, bool, error) {
	var (
		id  uuid.UUID
		val []byte
	)
	var scopeArg any
	if scopeID != uuid.Nil {
		scopeArg = scopeID
	}
	err := db.QueryRow(ctx,
		`SELECT id, value FROM rule_versions
          WHERE rule_key = $1 AND scope_kind = $2
            AND (scope_id = $3 OR ($3 IS NULL AND scope_id IS NULL))
            AND status IN ('active','superseded')
            AND effective_from <= $4 AND (effective_to IS NULL OR effective_to > $4)
          ORDER BY effective_from DESC
          LIMIT 1`,
		key, string(scope), scopeArg, at).Scan(&id, &val)
	if err != nil {
		if isNoRows(err) {
			return Resolved{}, false, nil
		}
		return Resolved{}, false, kerr.Wrapf(err, "rules.Resolve", "lookup %s at %s", key, scope)
	}
	return Resolved{Key: key, Value: json.RawMessage(val), Scope: scope, VersionID: id}, true, nil
}
