package i18n_test

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/i18n"
)

// Registry.Freeze must seal the underlying catalog so request-time reads never
// race a late write: after Freeze, a Catalog.Add is a silent no-op (the
// boot-sealed invariant) while existing lookups keep working.
func TestRegistryFreezeSealsCatalog(t *testing.T) {
	reg := i18n.NewRegistry()
	cat := reg.Catalog()
	cat.Add("en", "kernel.probe", "before")

	reg.Freeze()

	cat.Add("en", "kernel.probe", "after") // must be ignored post-freeze
	if got, _ := cat.Lookup("en", "kernel.probe"); got != "before" {
		t.Fatalf("post-freeze Add mutated the catalog: got %q, want %q", got, "before")
	}
}
