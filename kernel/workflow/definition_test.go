package workflow

import (
	"context"
	"strings"
	"testing"
)

// linearApprovalYAML is a valid linear approval → auto → terminal definition.
const linearApprovalYAML = `
key: requests.approval
version: 1
applies_to: requests.request
initial_step: manager_review
steps:
  manager_review:
    type: approval
    assignees:
      - kind: resolver
        resolver: test.approver
    on_approve: { next: provision }
    on_reject:  { next: end_rejected, require_comment: true }
  provision:
    type: auto
    action: requests.provision
    next: { next: end_done }
    on_error: { then: end_rejected }
  end_done:
    type: terminal
    outcome: completed
  end_rejected:
    type: terminal
    outcome: rejected
`

func autos(keys ...string) map[string]bool {
	m := map[string]bool{}
	for _, k := range keys {
		m[k] = true
	}
	return m
}

func TestParseDefinitionStrictUnknownKey(t *testing.T) {
	raw := `
key: k.v
version: 1
initial_step: a
bogus_field: nope
steps:
  a: { type: terminal, outcome: completed }
`
	if _, err := ParseDefinition([]byte(raw)); err == nil {
		t.Fatal("expected unknown-key parse error, got nil")
	}
}

func TestParseDefinitionValidLinear(t *testing.T) {
	def, err := ParseDefinition([]byte(linearApprovalYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := def.Validate(autos("requests.provision"), autos("test.approver")); err != nil {
		t.Fatalf("valid linear def rejected: %v", err)
	}
	if def.InitialStep != "manager_review" || len(def.Steps) != 4 {
		t.Fatalf("unexpected parse result: %+v", def)
	}
}

func TestValidateMissingInitialStep(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, Steps: map[string]Step{
		"a": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "initial_step is required") {
		t.Fatalf("expected initial_step error, got %v", err)
	}
}

func TestValidateInitialStepDoesNotExist(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "ghost", Steps: map[string]Step{
		"a": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), `initial_step "ghost" does not exist`) {
		t.Fatalf("expected missing initial_step error, got %v", err)
	}
}

func TestValidateDanglingTransition(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":   {Type: StepApproval, OnApprove: &Transition{Next: "nowhere"}, OnReject: &Transition{Next: "end"}},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), `transitions to unknown step "nowhere"`) {
		t.Fatalf("expected dangling transition error, got %v", err)
	}
}

func TestValidateOrphanStep(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":      {Type: StepApproval, OnApprove: &Transition{Next: "end"}, OnReject: &Transition{Next: "end"}},
		"end":    {Type: StepTerminal, Outcome: "completed"},
		"orphan": {Type: StepTask, Next: &Transition{Next: "end"}},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), `step "orphan" is unreachable`) {
		t.Fatalf("expected orphan error, got %v", err)
	}
}

func TestValidateNoTerminalReachable(t *testing.T) {
	// a -> b -> a, no terminal reachable.
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a": {Type: StepTask, Next: &Transition{Next: "b"}},
		"b": {Type: StepTask, Next: &Transition{Next: "a"}},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "no terminal step is reachable") {
		t.Fatalf("expected no-terminal error, got %v", err)
	}
}

func TestValidateUnknownAutoAction(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":   {Type: StepAuto, Action: "missing.action", Next: &Transition{Next: "end"}},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(autos("other.action"), nil)
	if err == nil || !strings.Contains(err.Error(), `unregistered auto-action "missing.action"`) {
		t.Fatalf("expected unknown auto-action error, got %v", err)
	}
}

func TestValidateUnknownResolver(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a": {
			Type: StepApproval, Assignees: []AssigneeSpec{{Kind: SpecResolver, Resolver: "missing.resolver"}},
			OnApprove: &Transition{Next: "end"}, OnReject: &Transition{Next: "end"},
		},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, autos("other.resolver"))
	if err == nil || !strings.Contains(err.Error(), `unregistered resolver "missing.resolver"`) {
		t.Fatalf("expected unknown resolver error, got %v", err)
	}
}

func TestValidateAccumulatesAllErrors(t *testing.T) {
	// Two independent problems must both surface (accumulate ALL).
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":      {Type: StepApproval, OnApprove: &Transition{Next: "nowhere"}, OnReject: &Transition{Next: "end"}},
		"end":    {Type: StepTerminal, Outcome: "completed"},
		"orphan": {Type: StepTask, Next: &Transition{Next: "end"}},
	}}
	err := def.Validate(nil, nil)
	if err == nil {
		t.Fatal("expected errors")
	}
	msg := err.Error()
	if !strings.Contains(msg, "nowhere") || !strings.Contains(msg, "orphan") {
		t.Fatalf("expected both dangling and orphan errors, got %v", msg)
	}
}

func TestRegistryErrRunsValidation(t *testing.T) {
	reg := NewRegistry()
	// Register a def referencing an auto action that is never registered.
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":   {Type: StepAuto, Action: "never.registered", Next: &Transition{Next: "end"}},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	if err := reg.RegisterDefinition(def); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := reg.Err(); err == nil || !strings.Contains(err.Error(), "never.registered") {
		t.Fatalf("expected Err() to surface validation, got %v", err)
	}

	// With the action registered, Err() is clean.
	reg2 := NewRegistry()
	if err := reg2.RegisterDefinition(def); err != nil {
		t.Fatalf("register: %v", err)
	}
	reg2.RegisterAutoAction("never.registered", func(context.Context, AutoInput) (map[string]any, error) {
		return nil, nil
	})
	if err := reg2.Err(); err != nil {
		t.Fatalf("expected clean Err() once action registered, got %v", err)
	}
}

func TestRegistryDuplicateDefinition(t *testing.T) {
	reg := NewRegistry()
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a": {Type: StepTerminal, Outcome: "completed"},
	}}
	if err := reg.RegisterDefinition(def); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := reg.RegisterDefinition(def); err == nil {
		t.Fatal("expected duplicate (key,version) error")
	}
}

// bp returns a pointer to a bool literal (for Policy.SelfApproval).
func bp(b bool) *bool { return &b }

// TestValidateFailsClosedOnUnenforcedGating is the SEC-36/37/38 regression: a
// definition that RELIES on gating the runtime does not yet enforce must be
// rejected at boot, not silently accepted and mis-tallied at runtime.
func TestValidateFailsClosedOnUnenforcedGating(t *testing.T) {
	cases := []struct {
		name string
		step Step
		want string
	}{
		{
			name: "vote step",
			step: Step{Type: StepVote, Next: &Transition{Next: "end"}},
			want: "vote steps are not yet tallied",
		},
		{
			name: "min_approvals > 1",
			step: Step{
				Type: StepApproval, Policy: &Policy{MinApprovals: 2},
				OnApprove: &Transition{Next: "end"}, OnReject: &Transition{Next: "end"},
			},
			want: "min_approvals > 1 is not yet enforced",
		},
		{
			name: "self_approval:false",
			step: Step{
				Type: StepApproval, Policy: &Policy{SelfApproval: bp(false)},
				OnApprove: &Transition{Next: "end"}, OnReject: &Transition{Next: "end"},
			},
			want: "self_approval:false is not yet enforced",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
				"a":   tc.step,
				"end": {Type: StepTerminal, Outcome: "completed"},
			}}
			err := def.Validate(nil, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected fail-closed error %q, got %v", tc.want, err)
			}
		})
	}
}

// TestValidateApprovalRequiresBothTransitions is the ARCH-64 regression: an
// approval step missing on_reject would dead-end a rejection at runtime.
func TestValidateApprovalRequiresBothTransitions(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":   {Type: StepApproval, OnApprove: &Transition{Next: "end"}},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "must define on_reject.next") {
		t.Fatalf("expected missing on_reject error, got %v", err)
	}
}
