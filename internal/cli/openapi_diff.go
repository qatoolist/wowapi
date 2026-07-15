package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pb33f/libopenapi"
)

func runOpenAPIDiff(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi openapi diff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	baselinePath := fs.String("baseline", "", "previous released merged OpenAPI document")
	currentPath := fs.String("current", "", "current merged OpenAPI document")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *baselinePath == "" || *currentPath == "" {
		fmt.Fprintln(stderr, "wowapi openapi diff: --baseline and --current are required")
		return 2
	}
	baseline, err := os.ReadFile(*baselinePath) // #nosec G304 -- compatibility CLI intentionally reads caller-selected contracts
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi diff: baseline: %v\n", err)
		return 1
	}
	current, err := os.ReadFile(*currentPath) // #nosec G304 -- compatibility CLI intentionally reads caller-selected contracts
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi diff: current: %v\n", err)
		return 1
	}
	if err := validateOpenAPI31(baseline); err != nil {
		fmt.Fprintf(stderr, "wowapi openapi diff: baseline %v\n", err)
		return 1
	}
	if err := validateOpenAPI31(current); err != nil {
		fmt.Fprintf(stderr, "wowapi openapi diff: current %v\n", err)
		return 1
	}
	oldDocument, err := libopenapi.NewDocument(baseline)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi diff: baseline parse: %v\n", err)
		return 1
	}
	newDocument, err := libopenapi.NewDocument(current)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi diff: current parse: %v\n", err)
		return 1
	}
	changes, err := libopenapi.CompareDocuments(oldDocument, newDocument)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi diff: compare: %v\n", err)
		return 1
	}
	breaking := changes.TotalBreakingChanges()
	total := changes.TotalChanges()
	if breaking == 0 {
		fmt.Fprintf(stdout, "openapi diff: 0 breaking, %d total changes\n", total)
		return 0
	}
	fmt.Fprintf(stderr, "openapi diff: %d breaking, %d total changes\n", breaking, total)
	for _, change := range changes.GetAllChanges() {
		if change.Breaking {
			fmt.Fprintf(stderr, "- %s: %s -> %s\n", change.Property, change.Original, change.New)
		}
	}
	return 1
}
