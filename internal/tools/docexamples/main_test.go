package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/appmodel"
)

func TestExtractExamplesRequiresExactAdjacentMarker(t *testing.T) {
	const markdown = `intro
<!-- doc-example: compile -->
` + "```go" + `
package main
func main() {}
` + "```" + `

<!-- doc-example: illustrative -->
` + "```go" + `
this is intentionally not Go
` + "```" + `
`

	examples, err := extractExamples("guide.md", []byte(markdown))
	if err != nil {
		t.Fatalf("extractExamples: %v", err)
	}
	if len(examples) != 1 {
		t.Fatalf("examples = %d, want 1", len(examples))
	}
	if examples[0].path != "guide.md" || examples[0].line != 4 {
		t.Fatalf("location = %s:%d, want guide.md:4", examples[0].path, examples[0].line)
	}
	if got := string(examples[0].source); got != "package main\nfunc main() {}\n" {
		t.Fatalf("source = %q", got)
	}

	_, err = extractExamples("broken.md", []byte("<!-- doc-example: compile -->\n\n```go\npackage main\n```\n"))
	if err == nil || !strings.Contains(err.Error(), "broken.md:1") || !strings.Contains(err.Error(), "immediately followed") {
		t.Fatalf("non-adjacent marker error = %v", err)
	}

	_, err = extractExamples("unclassified.md", []byte("```go\npackage main\n```\n"))
	if err == nil || !strings.Contains(err.Error(), "unclassified.md:1") || !strings.Contains(err.Error(), "must be classified") {
		t.Fatalf("unclassified fence error = %v", err)
	}
}

func TestCompileExamplesUsesIsolatedThrowawayPackages(t *testing.T) {
	root := repoRoot(t)
	examples := []example{
		{path: "one.md", line: 10, source: []byte("package main\nfunc main() {}\n")},
		{path: "two.md", line: 20, source: []byte("package main\nfunc main() {}\n")},
	}
	if err := compileExamples(context.Background(), root, examples); err != nil {
		t.Fatalf("compileExamples: %v", err)
	}
	for _, name := range []string{"example-001", "example-002"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("go build leaked %s into repository root (stat error %v)", name, err)
		}
	}
}

func TestRemovedSymbolFixtureFailsAtDocumentationLocation(t *testing.T) {
	root := repoRoot(t)
	fixture := filepath.Join(root, "internal", "tools", "docexamples", "testdata", "stale-example.md")
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	examples, err := extractExamples("internal/tools/docexamples/testdata/stale-example.md", data)
	if err != nil {
		t.Fatalf("extract fixture: %v", err)
	}
	if len(examples) != 1 {
		t.Fatalf("examples = %d, want 1", len(examples))
	}
	err = compileExamples(context.Background(), root, examples)
	if err == nil {
		t.Fatal("removed-symbol fixture unexpectedly compiled")
	}
	t.Logf("expected compile failure:\n%s", err)
	for _, want := range []string{"internal/tools/docexamples/testdata/stale-example.md:7", "undefined: app.RunAPI"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q:\n%s", want, err)
		}
	}
	if strings.Contains(err.Error(), ".docexamples-") {
		t.Fatalf("diagnostic contains nondeterministic temporary path:\n%s", err)
	}
}

func TestFutureStateLintRequiresLabelAfterFutureHeading(t *testing.T) {
	unlabeled, err := os.ReadFile("testdata/future-unlabeled.md")
	if err != nil {
		t.Fatal(err)
	}
	err = lintFutureState("testdata/future-unlabeled.md", unlabeled)
	if err == nil {
		t.Fatal("unlabeled future-state fixture unexpectedly passed")
	}
	t.Logf("expected lint failure: %s", err)
	if !strings.Contains(err.Error(), "testdata/future-unlabeled.md:3") ||
		!strings.Contains(err.Error(), "Target, not implemented") {
		t.Fatalf("lint error = %v", err)
	}

	labeled, err := os.ReadFile("testdata/future-labeled.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := lintFutureState("testdata/future-labeled.md", labeled); err != nil {
		t.Fatalf("labeled fixture: %v", err)
	}
}

func TestFutureStateLintIgnoresCodeAndCurrentState(t *testing.T) {
	const markdown = "# Current API\n\nThe client supports retries now.\n\n```text\n## Future API\nwill expose magic\n```\n"
	if err := lintFutureState("current.md", []byte(markdown)); err != nil {
		t.Fatalf("lint current prose: %v", err)
	}
}

func TestGeneratedReferenceByteMatchesAuthoritativeExport(t *testing.T) {
	want := []byte(appmodel.GenerateProjections(canonicalManifest()).Doc + "\n")
	if got := renderReference(); !bytes.Equal(got, want) {
		t.Fatalf("rendered reference does not byte-match ApplicationModel export\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}

	root := repoRoot(t)
	onDisk, err := os.ReadFile(filepath.Join(root, referencePath))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(onDisk, want) {
		t.Fatalf("%s is stale; run go run ./internal/tools/docexamples -write-reference", referencePath)
	}
	if err := checkReference(root); err != nil {
		t.Fatalf("checkReference: %v", err)
	}
}

func TestRepositoryDocumentationPassesAllGates(t *testing.T) {
	report, err := run(context.Background(), repoRoot(t))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if report.examples == 0 {
		t.Fatal("repository documentation gate compiled zero examples")
	}
	if report.futureDocs == 0 {
		t.Fatal("repository documentation gate linted zero future-state documents")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
