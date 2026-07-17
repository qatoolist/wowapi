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
// The clean v1.2.0 line starts at 00001_baseline.sql. Future kernel migrations
// resume at 00002 and use the generic online-migration manifest machinery when
// their rollout requires it; abandoned pre-v1.2 migration history is not an
// upgrade source.
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
