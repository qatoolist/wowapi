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
