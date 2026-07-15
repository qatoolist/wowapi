package requestbench

import (
	"path/filepath"
	"testing"
)

func TestPublicationPathIsRepositoryRelative(t *testing.T) {
	got, err := publicationPath("perf/results/reference.json")
	if err != nil {
		t.Fatalf("publication path: %v", err)
	}
	wantSuffix := filepath.Join("perf", "results", "reference.json")
	if filepath.IsAbs(got) == false {
		t.Fatalf("publication path %q is not absolute", got)
	}
	if filepath.Clean(got[len(got)-len(wantSuffix):]) != wantSuffix {
		t.Fatalf("publication path %q does not end in %q", got, wantSuffix)
	}
}
