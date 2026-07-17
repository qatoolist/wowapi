package workflow

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/qatoolist/wowapi/internal/sealer"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// AutoInput is what a registered auto-action receives: the instance context and
// its target resource. The returned map is merged into the instance context.
//
// Context is a deep canonical COPY of the instance context (mutating or
// retaining it never affects the framework), and its values are canonical
// JSON shapes: every NUMBER is a json.Number, never float64/int — assert
// numeric values as json.Number (n.Float64()/n.Int64()), not as native
// numeric types. Returned output is the only mutation channel and is
// canonicalized the same way before merging.
type AutoInput struct {
	InstanceID string
	Resource   resource.Ref
	Step       string
	Context    map[string]any
}

// AutoAction is a module Go action bound to an `auto` step. On success the
// runtime merges its output into the instance context and advances; on error it
// follows the step's on_error transition.
type AutoAction func(ctx context.Context, in AutoInput) (map[string]any, error)

// ResolveInput is what an assignee resolver receives when a task is created.
// Context follows the same contract as AutoInput.Context: a deep canonical
// copy whose numbers are json.Number values.
type ResolveInput struct {
	InstanceID string
	Resource   resource.Ref
	Step       string
	Context    map[string]any
}

// ResolvedKind is the concrete assignee kind persisted in
// workflow_task_assignees (capacity|role|relationship|system).
type ResolvedKind string

const (
	// KindCapacity addresses a specific acting capacity (the assignee-check unit).
	KindCapacity ResolvedKind = "capacity"
	// KindRole addresses anyone holding a role (authz-resolved at decide time).
	KindRole ResolvedKind = "role"
	// KindRelationship addresses relationship-holders on the resource.
	KindRelationship ResolvedKind = "relationship"
	// KindSystem addresses an automated principal.
	KindSystem ResolvedKind = "system"
)

// Assignee is a concrete, resolved assignee row for a task.
type Assignee struct {
	Kind ResolvedKind
	Ref  string
}

// AssigneeResolver resolves a `resolver`-kind AssigneeSpec into concrete
// assignees at task-creation time.
type AssigneeResolver func(ctx context.Context, in ResolveInput) ([]Assignee, error)

// Registry is the boot-time workflow catalog: definitions plus the module Go
// actions and assignee resolvers they reference. Like authz.Registry it
// accumulates registration errors and validates every definition in Err(), so a
// dangling transition or unknown auto-action fails boot, never a running
// instance (D-0053).
type Registry struct {
	// mu guards every map and the validation state: public-API callers may
	// interleave registration with runtime execution, and the runtime's read
	// paths must never race a mutation (fifth closure audit 2026-07-17).
	mu        sync.RWMutex
	defs      map[string]Definition // key\x00version -> def
	latest    map[string]int        // key -> highest registered version
	autos     map[string]AutoAction
	resolvers map[string]AssigneeResolver
	errs      []error
	sealed    bool
	// Validation state is GENERATION-KEYED (fifth closure audit 2026-07-17):
	// every mutation attempt bumps gen and clears validated at entry, so a
	// clean Err() result can never go stale — the runtime executes only when
	// the last clean validation covered the registry's current contents
	// (validated && validatedGen == gen).
	gen          uint64
	validatedGen uint64
	validated    bool
}

// Seal freezes the registry once boot validation completes: any later
// registration panics rather than silently adding a definition, auto-action,
// or resolver the boot gates never saw (closure review 2026-07-17, F-10).
// The sealer.Authority parameter restricts sealing to the framework's boot
// path: internal/sealer is unimportable outside the wowapi module, so a
// product module cannot prematurely seal a shared registry during Register.
func (r *Registry) Seal(sealer.Authority) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sealed = true
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		defs:      map[string]Definition{},
		latest:    map[string]int{},
		autos:     map[string]AutoAction{},
		resolvers: map[string]AssigneeResolver{},
	}
}

func defKey(key string, version int) string { return key + "\x00" + strconv.Itoa(version) }

// RegisterDefinition adds a definition. A duplicate (key, version) is an error.
// Full graph validation is deferred to Err() so it runs after all auto-actions
// and resolvers are registered.
func (r *Registry) RegisterDefinition(def Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// A rejected mutation must leave generation, contents, AND validation
	// UNCHANGED (sixth review, C-01): invalidate only once the mutation is
	// certain to take effect, never before the sealed-panic or a rejection —
	// a caller recovering the post-seal panic must not be able to strand a
	// previously validated registry as stale.
	if r.sealed {
		panic("workflow: definition registration after boot: the extension model is sealed")
	}
	if def.Key == "" || def.Version <= 0 {
		err := kerr.E(kerr.KindValidation, "workflow_definition_invalid",
			"workflow definition requires a non-empty key and positive version")
		r.errs = append(r.errs, err)
		return err
	}
	k := defKey(def.Key, def.Version)
	if _, dup := r.defs[k]; dup {
		err := kerr.E(kerr.KindConflict, "duplicate_workflow_definition",
			"workflow definition registered more than once: "+k)
		r.errs = append(r.errs, err)
		return err
	}
	// Gateway condition values are validated SYNCHRONOUSLY, before storage
	// (fourth closure audit 2026-07-17): full graph validation is deferred to
	// Err() because it needs the complete auto-action/resolver sets, but a
	// mutable Equals value must never be STORED — a caller on the public API
	// path that ignores Err() would otherwise retain an alias into the
	// registry's cloned definition and could steer gateway routing after
	// registration.
	for _, name := range sortedStepKeys(def.Steps) {
		step := def.Steps[name]
		for i, b := range step.Branches {
			if b.When == nil {
				continue
			}
			if b.When.Key == "" || !scalarConditionValue(b.When.Equals) {
				err := kerr.E(kerr.KindValidation, "workflow_condition_invalid",
					fmt.Sprintf("definition %s step %q branch[%d]: when requires a key and an immutable scalar equals value (string, bool, or number), not %T",
						def.Key, name, i, b.When.Equals))
				r.errs = append(r.errs, err)
				return err
			}
		}
	}
	// Invalidate only on the successful mutation (sixth review, C-01).
	r.gen++
	r.validated = false
	r.defs[k] = def.clone()
	if def.Version > r.latest[def.Key] {
		r.latest[def.Key] = def.Version
	}
	return nil
}

// clone returns a deep copy of d down to every nested step map, slice, and
// pointer: the registry must not alias a caller's Definition — a module
// mutating the value it registered (its Steps map, a Transition, a Policy)
// must never alter the validated graph running instances resolve against
// (second closure audit 2026-07-17, F-10).
func (d Definition) clone() Definition {
	out := d
	if d.Steps != nil {
		out.Steps = make(map[string]Step, len(d.Steps))
		for k, s := range d.Steps {
			out.Steps[k] = s.clone()
		}
	}
	return out
}

func (s Step) clone() Step {
	out := s
	if s.Assignees != nil {
		out.Assignees = append([]AssigneeSpec(nil), s.Assignees...)
	}
	if s.Policy != nil {
		p := *s.Policy
		if s.Policy.SelfApproval != nil {
			b := *s.Policy.SelfApproval
			p.SelfApproval = &b
		}
		out.Policy = &p
	}
	if s.SLA != nil {
		sla := *s.SLA
		out.SLA = &sla
	}
	out.OnApprove = s.OnApprove.clone()
	out.OnReject = s.OnReject.clone()
	out.Next = s.Next.clone()
	out.OnError = s.OnError.clone()
	if s.Branches != nil {
		out.Branches = make([]Branch, len(s.Branches))
		for i, b := range s.Branches {
			nb := b
			if b.When != nil {
				w := *b.When
				nb.When = &w
			}
			out.Branches[i] = nb
		}
	}
	if s.Electorate != nil {
		e := *s.Electorate
		out.Electorate = &e
	}
	if s.Quorum != nil {
		q := *s.Quorum
		out.Quorum = &q
	}
	if s.Pass != nil {
		p := *s.Pass
		out.Pass = &p
	}
	return out
}

func (t *Transition) clone() *Transition {
	if t == nil {
		return nil
	}
	c := *t
	return &c
}

// RegisterAutoAction binds a Go action to an auto step's action key.
func (r *Registry) RegisterAutoAction(key string, fn AutoAction) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		panic("workflow: auto-action registration after boot: the extension model is sealed")
	}
	if key == "" || fn == nil {
		r.errs = append(r.errs, kerr.E(kerr.KindValidation, "invalid_auto_action",
			"RegisterAutoAction requires a key and fn"))
		return
	}
	if _, dup := r.autos[key]; dup {
		r.errs = append(r.errs, kerr.E(kerr.KindConflict, "duplicate_auto_action",
			"auto-action registered more than once: "+key))
		return
	}
	// Invalidate only on the successful mutation (sixth review, C-01).
	r.gen++
	r.validated = false
	r.autos[key] = fn
}

// RegisterAssigneeResolver binds a resolver func to a resolver key.
func (r *Registry) RegisterAssigneeResolver(key string, fn AssigneeResolver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sealed {
		panic("workflow: assignee-resolver registration after boot: the extension model is sealed")
	}
	if key == "" || fn == nil {
		r.errs = append(r.errs, kerr.E(kerr.KindValidation, "invalid_resolver",
			"RegisterAssigneeResolver requires a key and fn"))
		return
	}
	if _, dup := r.resolvers[key]; dup {
		r.errs = append(r.errs, kerr.E(kerr.KindConflict, "duplicate_resolver",
			"assignee resolver registered more than once: "+key))
		return
	}
	// Invalidate only on the successful mutation (sixth review, C-01).
	r.gen++
	r.validated = false
	r.resolvers[key] = fn
}

// Err returns accumulated registration errors AND each definition's Validate()
// error, joined, or nil. It must gate boot.
func (r *Registry) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	autoKeys := make(map[string]bool, len(r.autos))
	for k := range r.autos {
		autoKeys[k] = true
	}
	resolverKeys := make(map[string]bool, len(r.resolvers))
	for k := range r.resolvers {
		resolverKeys[k] = true
	}

	var msgs []string
	for _, e := range r.errs {
		msgs = append(msgs, e.Error())
	}
	for k := range r.defs {
		if err := r.defs[k].Validate(autoKeys, resolverKeys); err != nil {
			msgs = append(msgs, err.Error())
		}
	}
	if len(msgs) == 0 {
		// Validation completed clean FOR THIS GENERATION: the runtime may
		// execute until the next mutation attempt (fourth/fifth closure
		// audits 2026-07-17).
		r.validated = true
		r.validatedGen = r.gen
		return nil
	}
	r.validated = false
	return kerr.E(kerr.KindInternal, "workflow_registration_failed",
		"workflow registration failed: "+strings.Join(msgs, "; "))
}

// validatedOK reports whether the LAST clean validation covers the registry's
// CURRENT contents; any mutation attempt since invalidates it.
func (r *Registry) validatedOK() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.validated && r.validatedGen == r.gen
}

// resolveValidated atomically returns the registered definition for
// (key, version) UNDER ONE READ LOCK together with the validation check
// (sixth review, C-02): a separate validatedOK() then definition() opens a
// window where a concurrent registration invalidates the registry between the
// two calls, letting an unvalidated graph execute. `validated` reports whether
// the current generation is validated; `found` whether the definition exists.
func (r *Registry) resolveValidated(key string, version int) (def Definition, validated, found bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.validated || r.validatedGen != r.gen {
		return Definition{}, false, false
	}
	d, ok := r.defs[defKey(key, version)]
	return d, true, ok
}

// callbackKeys returns the registered auto-action and resolver key sets, for
// validating PERSISTED definitions against the same semantic rules as
// registered ones (fifth closure audit 2026-07-17).
func (r *Registry) callbackKeys() (autos, resolvers map[string]bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	autos = make(map[string]bool, len(r.autos))
	for k := range r.autos {
		autos[k] = true
	}
	resolvers = make(map[string]bool, len(r.resolvers))
	for k := range r.resolvers {
		resolvers[k] = true
	}
	return autos, resolvers
}

// definition returns the registered definition for (key, version).
func (r *Registry) definition(key string, version int) (Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.defs[defKey(key, version)]
	return d, ok
}

// latestVersion returns the highest registered version for a key.
func (r *Registry) latestVersion(key string) (int, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.latest[key]
	return v, ok
}

func (r *Registry) auto(key string) (AutoAction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.autos[key]
	return fn, ok
}

func (r *Registry) resolver(key string) (AssigneeResolver, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.resolvers[key]
	return fn, ok
}
