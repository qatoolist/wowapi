package workflowcontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func root(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate repository root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

func workflow(t *testing.T, name string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root(t), ".github/workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := yaml.Unmarshal(content, &value); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return value
}

func jobs(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	result, ok := value["jobs"].(map[string]any)
	if !ok {
		t.Fatal("workflow jobs is not a mapping")
	}
	return result
}

func TestCIOwnsNativeJobsOnceAndScopesOnlyPullRequests(t *testing.T) {
	ci := workflow(t, "ci.yml")
	ciJobs := jobs(t, ci)
	if _, duplicate := ciJobs["required-gates"]; duplicate {
		t.Fatal("CI must not also execute the release manifest matrix")
	}
	for _, id := range []string{"gate", "gate-bench", "tenantfk-gate", "reference-smoke", "coverage", "golden-consumer"} {
		job, ok := ciJobs[id].(map[string]any)
		if !ok {
			t.Fatalf("missing native CI job %s", id)
		}
		if _, ok := job["needs"]; !ok {
			t.Fatalf("expensive job %s bypasses changes classification", id)
		}
		if _, ok := job["if"]; !ok {
			t.Fatalf("expensive job %s lacks a path-scope condition", id)
		}
	}
	changes := ciJobs["changes"].(map[string]any)
	text, _ := yaml.Marshal(changes)
	if !strings.Contains(string(text), `EVENT_NAME" != "pull_request`) ||
		!strings.Contains(string(text), "code=true") || !strings.Contains(string(text), "bench=true") {
		t.Fatal("non-PR events are not deterministically forced to the full native gate set")
	}
}

func TestReleaseEvidenceDeclarationsHaveOneProducerAndDigestVerifier(t *testing.T) {
	type gate struct {
		ID       string `json:"id"`
		Evidence string `json:"evidence_artifact"`
	}
	var manifest struct {
		Schema int    `json:"schema_version"`
		Gates  []gate `json:"gates"`
	}
	content, err := os.ReadFile(filepath.Join(root(t), "ci/release-gates.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != 2 {
		t.Fatalf("gate manifest schema = %d, want 2", manifest.Schema)
	}
	seen := map[string]bool{}
	for _, item := range manifest.Gates {
		want := "gate-evidence/" + item.ID + ".json"
		if item.Evidence != want || seen[item.ID] {
			t.Fatalf("non-canonical or duplicate evidence declaration for %s", item.ID)
		}
		seen[item.ID] = true
	}
	required, err := os.ReadFile(filepath.Join(root(t), ".github/workflows/required-gates.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(required)
	for _, marker := range []string{"write-gate-evidence", "assemble-gate-results", "gate-bundle", "subject-path: gate-bundle/gate-results.json"} {
		if strings.Count(text, marker) == 0 {
			t.Fatalf("required gate workflow lacks evidence producer/verifier marker %q", marker)
		}
	}
	release, err := os.ReadFile(filepath.Join(root(t), ".github/workflows/release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(release), "verify-gate-bundle") ||
		!strings.Contains(string(release), "source_sha: ${{ github.sha }}") ||
		!strings.Contains(string(release), "release_tag: ${{ github.ref_name }}") {
		t.Fatal("release does not re-verify the complete exact-SHA evidence bundle")
	}
}

func TestAllWorkflowYAMLParses(t *testing.T) {
	entries, err := filepath.Glob(filepath.Join(root(t), ".github/workflows/*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		t.Run(filepath.Base(entry), func(t *testing.T) {
			workflow(t, filepath.Base(entry))
		})
	}
}
