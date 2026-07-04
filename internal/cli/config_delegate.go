// config_delegate.go — product-local config delegation + `config diff`.
//
// The installed wowapi binary cannot import product config types (it is prebuilt),
// so `wowapi config <mode>` delegates to the product-local checker scaffolded by
// `wowapi init` at tools/configcheck (blueprint 12 §8). When that checker is
// absent (e.g. inside the framework repo) the commands fall back to validating
// config.Framework alone.
package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/qatoolist/wowapi/kernel/config"
)

// delegateConfigCheck runs the product-local tools/configcheck for mode when it
// exists in the working directory, forwarding args and stdio. It returns
// handled=false when no product checker is present so the caller runs the
// framework-only path.
func delegateConfigCheck(mode string, args []string, stdout, stderr io.Writer) (handled bool, code int) {
	if _, err := os.Stat(filepath.Join("tools", "configcheck", "main.go")); err != nil {
		return false, 0 // no product checker → framework-only handling
	}
	runArgs := append([]string{"run", "./tools/configcheck", mode}, args...)
	cmd := exec.Command("go", runArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
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
