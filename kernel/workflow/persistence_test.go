package workflow

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
)

const maximalDefinitionYAML = `
key: catalog.maximal
version: 7
applies_to: catalog.item
initial_step: review
steps:
  review:
    type: approval
    assignees:
      - {kind: actor, actor: 00000000-0000-0000-0000-000000000001}
      - {kind: role, role: reviewer, scope: resource}
      - {kind: relationship, rel: owner}
      - {kind: resource_owner}
      - {kind: resolver, resolver: catalog.reviewers}
    sla: {due: P2D, remind_after: PT1H, escalate_to: step:manual}
    on_approve: {next: automate, require_comment: true}
    on_reject: {next: rejected}
  automate:
    type: auto
    action: catalog.provision
    next: {next: route}
    on_error: {retry: bounded, then: manual}
  manual:
    type: task
    assignees: [{kind: role, role: operator}]
    next: {next: route}
  route:
    type: gateway
    branches:
      - when: {key: amount, equals: 9223372036854775807}
        next: completed
      - next: rejected
  completed: {type: terminal, outcome: completed}
  rejected: {type: terminal, outcome: rejected}
`

func mustMaximalDefinition(t *testing.T) Definition {
	t.Helper()
	def, err := ParseDefinition([]byte(maximalDefinitionYAML))
	if err != nil {
		t.Fatal(err)
	}
	return def
}

func TestCanonicalDefinitionV1Golden(t *testing.T) {
	def := mustMaximalDefinition(t)
	canonical, err := canonicalDefinitionV1(def)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := definitionDigestV1(def)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := os.ReadFile("testdata/canonical_definition_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	wantJSON = bytes.TrimSpace(wantJSON)
	wantDigestBytes, err := os.ReadFile("testdata/canonical_definition_v1.sha256")
	if err != nil {
		t.Fatal(err)
	}
	wantDigest := strings.TrimSpace(string(wantDigestBytes))
	if !bytes.Equal(canonical, wantJSON) {
		t.Fatalf("canonical JSON changed:\n got: %s\nwant: %s", canonical, wantJSON)
	}
	if digest != wantDigest {
		t.Fatalf("canonical digest changed: got %s want %s", digest, wantDigest)
	}

	parsed, err := ParseDefinition(canonical)
	if err != nil {
		t.Fatalf("strict parse canonical JSON: %v", err)
	}
	remarshal, err := canonicalDefinitionV1(parsed)
	if err != nil {
		t.Fatal(err)
	}
	if string(remarshal) != string(canonical) {
		t.Fatalf("canonical remarshal drifted:\nfirst: %s\nagain: %s", canonical, remarshal)
	}
}

func TestDefinitionRowIDV1Golden(t *testing.T) {
	const want = "667ad7aa-aa51-5b8b-b9b1-bed2dc91df69"
	if got := definitionRowID("catalog.maximal", 7).String(); got != want {
		t.Fatalf("definition row identity changed: got %s want %s", got, want)
	}
}

func TestCanonicalDefinitionV1YAMLJSONAndMapOrderEquivalent(t *testing.T) {
	def := mustMaximalDefinition(t)
	want, err := canonicalDefinitionV1(def)
	if err != nil {
		t.Fatal(err)
	}
	var generic map[string]any
	decoder := json.NewDecoder(bytes.NewReader(want))
	decoder.UseNumber()
	if err := decoder.Decode(&generic); err != nil {
		t.Fatal(err)
	}
	// json.Marshal deliberately rebuilds the document through generic maps;
	// strict parsing must recover the same typed definition and canonical bytes.
	reorderedJSON, err := json.Marshal(generic)
	if err != nil {
		t.Fatal(err)
	}
	fromJSON, err := ParseDefinition(reorderedJSON)
	if err != nil {
		t.Fatal(err)
	}
	got, err := canonicalDefinitionV1(fromJSON)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("equivalent JSON canonicalized differently:\nYAML: %s\nJSON: %s", want, got)
	}

	// Reinsert steps in reverse order; map insertion order must be irrelevant.
	reversed := def.clone()
	reversed.Steps = make(map[string]Step, len(def.Steps))
	keys := sortedStepKeys(def.Steps)
	for i := len(keys) - 1; i >= 0; i-- {
		reversed.Steps[keys[i]] = def.Steps[keys[i]]
	}
	reordered, err := canonicalDefinitionV1(reversed)
	if err != nil {
		t.Fatal(err)
	}
	if string(reordered) != string(want) {
		t.Fatal("step map insertion order changed canonical representation")
	}
}

func TestCanonicalDefinitionV1NormalizesCollections(t *testing.T) {
	base := Definition{
		Key: "normalization", Version: 1, AppliesTo: "thing", InitialStep: "done",
		Steps: map[string]Step{"done": {Type: StepTerminal, Outcome: "done"}},
	}
	a := base.clone()
	a.Steps["done"] = Step{Type: StepTerminal, Outcome: "done", Assignees: nil, Branches: nil}
	b := base.clone()
	b.Steps["done"] = Step{Type: StepTerminal, Outcome: "done", Assignees: []AssigneeSpec{}, Branches: []Branch{}}
	ca, err := canonicalDefinitionV1(a)
	if err != nil {
		t.Fatal(err)
	}
	cb, err := canonicalDefinitionV1(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(ca) != string(cb) {
		t.Fatalf("nil and empty collections differ: %s != %s", ca, cb)
	}
}

func TestCanonicalDefinitionV1ConditionScalarBoundaries(t *testing.T) {
	values := []any{
		json.Number("-9223372036854775808"), json.Number("9223372036854775807"),
		json.Number("18446744073709551615"), json.Number("0.000000000000000001"),
		json.Number("-0.0"),
		"gold", true, false,
	}
	for _, value := range values {
		t.Run(strings.ReplaceAll(strings.TrimSpace(toString(value)), "/", "_"), func(t *testing.T) {
			def := gatewayDefinition(value)
			first, err := canonicalDefinitionV1(def)
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := ParseDefinition(first)
			if err != nil {
				t.Fatal(err)
			}
			second, err := canonicalDefinitionV1(parsed)
			if err != nil {
				t.Fatal(err)
			}
			if string(first) != string(second) {
				t.Fatalf("scalar %v changed across round trip: %s != %s", value, first, second)
			}
		})
	}
}

func TestCanonicalDefinitionV1PreservesArbitraryDecimalPrecision(t *testing.T) {
	equivalent := []json.Number{
		"0.1234567890123456789012345678900",
		"123456789012345678901234567890e-30",
		"12345678901234567890123456789e-29",
	}
	var want []byte
	for _, value := range equivalent {
		got, err := canonicalDefinitionV1(gatewayDefinition(value))
		if err != nil {
			t.Fatal(err)
		}
		if want == nil {
			want = got
		} else if !bytes.Equal(got, want) {
			t.Fatalf("equivalent exact decimals canonicalized differently:\nwant %s\n got %s", want, got)
		}
	}
	if !bytes.Contains(want, []byte(`12345678901234567890123456789e-29`)) {
		t.Fatalf("canonical form lost exact decimal precision: %s", want)
	}
}

func TestCanonicalDefinitionV1RejectsInvalidValuesBeforeDigest(t *testing.T) {
	for name, value := range map[string]any{
		"mutable map":   map[string]any{"x": 1},
		"mutable slice": []any{1},
		"non finite":    math.Inf(1),
		"bad number":    json.Number("01"),
	} {
		t.Run(name, func(t *testing.T) {
			if digest, err := definitionDigestV1(gatewayDefinition(value)); err == nil || digest != "" {
				t.Fatalf("definitionDigestV1(%T) = %q, %v; want rejection before digest", value, digest, err)
			}
		})
	}
	if _, err := ParseDefinition([]byte(maximalDefinitionYAML + "\n---\n{}\n")); err == nil {
		t.Fatal("multiple definition documents were accepted")
	}
}

func gatewayDefinition(value any) Definition {
	return Definition{
		Key: "condition.boundary", Version: 1, AppliesTo: "thing", InitialStep: "gate",
		Steps: map[string]Step{
			"gate": {Type: StepGateway, Branches: []Branch{
				{When: &Condition{Key: "value", Equals: value}, Next: "done"},
				{Next: "done"},
			}},
			"done": {Type: StepTerminal, Outcome: "done"},
		},
	}
}

func toString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
