package cli

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------- dlq: missing DSN (no DB needed) ----------

func TestDLQMissingDSN(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	var out, errb bytes.Buffer
	if code := runDLQ([]string{"jobs", "list"}, &out, &errb); code != 1 {
		t.Fatalf("missing DATABASE_URL should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "DATABASE_URL is not set") {
		t.Fatalf("expected DSN error, got %q", errb.String())
	}
}

// insertDeadJob inserts a discarded job with a multiline last_error (to exercise
// oneLine) and returns its id. It is removed on cleanup.
func insertDeadJob(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO jobs_queue (kind, payload, status, attempts, max_attempts, last_error, finished_at)
		 VALUES ('clitest.deadjob', '{"marker":true}', 'discarded', 5, 5, 'line one'||chr(10)||'line two', now())
		 RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatalf("insert dead job: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM jobs_queue WHERE id = $1`, id)
	})
	return id
}

// insertDeadEvent inserts a dead outbox event and returns its id.
func insertDeadEvent(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	execAdmin(t, pool,
		`INSERT INTO events_outbox (id, tenant_id, event_type, payload, dispatch_status, attempts, max_attempts, last_error, failed_at, created_by)
		 VALUES ($1, $2, 'clitest.deadevent', '{}', 'dead', 10, 10, 'event boom', now(), $3)`,
		id, uuid.New(), uuid.New())
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM events_outbox WHERE id = $1`, id)
	})
	return id
}

// ---------- dlq jobs ----------

func TestDLQJobsListInspectDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	id := insertDeadJob(t, pool)

	var out, errb bytes.Buffer
	if code := runDLQ([]string{"jobs", "list", "--limit", "500"}, &out, &errb); code != 0 {
		t.Fatalf("jobs list exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "clitest.deadjob") {
		t.Fatalf("jobs list missing our dead job: %q", out.String())
	}
	// oneLine must collapse the multiline error onto one line.
	if strings.Contains(out.String(), "line one\nline two") {
		t.Fatalf("last_error should be collapsed to one line: %q", out.String())
	}

	// inspect the specific id → full payload.
	out.Reset()
	errb.Reset()
	if code := runDLQ([]string{"jobs", "inspect", strconv.FormatInt(id, 10), "--limit", "500"}, &out, &errb); code != 0 {
		t.Fatalf("jobs inspect exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), `"marker"`) || !strings.Contains(out.String(), "payload:") {
		t.Fatalf("inspect should show payload, got %q", out.String())
	}
}

func TestDLQJobsInspectNotFoundDB(t *testing.T) {
	requireDSN(t)
	var out, errb bytes.Buffer
	// A job id that (almost certainly) is not a discarded job.
	code := runDLQ([]string{"jobs", "inspect", "-999999"}, &out, &errb)
	if code != 1 {
		t.Fatalf("inspect of missing job should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "no discarded job with id") {
		t.Fatalf("expected not-found message, got %q", errb.String())
	}
}

func TestDLQJobsReplayThenDiscardDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	id := insertDeadJob(t, pool)
	idStr := strconv.FormatInt(id, 10)

	// replay: discarded → available.
	var out, errb bytes.Buffer
	if code := runDLQ([]string{"jobs", "replay", idStr}, &out, &errb); code != 0 {
		t.Fatalf("jobs replay exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), fmt.Sprintf("job %d replayed", id)) {
		t.Fatalf("replay output unexpected: %q", out.String())
	}
	var status string
	if err := pool.QueryRow(context.Background(), `SELECT status FROM jobs_queue WHERE id = $1`, id).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != "available" {
		t.Fatalf("replayed job status = %q, want available", status)
	}

	// A replayed (no longer discarded) job cannot be discarded → not found (exit 1).
	out.Reset()
	errb.Reset()
	if code := runDLQ([]string{"jobs", "discard", idStr}, &out, &errb); code != 1 {
		t.Fatalf("discard of non-discarded job should exit 1, got %d", code)
	}

	// Put it back to discarded and discard it for real.
	execAdmin(t, pool, `UPDATE jobs_queue SET status = 'discarded' WHERE id = $1`, id)
	out.Reset()
	errb.Reset()
	if code := runDLQ([]string{"jobs", "discard", idStr}, &out, &errb); code != 0 {
		t.Fatalf("jobs discard exit %d: %s", code, errb.String())
	}
	var n int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM jobs_queue WHERE id = $1`, id).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("discarded job should be deleted, count = %d", n)
	}
}

// ---------- dlq events ----------

func TestDLQEventsListInspectDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	id := insertDeadEvent(t, pool)

	var out, errb bytes.Buffer
	if code := runDLQ([]string{"events", "list", "--limit", "500"}, &out, &errb); code != 0 {
		t.Fatalf("events list exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "clitest.deadevent") {
		t.Fatalf("events list missing our dead event: %q", out.String())
	}

	out.Reset()
	errb.Reset()
	if code := runDLQ([]string{"events", "inspect", id.String(), "--limit", "500"}, &out, &errb); code != 0 {
		t.Fatalf("events inspect exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), id.String()) || !strings.Contains(out.String(), "clitest.deadevent") {
		t.Fatalf("inspect should show event detail, got %q", out.String())
	}
}

func TestDLQEventsInspectNotFoundDB(t *testing.T) {
	requireDSN(t)
	var out, errb bytes.Buffer
	code := runDLQ([]string{"events", "inspect", uuid.New().String()}, &out, &errb)
	if code != 1 {
		t.Fatalf("inspect of missing event should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "no dead event with id") {
		t.Fatalf("expected not-found message, got %q", errb.String())
	}
}

func TestDLQEventsReplayThenDiscardDB(t *testing.T) {
	dsn := requireDSN(t)
	pool := adminPool(t, dsn)
	id := insertDeadEvent(t, pool)

	var out, errb bytes.Buffer
	if code := runDLQ([]string{"events", "replay", id.String()}, &out, &errb); code != 0 {
		t.Fatalf("events replay exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "replayed") {
		t.Fatalf("replay output unexpected: %q", out.String())
	}
	var status string
	if err := pool.QueryRow(context.Background(), `SELECT dispatch_status FROM events_outbox WHERE id = $1`, id).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != "pending" {
		t.Fatalf("replayed event status = %q, want pending", status)
	}

	// Restore to dead and discard.
	execAdmin(t, pool, `UPDATE events_outbox SET dispatch_status = 'dead' WHERE id = $1`, id)
	out.Reset()
	errb.Reset()
	if code := runDLQ([]string{"events", "discard", id.String()}, &out, &errb); code != 0 {
		t.Fatalf("events discard exit %d: %s", code, errb.String())
	}
	var n int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM events_outbox WHERE id = $1`, id).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("discarded event should be deleted, count = %d", n)
	}
}
