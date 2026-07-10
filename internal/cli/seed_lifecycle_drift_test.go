// seed_lifecycle_drift_test.go — B4: a drift guard comparing the standalone
// `wowapi seed sync` CLI path against the generated product migrate main's
// lifecycle (internal/cli/templates/init/cmd_migrate_main.go.tmpl), the
// production path GAP-003/GAP-007 built. The two paths are intentionally
// asymmetric (the CLI has no product config or rule registry to draw on) —
// this test locks that asymmetry so a future change to either side that
// silently narrows/widens the gap without updating the other's disclosure
// gets caught here instead of at deploy time.
package cli

import (
	"os"
	"strings"
	"testing"
)

const migrateTemplatePath = "templates/init/cmd_migrate_main.go.tmpl"

// TestGeneratedMigrateTemplateRunsFullLifecycle proves the generated migrate
// main still does what the docs/help text claim it does on the CLI's behalf:
// load the composed product config via appcfg.Load() (configs/<env>.yaml +
// secretref resolution), sync seeds, and sync rule definitions — in that
// order. If this template ever drops one of these steps, the CLI's escape-
// hatch framing ("the generated migrate does X, Y, Z; this command only does
// Y") becomes a lie, and this test catches it.
func TestGeneratedMigrateTemplateRunsFullLifecycle(t *testing.T) {
	body, err := os.ReadFile(migrateTemplatePath)
	if err != nil {
		t.Fatalf("read migrate template: %v", err)
	}
	src := string(body)

	mustContain := []string{
		"appcfg.Load()",          // product config + secretref resolution (not bare DATABASE_URL)
		"seeds.Sync(",            // GAP-003 seed catalog sync
		"rules.SyncDefinitions(", // GAP-007 rule definition sync
	}
	for _, want := range mustContain {
		if !strings.Contains(src, want) {
			t.Fatalf("generated migrate template no longer contains %q — CLI escape-hatch\n"+
				"disclosure (seedUsage) would become inaccurate; template:\n%s", want, src)
		}
	}

	// Order matters for the lifecycle contract (blueprint 06 §2 / GAP-007
	// mirrors GAP-003's position exactly): within func run(), the config load
	// (via loadConfig(), a thin wrapper around appcfg.Load() defined later in
	// the file) happens before seeds.Sync, which happens before
	// rules.SyncDefinitions. Scoped to func run(...)'s body specifically,
	// since loadConfig's own definition (containing the literal
	// "appcfg.Load()" call) is declared textually AFTER run() in the file.
	runStart := strings.Index(src, "func run(ctx context.Context")
	runEnd := strings.Index(src, "\n// loadConfig reads")
	if runStart < 0 || runEnd < 0 || runEnd < runStart {
		t.Fatalf("could not locate func run(...) body in template (markers moved?)")
	}
	runBody := src[runStart:runEnd]

	loadIdx := strings.Index(runBody, "loadConfig()")
	seedIdx := strings.Index(runBody, "seeds.Sync(")
	rulesIdx := strings.Index(runBody, "rules.SyncDefinitions(")
	if loadIdx < 0 || seedIdx < 0 || rulesIdx < 0 {
		t.Fatalf("expected loadConfig()/seeds.Sync()/rules.SyncDefinitions() all within func run(...), got indices %d,%d,%d",
			loadIdx, seedIdx, rulesIdx)
	}
	if loadIdx >= seedIdx || seedIdx >= rulesIdx {
		t.Fatalf("expected lifecycle order loadConfig() < seeds.Sync() < rules.SyncDefinitions() within func run(...), got indices %d,%d,%d",
			loadIdx, seedIdx, rulesIdx)
	}
}

// TestStandaloneSeedSyncDoesNotClaimRuleSync is the negative half of the
// drift guard: the standalone CLI path (seed_cmd.go) must NOT import
// kernel/rules or call rules.SyncDefinitions — it has no product rule
// registry to source definitions from (kernel/rules.SyncDefinitions takes a
// *rules.Registry populated only by product Go code via module Register(),
// which a framework-only binary never loads). Asserted via the import line
// specifically (rather than a plain substring search) so the many
// documentation/help-text mentions of "rules.SyncDefinitions" this file
// deliberately carries (seedUsage, comments) don't trip the guard — only an
// actual `"github.com/qatoolist/wowapi/kernel/rules"` import would let
// seed_cmd.go call the real function. If this ever changes (e.g. a
// framework-kernel-only rule registry becomes syncable), this test — and the
// seedUsage/warning text it pairs with — must be updated together.
func TestStandaloneSeedSyncDoesNotClaimRuleSync(t *testing.T) {
	body, err := os.ReadFile("seed_cmd.go")
	if err != nil {
		t.Fatalf("read seed_cmd.go: %v", err)
	}
	src := string(body)
	if strings.Contains(src, `"github.com/qatoolist/wowapi/kernel/rules"`) {
		t.Fatalf("seed_cmd.go now imports kernel/rules — update seedUsage()'s " +
			"disclosure and the drift test's positive assertions to match reality " +
			"(the CLI would need a source of rules.Point declarations, which today only " +
			"a booted product process has)")
	}
}
