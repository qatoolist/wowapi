package workflow_test

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/workflow"
)

// BenchmarkDefinitionValidate measures the graph/reference validation that
// gates every workflow definition at registration before it can serve requests.
func BenchmarkDefinitionValidate(b *testing.B) {
	def := workflow.Definition{
		Key: "expense.approval", Version: 1, AppliesTo: "expense.claim", InitialStep: "manager",
		Steps: map[string]workflow.Step{
			"manager": {
				Type:      workflow.StepApproval,
				Assignees: []workflow.AssigneeSpec{{Kind: workflow.SpecRole, Role: "manager"}},
				OnApprove: &workflow.Transition{Next: "risk"}, OnReject: &workflow.Transition{Next: "rejected"},
			},
			"risk": {
				Type: workflow.StepGateway,
				Branches: []workflow.Branch{
					{When: &workflow.Condition{Key: "high_risk", Equals: true}, Next: "review"},
					{Next: "approved"},
				},
			},
			"review": {
				Type:      workflow.StepTask,
				Assignees: []workflow.AssigneeSpec{{Kind: workflow.SpecResolver, Resolver: "risk.reviewers"}},
				Next:      &workflow.Transition{Next: "approved"},
			},
			"approved": {Type: workflow.StepTerminal, Outcome: "approved"},
			"rejected": {Type: workflow.StepTerminal, Outcome: "rejected"},
		},
	}
	resolvers := map[string]bool{"risk.reviewers": true}

	b.ReportAllocs()
	for b.Loop() {
		if err := def.Validate(nil, resolvers); err != nil {
			b.Fatalf("validate definition: %v", err)
		}
	}
}
