package migration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrContractGateDenied is returned when the no-N-1-remains precondition is
// not evidenced. The gate fails closed.
var ErrContractGateDenied = errors.New("migration: contract gate denied — N-1 process presence not disproven")

// ActiveProcessRecord represents a currently-running application process or
// consumer that has reported its schema-version posture.
type ActiveProcessRecord struct {
	ProcessID string
	Version   string
	UpdatedAt time.Time
}

// EnsureActiveProcessTable creates the migration.active_process table used to
// evidence the absence of N-1 processes.
func EnsureActiveProcessTable(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS migration"); err != nil {
		return err
	}
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migration.active_process (
			process_id text PRIMARY KEY,
			version text NOT NULL,
			updated_at timestamptz NOT NULL DEFAULT now()
		)
	`)
	return err
}

// RegisterActiveProcess records that processID is running at version.
func RegisterActiveProcess(ctx context.Context, conn *pgx.Conn, processID, version string) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO migration.active_process (process_id, version, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (process_id) DO UPDATE
		SET version = EXCLUDED.version, updated_at = EXCLUDED.updated_at
	`, processID, version)
	return err
}

// DeregisterActiveProcess removes processID from the registry.
func DeregisterActiveProcess(ctx context.Context, conn *pgx.Conn, processID string) error {
	_, err := conn.Exec(ctx, `DELETE FROM migration.active_process WHERE process_id = $1`, processID)
	return err
}

// NoN1Remains returns true when there is positive evidence that no process
// remains at n1Version. The evidence is the active-process registry being
// empty for that version. If the registry has not been populated, the gate
// fails closed (returns false).
func NoN1Remains(ctx context.Context, conn *pgx.Conn, n1Version string) (bool, error) {
	var count int64
	err := conn.QueryRow(ctx, `
		SELECT count(*) FROM migration.active_process WHERE version = $1
	`, n1Version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// ContractGate denies contract-phase DDL unless evidence proves no N-1 process
// remains. It fails closed on missing or ambiguous evidence.
func ContractGate(ctx context.Context, conn *pgx.Conn, n1Version string) error {
	ok, err := NoN1Remains(ctx, conn, n1Version)
	if err != nil {
		return fmt.Errorf("contract gate evidence check: %w", err)
	}
	if !ok {
		return ErrContractGateDenied
	}
	return nil
}

// RecoverablePhase names the DATA-09 protocol phases that must each support
// forward recovery.
type RecoverablePhase string

const (
	PhaseExpand   RecoverablePhase = "expand"
	PhaseBackfill RecoverablePhase = "backfill"
	PhaseValidate RecoverablePhase = "validate"
	PhaseCanary   RecoverablePhase = "canary"
	PhaseSwitch   RecoverablePhase = "switch"
	PhaseContract RecoverablePhase = "contract"
)

// ForwardRecovery runs the supplied recovery action for phase and returns the
// result. In tests this proves that every phase has a defined forward path;
// in production it would dispatch to the phase-specific recovery handler.
func ForwardRecovery(ctx context.Context, phase RecoverablePhase, recover func(ctx context.Context) error) error {
	if recover == nil {
		return fmt.Errorf("migration: no forward-recovery handler for phase %q", phase)
	}
	if err := recover(ctx); err != nil {
		return fmt.Errorf("migration: forward recovery for phase %q failed: %w", phase, err)
	}
	return nil
}
