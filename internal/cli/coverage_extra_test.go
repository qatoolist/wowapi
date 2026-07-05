package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------- Run dispatch: exercise every switch arm through the entrypoint ----------

func TestRunDispatchesAllCommands(t *testing.T) {
	// Each command, invoked with no sub-args, returns a usage error (exit 2)
	// without side effects — enough to cover the dispatch arm in Run.
	for _, cmd := range []string{
		"migrate", "seed", "openapi", "lint", "deploy",
		"new-module", "gen", "dlq",
	} {
		code, _, _ := run(t, cmd)
		if code != 2 {
			t.Errorf("Run(%q) = %d, want 2 (usage error)", cmd, code)
		}
	}
	// init with no --module is a usage error too.
	if code, _, _ := run(t, "init"); code != 2 {
		t.Errorf("Run(init) = %d, want 2", code)
	}
}

func TestRunApikeyAuditDispatch(t *testing.T) {
	// apikey/audit reach their handlers (bad tenant → exit 2) via Run.
	if code, _, _ := run(t, "apikey", "list", "--tenant", "bad"); code != 2 {
		t.Errorf("Run(apikey ...) bad tenant = %d, want 2", code)
	}
	if code, _, _ := run(t, "audit", "verify", "--tenant", "bad"); code != 2 {
		t.Errorf("Run(audit ...) bad tenant = %d, want 2", code)
	}
}

// ---------- flag-parse errors: an unknown flag is a usage error (exit 2) ----------

func TestUnknownFlagUsageErrors(t *testing.T) {
	cases := []struct {
		name string
		fn   func(args []string, o, e *bytes.Buffer) int
		args []string
	}{
		{"apikey", func(a []string, o, e *bytes.Buffer) int { return runApikey(a, o, e) }, []string{"list", "--bogus"}},
		{"audit", func(a []string, o, e *bytes.Buffer) int { return runAudit(a, o, e) }, []string{"verify", "--bogus"}},
		{"deploy", func(a []string, o, e *bytes.Buffer) int { return runDeploy(a, o, e) }, []string{"render", "--bogus"}},
		{"gencrud", func(a []string, o, e *bytes.Buffer) int { return runGenCRUD(a, o, e) }, []string{"--bogus"}},
		{"init", func(a []string, o, e *bytes.Buffer) int { return runInit(a, o, e) }, []string{"--bogus"}},
		{"migrate", func(a []string, o, e *bytes.Buffer) int { return runMigrateCreate(a, o, e) }, []string{"--bogus"}},
		{"newmodule", func(a []string, o, e *bytes.Buffer) int { return runNewModule(a, o, e) }, []string{"--bogus"}},
		{"openapi", func(a []string, o, e *bytes.Buffer) int { return runOpenAPI(a, o, e) }, []string{"merge", "--bogus"}},
		{"seed", func(a []string, o, e *bytes.Buffer) int { return runSeedValidate(a, o, e) }, []string{"--bogus"}},
		{"lint", func(a []string, o, e *bytes.Buffer) int { return runLintBoundaries(a, o, e) }, []string{"--bogus"}},
		{"configvalidate", func(a []string, o, e *bytes.Buffer) int { return runConfigValidate(a, o, e) }, []string{"--bogus"}},
		{"configprint", func(a []string, o, e *bytes.Buffer) int { return runConfigPrint(a, o, e) }, []string{"--bogus"}},
		{"configdoctor", func(a []string, o, e *bytes.Buffer) int { return runConfigDoctor(a, o, e) }, []string{"--bogus"}},
		{"configdiff", func(a []string, o, e *bytes.Buffer) int { return runConfigDiff(a, o, e) }, []string{"--bogus"}},
	}
	for _, c := range cases {
		var out, errb bytes.Buffer
		if code := c.fn(c.args, &out, &errb); code != 2 {
			t.Errorf("%s unknown flag = %d, want 2 (stderr=%q)", c.name, code, errb.String())
		}
	}
}

// ---------- DB error paths (need a reachable-but-failing DSN) ----------

// TestApikeyConnectError exercises the pool-connect failure branch: a DSN that
// parses but cannot connect must exit 1 with an error, not panic.
func TestApikeyConnectError(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://nouser:nopass@127.0.0.1:1/none?sslmode=disable")
	var out, errb bytes.Buffer
	if code := runApikey([]string{"list", "--tenant", uuid.New().String()}, &out, &errb); code != 1 {
		t.Fatalf("connect failure should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "wowapi apikey") {
		t.Fatalf("expected apikey error prefix, got %q", errb.String())
	}
}

func TestAuditConnectError(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://nouser:nopass@127.0.0.1:1/none?sslmode=disable")
	var out, errb bytes.Buffer
	if code := runAudit([]string{"verify", "--tenant", uuid.New().String()}, &out, &errb); code != 1 {
		t.Fatalf("connect failure should exit 1, got %d", code)
	}
}

func TestDLQConnectError(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://nouser:nopass@127.0.0.1:1/none?sslmode=disable")
	var out, errb bytes.Buffer
	if code := runDLQ([]string{"jobs", "list"}, &out, &errb); code != 1 {
		t.Fatalf("connect failure should exit 1, got %d", code)
	}
}

// ---------- apikey store-operation errors (valid but absent key id) ----------

func TestApikeyRotateAbsentKeyDB(t *testing.T) {
	requireDSN(t)
	var out, errb bytes.Buffer
	// A well-formed uuid that is not a key of this tenant → store.Rotate errors.
	if code := runApikey([]string{"rotate", "--tenant", uuid.New().String(), "--id", uuid.New().String()}, &out, &errb); code != 1 {
		t.Fatalf("rotate of absent key should exit 1, got %d (stderr=%q)", code, errb.String())
	}
	if !strings.Contains(errb.String(), "wowapi apikey rotate") {
		t.Fatalf("expected rotate error prefix, got %q", errb.String())
	}
}

func TestApikeyRevokeAbsentKeyDB(t *testing.T) {
	requireDSN(t)
	var out, errb bytes.Buffer
	if code := runApikey([]string{"revoke", "--tenant", uuid.New().String(), "--id", uuid.New().String()}, &out, &errb); code != 1 {
		t.Fatalf("revoke of absent key should exit 1, got %d (stderr=%q)", code, errb.String())
	}
}
