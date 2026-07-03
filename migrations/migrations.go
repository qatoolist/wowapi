// Package migrations embeds the wowapi kernel SQL migrations and exposes them
// as an fs.FS for the migration runner (kernel/database.Migrate) and the
// wowapi CLI.
//
// Kernel migrations always run before product-module migrations. Each source
// (the kernel, then each module) is applied with its OWN goose history table
// keyed by a source name — database.Migrate(ctx, pool, src, source). This is
// what lets independently-numbered sources coexist: a module numbering its
// files 0001.. does not collide with the kernel's 00001.. because their
// version histories are separate tables (see docs/blueprint/03 §5 and
// docs/blueprint/11 §4). The ordering contract is: apply the "wowapi" source
// first, then each module's source in dependency-graph order.
//
// # File naming
//
// goose requires numeric-ascending file names. The blueprint's "000/001"
// logical names map to on-disk "00001/00002":
//
//	Blueprint 000 → 00001_bootstrap.sql    (extensions, roles, app_tenant_id)
//	Blueprint 001 → 00002_core_identity.sql (tenants, users, user_tenant_access)
//
// Phase 2 ships only these two migrations (see decision D-0025).
package migrations

import (
	"embed"
	"io/fs"
)

// SourceName is the kernel's migration source identifier — the history-table
// key passed to database.Migrate for the framework's own migrations.
const SourceName = "wowapi"

//go:embed *.sql
var kernelFS embed.FS

// Kernel returns the framework's migrations as an fs.FS for the migration
// runner (database.Migrate) and the wowapi CLI.
func Kernel() fs.FS { return kernelFS }
