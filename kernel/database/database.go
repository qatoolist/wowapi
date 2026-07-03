// Package database is wowapi's persistence kernel: the pgx pool, the
// TxManager that is the ONLY door to tenant data, and the RLS session
// plumbing (SET LOCAL app.tenant_id inside a transaction, never on a pooled
// connection). Contracts in docs/blueprint/05-http-and-persistence.md §2;
// tenant-isolation model in 03 §1.
//
// TenantDB starts as the sqlc DBTX facade and grows the per-tx service
// bundle (Outbox/Audit/Resources) alongside the phases that deliver those
// capabilities (D-0024).
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/config"
)

// DBTX is the query surface sqlc-generated code targets. Both TenantDB and
// DB satisfy it; nothing in module code ever sees a raw pool or connection.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TenantDB is the facade module repositories receive inside a tenant
// transaction. It cannot outlive the tx and cannot be constructed outside
// this package. Per-tx services (Outbox, Audit, Resources) attach here in
// Phases 4/6.
type TenantDB interface {
	DBTX
	tenantSealed()
}

// DB is the platform-scope facade (global tables only). Kernel services
// only; it is never exposed through module.Context.
type DB interface {
	DBTX
	platformSealed()
}

// Option customizes pool construction.
type Option func(*pgxpool.Config)

// WithSetRole makes every pooled connection assume the given role after
// connecting (SET ROLE). This establishes the session-level baseline role so
// that even queries outside a TxManager transaction (e.g. testkit raw probes)
// run RLS-constrained. It is how local/test environments run as app_rt without
// a second login (D-0023); production may instead provision a dedicated login
// in the DSN. It is NOT sufficient on its own — the TxManager re-asserts the
// role per transaction (WithRole) to survive session-state leaks (SEC-11).
func WithSetRole(role string) Option {
	return func(pc *pgxpool.Config) {
		chainAfterConnect(pc, func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, "SET ROLE "+pgx.Identifier{role}.Sanitize())
			return err
		})
	}
}

// WithConnRLSGuard rejects, at connect time, any connection whose effective
// role is a superuser or has BYPASSRLS — such a role silently defeats FORCE
// row level security. Chain it AFTER WithSetRole so it checks the assumed
// role. This backstops the per-transaction guard (Manager's WithRLSGuard) for
// the "over-privileged DSN, no role set" misconfiguration, failing pool
// construction instead of serving tenant traffic with RLS disabled (SEC-12).
func WithConnRLSGuard() Option {
	return func(pc *pgxpool.Config) {
		chainAfterConnect(pc, func(ctx context.Context, conn *pgx.Conn) error {
			var enforced bool
			if err := conn.QueryRow(ctx,
				`SELECT current_setting('is_superuser') = 'off' AND NOT rolbypassrls
                   FROM pg_roles WHERE rolname = current_user`).Scan(&enforced); err != nil {
				return fmt.Errorf("RLS enforcement check: %w", err)
			}
			if !enforced {
				return fmt.Errorf("effective role is superuser or BYPASSRLS; RLS would not be enforced (set a non-privileged runtime role)")
			}
			return nil
		})
	}
}

func chainAfterConnect(pc *pgxpool.Config, step func(context.Context, *pgx.Conn) error) {
	prev := pc.AfterConnect
	pc.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if prev != nil {
			if err := prev(ctx, conn); err != nil {
				return err
			}
		}
		return step(ctx, conn)
	}
}

// NewPool builds the process pool. The DSN arrives as a plain string: the
// composition root (app / cmd) reveals the config Secret — kernel code never
// calls Reveal (boundary lint).
func NewPool(ctx context.Context, dsn string, cfg config.DB, opts ...Option) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		// pgx wraps the DSN (credentials included) into parse errors — never
		// propagate that text.
		return nil, fmt.Errorf("database: invalid DSN (parse failed)")
	}
	pc.MaxConns = int32(cfg.MaxConns)
	for _, o := range opts {
		o(pc)
	}
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return pool, nil
}
