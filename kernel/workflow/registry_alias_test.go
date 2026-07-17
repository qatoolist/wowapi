package workflow

import (
	"strings"
	"testing"
)

// Second closure-audit regression (2026-07-17, F-10): Definition carries a
// Steps map with nested slices and pointers. The registry must deep-copy at
// registration — a module mutating the value it registered (its Steps map, a
// Transition, a Policy) must never alter the validated graph running
// instances resolve against.
func TestDefinitionNestedDataIsNotAliased(t *testing.T) {
	r := NewRegistry()
	selfApprove := false
	def := Definition{
		Key: "widgets.approval", Version: 1, AppliesTo: "widgets.thing", InitialStep: "review",
		Steps: map[string]Step{
			"review": {
				Type:      StepApproval,
				Assignees: []AssigneeSpec{{Kind: "role", Role: "approver"}},
				Policy:    &Policy{MinApprovals: 2, SelfApproval: &selfApprove},
				OnApprove: &Transition{Next: "done"},
				OnReject:  &Transition{Next: "rejected"},
			},
			"done":     {Type: StepTerminal, Outcome: "approved"},
			"rejected": {Type: StepTerminal, Outcome: "rejected"},
		},
	}
	if err := r.RegisterDefinition(def); err != nil {
		t.Fatal(err)
	}

	// Mutate the RETAINED registration value: replace a step, retarget a
	// transition, weaken the policy through the shared pointers.
	def.Steps["evil"] = Step{Type: StepTerminal, Outcome: "backdoor"}
	def.Steps["review"].OnApprove.Next = "rejected"
	def.Steps["review"].Policy.MinApprovals = 0
	*def.Steps["review"].Policy.SelfApproval = true
	def.Steps["review"].Assignees[0].Role = "anyone"

	got, ok := r.definition("widgets.approval", 1)
	if !ok {
		t.Fatal("definition missing")
	}
	if _, ok := got.Steps["evil"]; ok {
		t.Fatal("retained registration value injected a step into the validated graph")
	}
	review := got.Steps["review"]
	if review.OnApprove.Next != "done" {
		t.Fatalf("retained alias retargeted a transition: %+v", review.OnApprove)
	}
	if review.Policy.MinApprovals != 2 || *review.Policy.SelfApproval {
		t.Fatalf("retained alias weakened the approval policy: %+v", review.Policy)
	}
	if review.Assignees[0].Role != "approver" {
		t.Fatalf("retained alias changed the assignees: %+v", review.Assignees)
	}
}

// Third closure-audit regression (2026-07-17, F-10): Condition.Equals is `any`
// — a mutable value (map/slice/pointer) would survive the definition clone as
// a shared reference and let a module change gateway routing after boot. The
// invalid state is unrepresentable: registration validation rejects every
// non-scalar Equals.
func TestGatewayConditionRejectsMutableEqualsValues(t *testing.T) {
	for name, equals := range map[string]any{
		"map":      map[string]any{"tier": "gold"},
		"slice":    []string{"gold"},
		"pointer":  &struct{ V string }{"gold"},
		"func":     func() {},
		"nil":      nil,
		"any-map":  map[any]any{1: 2},
		"struct{}": struct{ V string }{"gold"},
	} {
		t.Run(name, func(t *testing.T) {
			r := NewRegistry()
			if err := r.RegisterDefinition(Definition{
				Key: "widgets.gw", Version: 1, AppliesTo: "widgets.thing", InitialStep: "gate",
				Steps: map[string]Step{
					"gate": {Type: StepGateway, Branches: []Branch{
						{When: &Condition{Key: "tier", Equals: equals}, Next: "done"},
						{Next: "done"},
					}},
					"done": {Type: StepTerminal, Outcome: "ok"},
				},
			}); err != nil {
				return // rejected at registration — also acceptable
			}
			err := r.Err()
			if err == nil {
				t.Fatalf("a %s when.equals value passed validation — it aliases module-owned mutable memory", name)
			}
			if !strings.Contains(err.Error(), "immutable scalar") {
				t.Fatalf("validation error does not explain the scalar restriction: %v", err)
			}
		})
	}
}

// With Equals restricted to scalars, the definition clone is provably
// alias-free end to end: mutate everything reachable in the RETAINED
// registration value and prove gateway target selection over the compiled
// definition is unchanged. (Runs under -race in the race gate like every
// other test.)
func TestGatewayRoutingImmuneToRetainedDefinitionMutation(t *testing.T) {
	r := NewRegistry()
	def := Definition{
		Key: "widgets.gw", Version: 1, AppliesTo: "widgets.thing", InitialStep: "gate",
		Steps: map[string]Step{
			"gate": {Type: StepGateway, Branches: []Branch{
				{When: &Condition{Key: "tier", Equals: "gold"}, Next: "fast"},
				{Next: "slow"},
			}},
			"fast": {Type: StepTerminal, Outcome: "fast"},
			"slow": {Type: StepTerminal, Outcome: "slow"},
		},
	}
	if err := r.RegisterDefinition(def); err != nil {
		t.Fatal(err)
	}
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}

	// Mutate every reachable piece of the retained declaration.
	def.Steps["gate"].Branches[0].When.Equals = "platinum"
	def.Steps["gate"].Branches[0].When.Key = "rank"
	def.Steps["gate"].Branches[0].Next = "slow"
	def.Steps["evil"] = Step{Type: StepTerminal, Outcome: "backdoor"}

	got, ok := r.definition("widgets.gw", 1)
	if !ok {
		t.Fatal("definition missing")
	}
	rt := &Runtime{}
	if target := rt.gatewayTarget(got.Steps["gate"], map[string]any{"tier": "gold"}); target != "fast" {
		t.Fatalf("gateway routing changed after retained-declaration mutation: gold -> %q, want fast", target)
	}
	if target := rt.gatewayTarget(got.Steps["gate"], map[string]any{"tier": "silver"}); target != "slow" {
		t.Fatalf("default branch changed after retained-declaration mutation: silver -> %q, want slow", target)
	}
}
