package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type target struct {
	Package string `json:"package"`
	Name    string `json:"name"`
}

var targets = []target{
	{Package: "./kernel/filtering", Name: "FuzzFilterParse"},
	{Package: "./kernel/filtering", Name: "FuzzParseSort"},
	{Package: "./kernel/pagination", Name: "FuzzDecodeCursor"},
}

type progress struct {
	ElapsedSeconds float64 `json:"elapsed_seconds"`
	Executions     int64   `json:"executions"`
}

type corpusSnapshot struct {
	Files         int       `json:"files"`
	LatestModTime time.Time `json:"latest_mod_time,omitempty"`
}

type targetResult struct {
	Target   target   `json:"target"`
	Progress progress `json:"progress"`
	Log      string   `json:"log"`
}

type report struct {
	Profile              string         `json:"profile"`
	FuzzTime             string         `json:"fuzz_time"`
	StartedAt            time.Time      `json:"started_at"`
	FinishedAt           time.Time      `json:"finished_at"`
	WallElapsedSeconds   float64        `json:"wall_elapsed_seconds"`
	CorpusBefore         corpusSnapshot `json:"corpus_before"`
	CorpusAfter          corpusSnapshot `json:"corpus_after"`
	RetainedFromPriorRun bool           `json:"retained_from_prior_run"`
	Targets              []targetResult `json:"targets"`
}

var elapsedPattern = regexp.MustCompile(`fuzz: elapsed: ([0-9]+m[0-9]+(?:\.[0-9]+)?s|[0-9]+(?:\.[0-9]+)?s), execs: ([0-9]+)`) // positive progress lines only

func main() {
	profile := flag.String("profile", "pr", "fuzz profile: pr or scheduled")
	fuzzTime := flag.String("fuzztime", "10s", "per-target Go fuzz duration")
	cache := flag.String("cache", ".fuzzcache/go-build", "persistent GOCACHE path; its fuzz subtree is retained by CI")
	output := flag.String("output", "fuzz-report", "directory for logs and report.json")
	expectRetained := flag.Bool("expect-retained", false, "require corpus restored from an earlier run")
	flag.Parse()

	if *profile != "pr" && *profile != "scheduled" {
		fatalf("profile must be pr or scheduled, got %q", *profile)
	}
	duration, err := time.ParseDuration(*fuzzTime)
	if err != nil || duration <= 0 {
		fatalf("invalid fuzztime %q", *fuzzTime)
	}
	cacheAbs, err := filepath.Abs(*cache)
	if err != nil {
		fatalf("resolve cache path: %v", err)
	}
	if err := os.MkdirAll(cacheAbs, 0o750); err != nil {
		fatalf("create fuzz cache: %v", err)
	}
	if err := os.MkdirAll(*output, 0o750); err != nil {
		fatalf("create output directory: %v", err)
	}

	before, err := snapshotCorpus(cacheAbs)
	if err != nil {
		fatalf("snapshot retained corpus: %v", err)
	}
	if *expectRetained && before.Files == 0 {
		fatalf("expected a retained fuzz corpus, but %s/fuzz is empty", cacheAbs)
	}

	started := time.Now().UTC()
	results := make([]targetResult, 0, len(targets))
	for _, item := range targets {
		result, err := runTarget(item, *fuzzTime, cacheAbs, *output)
		if err != nil {
			fatalf("%s/%s: %v", item.Package, item.Name, err)
		}
		results = append(results, result)
	}
	finished := time.Now().UTC()
	after, err := snapshotCorpus(cacheAbs)
	if err != nil {
		fatalf("snapshot generated corpus: %v", err)
	}
	if after.Files == 0 {
		fatalf("coverage-guided fuzzing completed but produced no retained GOCACHE/fuzz corpus files")
	}

	summary := report{
		Profile:              *profile,
		FuzzTime:             *fuzzTime,
		StartedAt:            started,
		FinishedAt:           finished,
		WallElapsedSeconds:   finished.Sub(started).Seconds(),
		CorpusBefore:         before,
		CorpusAfter:          after,
		RetainedFromPriorRun: before.Files > 0,
		Targets:              results,
	}
	encoded, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		fatalf("encode report: %v", err)
	}
	reportPath := filepath.Join(*output, "report.json")
	if err := os.WriteFile(reportPath, append(encoded, '\n'), 0o600); err != nil {
		fatalf("write report: %v", err)
	}
	fmt.Printf("PASS: %s fuzz profile ran %d coverage-guided targets for positive elapsed time; corpus files before=%d after=%d; report=%s\n", *profile, len(results), before.Files, after.Files, reportPath)
}

// maxFuzzAttempts bounds retries for a target whose `go test -fuzz` run fails
// with a TRANSIENT Go-fuzzing worker error (a coordinator/worker RPC timeout —
// "context deadline exceeded" — under CI CPU starvation, not a discovered bug).
// A real crash writes a reproducer to the package corpus and is never retried.
const maxFuzzAttempts = 3

// crashSignature marks a genuine fuzz-discovered failure: Go writes the failing
// input to the package's testdata/fuzz corpus and prints "Failing input written
// to ...", and/or the target panics / hits a runtime fatal error. These are real
// bugs — they must NEVER be retried or masked.
var crashSignature = regexp.MustCompile(`(?m)Failing input written to|^\s*panic:|^\s*fatal error:|test execution corrupted`)

// transientSignature marks a fuzzing-infrastructure failure (worker coordination
// under load), safe to retry because it reflects the runner, not the code. Kept
// deliberately narrow so unclassified failures fall through to fail-closed.
var transientSignature = regexp.MustCompile(`context deadline exceeded|fuzzing process (?:hung or )?terminated|communicating with fuzzing process`)

// retryableFuzzFailure reports whether a failed run is a transient worker
// timeout rather than a real crash. It fails closed: a crash signature (or any
// output that does not positively match a known transient pattern) is treated
// as a genuine failure and never retried.
func retryableFuzzFailure(output string) bool {
	if crashSignature.MatchString(output) {
		return false
	}
	return transientSignature.MatchString(output)
}

func runTarget(item target, fuzzTime, cache, output string) (targetResult, error) {
	logPath := filepath.Join(output, item.Name+".txt")
	var lastErr error
	for attempt := 1; attempt <= maxFuzzAttempts; attempt++ {
		result, combined, err := runFuzzOnce(item, fuzzTime, cache, logPath)
		if err == nil {
			return result, nil
		}
		if !retryableFuzzFailure(combined) {
			// Real crash (reproducer written) or an unclassified error: fail
			// hard and surface it — never retry, never mask a discovered bug.
			return targetResult{}, fmt.Errorf("go fuzz failed: %w", err)
		}
		lastErr = err
		fmt.Fprintf(os.Stderr, "fuzz proof: %s/%s transient fuzzing-worker failure on attempt %d/%d (%v); retrying\n",
			item.Package, item.Name, attempt, maxFuzzAttempts, err)
	}
	return targetResult{}, fmt.Errorf("go fuzz failed after %d attempts with only transient worker timeouts (no crash found): %w", maxFuzzAttempts, lastErr)
}

func runFuzzOnce(item target, fuzzTime, cache, logPath string) (targetResult, string, error) {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- fuzz tool writes the selected output path by design
	if err != nil {
		return targetResult{}, "", err
	}
	defer func() { _ = logFile.Close() }()

	var captured bytes.Buffer
	command := exec.CommandContext(context.Background(), "go", "test", item.Package, "-run", "^$", "-fuzz", "^"+item.Name+"$", "-fuzztime", fuzzTime, "-v") // #nosec G204 -- package and fuzz target come from the tool's fixed target table
	command.Env = replaceEnv(os.Environ(), "GOCACHE", cache)
	command.Stdout = io.MultiWriter(os.Stdout, logFile, &captured)
	command.Stderr = io.MultiWriter(os.Stderr, logFile, &captured)
	if err := command.Run(); err != nil {
		return targetResult{}, captured.String(), fmt.Errorf("go fuzz failed: %w", err)
	}
	observed, err := parseProgress(captured.String())
	if err != nil {
		return targetResult{}, captured.String(), err
	}
	return targetResult{Target: item, Progress: observed, Log: logPath}, captured.String(), nil
}

func parseProgress(output string) (progress, error) {
	matches := elapsedPattern.FindAllStringSubmatch(output, -1)
	var result progress
	for _, match := range matches {
		elapsed, err := time.ParseDuration(match[1])
		if err != nil {
			return progress{}, err
		}
		executions, err := strconv.ParseInt(match[2], 10, 64)
		if err != nil {
			return progress{}, err
		}
		if elapsed.Seconds() > result.ElapsedSeconds {
			result.ElapsedSeconds = elapsed.Seconds()
			result.Executions = executions
		}
	}
	if result.ElapsedSeconds <= 0 || result.Executions <= 0 {
		return progress{}, errors.New("output proves seed replay only; no positive elapsed coverage-guided executions found")
	}
	return result, nil
}

func snapshotCorpus(cache string) (corpusSnapshot, error) {
	root := filepath.Join(cache, "fuzz")
	var snapshot corpusSnapshot
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		snapshot.Files++
		if info.ModTime().After(snapshot.LatestModTime) {
			snapshot.LatestModTime = info.ModTime().UTC()
		}
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return snapshot, nil
	}
	return snapshot, err
}

func replaceEnv(environment []string, name, value string) []string {
	prefix := name + "="
	result := make([]string, 0, len(environment)+1)
	for _, item := range environment {
		if !strings.HasPrefix(item, prefix) {
			result = append(result, item)
		}
	}
	return append(result, prefix+value)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fuzz proof: "+format+"\n", args...)
	os.Exit(1)
}
