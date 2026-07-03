package workflow

import (
	"context"
	"strconv"
	"strings"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/resource"
)

// AutoInput is what a registered auto-action receives: the instance context and
// its target resource. The returned map is merged into the instance context.
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
	defs      map[string]Definition // key\x00version -> def
	latest    map[string]int        // key -> highest registered version
	autos     map[string]AutoAction
	resolvers map[string]AssigneeResolver
	errs      []error
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
	r.defs[k] = def
	if def.Version > r.latest[def.Key] {
		r.latest[def.Key] = def.Version
	}
	return nil
}

// RegisterAutoAction binds a Go action to an auto step's action key.
func (r *Registry) RegisterAutoAction(key string, fn AutoAction) {
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
	r.autos[key] = fn
}

// RegisterAssigneeResolver binds a resolver func to a resolver key.
func (r *Registry) RegisterAssigneeResolver(key string, fn AssigneeResolver) {
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
	r.resolvers[key] = fn
}

// Err returns accumulated registration errors AND each definition's Validate()
// error, joined, or nil. It must gate boot.
func (r *Registry) Err() error {
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
		return nil
	}
	return kerr.E(kerr.KindInternal, "workflow_registration_failed",
		"workflow registration failed: "+strings.Join(msgs, "; "))
}

// definition returns the registered definition for (key, version).
func (r *Registry) definition(key string, version int) (Definition, bool) {
	d, ok := r.defs[defKey(key, version)]
	return d, ok
}

// latestVersion returns the highest registered version for a key.
func (r *Registry) latestVersion(key string) (int, bool) {
	v, ok := r.latest[key]
	return v, ok
}

func (r *Registry) auto(key string) (AutoAction, bool) { fn, ok := r.autos[key]; return fn, ok }
func (r *Registry) resolver(key string) (AssigneeResolver, bool) {
	fn, ok := r.resolvers[key]
	return fn, ok
}
