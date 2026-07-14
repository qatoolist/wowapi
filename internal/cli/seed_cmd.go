// seed_cmd.go — `wowapi seed validate`/`wowapi seed sync` (Phase 10, GAP-003;
// escape-hatch framing per B4). validate loads a module's seed bundle through
// the same kernel/seeds.Load the app uses at boot, so a seed error is caught
// in CI (exit 1) rather than at deploy time. sync loads one or more modules'
// seed bundles and applies them to a real database with kernel/seeds.Sync, on
// a platform-privileged connection — but it is a LOW-LEVEL STANDALONE ESCAPE
// HATCH, not the production lifecycle path: the generated cmd/migrate is
// (loads the composed product config via appcfg.Load(), runs migrations, then
// seeds.Sync, then rules.SyncDefinitions — see
// internal/cli/templates/init/cmd_migrate_main.go.tmpl and
// docs/user-guide/database-migrations.md). This command connects via a bare
// DATABASE_URL env var (no product config layering, no secretref://
// resolution, hardcoded pool defaults) and never calls rules.SyncDefinitions:
// rule points exist only as Go declarations registered by a booted product
// process (mc.Rules().Register(...)), so a framework-only binary has no
// registry to read — there is nothing for this command to sync even in
// principle (see kernel/rules.SyncDefinitions's doc comment). Use it to
// re-sync seed catalogs without a full migrate run; it is not a substitute
// for the generated migrate on a fresh environment.
package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/seeds"
)

func seedUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi seed <validate|sync> [flags]

Subcommands:
  validate   load and validate a module's seed bundle (no database needed)
  sync       load one or more modules' seed bundles and apply them to a
             database (LOW-LEVEL ESCAPE HATCH — see below)

Flags (validate):
  --dir      directory holding the module's seed YAML (default "seeds")
  --module   module name that owns these seeds (required; keys must be prefixed)

Flags (sync):
  --module   name=dir pair identifying a module's seed directory; repeatable,
             e.g. --module widgets=modules/widgets/seeds (at least one required)
  --dry-run  compute and print a change plan without writing to the database

THE PRODUCTION PATH IS THE GENERATED "cmd/migrate": it loads the composed
product config via appcfg.Load() (configs/base.yaml + configs/<env>.yaml +
secretref:// resolution), then runs migrations, seeds.Sync, AND
rules.SyncDefinitions, in that order — see
docs/user-guide/database-migrations.md#seeds and #rule-definitions.

'wowapi seed sync' is a low-level escape hatch (e.g. to re-sync catalogs
without a full migrate run) with real limitations vs. that path:
  - connects to a bare DATABASE_URL env var as app_platform — no product
    config layering, no secretref:// resolution, hardcoded pool defaults
    (config.Defaults().DB), unlike appcfg.Load().
  - does NOT sync rule definitions (rule_definitions / GAP-007). Rule points
    exist only as Go declarations registered by a booted product process
    (mc.Rules().Register(...)); this framework-only binary has no product
    rule registry to read, so there is nothing to sync here even in
    principle. Run the generated migrate (or call
    rules.SyncDefinitions(ctx, pool, registry) from a custom migrate main)
    to keep rule_definitions converged — otherwise any rule_versions.Propose
    for a new key fails its rule_key foreign key.

sync connects to DATABASE_URL as app_platform (the kernel maintenance role,
same convention as 'wowapi dlq') and upserts the merged seed bundle
idempotently.
`)
}

func runSeed(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		seedUsage(stderr)
		return 2
	}
	switch args[0] {
	case "validate":
		return runSeedValidate(args[1:], stdout, stderr)
	case "sync":
		return runSeedSync(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		seedUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "wowapi seed: unknown subcommand %q\n", args[0])
		seedUsage(stderr)
		return 2
	}
}

func runSeedValidate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi seed validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "seeds", "directory holding the module's seed YAML")
	module := fs.String("module", "", "module name that owns these seeds")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *module == "" {
		fmt.Fprintln(stderr, "wowapi seed validate: --module is required")
		return 2
	}
	if info, err := os.Stat(*dir); err != nil || !info.IsDir() {
		fmt.Fprintf(stderr, "wowapi seed validate: %s is not a directory\n", *dir)
		return 1
	}
	bundle, err := seeds.Load(os.DirFS(*dir), *module)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi seed validate: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "OK: %s seeds valid — %d permissions, %d roles, %d resource types, %d relationship types\n",
		*module, len(bundle.Permissions), len(bundle.Roles), len(bundle.ResourceTypes), len(bundle.RelationshipTypes))
	return 0
}

// moduleSeedDir is one --module name=dir pair for `wowapi seed sync`.
type moduleSeedDir struct {
	name, dir string
}

// moduleFlagList accumulates repeated --module name=dir flags via flag.Func.
type moduleFlagList struct{ entries []moduleSeedDir }

func (l *moduleFlagList) set(v string) error {
	name, dir, ok := strings.Cut(v, "=")
	if !ok || name == "" || dir == "" {
		return fmt.Errorf("must be name=dir (got %q)", v)
	}
	l.entries = append(l.entries, moduleSeedDir{name: name, dir: dir})
	return nil
}

func runSeedSync(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi seed sync", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var modules moduleFlagList
	var dryRun bool
	fs.Func("module", "name=dir pair identifying a module's seed directory (repeatable)", modules.set)
	fs.BoolVar(&dryRun, "dry-run", false, "compute and print a change plan without writing to the database")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(modules.entries) == 0 {
		fmt.Fprintln(stderr, "wowapi seed sync: at least one --module name=dir is required")
		return 2
	}

	// Escape-hatch disclosure (B4): print unconditionally, before any work,
	// so a user who never reads --help still sees it. This command is not the
	// production lifecycle path — it has no product config/secretref
	// resolution and, unlike the generated migrate, rule definitions are NOT
	// synced by this command (no product rule registry to read from a
	// framework-only binary). See docs/user-guide/database-migrations.md.
	fmt.Fprintln(stderr, "wowapi seed sync: WARNING: low-level escape hatch — rule definitions are NOT synced "+
		"by this command (no product rule registry available here); run the generated 'migrate' for the full "+
		"production lifecycle (migrations -> seeds.Sync -> rules.SyncDefinitions).")

	// Load and merge every module's bundle BEFORE touching the database, so a
	// seed error (typo, ownership violation) is reported without a partial
	// sync — mirrors validate's strict-load behavior.
	var bundle seeds.Bundle
	for _, m := range modules.entries {
		info, err := os.Stat(m.dir)
		if err != nil || !info.IsDir() {
			fmt.Fprintf(stderr, "wowapi seed sync: %s is not a directory\n", m.dir)
			return 1
		}
		b, err := seeds.Load(os.DirFS(m.dir), m.name)
		if err != nil {
			fmt.Fprintf(stderr, "wowapi seed sync: %v\n", err)
			return 1
		}
		bundle.Permissions = append(bundle.Permissions, b.Permissions...)
		bundle.Roles = append(bundle.Roles, b.Roles...)
		bundle.ResourceTypes = append(bundle.ResourceTypes, b.ResourceTypes...)
		bundle.RelationshipTypes = append(bundle.RelationshipTypes, b.RelationshipTypes...)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(stderr, "wowapi seed sync: DATABASE_URL is not set")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Same connection convention as `wowapi dlq` (kernel/database.NewPool +
	// WithSetRole("app_platform") + WithConnRLSGuard): seeds.Apply writes the
	// global catalog tables (permissions, roles, role_permissions,
	// resource_types, relationship_types), which are app_platform-writable and
	// app_rt-read-only by design (SEC-13/D-0026) — never a superuser/BYPASSRLS
	// DSN. This CLI invocation is a one-shot process with no long-lived authz
	// cache to invalidate (unlike a running api/worker with AuthzCacheTTL set),
	// so no SpineInvalidator is passed; Apply behaves exactly as it does with
	// caching off.
	pool, err := database.NewPool(ctx, dsn, config.Defaults().DB,
		database.WithSetRole("app_platform"), database.WithConnRLSGuard())
	if err != nil {
		fmt.Fprintf(stderr, "wowapi seed sync: %v\n", err)
		return 1
	}
	defer pool.Close()

	report, err := seeds.Apply(ctx, pool, bundle, seeds.ApplyOptions{
		DryRun: dryRun,
		Actor:  "wowapi-cli",
		Out:    stdout,
	})
	if err != nil {
		fmt.Fprintf(stderr, "wowapi seed sync: %v\n", err)
		return 1
	}

	switch report.Outcome {
	case "dry_run":
		fmt.Fprintf(stdout, "OK: dry-run complete (manifest hash %s)\n", report.Hash)
	case "noop":
		fmt.Fprintf(stdout, "OK: seed catalogs already up to date (hash %s)\n", report.Hash)
	default:
		fmt.Fprintf(stdout, "OK: synced %d permissions, %d roles, %d resource types, %d relationship types across %d module(s) (hash %s)\n",
			len(bundle.Permissions), len(bundle.Roles), len(bundle.ResourceTypes), len(bundle.RelationshipTypes), len(modules.entries), report.Hash)
	}
	return 0
}
