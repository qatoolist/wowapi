// config_delegate.go — product-local config delegation + `config diff`.
//
// The installed wowapi binary cannot import product config types (it is prebuilt),
// so `wowapi config <mode>` delegates to the product-local checker scaffolded by
// `wowapi init` at tools/configcheck (blueprint 12 §8). When that checker is
// absent (e.g. inside the framework repo) the commands fall back to validating
// config.Framework alone.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/qatoolist/wowapi/adapters/secrets/envprovider"
	"github.com/qatoolist/wowapi/kernel/config"
)

// delegateConfigCheck runs the product-local tools/configcheck for mode when it
// exists in the product root. The product root is discovered via --project, or
// by running `go env GOMOD` in the current directory, so delegation works from
// nested subdirectories or outside the repo. It prints an explicit message to
// stderr reporting whether product validation ran.
//
// It returns handled=false when no product checker is present so the caller
// runs the framework-only path.
func delegateConfigCheck(mode string, args []string, stdout, stderr io.Writer) (handled bool, code int) {
	root, args, err := resolveProductRoot(args)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi config %s: product root discovery: %v\n", mode, err)
		return false, 0 // fail open to framework-only path
	}

	checker := filepath.Join(root, "tools", "configcheck", "main.go")
	if _, err := os.Stat(checker); err != nil {
		fmt.Fprintf(stderr, "product validation: skipped (no product checker at %s)\n", checker)
		return false, 0
	}

	runArgs := append([]string{"run", "./tools/configcheck", mode}, args...)
	// Deliberate subprocess: the CLI's job here IS to run the product's own
	// configcheck via the local go toolchain with caller-supplied args (W01-E01
	// gosec triage). Context-bound so the launch is cancellable in principle.
	cmd := exec.CommandContext(context.Background(), "go", runArgs...) // #nosec G204 -- runs the repo's own `go run ./tools/configcheck`; args are the CLI caller's own
	cmd.Dir = root
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	fmt.Fprintf(stderr, "product validation: engaged (%s)\n", root)
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return true, ee.ExitCode()
		}
		fmt.Fprintf(stderr, "wowapi config %s: running product configcheck: %v\n", mode, err)
		return true, 1
	}
	return true, 0
}

// resolveProductRoot extracts an explicit --project flag from args and, if
// absent, asks the local Go toolchain for the module root. It returns the root
// and the remaining args (with --project removed).
func resolveProductRoot(args []string) (string, []string, error) {
	root := ""
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--project":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--project requires a value")
			}
			root = args[i+1]
			i++
		case strings.HasPrefix(a, "--project="):
			root = strings.TrimPrefix(a, "--project=")
		default:
			out = append(out, a)
		}
	}
	if root != "" {
		abs, err := filepath.Abs(root)
		if err != nil {
			return "", nil, err
		}
		return abs, out, nil
	}
	cmd := exec.CommandContext(context.Background(), "go", "env", "GOMOD")
	outBytes, err := cmd.Output()
	if err != nil {
		return "", nil, fmt.Errorf("go env GOMOD: %w", err)
	}
	mod := strings.TrimSpace(string(outBytes))
	if mod == "" || mod == "/dev/null" {
		return "", nil, fmt.Errorf("not inside a Go module (and --project not given)")
	}
	return filepath.Dir(mod), out, nil
}

// runConfigDiff implements `wowapi config diff --from <env> --to <env>`: a
// redacted effective-config diff between two environments (framework-side; the
// product checker handles it when present).
func runConfigDiff(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi config diff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "configs", "directory holding base.yaml + <env>.yaml")
	prefix := fs.String("env-prefix", "WOWAPI__", "environment variable prefix")
	from := fs.String("from", "", "source environment (required)")
	to := fs.String("to", "", "target environment (required)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *from == "" || *to == "" {
		fmt.Fprintln(stderr, "wowapi config diff: both --from and --to are required")
		return 2
	}

	load := func(env string) (string, error) {
		loaded, err := config.LoadDetailed[config.Framework](config.Options{
			BaseFile:  filepath.Join(*dir, "base.yaml"),
			EnvFile:   filepath.Join(*dir, env+".yaml"),
			EnvPrefix: *prefix,
			// Wire the env secret provider so secretref://env/<VAR> values resolve;
			// without it a normal DSN config errors out during the diff load.
			Secrets: envprovider.New(),
		})
		if err != nil {
			return "", err
		}
		js, err := json.MarshalIndent(loaded.Config, "", "  ")
		return string(js), err
	}

	fromJSON, err := load(*from)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi config diff: load %s: %v\n", *from, err)
		return 1
	}
	toJSON, err := load(*to)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi config diff: load %s: %v\n", *to, err)
		return 1
	}

	if fromJSON == toJSON {
		fmt.Fprintf(stdout, "config diff %s -> %s: identical\n", *from, *to)
		return 0
	}
	fmt.Fprintf(stdout, "--- %s\n+++ %s\n", *from, *to)
	printLineDiff(stdout, fromJSON, toJSON)
	return 0
}

// printLineDiff emits a minimal line-oriented diff: lines only in `a` are
// prefixed "-", lines only in `b` "+". Both inputs are already redacted JSON, so
// no secret can appear. (A full LCS diff is overkill for a canonical config dump
// whose key order is stable.)
func printLineDiff(w io.Writer, a, b string) {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	inB := make(map[string]int, len(bl))
	for _, l := range bl {
		inB[l]++
	}
	inA := make(map[string]int, len(al))
	for _, l := range al {
		inA[l]++
	}
	for _, l := range al {
		if inB[l] == 0 {
			fmt.Fprintf(w, "- %s\n", strings.TrimSpace(l))
		}
	}
	for _, l := range bl {
		if inA[l] == 0 {
			fmt.Fprintf(w, "+ %s\n", strings.TrimSpace(l))
		}
	}
}
