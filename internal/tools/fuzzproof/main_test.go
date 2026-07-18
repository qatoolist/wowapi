package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProgressRejectsSeedReplayOnly(t *testing.T) {
	log := "fuzz: elapsed: 0s, gathering baseline coverage: 12/12 completed\nPASS\n"
	if _, err := parseProgress(log); err == nil {
		t.Fatal("parseProgress accepted seed replay without positive elapsed fuzzing")
	}
}

func TestParseProgressAcceptsPositiveElapsedExecutions(t *testing.T) {
	log := "fuzz: elapsed: 0s, gathering baseline coverage: 12/12 completed\n" +
		"fuzz: elapsed: 2s, execs: 191389 (95193/sec), new interesting: 93 (total: 105)\n"
	progress, err := parseProgress(log)
	if err != nil {
		t.Fatalf("parseProgress: %v", err)
	}
	if progress.ElapsedSeconds != 2 || progress.Executions != 191389 {
		t.Fatalf("progress = %+v", progress)
	}
}

func TestParseProgressHandlesMinuteDuration(t *testing.T) {
	log := "fuzz: elapsed: 57s, execs: 100\n" +
		"fuzz: elapsed: 1m0s, execs: 200\n"
	progress, err := parseProgress(log)
	if err != nil {
		t.Fatalf("parseProgress: %v", err)
	}
	if progress.ElapsedSeconds != 60 || progress.Executions != 200 {
		t.Fatalf("progress = %+v, want 60 seconds and 200 executions", progress)
	}
}

func TestCorpusSnapshotCountsRetainedFiles(t *testing.T) {
	cache := t.TempDir()
	corpus := filepath.Join(cache, "fuzz", "example", "FuzzThing")
	if err := os.MkdirAll(corpus, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(corpus, "seed"), []byte("corpus"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := snapshotCorpus(cache)
	if err != nil {
		t.Fatalf("snapshotCorpus: %v", err)
	}
	if snapshot.Files != 1 || snapshot.LatestModTime.IsZero() {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}

// The FuzzParseSort output that flaked CI on 2026-07-18: 10s of positive
// progress, no crash, then a worker-coordination timeout. Must be retryable.
const observedTransientFailure = `=== RUN   FuzzParseSort
fuzz: elapsed: 3s, execs: 104580 (34857/sec), new interesting: 4 (total: 247)
fuzz: elapsed: 9s, execs: 332904 (37091/sec), new interesting: 12 (total: 255)
fuzz: elapsed: 10s, execs: 373791 (37455/sec), new interesting: 14 (total: 257)
--- FAIL: FuzzParseSort (10.10s)
    context deadline exceeded
FAIL`

// A genuine fuzz-discovered crash writes a reproducer. Must NEVER be retried.
const realCrashFailure = `--- FAIL: FuzzParseSort (2.31s)
    parse_test.go:42: mismatch on input
    Failing input written to testdata/fuzz/FuzzParseSort/abc123
    To re-run:
    go test -run=FuzzParseSort/abc123
FAIL`

func TestRetryableFuzzFailure(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   bool
	}{
		{"observed transient timeout", observedTransientFailure, true},
		{"real crash with reproducer", realCrashFailure, false},
		{"panic is a real failure", "some log\npanic: runtime error: index out of range\ngoroutine 1", false},
		{"runtime fatal error is real", "fatal error: concurrent map writes", false},
		{"transient text but crash present is not retried", "context deadline exceeded\nFailing input written to testdata/fuzz/X/y", false},
		{"worker process terminated is transient", "fuzzing process hung or terminated unexpectedly", true},
		{"unclassified failure fails closed", "--- FAIL: FuzzX (1s)\n    assertion failed", false},
		{"clean output has nothing transient to retry", "PASS\nok", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := retryableFuzzFailure(tc.output); got != tc.want {
				t.Fatalf("retryableFuzzFailure(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
