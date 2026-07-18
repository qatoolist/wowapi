package workflow_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/workflow"
	"github.com/qatoolist/wowapi/testkit"
)

func terminalDefinition(key, outcome string) workflow.Definition {
	return workflow.Definition{
		Key: key, Version: 1, AppliesTo: "catalog.item", InitialStep: "done",
		Steps: map[string]workflow.Step{"done": {Type: workflow.StepTerminal, Outcome: outcome}},
	}
}

func validatedRegistry(t *testing.T, defs ...workflow.Definition) *workflow.Registry {
	t.Helper()
	reg := workflow.NewRegistry()
	for _, def := range defs {
		if err := reg.RegisterDefinition(def); err != nil {
			t.Fatal(err)
		}
	}
	if err := reg.Err(); err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestIntegrationSyncDefinitionsIdempotentStableIdentityAndJSONBRoundTrip(t *testing.T) {
	h := testkit.NewDB(t)
	reg := validatedRegistry(t, terminalDefinition("catalog.stable", "completed"))
	ctx := context.Background()
	if err := workflow.SyncDefinitions(ctx, h.Platform, reg); err != nil {
		t.Fatal(err)
	}
	var firstID uuid.UUID
	var firstDigest string
	var firstRaw []byte
	if err := h.Admin.QueryRow(ctx, `SELECT id, definition_digest, definition
		FROM workflow_definitions WHERE key='catalog.stable' AND version=1`).
		Scan(&firstID, &firstDigest, &firstRaw); err != nil {
		t.Fatal(err)
	}
	if err := workflow.SyncDefinitions(ctx, h.Platform, reg); err != nil {
		t.Fatalf("second sync after jsonb retrieval: %v", err)
	}
	var secondID uuid.UUID
	var secondDigest string
	var secondRaw []byte
	if err := h.Admin.QueryRow(ctx, `SELECT id, definition_digest, definition
		FROM workflow_definitions WHERE key='catalog.stable' AND version=1`).
		Scan(&secondID, &secondDigest, &secondRaw); err != nil {
		t.Fatal(err)
	}
	if firstID != secondID {
		t.Fatalf("idempotent sync changed row id: %s -> %s", firstID, secondID)
	}
	if firstDigest != secondDigest || len(firstDigest) != 64 {
		t.Fatalf("digest drifted across jsonb round trip: %q -> %q", firstDigest, secondDigest)
	}
	if !json.Valid(firstRaw) || !json.Valid(secondRaw) {
		t.Fatal("synchronized definition was not JSON")
	}
}

func TestIntegrationSyncDefinitionsRejectsSameVersionDivergence(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	if err := workflow.SyncDefinitions(ctx, h.Platform,
		validatedRegistry(t, terminalDefinition("catalog.immutable", "completed"))); err != nil {
		t.Fatal(err)
	}
	var id uuid.UUID
	var digest string
	if err := h.Admin.QueryRow(ctx, `SELECT id, definition_digest FROM workflow_definitions
		WHERE key='catalog.immutable' AND version=1`).Scan(&id, &digest); err != nil {
		t.Fatal(err)
	}
	err := workflow.SyncDefinitions(ctx, h.Platform,
		validatedRegistry(t, terminalDefinition("catalog.immutable", "rejected")))
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("same-version divergence = %v, want conflict", err)
	}
	var gotID uuid.UUID
	var gotDigest string
	if err := h.Admin.QueryRow(ctx, `SELECT id, definition_digest FROM workflow_definitions
		WHERE key='catalog.immutable' AND version=1`).Scan(&gotID, &gotDigest); err != nil {
		t.Fatal(err)
	}
	if gotID != id || gotDigest != digest {
		t.Fatalf("conflict mutated identity: id %s/%s digest %s/%s", id, gotID, digest, gotDigest)
	}
}

func TestIntegrationSyncDefinitionsLateConflictRollsBackEarlierInsert(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()
	if err := workflow.SyncDefinitions(ctx, h.Platform,
		validatedRegistry(t, terminalDefinition("z.last", "completed"))); err != nil {
		t.Fatal(err)
	}
	err := workflow.SyncDefinitions(ctx, h.Platform, validatedRegistry(t,
		terminalDefinition("a.first", "completed"),
		terminalDefinition("z.last", "rejected"),
	))
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("late divergence = %v, want conflict", err)
	}
	var count int
	if err := h.Admin.QueryRow(ctx, `SELECT count(*) FROM workflow_definitions WHERE key='a.first'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("earlier insert survived late conflict: count=%d", count)
	}
}

func TestIntegrationWorkflowDefinitionDigestConstraintRejectsMalformedValues(t *testing.T) {
	for _, digest := range []string{"", strings.Repeat("A", 64), strings.Repeat("a", 63), strings.Repeat("g", 64)} {
		t.Run(digestName(digest), func(t *testing.T) {
			h := testkit.NewDB(t)
			_, err := h.Admin.Exec(context.Background(), `INSERT INTO workflow_definitions
				(id,key,version,applies_to,definition,definition_digest,status,created_by)
				VALUES ($1,$2,1,'thing','{}',$3,'active',$4)`, uuid.New(), "bad."+uuid.NewString(), digest, uuid.Nil)
			if err == nil {
				t.Fatalf("database accepted malformed digest %q", digest)
			}
		})
	}
}

func digestName(digest string) string {
	if digest == "" {
		return "empty"
	}
	if len(digest) != 64 {
		return "wrong_length"
	}
	if strings.ToUpper(digest) == digest {
		return "uppercase"
	}
	return "non_hex"
}

func TestIntegrationStartRejectsDefinitionCorruptionBeforeInsert(t *testing.T) {
	for _, tc := range []struct {
		name      string
		recompute bool
	}{
		{name: "changed_json_unchanged_digest"},
		{name: "changed_json_recomputed_digest", recompute: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := testkit.NewDB(t)
			tn := testkit.CreateTenant(t, h)
			userID := testkit.CreateUser(t, h)
			cap := testkit.CreateCapacity(t, h, tn.ID, userID)
			res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
			rt := buildRuntime(t, h, cap, linearDef)
			if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_definitions
				SET definition = jsonb_set(definition, '{steps,end_done,outcome}', '"tampered"'::jsonb)
				WHERE key='requests.approval' AND version=1`); err != nil {
				t.Fatal(err)
			}
			if tc.recompute {
				var raw []byte
				if err := h.Admin.QueryRow(context.Background(), `SELECT definition FROM workflow_definitions
					WHERE key='requests.approval' AND version=1`).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				def, err := workflow.ParseDefinition(raw)
				if err != nil {
					t.Fatal(err)
				}
				canonical, err := json.Marshal(def)
				if err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(canonical)
				if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_definitions
					SET definition_digest=$1 WHERE key='requests.approval' AND version=1`, hex.EncodeToString(sum[:])); err != nil {
					t.Fatal(err)
				}
			}
			err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
				_, err := rt.StartIn(ctx, db, "requests.approval", res, nil)
				return err
			})
			if kerr.KindOf(err) != kerr.KindConflict {
				t.Fatalf("StartIn corruption = %v, want identity conflict", err)
			}
			var instances int
			if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM workflow_instances`).Scan(&instances); err != nil {
				t.Fatal(err)
			}
			if instances != 0 {
				t.Fatalf("StartIn inserted %d instance(s) before rejecting definition", instances)
			}
		})
	}
}

func TestIntegrationVerifiedLoaderRejectsMalformedDigestAndJSON(t *testing.T) {
	cases := []struct {
		name   string
		mutate string
		arg    any
	}{
		{name: "empty_digest", mutate: "definition_digest=$1", arg: ""},
		{name: "uppercase_digest", mutate: "definition_digest=$1", arg: strings.Repeat("A", 64)},
		{name: "wrong_length_digest", mutate: "definition_digest=$1", arg: strings.Repeat("a", 63)},
		{name: "non_hex_digest", mutate: "definition_digest=$1", arg: strings.Repeat("g", 64)},
		{name: "malformed_definition", mutate: "definition=$1", arg: []byte(`{}`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := testkit.NewDB(t)
			tn := testkit.CreateTenant(t, h)
			user := testkit.CreateUser(t, h)
			cap := testkit.CreateCapacity(t, h, tn.ID, user)
			res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
			rt := buildRuntime(t, h, cap, linearDef)
			if strings.Contains(tc.name, "digest") {
				if _, err := h.Admin.Exec(context.Background(), `ALTER TABLE workflow_definitions
					DROP CONSTRAINT workflow_definitions_definition_digest_check`); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_definitions SET `+tc.mutate+
				` WHERE key='requests.approval' AND version=1`, tc.arg); err != nil {
				t.Fatal(err)
			}
			err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
				_, err := rt.StartIn(ctx, db, "requests.approval", res, nil)
				return err
			})
			if kerr.KindOf(err) != kerr.KindConflict {
				t.Fatalf("malformed persisted identity = %v, want conflict", err)
			}
			var count int
			if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM workflow_instances`).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Fatalf("malformed identity inserted %d instances", count)
			}
		})
	}
}

func TestIntegrationStartRequiresSynchronizedDefinitionRow(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	user := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, user)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
	rt := buildRuntime(t, h, cap, linearDef)
	if _, err := h.Admin.Exec(context.Background(), `DELETE FROM workflow_definitions
		WHERE key='requests.approval' AND version=1`); err != nil {
		t.Fatal(err)
	}
	err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, err := rt.StartIn(ctx, db, "requests.approval", res, nil)
		return err
	})
	if kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("missing synchronized row = %v, want not found", err)
	}
}

func TestIntegrationMutationMethodsRejectDefinitionMismatchWithoutEffects(t *testing.T) {
	tests := []struct {
		name string
		def  string
		run  func(*workflow.Runtime, context.Context, uuid.UUID, uuid.UUID, authzFixture) error
	}{
		{name: "Decide", def: linearDef, run: func(rt *workflow.Runtime, ctx context.Context, taskID, _ uuid.UUID, a authzFixture) error {
			return rt.Decide(ctx, taskID, workflow.Decision{Actor: actor(a.tenant, a.user, a.capacity), Type: workflow.DecisionApprove})
		}},
		{name: "CompleteTask", def: taskDef, run: func(rt *workflow.Runtime, ctx context.Context, taskID, _ uuid.UUID, _ authzFixture) error {
			return rt.CompleteTask(ctx, taskID, map[string]any{"done": true})
		}},
		{name: "Delegate", def: linearDef, run: func(rt *workflow.Runtime, ctx context.Context, taskID, _ uuid.UUID, _ authzFixture) error {
			return rt.Delegate(ctx, taskID, uuid.New(), time.Now().Add(time.Hour))
		}},
		{name: "Override", def: linearDef, run: func(rt *workflow.Runtime, ctx context.Context, _ uuid.UUID, instanceID uuid.UUID, a authzFixture) error {
			return rt.Override(ctx, actor(a.tenant, a.user, a.capacity), instanceID, "end_rejected", "test")
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := testkit.NewDB(t)
			tn := testkit.CreateTenant(t, h)
			user := testkit.CreateUser(t, h)
			cap := testkit.CreateCapacity(t, h, tn.ID, user)
			res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
			rt := buildRuntime(t, h, cap, tc.def)
			key := "requests.approval"
			step := "manager_review"
			if tc.def == taskDef {
				key, step = "requests.task", "do_work"
			}
			sim := testkit.NewWorkflowSim(t, h, rt).Start(key, res, nil)
			taskID := openTaskID(t, h, sim.InstanceID(), step)
			var beforeEvents int
			if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM events_outbox`).Scan(&beforeEvents); err != nil {
				t.Fatal(err)
			}
			if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_definitions SET applies_to='tampered'
				WHERE key=$1 AND version=1`, key); err != nil {
				t.Fatal(err)
			}
			err := tc.run(rt, testkit.TenantCtx(tn.ID), taskID, sim.InstanceID(), authzFixture{tn.ID, user, cap})
			if kerr.KindOf(err) != kerr.KindConflict {
				t.Fatalf("%s mismatch = %v, want conflict", tc.name, err)
			}
			var taskStatus, instanceStatus, currentStep string
			var afterEvents int
			if err := h.Admin.QueryRow(context.Background(), `SELECT status FROM workflow_tasks WHERE id=$1`, taskID).Scan(&taskStatus); err != nil {
				t.Fatal(err)
			}
			if err := h.Admin.QueryRow(context.Background(), `SELECT status,current_step FROM workflow_instances WHERE id=$1`, sim.InstanceID()).Scan(&instanceStatus, &currentStep); err != nil {
				t.Fatal(err)
			}
			if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM events_outbox`).Scan(&afterEvents); err != nil {
				t.Fatal(err)
			}
			if taskStatus != "open" || instanceStatus != "running" || currentStep != step || afterEvents != beforeEvents {
				t.Fatalf("partial effect: task=%s instance=%s step=%s events=%d/%d", taskStatus, instanceStatus, currentStep, beforeEvents, afterEvents)
			}
		})
	}
}

type authzFixture struct{ tenant, user, capacity uuid.UUID }

func TestIntegrationMissingRegistryDefinitionAndSLACorruptionFailClosed(t *testing.T) {
	h := testkit.NewDB(t)
	tn := testkit.CreateTenant(t, h)
	user := testkit.CreateUser(t, h)
	cap := testkit.CreateCapacity(t, h, tn.ID, user)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
	rt := buildRuntime(t, h, cap, linearDef)
	sim := testkit.NewWorkflowSim(t, h, rt).Start("requests.approval", res, nil)
	taskID := openTaskID(t, h, sim.InstanceID(), "manager_review")

	empty := workflow.NewRegistry()
	if err := empty.Err(); err != nil {
		t.Fatal(err)
	}
	missingRT := workflow.NewRuntime(h.TxM, empty, fakeEvaluator{}, outbox.NewWriter(model.UUIDv7()), model.UUIDv7(), audit.New(model.UUIDv7(), nil))
	err := missingRT.Decide(testkit.TenantCtx(tn.ID), taskID,
		workflow.Decision{Actor: actor(tn.ID, user, cap), Type: workflow.DecisionApprove})
	if kerr.KindOf(err) != kerr.KindInternal || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("missing registry definition = %v", err)
	}

	if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_tasks
		SET due_at=now()-interval '1 hour', remind_after=now()-interval '1 hour' WHERE id=$1`, taskID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_definitions SET applies_to='tampered'
		WHERE key='requests.approval' AND version=1`); err != nil {
		t.Fatal(err)
	}
	err = h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		_, _, err := rt.SweepSLA(ctx, db, time.Now())
		return err
	})
	if kerr.KindOf(err) != kerr.KindConflict {
		t.Fatalf("SweepSLA identity mismatch = %v, want conflict", err)
	}
	var status string
	var reminded *time.Time
	var escalationEvents int
	if err := h.Admin.QueryRow(context.Background(), `SELECT status,last_reminded_at FROM workflow_tasks WHERE id=$1`, taskID).Scan(&status, &reminded); err != nil {
		t.Fatal(err)
	}
	if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM events_outbox
		WHERE event_type IN ('workflow.requests.approval.reminded','workflow.requests.approval.escalated')`).Scan(&escalationEvents); err != nil {
		t.Fatal(err)
	}
	if status != "open" || reminded != nil || escalationEvents != 0 {
		t.Fatalf("SLA corruption caused effects: status=%s reminded=%v events=%d", status, reminded, escalationEvents)
	}
}

func TestIntegrationSweepSLAVerifiesIdentityBeforeCallerOwnedTransactionEffects(t *testing.T) {
	for _, tc := range []struct {
		name   string
		dueSQL string
	}{
		{name: "reminder", dueSQL: `UPDATE workflow_tasks
			SET remind_after=now()-interval '1 hour', due_at=NULL WHERE id=$1`},
		{name: "escalation", dueSQL: `UPDATE workflow_tasks
			SET due_at=now()-interval '1 hour', remind_after=NULL WHERE id=$1`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := testkit.NewDB(t)
			tn := testkit.CreateTenant(t, h)
			user := testkit.CreateUser(t, h)
			cap := testkit.CreateCapacity(t, h, tn.ID, user)
			res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
			rt := buildRuntime(t, h, cap, linearDef)
			sim := testkit.NewWorkflowSim(t, h, rt).Start("requests.approval", res, nil)
			taskID := openTaskID(t, h, sim.InstanceID(), "manager_review")

			if _, err := h.Admin.Exec(context.Background(), tc.dueSQL, taskID); err != nil {
				t.Fatalf("make %s due: %v", tc.name, err)
			}
			if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_definitions
				SET applies_to='tampered' WHERE key='requests.approval' AND version=1`); err != nil {
				t.Fatal(err)
			}

			var identityErr error
			// The caller owns this tenant transaction, deliberately ignores the
			// SweepSLA error, and returns nil so the transaction commits. Any
			// pre-verification write therefore becomes observable after commit.
			if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
				_, _, identityErr = rt.SweepSLA(ctx, db, time.Now())
				return nil
			}); err != nil {
				t.Fatalf("caller-owned transaction: %v", err)
			}
			if kerr.KindOf(identityErr) != kerr.KindConflict {
				t.Fatalf("SweepSLA identity mismatch = %v, want conflict", identityErr)
			}

			var status string
			var reminded *time.Time
			if err := h.Admin.QueryRow(context.Background(), `SELECT status,last_reminded_at
				FROM workflow_tasks WHERE id=$1`, taskID).Scan(&status, &reminded); err != nil {
				t.Fatal(err)
			}
			var effects int
			if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM events_outbox
				WHERE tenant_id=$1 AND event_type IN
				('workflow.requests.approval.reminded','workflow.requests.approval.escalated')`, tn.ID).Scan(&effects); err != nil {
				t.Fatal(err)
			}
			if status != "open" || reminded != nil || effects != 0 {
				t.Fatalf("%s identity error committed partial effects: status=%s reminded=%v outbox=%d",
					tc.name, status, reminded, effects)
			}
		})
	}
}

func TestIntegrationSweepSLARejectsCorruptTaskAssociationBeforeEffects(t *testing.T) {
	for _, tc := range []struct {
		name    string
		corrupt func(*testing.T, *testkit.DBHandle, uuid.UUID, uuid.UUID)
	}{
		{name: "unknown task step", corrupt: func(t *testing.T, h *testkit.DBHandle, taskID, _ uuid.UUID) {
			if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_tasks
				SET step_key='missing_step', due_at=now()-interval '1 hour', remind_after=NULL WHERE id=$1`, taskID); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "non-running instance", corrupt: func(t *testing.T, h *testkit.DBHandle, taskID, instanceID uuid.UUID) {
			if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_tasks
				SET due_at=now()-interval '1 hour', remind_after=NULL WHERE id=$1`, taskID); err != nil {
				t.Fatal(err)
			}
			if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_instances SET status='completed' WHERE id=$1`, instanceID); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := testkit.NewDB(t)
			tn := testkit.CreateTenant(t, h)
			user := testkit.CreateUser(t, h)
			cap := testkit.CreateCapacity(t, h, tn.ID, user)
			res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")
			rt := buildRuntime(t, h, cap, linearDef)
			sim := testkit.NewWorkflowSim(t, h, rt).Start("requests.approval", res, nil)
			taskID := openTaskID(t, h, sim.InstanceID(), "manager_review")
			tc.corrupt(t, h, taskID, sim.InstanceID())

			var associationErr error
			if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
				_, _, associationErr = rt.SweepSLA(ctx, db, time.Now())
				return nil // deliberately commit after ignoring the sweep error
			}); err != nil {
				t.Fatalf("caller-owned transaction: %v", err)
			}
			if kerr.KindOf(associationErr) != kerr.KindConflict {
				t.Fatalf("corrupt association = %v, want conflict", associationErr)
			}

			var status string
			var reminded *time.Time
			if err := h.Admin.QueryRow(context.Background(), `SELECT status,last_reminded_at
				FROM workflow_tasks WHERE id=$1`, taskID).Scan(&status, &reminded); err != nil {
				t.Fatal(err)
			}
			var effects int
			if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM events_outbox
				WHERE tenant_id=$1 AND event_type IN
				('workflow.requests.approval.reminded','workflow.requests.approval.escalated')`, tn.ID).Scan(&effects); err != nil {
				t.Fatal(err)
			}
			if status != "open" || reminded != nil || effects != 0 {
				t.Fatalf("corrupt association committed effects: status=%s reminded=%v outbox=%d", status, reminded, effects)
			}
		})
	}
}
