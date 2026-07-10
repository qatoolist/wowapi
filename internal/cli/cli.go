// Package cli implements the wowapi command dispatcher. cmd/wowapi is a thin
// main over Run so behavior is unit-testable. Private implementation detail.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/qatoolist/wowapi/internal/buildinfo"
)

// planned maps not-yet-implemented commands to the phase that delivers them
// (docs/implementation/phase-plan.md). Keeping the full surface visible from
// day one makes `wowapi help` an honest roadmap.
var planned = map[string]string{}

// Run executes the CLI and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stdout)
		return 0
	}
	switch cmd := args[0]; cmd {
	case "version":
		return runVersion(stdout, stderr)
	case "help", "-h", "--help":
		usage(stdout)
		return 0
	case "config":
		return runConfig(args[1:], stdout, stderr)
	case "migrate":
		return runMigrate(args[1:], stdout, stderr)
	case "seed":
		return runSeed(args[1:], stdout, stderr)
	case "i18n":
		return runI18n(args[1:], stdout, stderr)
	case "openapi":
		return runOpenAPI(args[1:], stdout, stderr)
	case "lint":
		return runLint(args[1:], stdout, stderr)
	case "deploy":
		return runDeploy(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "new-module":
		return runNewModule(args[1:], stdout, stderr)
	case "gen":
		return runGen(args[1:], stdout, stderr)
	case "dlq":
		return runDLQ(args[1:], stdout, stderr)
	case "apikey":
		return runApikey(args[1:], stdout, stderr)
	case "audit":
		return runAudit(args[1:], stdout, stderr)
	default:
		if phase, ok := planned[cmd]; ok {
			fmt.Fprintf(stderr, "wowapi %s: not implemented yet — planned in %s.\n", cmd, phase)
			fmt.Fprintf(stderr, "See docs/implementation/phase-plan.md in %s.\n", buildinfo.ModulePath)
			return 2
		}
		fmt.Fprintf(stderr, "wowapi: unknown command %q — run `wowapi help`.\n", cmd)
		return 2
	}
}

func runVersion(stdout, stderr io.Writer) int {
	cliVersion := buildinfo.Version()
	fmt.Fprintf(stdout, "wowapi %s\n", cliVersion)

	wd, err := os.Getwd()
	if err != nil {
		return 0
	}
	gm, ok := buildinfo.FindGoMod(wd)
	if !ok {
		return 0
	}
	switch {
	case gm.IsFramework():
		fmt.Fprintln(stdout, "context: wowapi framework repository")
	case gm.WowapiVersion != "":
		fmt.Fprintf(stdout, "dependency: %s %s (module %s)\n", buildinfo.ModulePath, gm.WowapiVersion, gm.ModulePath)
		if gm.WowapiVersion != cliVersion {
			fmt.Fprintf(stderr, "warning: CLI version %s differs from the %s dependency %s — install the matching CLI:\n", cliVersion, buildinfo.ModulePath, gm.WowapiVersion)
			fmt.Fprintf(stderr, "  go install %s/cmd/wowapi@%s\n", buildinfo.ModulePath, gm.WowapiVersion)
		}
	}
	return 0
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `wowapi — framework CLI for %s

Usage:
  wowapi <command> [flags]

Available commands:
  version      print CLI version and check the go.mod dependency version
  help         this help
  config       validate|print|schema|doctor  (run `+"`wowapi config`"+` for details)
  init         scaffold a product repository
  new-module   scaffold a product module
  gen          run code generators (crud)
  migrate      create the next-numbered migration file
  seed         validate a module's seed bundle
  i18n         validate a product's locale catalogs (coverage, ownership, placeholders)
  openapi      merge OpenAPI fragments into one document
  lint         boundaries — module isolation + layering check
  deploy       render deployment manifests (compose|env)
  dlq          inspect/replay/discard dead-letter jobs and events
`, buildinfo.ModulePath)
}
