package compatcli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/qatoolist/wowapi/internal/compat"
)

// Run executes the compatibility gate CLI and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "config" {
		_, _ = fmt.Fprintln(stderr, "usage: compatcheck config --baseline <schema.json> --current <schema.json>")
		return 2
	}
	fs := flag.NewFlagSet("compatcheck config", flag.ContinueOnError)
	fs.SetOutput(stderr)
	baselinePath := fs.String("baseline", "", "previous released JSON Schema")
	currentPath := fs.String("current", "", "current generated JSON Schema")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *baselinePath == "" || *currentPath == "" {
		_, _ = fmt.Fprintln(stderr, "compatcheck config: --baseline and --current are required")
		return 2
	}
	baseline, err := os.ReadFile(*baselinePath) // #nosec G304 -- CI tool intentionally reads caller-selected schema files
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "compatcheck config: baseline: %v\n", err)
		return 1
	}
	current, err := os.ReadFile(*currentPath) // #nosec G304 -- CI tool intentionally reads caller-selected schema files
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "compatcheck config: current: %v\n", err)
		return 1
	}
	if err := compat.CheckConfigSchemaCompatibility(baseline, current); err != nil {
		_, _ = fmt.Fprintf(stderr, "compatcheck config: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintln(stdout, "config schemas are backward compatible")
	return 0
}
