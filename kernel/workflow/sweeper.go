package workflow

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// SweepSLA processes SLA timers for open tasks in the caller's tenant tx and is
// idempotent: reminders are guarded by last_reminded_at, escalations by the
// task transitioning out of 'open'. It is invoked by a registered per-tenant
// job (the lead wires the job; this is the method).
//
//   - reminder: an open task past remind_after that has not been reminded since
//     that time gets a workflow.<def>.reminded event and last_reminded_at = now.
//     Running twice does not double-remind.
//   - escalation: an open task past due_at is marked expired, a
//     workflow.<def>.escalated event is emitted, and if its step declares an
//     escalate_to step an escalation task is created there.
func (rt *Runtime) SweepSLA(ctx context.Context, db database.TenantDB, now time.Time) (reminders, escalations int, err error) {
	now = now.UTC()

	// --- Reminders. The last_reminded_at guard makes this idempotent: once set
	// to >= remind_after, the row no longer qualifies. ---
	remRows, err := db.Query(ctx,
		`SELECT id, instance_id, step_key FROM workflow_tasks
		  WHERE status = 'open' AND remind_after IS NOT NULL AND remind_after <= $1
		    AND (last_reminded_at IS NULL OR last_reminded_at < remind_after)`, now)
	if err != nil {
		return 0, 0, kerr.Wrapf(err, "workflow.SweepSLA", "query reminders")
	}
	type ref struct {
		id       uuid.UUID
		instance uuid.UUID
		step     string
	}
	var toRemind []ref
	for remRows.Next() {
		var r ref
		if e := remRows.Scan(&r.id, &r.instance, &r.step); e != nil {
			remRows.Close()
			return 0, 0, kerr.Wrapf(e, "workflow.SweepSLA", "scan reminder")
		}
		toRemind = append(toRemind, r)
	}
	remRows.Close()
	if e := remRows.Err(); e != nil {
		return 0, 0, kerr.Wrapf(e, "workflow.SweepSLA", "iterate reminders")
	}
	for _, r := range toRemind {
		tag, e := db.Exec(ctx,
			`UPDATE workflow_tasks SET last_reminded_at = $2, updated_at = now()
			  WHERE id = $1 AND status = 'open'
			    AND remind_after <= $2 AND (last_reminded_at IS NULL OR last_reminded_at < remind_after)`,
			r.id, now)
		if e != nil {
			return reminders, escalations, kerr.Wrapf(e, "workflow.SweepSLA", "mark reminded")
		}
		if tag.RowsAffected() == 0 {
			continue // concurrent sweep already handled it — no double reminder
		}
		inst, def, e := rt.loadInstanceAndDef(ctx, db, r.instance)
		if e != nil {
			return reminders, escalations, e
		}
		if e := rt.emit(ctx, db, inst, def, "reminded", map[string]any{
			"instance_id": r.instance.String(), "task_id": r.id.String(), "step": r.step,
		}); e != nil {
			return reminders, escalations, e
		}
		reminders++
	}

	// --- Escalations. Marking the task 'expired' removes it from the open set,
	// so a re-run cannot escalate it twice. ---
	escRows, err := db.Query(ctx,
		`SELECT id, instance_id, step_key FROM workflow_tasks
		  WHERE status = 'open' AND due_at IS NOT NULL AND due_at <= $1`, now)
	if err != nil {
		return reminders, escalations, kerr.Wrapf(err, "workflow.SweepSLA", "query escalations")
	}
	var toEsc []ref
	for escRows.Next() {
		var r ref
		if e := escRows.Scan(&r.id, &r.instance, &r.step); e != nil {
			escRows.Close()
			return reminders, escalations, kerr.Wrapf(e, "workflow.SweepSLA", "scan escalation")
		}
		toEsc = append(toEsc, r)
	}
	escRows.Close()
	if e := escRows.Err(); e != nil {
		return reminders, escalations, kerr.Wrapf(e, "workflow.SweepSLA", "iterate escalations")
	}
	for _, r := range toEsc {
		tag, e := db.Exec(ctx,
			`UPDATE workflow_tasks SET status = 'expired', version = version + 1, updated_at = now()
			  WHERE id = $1 AND status = 'open'`, r.id)
		if e != nil {
			return reminders, escalations, kerr.Wrapf(e, "workflow.SweepSLA", "mark expired")
		}
		if tag.RowsAffected() == 0 {
			continue // already handled
		}
		inst, def, e := rt.loadInstanceAndDef(ctx, db, r.instance)
		if e != nil {
			return reminders, escalations, e
		}
		if e := rt.emit(ctx, db, inst, def, "escalated", map[string]any{
			"instance_id": r.instance.String(), "task_id": r.id.String(), "step": r.step,
		}); e != nil {
			return reminders, escalations, e
		}
		// If the step declares an escalation target, create a task there and move
		// the instance pointer to it.
		step := def.Steps[r.step]
		if step.SLA != nil && step.SLA.EscalateTo != "" {
			target := stripStepPrefix(step.SLA.EscalateTo)
			if _, ok := def.Steps[target]; ok {
				if e := rt.setCurrentStep(ctx, db, &inst, target); e != nil {
					return reminders, escalations, e
				}
				if e := rt.enterStep(ctx, db, &inst, def, target, uuid.Nil); e != nil {
					return reminders, escalations, e
				}
			}
		}
		escalations++
	}

	return reminders, escalations, nil
}

// parseISODuration parses the ISO-8601 duration subset the SLA fields use:
// PnW, PnD, and PnDTnHnMnS combinations (days/hours/minutes/seconds/weeks).
// Months/years are intentionally unsupported (ambiguous length). An empty
// string yields (0, nil).
func parseISODuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	if s[0] != 'P' {
		return 0, fmt.Errorf("workflow: ISO duration must start with P: %q", s)
	}
	body := s[1:]
	datePart, timePart := body, ""
	if i := strings.IndexByte(body, 'T'); i >= 0 {
		datePart, timePart = body[:i], body[i+1:]
	}
	var total time.Duration
	consume := func(part string, units map[byte]time.Duration) error {
		num := ""
		for i := 0; i < len(part); i++ {
			c := part[i]
			if c >= '0' && c <= '9' {
				num += string(c)
				continue
			}
			unit, ok := units[c]
			if !ok {
				return fmt.Errorf("workflow: bad ISO duration unit %q in %q", string(c), s)
			}
			if num == "" {
				return fmt.Errorf("workflow: ISO duration unit %q without a number in %q", string(c), s)
			}
			n, err := strconv.Atoi(num)
			if err != nil {
				return fmt.Errorf("workflow: ISO duration number %q: %w", num, err)
			}
			total += time.Duration(n) * unit
			num = ""
		}
		if num != "" {
			return fmt.Errorf("workflow: trailing number %q in ISO duration %q", num, s)
		}
		return nil
	}
	if err := consume(datePart, map[byte]time.Duration{
		'W': 7 * 24 * time.Hour, 'D': 24 * time.Hour,
	}); err != nil {
		return 0, err
	}
	if err := consume(timePart, map[byte]time.Duration{
		'H': time.Hour, 'M': time.Minute, 'S': time.Second,
	}); err != nil {
		return 0, err
	}
	return total, nil
}
