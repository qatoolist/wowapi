// Package privileged is wowapi's scoped privileged-service surface: the
// sanctioned, audited way a module performs a valid tenant-scoped operation that
// requires PLATFORM privilege at the database, WITHOUT the module writing its
// own SECURITY DEFINER SQL and WITHOUT ever seeing a platform pool or raw SQL
// door (SEC-24 / SEC-13; GAP-006).
//
// # Why this exists
//
// Two framework tables are deliberately off-limits to the shared app_rt role
// modules run as:
//
//   - relationships — a granted_via edge is an AUTHORIZATION INPUT, so app_rt
//     holds SELECT only; writes are app_platform (migration 00005).
//   - rule_versions — ACTIVATION changes runtime behavior, so app_rt holds
//     SELECT,INSERT (propose drafts) only; activation UPDATE is app_platform
//     (migration 00008).
//
// A product that needs to grant an edge or activate a tenant rule version could
// previously only bridge the gap with a per-product SECURITY DEFINER function,
// re-implementing tenant binding, resource existence, type/key ownership, scope
// restriction, audit, and race handling every time — risky and unaudited.
//
// # How it stays safe
//
// Each Services value is bound to ONE owning module at construction. Every
// operation runs in a PLATFORM transaction that is nonetheless TENANT-BOUND
// (TxManager.WithTenant over the app_platform pool): app_tenant_id() resolves to
// the caller's tenant, so the relationships/rule_versions RLS WITH CHECK holds
// exactly as it did for the SECURITY DEFINER bridges, while the platform grants
// permit the write. In Go, before the write, the service enforces:
//
//   - tenant binding (caller ctx must carry a tenant; else fail closed);
//   - relationship-type / rule-key OWNERSHIP — the key must be prefixed with the
//     owning module name, or appear in a declared allow-list (mirrors how seeds,
//     resource types, and rule points validate key ownership);
//   - subject/object RESOURCE EXISTENCE in the bound tenant;
//   - SCOPE restriction (rule versions must be tenant-scope and belong to the
//     bound tenant);
//   - AUDIT metadata via the kernel audit hash chain, in the same tx;
//
// and it relies on the existing DB invariants — RLS tenant isolation, the
// rule_versions one-active-per-instant EXCLUDE constraint, row locks — for the
// concurrency guarantees the bridges depended on. No new GRANT is added to any
// table; the security posture of migration 00005/00008 is untouched.
package privileged

import (
	"context"
	"strings"

	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/rules"
)

// Services is the per-module bundle of scoped privileged services handed to a
// module through module.Context. It is bound to a single owning module (name)
// and enforces that module's ownership of every key it operates on.
type Services struct {
	module string
	rels   *Relationships
	rules  *Rules
}

// Config declares a module's extra (non-prefixed) ownership grants. A module
// always owns keys prefixed "<module>."; AllowRelTypes / AllowRuleKeys widen
// that set with explicit keys it is permitted to operate on (e.g. a kernel
// "core." relationship type a module is sanctioned to grant). Empty is the
// common case: prefix ownership only.
type Config struct {
	AllowRelTypes []string
	AllowRuleKeys []string
}

// New builds the privileged services for one module over the PLATFORM
// transaction manager (the app_platform pool). platformTx MUST be a tenant-bindable
// manager whose WithTenant runs as app_platform — the role that holds the
// relationships/rule_versions write grants; passing the app_rt manager would fail
// closed at the DB. audit and idgen are the shared kernel instances.
func New(module string, platformTx database.TxManager, store *rules.Store, audit *kaudit.Writer, idgen model.IDGen, cfg Config) *Services {
	own := &ownership{
		module:   module,
		relTypes: sliceToSet(cfg.AllowRelTypes),
		ruleKeys: sliceToSet(cfg.AllowRuleKeys),
	}
	return &Services{
		module: module,
		rels:   &Relationships{tx: platformTx, audit: audit, idgen: idgen, own: own},
		rules:  &Rules{tx: platformTx, store: store, audit: audit, own: own},
	}
}

// Relationships returns the ReBAC relationship-edge service (Grant / Revoke).
func (s *Services) Relationships() *Relationships { return s.rels }

// Rules returns the tenant-scope rule-version activation service.
func (s *Services) Rules() *Rules { return s.rules }

// ownership decides whether the bound module may operate on a given
// relationship type or rule key. A key is owned when it is prefixed with the
// module name ("<module>.…") or explicitly allow-listed. This mirrors the key
// ownership rule enforced by kernel/seeds, kernel/resource, and kernel/rules.
type ownership struct {
	module   string
	relTypes map[string]struct{}
	ruleKeys map[string]struct{}
}

func (o *ownership) ownsRelType(key string) bool { return o.owns(key, o.relTypes) }
func (o *ownership) ownsRuleKey(key string) bool { return o.owns(key, o.ruleKeys) }

func (o *ownership) owns(key string, allow map[string]struct{}) bool {
	if strings.HasPrefix(key, o.module+".") {
		return true
	}
	_, ok := allow[key]
	return ok
}

// denyRelType / denyRuleKey build the standard ownership-denied error. Modeled
// KindForbidden so it maps to 403 and can never be confused with a NotFound
// probe result.
func (o *ownership) denyRelType(key string) error {
	return kerr.E(kerr.KindForbidden, "ownership_denied",
		"module "+o.module+" may not manage relationship type "+key)
}

func (o *ownership) denyRuleKey(key string) error {
	return kerr.E(kerr.KindForbidden, "ownership_denied",
		"module "+o.module+" may not activate rule key "+key)
}

func sliceToSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}

// requireTenant fails closed when no tenant is bound in ctx. Every privileged
// operation is tenant-scoped, so this is checked before any DB access — matching
// the bridges' app_tenant_id() "raise if unbound" fail-closed posture, but in Go
// so the caller gets a clean error rather than a raw SQL exception.
func requireTenant(ctx context.Context) error {
	if _, ok := database.TenantIDFrom(ctx); !ok {
		return kerr.E(kerr.KindUnauthenticated, "no_tenant",
			"privileged service requires a tenant-bound context")
	}
	return nil
}
