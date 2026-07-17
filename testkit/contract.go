package testkit

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/seeds"
	"github.com/qatoolist/wowapi/v2/module"
)

// RunModuleContract is the kernel's module conformance suite (blueprint 08 §2,
// 11): it registers the module ALONE on a fresh kernel and asserts it
//   - boots and validates (routes have metadata, permissions are declared,
//     no dependency/registry errors) on an EMPTY config namespace — defaults
//     must be complete;
//   - migrates and seeds IDEMPOTENTLY (running each twice is a no-op);
//   - enforces RLS on every module-owned table;
//   - REJECTS an invalid config namespace (unknown key) at boot.
//
// It requires a real Postgres (skips without a DSN, like NewDB).
func RunModuleContract(t *testing.T, m module.Module) {
	t.Helper()
	h := NewDB(t)
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(discard{}, nil))

	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	if err != nil {
		t.Fatalf("contract: kernel.New: %v", err)
	}

	// 1. Boots + validates on an empty namespace (defaults complete).
	a := app.New()
	a.Register(m)
	booted, err := a.Boot(ctx, k, nil)
	if err != nil {
		t.Fatalf("contract: module %q must boot on an empty config namespace: %v", m.Name(), err)
	}

	// 2. Migrations apply, and re-applying is a no-op (idempotent). Snapshot
	// the public tables before/after so the RLS check (step 4) inspects the
	// tables THIS module created — not a name-prefix convention it could evade.
	before := publicTables(t, h)
	migFS, ok := booted.RuntimeMigrations()[m.Name()]
	if ok {
		r1, err := database.Migrate(ctx, h.Admin, migFS, m.Name())
		if err != nil {
			t.Fatalf("contract: module migrate: %v", err)
		}
		if r1.Applied == 0 {
			t.Fatalf("contract: module %q declared migrations but none applied", m.Name())
		}
		r2, err := database.Migrate(ctx, h.Admin, migFS, m.Name())
		if err != nil {
			t.Fatalf("contract: module migrate rerun: %v", err)
		}
		if r2.Applied != 0 {
			t.Fatalf("contract: module migrations are not idempotent (rerun applied %d)", r2.Applied)
		}
	}
	created := diffTables(before, publicTables(t, h))

	// 3. Seeds sync under app_platform privilege — the real posture (SEC-33), not
	// superuser — so a seed needing a grant app_platform lacks fails here. Then
	// re-sync and assert it changed NOTHING (idempotent in EFFECT, not just
	// no-error): the catalog checksum before and after the second sync matches
	// (review finding ARCH-49).
	if err := seeds.Sync(ctx, h.Platform, booted.RuntimeSeeds()); err != nil {
		t.Fatalf("contract: seed sync (as app_platform): %v", err)
	}
	sum1 := catalogChecksum(t, h)
	if err := seeds.Sync(ctx, h.Platform, booted.RuntimeSeeds()); err != nil {
		t.Fatalf("contract: seed sync rerun (must be idempotent): %v", err)
	}
	if sum2 := catalogChecksum(t, h); sum1 != sum2 {
		t.Fatalf("contract: seed sync is not idempotent in effect — the catalog changed on rerun")
	}

	// 4. RLS forced on every table the module's migration created — inspected
	// by before/after diff, so a non-conforming table name cannot evade the
	// check, and a module that declared migrations but produced no RLS-forced
	// table fails (review finding ARCH-48).
	assertTablesRLS(t, h, m.Name(), created, ok)

	// 5. Rejects an invalid config namespace (unknown key) at boot.
	bad := config.Namespaces{m.Name(): config.MapView{"this_key_does_not_exist_in_the_module_config": true}}
	k2, _ := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	a2 := app.New()
	a2.Register(m)
	if _, err := a2.Boot(ctx, k2, bad); err == nil {
		t.Fatalf("contract: module %q must reject an unknown config key at boot", m.Name())
	}
}

// publicTables returns the set of public schema table names.
func publicTables(t *testing.T, h *DBHandle) map[string]bool {
	t.Helper()
	rows, err := h.Admin.Query(context.Background(),
		`SELECT tablename FROM pg_tables WHERE schemaname = 'public'`)
	if err != nil {
		t.Fatalf("contract: list tables: %v", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		out[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func diffTables(before, after map[string]bool) []string {
	var created []string
	for name := range after {
		if before[name] {
			continue
		}
		// goose's own version-history table is bookkeeping, not tenant data.
		if strings.HasPrefix(name, "goose_") {
			continue
		}
		created = append(created, name)
	}
	sort.Strings(created)
	return created
}

// assertTablesRLS asserts every table the module created has FORCE row-level
// security. A module that declared migrations but created no RLS-forced table
// is a contract failure — the RLS invariant is not optional.
func assertTablesRLS(t *testing.T, h *DBHandle, moduleName string, created []string, declaredMigrations bool) {
	t.Helper()
	if len(created) == 0 {
		if declaredMigrations {
			t.Fatalf("contract: module %q ran migrations but created no tables to RLS-check", moduleName)
		}
		return
	}
	for _, name := range created {
		var rls, force bool
		if err := h.Admin.QueryRow(context.Background(),
			`SELECT c.relrowsecurity, c.relforcerowsecurity
               FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
              WHERE n.nspname = 'public' AND c.relname = $1`, name).Scan(&rls, &force); err != nil {
			t.Fatalf("contract: introspect RLS for %q: %v", name, err)
		}
		if !rls || !force {
			t.Errorf("contract: table %q created by module %q must ENABLE and FORCE row-level security (enabled=%v forced=%v)",
				name, moduleName, rls, force)
		}
	}
}

// catalogChecksum returns a stable hash of the seed-managed catalogs, so a
// re-sync that rewrites any row (not merely succeeds) is detectable.
func catalogChecksum(t *testing.T, h *DBHandle) string {
	t.Helper()
	var sum string
	err := h.Admin.QueryRow(context.Background(), `
SELECT md5(string_agg(x, '|' ORDER BY x)) FROM (
    SELECT key||':'||coalesce(description,'')||':'||sensitive::text AS x FROM permissions
    UNION ALL SELECT 'rt:'||key||':'||coalesce(description,'') FROM resource_types
    UNION ALL SELECT 'rel:'||key||':'||subject_kind||':'||object_kind FROM relationship_types
    UNION ALL SELECT 'rp:'||rp.role_id::text||':'||rp.permission_key
              FROM role_permissions rp JOIN roles r ON r.id = rp.role_id WHERE r.tenant_id IS NULL
) s`).Scan(&sum)
	if err != nil {
		t.Fatalf("contract: catalog checksum: %v", err)
	}
	return sum
}

// discard is an io.Writer sink for the contract's throwaway logger.
type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
