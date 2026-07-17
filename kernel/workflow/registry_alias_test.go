package workflow

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/resource"
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

// Fourth closure-audit regression (2026-07-17): the PUBLIC API path — no
// App.Boot, no Err() — must reject every mutable Equals shape SYNCHRONOUSLY
// at RegisterDefinition, before storage, so an ignoring caller can never
// retain an alias into the stored definition.
func TestRegisterDefinitionRejectsMutableEqualsSynchronously(t *testing.T) {
	for name, equals := range map[string]any{
		"map":     map[string]string{"tier": "gold"},
		"slice":   []string{"gold"},
		"pointer": &struct{ V string }{"gold"},
		"func":    func() {},
		"nil":     nil,
		"struct":  struct{ V string }{"gold"},
	} {
		t.Run(name, func(t *testing.T) {
			r := NewRegistry()
			err := r.RegisterDefinition(Definition{
				Key: "widgets.gw", Version: 1, AppliesTo: "widgets.thing", InitialStep: "gate",
				Steps: map[string]Step{
					"gate": {Type: StepGateway, Branches: []Branch{
						{When: &Condition{Key: "tier", Equals: equals}, Next: "done"},
						{Next: "done"},
					}},
					"done": {Type: StepTerminal, Outcome: "ok"},
				},
			})
			if err == nil {
				t.Fatalf("RegisterDefinition returned nil for a %s when.equals — the alias is already stored", name)
			}
			if !strings.Contains(err.Error(), "immutable scalar") {
				t.Fatalf("error does not explain the scalar restriction: %v", err)
			}
			if _, ok := r.definition("widgets.gw", 1); ok {
				t.Fatal("rejected definition was stored anyway")
			}
		})
	}
}

// Fourth closure-audit regression (2026-07-17): the runtime refuses to execute
// against a registry that has not completed validation — an unvalidated OR
// invalid registry must fail before any instance work, closing the
// NewRuntime-without-Err() route around the boot gates.
func TestRuntimeRefusesUnvalidatedRegistry(t *testing.T) {
	valid := Definition{
		Key: "widgets.flow", Version: 1, AppliesTo: "widgets.thing", InitialStep: "done",
		Steps: map[string]Step{"done": {Type: StepTerminal, Outcome: "ok"}},
	}

	t.Run("never validated", func(t *testing.T) {
		r := NewRegistry()
		if err := r.RegisterDefinition(valid); err != nil {
			t.Fatal(err)
		}
		rt := &Runtime{registry: r}
		_, err := rt.StartIn(t.Context(), nil, "widgets.flow", resource.Ref{}, nil)
		if err == nil || !strings.Contains(err.Error(), "not completed validation") {
			t.Fatalf("StartIn on an unvalidated registry = %v, want the validation-gate error", err)
		}
	})

	t.Run("validation failed", func(t *testing.T) {
		r := NewRegistry()
		if err := r.RegisterDefinition(Definition{
			Key: "widgets.broken", Version: 1, InitialStep: "missing",
			Steps: map[string]Step{"done": {Type: StepTerminal, Outcome: "ok"}},
		}); err != nil {
			t.Fatal(err)
		}
		if err := r.Err(); err == nil {
			t.Fatal("broken definition passed validation")
		}
		rt := &Runtime{registry: r}
		if _, err := rt.StartIn(t.Context(), nil, "widgets.broken", resource.Ref{}, nil); err == nil ||
			!strings.Contains(err.Error(), "not completed validation") {
			t.Fatalf("StartIn on an invalid registry = %v, want the validation-gate error", err)
		}
	})

	t.Run("validated registry passes the gate", func(t *testing.T) {
		r := NewRegistry()
		if err := r.RegisterDefinition(valid); err != nil {
			t.Fatal(err)
		}
		if err := r.Err(); err != nil {
			t.Fatal(err)
		}
		rt := &Runtime{registry: r}
		if err := rt.requireValidated(); err != nil {
			t.Fatalf("validated registry rejected: %v", err)
		}
	})
}

type panickingStringer struct{}

func (panickingStringer) String() string { panic("String() must never be invoked for routing") }

// Fourth closure-audit regression (2026-07-17): gateway comparison is
// type-preserving and canonical — never fmt.Sprint. String "1" must not match
// number 1, "true" must not match true, named scalar types compare by kind,
// Stringer implementations are never invoked, and routing over the canonical
// context is identical before and after a JSON reload.
func TestGatewayComparisonIsCanonicalAndTypeSafe(t *testing.T) {
	type tier string
	rt := &Runtime{}
	step := func(equals any) Step {
		return Step{Type: StepGateway, Branches: []Branch{
			{When: &Condition{Key: "v", Equals: equals}, Next: "match"},
			{Next: "default"},
		}}
	}

	for name, tc := range map[string]struct {
		equals any
		ctxVal any
		want   string
	}{
		"string 1 vs number 1":        {equals: "1", ctxVal: float64(1), want: "default"},
		"number 1 vs string 1":        {equals: 1, ctxVal: "1", want: "default"},
		"string true vs bool true":    {equals: "true", ctxVal: true, want: "default"},
		"bool true vs bool true":      {equals: true, ctxVal: true, want: "match"},
		"int condition vs json float": {equals: 1, ctxVal: float64(1), want: "match"},
		"named scalar type":           {equals: "gold", ctxVal: tier("gold"), want: "match"},
		"stringer never invoked":      {equals: "gold", ctxVal: panickingStringer{}, want: "default"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := rt.gatewayTarget(step(tc.equals), map[string]any{"v": tc.ctxVal}); got != tc.want {
				t.Fatalf("routing %s: got %q, want %q", name, got, tc.want)
			}
		})
	}

	// Before/after-reload equivalence: routing over the canonicalized context
	// equals routing over its JSON round-trip for every scalar kind.
	in := map[string]any{"s": tier("gold"), "n": 7, "b": true, "f": 1.5}
	canon, err := canonicalizeContext(in)
	if err != nil {
		t.Fatal(err)
	}
	reloadedRaw, _ := json.Marshal(canon)
	var reloaded map[string]any
	if err := json.Unmarshal(reloadedRaw, &reloaded); err != nil {
		t.Fatal(err)
	}
	for key, equals := range map[string]any{"s": "gold", "n": 7, "b": true, "f": 1.5} {
		st := Step{Type: StepGateway, Branches: []Branch{
			{When: &Condition{Key: key, Equals: equals}, Next: "match"},
			{Next: "default"},
		}}
		before := rt.gatewayTarget(st, canon)
		after := rt.gatewayTarget(st, reloaded)
		if before != "match" || after != "match" {
			t.Fatalf("key %s: routing diverges or misses across reload: before=%q after=%q", key, before, after)
		}
	}
}

// Fifth closure-audit regression (2026-07-17): a clean validation result must
// never go STALE — every mutation attempt (all three registration methods)
// invalidates it, and the runtime refuses to execute until a NEW clean
// validation pass covers the current contents.
func TestValidationInvalidatedByEveryMutation(t *testing.T) {
	valid := func(key string) Definition {
		return Definition{
			Key: key, Version: 1, AppliesTo: "widgets.thing", InitialStep: "done",
			Steps: map[string]Step{"done": {Type: StepTerminal, Outcome: "ok"}},
		}
	}
	mutations := []struct {
		name string
		fn   func(r *Registry, i int)
	}{
		{"RegisterDefinition", func(r *Registry, i int) {
			_ = r.RegisterDefinition(valid("widgets.mut" + strconv.Itoa(i)))
		}},
		{"RegisterAutoAction", func(r *Registry, i int) {
			r.RegisterAutoAction("late.act"+strconv.Itoa(i), func(context.Context, AutoInput) (map[string]any, error) { return nil, nil })
		}},
		{"RegisterAssigneeResolver", func(r *Registry, i int) {
			r.RegisterAssigneeResolver("late.res"+strconv.Itoa(i), func(context.Context, ResolveInput) ([]Assignee, error) { return nil, nil })
		}},
	}
	r := NewRegistry()
	if err := r.RegisterDefinition(valid("widgets.base")); err != nil {
		t.Fatal(err)
	}
	rt := &Runtime{registry: r}
	for i, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			if err := r.Err(); err != nil {
				t.Fatalf("validation: %v", err)
			}
			if err := rt.requireValidated(); err != nil {
				t.Fatalf("freshly validated registry rejected: %v", err)
			}
			m.fn(r, i)
			if err := rt.requireValidated(); err == nil {
				t.Fatalf("%s did not invalidate the previous clean validation — stale validated state", m.name)
			}
			if _, err := rt.StartIn(t.Context(), nil, "widgets.base", resource.Ref{}, nil); err == nil ||
				!strings.Contains(err.Error(), "not completed validation") {
				t.Fatalf("StartIn after %s = %v, want the validation-gate error until revalidation", m.name, err)
			}
		})
	}
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rt.requireValidated(); err != nil {
		t.Fatalf("revalidated registry rejected: %v", err)
	}
}

// Fifth closure-audit regression (2026-07-17): concurrent registration and
// execution-path reads must be race-free (the registry is RWMutex-guarded) —
// run under the -race gate.
func TestConcurrentRegistrationAndExecutionIsSafe(t *testing.T) {
	r := NewRegistry()
	rt := &Runtime{registry: r}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := range 200 {
			_ = r.RegisterDefinition(Definition{
				Key: "widgets.c" + strconv.Itoa(i), Version: 1, AppliesTo: "widgets.thing", InitialStep: "done",
				Steps: map[string]Step{"done": {Type: StepTerminal, Outcome: "ok"}},
			})
			if i%10 == 0 {
				_ = r.Err()
			}
		}
	}()
	for i := range 200 {
		_ = rt.requireValidated()
		_, _ = r.definition("widgets.c"+strconv.Itoa(i%50), 1)
		_, _ = r.latestVersion("widgets.c0")
		_, _ = r.auto("none")
		_, _ = r.resolver("none")
	}
	<-done
}

// Fifth closure-audit regression (2026-07-17): numeric comparison is EXACT —
// json.Number context values against big.Rat condition semantics. Boundaries
// the float64 model got wrong: integers beyond 2^53, int64/uint64 maxima,
// float32 shortest-decimal, NaN/Inf.
func TestNumericComparisonExactBoundaries(t *testing.T) {
	for name, tc := range map[string]struct {
		ctxVal any
		equals any
		want   bool
	}{
		"2^53 equals itself":            {json.Number("9007199254740992"), int64(9007199254740992), true},
		"2^53+1 distinct from 2^53":     {json.Number("9007199254740993"), int64(9007199254740992), false},
		"2^53+1 equals itself":          {json.Number("9007199254740993"), int64(9007199254740993), true},
		"max int64 exact":               {json.Number("9223372036854775807"), int64(9223372036854775807), true},
		"max int64 vs max-1":            {json.Number("9223372036854775807"), int64(9223372036854775806), false},
		"max uint64 exact":              {json.Number("18446744073709551615"), uint64(18446744073709551615), true},
		"float32(0.1) matches JSON 0.1": {json.Number("0.1"), float32(0.1), true},
		"float64 0.1 matches JSON 0.1":  {json.Number("0.1"), 0.1, true},
		"NaN never matches":             {json.Number("1"), math.NaN(), false},
		"+Inf never matches":            {json.Number("1"), math.Inf(1), false},
		"scientific notation exact":     {json.Number("1e3"), int64(1000), true},
	} {
		t.Run(name, func(t *testing.T) {
			if got := conditionMatches(tc.ctxVal, tc.equals); got != tc.want {
				t.Fatalf("conditionMatches(%v, %v) = %v, want %v", tc.ctxVal, tc.equals, got, tc.want)
			}
		})
	}

	// NaN/Inf are also unrepresentable in declarations: registration rejects.
	for name, v := range map[string]any{"NaN": math.NaN(), "+Inf": math.Inf(1), "-Inf": math.Inf(-1)} {
		r := NewRegistry()
		err := r.RegisterDefinition(Definition{
			Key: "widgets.gw", Version: 1, AppliesTo: "widgets.thing", InitialStep: "gate",
			Steps: map[string]Step{
				"gate": {Type: StepGateway, Branches: []Branch{
					{When: &Condition{Key: "v", Equals: v}, Next: "done"}, {Next: "done"},
				}},
				"done": {Type: StepTerminal, Outcome: "ok"},
			},
		})
		if err == nil {
			t.Fatalf("%s condition value passed registration", name)
		}
	}

	// Reload equivalence at the boundary: a big integer survives the canonical
	// round trip bit-for-bit.
	canon, err := canonicalizeContext(map[string]any{"n": int64(9007199254740993)})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(canon)
	reloaded, err := decodeCanonicalContext(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !conditionMatches(reloaded["n"], int64(9007199254740993)) ||
		conditionMatches(reloaded["n"], int64(9007199254740992)) {
		t.Fatalf("2^53+1 lost precision across the canonical round trip: %v", reloaded["n"])
	}
}
