package appmodel

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// State represents the lifecycle state of the ApplicationModel compile process.
type State string

const (
	StateCollecting State = "collecting"
	StateValidating State = "validating"
	StateSealed     State = "sealed"
)

// ErrInvalidTransition is returned when a state-machine transition is invalid.
var ErrInvalidTransition = errors.New("invalid state-machine transition")

// ErrPostSealMutation is returned when attempting to register a declaration after the model has been sealed.
var ErrPostSealMutation = errors.New("cannot mutate application model after it has been sealed")

// ErrInvalidRegistrar is returned when a zero-value or uninitialized registrar is used.
var ErrInvalidRegistrar = errors.New("invalid registrar: zero value or uninitialized")

// PortDef represents a port definition.
type PortDef struct {
	ID    string
	Owner string
	Type  reflect.Type
}

// PortProvider represents an implementation provider for a port.
type PortProvider struct {
	ID    string
	Owner string
	Impl  any
}

// PortRequirement represents a module's requirement for a port.
type PortRequirement struct {
	ID    string
	Owner string
}

// ApplicationModel is the immutable compiled representation of the application.
type ApplicationModel struct {
	mu           sync.RWMutex
	state        State
	ports        map[string]PortDef
	providers    map[string]PortProvider
	requirements map[string][]PortRequirement
}

// Snapshot returns a deep copy of the ApplicationModel.
func (m *ApplicationModel) Snapshot() *ApplicationModel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snapshotNoLock()
}

func (m *ApplicationModel) snapshotNoLock() *ApplicationModel {
	ports := make(map[string]PortDef)
	for k, v := range m.ports {
		ports[k] = v
	}

	providers := make(map[string]PortProvider)
	for k, v := range m.providers {
		providers[k] = v
	}

	requirements := make(map[string][]PortRequirement)
	for k, v := range m.requirements {
		requirements[k] = append([]PortRequirement(nil), v...)
	}

	return &ApplicationModel{
		state:        m.state,
		ports:        ports,
		providers:    providers,
		requirements: requirements,
	}
}

// State returns the current state of the model.
func (m *ApplicationModel) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// Compiler accumulates module declarations through owner-bound calls and seals them.
type Compiler struct {
	mu    sync.Mutex
	model *ApplicationModel
}

// NewCompiler returns a new Compiler.
func NewCompiler() *Compiler {
	return &Compiler{
		model: &ApplicationModel{
			state:        StateCollecting,
			ports:        make(map[string]PortDef),
			providers:    make(map[string]PortProvider),
			requirements: make(map[string][]PortRequirement),
		},
	}
}

// privateToken prevents other packages from constructing a Registrar.
type privateToken struct{}

// Registrar is a capability type representing module ownership.
type Registrar[T any] struct {
	owner    string
	compiler *Compiler
	_        privateToken
}

// GetRegistrar mints a Registrar with immutable owner identity.
// This is the sole public mint authority.
func (c *Compiler) GetRegistrar(owner string) Registrar[any] {
	if owner == "" {
		panic("owner identity cannot be empty")
	}
	return Registrar[any]{
		owner:    owner,
		compiler: c,
	}
}

// Owner returns the owner of the Registrar.
func (r Registrar[T]) Owner() string {
	return r.owner
}

// DefinePort registers a port definition under the given key.
func (r Registrar[T]) DefinePort(id string, t reflect.Type) error {
	if r.owner == "" || r.compiler == nil {
		return ErrInvalidRegistrar
	}
	return r.compiler.definePort(r.owner, id, t)
}

// ProvidePort registers an implementation provider for a port.
func (r Registrar[T]) ProvidePort(id string, impl any, t reflect.Type) error {
	if r.owner == "" || r.compiler == nil {
		return ErrInvalidRegistrar
	}
	return r.compiler.providePort(r.owner, id, impl, t)
}

// RequirePort registers a requirement for a port.
func (r Registrar[T]) RequirePort(id string, t reflect.Type) error {
	if r.owner == "" || r.compiler == nil {
		return ErrInvalidRegistrar
	}
	return r.compiler.requirePort(r.owner, id, t)
}

// ResolvePort resolves the implementation of a port.
func (r Registrar[T]) ResolvePort(id string) (any, error) {
	if r.owner == "" || r.compiler == nil {
		return nil, ErrInvalidRegistrar
	}
	return r.compiler.resolvePort(r.owner, id)
}

// definePort registers a port definition under the given key.
func (c *Compiler) definePort(owner string, id string, t reflect.Type) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.mu.Lock()
	defer c.model.mu.Unlock()

	if c.model.state == StateSealed {
		return handlePostSealMutation(ErrPostSealMutation)
	}

	if id == "" {
		return errors.New("port ID cannot be empty")
	}

	if t == nil {
		return errors.New("port type cannot be nil")
	}

	if existing, exists := c.model.ports[id]; exists {
		if existing.Type != t {
			return fmt.Errorf("port %s type conflict: already defined with type %v, got %v", id, existing.Type, t)
		}
		return fmt.Errorf("port %s is already defined", id)
	}

	c.model.ports[id] = PortDef{ID: id, Owner: owner, Type: t}
	return nil
}

// providePort registers an implementation provider for a port.
func (c *Compiler) providePort(owner string, id string, impl any, t reflect.Type) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.mu.Lock()
	defer c.model.mu.Unlock()

	if c.model.state == StateSealed {
		return handlePostSealMutation(ErrPostSealMutation)
	}

	if id == "" {
		return errors.New("port ID cannot be empty")
	}

	if t == nil {
		return errors.New("port type cannot be nil")
	}

	// Verify that the port is defined and types match
	existing, exists := c.model.ports[id]
	if !exists {
		return fmt.Errorf("port %s is not defined", id)
	}
	if existing.Type != t {
		return fmt.Errorf("port %s type conflict: defined with type %v, provided with %v", id, existing.Type, t)
	}

	if _, exists := c.model.providers[id]; exists {
		return fmt.Errorf("port %s already has a provider", id)
	}

	c.model.providers[id] = PortProvider{ID: id, Owner: owner, Impl: impl}
	return nil
}

// requirePort registers a requirement for a port.
func (c *Compiler) requirePort(owner string, id string, t reflect.Type) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.mu.Lock()
	defer c.model.mu.Unlock()

	if c.model.state == StateSealed {
		return handlePostSealMutation(ErrPostSealMutation)
	}

	if id == "" {
		return errors.New("port ID cannot be empty")
	}

	if t == nil {
		return errors.New("port type cannot be nil")
	}

	// Verify that the port is defined and types match
	existing, exists := c.model.ports[id]
	if !exists {
		return fmt.Errorf("port %s is not defined", id)
	}
	if existing.Type != t {
		return fmt.Errorf("port %s type conflict: defined with type %v, required with %v", id, existing.Type, t)
	}

	reqs := c.model.requirements[id]
	reqs = append(reqs, PortRequirement{ID: id, Owner: owner})
	c.model.requirements[id] = reqs
	return nil
}

// resolvePort resolves the implementation of a port.
func (c *Compiler) resolvePort(_ string, id string) (any, error) {
	c.model.mu.RLock()
	defer c.model.mu.RUnlock()

	if c.model.state != StateSealed {
		return nil, errors.New("cannot resolve ports until the model is sealed")
	}

	// Verify that the port is defined
	_, exists := c.model.ports[id]
	if !exists {
		return nil, fmt.Errorf("port %s is not defined", id)
	}

	prov, exists := c.model.providers[id]
	if !exists {
		return nil, fmt.Errorf("no provider registered for port %s", id)
	}

	return prov.Impl, nil
}

// Compile validates then seals the accumulated declarations into an immutable ApplicationModel.
func (c *Compiler) Compile() (*ApplicationModel, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.model.mu.Lock()
	defer c.model.mu.Unlock()

	if c.model.state != StateCollecting {
		return nil, fmt.Errorf("%w: cannot compile from state %s", ErrInvalidTransition, c.model.state)
	}

	c.model.state = StateValidating

	// Run validation checks
	if err := c.validate(); err != nil {
		c.model.state = StateCollecting // Revert on failure
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	c.model.state = StateSealed
	return c.model.snapshotNoLock(), nil
}

// validate performs basic sanity and consistency checks over the accumulated declarations.
func (c *Compiler) validate() error {
	// Already called under c.model.mu Lock in Compile()
	// 1. Verify every provided port has a definition.
	for id, prov := range c.model.providers {
		if _, defined := c.model.ports[id]; !defined {
			return fmt.Errorf("port %s is provided by %s but has no port definition", id, prov.Owner)
		}
	}

	// 2. Verify every required port has a definition and a provider.
	for id, reqs := range c.model.requirements {
		if _, defined := c.model.ports[id]; !defined {
			return fmt.Errorf("port %s is required but has no port definition", id)
		}
		if _, provided := c.model.providers[id]; !provided {
			return fmt.Errorf("port %s is required by %s but has no provider", id, reqs[0].Owner)
		}
	}

	return nil
}
