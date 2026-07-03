// openapi_cmd.go — `wowapi openapi merge` (Phase 10). Merges per-module OpenAPI
// fragments (the JSON each module registers via ctx.OpenAPI) into one document,
// failing loudly on a duplicate path or schema so two modules cannot silently
// clobber each other's API surface.
package cli

import (
	"bytes"
	"encoding/json"
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

Merge OpenAPI 3.1 fragments (module openapi.json files) into one document.
Fragments are taken from --dir (all *.json) and/or explicit file arguments.

Flags:
  --dir       directory of *.json fragments (default "."; set "" to use only args)
  --title     info.title for the merged doc (default "wowapi API")
  --version   info.version for the merged doc (default "0.0.0")
  --out       output file (default: stdout)
`)
}

func runOpenAPI(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "merge" && args[0] != "-h" && args[0] != "--help" && args[0] != "help") {
		openapiUsage(stderr)
		return 2
	}
	if args[0] != "merge" {
		openapiUsage(stdout)
		return 0
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

	paths := map[string]any{}
	schemas := map[string]any{}
	for _, f := range files {
		if err := mergeFragment(f, paths, schemas); err != nil {
			fmt.Fprintf(stderr, "wowapi openapi merge: %v\n", err)
			return 1
		}
	}

	doc := map[string]any{
		"openapi": "3.1.0",
		"info":    map[string]any{"title": *title, "version": *version},
		"paths":   paths,
	}
	if len(schemas) > 0 {
		doc["components"] = map[string]any{"schemas": schemas}
	}
	enc, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "wowapi openapi merge: %v\n", err)
		return 1
	}
	enc = append(enc, '\n')
	if *out == "" {
		if _, err := stdout.Write(enc); err != nil {
			fmt.Fprintf(stderr, "wowapi openapi merge: write: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*out, enc, 0o644); err != nil {
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

func mergeFragment(path string, paths, schemas map[string]any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// The fragment must be a JSON OBJECT — `null`, an array, or a scalar
	// unmarshals into the struct below with nil fields and would contribute
	// nothing SILENTLY (CLI-02). Reject anything that is not an object.
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("%s: expected a JSON object (OpenAPI fragment), got %s", path, firstToken(trimmed))
	}
	var frag struct {
		Paths      map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &frag); err != nil {
		return fmt.Errorf("%s: invalid JSON: %w", path, err)
	}
	for p, v := range frag.Paths {
		if _, dup := paths[p]; dup {
			return fmt.Errorf("%s: duplicate path %q already defined by another fragment", path, p)
		}
		paths[p] = v
	}
	for name, v := range frag.Components.Schemas {
		if _, dup := schemas[name]; dup {
			return fmt.Errorf("%s: duplicate component schema %q already defined by another fragment", path, name)
		}
		schemas[name] = v
	}
	return nil
}
