package rules

import (
	"context"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// SyncDefinitions upserts every point in the registry into rule_definitions —
// the persisted mirror blueprint 02 §2.1 describes ("makes points
// introspectable/auditable in the DB"), and the FK rule_versions.rule_key
// depends on. It is the rule-registry analogue of kernel/seeds.Sync: it must
// run on a platform-privileged connection (rule_definitions is app_platform
// SELECT/INSERT/UPDATE, app_rt SELECT-only — migration 00008), and it is
// idempotent — re-running converges the schema/default/scopes/approval/
// description columns onto whatever the Go registry currently declares, never
// producing duplicate rows (ON CONFLICT (key) DO UPDATE).
//
// Call this from the generated migrate main after module migrations (so the
// table exists) and before any rule_versions writes — mirroring seeds.Sync's
// lifecycle position exactly (GAP-003 → GAP-007). A standalone `wowapi rules
// sync` also runs it outside a full migrate, for re-syncing definitions
// without a schema change.
//
// SyncDefinitions does not itself check module ownership beyond what
// Registry.Register already enforced at registration time (a rule point can
// only be registered under its own module prefix); a registry that failed
// Err() should never be passed here.
func SyncDefinitions(ctx context.Context, db database.DBTX, reg *Registry) error {
	for _, key := range reg.Keys() {
		p, ok := reg.Get(key)
		if !ok {
			continue // defensive: Keys() is derived from the same map
		}
		scopes := make([]string, len(p.AllowedScopes))
		for i, s := range p.AllowedScopes {
			scopes[i] = string(s)
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO rule_definitions (key, module, value_schema, default_value, allowed_scopes, requires_approval, description)
                  VALUES ($1, $2, $3, $4, $5, $6, $7)
             ON CONFLICT (key) DO UPDATE SET
                   module = EXCLUDED.module,
                   value_schema = EXCLUDED.value_schema,
                   default_value = EXCLUDED.default_value,
                   allowed_scopes = EXCLUDED.allowed_scopes,
                   requires_approval = EXCLUDED.requires_approval,
                   description = EXCLUDED.description`,
			p.Key, p.Module, p.ValueSchema, p.Default, scopes, p.RequiresApproval, p.Description); err != nil {
			return kerr.Wrapf(err, "rules.SyncDefinitions", "upsert rule_definitions %s", p.Key)
		}
	}
	return nil
}
