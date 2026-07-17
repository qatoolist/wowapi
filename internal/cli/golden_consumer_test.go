package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	goldenConsumerModulePath = "example.com/wowapi-golden-consumer"
	// BUMP the numeric suffix whenever the checkout's package SET changes — see
	// e2eReleaseVersion's comment (the go module index is path-keyed and
	// immutable-by-assumption; only a new version path gets a fresh index).
	goldenConsumerFrameworkVersion = "v1.2.0-w06e01s002.12"
)

func goldenConsumerScaffold(t *testing.T) string {
	t.Helper()

	gobin := t.TempDir()
	proxy := buildFrameworkProxy(t, goldenConsumerFrameworkVersion)
	goEnv := hermeticGoEnv(proxy + "," + modCacheProxyURL(t))
	install := exec.Command(
		"go", "install", "-buildvcs=false",
		"github.com/qatoolist/wowapi/cmd/wowapi@"+goldenConsumerFrameworkVersion,
	)
	install.Dir = wowapiCheckoutRoot(t)
	install.Env = append(os.Environ(), goEnv...)
	install.Env = append(install.Env, "GOBIN="+gobin)
	if out, err := install.CombinedOutput(); err != nil {
		t.Fatalf("go install wowapi CLI: %v\n%s", err, out)
	}
	t.Log("pipeline step \"go install versioned CLI\" ok")

	cli := filepath.Join(gobin, "wowapi")
	provenance := runPipelineStep(t, "verify installed CLI provenance", install.Dir, goEnv,
		"go", "version", "-m", cli)
	if !strings.Contains(provenance, "github.com/qatoolist/wowapi/cmd/wowapi") ||
		!strings.Contains(provenance, goldenConsumerFrameworkVersion) {
		t.Fatalf("installed CLI provenance does not name versioned wowapi module:\n%s", provenance)
	}
	productDir := scaffoldPipeline(t, cli, goldenConsumerModulePath, nil, goEnv)

	runPipelineStep(t, "generate catalog module", productDir, nil, cli,
		"new-module", "--name", "catalog")
	runPipelineStep(t, "generate fulfillment module", productDir, nil, cli,
		"new-module", "--name", "fulfillment")
	runPipelineStep(t, "generate catalog CRUD", productDir, nil, cli,
		"gen", "crud",
		"--module", "internal/modules/catalog",
		"--resource", "item",
		"--fields", "name:string,stock:int")
	runPipelineStep(t, "generate fulfillment CRUD", productDir, nil, cli,
		"gen", "crud",
		"--module", "internal/modules/fulfillment",
		"--resource", "shipment",
		"--fields", "reference:string,attempts:int")
	runPipelineStep(t, "generate catalog rule", productDir, nil, cli,
		"gen", "rule", "--module", "internal/modules/catalog", "--name", "stock_limit")
	runPipelineStep(t, "generate catalog workflow", productDir, nil, cli,
		"gen", "workflow", "--module", "internal/modules/catalog", "--name", "item_review")
	runPipelineStep(t, "generate catalog event handler", productDir, nil, cli,
		"gen", "event-handler", "--module", "internal/modules/catalog", "--name", "item_created")
	runPipelineStep(t, "generate fulfillment recurring job", productDir, nil, cli,
		"gen", "recurring-job", "--module", "internal/modules/fulfillment", "--name", "shipment_retry")
	runPipelineStep(t, "generate catalog document flow", productDir, nil, cli,
		"gen", "document-flow", "--module", "internal/modules/catalog", "--name", "item_attachment")
	runPipelineStep(t, "generate fulfillment notification", productDir, nil, cli,
		"gen", "notification", "--module", "internal/modules/fulfillment", "--name", "shipment_ready")
	runPipelineStep(t, "generate fulfillment webhook", productDir, nil, cli,
		"gen", "webhook", "--module", "internal/modules/fulfillment", "--name", "shipment_update")

	runPipelineStep(t, "tidy generated two-module consumer", productDir, goEnv,
		"go", "mod", "tidy")
	runPipelineStep(t, "build generated two-module consumer", productDir, goEnv,
		"go", "build", "./...")
	runPipelineStep(t, "boot generated eight-subsystem consumer", productDir, goEnv,
		"go", "test", "./internal/boottest/")
	return productDir
}

var goldenConsumerRequiredArtifacts = []string{
	"internal/modules/catalog/module.go",
	"internal/modules/catalog/item.go",
	"internal/modules/fulfillment/module.go",
	"internal/modules/fulfillment/shipment.go",
	"internal/modules/catalog/stock_limit_rule.go",
	"internal/modules/catalog/item_review_workflow.go",
	"internal/modules/catalog/item_created_event_handler.go",
	"internal/modules/catalog/item_attachment_document_flow.go",
	"internal/modules/fulfillment/shipment_retry_recurring_job.go",
	"internal/modules/fulfillment/shipment_ready_notification.go",
	"internal/modules/fulfillment/shipment_update_webhook.go",
}

func goldenConsumerArtifactError(productDir string) error {
	for _, rel := range goldenConsumerRequiredArtifacts {
		if _, err := os.Stat(filepath.Join(productDir, rel)); err != nil {
			return fmt.Errorf("installed-binary scaffold missing %s: %w", rel, err)
		}
	}
	return nil
}

// TestGoldenConsumerFailingFixture proves the release gate rejects an
// incomplete consumer rather than passing because the remaining Go code builds.
func TestGoldenConsumerFailingFixture(t *testing.T) {
	productDir := t.TempDir()
	for _, rel := range goldenConsumerRequiredArtifacts[1:] {
		path := filepath.Join(productDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := goldenConsumerArtifactError(productDir); err == nil {
		t.Fatal("deliberately incomplete golden fixture passed the artifact gate")
	}
}

// TestGoldenConsumerInstalledBinaryTwoModules installs the real CLI with
// `go install`, scaffolds a third-party module, generates and boots all eight
// required subsystem types across two modules, and needs no manual
// post-generation edits.
func TestGoldenConsumerInstalledBinaryTwoModules(t *testing.T) {
	productDir := goldenConsumerScaffold(t)

	if err := goldenConsumerArtifactError(productDir); err != nil {
		t.Error(err)
	}

	wireSource, err := os.ReadFile(filepath.Join(productDir, "internal", "wire", "modules.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		goldenConsumerModulePath + "/internal/modules/catalog",
		goldenConsumerModulePath + "/internal/modules/fulfillment",
		"&catalog.Module{}",
		"&fulfillment.Module{}",
	} {
		if !strings.Contains(string(wireSource), want) {
			t.Errorf("generated wire registry missing %q:\n%s", want, wireSource)
		}
	}

	gomod, err := os.ReadFile(filepath.Join(productDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(gomod), "replace github.com/qatoolist/wowapi") {
		t.Fatal("golden consumer must resolve wowapi as a versioned dependency, not a checkout replace")
	}
}

func installGoldenConsumerCLI(t *testing.T, version string) string {
	t.Helper()

	gobin := t.TempDir()
	install := exec.Command(
		"go", "install", "-buildvcs=false",
		"github.com/qatoolist/wowapi/cmd/wowapi@"+version,
	)
	install.Dir = wowapiCheckoutRoot(t)
	install.Env = append(os.Environ(), "GOBIN="+gobin)
	if out, err := install.CombinedOutput(); err != nil {
		t.Fatalf("go install wowapi CLI %s: %v\n%s", version, err, out)
	}

	cli := filepath.Join(gobin, "wowapi")
	provenance := runPipelineStep(t, "verify installed CLI provenance "+version, install.Dir, nil,
		"go", "version", "-m", cli)
	if !strings.Contains(provenance, "github.com/qatoolist/wowapi/cmd/wowapi") ||
		!strings.Contains(provenance, version) {
		t.Fatalf("installed CLI provenance does not name wowapi %s:\n%s", version, provenance)
	}
	primeReleasedModuleCacheProxy(t, version)
	return cli
}

func installGoldenConsumerCandidateCLI(t *testing.T, version string) (string, []string) {
	t.Helper()
	gobin := t.TempDir()
	proxy := buildFrameworkProxy(t, version)
	goEnv := hermeticGoEnv(proxy + "," + modCacheProxyURL(t))
	install := exec.Command("go", "install", "-buildvcs=false",
		"github.com/qatoolist/wowapi/cmd/wowapi@"+version)
	install.Dir = wowapiCheckoutRoot(t)
	install.Env = append(os.Environ(), append(goEnv, "GOBIN="+gobin)...)
	if out, err := install.CombinedOutput(); err != nil {
		t.Fatalf("go install candidate CLI %s: %v\n%s", version, err, out)
	}
	return filepath.Join(gobin, "wowapi"), goEnv
}

func generateGoldenConsumerModules(
	t *testing.T,
	cli string,
	productDir string,
	goEnv []string,
	force bool,
) {
	t.Helper()

	if !force {
		runPipelineStep(t, "generate catalog module", productDir, nil, cli,
			"new-module", "--name", "catalog")
		runPipelineStep(t, "generate fulfillment module", productDir, nil, cli,
			"new-module", "--name", "fulfillment")
	}

	forceArg := []string{}
	if force {
		forceArg = append(forceArg, "--force")
	}
	runPipelineStep(t, "generate catalog CRUD", productDir, nil, cli,
		append([]string{
			"gen", "crud",
			"--module", "internal/modules/catalog",
			"--resource", "item",
			"--fields", "name:string,stock:int",
		}, forceArg...)...)
	runPipelineStep(t, "generate fulfillment CRUD", productDir, nil, cli,
		append([]string{
			"gen", "crud",
			"--module", "internal/modules/fulfillment",
			"--resource", "shipment",
			"--fields", "reference:string,attempts:int",
		}, forceArg...)...)
	if force {
		for _, command := range [][]string{
			{"rule", "--module", "internal/modules/catalog", "--name", "stock_limit", "--force"},
			{"workflow", "--module", "internal/modules/catalog", "--name", "item_review", "--force"},
			{"event-handler", "--module", "internal/modules/catalog", "--name", "item_created", "--force"},
			{"recurring-job", "--module", "internal/modules/fulfillment", "--name", "shipment_retry", "--force"},
			{"document-flow", "--module", "internal/modules/catalog", "--name", "item_attachment", "--force"},
			{"notification", "--module", "internal/modules/fulfillment", "--name", "shipment_ready", "--force"},
			{"webhook", "--module", "internal/modules/fulfillment", "--name", "shipment_update", "--force"},
		} {
			runPipelineStep(t, "regenerate "+command[0], productDir, nil, cli,
				append([]string{"gen"}, command...)...)
		}
	}

	runPipelineStep(t, "tidy generated two-module consumer", productDir, goEnv,
		"go", "mod", "tidy")
}

func assertGoldenConsumerContract(t *testing.T, productDir string, goEnv []string, version string) {
	t.Helper()

	runPipelineStep(t, "build generated two-module consumer at "+version, productDir, goEnv,
		"go", "build", "./...")
	runPipelineStep(t, "boot generated two-module consumer at "+version, productDir, goEnv,
		"go", "test", "./internal/boottest/")

	gomod, err := os.ReadFile(filepath.Join(productDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gomod), "github.com/qatoolist/wowapi "+version) {
		t.Fatalf("golden consumer go.mod does not pin wowapi %s:\n%s", version, gomod)
	}
	if strings.Contains(string(gomod), "replace github.com/qatoolist/wowapi") {
		t.Fatal("upgrade replay must use tagged dependencies, not a checkout replace")
	}
}

// TestGoldenConsumerUpgradeReplay proves the supported v1 N-1/N contract:
// generate and exercise a consumer with the tagged v1.1.0 CLI and dependency,
// upgrade both to the locally packaged release candidate, then rerun the same
// contract checks against the upgraded fixture.
func TestGoldenConsumerUpgradeReplay(t *testing.T) {
	const (
		previous = "v1.1.0"
		current  = goldenConsumerFrameworkVersion
	)

	previousCLI := installGoldenConsumerCLI(t, previous)
	currentCLI, goEnv := installGoldenConsumerCandidateCLI(t, current)

	productDir := scaffoldPipeline(t, previousCLI, goldenConsumerModulePath, nil, goEnv)
	generateGoldenConsumerModules(t, previousCLI, productDir, goEnv, false)
	assertGoldenConsumerContract(t, productDir, goEnv, previous)

	runPipelineStep(t, "upgrade framework dependency N-1 to N", productDir, goEnv,
		"go", "get", "github.com/qatoolist/wowapi@"+current)
	runPipelineStep(t, "upgrade product scaffold N-1 to N", productDir, goEnv,
		currentCLI, "init", "--module", goldenConsumerModulePath, "--dir", ".", "--force")
	generateGoldenConsumerModules(t, currentCLI, productDir, goEnv, true)
	assertGoldenConsumerContract(t, productDir, goEnv, current)
	exerciseGoldenConsumerRealInfrastructure(t, productDir, goEnv)
}
