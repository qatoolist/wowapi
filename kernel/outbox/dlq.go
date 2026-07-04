package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// DLQ operability for events (roadmap R4). An event that exhausts its dispatch
// attempts is dead-lettered as dispatch_status='dead'; these functions let an
// operator (`wowapi dlq events …`) list, inspect, replay, and purge them. They
// run on the platform pool (app_platform, which reads/writes events_outbox
// cross-tenant via the outbox_relay_all policy).
//
// Replay is exactly-once-safe: re-dispatched events pass back through the relay,
// and each handler's processed_events inbox dedups by event id, so a replayed
// event cannot double-apply a handler's DB effect.

// DeadEventEntry is a dead-lettered outbox event for inspection.
type DeadEventEntry struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	EventType   string
	Attempts    int
	MaxAttempts int
	LastError   string
	FailedAt    *time.Time
	Payload     []byte
}

// ListDeadEvents returns dead-lettered events, most recently failed first.
func ListDeadEvents(ctx context.Context, pool *pgxpool.Pool, limit int) ([]DeadEventEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := pool.Query(ctx,
		`SELECT id, tenant_id, event_type, attempts, max_attempts, COALESCE(last_error,''), failed_at, payload
		   FROM events_outbox
		  WHERE dispatch_status = 'dead'
		  ORDER BY failed_at DESC NULLS LAST, id DESC
		  LIMIT $1`, limit)
	if err != nil {
		return nil, kerr.Wrapf(err, "outbox.ListDeadEvents", "query dlq")
	}
	defer rows.Close()
	var out []DeadEventEntry
	for rows.Next() {
		var e DeadEventEntry
		if err := rows.Scan(&e.ID, &e.TenantID, &e.EventType, &e.Attempts, &e.MaxAttempts, &e.LastError, &e.FailedAt, &e.Payload); err != nil {
			return nil, kerr.Wrapf(err, "outbox.ListDeadEvents", "scan dlq row")
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "outbox.ListDeadEvents", "iterate dlq")
	}
	return out, nil
}

// ReplayDeadEvent resets a dead event to 'pending' for re-dispatch: attempts
// back to 0, failure/error cleared. Returns KindNotFound if id is not a dead
// event.
func ReplayDeadEvent(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	tag, err := pool.Exec(ctx,
		`UPDATE events_outbox
		    SET dispatch_status = 'pending', attempts = 0, failed_at = NULL, last_error = NULL
		  WHERE id = $1 AND dispatch_status = 'dead'`, id)
	if err != nil {
		return kerr.Wrapf(err, "outbox.ReplayDeadEvent", "replay event %s", id)
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindNotFound, "not_found", "no dead event with that id")
	}
	return nil
}

// DiscardDeadEvent permanently deletes a dead event. Returns KindNotFound if id
// is not a dead event.
func DiscardDeadEvent(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	tag, err := pool.Exec(ctx,
		`DELETE FROM events_outbox WHERE id = $1 AND dispatch_status = 'dead'`, id)
	if err != nil {
		return kerr.Wrapf(err, "outbox.DiscardDeadEvent", "discard event %s", id)
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindNotFound, "not_found", "no dead event with that id")
	}
	return nil
}
