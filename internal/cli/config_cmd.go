// config_cmd.go — wowapi config subcommands (framework-side, Phase 1).
//
// Framework-repo fallback: all subcommands run against config.Framework alone.
// Product-repo configcheck integration (tools/configcheck) arrives in Phase 10.
//
// Precedence used by the loader: compiled defaults ← base.yaml ← env overlay ←
// env vars ← secret resolution. See docs/blueprint/12-configuration-and-deployment.md §3.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/qatoolist/wowapi/adapters/secrets/envprovider"
	"github.com/qatoolist/wowapi/kernel/config"
)

// configUsage prints help for `wowapi config` and its subcommands.
func configUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi config <subcommand> [flags]

In a product repo (with tools/configcheck, scaffolded by 'wowapi init'), these
commands run the product-local checker so they see the product's Config type; in
the framework repo they run against the framework config alone.

Subcommands:
  validate    load and validate config; CI gate (exit 0 = OK, 1 = invalid)
  print       print redacted effective config as JSON (--redacted is required)
  schema      print JSON Schema derived from struct tags (no config files needed)
  doctor      show per-key provenance table and fingerprint
              (note: environment-variable sanity probes arrive in Phase 10)
  diff        redacted effective-config diff (--from <env> --to <env>)
  capacity    check the concurrency capacity budget (backlog B6); CI lint
              (exit 0 = within budget or shape not configured, 1 = oversubscribed)

Shared flags (validate, print, doctor):
  --dir         directory holding base.yaml + <env>.yaml (default "configs")
  --base        explicit base file path; absent default skips the layer, not an error
  --env         environment name; overlay <dir>/<env>.yaml must exist when set
  --env-prefix  environment variable prefix (default "WOWAPI__")
`)
}

// cfgFlags holds the parsed shared config flags for a subcommand.
type cfgFlags struct {
	dir    string
	base   string
	env    string
	prefix string
}

// newCfgFlagSet returns a flag.FlagSet with shared config flags pre-registered.
// The caller must call fs.Parse(args) and then cfgFlags.resolve().
func newCfgFlagSet(subcmd string, stderr io.Writer) (*flag.FlagSet, *cfgFlags) {
	fs := flag.NewFlagSet("wowapi config "+subcmd, flag.ContinueOnError)
	fs.SetOutput(stderr)
	f := &cfgFlags{}
	fs.StringVar(&f.dir, "dir", "configs", "directory holding base.yaml + <env>.yaml")
	fs.StringVar(&f.base, "base", "", "explicit base file path")
	fs.StringVar(&f.env, "env", "", "environment name; overlay <dir>/<env>.yaml")
	fs.StringVar(&f.prefix, "env-prefix", "WOWAPI__", "env var prefix")
	return fs, f
}

// resolve computes the actual BaseFile and EnvFile paths after flag parsing.
//
// --base explicit: use the supplied path verbatim.
// --base absent: use <dir>/base.yaml if it exists; silently skip if not
// (allows env-var-only validation without requiring a file on disk).
// --env set: <dir>/<env>.yaml must exist; missing = clear error.
func (f *cfgFlags) resolve(fs *flag.FlagSet, subcmd string, stderr io.Writer) (baseFile, envFile string, ok bool) {
	baseExplicit := false
	fs.Visit(func(fl *flag.Flag) {
		if fl.Name == "base" {
			baseExplicit = true
		}
	})

	if baseExplicit {
		baseFile = f.base
	} else {
		candidate := filepath.Join(f.dir, "base.yaml")
		if _, err := os.Stat(candidate); err == nil {
			baseFile = candidate
		}
		// absent default → BaseFile="" → loader skips this layer (fail-closed
		// environment check still applies to whatever layers do supply values)
	}

	if f.env != "" {
		overlay := filepath.Join(f.dir, f.env+".yaml")
		if _, err := os.Stat(overlay); err != nil {
			fmt.Fprintf(stderr, "wowapi config %s: env overlay %q not found\n", subcmd, overlay)
			return "", "", false
		}
		envFile = overlay
	}

	return baseFile, envFile, true
}

// loaderOpts builds config.Options from resolved paths and the prefix.
// Secrets is always wired to envprovider.New() so that secretref://env/<VAR>
// references in config files resolve from the process environment.
func (f *cfgFlags) loaderOpts(baseFile, envFile string) config.Options {
	return config.Options{
		BaseFile:  baseFile,
		EnvFile:   envFile,
		EnvPrefix: f.prefix,
		Secrets:   envprovider.New(),
	}
}

// assertEnv enforces that the loaded environment matches --env when given:
// `config validate --env prod` is a CI gate promising prod rules were applied
// — a prod.yaml that (mis)declares another environment must fail the gate,
// not silently validate under laxer rules (review finding SEC-6).
func (f *cfgFlags) assertEnv(loadedEnv config.Env, subcmd string, stderr io.Writer) bool {
	if f.env == "" || loadedEnv == config.Env(f.env) {
		return true
	}
	fmt.Fprintf(stderr, "wowapi config %s: config declares environment %q but --env %s was requested — the overlay must set the environment it is named for\n",
		subcmd, string(loadedEnv), f.env)
	return false
}

// runConfig dispatches `wowapi config <subcommand>`.
func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		configUsage(stderr)
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "validate", "print", "schema", "doctor":
		// Prefer the product-local checker (it knows the product's Config type);
		// fall back to framework-only handling when it is absent.
		if handled, code := delegateConfigCheck(sub, rest, stdout, stderr); handled {
			return code
		}
		switch sub {
		case "validate":
			return runConfigValidate(rest, stdout, stderr)
		case "print":
			return runConfigPrint(rest, stdout, stderr)
		case "schema":
			return runConfigSchema(rest, stdout, stderr)
		case "doctor":
			return runConfigDoctor(rest, stdout, stderr)
		}
		return 2 // unreachable
	case "diff":
		// Prefer the product-local checker so the diff covers the product's Config
		// type (module + product-specific fields), matching validate/print/doctor;
		// fall back to the framework-only diff when no product checker is present.
		if handled, code := delegateConfigCheck("diff", rest, stdout, stderr); handled {
			return code
		}
		return runConfigDiff(rest, stdout, stderr)
	case "capacity":
		// Framework-only for now (backlog B6): the capacity-budget formula only
		// needs config.Framework's own fields (concurrency.* + db.max_conns), so
		// there is no product-specific type to delegate to yet.
		return runConfigCapacity(rest, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "wowapi config: unknown subcommand %q\n", sub)
		configUsage(stderr)
		return 2
	}
}

// runConfigValidate implements `wowapi config validate`.
//
// Exit 0: config OK — prints "config OK  fingerprint=<12hex>" to stdout and any
// warnings (prefixed "warning: ") to stderr.
// Exit 1: invalid config — prints the full accumulated error list to stderr.
func runConfigValidate(args []string, stdout, stderr io.Writer) int {
	fs, f := newCfgFlagSet("validate", stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	baseFile, envFile, ok := f.resolve(fs, "validate", stderr)
	if !ok {
		return 1
	}

	loaded, err := config.LoadDetailed[config.Framework](f.loaderOpts(baseFile, envFile))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !f.assertEnv(loaded.Config.Environment, "validate", stderr) {
		return 1
	}
	for _, w := range loaded.Warnings {
		fmt.Fprintf(stderr, "warning: %s\n", w)
	}
	fmt.Fprintf(stdout, "config OK  fingerprint=%s\n", loaded.Fingerprint.Short())
	return 0
}

// runConfigPrint implements `wowapi config print --redacted`.
//
// --redacted is required: the flag must be supplied (redaction is not optional).
// Without it the command exits 2. On load failure it exits 1. On success it
// prints json.MarshalIndent of the loaded Framework to stdout and exits 0.
func runConfigPrint(args []string, stdout, stderr io.Writer) int {
	fs, f := newCfgFlagSet("print", stderr)
	redacted := fs.Bool("redacted", false, "print redacted config (required; redaction is not optional)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if !*redacted {
		fmt.Fprintln(stderr, "wowapi config print: requires --redacted (redaction is not optional)")
		return 2
	}
	baseFile, envFile, ok := f.resolve(fs, "print", stderr)
	if !ok {
		return 1
	}

	loaded, err := config.LoadDetailed[config.Framework](f.loaderOpts(baseFile, envFile))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !f.assertEnv(loaded.Config.Environment, "print", stderr) {
		return 1
	}
	js, err := json.MarshalIndent(loaded.Config, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "wowapi config print: marshal: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(js))
	return 0
}

// runConfigSchema implements `wowapi config schema`.
//
// Derives a JSON Schema from config.Framework struct tags and prints it to
// stdout. Works without any config files on disk.
func runConfigSchema(_ []string, stdout, stderr io.Writer) int {
	js, err := config.Schema[config.Framework]()
	if err != nil {
		fmt.Fprintf(stderr, "wowapi config schema: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(js))
	return 0
}

// runConfigDoctor implements `wowapi config doctor`.
//
// Loads config, then prints a stable sorted per-key provenance table to stdout:
//
//	KEY         LAYER
//	environment base-file
//	http.addr   default
//	...
//
// followed by a fingerprint line. Warnings go to stderr.
// On load failure: prints the error to stderr and exits 1 (the table cannot
// be rendered without a valid load). Env-variable sanity probes arrive in Phase 10.
func runConfigDoctor(args []string, stdout, stderr io.Writer) int {
	fs, f := newCfgFlagSet("doctor", stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	baseFile, envFile, ok := f.resolve(fs, "doctor", stderr)
	if !ok {
		return 1
	}

	loaded, err := config.LoadDetailed[config.Framework](f.loaderOpts(baseFile, envFile))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !f.assertEnv(loaded.Config.Environment, "doctor", stderr) {
		return 1
	}

	keys := make([]string, 0, len(loaded.Provenance))
	for k := range loaded.Provenance {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tLAYER")
	for _, k := range keys {
		fmt.Fprintf(tw, "%s\t%s\n", k, loaded.Provenance[k])
	}
	_ = tw.Flush() // best-effort flush to stdout; a terminal write error has no recovery

	fmt.Fprintf(stdout, "fingerprint=%s\n", loaded.Fingerprint.Short())
	for _, w := range loaded.Warnings {
		fmt.Fprintf(stderr, "warning: %s\n", w)
	}
	return 0
}

// runConfigCapacity implements `wowapi config capacity` (backlog B6): a CI
// lint independent of concurrency.capacity_mode — it always fails (exit 1) on
// an oversubscribed deployment shape, regardless of whether the deployment
// itself runs in advisory or enforced mode. capacity_mode governs whether
// `wowapi config validate` / process boot fails on the same condition; this
// subcommand exists so a CI pipeline can catch the problem early even while
// production boot stays advisory (rollout guard: advisory-then-enforced).
//
// Exit 0: the shape fits under db.max_conns, OR no shape is declared
// (concurrency.replicas == 0) — printed as "capacity OK ... (shape not
// configured)" so the difference is visible without being a failure.
// Exit 1: the computed demand exceeds db.max_conns; the formula and the
// numbers are printed to stderr.
func runConfigCapacity(args []string, stdout, stderr io.Writer) int {
	fs, f := newCfgFlagSet("capacity", stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	baseFile, envFile, ok := f.resolve(fs, "capacity", stderr)
	if !ok {
		return 1
	}

	loaded, err := config.LoadDetailed[config.Framework](f.loaderOpts(baseFile, envFile))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !f.assertEnv(loaded.Config.Environment, "capacity", stderr) {
		return 1
	}
	for _, w := range loaded.Warnings {
		fmt.Fprintf(stderr, "warning: %s\n", w)
	}

	if loaded.Config.Concurrency.Replicas == 0 {
		fmt.Fprintln(stdout, "capacity OK  (deployment shape not configured: concurrency.replicas is 0, check skipped)")
		return 0
	}
	if problem := config.CheckCapacity(loaded.Config); problem != nil {
		fmt.Fprintln(stderr, problem.Error())
		return 1
	}
	fmt.Fprintf(stdout, "capacity OK  replicas=%d runtime_pool_max=%d platform_pool_max=%d migrate_pool_max=%d reserved_admin=%d db.max_conns=%d\n",
		loaded.Config.Concurrency.Replicas, loaded.Config.Concurrency.RuntimePoolMax, loaded.Config.Concurrency.PlatformPoolMax,
		loaded.Config.Concurrency.MigratePoolMax, loaded.Config.Concurrency.ReservedAdmin, loaded.Config.DB.MaxConns)
	return 0
}
