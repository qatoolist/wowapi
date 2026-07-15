package migration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func timeoutSQL(name string, d time.Duration) string {
	return fmt.Sprintf("SET %s = '%s'", name, d.String())
}

// DefaultLockBudget is the DATA-09 online DDL lock budget.
const DefaultLockBudget = 2 * time.Second

// DefaultRetryCeiling is the maximum number of retries after a lock-timeout
// abort. The total number of attempts is 1 + DefaultRetryCeiling.
const DefaultRetryCeiling = 3

// ErrLockTimeout is returned when a DDL statement exhausts every attempt
// without acquiring its lock within the budget.
var ErrLockTimeout = errors.New("migration: DDL lock-timeout budget exhausted")

// ExecDDL executes a DDL statement with a bounded lock-timeout budget.
// It sets session lock_timeout and statement_timeout, retries on lock-timeout
// aborts (SQLSTATE 55P03), and leaves the session timeouts cleared on return.
// The statement itself is atomic, so a lock-timeout abort never applies a
// partial schema change.
func ExecDDL(ctx context.Context, conn *pgx.Conn, stmt string, budget time.Duration, maxRetries int) error {
	if budget <= 0 {
		budget = DefaultLockBudget
	}
	if maxRetries < 0 {
		maxRetries = DefaultRetryCeiling
	}

	logger := slog.With("stmt", stmt, "budget_ms", budget.Milliseconds())

	// Ensure timeouts are cleared when we leave so the pooled connection is not
	// left in a constrained state.
	defer func() {
		_, _ = conn.Exec(ctx, "SET lock_timeout = DEFAULT")
		_, _ = conn.Exec(ctx, "SET statement_timeout = DEFAULT")
	}()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := setTimeouts(ctx, conn, budget); err != nil {
			return fmt.Errorf("migration: set timeouts: %w", err)
		}
		_, err := conn.Exec(ctx, stmt)
		if err == nil {
			if attempt > 0 {
				logger.Info("DDL succeeded after retry", "attempt", attempt+1)
			}
			return nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "55P03" {
			logger.Info("DDL lock-timeout abort", "attempt", attempt+1, "max_attempts", maxRetries+1)
			if attempt == maxRetries {
				return fmt.Errorf("%w after %d attempt(s)", ErrLockTimeout, attempt+1)
			}
			// Bounded backoff: 50ms, 100ms, 150ms... short because the budget
			// itself is the primary pacing control.
			select {
			case <-time.After(time.Duration(attempt+1) * 50 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		return fmt.Errorf("migration: DDL failed: %w", err)
	}
	return ErrLockTimeout
}

func setTimeouts(ctx context.Context, conn *pgx.Conn, budget time.Duration) error {
	if _, err := conn.Exec(ctx, timeoutSQL("lock_timeout", budget)); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, timeoutSQL("statement_timeout", 10*budget)); err != nil {
		return err
	}
	return nil
}
