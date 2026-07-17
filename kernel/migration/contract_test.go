package migration

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestContractGateAndForwardRecovery proves both required contract-phase
// properties: (1) the contract gate fails closed until evidence shows no N-1
// process remains, and (2) every DATA-09 phase has a defined forward-recovery
// path.
func TestContractGateAndForwardRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test needs Postgres")
	}
	db := testkit.NewDB(t)
	ctx := context.Background()

	admin, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer admin.Release()

	if err := EnsureActiveProcessTable(ctx, admin.Conn()); err != nil {
		t.Fatalf("ensure active process table: %v", err)
	}

	const n1Version = "N-1"
	const n1Process = "worker-1"

	// With an N-1 process registered, the gate must deny contract.
	if err := RegisterActiveProcess(ctx, admin.Conn(), n1Process, n1Version); err != nil {
		t.Fatalf("register N-1 process: %v", err)
	}
	if err := ContractGate(ctx, admin.Conn(), n1Version); !errors.Is(err, ErrContractGateDenied) {
		t.Fatalf("expected ErrContractGateDenied with N-1 active, got %v", err)
	}

	// After the N-1 process exits, the gate passes.
	if err := DeregisterActiveProcess(ctx, admin.Conn(), n1Process); err != nil {
		t.Fatalf("deregister N-1 process: %v", err)
	}
	if err := ContractGate(ctx, admin.Conn(), n1Version); err != nil {
		t.Fatalf("expected gate to pass after N-1 deregistered, got %v", err)
	}

	// Forward recovery: every phase must accept a recovery handler and report
	// success when the handler succeeds.
	phases := []RecoverablePhase{PhaseExpand, PhaseBackfill, PhaseValidate, PhaseCanary, PhaseSwitch, PhaseContract}
	for _, phase := range phases {
		called := false
		if err := ForwardRecovery(ctx, phase, func(ctx context.Context) error {
			called = true
			return nil
		}); err != nil {
			t.Fatalf("forward recovery for %s: %v", phase, err)
		}
		if !called {
			t.Fatalf("forward recovery handler for %s was not called", phase)
		}
	}

	// Forward recovery must surface handler failures.
	if err := ForwardRecovery(ctx, PhaseBackfill, func(ctx context.Context) error {
		return errors.New("simulated recovery failure")
	}); err == nil {
		t.Fatal("expected forward recovery failure to propagate")
	}

	// Missing recovery handler must fail closed.
	if err := ForwardRecovery(ctx, PhaseContract, nil); err == nil {
		t.Fatal("expected missing recovery handler to fail")
	}
}

var _ = pgx.ErrNoRows
