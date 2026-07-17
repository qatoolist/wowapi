package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The W01-E04-S001-T004 (DX-02) generator-output-boots acceptance test.
//
// `wowapi gen crud` output is only correct if the module it produces actually
// BOOTS — string-inspecting the rendered template (TestGenCRUDPermissionKeys)
// cannot catch a permission verb outside the kernel's closed authorization-verb
// set (kernel/authz/registry.go), because that rejection only surfaces via
// Registry.Err() when app.Boot validates the registration graph. This test
// closes that gap end-to-end: scaffold a product (reusing the shared
// buildRenderedProduct primitive), scaffold a module, run `gen crud` into it,
// wire the generated routes exactly as the template's own TODO instructs a
// developer to, declare the permission keys the generator ACTUALLY emitted in
// the module's seed catalog, and boot the product's module set.
//
// Fail-first history (RISK-W01-005 / DX-02): before the `.delete` →
// `.deactivate` template fix, this test fails with the closed-verb-set
// rejection from kernel/authz/registry.go ("permission action %q is not in the
// closed verb set: widgets.widget.delete"). It is a permanent regression guard:
// any future template change (or verb-set change) that makes generated CRUD
// output dead-on-arrival at boot fails here before merge.

// routePermRE extracts RouteMeta permission keys from a generated resource file.
var routePermRE = regexp.MustCompile(`Permission:\s*"([^"]+)"`)

// extractRoutePermissions returns every RouteMeta permission key in the
// generated Go file, in source order. The seed catalog written below is derived
// from these — verbatim, never hardcoded — so the boot proof always tracks what
// the template actually emitted, not what this test wishes it emitted.
func extractRoutePermissions(t *testing.T, goFile string) []string {
	t.Helper()
	src, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatal(err)
	}
	var perms []string
	for _, m := range routePermRE.FindAllStringSubmatch(string(src), -1) {
		perms = append(perms, m[1])
	}
	if len(perms) == 0 {
		t.Fatalf("no RouteMeta permissions found in %s — generator output changed shape?", goFile)
	}
	return perms
}

// TestGenCRUDOutputBoots proves a freshly generated `gen crud` module boots:
// generate → wire → declare seeds → app.Boot, asserting no
// closed-authorization-verb-set rejection (AC-W01-E04-S001-04).
func TestGenCRUDOutputBoots(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles and boots the rendered product against the real framework; skipped in -short")
	}

	// Step 1: scaffold a product wired to THIS framework checkout (shared
	// scaffold primitive — same one the DX-01 T5 harness work builds on).
	dir := buildRenderedProduct(t)

	// Step 2: scaffold a module, exactly as a developer would.
	modParent := filepath.Join(dir, "internal", "modules")
	var nmOut, nmErr bytes.Buffer
	if code := runNewModule([]string{"--name", "widgets", "--dir", modParent}, &nmOut, &nmErr); code != 0 {
		t.Fatalf("wowapi new-module exit %d: %s", code, nmErr.String())
	}
	modDir := filepath.Join(modParent, "widgets")

	// Step 3: generate the CRUD resource under test.
	if code, _, genErr := callGenCRUD(t, "--module", modDir, "--resource", "widget"); code != 0 {
		t.Fatalf("wowapi gen crud exit %d: %s", code, genErr)
	}
	if code, _, genErr := callGenCRUD(t, "--module", modDir, "--resource", "gadget"); code != 0 {
		t.Fatalf("wowapi gen second crud exit %d: %s", code, genErr)
	}
	assertFileExists(t, filepath.Join(modDir, "gadget.go"))
	resourceFile := filepath.Join(modDir, "widget.go")
	assertFileExists(t, resourceFile)

	// Step 4 is automatic: every generated route permission is added to the
	// module seed catalog, so the output boots without hand editing.
	perms := extractRoutePermissions(t, resourceFile)
	seed, err := os.ReadFile(filepath.Join(modDir, "seeds", "permissions.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, permission := range perms {
		if !strings.Contains(string(seed), "key: "+permission) {
			t.Errorf("generated seed catalog missing route permission %q:\n%s", permission, seed)
		}
	}

	// Steps 5-6 are automatic: generated declarations register through the
	// module template, and new-module adds the module to the product wire set.
	moduleGo := filepath.Join(modDir, "module.go")
	moduleSource, err := os.ReadFile(moduleGo)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(moduleSource), "generatedRegistrations") {
		t.Fatal("generated module does not execute generated registrations")
	}
	wireGo := filepath.Join(dir, "internal", "wire", "modules.go")
	wireSource, err := os.ReadFile(wireGo)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"github.com/acme/compiletest/internal/modules/widgets",
		"&widgets.Module{}",
	} {
		if !strings.Contains(string(wireSource), want) {
			t.Fatalf("new-module did not wire %q:\n%s", want, wireSource)
		}
	}

	// Step 7: a product-side boot test — boots the wired module set exactly as
	// the generated binaries do (app.New → Register(wire.Modules()) → Boot).
	// No DB is needed: Boot's registration validation (permission key shape,
	// closed verb set, route-permission declaration) runs before any pool use,
	// and nil pools skip the RLS liveness assertions.
	bootTest := `package boottest

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"

	"github.com/acme/compiletest/internal/wire"
)

// noopTxM satisfies kernel.Deps.Tx (required by kernel.New) without a
// database. Boot's registration validation never opens a transaction, so
// these methods are unreachable in this test.
type noopTxM struct{}

var errNoDB = errors.New("boot test runs without a database")

func (noopTxM) WithTenant(context.Context, func(context.Context, database.TenantDB) error) error {
	return errNoDB
}

func (noopTxM) WithTenantRO(context.Context, func(context.Context, database.TenantDB) error) error {
	return errNoDB
}

func (noopTxM) Platform(context.Context, func(context.Context, database.DB) error) error {
	return errNoDB
}

func TestGeneratedModuleBoots(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Tx: noopTxM{}})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	a.Register(wire.Modules()...)
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("boot: %v", err)
	}
}
`
	bootDir := filepath.Join(dir, "internal", "boottest")
	if err := os.MkdirAll(bootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bootDir, "boot_test.go"), []byte(bootTest), 0o644); err != nil {
		t.Fatal(err)
	}

	// Step 8: boot it. Everything resolves from the module cache primed by
	// buildRenderedProduct's tidy, so this runs offline.
	cmd := exec.Command("go", "test", "./internal/boottest/")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "not in the closed verb set") {
			t.Fatalf("gen crud output is dead-on-arrival at boot: generated permission verb rejected by the kernel's closed authorization-verb set (kernel/authz/registry.go):\n%s", out)
		}
		t.Fatalf("generated module failed to boot (not the closed-verb rejection — investigate separately):\n%s", out)
	}

	// Belt and suspenders: the generated permission verbs must all be in the
	// closed set — the boot above already enforces this, but assert the
	// specific bad key never reappears in generated output.
	for _, p := range perms {
		if strings.HasSuffix(p, ".delete") {
			t.Errorf("generator emitted out-of-set permission verb %q (closed set has no %q)", p, "delete")
		}
	}
}

// TestInitScaffoldConfigValidates (W01-E04-S001 scope-add, conductor-approved
// 2026-07-13): `wowapi config validate` must pass on a PRISTINE scaffold's
// configs. The prebuilt wowapi binary validates config.Framework alone whenever
// the product-local tools/configcheck delegation is unavailable (D-0002 /
// config_delegate.go — e.g. cwd outside the product root, or the checker not
// yet buildable), so every ACTIVE key init writes into configs/*.yaml must be
// framework-schema-valid. Product-owned sections (auth, storage, security,
// i18n, ...) belong in the scaffold as COMMENTED examples — the file's own
// established convention. Fail-first history: before the fix, base.yaml shipped
// an active `i18n:` block whose keys the framework schema rejects, so this
// test failed with `unknown key "i18n.default_locale"` (et al.) on a fresh
// scaffold — the exact defect reported in W01-E04-S002's DEV-03(a).
func TestInitScaffoldConfigValidates(t *testing.T) {
	dir := t.TempDir()
	if code, _, errOut := callInit(t, "--module", "github.com/acme/validatetest", "--dir", dir); code != 0 {
		t.Fatalf("init exit %d: %s", code, errOut)
	}
	// local.yaml's DSNs are secretref://env/<VAR> references; validation
	// resolves them from the process env (no connection is made).
	t.Setenv("DATABASE_URL", "postgres://app_rt:x@localhost:5432/validatetest?sslmode=disable")
	t.Setenv("MIGRATE_URL", "postgres://app_migrate:x@localhost:5432/validatetest?sslmode=disable")
	t.Setenv("PLATFORM_URL", "postgres://app_platform:x@localhost:5432/validatetest?sslmode=disable")

	code, out, errOut := run(t, "config", "validate", "--dir", filepath.Join(dir, "configs"), "--env", "local")
	if code != 0 {
		t.Fatalf("`config validate --env local` must pass on a pristine scaffold (framework-only path); exit %d:\n%s", code, errOut)
	}
	if !strings.Contains(out, "config OK") {
		t.Errorf("stdout missing 'config OK': %q", out)
	}
}
