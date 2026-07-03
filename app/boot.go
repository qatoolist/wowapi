package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// Booted is the result of App.Boot: everything the process layer needs to
// start serving (or to migrate/seed). Modules have registered; the whole graph
// and registries are validated; nothing has started yet.
type Booted struct {
	Kernel     *kernel.Kernel
	Router     *httpx.Router
	Events     *outbox.HandlerRegistry // event subscriptions (drives the relay)
	Jobs       *jobs.Registry          // job kinds (drives the worker pools)
	OpenAPI    map[string][]byte
	Health     map[string]func(context.Context) error
	Migrations map[string]fs.FS
	Seeds      seeds.Bundle // merged catalog seeds, ready for SeedSync
}

// Boot runs the module lifecycle up to (not including) Start: it registers every
// module against a capability-scoped context built from k — in dependency order
// so a module's ports are available to its dependents — then validates the whole
// graph and the shared registries (blueprint 06 §2). Boot fails, before anything
// serves, on: a module graph error (dup/unknown/cycle), a registration error, a
// route whose permission is not registered, a duplicate/invalid permission or
// resource type, or a seed ownership/parse error.
//
// namespaces is the loaded product config's module.* subtree; each module sees
// only its own slice via Context.Config().
func (a *App) Boot(ctx context.Context, k *kernel.Kernel, namespaces config.Namespaces) (*Booted, error) {
	ordered, err := a.validateAndOrder()
	if err != nil {
		return nil, err
	}

	boot := newBootState()
	router := httpx.NewRouter()
	val := validation.New()
	idgen := model.UUIDv7()
	events := outbox.NewHandlerRegistry()
	writer := outbox.NewWriter(idgen)
	jobReg := jobs.NewRegistry()

	var regErrs []error
	for _, m := range ordered {
		var view config.ModuleView
		if namespaces != nil {
			if v, ok := namespaces[m.Name()]; ok {
				view = v
			}
		}
		mc := newModuleContext(m.Name(), k.Log, view, moduleDeps{
			router: router, val: val, perms: k.Perms, rtypes: k.Resources,
			eval: k.Authz, tx: k.Tx, idgen: idgen,
			events: events, writer: writer, jobs: jobReg,
			rules: k.Rules, resolver: k.RulesResolver, wfReg: k.Workflows, wfRT: k.WorkflowRuntime,
			docClass: k.DocumentClasses, docHooks: k.DocumentHooks, docs: k.Documents,
			comments: k.Comments, attaches: k.Attachments,
			boot: boot,
		})
		if err := m.Register(mc); err != nil {
			regErrs = append(regErrs, fmt.Errorf("module %q: Register: %w", m.Name(), err))
		}
	}

	// Load and merge each module's seed bundle (strict, ownership-checked), and
	// register seed-declared permissions into the shared registry so the
	// evaluator recognizes them. Iterate module names in sorted order so the
	// merged bundle and any error messages are deterministic (ARCH-52).
	var bundle seeds.Bundle
	seedModules := make([]string, 0, len(boot.seeds))
	for name := range boot.seeds {
		seedModules = append(seedModules, name)
	}
	sort.Strings(seedModules)
	for _, name := range seedModules {
		fsys := boot.seeds[name]
		b, err := seeds.Load(fsys, name)
		if err != nil {
			regErrs = append(regErrs, err)
			continue
		}
		bundle.Permissions = append(bundle.Permissions, b.Permissions...)
		bundle.Roles = append(bundle.Roles, b.Roles...)
		bundle.ResourceTypes = append(bundle.ResourceTypes, b.ResourceTypes...)
		bundle.RelationshipTypes = append(bundle.RelationshipTypes, b.RelationshipTypes...)
		for _, p := range b.Permissions {
			k.Perms.Register(authz.Permission{Key: p.Key, Sensitive: p.Sensitive, GrantedVia: p.GrantedVia})
		}
		for _, rt := range b.ResourceTypes {
			k.Resources.Register(name, resource.TypeSpec{Key: rt.Key, Description: rt.Description})
		}
	}

	// Whole-graph validation gates (all accumulated, boot fails with the full list).
	if err := k.Perms.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.Resources.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := router.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := events.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := jobReg.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.Rules.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.Workflows.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.DocumentClasses.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	// A module that registered a document class needs a document service to use
	// it; the service is nil when no object-storage adapter was wired. Fail boot
	// loudly rather than hand modules a nil Documents() at runtime.
	if len(k.DocumentClasses.Keys()) > 0 && k.Documents == nil {
		regErrs = append(regErrs, fmt.Errorf("document classes are registered (%v) but no storage adapter is wired: pass kernel.Deps.Storage", k.DocumentClasses.Keys()))
	}
	// Every route's permission must be a registered permission (deny-by-default
	// depends on the registry knowing it; an unknown permission is a boot bug).
	for _, p := range router.Permissions() {
		if !k.Perms.Has(p) {
			regErrs = append(regErrs, fmt.Errorf("route permission %q is not declared by any module seed or registration", p))
		}
	}

	if len(regErrs) > 0 {
		return nil, fmt.Errorf("app: boot validation failed: %w", errors.Join(regErrs...))
	}

	return &Booted{
		Kernel:     k,
		Router:     router,
		Events:     events,
		Jobs:       jobReg,
		OpenAPI:    boot.openapi,
		Health:     boot.health,
		Migrations: boot.migrations,
		Seeds:      bundle,
	}, nil
}
