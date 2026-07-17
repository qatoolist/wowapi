package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/config"
)

// TxManager is the only door to the database for tenant work: one
// transaction per unit of work, tenant identity bound with SET LOCAL so RLS
// scopes every statement, automatic rollback on error or panic.
type TxManager interface {
	// WithTenant runs fn in a read-write transaction bound to the tenant in
	// ctx (database.WithTenantID). Missing tenant = ErrNoTenantContext.
	WithTenant(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error
	// WithTenantRO is WithTenant with BEGIN READ ONLY — list/get paths.
	WithTenantRO(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error
	// Platform runs fn against global tables with NO tenant binding. Kernel
	// services only; never exposed through module.Context.
	Platform(ctx context.Context, fn func(ctx context.Context, db DB) error) error
}

// Manager implements TxManager over a pgx pool.
type Manager struct {
	pool     *pgxpool.Pool
	cfg      config.DB
	role     string // re-asserted per tenant tx with SET LOCAL ROLE (SEC-11)
	guardRLS bool   // verify the effective role actually enforces RLS (SEC-12)
}

// ManagerOption customizes a Manager.
type ManagerOption func(*Manager)

// WithRole re-binds the given role at the start of every tenant transaction
// with SET LOCAL ROLE. Unlike a once-per-connection session SET ROLE, this is
// transaction-scoped, so a prior transaction that left the pooled connection
// in a different role (a buggy or hostile module issuing RESET ROLE / SET
// ROLE) cannot leak that state into the next tenant's work — the role is
// re-established from scratch each tx and reverts at COMMIT/ROLLBACK (SEC-11).
func WithRole(role string) ManagerOption {
	return func(m *Manager) { m.role = role }
}

// WithRLSGuard makes every tenant transaction assert that its effective role
// is non-superuser and lacks BYPASSRLS before running caller code. FORCE row
// level security does not apply to superusers or BYPASSRLS roles, so without
// this a pool wired against an over-privileged DSN (or with no role set) would
// silently execute tenant queries with RLS disabled and no signal (SEC-12).
// Deployed processes MUST enable this.
func WithRLSGuard() ManagerOption {
	return func(m *Manager) { m.guardRLS = true }
}

// NewManager wires the manager; cfg travels by value (immutable, 12 §6).
// QueryTimeout out of the validated 100ms..60s band is clamped to the compiled
// default rather than silently disabling the server-side statement ceiling —
// NewManager takes a raw struct and bypasses config.Validate (SEC-14).
func NewManager(pool *pgxpool.Pool, cfg config.DB, opts ...ManagerOption) *Manager {
	if cfg.QueryTimeout <= 0 {
		cfg.QueryTimeout = config.Defaults().DB.QueryTimeout
	}
	m := &Manager{pool: pool, cfg: cfg}
	for _, o := range opts {
		o(m)
	}
	return m
}

var _ TxManager = (*Manager)(nil)

func (m *Manager) WithTenant(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error {
	return m.tenantTx(ctx, pgx.TxOptions{}, fn)
}

func (m *Manager) WithTenantRO(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error {
	return m.tenantTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly}, fn)
}

func (m *Manager) tenantTx(ctx context.Context, opts pgx.TxOptions, fn func(ctx context.Context, db TenantDB) error) error {
	tenantID, ok := TenantIDFrom(ctx)
	if !ok {
		return ErrNoTenantContext // fail closed BEFORE any connection is used
	}
	return m.inTx(ctx, opts, func(tx pgx.Tx) error {
		// Re-assert the RLS role transaction-locally FIRST, so it holds
		// regardless of any state a previous checkout left on this pooled
		// connection (SEC-11). pgx.Identifier.Sanitize quotes the role name.
		if m.role != "" {
			if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+pgx.Identifier{m.role}.Sanitize()); err != nil {
				return fmt.Errorf("database: bind role: %w", err)
			}
		}
		if m.guardRLS {
			var enforced bool
			// A superuser or BYPASSRLS role silently defeats FORCE RLS. Fail
			// closed if the effective role is not RLS-bound (SEC-12).
			if err := tx.QueryRow(ctx,
				`SELECT current_setting('is_superuser') = 'off' AND NOT rolbypassrls
                   FROM pg_roles WHERE rolname = current_user`).Scan(&enforced); err != nil {
				return fmt.Errorf("database: RLS enforcement check failed: %w", err)
			}
			if !enforced {
				return fmt.Errorf("database: refusing tenant transaction — effective role %q is superuser or BYPASSRLS; RLS would not be enforced", m.role)
			}
		}
		// set_config(..., true) = SET LOCAL: scoped to this transaction, so a
		// pooled connection can never leak one tenant's binding to the next
		// checkout (risk R1). Parameterized — no identifier interpolation.
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID.String()); err != nil {
			return fmt.Errorf("database: bind tenant: %w", err)
		}
		if actorID, ok := ActorIDFrom(ctx); ok {
			if _, err := tx.Exec(ctx, "SELECT set_config('app.actor_id', $1, true)", actorID.String()); err != nil {
				return fmt.Errorf("database: bind actor: %w", err)
			}
		}
		return fn(ctx, tenantTx{tx: tx})
	})
}

func (m *Manager) Platform(ctx context.Context, fn func(ctx context.Context, db DB) error) error {
	return m.inTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		return fn(ctx, platformTx{tx: tx})
	})
}

// inTx owns begin/rollback/commit and the server-side statement timeout.
func (m *Manager) inTx(ctx context.Context, opts pgx.TxOptions, fn func(tx pgx.Tx) error) error {
	tx, err := m.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("database: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after commit; covers error AND panic paths

	if m.cfg.QueryTimeout > 0 {
		// Server-side per-statement ceiling; SET LOCAL keeps it tx-scoped.
		// More robust than per-call context juggling (rows must outlive the
		// Query call) and it also bounds statements issued by SQL functions.
		if _, err := tx.Exec(ctx, "SELECT set_config('statement_timeout', $1, true)",
			fmt.Sprintf("%d", m.cfg.QueryTimeout.Milliseconds())); err != nil {
			return fmt.Errorf("database: set statement_timeout: %w", err)
		}
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("database: commit: %w", err)
	}
	return nil
}

// tenantTx seals pgx.Tx behind the TenantDB facade for the duration of fn.
type tenantTx struct{ tx pgx.Tx }

func (t tenantTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, sql, args...)
}

func (t tenantTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

func (t tenantTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}
func (tenantTx) tenantSealed() {}

// platformTx seals pgx.Tx behind the platform DB facade.
type platformTx struct{ tx pgx.Tx }

func (t platformTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, sql, args...)
}

func (t platformTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

func (t platformTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}
func (platformTx) platformSealed() {}
