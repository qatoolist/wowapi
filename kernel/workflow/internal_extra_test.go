package workflow

import (
	"context"
	"strings"
	"testing"
)

// internal_extra_test.go — white-box coverage for the definition validator's
// remaining error branches, the registry's registration guards, latestVersion,
// and the NewRuntime precondition panic. These are pure/unexported units better
// exercised in-package than through the DB-backed runtime.

func TestValidateEmptyKeyAndVersion(t *testing.T) {
	// Missing key and non-positive version both accumulate.
	def := Definition{Version: 0, InitialStep: "a", Steps: map[string]Step{
		"a": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "key is required") {
		t.Errorf("missing 'key is required': %v", msg)
	}
	if !strings.Contains(msg, "version must be a positive integer") {
		t.Errorf("missing version error: %v", msg)
	}
}

func TestValidateNoSteps(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a"}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "definition has no steps") {
		t.Fatalf("expected no-steps error, got %v", err)
	}
}

func TestValidateUnknownStepType(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a": {Type: StepType("mystery"), Next: &Transition{Next: "end"}},
		// "end" is a terminal so a terminal is reachable; the only complaint is the
		// unknown type of "a".
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), `has unknown type "mystery"`) {
		t.Fatalf("expected unknown-type error, got %v", err)
	}
}

func TestValidateUnknownAssigneeKind(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a": {
			Type:      StepApproval,
			Assignees: []AssigneeSpec{{Kind: "wat"}},
			OnApprove: &Transition{Next: "end"}, OnReject: &Transition{Next: "end"},
		},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), `unknown kind "wat"`) {
		t.Fatalf("expected unknown-assignee-kind error, got %v", err)
	}
}

func TestValidateResolverKindMissingResolverKey(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a": {
			Type:      StepApproval,
			Assignees: []AssigneeSpec{{Kind: SpecResolver}}, // Resolver == ""
			OnApprove: &Transition{Next: "end"}, OnReject: &Transition{Next: "end"},
		},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "resolver kind requires a resolver key") {
		t.Fatalf("expected missing-resolver-key error, got %v", err)
	}
}

func TestValidateAutoStepMissingAction(t *testing.T) {
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":   {Type: StepAuto, Next: &Transition{Next: "end"}}, // Action == ""
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "requires an action") {
		t.Fatalf("expected missing-action error, got %v", err)
	}
}

// TestValidateGatewayOutgoing exercises the gateway arm of Step.outgoing() (a
// gateway's branch targets must exist and count toward reachability).
func TestValidateGatewayValidAndDangling(t *testing.T) {
	// Valid: gateway branches reach a terminal.
	valid := Definition{Key: "k.v", Version: 1, InitialStep: "g", Steps: map[string]Step{
		"g": {Type: StepGateway, Branches: []Branch{
			{When: &Condition{Key: "tier", Equals: "gold"}, Next: "end"},
			{Next: "end"},
		}},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	if err := valid.Validate(nil, nil); err != nil {
		t.Fatalf("valid gateway rejected: %v", err)
	}
	// Dangling: a gateway branch points at a missing step.
	bad := Definition{Key: "k.v", Version: 1, InitialStep: "g", Steps: map[string]Step{
		"g":   {Type: StepGateway, Branches: []Branch{{Next: "ghost"}, {Next: "end"}}},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := bad.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), `transitions to unknown step "ghost"`) {
		t.Fatalf("expected dangling gateway branch error, got %v", err)
	}
}

func TestRegisterDefinitionEmptyKey(t *testing.T) {
	reg := NewRegistry()
	if err := reg.RegisterDefinition(Definition{Version: 1}); err == nil {
		t.Fatal("expected empty-key registration error")
	}
	// The error is also accumulated for Err().
	if err := reg.Err(); err == nil {
		t.Fatal("expected Err() to surface the registration error")
	}
}

func TestRegisterAutoActionGuards(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterAutoAction("", func(context.Context, AutoInput) (map[string]any, error) { return nil, nil })
	if err := reg.Err(); err == nil || !strings.Contains(err.Error(), "requires a key and fn") {
		t.Fatalf("expected empty-key auto-action error, got %v", err)
	}

	reg2 := NewRegistry()
	fn := func(context.Context, AutoInput) (map[string]any, error) { return nil, nil }
	reg2.RegisterAutoAction("dup", fn)
	reg2.RegisterAutoAction("dup", fn)
	if err := reg2.Err(); err == nil || !strings.Contains(err.Error(), "registered more than once") {
		t.Fatalf("expected duplicate auto-action error, got %v", err)
	}
}

func TestRegisterAssigneeResolverGuards(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterAssigneeResolver("nofn", nil)
	if err := reg.Err(); err == nil || !strings.Contains(err.Error(), "requires a key and fn") {
		t.Fatalf("expected nil-fn resolver error, got %v", err)
	}

	reg2 := NewRegistry()
	fn := func(context.Context, ResolveInput) ([]Assignee, error) { return nil, nil }
	reg2.RegisterAssigneeResolver("dup", fn)
	reg2.RegisterAssigneeResolver("dup", fn)
	if err := reg2.Err(); err == nil || !strings.Contains(err.Error(), "registered more than once") {
		t.Fatalf("expected duplicate resolver error, got %v", err)
	}
}

func TestLatestVersion(t *testing.T) {
	reg := NewRegistry()
	term := map[string]Step{"a": {Type: StepTerminal, Outcome: "completed"}}
	if err := reg.RegisterDefinition(Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: term}); err != nil {
		t.Fatalf("register v1: %v", err)
	}
	if err := reg.RegisterDefinition(Definition{Key: "k.v", Version: 3, InitialStep: "a", Steps: term}); err != nil {
		t.Fatalf("register v3: %v", err)
	}
	// Registering an older version must NOT lower the recorded latest.
	if err := reg.RegisterDefinition(Definition{Key: "k.v", Version: 2, InitialStep: "a", Steps: term}); err != nil {
		t.Fatalf("register v2: %v", err)
	}
	if v, ok := reg.latestVersion("k.v"); !ok || v != 3 {
		t.Fatalf("latestVersion = %d, %v; want 3, true", v, ok)
	}
	if _, ok := reg.latestVersion("does.not.exist"); ok {
		t.Fatal("latestVersion for unknown key should be (0, false)")
	}
}

func TestValidateApprovalRequiresOnApprove(t *testing.T) {
	// Mirror of the existing on_reject regression: a missing on_approve must also
	// be rejected at boot (an approval could otherwise dead-end).
	def := Definition{Key: "k.v", Version: 1, InitialStep: "a", Steps: map[string]Step{
		"a":   {Type: StepApproval, OnReject: &Transition{Next: "end"}},
		"end": {Type: StepTerminal, Outcome: "completed"},
	}}
	err := def.Validate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "must define on_approve.next") {
		t.Fatalf("expected missing on_approve error, got %v", err)
	}
}

func TestParseISODurationTrailingAndOverflow(t *testing.T) {
	// A trailing number with no unit is an error (e.g. "P1D3").
	if _, err := parseISODuration("P1D3"); err == nil || !strings.Contains(err.Error(), "trailing number") {
		t.Fatalf("expected trailing-number error, got %v", err)
	}
	// A number too large for int overflows strconv.Atoi and must surface as an error.
	if _, err := parseISODuration("P99999999999999999999D"); err == nil {
		t.Fatal("expected overflow error for an oversized duration number")
	}
}

func TestNewRuntimePanicsOnNilDeps(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewRuntime with nil deps must panic")
		}
	}()
	NewRuntime(nil, nil, nil, nil, nil, nil)
}

func TestStatusForOutcomeMapping(t *testing.T) {
	cases := map[string]string{
		"rejected":   "rejected",
		"cancelled":  "cancelled",
		"canceled":   "cancelled",
		"overridden": "overridden",
		"completed":  "completed",
		"":           "completed",
		"anything":   "completed",
	}
	for outcome, want := range cases {
		if got := statusForOutcome(outcome); got != want {
			t.Errorf("statusForOutcome(%q) = %q, want %q", outcome, got, want)
		}
	}
}
