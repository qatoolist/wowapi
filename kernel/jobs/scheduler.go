package jobs

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// Scheduler runs registered maintenance tasks on fixed intervals, leader-safe
// across worker replicas (roadmap E5 + R3). Each task has a row in `schedules`;
// a due tick is claimed by an atomic conditional UPDATE (`next_run_at <= now`
// under FOR UPDATE SKIP LOCKED), so exactly one replica runs a given task per
// interval — no separate leader election. Tasks run OUTSIDE the claim tx, so a
// slow task never holds the row lock; because the claim already advanced
// next_run_at, a failed task simply retries next interval (tasks must be
// idempotent, which the kernel sweeps are).
type Scheduler struct {
	pool  *pgxpool.Pool // app_platform pool that owns the schedules table
	log   *slog.Logger
	tasks []task
	// onRun is invoked after each claimed run with the observed lag (how late the
	// run started vs its scheduled time) — wire it to a metric (R3 "sweeper lag").
	onRun func(name string, lag time.Duration, err error)
}

type task struct {
	name  string
	every time.Duration
	run   func(ctx context.Context) error
}

// NewScheduler builds a scheduler over the platform pool.
func NewScheduler(pool *pgxpool.Pool, log *slog.Logger) *Scheduler {
	if log == nil {
		log = slog.Default()
	}
	return &Scheduler{pool: pool, log: log}
}

// OnRun sets the per-run observer (metric hook). Optional.
func (s *Scheduler) OnRun(fn func(name string, lag time.Duration, err error)) { s.onRun = fn }

// Register adds a recurring task. every is clamped to >= 1s. Call before Run.
func (s *Scheduler) Register(name string, every time.Duration, run func(ctx context.Context) error) {
	if every < time.Second {
		every = time.Second
	}
	s.tasks = append(s.tasks, task{name: name, every: every, run: run})
}

// Run ensures each task's schedule row exists, then polls: on every tick it
// tries to claim and run each due task. It blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context, poll time.Duration) error {
	if poll <= 0 {
		poll = 30 * time.Second
	}
	if err := s.Ensure(ctx); err != nil {
		// A canceled context during the startup upsert means shutdown raced
		// startup — a clean exit, not a scheduler failure. (Otherwise the more
		// schedules registered, the wider this race, e.g. StartWorker on a
		// fast-cancelled worker.)
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	t := time.NewTicker(poll)
	defer t.Stop()
	for {
		s.Tick(ctx)
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}

// Tick attempts every registered task once: it claims each due task (leader-safe)
// and runs the ones this replica won. Exposed so callers can drive the scheduler
// manually (and for tests). Errors are logged, never fatal.
func (s *Scheduler) Tick(ctx context.Context) {
	for _, tk := range s.tasks {
		if ctx.Err() != nil {
			return
		}
		claimed, lag, err := s.claim(ctx, tk)
		if err != nil {
			s.log.WarnContext(ctx, "scheduler: claim failed", "task", tk.name, "err", err)
			continue
		}
		if !claimed {
			continue
		}
		runErr := tk.run(ctx)
		if runErr != nil {
			s.log.WarnContext(ctx, "scheduler: task failed", "task", tk.name, "err", runErr)
		}
		if s.onRun != nil {
			s.onRun(tk.name, lag, runErr)
		}
	}
}

// Ensure upserts a schedule row for every registered task, preserving next_run_at
// for an existing row (so a redeploy does not reset the clock) while updating its
// interval. Run calls it; tests may call it before Tick.
func (s *Scheduler) Ensure(ctx context.Context) error {
	for _, t := range s.tasks {
		if _, err := s.pool.Exec(ctx,
			`INSERT INTO schedules (name, interval_seconds, next_run_at)
			 VALUES ($1, $2, now())
			 ON CONFLICT (name) DO UPDATE SET interval_seconds = EXCLUDED.interval_seconds`,
			t.name, int(t.every.Seconds())); err != nil {
			return kerr.Wrapf(err, "scheduler.Ensure", "upsert schedule %s", t.name)
		}
	}
	return nil
}

// claim atomically claims a due task for THIS replica. It locks the due row with
// SKIP LOCKED (so a replica mid-claim is simply skipped), advances next_run_at,
// and returns the lag (now - the scheduled time). Not due / locked / disabled →
// claimed=false. All in one short transaction.
func (s *Scheduler) claim(ctx context.Context, t task) (claimed bool, lag time.Duration, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, 0, kerr.Wrapf(err, "scheduler.claim", "begin")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var scheduled time.Time
	row := tx.QueryRow(ctx,
		`SELECT next_run_at FROM schedules
		  WHERE name = $1 AND enabled AND next_run_at <= now()
		  FOR UPDATE SKIP LOCKED`, t.name)
	if err := row.Scan(&scheduled); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, 0, nil // not due, or another replica holds it
		}
		return false, 0, kerr.Wrapf(err, "scheduler.claim", "select due %s", t.name)
	}

	var now time.Time
	if err := tx.QueryRow(ctx,
		`UPDATE schedules
		    SET last_run_at = now(), next_run_at = now() + make_interval(secs => interval_seconds)
		  WHERE name = $1
		  RETURNING last_run_at`, t.name).Scan(&now); err != nil {
		return false, 0, kerr.Wrapf(err, "scheduler.claim", "advance %s", t.name)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, 0, kerr.Wrapf(err, "scheduler.claim", "commit %s", t.name)
	}
	lag = max(now.Sub(scheduled), 0)
	return true, lag, nil
}
