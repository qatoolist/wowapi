package app_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationBootFailsOnRLSBypassRuntimePool is the M3 backstop: app.Boot must
// refuse to start when the runtime pool runs as a superuser/BYPASSRLS role — FORCE
// RLS does not apply to such roles, so every tenant query would silently run
// unfiltered. This makes RLS enforcement safe-by-default even if a product forgets
// to wire the per-connection (WithConnRLSGuard) / per-tx (WithRLSGuard) guards.
func TestIntegrationBootFailsOnRLSBypassRuntimePool(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	// h.Admin connects as the DB owner (a superuser) with no SET ROLE — exactly the
	// "over-privileged runtime DSN" misconfiguration M3 warns about.
	kBad, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Admin, Platform: h.Platform, Tx: h.TxM})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.New().Boot(context.Background(), kBad, nil); err == nil {
		t.Fatal("Boot must fail when the runtime pool is a superuser/BYPASSRLS role — RLS would be inert")
	} else if !strings.Contains(err.Error(), "superuser or BYPASSRLS") {
		t.Fatalf("expected an RLS-enforcement error, got: %v", err)
	}

	// Control: the non-privileged app_rt runtime pool boots cleanly.
	kGood, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.New().Boot(context.Background(), kGood, nil); err != nil {
		t.Fatalf("Boot with an app_rt runtime pool should succeed: %v", err)
	}
}
