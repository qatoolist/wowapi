package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qatoolist/wowapi/kernel/database"

	"github.com/qatoolist/wowapi/testkit"
)

func TestGoldenConsumerRealInfrastructure(t *testing.T) {
	productDir := goldenConsumerScaffold(t)
	exerciseGoldenConsumerRealInfrastructure(t, productDir, nil)
}

// exerciseGoldenConsumerRealInfrastructure boots the generated API and worker against
// the same Postgres/MinIO/Mailpit/OTel stack used by CI. It proves authenticated CRUD,
// tenant RLS, transactional-outbox delivery, and worker restart recovery end to end.
func exerciseGoldenConsumerRealInfrastructure(t *testing.T, productDir string, goEnv []string) {
	t.Helper()
	if os.Getenv("DATABASE_URL") == "" {
		if os.Getenv("WOWAPI_REQUIRE_DB") == "1" {
			t.Fatal("WOWAPI_REQUIRE_DB=1 but DATABASE_URL is empty")
		}
		t.Skip("golden real-infrastructure contract needs DATABASE_URL")
	}

	h := testkit.NewDB(t)
	dsn := goldenConsumerDatabaseURL(t, h.Name)
	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("MIGRATE_URL", dsn)
	t.Setenv("PLATFORM_URL", dsn)
	t.Setenv("APP_ENV", "local")
	t.Setenv("S3_ACCESS_KEY", "wowapi")
	t.Setenv("S3_SECRET_KEY", "wowapi-local-only")
	t.Setenv("WOWAPI__STORAGE__ENDPOINT", "http://localhost:9000")
	t.Setenv("WOWAPI__STORAGE__BUCKET", "golden-"+strings.ToLower(uuid.NewString()))
	t.Setenv("WOWAPI__STORAGE__ACCESS_KEY", "secretref://env/S3_ACCESS_KEY")
	t.Setenv("WOWAPI__STORAGE__SECRET_KEY", "secretref://env/S3_SECRET_KEY")
	t.Setenv("WOWAPI__STORAGE__CREATE_BUCKET", "true")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")

	assertGoldenService(t, "MinIO", "http://localhost:9000/minio/health/live")
	assertGoldenService(t, "Mailpit", "http://localhost:8025/api/v1/info")
	assertGoldenService(t, "Jaeger", "http://localhost:16686/api/services")

	runPipelineStep(t, "migrate generated consumer database", productDir, goEnv,
		"go", "run", "./cmd/migrate", "up")

	tenantA := testkit.CreateTenant(t, h)
	tenantB := testkit.CreateTenant(t, h)
	tokenA := issueGoldenConsumerKey(t, tenantA.ID, "golden-a")
	tokenB := issueGoldenConsumerKey(t, tenantB.ID, "golden-b")

	apiPort := goldenFreePort(t)
	workerPort := goldenFreePort(t)
	t.Setenv("WOWAPI__HTTP__ADDR", fmt.Sprintf("127.0.0.1:%d", apiPort))
	t.Setenv("WORKER_METRICS_ADDR", fmt.Sprintf("127.0.0.1:%d", workerPort))

	api := startGoldenProcess(t, productDir, goEnv, "api", "go", "run", "./cmd/api")
	defer api.stop(t)
	waitGoldenHTTP(t, fmt.Sprintf("http://127.0.0.1:%d/healthz", apiPort), 30*time.Second)

	base := fmt.Sprintf("http://127.0.0.1:%d", apiPort)
	created := goldenJSONRequest(t, http.MethodPost, base+"/item", tokenA,
		`{"name":"golden","stock":7}`, http.StatusCreated)
	id := goldenResponseID(t, created)
	goldenJSONRequest(t, http.MethodGet, base+"/item/"+id, tokenA, "", http.StatusOK)
	goldenJSONRequest(t, http.MethodPut, base+"/item/"+id, tokenA,
		`{"name":"updated","stock":9}`, http.StatusOK)
	goldenJSONRequest(t, http.MethodGet, base+"/item", tokenA, "", http.StatusOK)

	if status := goldenRequestStatus(t, http.MethodGet, base+"/item/"+id, tokenB, ""); status == http.StatusOK {
		t.Fatal("tenant B read tenant A's generated CRUD row: RLS isolation failed")
	}
	goldenAssertTenantRows(t, h, tenantA.ID, tenantB.ID)

	// The create transaction also emits catalog.item_created. Start the generated
	// worker only after the event is committed: recovery from downtime must drain it.
	worker := startGoldenProcess(t, productDir, goEnv, "worker", "go", "run", "./cmd/worker")
	waitGoldenHTTP(t, fmt.Sprintf("http://127.0.0.1:%d/metrics", workerPort), 30*time.Second)
	waitGoldenEventStatus(t, h, tenantA.ID, 1, "dispatched", 30*time.Second)

	// Stop the worker, commit another event while it is unavailable, then restart.
	// The pending outbox row is the retry/recovery contract: no event is lost, and
	// the restarted worker idempotently dispatches it through the generated handler.
	worker.stop(t)
	goldenJSONRequest(t, http.MethodPost, base+"/item", tokenA,
		`{"name":"after-restart","stock":1}`, http.StatusCreated)
	waitGoldenEventStatus(t, h, tenantA.ID, 1, "pending", 5*time.Second)
	worker = startGoldenProcess(t, productDir, goEnv, "worker-restart", "go", "run", "./cmd/worker")
	defer worker.stop(t)
	waitGoldenHTTP(t, fmt.Sprintf("http://127.0.0.1:%d/metrics", workerPort), 30*time.Second)
	waitGoldenEventStatus(t, h, tenantA.ID, 2, "dispatched", 30*time.Second)

	goldenJSONRequest(t, http.MethodDelete, base+"/item/"+id, tokenA, "", http.StatusNoContent)
}

func goldenConsumerDatabaseURL(t *testing.T, database string) string {
	t.Helper()
	u, err := url.Parse(os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + database
	return u.String()
}

func issueGoldenConsumerKey(t *testing.T, tenant uuid.UUID, name string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	code := runApikey([]string{
		"issue", "--tenant", tenant.String(), "--name", name,
		"--scopes", "catalog.item.create,catalog.item.read,catalog.item.list,catalog.item.update,catalog.item.deactivate",
	}, &out, &errOut)
	if code != 0 {
		t.Fatalf("issue API key: exit %d: %s", code, errOut.String())
	}
	const marker = "token (shown once): "
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(line, marker) {
			return strings.TrimPrefix(line, marker)
		}
	}
	t.Fatalf("API key output omitted one-time token: %q", out.String())
	return ""
}

type goldenProcess struct {
	name   string
	cmd    *exec.Cmd
	cancel context.CancelFunc
	output bytes.Buffer
	waited bool
}

func startGoldenProcess(t *testing.T, dir string, goEnv []string, name string, command ...string) *goldenProcess {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), goEnv...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	p := &goldenProcess{name: name, cmd: cmd, cancel: cancel}
	cmd.Stdout, cmd.Stderr = &p.output, &p.output
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start %s: %v", name, err)
	}
	return p
}

func (p *goldenProcess) stop(t *testing.T) {
	t.Helper()
	if p == nil || p.waited {
		return
	}
	p.cancel()
	if p.cmd.Process != nil {
		_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
	}
	_ = p.cmd.Wait()
	p.waited = true
}

func goldenFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func assertGoldenService(t *testing.T, name, endpoint string) {
	t.Helper()
	resp, err := http.Get(endpoint) // #nosec G107 -- fixed local CI endpoints
	if err != nil {
		t.Fatalf("%s is required for golden-consumer: %v", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		t.Fatalf("%s health %s: %s", name, endpoint, resp.Status)
	}
}

func waitGoldenHTTP(t *testing.T, endpoint string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(endpoint) // #nosec G107 -- loopback endpoint
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode/100 == 2 {
				return
			}
			last = fmt.Errorf("status %s", resp.Status)
		} else {
			last = err
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("wait for %s: %v", endpoint, last)
}

func goldenJSONRequest(t *testing.T, method, endpoint, token, body string, want int) []byte {
	t.Helper()
	status, response := goldenRequest(t, method, endpoint, token, body)
	if status != want {
		t.Fatalf("%s %s: status %d, want %d: %s", method, endpoint, status, want, response)
	}
	return response
}

func goldenRequestStatus(t *testing.T, method, endpoint, token, body string) int {
	t.Helper()
	status, _ := goldenRequest(t, method, endpoint, token, body)
	return status
}

func goldenRequest(t *testing.T, method, endpoint, token, body string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, endpoint, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, payload
}

func goldenResponseID(t *testing.T, payload []byte) string {
	t.Helper()
	var envelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("decode create response: %v: %s", err, payload)
	}
	if _, err := uuid.Parse(envelope.Data.ID); err != nil {
		t.Fatalf("create response id %q is not a UUID: %s", envelope.Data.ID, payload)
	}
	return envelope.Data.ID
}

func goldenAssertTenantRows(t *testing.T, h *testkit.DBHandle, tenantA, tenantB uuid.UUID) {
	t.Helper()
	count := func(tenant uuid.UUID) int {
		var n int
		err := h.TxM.WithTenant(testkit.TenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
			return db.QueryRow(ctx, `SELECT count(*) FROM catalog_item`).Scan(&n)
		})
		if err != nil {
			t.Fatalf("RLS count for %s: %v", tenant, err)
		}
		return n
	}
	if got := count(tenantA); got != 1 {
		t.Fatalf("tenant A catalog_item = %d, want 1", got)
	}
	if got := count(tenantB); got != 0 {
		t.Fatalf("tenant B catalog_item = %d, want 0", got)
	}
}

func waitGoldenEventStatus(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, wantCount int, wantStatus string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var count int
	for time.Now().Before(deadline) {
		if err := h.Admin.QueryRow(context.Background(),
			`SELECT count(*) FROM events_outbox WHERE tenant_id=$1 AND event_type='catalog.item_created' AND dispatch_status=$2`,
			tenant, wantStatus).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count >= wantCount {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("catalog.item_created events in %s = %d, want at least %d", wantStatus, count, wantCount)
}
