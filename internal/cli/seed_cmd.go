// seed_cmd.go — `wowapi seed validate` (Phase 10). Loads a module's seed bundle
// through the same kernel/seeds.Load the app uses at boot, so a seed error is
// caught in CI (exit 1) rather than at deploy time.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/qatoolist/wowapi/kernel/seeds"
)

func seedUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi seed validate [flags]

Load and validate a module's seed bundle (permissions, roles, resource types,
relationship types) with the same strict rules the app applies at boot.

Flags:
  --dir      directory holding the module's seed YAML (default "seeds")
  --module   module name that owns these seeds (required; keys must be prefixed)
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
