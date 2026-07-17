package kernel_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationAuthzCacheWiring is the R1/CA-2 regression: the per-actor authz
// cache is OFF by default (no stale-allow risk unless opted in) and wired — with
// an exposed Invalidate handle — when Deps.AuthzCacheTTL > 0.
func TestIntegrationAuthzCacheWiring(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	off, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	if err != nil {
		t.Fatalf("kernel.New (cache off): %v", err)
	}
	if off.AuthzCache != nil {
		t.Fatal("AuthzCache must be nil when AuthzCacheTTL is zero (opt-in only)")
	}

	on, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Tx: h.TxM, AuthzCacheTTL: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("kernel.New (cache on): %v", err)
	}
	if on.AuthzCache == nil {
		t.Fatal("AuthzCache must be wired when AuthzCacheTTL > 0")
	}
}
