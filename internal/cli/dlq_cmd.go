// dlq_cmd.go — wowapi dlq: inspect and operate the dead-letter queues (R4).
package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/jobs"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
)

func dlqUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi dlq <jobs|events> <list|inspect|replay|discard> [args]

Inspect and operate the dead-letter queues. Requires DATABASE_URL; connects as
app_platform (the kernel maintenance role).

  wowapi dlq jobs   list [--limit N]      list discarded jobs
  wowapi dlq jobs   inspect <id>          show one discarded job (full payload)
  wowapi dlq jobs   replay  <id>          requeue a discarded job (attempts reset)
  wowapi dlq jobs   discard <id>          permanently delete a discarded job

  wowapi dlq events list [--limit N]      list dead events
  wowapi dlq events inspect <uuid>        show one dead event (full payload)
  wowapi dlq events replay  <uuid>        re-dispatch a dead event (inbox dedups)
  wowapi dlq events discard <uuid>        permanently delete a dead event

Replay is safe: jobs are at-least-once with idempotent workers; events dedup via
the processed_events inbox.
`)
}

// runDLQ implements `wowapi dlq`.
func runDLQ(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		dlqUsage(stderr)
		return 2
	}
	domain, action := args[0], args[1]
	rest := args[2:]

	// Validate arguments BEFORE touching the database, so bad invocations fail
	// fast (exit 2) without needing DATABASE_URL.
	if domain != "jobs" && domain != "events" {
		fmt.Fprintf(stderr, "wowapi dlq: unknown domain %q (want jobs|events)\n", domain)
		return 2
	}
	switch action {
	case "list", "inspect", "replay", "discard":
	default:
		dlqUsage(stderr)
		return 2
	}
	if action == "inspect" || action == "replay" || action == "discard" {
		var ok bool
		if domain == "jobs" {
			_, ok = parseInt64(rest, stderr)
		} else {
			_, ok = parseUUID(rest, stderr)
		}
		if !ok {
			return 2
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := dlqPool(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi dlq: %v\n", err)
		return 1
	}
	defer pool.Close()

	switch domain {
	case "jobs":
		return dlqJobs(ctx, pool, action, rest, stdout, stderr)
	case "events":
		return dlqEvents(ctx, pool, action, rest, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "wowapi dlq: unknown domain %q (want jobs|events)\n", domain)
		return 2
	}
}

func dlqPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	// Reject a superuser/BYPASSRLS DSN at connect time: app_platform reads
	// cross-tenant via permissive RLS policies, not by bypassing RLS. Mirrors the
	// app_rt guard on the audit/apikey CLIs and the generated api/worker platform pools.
	return database.NewPool(ctx, dsn, config.Defaults().DB,
		database.WithSetRole("app_platform"), database.WithConnRLSGuard())
}

func dlqLimit(rest []string) int {
	fs := flag.NewFlagSet("dlq list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	limit := fs.Int("limit", 50, "max rows")
	_ = fs.Parse(rest)
	return *limit
}

func dlqJobs(ctx context.Context, pool *pgxpool.Pool, action string, rest []string, stdout, stderr io.Writer) int {
	switch action {
	case "list", "inspect":
		entries, err := jobs.ListDead(ctx, pool, dlqLimit(rest))
		if err != nil {
			fmt.Fprintf(stderr, "wowapi dlq: %v\n", err)
			return 1
		}
		if action == "inspect" {
			return dlqInspectJob(entries, rest, stdout, stderr)
		}
		fmt.Fprintf(stdout, "%-8s %-32s %-8s %s\n", "ID", "KIND", "ATTEMPTS", "LAST_ERROR")
		for _, e := range entries {
			fmt.Fprintf(stdout, "%-8d %-32s %d/%-6d %s\n", e.ID, e.Kind, e.Attempts, e.MaxAttempts, oneLine(e.LastError))
		}
		fmt.Fprintf(stdout, "(%d discarded job(s))\n", len(entries))
		return 0
	case "replay", "discard":
		id, ok := parseInt64(rest, stderr)
		if !ok {
			return 2
		}
		var err error
		if action == "replay" {
			err = jobs.ReplayDead(ctx, pool, id)
		} else {
			err = jobs.DiscardDead(ctx, pool, id)
		}
		if err != nil {
			fmt.Fprintf(stderr, "wowapi dlq: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "job %d %sed\n", id, action)
		return 0
	default:
		dlqUsage(stderr)
		return 2
	}
}

func dlqEvents(ctx context.Context, pool *pgxpool.Pool, action string, rest []string, stdout, stderr io.Writer) int {
	switch action {
	case "list", "inspect":
		entries, err := outbox.ListDeadEvents(ctx, pool, dlqLimit(rest))
		if err != nil {
			fmt.Fprintf(stderr, "wowapi dlq: %v\n", err)
			return 1
		}
		if action == "inspect" {
			return dlqInspectEvent(entries, rest, stdout, stderr)
		}
		fmt.Fprintf(stdout, "%-38s %-28s %-8s %s\n", "ID", "EVENT_TYPE", "ATTEMPTS", "LAST_ERROR")
		for _, e := range entries {
			fmt.Fprintf(stdout, "%-38s %-28s %d/%-6d %s\n", e.ID, e.EventType, e.Attempts, e.MaxAttempts, oneLine(e.LastError))
		}
		fmt.Fprintf(stdout, "(%d dead event(s))\n", len(entries))
		return 0
	case "replay", "discard":
		id, ok := parseUUID(rest, stderr)
		if !ok {
			return 2
		}
		var err error
		if action == "replay" {
			err = outbox.ReplayDeadEvent(ctx, pool, id)
		} else {
			err = outbox.DiscardDeadEvent(ctx, pool, id)
		}
		if err != nil {
			fmt.Fprintf(stderr, "wowapi dlq: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "event %s %sed\n", id, action)
		return 0
	default:
		dlqUsage(stderr)
		return 2
	}
}

func dlqInspectJob(entries []jobs.DeadJobEntry, rest []string, stdout, stderr io.Writer) int {
	id, ok := parseInt64(rest, stderr)
	if !ok {
		return 2
	}
	for _, e := range entries {
		if e.ID == id {
			fmt.Fprintf(stdout, "id:          %d\nkind:        %s\nattempts:    %d/%d\nlast_error:  %s\npayload:     %s\n",
				e.ID, e.Kind, e.Attempts, e.MaxAttempts, e.LastError, string(e.Payload))
			return 0
		}
	}
	fmt.Fprintf(stderr, "wowapi dlq: no discarded job with id %d\n", id)
	return 1
}

func dlqInspectEvent(entries []outbox.DeadEventEntry, rest []string, stdout, stderr io.Writer) int {
	id, ok := parseUUID(rest, stderr)
	if !ok {
		return 2
	}
	for _, e := range entries {
		if e.ID == id {
			fmt.Fprintf(stdout, "id:          %s\ntenant:      %s\nevent_type:  %s\nattempts:    %d/%d\nlast_error:  %s\npayload:     %s\n",
				e.ID, e.TenantID, e.EventType, e.Attempts, e.MaxAttempts, e.LastError, string(e.Payload))
			return 0
		}
	}
	fmt.Fprintf(stderr, "wowapi dlq: no dead event with id %s\n", id)
	return 1
}

func parseInt64(rest []string, stderr io.Writer) (int64, bool) {
	if len(rest) < 1 {
		fmt.Fprintln(stderr, "wowapi dlq: an id is required")
		return 0, false
	}
	id, err := strconv.ParseInt(rest[0], 10, 64)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi dlq: %q is not a valid job id\n", rest[0])
		return 0, false
	}
	return id, true
}

func parseUUID(rest []string, stderr io.Writer) (uuid.UUID, bool) {
	if len(rest) < 1 {
		fmt.Fprintln(stderr, "wowapi dlq: an id is required")
		return uuid.Nil, false
	}
	id, err := uuid.Parse(rest[0])
	if err != nil {
		fmt.Fprintf(stderr, "wowapi dlq: %q is not a valid event uuid\n", rest[0])
		return uuid.Nil, false
	}
	return id, true
}

// oneLine collapses a possibly-multiline error into a single bounded line for
// the list view.
func oneLine(s string) string {
	const max = 80
	out := make([]rune, 0, max)
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			r = ' '
		}
		out = append(out, r)
		if len(out) >= max {
			out = append(out, '…')
			break
		}
	}
	return string(out)
}
