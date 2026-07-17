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
	"github.com/qatoolist/wowapi/kernel/observability"
)

const sweepSLABatchSize = 100

var workflowSLAMetricLabels = map[string]string{"worker": "workflow_sla"}

type slaRef struct {
	id       uuid.UUID
	instance uuid.UUID
	step     string
	dueAt    time.Time
}

type slaState struct {
	instance Instance
	def      Definition
}

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
	if err := rt.requireValidated(); err != nil {
		return 0, 0, err
	}
	started := time.Now()
	maxLag := time.Duration(0)
	defer func() {
		rt.metrics.SetGauge("worker_queue_lag_seconds", maxLag.Seconds(), workflowSLAMetricLabels)
		observability.ObserveHistogram(rt.metrics, "worker_batch_duration_seconds", time.Since(started).Seconds(), workflowSLAMetricLabels)
	}()
	now = now.UTC()

	// Claim and guard-flip one bounded reminder batch atomically. SKIP LOCKED
	// lets another invocation make progress while preserving no-double-remind.
	toRemind, err := claimReminderBatch(ctx, db, now)
	if err != nil {
		return 0, 0, err
	}
	reminderState, err := rt.loadSLAState(ctx, db, toRemind)
	if err != nil {
		return 0, 0, err
	}
	for _, ref := range toRemind {
		if lag := now.Sub(ref.dueAt); lag > maxLag {
			maxLag = lag
		}
		state := reminderState[ref.instance]
		if err := rt.emit(ctx, db, state.instance, state.def, "reminded", map[string]any{
			"instance_id": ref.instance.String(), "task_id": ref.id.String(), "step": ref.step,
		}); err != nil {
			return reminders, escalations, err
		}
		reminders++
	}

	// Expiry is the escalation guard. As above, the state transition and
	// bounded claim are one statement, so concurrent invocations cannot emit
	// the same escalation.
	toEscalate, err := claimEscalationBatch(ctx, db, now)
	if err != nil {
		return reminders, escalations, err
	}
	escalationState, err := rt.loadSLAState(ctx, db, toEscalate)
	if err != nil {
		return reminders, escalations, err
	}
	for _, ref := range toEscalate {
		if lag := now.Sub(ref.dueAt); lag > maxLag {
			maxLag = lag
		}
		state := escalationState[ref.instance]
		inst, def := state.instance, state.def
		if err := rt.emit(ctx, db, inst, def, "escalated", map[string]any{
			"instance_id": ref.instance.String(), "task_id": ref.id.String(), "step": ref.step,
		}); err != nil {
			return reminders, escalations, err
		}
		step := def.Steps[ref.step]
		if step.SLA != nil && step.SLA.EscalateTo != "" {
			target := stripStepPrefix(step.SLA.EscalateTo)
			if _, ok := def.Steps[target]; ok {
				if err := rt.setCurrentStep(ctx, db, &inst, target); err != nil {
					return reminders, escalations, err
				}
				if err := rt.enterStep(ctx, db, &inst, def, target, uuid.Nil); err != nil {
					return reminders, escalations, err
				}
			}
		}
		escalations++
	}
	return reminders, escalations, nil
}

func claimReminderBatch(ctx context.Context, db database.TenantDB, now time.Time) ([]slaRef, error) {
	rows, err := db.Query(ctx, `WITH due AS (
		SELECT id
		  FROM workflow_tasks
		 WHERE status = 'open'
		   AND remind_after IS NOT NULL
		   AND remind_after <= $1
		   AND (last_reminded_at IS NULL OR last_reminded_at < remind_after)
		 ORDER BY remind_after, id
		 FOR UPDATE SKIP LOCKED
		 LIMIT $2
	)
	UPDATE workflow_tasks AS task
	   SET last_reminded_at = $1, updated_at = now()
	  FROM due
	 WHERE task.id = due.id
	   AND task.status = 'open'
	   AND task.remind_after <= $1
	   AND (task.last_reminded_at IS NULL OR task.last_reminded_at < task.remind_after)
	RETURNING task.id, task.instance_id, task.step_key, task.remind_after`, now, sweepSLABatchSize)
	if err != nil {
		return nil, kerr.Wrapf(err, "workflow.SweepSLA", "claim reminders")
	}
	return scanSLARefs(rows, "reminders")
}

func claimEscalationBatch(ctx context.Context, db database.TenantDB, now time.Time) ([]slaRef, error) {
	rows, err := db.Query(ctx, `WITH due AS (
		SELECT id
		  FROM workflow_tasks
		 WHERE status = 'open' AND due_at IS NOT NULL AND due_at <= $1
		 ORDER BY due_at, id
		 FOR UPDATE SKIP LOCKED
		 LIMIT $2
	)
	UPDATE workflow_tasks AS task
	   SET status = 'expired', version = version + 1, updated_at = now()
	  FROM due
	 WHERE task.id = due.id AND task.status = 'open'
	RETURNING task.id, task.instance_id, task.step_key, task.due_at`, now, sweepSLABatchSize)
	if err != nil {
		return nil, kerr.Wrapf(err, "workflow.SweepSLA", "claim escalations")
	}
	return scanSLARefs(rows, "escalations")
}

func scanSLARefs(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}, operation string,
) ([]slaRef, error) {
	refs := make([]slaRef, 0, sweepSLABatchSize)
	for rows.Next() {
		var ref slaRef
		if err := rows.Scan(&ref.id, &ref.instance, &ref.step, &ref.dueAt); err != nil {
			rows.Close()
			return nil, kerr.Wrapf(err, "workflow.SweepSLA", "scan %s", operation)
		}
		refs = append(refs, ref)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "workflow.SweepSLA", "iterate %s", operation)
	}
	return refs, nil
}

func (rt *Runtime) loadSLAState(ctx context.Context, db database.TenantDB, refs []slaRef) (map[uuid.UUID]slaState, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	instanceIDs := make([]uuid.UUID, 0, len(refs))
	seenInstances := make(map[uuid.UUID]struct{}, len(refs))
	for _, ref := range refs {
		if _, ok := seenInstances[ref.instance]; ok {
			continue
		}
		seenInstances[ref.instance] = struct{}{}
		instanceIDs = append(instanceIDs, ref.instance)
	}
	rows, err := db.Query(ctx, `SELECT id, definition_id, resource_type, resource_id, current_step, status, context, version
		FROM workflow_instances WHERE id = ANY($1::uuid[])`, instanceIDs)
	if err != nil {
		return nil, kerr.Wrapf(err, "workflow.SweepSLA", "batch load instances")
	}
	instances := make(map[uuid.UUID]Instance, len(instanceIDs))
	definitionIDs := make([]uuid.UUID, 0, len(instanceIDs))
	seenDefinitions := make(map[uuid.UUID]struct{}, len(instanceIDs))
	for rows.Next() {
		var inst Instance
		var rawContext []byte
		if err := rows.Scan(&inst.ID, &inst.DefinitionID, &inst.Resource.Type, &inst.Resource.ID,
			&inst.CurrentStep, &inst.Status, &rawContext, &inst.Version); err != nil {
			rows.Close()
			return nil, kerr.Wrapf(err, "workflow.SweepSLA", "scan instance")
		}
		inst.Context = decodeJSONMap(rawContext)
		instances[inst.ID] = inst
		if _, ok := seenDefinitions[inst.DefinitionID]; !ok {
			seenDefinitions[inst.DefinitionID] = struct{}{}
			definitionIDs = append(definitionIDs, inst.DefinitionID)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "workflow.SweepSLA", "iterate instances")
	}
	if len(instances) != len(instanceIDs) {
		return nil, kerr.E(kerr.KindNotFound, "workflow_instance_not_found", "workflow instance not found during SLA sweep")
	}

	defRows, err := db.Query(ctx, `SELECT id, key, version, definition
		FROM workflow_definitions WHERE id = ANY($1::uuid[])`, definitionIDs)
	if err != nil {
		return nil, kerr.Wrapf(err, "workflow.SweepSLA", "batch load definitions")
	}
	definitions := make(map[uuid.UUID]Definition, len(definitionIDs))
	for defRows.Next() {
		var id uuid.UUID
		var key string
		var version int
		var raw []byte
		if err := defRows.Scan(&id, &key, &version, &raw); err != nil {
			defRows.Close()
			return nil, kerr.Wrapf(err, "workflow.SweepSLA", "scan definition")
		}
		def, ok := rt.registry.definition(key, version)
		if !ok {
			var err error
			def, err = ParseDefinition(raw)
			if err != nil {
				defRows.Close()
				return nil, err
			}
		}
		definitions[id] = def
	}
	defRows.Close()
	if err := defRows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "workflow.SweepSLA", "iterate definitions")
	}
	if len(definitions) != len(definitionIDs) {
		return nil, kerr.E(kerr.KindNotFound, "workflow_definition_not_found", "workflow definition not found during SLA sweep")
	}

	state := make(map[uuid.UUID]slaState, len(instances))
	for id, inst := range instances {
		state[id] = slaState{instance: inst, def: definitions[inst.DefinitionID]}
	}
	return state, nil
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
