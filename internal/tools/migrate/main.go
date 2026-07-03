// Command migrate applies the kernel migrations to the database named by
// DATABASE_URL — the framework repo's local/CI migration runner behind
// `make migrate`. Product repositories get a real cmd/migrate process via
// app.RunMigrate (Phase 5+); this private tool exists so the compose stack
// can be migrated without one.
//
// SECURITY: this tool intentionally connects as the raw DATABASE_URL login
// (the local compose superuser) and uses config.Defaults() — acceptable ONLY
// because it is a dev/CI convenience. Product code (app.RunMigrate) MUST use
// the dedicated app_migrate owner DSN from the narrowed MigrateConfig and MUST
// NOT copy this DATABASE_URL/Defaults() shortcut (SEC-15).
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/migrations"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "migrate: DATABASE_URL is not set (run `make up`, or use `make migrate` which supplies the compose default)")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.NewPool(ctx, dsn, config.Defaults().DB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	res, err := database.Migrate(ctx, pool, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("kernel migrations applied (%d this run); %s source at version %d\n",
		res.Applied, migrations.SourceName, res.Version)
}
