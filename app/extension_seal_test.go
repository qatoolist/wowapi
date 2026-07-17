package app_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/foundation/integration"
	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/retention"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/workflow"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// Closure-review regression (adversarial closure review 2026-07-17, F-10):
// after Boot returns, the extension model is SEALED for EVERY registry class —
// not only the collectors guarded by moduleContext.mustBeUnsealed. A retained
// module.Context hands out the live shared registries (routes, permissions,
// resources, events, jobs, rules, workflows, retention/document classes,
// hooks, templates, providers), and Booted intentionally exposes the live
// Router/Events/Jobs for serving; the runtime does LIVE lookups against Jobs
// (kernel/jobs/runner.go) and Events (relay dispatch), so an unsealed mutator
// would let post-boot code introduce routes, job kinds, or subscriptions that
// boot validation never saw. Every mutator must panic after boot.
func TestSealedExtensionModelRejectsEveryPostBootRegistration(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	var retained module.Context
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		retained = mc
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	noopHandler := func(http.ResponseWriter, *http.Request) {}
	mutations := []struct {
		name string
		fn   func()
	}{
		// Registries reachable from a retained module context.
		{"Routes.Handle", func() {
			retained.Routes().Handle(http.MethodGet, "/late", httpx.RouteMeta{}, noopHandler)
		}},
		{"Permissions.Register", func() { retained.Permissions().Register(authz.Permission{}) }},
		{"Resources.Register", func() { retained.Resources().Register("widgets", resource.TypeSpec{}) }},
		{"Events.Subscribe", func() { retained.Events().Subscribe("late.event", "late", nil) }},
		{"Jobs.RegisterKind", func() { retained.Jobs().RegisterKind("late.kind", nil, jobs.RetryPolicy{}) }},
		{"Rules.Register", func() { retained.Rules().Register("widgets", rules.Point{}) }},
		{"Workflows.RegisterDefinition", func() { _ = retained.Workflows().RegisterDefinition(workflow.Definition{}) }},
		{"Workflows.RegisterAutoAction", func() { retained.Workflows().RegisterAutoAction("late", nil) }},
		{"Workflows.RegisterAssigneeResolver", func() { retained.Workflows().RegisterAssigneeResolver("late", nil) }},
		{"RetentionClasses.Register", func() { retained.RetentionClasses().Register(retention.RecordClass{}) }},
		{"DocumentClasses.Register", func() { retained.DocumentClasses().Register("widgets", document.Class{}) }},
		{"DocumentHooks.OnFileUpload", func() { retained.DocumentHooks().OnFileUpload(nil) }},
		{"DocumentHooks.OnDocumentAccess", func() { retained.DocumentHooks().OnDocumentAccess(nil) }},
		{"NotifyTemplates.Register", func() { retained.NotifyTemplates().Register("widgets", notify.TemplateSpec{}) }},
		{"IntegrationProviders.Register", func() { retained.IntegrationProviders().Register("widgets", integration.Provider(nil)) }},
		// The live collectors Booted itself exposes for serving.
		{"Booted.Router.Handle", func() {
			booted.Router.Handle(http.MethodGet, "/late2", httpx.RouteMeta{}, noopHandler)
		}},
		{"Booted.Events.Subscribe", func() { booted.Events.Subscribe("late.event2", "late2", nil) }},
		{"Booted.Jobs.RegisterKind", func() { booted.Jobs.RegisterKind("late.kind2", nil, jobs.RetryPolicy{}) }},
	}
	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s after boot did not panic — the extension model is not sealed for this class", m.name)
				}
			}()
			m.fn()
		})
	}
}
