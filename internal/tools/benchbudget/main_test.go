package main

import (
	"os"
	"path/filepath"
	"testing"
)

// main_test.go — QA G8 (tooling reliability): benchbudget is the CI performance
// gate. If its parsing or thresholding is wrong, the gate silently passes
// regressions or fails spuriously. Its pure logic (baseName, loadBudgets,
// parseBenchOutput) had no test.

func TestBaseNameStripsGomaxprocsSuffix(t *testing.T) {
	cases := map[string]string{
		"BenchmarkFoo-16":   "BenchmarkFoo",
		"BenchmarkFoo-1":    "BenchmarkFoo",
		"BenchmarkFoo":      "BenchmarkFoo", // no suffix
		"Benchmark-Bar-8":   "Benchmark-Bar",
		"BenchmarkX-notnum": "BenchmarkX-notnum", // suffix isn't a number → kept
	}
	for in, want := range cases {
		if got := baseName(in); got != want {
			t.Errorf("baseName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadBudgetsValid(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "budgets.txt")
	os.WriteFile(p, []byte("# a comment\n\nBenchmarkA 300 0\nBenchmarkB 8000 -1\n"), 0o644)
	b, err := loadBudgets(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 2 {
		t.Fatalf("want 2 budgets (comments/blanks ignored), got %d", len(b))
	}
	if b["BenchmarkA"].maxNsPerOp != 300 || b["BenchmarkA"].maxAllocsPerOp != 0 {
		t.Fatalf("BenchmarkA parsed wrong: %+v", b["BenchmarkA"])
	}
	if b["BenchmarkB"].maxAllocsPerOp != -1 {
		t.Fatalf("-1 (unchecked) allocs not preserved: %+v", b["BenchmarkB"])
	}
}

func TestLoadBudgetsRejectsMalformed(t *testing.T) {
	dir := t.TempDir()
	for _, bad := range []string{
		"BenchmarkA 300\n",         // too few fields
		"BenchmarkA 300 0 extra\n", // too many fields
		"BenchmarkA notnum 0\n",    // bad ns
		"BenchmarkA 300 notnum\n",  // bad allocs
	} {
		p := filepath.Join(dir, "b.txt")
		os.WriteFile(p, []byte(bad), 0o644)
		if _, err := loadBudgets(p); err == nil {
			t.Errorf("loadBudgets should reject %q", bad)
		}
	}
}

func TestParseBenchOutput(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bench.txt")
	// Realistic `go test -bench -benchmem` lines (tab-separated), plus noise.
	content := "goos: darwin\n" +
		"BenchmarkFoo-16   \t 1000000 \t   30.5 ns/op \t 0 B/op \t 0 allocs/op\n" +
		"BenchmarkBar-16   \t   50000 \t  724.0 ns/op \t 512 B/op \t 22 allocs/op\n" +
		"PASS\n" +
		"BenchmarkFoo-16   \t 2000000 \t   45.0 ns/op \t 0 B/op \t 1 allocs/op\n" // duplicate → keep worst
	os.WriteFile(p, []byte(content), 0o644)
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	res, err := parseBenchOutput(f)
	if err != nil {
		t.Fatal(err)
	}
	// Names are base names (suffix stripped); non-benchmark lines ignored.
	if len(res) != 2 {
		t.Fatalf("want 2 benchmark results, got %d (%v)", len(res), res)
	}
	// Foo: the WORST (highest ns/op) of the two lines is kept.
	if res["BenchmarkFoo"].nsPerOp != 45.0 || res["BenchmarkFoo"].allocsPerOp != 1 {
		t.Fatalf("BenchmarkFoo should keep the worst result, got %+v", res["BenchmarkFoo"])
	}
	if res["BenchmarkBar"].nsPerOp != 724.0 || res["BenchmarkBar"].allocsPerOp != 22 {
		t.Fatalf("BenchmarkBar parsed wrong: %+v", res["BenchmarkBar"])
	}
}
