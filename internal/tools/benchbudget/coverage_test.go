package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// coverage_test.go — QA G8: completes coverage of the CI performance gate.
// Covers the parseBenchOutput edge branches (malformed/short lines, unparsable
// values), the loadBudgets scanner-error branch, and every exit path of main()
// (usage error, load error, parse error, budget violation — including a
// budgeted-but-missing benchmark, PERF-06 T1) plus its in-process happy path.

// writeFile writes content to a fresh file in dir and returns its path.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

// openFile opens p read-only, registering cleanup.
func openFile(t *testing.T, p string) *os.File {
	t.Helper()
	f, err := os.Open(p)
	if err != nil {
		t.Fatalf("open %s: %v", p, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func TestLoadBudgetsFileNotFound(t *testing.T) {
	_, err := loadBudgets(filepath.Join(t.TempDir(), "does-not-exist.txt"))
	if err == nil {
		t.Fatal("loadBudgets should error on missing file")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("want not-exist error, got %v", err)
	}
}

// TestLoadBudgetsScannerError triggers the bufio.Scanner error branch
// (line "sc.Err()") using a single token longer than bufio's 64KiB buffer.
func TestLoadBudgetsScannerError(t *testing.T) {
	dir := t.TempDir()
	huge := strings.Repeat("A", 128*1024) // no newline, no spaces -> ErrTooLong
	p := writeFile(t, dir, "huge.txt", huge)
	_, err := loadBudgets(p)
	if err == nil {
		t.Fatal("loadBudgets should surface the scanner error on an over-long line")
	}
}

func TestParseBenchOutputEdgeCases(t *testing.T) {
	dir := t.TempDir()
	content := strings.Join([]string{
		"Benchmark short", // starts with Benchmark but <3 fields -> skipped
		"BenchmarkBadNs-8 \t 10 \t notanumber ns/op",             // ns/op unparsable -> nsPerOp stays <0 -> line skipped
		"BenchmarkNoMem-8 \t 100 \t 12.5 ns/op",                  // no -benchmem -> allocs defaults to 0
		"BenchmarkBadAlloc-8 \t 100 \t 7.0 ns/op \t x allocs/op", // allocs unparsable -> defaults to 0, ns kept
		"not a benchmark line at all",                            // ignored (no Benchmark prefix)
	}, "\n") + "\n"
	p := writeFile(t, dir, "bench.txt", content)

	res, err := parseBenchOutput(openFile(t, p))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := res["BenchmarkShort"]; ok {
		t.Error("short line should not produce a result")
	}
	if _, ok := res["BenchmarkBadNs"]; ok {
		t.Error("line with unparsable ns/op should be skipped")
	}
	nm, ok := res["BenchmarkNoMem"]
	if !ok {
		t.Fatal("BenchmarkNoMem missing")
	}
	if nm.nsPerOp != 12.5 || nm.allocsPerOp != 0 {
		t.Fatalf("BenchmarkNoMem = %+v, want {12.5 0}", nm)
	}
	ba, ok := res["BenchmarkBadAlloc"]
	if !ok {
		t.Fatal("BenchmarkBadAlloc missing")
	}
	if ba.nsPerOp != 7.0 || ba.allocsPerOp != 0 {
		t.Fatalf("BenchmarkBadAlloc = %+v, want {7 0} (unparsable allocs -> 0)", ba)
	}
	if len(res) != 2 {
		t.Fatalf("want exactly 2 valid results, got %d: %+v", len(res), res)
	}
}

// TestMainMissingBenchmarkFails re-exercises the exit-path harness below to
// confirm a budgeted-but-absent benchmark causes a real CI failure (exit 1),
// not just a warning. See PERF-06 T1.
func TestMainMissingBenchmarkFails(t *testing.T) {
	if mode := os.Getenv("BB_MAIN_MODE"); mode != "" {
		runMainChild(t, mode)
		return
	}

	dir := t.TempDir()
	// BenchmarkGhost is budgeted but never appears in the bench output —
	// simulates a renamed/deleted benchmark.
	budgets := writeFile(t, dir, "budgets.txt", "BenchmarkGhost 100 0\n")
	bench := writeFile(t, dir, "bench.txt",
		"BenchmarkOther-8 \t 1000000 \t 30.0 ns/op \t 0 B/op \t 2 allocs/op\n")

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainMissingBenchmarkFails$")
	cmd.Env = append(os.Environ(), "BB_MAIN_MODE=violation:"+budgets)
	cmd.Stdin = openFile(t, bench)
	out, err := cmd.CombinedOutput()

	ee, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("missing budgeted benchmark should exit non-zero (%v); output: %s", err, out)
	}
	if got := ee.ExitCode(); got != 1 {
		t.Fatalf("exit code = %d, want 1; output: %s", got, out)
	}
	if !strings.Contains(string(out), "BenchmarkGhost") || !strings.Contains(string(out), "budgeted but not found in bench output") {
		t.Fatalf("output should explain the missing-benchmark failure, got: %s", out)
	}
}

// TestParseBenchOutputScannerError covers parseBenchOutput's sc.Err() branch.
func TestParseBenchOutputScannerError(t *testing.T) {
	dir := t.TempDir()
	// Over-long line beginning with "Benchmark" so it reaches the parse loop
	// path in spirit, but the scanner errors before yielding a full token.
	huge := "Benchmark" + strings.Repeat("A", 128*1024)
	p := writeFile(t, dir, "huge.txt", huge)
	_, err := parseBenchOutput(openFile(t, p))
	if err == nil {
		t.Fatal("parseBenchOutput should surface the scanner error")
	}
}

// ---- main() in-process happy path (no os.Exit is reached) -------------------

// captureMain runs main() with the given argv and stdin file, capturing
// stdout+stderr. It only returns normally when main() does not call os.Exit.
func captureMain(t *testing.T, argv []string, stdinPath string) (stdout, stderr string) {
	t.Helper()
	origArgs, origOut, origErr, origIn := os.Args, os.Stdout, os.Stderr, os.Stdin
	defer func() { os.Args, os.Stdout, os.Stderr, os.Stdin = origArgs, origOut, origErr, origIn }()

	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout, os.Stderr = outW, errW
	os.Stdin = openFile(t, stdinPath)
	os.Args = argv

	main()

	outW.Close()
	errW.Close()
	return readAll(outR), readAll(errR)
}

func readAll(f *os.File) string {
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return sb.String()
}

func TestMainHappyPathOK(t *testing.T) {
	dir := t.TempDir()
	// Both budgeted benchmarks are present and within budget: no violations,
	// so main() returns without calling os.Exit.
	budgets := writeFile(t, dir, "budgets.txt",
		"BenchmarkFast 1000 5\nBenchmarkFaster 1000 5\n")
	bench := writeFile(t, dir, "bench.txt",
		"BenchmarkFast-8 \t 1000000 \t 30.0 ns/op \t 0 B/op \t 2 allocs/op\n"+
			"BenchmarkFaster-8 \t 1000000 \t 10.0 ns/op \t 0 B/op \t 1 allocs/op\n")

	stdout, _ := captureMain(t, []string{"benchbudget", budgets}, bench)

	if !strings.Contains(stdout, "OK") || !strings.Contains(stdout, "BenchmarkFast") {
		t.Errorf("stdout should report OK for BenchmarkFast, got %q", stdout)
	}
	if !strings.Contains(stdout, "OK") || !strings.Contains(stdout, "BenchmarkFaster") {
		t.Errorf("stdout should report OK for BenchmarkFaster, got %q", stdout)
	}
}

// ---- main() os.Exit paths via subprocess re-exec ---------------------------
//
// The child process is this same coverage-instrumented test binary; it inherits
// GOCOVERDIR through os.Environ(), so its coverage merges into the profile.

func TestMainExitPaths(t *testing.T) {
	if mode := os.Getenv("BB_MAIN_MODE"); mode != "" {
		runMainChild(t, mode)
		return
	}

	dir := t.TempDir()
	budgets := writeFile(t, dir, "budgets.txt", "BenchmarkSlow 100 0\n")
	overBench := writeFile(t, dir, "over.txt",
		"BenchmarkSlow-8 \t 1000 \t 999.0 ns/op \t 64 B/op \t 9 allocs/op\n")

	cases := []struct {
		name     string
		mode     string
		stdin    string
		wantExit int
		wantErr  string
	}{
		{"usage", "usage", overBench, 2, "usage: benchbudget"},
		{"loaderr", "loaderr", overBench, 2, "load budgets"},
		{"parseerr", "parseerr", overBench, 2, "parse bench output"},
		{"violation", "violation:" + budgets, overBench, 1, "budget violation"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestMainExitPaths$")
			cmd.Env = append(os.Environ(), "BB_MAIN_MODE="+tc.mode)
			f := openFile(t, tc.stdin)
			cmd.Stdin = f
			out, err := cmd.CombinedOutput()

			ee, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("child should exit non-zero (%v); output: %s", err, out)
			}
			if got := ee.ExitCode(); got != tc.wantExit {
				t.Fatalf("exit code = %d, want %d; output: %s", got, tc.wantExit, out)
			}
			if !strings.Contains(string(out), tc.wantErr) {
				t.Fatalf("output %q should contain %q", out, tc.wantErr)
			}
		})
	}
}

// runMainChild is the subprocess entry point: it configures os.Args for the
// requested mode and calls main(), which exits the process.
func runMainChild(t *testing.T, mode string) {
	switch {
	case mode == "usage":
		os.Args = []string{"benchbudget"} // no budget-file arg -> exit 2
	case mode == "loaderr":
		os.Args = []string{"benchbudget", filepath.Join(t.TempDir(), "nope.txt")}
	case mode == "parseerr":
		// stdin is a directory: os.File.Read on a dir yields a read error that
		// the scanner surfaces from parseBenchOutput.
		dir := t.TempDir()
		budgets := writeFile(t, dir, "b.txt", "BenchmarkSlow 100 0\n")
		d, err := os.Open(dir)
		if err != nil {
			t.Fatalf("open dir: %v", err)
		}
		os.Stdin = d
		os.Args = []string{"benchbudget", budgets}
	case strings.HasPrefix(mode, "violation:"):
		os.Args = []string{"benchbudget", strings.TrimPrefix(mode, "violation:")}
	default:
		t.Fatalf("unknown child mode %q", mode)
	}
	main()
}
