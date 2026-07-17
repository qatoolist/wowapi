// openapi_cmd.go — `wowapi openapi merge` (Phase 10). Merges per-module OpenAPI
// fragments (the JSON each module registers via ctx.OpenAPI) into one document,
// failing loudly on a duplicate path or schema so two modules cannot silently
// clobber each other's API surface.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// firstToken returns a short, safe description of a fragment's leading bytes for
// an error message (never echoes a large body).
func firstToken(b []byte) string {
	if len(b) == 0 {
		return "empty file"
	}
	return strings.TrimSpace(string(b[:min(len(b), 12)]))
}

func openapiUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi openapi merge [flags] [fragment.json ...]
       wowapi openapi diff --baseline <released.json> --current <merged.json>

Merge OpenAPI 3.1 fragments complete-or-loud, or enforce the supported-line semantic
compatibility policy against a previous released merged document.

merge flags:
  --dir DIR       directory containing fragment JSON files (default ".")
  --title TITLE   merged info.title (default "wowapi API")
  --version VER   merged info.version (default "0.0.0")
  --out FILE      write merged JSON to FILE instead of stdout
`)
}

func runOpenAPI(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		openapiUsage(stderr)
		return 2
	}
	switch args[0] {
	case "-h", "--help", "help":
		openapiUsage(stdout)
		return 0
	case "diff":
		return runOpenAPIDiff(args[1:], stdout, stderr)
	case "merge":
	default:
		openapiUsage(stderr)
		return 2
	}

	fs := flag.NewFlagSet("wowapi openapi merge", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", ".", "directory of *.json fragments")
	title := fs.String("title", "wowapi API", "info.title")
	version := fs.String("version", "0.0.0", "info.version")
	out := fs.String("out", "", "output file (default stdout)")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	files, err := gatherFragments(*dir, fs.Args())
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi merge: %v\n", err)
		return 1
	}
	if len(files) == 0 {
		fmt.Fprintln(stderr, "wowapi openapi merge: no fragments found")
		return 1
	}

	merged, err := newOpenAPIMergeState(*title, *version)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi merge: %v\n", err)
		return 1
	}
	for _, file := range files {
		if err := merged.mergeFile(file); err != nil {
			fmt.Fprintf(stderr, "wowapi openapi merge: %v\n", err)
			return 1
		}
	}
	encoded, err := merged.document()
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi merge: %v\n", err)
		return 1
	}
	if *out == "" {
		if _, err := stdout.Write(encoded); err != nil {
			fmt.Fprintf(stderr, "wowapi openapi merge: write: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*out, encoded, 0o644); err != nil { // #nosec G306 -- merged OpenAPI spec is a build artifact meant to be world-readable, like source
		fmt.Fprintf(stderr, "wowapi openapi merge: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, *out)
	return 0
}

func gatherFragments(dir string, extra []string) ([]string, error) {
	var files []string
	if dir != "" {
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}
	files = append(files, extra...)
	sort.Strings(files)
	return files, nil
}
