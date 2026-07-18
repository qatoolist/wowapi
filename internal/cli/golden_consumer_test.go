package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const goldenConsumerModulePath = "example.com/wowapi-golden-consumer"

// goldenConsumerFrameworkVersion is the synthetic version the golden-consumer
// proxy serves this checkout as. The content-derived suffix keeps it in
// lock-step with the packaged source (see frameworkSourceSuffix): the go
// module index is path-keyed and immutable-by-assumption, so only a new
// version path gets a fresh index.
func goldenConsumerFrameworkVersion(t *testing.T) string {
	return "v1.2.0-golden-" + frameworkSourceSuffix(t)
}

func goldenConsumerScaffold(t *testing.T) string {
	t.Helper()

	gobin := t.TempDir()
	proxy := buildFrameworkProxy(t, goldenConsumerFrameworkVersion(t))
	goEnv := hermeticGoEnv(proxy + "," + modCacheProxyURL(t))
	install := exec.Command(
		"go", "install", "-buildvcs=false",
		"github.com/qatoolist/wowapi/cmd/wowapi@"+goldenConsumerFrameworkVersion(t),
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
		!strings.Contains(provenance, goldenConsumerFrameworkVersion(t)) {
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
	webhookClosureTest := strings.ReplaceAll(goldenWebhookClosureSource, "MODULE_PATH", goldenConsumerModulePath)
	if err := os.WriteFile(filepath.Join(productDir, "internal", "boottest", "webhook_route_test.go"), []byte(webhookClosureTest), 0o644); err != nil {
		t.Fatalf("write generated webhook closure test: %v", err)
	}

	runPipelineStep(t, "tidy generated two-module consumer", productDir, goEnv,
		"go", "mod", "tidy")
	runPipelineStep(t, "build generated two-module consumer", productDir, goEnv,
		"go", "build", "./...")
	runPipelineStep(t, "boot generated eight-subsystem consumer", productDir, goEnv,
		"go", "test", "./internal/boottest/")
	return productDir
}

const goldenWebhookClosureSource = `package boottest

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/secrets"
	"github.com/qatoolist/wowapi/kernel/storage"
	"github.com/qatoolist/wowapi/testkit"
	"MODULE_PATH/internal/wire"
)

const generatedWebhookSecret = "golden-webhook-secret"

type generatedWebhookSecrets struct{}

func (generatedWebhookSecrets) Resolve(context.Context, secrets.Ref) (string, error) {
	return generatedWebhookSecret, nil
}

func TestGeneratedInboundWebhookRouteIsBooted(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Tx: noopTxM{}, Storage: storage.NewMemory()})
	if err != nil { t.Fatal(err) }
	a := app.New()
	a.Register(wire.Modules()...)
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil { t.Fatal(err) }
	for _, route := range booted.RuntimeRouter().Routes() {
		if route.Method == http.MethodPost && route.Pattern == "/webhooks/fulfillment/shipment_update/{tenant_id}/{endpoint_id}" && route.Meta.Public {
			return
		}
	}
	t.Fatal("generated inbound webhook POST route was not present after boot")
}

func TestGeneratedInboundWebhookRoutePersistsVerifiedEvent(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	endpointID := uuid.New()
	if _, err := h.Admin.Exec(context.Background(), ` + "`" + `INSERT INTO webhook_endpoints
		(id, tenant_id, direction, secret_ref, signature_scheme, status, created_by)
		VALUES ($1,$2,'inbound','secretref://test/webhook','hmac-sha256','active',$3)` + "`" + `,
		endpointID, tenant.ID, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Tx: h.TxM, Storage: storage.NewMemory(), Secrets: generatedWebhookSecrets{},
	})
	if err != nil { t.Fatal(err) }
	a := app.New()
	a.Register(wire.Modules()...)
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil { t.Fatal(err) }

	body := []byte(` + "`" + `{"shipment":"updated"}` + "`" + `)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(generatedWebhookSecret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(body)
	req := httptest.NewRequest(http.MethodPost,
		"/webhooks/fulfillment/shipment_update/"+tenant.ID.String()+"/"+endpointID.String(),
		bytes.NewReader(body))
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	rec := httptest.NewRecorder()
	booted.RuntimeRouter().SecureHandler(nil, nil, h.TxM).ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("generated inbound route = %d: %s", rec.Code, rec.Body.String())
	}
	var status, eventType string
	var signatureOK bool
	if err := h.Admin.QueryRow(context.Background(), ` + "`" + `SELECT delivery_status,event_type,signature_ok
		FROM webhook_events WHERE tenant_id=$1 AND endpoint_id=$2` + "`" + `,
		tenant.ID, endpointID).Scan(&status, &eventType, &signatureOK); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || eventType != "fulfillment.shipment_update" || !signatureOK {
		t.Fatalf("generated inbound event status=%s type=%s signature_ok=%v", status, eventType, signatureOK)
	}
}
`

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
