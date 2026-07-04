package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// DLQ operability (roadmap R4). Jobs that exhaust their attempts are
// dead-lettered as status='discarded'; these functions let an operator (via
// `wowapi dlq jobs …`) list, inspect, replay, and purge them. They run on the
// platform pool (app_platform) since jobs_queue is a global kernel table.
//
// Replay is safe by construction: job delivery is at-least-once and workers must
// be idempotent (see the Worker contract), so re-running a discarded job cannot
// double-apply a correct worker's effect.

// DeadJob is a dead-lettered (discarded) job row for inspection.
type DeadJobEntry struct {
	ID          int64
	Kind        string
	TenantID    *uuid.UUID // nil for a global job
	Attempts    int
	MaxAttempts int
	LastError   string
	FinishedAt  *time.Time
	Payload     []byte
}

// ListDead returns dead-lettered jobs, most recently failed first, up to limit.
func ListDead(ctx context.Context, pool *pgxpool.Pool, limit int) ([]DeadJobEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := pool.Query(ctx,
		`SELECT id, kind, tenant_id, attempts, max_attempts, COALESCE(last_error,''), finished_at, payload
		   FROM jobs_queue
		  WHERE status = 'discarded'
		  ORDER BY finished_at DESC NULLS LAST, id DESC
		  LIMIT $1`, limit)
	if err != nil {
		return nil, kerr.Wrapf(err, "jobs.ListDead", "query dlq")
	}
	defer rows.Close()
	var out []DeadJobEntry
	for rows.Next() {
		var e DeadJobEntry
		if err := rows.Scan(&e.ID, &e.Kind, &e.TenantID, &e.Attempts, &e.MaxAttempts, &e.LastError, &e.FinishedAt, &e.Payload); err != nil {
			return nil, kerr.Wrapf(err, "jobs.ListDead", "scan dlq row")
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "jobs.ListDead", "iterate dlq")
	}
	return out, nil
}

// ReplayDead resets a discarded job to 'available' for another run: attempts
// back to 0, run_at now, error/lock cleared. Returns KindNotFound if id is not a
// discarded job (already replayed, running, or never existed).
func ReplayDead(ctx context.Context, pool *pgxpool.Pool, id int64) error {
	tag, err := pool.Exec(ctx,
		`UPDATE jobs_queue
		    SET status = 'available', attempts = 0, run_at = now(),
		        locked_at = NULL, last_error = NULL, finished_at = NULL
		  WHERE id = $1 AND status = 'discarded'`, id)
	if err != nil {
		return kerr.Wrapf(err, "jobs.ReplayDead", "replay job %d", id)
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindNotFound, "not_found", "no discarded job with that id")
	}
	return nil
}

// DiscardDead permanently deletes a discarded job from the queue. Returns
// KindNotFound if id is not a discarded job.
func DiscardDead(ctx context.Context, pool *pgxpool.Pool, id int64) error {
	tag, err := pool.Exec(ctx,
		`DELETE FROM jobs_queue WHERE id = $1 AND status = 'discarded'`, id)
	if err != nil {
		return kerr.Wrapf(err, "jobs.DiscardDead", "discard job %d", id)
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindNotFound, "not_found", "no discarded job with that id")
	}
	return nil
}
