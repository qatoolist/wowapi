// Package workflow is wowapi's small custom Postgres-backed workflow engine:
// a closed-step-type approval/state-machine runtime that shares the caller's
// tenant transaction (RLS + outbox + audit) exactly as blueprint 02 §1 and
// decisions D-0051/D-0053 specify.
//
// The kernel owns the runtime; modules own the definitions (seeded JSON/YAML)
// via a boot-validated Registry. Definitions are immutable per version and
// running instances pin their version. Every transition re-checks the actor
// (assignee + optional `workflow.task.decide` permission), mutates instance and
// task rows with optimistic locking, and writes the matching outbox event in
// the SAME tenant transaction as the state change.
//
// Import boundary (depguard): stdlib + kernel/{database,authz,resource,outbox,
// audit,errors,model,pagination} + pgx + uuid + yaml. NEVER module/app/adapters/
// testkit in production. Domain-neutral vocabulary only.
package workflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"gopkg.in/yaml.v3"
)

// StepType is the closed set of workflow step kinds (D-0053). A definition that
// carries any other type fails validation.
type StepType string

const (
	// StepApproval is an approve/reject decision by one or more assignees.
	StepApproval StepType = "approval"
	// StepTask is a do-something step marked done (with optional output).
	StepTask StepType = "task"
	// StepAuto invokes a registered module Go action, then advances.
	StepAuto StepType = "auto"
	// StepGateway branches on a simple predicate over the instance context.
	StepGateway StepType = "gateway"
	// StepTerminal ends the instance with an outcome.
	StepTerminal StepType = "terminal"
)

// validStepTypes is the closed step-type set for validation.
var validStepTypes = map[StepType]bool{
	StepApproval: true, StepTask: true, StepAuto: true,
	StepGateway: true, StepTerminal: true,
}

// Assignee spec kinds (closed set). These describe how a step's assignees are
// derived at task-creation time; they are resolved into concrete Assignee rows.
const (
	SpecActor         = "actor"          // explicit acting capacity
	SpecRole          = "role"           // role-at-scope
	SpecRelationship  = "relationship"   // relationship-holder
	SpecResourceOwner = "resource_owner" // owner of the target resource
	SpecResolver      = "resolver"       // module-registered resolver func
)

var validSpecKinds = map[string]bool{
	SpecActor: true, SpecRole: true, SpecRelationship: true,
	SpecResourceOwner: true, SpecResolver: true,
}

// Definition is the JSON/YAML workflow definition: a versioned, seedable graph
// of steps. It is immutable per (Key, Version); running instances pin a version.
type Definition struct {
	Key         string          `json:"key" yaml:"key"`
	Version     int             `json:"version" yaml:"version"`
	AppliesTo   string          `json:"applies_to" yaml:"applies_to"`
	InitialStep string          `json:"initial_step" yaml:"initial_step"`
	Steps       map[string]Step `json:"steps" yaml:"steps"`
}

// Step is one node in the definition graph. Which fields are meaningful depends
// on Type; validation and the runtime read only the relevant ones.
type Step struct {
	Type      StepType       `json:"type" yaml:"type"`
	Assignees []AssigneeSpec `json:"assignees,omitempty" yaml:"assignees,omitempty"`
	SLA       *SLA           `json:"sla,omitempty" yaml:"sla,omitempty"`

	// approval transitions.
	OnApprove *Transition `json:"on_approve,omitempty" yaml:"on_approve,omitempty"`
	OnReject  *Transition `json:"on_reject,omitempty" yaml:"on_reject,omitempty"`

	// task / auto / gateway default transition.
	Next *Transition `json:"next,omitempty" yaml:"next,omitempty"`

	// auto step.
	Action  string      `json:"action,omitempty" yaml:"action,omitempty"`
	OnError *Transition `json:"on_error,omitempty" yaml:"on_error,omitempty"`

	// gateway step.
	Branches []Branch `json:"branches,omitempty" yaml:"branches,omitempty"`

	// terminal step.
	Outcome string `json:"outcome,omitempty" yaml:"outcome,omitempty"`
}

// AssigneeSpec describes one source of assignees for a step.
type AssigneeSpec struct {
	Kind     string `json:"kind" yaml:"kind"`
	Actor    string `json:"actor,omitempty" yaml:"actor,omitempty"`       // capacity id (kind=actor)
	Role     string `json:"role,omitempty" yaml:"role,omitempty"`         // role key (kind=role)
	Scope    string `json:"scope,omitempty" yaml:"scope,omitempty"`       // scope hint (kind=role)
	Rel      string `json:"rel,omitempty" yaml:"rel,omitempty"`           // relationship type (kind=relationship)
	Resolver string `json:"resolver,omitempty" yaml:"resolver,omitempty"` // resolver key (kind=resolver)
}

// SLA carries the reminder/escalation timings for a step (ISO-8601 durations).
type SLA struct {
	Due         string `json:"due,omitempty" yaml:"due,omitempty"`
	RemindAfter string `json:"remind_after,omitempty" yaml:"remind_after,omitempty"`
	EscalateTo  string `json:"escalate_to,omitempty" yaml:"escalate_to,omitempty"` // "step:key" or "key"
}

// Transition is an edge to another step (Next) with optional decision flags.
type Transition struct {
	Next           string `json:"next,omitempty" yaml:"next,omitempty"`
	RequireComment bool   `json:"require_comment,omitempty" yaml:"require_comment,omitempty"`
	Retry          string `json:"retry,omitempty" yaml:"retry,omitempty"` // auto on_error retry policy (advisory)
	Then           string `json:"then,omitempty" yaml:"then,omitempty"`   // auto on_error target step
}

// target returns the step key this transition points at (Next, or Then for
// an on_error transition).
func (t *Transition) target() string {
	if t == nil {
		return ""
	}
	if t.Next != "" {
		return t.Next
	}
	return t.Then
}

// Branch is one gateway edge. A nil When is the default (fallthrough).
type Branch struct {
	When *Condition `json:"when,omitempty" yaml:"when,omitempty"`
	Next string     `json:"next" yaml:"next"`
}

// Condition is a minimal equality predicate over the instance context.
type Condition struct {
	Key    string `json:"key" yaml:"key"`
	Equals any    `json:"equals" yaml:"equals"`
}

// scalarConditionValue reports whether v is one of the immutable scalar kinds
// a compiled gateway condition supports. nil is excluded: a When without a
// value is meaningless (a default branch omits When entirely).
func scalarConditionValue(v any) bool {
	switch f := v.(type) {
	case json.Number:
		_, err := json.Marshal(f)
		return err == nil
	case float32:
		return !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0)
	case float64:
		return !math.IsNaN(f) && !math.IsInf(f, 0)
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

// stripStepPrefix normalizes an "step:key" escalate_to reference to "key".
func stripStepPrefix(s string) string { return strings.TrimPrefix(s, "step:") }

// ParseDefinition parses a strict JSON/YAML definition. Unknown keys are an
// error (KnownFields), so a typo in a seed fails loudly at load rather than
// silently dropping a step or transition. JSON is a subset of YAML, so this one
// path covers both.
func ParseDefinition(raw []byte) (Definition, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		dec.DisallowUnknownFields()
		dec.UseNumber()
		var d Definition
		if err := dec.Decode(&d); err != nil {
			return Definition{}, kerr.E(kerr.KindValidation, "workflow_definition_parse",
				"invalid workflow definition: "+err.Error())
		}
		var trailing any
		if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
			if err == nil {
				err = fmt.Errorf("multiple JSON values are not allowed")
			}
			return Definition{}, kerr.E(kerr.KindValidation, "workflow_definition_parse",
				"invalid workflow definition: "+err.Error())
		}
		return d, nil
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	var d Definition
	if err := dec.Decode(&d); err != nil {
		return Definition{}, kerr.E(kerr.KindValidation, "workflow_definition_parse",
			"invalid workflow definition: "+err.Error())
	}
	// A definition is exactly one document. Accepting a valid first document
	// and ignoring trailing content would make the persisted digest cover less
	// than the input a caller believed it registered.
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = fmt.Errorf("multiple documents are not allowed")
		}
		return Definition{}, kerr.E(kerr.KindValidation, "workflow_definition_parse",
			"invalid workflow definition: "+err.Error())
	}
	// Round-trip through JSON once so Condition.Equals uses the same immutable
	// scalar representation whether the source was YAML or JSON. UseNumber
	// preserves integer boundaries instead of converting through float64.
	b, err := json.Marshal(d)
	if err != nil {
		return Definition{}, kerr.E(kerr.KindValidation, "workflow_definition_parse",
			"invalid workflow definition: "+err.Error())
	}
	jd := json.NewDecoder(bytes.NewReader(b))
	jd.UseNumber()
	if err := jd.Decode(&d); err != nil {
		return Definition{}, kerr.E(kerr.KindValidation, "workflow_definition_parse",
			"invalid workflow definition: "+err.Error())
	}
	return d, nil
}

// Validate checks the definition graph and its external references, accumulating
// ALL problems into a single error. autoActions and resolvers are the sets of
// registered keys the definition may reference; unknown keys fail.
//
// Checks: initial_step exists; every step type is in the closed set; every
// transition target exists; every step is reachable from initial_step (no
// orphans); at least one terminal is reachable; every auto action key is
// registered; every resolver key is registered.
func (d Definition) Validate(autoActions, resolvers map[string]bool) error {
	var probs []string
	add := func(format string, a ...any) { probs = append(probs, fmt.Sprintf(format, a...)) }

	if d.Key == "" {
		add("key is required")
	}
	if d.Version <= 0 {
		add("version must be a positive integer")
	}
	if len(d.Steps) == 0 {
		add("definition has no steps")
		return joinProblems(d.Key, probs)
	}
	if d.InitialStep == "" {
		add("initial_step is required")
	} else if _, ok := d.Steps[d.InitialStep]; !ok {
		add("initial_step %q does not exist", d.InitialStep)
	}

	// Per-step validation: type, external refs, transition targets exist.
	for _, name := range sortedStepKeys(d.Steps) {
		step := d.Steps[name]
		if !validStepTypes[step.Type] {
			add("step %q has unknown type %q", name, step.Type)
		}
		for i, spec := range step.Assignees {
			if !validSpecKinds[spec.Kind] {
				add("step %q assignee[%d] has unknown kind %q", name, i, spec.Kind)
			}
			if spec.Kind == SpecResolver {
				if spec.Resolver == "" {
					add("step %q assignee[%d] resolver kind requires a resolver key", name, i)
				} else if !resolvers[spec.Resolver] {
					add("step %q references unregistered resolver %q", name, spec.Resolver)
				}
			}
		}
		if step.Type == StepAuto {
			if step.Action == "" {
				add("auto step %q requires an action", name)
			} else if !autoActions[step.Action] {
				add("auto step %q references unregistered auto-action %q", name, step.Action)
			}
		}
		if step.Type == StepApproval {
			// An approval step must define BOTH decision transitions, or a
			// reject would dead-end at runtime (review finding ARCH-64).
			if step.OnApprove == nil || step.OnApprove.target() == "" {
				add("approval step %q must define on_approve.next", name)
			}
			if step.OnReject == nil || step.OnReject.target() == "" {
				add("approval step %q must define on_reject.next", name)
			}
		}
		// Gateway conditions: Equals is `any` for YAML authoring flexibility,
		// but the COMPILED definition only accepts immutable scalars (third
		// closure audit 2026-07-17, F-10): a map, slice, pointer, or other
		// reference value would keep the registry's cloned definition aliased
		// to module-owned mutable memory — gateway routing could then change
		// after boot validation and race runtime readers. The framework only
		// accepts condition values it can compare, clone, and serialize
		// deterministically; anything else is unrepresentable, not cloned.
		for i, b := range step.Branches {
			if b.When == nil {
				continue
			}
			if b.When.Key == "" {
				add("step %q branch[%d]: when requires a key", name, i)
			}
			if !scalarConditionValue(b.When.Equals) {
				add("step %q branch[%d]: when.equals must be an immutable scalar (string, bool, or number), not %T", name, i, b.When.Equals)
			}
		}
		for _, tgt := range step.outgoing() {
			if _, ok := d.Steps[tgt]; !ok {
				add("step %q transitions to unknown step %q", name, tgt)
			}
		}
	}

	// Reachability from initial_step: orphans and terminal-reachability.
	if _, ok := d.Steps[d.InitialStep]; ok {
		reached := d.reachable()
		for _, name := range sortedStepKeys(d.Steps) {
			if !reached[name] {
				add("step %q is unreachable from initial_step (orphan)", name)
			}
		}
		terminalReached := false
		for name := range reached {
			if d.Steps[name].Type == StepTerminal {
				terminalReached = true
				break
			}
		}
		if !terminalReached {
			add("no terminal step is reachable from initial_step")
		}
	}

	return joinProblems(d.Key, probs)
}

// outgoing returns the step keys this step can transition to.
func (s Step) outgoing() []string {
	var out []string
	push := func(t *Transition) {
		if tgt := t.target(); tgt != "" {
			out = append(out, tgt)
		}
	}
	// Fail-closed by design (adjudicated, MATRIX CS-23): StepTerminal — and any
	// future unknown step type — has no outgoing transitions, so falling out of
	// this switch with an empty slice is the correct terminal behavior, not a
	// missed case. Do not convert to an exhaustive enumeration.
	//exhaustive:ignore
	switch s.Type {
	case StepApproval:
		push(s.OnApprove)
		push(s.OnReject)
	case StepTask:
		push(s.Next)
	case StepAuto:
		push(s.Next)
		push(s.OnError)
	case StepGateway:
		for _, b := range s.Branches {
			if b.Next != "" {
				out = append(out, b.Next)
			}
		}
	}
	return out
}

// reachable returns the set of step keys reachable from initial_step (BFS).
func (d Definition) reachable() map[string]bool {
	seen := map[string]bool{}
	queue := []string{d.InitialStep}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if seen[cur] {
			continue
		}
		seen[cur] = true
		step, ok := d.Steps[cur]
		if !ok {
			continue
		}
		for _, tgt := range step.outgoing() {
			if !seen[tgt] {
				queue = append(queue, tgt)
			}
		}
	}
	return seen
}

func sortedStepKeys(steps map[string]Step) []string {
	keys := make([]string, 0, len(steps))
	for k := range steps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func joinProblems(key string, probs []string) error {
	if len(probs) == 0 {
		return nil
	}
	sort.Strings(probs)
	return kerr.E(kerr.KindValidation, "workflow_definition_invalid",
		"workflow definition "+key+" invalid: "+strings.Join(probs, "; "))
}
