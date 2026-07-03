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
var planned = map[string]string{
	"init":       "Phase 10",
	"new-module": "Phase 10",
	"gen":        "Phase 10",
	"migrate":    "Phase 10",
	"seed":       "Phase 10",
	"openapi":    "Phase 10",
	"lint":       "Phase 10",
	"config":     "Phase 1 (validate/print/schema for framework config), Phase 10 (product repos)",
	"deploy":     "Phase 10",
}

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

Planned commands (see docs/implementation/phase-plan.md):
  init         scaffold a product repository            (%s)
  new-module   scaffold a product module                (%s)
  gen          run code generators (crud, sqlc, mocks)  (%s)
  migrate      migration helpers                        (%s)
  seed         seed validation                          (%s)
  openapi      merge/check OpenAPI fragments            (%s)
  lint         boundary lint                            (%s)
  config       config init/validate/doctor/print/diff/schema (%s)
  deploy       render deployment manifests              (%s)
`, buildinfo.ModulePath,
		planned["init"], planned["new-module"], planned["gen"], planned["migrate"],
		planned["seed"], planned["openapi"], planned["lint"], planned["config"], planned["deploy"])
}
