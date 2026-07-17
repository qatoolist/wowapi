// Package app is wowapi's composition root helpers. Product binaries
// (cmd/api, cmd/worker, cmd/migrate in the product repo) construct an App,
// register their modules, and run it.
//
// Phase 0 ships registration + whole-graph validation (duplicate names,
// invalid names, unknown dependencies, dependency cycles) and the
// deterministic registration order. Kernel construction and the
// RunAPI/RunWorker/RunMigrate helpers land in Phases 1–2.
package app

import (
	"errors"
	"fmt"
	"slices"

	"github.com/qatoolist/wowapi/v2/module"
)

// App holds the registered module set. Construct with New, add modules with
// Register, then call Validate before starting anything.
type App struct {
	modules []module.Module
}

// New returns an empty App.
func New() *App { return &App{} }

// Register adds modules. Order of Register calls does not matter: startup
// order is derived from each module's DependsOn graph during Validate.
func (a *App) Register(ms ...module.Module) {
	a.modules = append(a.modules, ms...)
}

// Validate performs whole-graph checks and returns ALL problems joined.
// It must pass before Start*; every failure names the offending module.
func (a *App) Validate() error {
	_, err := a.validateAndOrder()
	return err
}

// Ordered returns modules in dependency order (dependencies first, ties
// broken alphabetically for determinism), validating the graph on the way.
func (a *App) Ordered() ([]module.Module, error) {
	return a.validateAndOrder()
}

// validateAndOrder runs the whole-graph checks and the topo-sort exactly once
// (review finding ARCH-4: Ordered previously validated + sorted twice).
func (a *App) validateAndOrder() ([]module.Module, error) {
	var errs []error

	byName := make(map[string]module.Module, len(a.modules))
	for _, m := range a.modules {
		name := m.Name()
		if !module.ValidName(name) {
			errs = append(errs, fmt.Errorf("module %q: invalid name (want %s)", name, "[a-z][a-z0-9_]{0,63}"))
			continue
		}
		if _, dup := byName[name]; dup {
			errs = append(errs, fmt.Errorf("module %q: registered more than once", name))
			continue
		}
		byName[name] = m
	}

	for _, m := range a.modules {
		for _, dep := range m.DependsOn() {
			if dep == m.Name() {
				errs = append(errs, fmt.Errorf("module %q: depends on itself", m.Name()))
				continue
			}
			if _, ok := byName[dep]; !ok {
				errs = append(errs, fmt.Errorf("module %q: depends on unknown module %q", m.Name(), dep))
			}
		}
	}

	if len(errs) == 0 {
		ordered, err := order(byName)
		if err == nil {
			return ordered, nil
		}
		errs = append(errs, err)
	}
	return nil, errors.Join(errs...)
}

// order topo-sorts via DFS with cycle detection.
func order(byName map[string]module.Module) ([]module.Module, error) {
	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	slices.Sort(names)

	const (
		white = 0 // unvisited
		grey  = 1 // on stack
		black = 2 // done
	)
	state := make(map[string]int, len(byName))
	out := make([]module.Module, 0, len(byName))

	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		switch state[name] {
		case black:
			return nil
		case grey:
			return fmt.Errorf("module dependency cycle: %v", append(path, name))
		}
		state[name] = grey
		deps := slices.Clone(byName[name].DependsOn())
		slices.Sort(deps)
		for _, dep := range deps {
			if err := visit(dep, append(path, name)); err != nil {
				return err
			}
		}
		state[name] = black
		out = append(out, byName[name])
		return nil
	}

	for _, n := range names {
		if err := visit(n, nil); err != nil {
			return nil, err
		}
	}
	return out, nil
}
