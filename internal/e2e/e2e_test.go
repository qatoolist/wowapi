// Package e2e_test is the end-to-end acceptance suite for Phase 12.
//
// It proves:
//   - Acceptance criterion #19: a blank repo scaffolded with `wowapi init`
//     builds working api/worker/migrate binaries.
//   - Acceptance criterion #22: kernel migrations run from cmd/migrate (when
//     DATABASE_URL is set).
//
// The test shells out to the go toolchain and is intentionally slow; run it
// explicitly:
//
//	go test -run TestE2E -count=1 ./internal/e2e/
//
// Requires: go toolchain on PATH. Skips (never fails) when the module cache
// is cold (network-dependent resolution), matching the offline-skip pattern in
// testkit/consumer_test.go.
package e2e_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestE2EScaffoldedRepoBuild is the headline Phase 12 acceptance proof.
// It scaffolds a fresh product repo, replaces the framework dependency with
// the local tree, and verifies the repo builds and vets cleanly.
// With DATABASE_URL set it also runs migrate + polls api /healthz.
func TestE2EScaffoldedRepoBuild(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("e2e: go toolchain not found on PATH")
	}

	repoRoot := findRepoRoot(t)
	tmpDir := t.TempDir()

	// Step 1: build the wowapi CLI from the local tree.
	wowapiBin := filepath.Join(tmpDir, "wowapi")
	if err := runCmd(t, repoRoot, nil, "go", "build", "-o", wowapiBin, "./cmd/wowapi"); err != nil {
		if isOfflineErr(err.Error()) {
			t.Skipf("e2e: CLI build needs network (cold module cache): %v", err)
		}
		t.Fatalf("e2e: build wowapi CLI: %v", err)
	}

	// Step 2: scaffold a product repo.
	productDir := filepath.Join(tmpDir, "product")
	if err := os.MkdirAll(productDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(t, productDir, nil, wowapiBin, "init",
		"--module", "e2e.example/app",
		"--name", "app",
		"--dir", productDir,
	); err != nil {
		t.Fatalf("e2e: wowapi init: %v", err)
	}

	// Step 3: point the product's wowapi dependency at the local tree.
	if err := runCmd(t, productDir, nil, "go", "mod", "edit",
		"-replace", "github.com/qatoolist/wowapi="+repoRoot,
	); err != nil {
		t.Fatalf("e2e: go mod edit -replace: %v", err)
	}

	// Base env for all subsequent go commands in the product dir.
	// APP_ENV=local makes the mains load configs/local.yaml (which sets
	// environment: local and the secretref DSN stubs). WOWAPI__ENVIRONMENT=local
	// also sets the environment directly via the env-var layer (belt+suspenders).
	env := append(os.Environ(),
		"GOFLAGS=-mod=mod",
		"APP_ENV=local",
		"WOWAPI__ENVIRONMENT=local",
	)

	// Step 4: go mod tidy — skip on cold module cache.
	if err := runCmd(t, productDir, env, "go", "mod", "tidy"); err != nil {
		if isOfflineErr(err.Error()) {
			t.Skipf("e2e: go mod tidy needs network (cold module cache): %v", err)
		}
		t.Fatalf("e2e: go mod tidy: %v", err)
	}

	// Step 5: go build ./... — proves api/worker/migrate compile (criterion #19).
	if err := runCmd(t, productDir, env, "go", "build", "./..."); err != nil {
		t.Fatalf("e2e: go build ./...: %v", err)
	}

	// Step 6: go vet ./...
	if err := runCmd(t, productDir, env, "go", "vet", "./..."); err != nil {
		t.Fatalf("e2e: go vet ./...: %v", err)
	}

	// Step 6b: `config validate --env <e>` must honor --env, NOT fall back to
	// APP_ENV. No prod overlay is scaffolded, so validating --env prod must FAIL
	// even though APP_ENV=local would otherwise validate clean — the regression
	// guard for the "validated the wrong environment" bug (the prod CI gate).
	if err := runCmd(t, productDir, env, wowapiBin, "config", "validate", "--env", "prod"); err == nil {
		t.Fatal("e2e: `config validate --env prod` must fail when no prod overlay exists — --env must not fall back to APP_ENV")
	}
	t.Log("e2e: config validate --env prod correctly fails (honors --env ✓)")

	// DB smoke path: guarded by DATABASE_URL so it skips cleanly in offline CI —
	// but a CI/release gate sets WOWAPI_REQUIRE_DB=1 so the runtime proof (migrate
	// + api /healthz) is not silently skipped.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		if os.Getenv("WOWAPI_REQUIRE_DB") != "" {
			t.Fatal("WOWAPI_REQUIRE_DB is set but DATABASE_URL is empty — the E2E runtime smoke (migrate + /healthz) must run in this gate")
		}
		t.Log("e2e: skipping DB smoke — DATABASE_URL not set; build+vet passed (criterion #19 ✓)")
		return
	}

	// DATABASE_URL is resolved by the secretref://env/DATABASE_URL in local.yaml.
	// MIGRATE_URL falls back to DATABASE_URL inside the migrate main. PLATFORM_URL
	// is required since CF-1 (the api/worker fail closed without db.platform_dsn);
	// the gate DSN is the superuser, which can SET ROLE app_platform for the
	// platform pool just as it does SET ROLE app_rt for the runtime pool.
	dbEnv := append(env, "DATABASE_URL="+dsn, "MIGRATE_URL="+dsn, "PLATFORM_URL="+dsn)

	// Step 7: build and run cmd/migrate — criterion #22.
	migrateBin := filepath.Join(tmpDir, "migrate")
	if err := runCmd(t, productDir, dbEnv, "go", "build", "-o", migrateBin, "./cmd/migrate"); err != nil {
		t.Fatalf("e2e: build migrate: %v", err)
	}
	if err := runCmd(t, productDir, dbEnv, migrateBin); err != nil {
		t.Fatalf("e2e: migrate run: %v (criterion #22 FAIL)", err)
	}
	t.Log("e2e: migrate ran successfully (criterion #22 ✓)")

	// `config validate --env local` must pass now that the DSN secretrefs resolve
	// (the product checker wires the env secret provider) — the control for the
	// --env-prod failure asserted above.
	if err := runCmd(t, productDir, dbEnv, wowapiBin, "config", "validate", "--env", "local"); err != nil {
		t.Fatalf("e2e: `config validate --env local` should pass with DSNs set: %v", err)
	}

	// migrate down is a guarded FULL reset: it must refuse outside local/dev even
	// with a valid DSN, so a prod operator cannot wipe the database. Write a valid
	// prod overlay and run under APP_ENV=prod so the config loads cleanly AS prod —
	// then it is the runner's own environment guard (asserted via its message), not
	// a config-conflict error, that rejects the reset, before any DDL runs.
	prodOverlay := filepath.Join(productDir, "configs", "prod.yaml")
	prodYAML := "environment: prod\nlog:\n  level: info\n  format: json\ndb:\n" +
		"  dsn: \"secretref://env/DATABASE_URL\"\n" +
		"  migrate_dsn: \"secretref://env/MIGRATE_URL\"\n" +
		"  platform_dsn: \"secretref://env/PLATFORM_URL\"\n"
	if err := os.WriteFile(prodOverlay, []byte(prodYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(prodOverlay) }()
	// Override BOTH env layers to prod: APP_ENV=prod selects configs/prod.yaml and
	// WOWAPI__ENVIRONMENT=prod overrides the baseline env-var layer (else the two
	// disagree and the config loader rejects the conflict before the guard runs).
	prodEnv := append(append([]string{}, dbEnv...), "APP_ENV=prod", "WOWAPI__ENVIRONMENT=prod")
	derr := runCmd(t, productDir, prodEnv, migrateBin, "down")
	if derr == nil {
		t.Fatal("e2e: `migrate down` must refuse in a non-local/dev environment (guarded reset)")
	}
	if !strings.Contains(derr.Error(), "refusing to reset") {
		t.Fatalf("e2e: `migrate down` in prod must be rejected by the runner's environment guard, got: %v", derr)
	}
	t.Log("e2e: migrate down refuses outside local/dev via the runner guard (guarded reset ✓)")

	// Step 8: build cmd/api and start it on a free port.
	apiBin := filepath.Join(tmpDir, "api")
	if err := runCmd(t, productDir, dbEnv, "go", "build", "-o", apiBin, "./cmd/api"); err != nil {
		t.Fatalf("e2e: build api: %v", err)
	}

	port, err := freePort()
	if err != nil {
		t.Fatalf("e2e: find free port: %v", err)
	}

	// Boot the api from the manifest `wowapi deploy render` ACTUALLY emits — mapped to
	// real DSNs — not a hand-assembled env. This makes the acceptance proof cover the
	// artifact a consumer deploys: if the rendered manifest omits a fail-closed
	// requirement (e.g. db.platform_dsn), the api crash-loops here instead of the gap
	// shipping silently.
	manifestPath := filepath.Join(tmpDir, "app.env")
	if err := runCmd(t, productDir, env, wowapiBin, "deploy", "render", "--format", "env", "--env", "local", "--out", manifestPath); err != nil {
		t.Fatalf("e2e: deploy render: %v", err)
	}
	manifestBytes, rerr := os.ReadFile(manifestPath)
	if rerr != nil {
		t.Fatalf("e2e: read rendered manifest: %v", rerr)
	}
	// Take the manifest's WOWAPI__ config keys (secretref values) verbatim, then
	// supply the referenced secret vars + a listen addr. No DATABASE_URL/PLATFORM_URL
	// hand-injection — the manifest alone must carry every DSN the api requires.
	renderedEnv := append([]string{}, os.Environ()...)
	for line := range strings.SplitSeq(string(manifestBytes), "\n") {
		if strings.HasPrefix(line, "WOWAPI__") {
			renderedEnv = append(renderedEnv, line)
		}
	}
	renderedEnv = append(renderedEnv,
		"WOWAPI_DB_DSN="+dsn,
		"WOWAPI_PLATFORM_DSN="+dsn,
		"WOWAPI_MIGRATE_DSN="+dsn,
		"WOWAPI__HTTP__ADDR=:"+port,
	)
	apiCmd := exec.Command(apiBin)
	apiCmd.Dir = productDir
	apiCmd.Env = renderedEnv
	apiCmd.Stdout = os.Stdout
	apiCmd.Stderr = os.Stderr
	if err := apiCmd.Start(); err != nil {
		t.Fatalf("e2e: start api: %v", err)
	}
	t.Cleanup(func() {
		_ = apiCmd.Process.Signal(os.Interrupt)
		_ = apiCmd.Wait()
	})

	// Step 9: poll /healthz until 200 (criterion #19 runtime proof).
	healthURL := "http://localhost:" + port + "/healthz"
	pollCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := pollHealthz(pollCtx, healthURL); err != nil {
		t.Fatalf("e2e: api /healthz: %v (criterion #19 runtime FAIL)", err)
	}
	t.Logf("e2e: api /healthz 200 OK at %s (criterion #19 runtime ✓)", healthURL)

	// H-1 regression: a rate-limited (429) response must STILL carry the security
	// headers SecureHeaders sets — proving SecureHeaders/CORS are OUTER to RateLimit,
	// which short-circuits without calling inner middleware. Fire a burst past the
	// default capacity and require the first 429 to carry X-Content-Type-Options.
	if err := assert429CarriesSecurityHeaders(healthURL); err != nil {
		t.Fatalf("e2e: %v (middleware order regression)", err)
	}
	t.Log("e2e: rate-limited 429 still carries security headers (middleware order ✓)")
}

// assert429CarriesSecurityHeaders bursts requests past the default rate-limit
// capacity (burst 40 / 20 rps) and verifies the first 429 still carries the
// X-Content-Type-Options header SecureHeaders sets. If SecureHeaders were inner to
// RateLimit (the H-1 bug), the short-circuited 429 would omit it.
func assert429CarriesSecurityHeaders(url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	for range 100 {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		code, hdr := resp.StatusCode, resp.Header.Get("X-Content-Type-Options")
		_ = resp.Body.Close()
		if code == http.StatusTooManyRequests {
			if hdr != "nosniff" {
				return fmt.Errorf("rate-limited 429 is missing X-Content-Type-Options (SecureHeaders is inner to RateLimit): got %q", hdr)
			}
			return nil
		}
	}
	return fmt.Errorf("could not trigger a 429 within 100 requests to exercise the rate-limit path")
}

// findRepoRoot walks up from the test package directory to find the go.mod.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("e2e: cannot locate repo root (no go.mod found walking up from %s)", wd)
		}
		dir = parent
	}
}

// runCmd runs a command and returns a combined error string on failure.
func runCmd(t *testing.T, dir string, env []string, name string, args ...string) error {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return nil
}

// isOfflineErr reports whether the error message indicates a network/proxy
// failure that should trigger a skip rather than a test failure.
func isOfflineErr(msg string) bool {
	for _, marker := range []string{"dial tcp", "lookup ", "proxy", "cannot find module", "no such host"} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// freePort returns a random free TCP port on localhost.
func freePort() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	port := fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()
	return port, nil
}

// pollHealthz polls url until it returns HTTP 200 or ctx expires.
func pollHealthz(ctx context.Context, url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s: %w", url, ctx.Err())
		case <-time.After(200 * time.Millisecond):
			resp, err := client.Get(url)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		}
	}
}
