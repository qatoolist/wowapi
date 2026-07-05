package kernel_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationRulesResolverOrgAncestry exercises the org-ancestry closure New
// wires into the rules resolver (it queries the authz store for an org's
// ancestors). With a non-nil org and no rule versions set, resolution walks
// ancestry (running the closure) and falls through to the code default.
func TestIntegrationRulesResolverOrgAncestry(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}

	// Register a rule point so the resolver gets past its registry lookup.
	k.Rules.Register("test", rules.Point{
		Key:         "test.rules.knob",
		ValueSchema: json.RawMessage(`{"type":"integer"}`),
		Default:     json.RawMessage(`7`),
		Description: "coverage knob",
	})
	if err := k.Rules.Err(); err != nil {
		t.Fatalf("register rule point: %v", err)
	}

	tenant := uuid.New()
	org := uuid.New() // non-nil org → the ancestry closure runs
	ctx := database.WithTenantID(context.Background(), tenant)

	var res rules.Resolved
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		r, e := k.RulesResolver.Resolve(ctx, db, "test.rules.knob", org, time.Now())
		if e != nil {
			return e
		}
		res = r
		return nil
	}); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// No versions exist for any scope → the code default is returned.
	if !res.IsDefault {
		t.Fatalf("expected the code default, got %+v", res)
	}
	if string(res.Value) != "7" {
		t.Fatalf("default value = %q, want %q", string(res.Value), "7")
	}
}
