package workflow

import "testing"

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
