package privileged

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/rules"
)

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

// Rules is the scoped privileged service for tenant-scope rule-version
// activation. It lets a module activate a draft version of a rule KEY it owns,
// but only a TENANT-SCOPE version belonging to the caller's bound tenant —
// platform-scope activation stays platform-tooling-only. It absorbs, framework-
// side, the checks the product SECURITY DEFINER bridge
// (policy_activate_rule_version) performed, and delegates the supersede+activate
// state machine to the kernel rules.Store so the one-active-per-instant EXCLUDE
// constraint keeps arbitrating races.
type Rules struct {
	tx    database.TxManager // the app_platform, tenant-bindable manager
	store *rules.Store
	audit *kaudit.Writer
	own   *ownership
}

// ActivateOptions carries optional, product-supplied activation gates. Gate, when
// set, runs INSIDE the activation transaction after the framework's ownership /
// scope / tenant checks pass but before the supersede+activate write — so a
// product can enforce a domain rule (e.g. "a verified citation must cover the
// effective date") atomically, without a SECURITY DEFINER bridge. Returning an
// error aborts and rolls back the activation. The framework stays domain-
// agnostic: it never interprets the gate, only runs it in the right tx position.
type ActivateOptions struct {
	Gate func(ctx context.Context, db database.TenantDB) error
}

// ActivateTenant activates a tenant-scope rule version the module owns, in a
// tenant-bound app_platform transaction, and writes an audit row. It enforces
// (in order): a bound tenant; that the version exists; module ownership of its
// rule_key; that the version is TENANT scope AND belongs to the bound tenant
// (bridge check policy_activation_scope_denied — cross-tenant / platform-scope
// activation is refused); the draft/pending transition (delegated to the store);
// an optional product Gate; then supersede+activate via rules.Store.Activate.
//
// Concurrency: two concurrent activations of overlapping versions at the same
// (key, tenant, scope) both call the store's supersede+activate; the rule_versions
// one-active-per-instant EXCLUDE constraint makes the loser fail with a conflict
// (23P01) rather than both becoming active — the same arbitration the bridge
// relied on.
func (r *Rules) ActivateTenant(ctx context.Context, versionID, approvedBy uuid.UUID, opts ActivateOptions) error {
	if err := requireTenant(ctx); err != nil {
		return err
	}
	if versionID == uuid.Nil {
		return kerr.E(kerr.KindValidation, "invalid_version", "rule version id is required")
	}
	boundTenant, _ := database.TenantIDFrom(ctx)

	return r.tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		// Lock + load the version to enforce ownership and scope before we let the
		// store mutate it. RLS on rule_versions admits this tenant's rows and
		// platform (NULL-tenant) rows; the explicit scope/tenant checks below then
		// refuse anything that is not a tenant-scope row of THIS tenant.
		// IMPORTANT: rule_versions carries an app_platform bypass RLS policy
		// (rule_versions_platform_all, USING/WITH CHECK true, migration 00008) so a
		// tenant-bound app_platform tx can STILL see and write every tenant's rows —
		// unlike relationships/resources/acting_capacities, whose tenant-isolation
		// policies bind app_platform too. The cross-tenant boundary for activation is
		// therefore enforced HERE, in Go, and must not be weakened. We load the row
		// unfiltered first to classify the failure precisely (not-found vs ownership
		// vs scope), then bind a DB-side tenant+scope guard onto the FOR UPDATE lock
		// as belt-and-suspenders so a future refactor cannot silently drop the check.
		var (
			ruleKey   string
			scopeKind string
			tenantID  *uuid.UUID
		)
		err := db.QueryRow(ctx,
			`SELECT rule_key, scope_kind, tenant_id FROM rule_versions WHERE id = $1`, versionID).
			Scan(&ruleKey, &scopeKind, &tenantID)
		if isNoRows(err) {
			return kerr.E(kerr.KindNotFound, "not_found", "rule version not found")
		}
		if err != nil {
			return kerr.Wrapf(err, "privileged.Rules.ActivateTenant", "load version")
		}
		if !r.own.ownsRuleKey(ruleKey) {
			return r.own.denyRuleKey(ruleKey)
		}
		// Scope restriction: tenant-scope only, and the row must belong to the bound
		// tenant. Masked as KindTenantIsolation (404) so a cross-tenant probe cannot
		// distinguish "exists elsewhere" from "does not exist".
		if scopeKind != string(rules.ScopeTenant) || tenantID == nil || *tenantID != boundTenant {
			return kerr.E(kerr.KindTenantIsolation, "scope_denied",
				"only a tenant-scope rule version of the caller's tenant can be activated here")
		}
		// DB backstop: take the row lock ONLY if it is genuinely a tenant-scope row of
		// the bound tenant. If the platform-bypass policy or the Go check above ever
		// regressed, a mismatched row yields no lock here and we fail closed rather
		// than activate cross-tenant.
		var locked uuid.UUID
		if err := db.QueryRow(ctx,
			`SELECT id FROM rule_versions
			  WHERE id = $1 AND scope_kind = 'tenant' AND tenant_id = $2 FOR UPDATE`,
			versionID, boundTenant).Scan(&locked); err != nil {
			if isNoRows(err) {
				return kerr.E(kerr.KindTenantIsolation, "scope_denied",
					"only a tenant-scope rule version of the caller's tenant can be activated here")
			}
			return kerr.Wrapf(err, "privileged.Rules.ActivateTenant", "lock version")
		}
		// Optional product gate, run atomically before the state transition.
		if opts.Gate != nil {
			if err := opts.Gate(ctx, db); err != nil {
				return err
			}
		}
		// Delegate the supersede + activate state machine to the kernel store; it
		// re-checks the draft/pending transition and the EXCLUDE constraint
		// arbitrates concurrent activations. TenantDB satisfies database.DBTX.
		if err := r.store.Activate(ctx, db, versionID, approvedBy); err != nil {
			return err
		}
		return r.audit.Record(ctx, db, kaudit.Entry{
			Action:     "rule_version.activate",
			EntityType: "rule_version",
			EntityID:   versionID,
			ActorKind:  "system",
			Metadata: map[string]any{
				"rule_key":    ruleKey,
				"approved_by": approvedBy.String(),
				"module":      r.own.module,
			},
		})
	})
}
