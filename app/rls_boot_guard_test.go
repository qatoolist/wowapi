package app_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/testkit"
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

// TestIntegrationBootFailsOnRLSBypassPlatformPool extends the M3 backstop to the
// platform pool: it does all cross-tenant kernel work (job runner, outbox relay,
// webhook dispatch) over FORCE-RLS tables and relies on app_platform being a
// non-privileged role served by permissive policies. A superuser/BYPASSRLS platform
// DSN would bypass those policies with no signal, so Boot must refuse it even when
// the runtime pool is correctly non-privileged.
func TestIntegrationBootFailsOnRLSBypassPlatformPool(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Runtime pool is the correct non-privileged app_rt; only the platform pool is
	// the over-privileged superuser owner (h.Admin) — the misconfiguration M3 must catch.
	kBad, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Platform: h.Admin, Tx: h.TxM})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.New().Boot(context.Background(), kBad, nil); err == nil {
		t.Fatal("Boot must fail when the platform pool is a superuser/BYPASSRLS role — RLS would be inert on cross-tenant kernel tables")
	} else if !strings.Contains(err.Error(), "platform pool") || !strings.Contains(err.Error(), "superuser or BYPASSRLS") {
		t.Fatalf("expected a platform-pool RLS-enforcement error, got: %v", err)
	}
}
