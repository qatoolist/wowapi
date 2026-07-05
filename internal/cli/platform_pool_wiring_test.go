package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitWiresGuardedPlatformPool is the revert-sensitive guard for D-0085 C-1 and
// the D-0084 platform-pool guards. A delivered-artifact BEHAVIORAL test cannot catch
// these regressions: the test DSN is a superuser, but the generated pools apply
// `SET ROLE app_platform`, which demotes the effective role to a non-privileged one —
// so `AssertRLSEnforced` and `WithConnRLSGuard` both PASS whether or not they are
// wired. Reverting the wiring would leave the e2e green. This structural assertion
// fails the moment the generated api/worker stops (a) building the platform pool with
// WithConnRLSGuard, or (b) wiring it into the kernel (Deps.Platform) so app.Boot's
// RLS-enforcement check actually covers it.
func TestInitWiresGuardedPlatformPool(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/app", "--dir", dir)
	if code != 0 {
		t.Fatalf("init exit %d: %s", code, errOut)
	}
	apiMain := filepath.Join(dir, "cmd", "api", "main.go")
	workerMain := filepath.Join(dir, "cmd", "worker", "main.go")

	// (a) The platform pool must carry WithConnRLSGuard close after the SET ROLE
	// app_platform (comments between are fine) — reject a superuser/BYPASSRLS platform
	// DSN at connect. The bounded window keeps the guard tied to this pool.
	guardedPlatform := `WithSetRole\("app_platform"\),[\s\S]{0,300}?database\.WithConnRLSGuard\(\)`
	assertFileMatches(t, apiMain, guardedPlatform)
	assertFileMatches(t, workerMain, guardedPlatform)

	// (b) The platform pool must be wired into the kernel (Deps.Platform) so app.Boot's
	// M3 RLS check reaches it. This is the C-1 regression guard: the api previously
	// built its platform pool AFTER Boot and never wired it, making the check dead code.
	wiredPlatform := `kernel\.Deps\{[^}]*Platform:\s*platformPool`
	assertFileMatches(t, apiMain, wiredPlatform)
	assertFileMatches(t, workerMain, wiredPlatform)
}

// TestCLIPoolsApplyRLSGuard locks the connect-time RLS-bypass guard onto every
// framework CLI DB pool. Same masking problem as above (SET ROLE demotes the test
// superuser), so this asserts the guard is present at the source level; it fails the
// moment WithConnRLSGuard is dropped from dlq (app_platform) or audit/apikey (app_rt).
func TestCLIPoolsApplyRLSGuard(t *testing.T) {
	for _, f := range []string{"dlq_cmd.go", "audit_cmd.go", "apikey_cmd.go"} {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if !strings.Contains(string(src), "WithConnRLSGuard()") {
			t.Errorf("%s: its DB pool must apply database.WithConnRLSGuard() (connect-time RLS-bypass guard); a superuser/BYPASSRLS DSN would otherwise run cross-tenant queries with RLS inert", f)
		}
	}
}
