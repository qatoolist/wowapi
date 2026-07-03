// benchbudget enforces performance budgets against go test -bench output.
//
// Usage (Makefile pipe form — preferred):
//
//	go test -bench=. -benchmem -run=^$ ./... | go run ./internal/tools/benchbudget bench-budgets.txt
//
// The tool reads benchmark output from stdin and a budget file from the first
// argument. It exits non-zero and prints a report of every benchmark that
// exceeds its budget. Benchmarks in the budget file that do not appear in the
// input emit a warning (not a failure) — the test binary may not have been run
// with -bench matching that name.
//
// Budget file format (bench-budgets.txt at repo root):
//
//	# comment lines and blank lines are ignored
//	BenchmarkName  max_ns_per_op  max_allocs_per_op
//
// BenchmarkName must match the base name (no -N suffix). max_allocs_per_op of
// -1 means "unchecked" (useful during initial measurement).
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: benchbudget <budget-file>  (bench output read from stdin)")
		os.Exit(2)
	}
	budgetFile := os.Args[1]

	budgets, err := loadBudgets(budgetFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "benchbudget: load budgets: %v\n", err)
		os.Exit(2)
	}

	results, err := parseBenchOutput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "benchbudget: parse bench output: %v\n", err)
		os.Exit(2)
	}

	var violations []string
	for name, budget := range budgets {
		res, ok := results[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "WARN benchbudget: %s is budgeted but not found in bench output\n", name)
			continue
		}
		var v []string
		if budget.maxNsPerOp >= 0 && res.nsPerOp > float64(budget.maxNsPerOp) {
			v = append(v, fmt.Sprintf("ns/op %.1f > budget %d", res.nsPerOp, budget.maxNsPerOp))
		}
		if budget.maxAllocsPerOp >= 0 && res.allocsPerOp > int64(budget.maxAllocsPerOp) {
			v = append(v, fmt.Sprintf("allocs/op %d > budget %d", res.allocsPerOp, budget.maxAllocsPerOp))
		}
		if len(v) > 0 {
			violations = append(violations, fmt.Sprintf("FAIL  %-55s  %s", name, strings.Join(v, ", ")))
		} else {
			fmt.Printf("OK    %-55s  %.1f ns/op  %d allocs/op\n", name, res.nsPerOp, res.allocsPerOp)
		}
	}

	if len(violations) > 0 {
		fmt.Fprintf(os.Stderr, "\nbenchbudget: %d budget violation(s):\n", len(violations))
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}
		os.Exit(1)
	}
}

// budget is the allowable ceiling for one benchmark.
type budget struct {
	maxNsPerOp     int64 // -1 = unchecked
	maxAllocsPerOp int64 // -1 = unchecked
}

// result is one parsed benchmark result from go test output.
type result struct {
	nsPerOp     float64
	allocsPerOp int64
}

func loadBudgets(path string) (map[string]budget, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	budgets := make(map[string]budget)
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("line %d: want 3 fields (name max_ns_per_op max_allocs_per_op), got %d: %q", lineNo, len(fields), line)
		}
		name := fields[0]
		ns, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: max_ns_per_op %q: %v", lineNo, fields[1], err)
		}
		allocs, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: max_allocs_per_op %q: %v", lineNo, fields[2], err)
		}
		budgets[name] = budget{maxNsPerOp: ns, maxAllocsPerOp: allocs}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return budgets, nil
}

// parseBenchOutput reads go test -benchmem output from r and extracts results.
// Format: BenchmarkName-N\tcount\tN ns/op\tN B/op\tN allocs/op
// Some fields (B/op) may be absent if -benchmem is not passed, but we only
// require ns/op (always present) and allocs/op (present with -benchmem).
func parseBenchOutput(r *os.File) (map[string]result, error) {
	results := make(map[string]result)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		// Tab-separated fields:
		// BenchmarkFoo-N <tab> iterations <tab> N ns/op [<tab> N B/op <tab> N allocs/op]
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		// Base name: strip the -N GOMAXPROCS suffix.
		fullName := parts[0]
		name := baseName(fullName)

		res := result{nsPerOp: -1, allocsPerOp: -1}
		// Walk pairs: "<value> <unit>"
		for i := 2; i+1 < len(parts); i += 2 {
			unit := parts[i+1]
			val := parts[i]
			switch unit {
			case "ns/op":
				f, err := strconv.ParseFloat(val, 64)
				if err == nil {
					res.nsPerOp = f
				}
			case "allocs/op":
				n, err := strconv.ParseInt(val, 10, 64)
				if err == nil {
					res.allocsPerOp = n
				}
			}
		}
		if res.nsPerOp < 0 {
			continue // not a valid bench line
		}
		if res.allocsPerOp < 0 {
			res.allocsPerOp = 0 // -benchmem not used; treat as 0
		}
		// Keep the worst (highest) result if the same name appears multiple times.
		if prev, ok := results[name]; ok {
			if res.nsPerOp > prev.nsPerOp {
				results[name] = res
			}
		} else {
			results[name] = res
		}
	}
	return results, sc.Err()
}

// baseName strips the -N GOMAXPROCS suffix from a benchmark name.
// "BenchmarkFoo-16" → "BenchmarkFoo"
func baseName(name string) string {
	if i := strings.LastIndex(name, "-"); i > 0 {
		suffix := name[i+1:]
		if _, err := strconv.Atoi(suffix); err == nil {
			return name[:i]
		}
	}
	return name
}
