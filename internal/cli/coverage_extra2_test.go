package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------- dlq oneLine truncation (long last_error) ----------

func TestDLQJobsListTruncatesLongErrorDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	long := strings.Repeat("x", 200)
	var id int64
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO jobs_queue (kind, payload, status, attempts, max_attempts, last_error, finished_at)
		 VALUES ('clitest.longerr', '{}', 'discarded', 5, 5, $1, now()) RETURNING id`, long).Scan(&id); err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM jobs_queue WHERE id=$1`, id) })

	var out, errb bytes.Buffer
	if code := runDLQ([]string{"jobs", "list", "--limit", "500"}, &out, &errb); code != 0 {
		t.Fatalf("list exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "…") {
		t.Fatalf("long error should be truncated with an ellipsis: %q", out.String())
	}
}

// ---------- apikey: expired-status branch in list ----------

func TestApikeyListExpiredStatusDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	tenant := uuid.New()
	t.Cleanup(func() { cleanupTenant(t, pool, tenant) })

	var out, errb bytes.Buffer
	if code := runApikey([]string{"issue", "--tenant", tenant.String(), "--name", "exp"}, &out, &errb); code != 0 {
		t.Fatalf("issue exit %d: %s", code, errb.String())
	}
	// Force the key into the past so list reports it as expired.
	execAdmin(t, pool, `UPDATE api_keys SET expires_at = now() - interval '1 day' WHERE tenant_id = $1`, tenant)

	out.Reset()
	errb.Reset()
	if code := runApikey([]string{"list", "--tenant", tenant.String()}, &out, &errb); code != 0 {
		t.Fatalf("list exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "expired") {
		t.Fatalf("expected expired status, got %q", out.String())
	}
}

// ---------- migrate: absent dir takes the IsNotExist path and is created ----------

func TestMigrateCreateAbsentDirCreated(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does", "not", "exist")
	var out, errb bytes.Buffer
	if code := runMigrateCreate([]string{"--dir", dir, "--name", "first"}, &out, &errb); code != 0 {
		t.Fatalf("absent dir should be created, exit %d: %s", code, errb.String())
	}
	if filepath.Base(strings.TrimSpace(out.String())) != "00001_first.sql" {
		t.Fatalf("expected 00001_first.sql, got %q", out.String())
	}
}

// ---------- openapi: --dir pointing at a file surfaces a read error ----------

func TestOpenAPIMergeDirIsFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"merge", "--dir", f}, &out, &errb); code != 1 {
		t.Fatalf("--dir as a file should exit 1, got %d", code)
	}
}

// ---------- gen crud: migrations path is a file → nextMigrationNumber errors ----------

func TestGenCRUDMigrationsPathIsFile(t *testing.T) {
	base := t.TempDir()
	modDir := filepath.Join(base, "widgets")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A regular file named "migrations" makes nextMigrationNumber fail.
	if err := os.WriteFile(filepath.Join(modDir, "migrations"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "widget")
	if code != 1 {
		t.Fatalf("migrations-as-file should exit 1, got %d (%s)", code, errOut)
	}
}

// Empty field parts (trailing/double commas) are skipped, not errors.
func TestGenCRUDEmptyFieldPartsSkipped(t *testing.T) {
	modDir := scaffoldModuleDir(t, "things")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "thing", "--fields", "a:string,,b:int,")
	if code != 0 {
		t.Fatalf("empty field parts should be skipped, exit %d: %s", code, errOut)
	}
	assertParseGo(t, filepath.Join(modDir, "thing.go"))
}

// ---------- init: --dir pointing at a file surfaces a ReadDir error ----------

func TestInitDirIsFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := runInit([]string{"--module", "github.com/acme/app", "--dir", f}, &out, &errb); code != 1 {
		t.Fatalf("--dir as a file should exit 1, got %d", code)
	}
}

// ---------- lint: boundaries arm reached via runLint dispatch ----------

func TestLintDispatchBoundaries(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/clean2\n\ngo 1.26\n")
	writeFile(t, dir, "foo.go", "package clean2\n")
	t.Chdir(dir)
	var out, errb bytes.Buffer
	if code := runLint([]string{"boundaries", "--pkgs", "./..."}, &out, &errb); code != 0 {
		t.Fatalf("lint boundaries dispatch exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "boundary lint: OK") {
		t.Fatalf("expected OK, got %q", out.String())
	}
}

// ---------- lint: lifecycle arm reached via runLint dispatch (backlog B9) ----------

func TestLintDispatchLifecycle(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runLint([]string{"lifecycle"}, &out, &errb); code != 0 {
		t.Fatalf("lint lifecycle dispatch exit %d: %s", code, errb.String())
	}
	got := out.String()
	if !strings.Contains(got, "lifecycle lint: OK") {
		t.Fatalf("expected OK, got %q", got)
	}
	if !strings.Contains(got, "kernel.Tx") {
		t.Fatalf("expected the manifest table to be printed, got %q", got)
	}
}

func TestLintLifecycleUnknownFlag(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runLintLifecycle([]string{"--bogus"}, &out, &errb); code != 2 {
		t.Fatalf("unknown flag should exit 2, got %d: %s", code, errb.String())
	}
}

func TestLintDispatchUnknownSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runLint([]string{"bogus"}, &out, &errb); code != 2 {
		t.Fatalf("unknown lint subcommand should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), `unknown subcommand "bogus"`) {
		t.Fatalf("expected unknown-subcommand error, got %q", errb.String())
	}
}

// ---------- config print/doctor: resolve failure + load error ----------

func TestConfigPrintEnvOverlayMissing(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	code, _, errOut := run(t, "config", "print", "--redacted", "--dir", dir, "--env", "ghost")
	if code != 1 {
		t.Fatalf("missing overlay should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "ghost") {
		t.Fatalf("expected missing-overlay error naming ghost, got %q", errOut)
	}
}

func TestConfigDoctorEnvOverlayMissing(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	code, _, errOut := run(t, "config", "doctor", "--dir", dir, "--env", "ghost")
	if code != 1 {
		t.Fatalf("missing overlay should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "ghost") {
		t.Fatalf("expected missing-overlay error, got %q", errOut)
	}
}

func TestConfigDoctorLoadError(t *testing.T) {
	dir := t.TempDir()
	// Missing environment → LoadDetailed fails, so the table cannot be rendered.
	writeYAML(t, dir, "base.yaml", "log:\n  level: info\n")
	code, _, errOut := run(t, "config", "doctor", "--dir", dir)
	if code != 1 {
		t.Fatalf("load error should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "environment") {
		t.Fatalf("expected environment error, got %q", errOut)
	}
}

// ---------- scaffold: writeEmpty creates parent dirs ----------

func TestWriteEmptyCreatesFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "nested", "deep", "keep")
	if err := writeEmpty(dest, false); err != nil {
		t.Fatalf("writeEmpty: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected file created: %v", err)
	}
}
