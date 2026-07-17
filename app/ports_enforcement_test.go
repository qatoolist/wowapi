package app_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/module"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// F-10 regressions (adversarial-framework-review-2026-07-17): the public
// extension contract (owner-prefixed ports, boot-checked) must be ENFORCED by
// the production boot path via the ownership-bound compiler — not left as
// unconditional map writes that any module can overwrite or bypass.

func bootModules(t *testing.T, mods ...module.Module) error {
	t.Helper()
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	a.Register(mods...)
	_, err = a.Boot(context.Background(), k, nil)
	return err
}

type widgetPort interface{ Tick() int }

type widgetImpl struct{}

func (widgetImpl) Tick() int { return 1 }

func TestBootRejectsForeignPrefixPortProvider(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets"},
		funcModule{name: "gadgets", reg: func(mc module.Context) error {
			mc.ProvidePort("widgets.clock", widgetImpl{}) // another module's prefix
			return nil
		}},
	)
	if err == nil {
		t.Fatal("Boot accepted a port provided under another module's prefix")
	}
	if !strings.Contains(err.Error(), "widgets.clock") {
		t.Fatalf("boot error does not name the offending port: %v", err)
	}
}

func TestBootRejectsDuplicatePortProvider(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			mc.ProvidePort("widgets.clock", widgetImpl{})
			mc.ProvidePort("widgets.clock", widgetImpl{}) // overwrite attempt
			return nil
		}},
	)
	if err == nil {
		t.Fatal("Boot accepted the same port provided twice (silent overwrite)")
	}
}

func TestBootRejectsNilPortImpl(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			mc.ProvidePort("widgets.clock", nil)
			return nil
		}},
	)
	if err == nil {
		t.Fatal("Boot accepted a nil port implementation")
	}
}

func TestPortResolutionRequiresDeclaredDependency(t *testing.T) {
	// gadgets does NOT declare widgets in DependsOn but resolves its port.
	// Registration order still makes the port available in the raw map, so
	// only dependency validation can reject this.
	var resolveErr error
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			mc.ProvidePort("widgets.clock", widgetImpl{})
			return nil
		}},
		// "zgadgets" sorts after "widgets": it registers second, so the port IS
		// in the raw map — only dependency validation can reject the resolve.
		funcModule{name: "zgadgets", reg: func(mc module.Context) error {
			_, resolveErr = mc.Port("widgets.clock")
			return nil
		}},
	)
	_ = err
	if resolveErr == nil {
		t.Fatal("Port resolved from a module that never declared the provider as a dependency")
	}
}

func TestPortResolutionWorksForDeclaredDependency(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			mc.ProvidePort("widgets.clock", widgetImpl{})
			return nil
		}},
		funcModule{name: "gadgets", deps: []string{"widgets"}, reg: func(mc module.Context) error {
			p, err := mc.Port("widgets.clock")
			if err != nil {
				return err
			}
			if p.(widgetPort).Tick() != 1 {
				t.Fatal("resolved port does not work")
			}
			return nil
		}},
	)
	if err != nil {
		t.Fatalf("legal dependency port flow failed: %v", err)
	}
}

func TestRetainedContextCannotMutatePortsAfterBoot(t *testing.T) {
	var retained module.Context
	if err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			retained = mc
			return nil
		}},
	); err != nil {
		t.Fatalf("boot: %v", err)
	}
	mustPanic := func(what string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatalf("%s on a retained context after boot did not panic — extensions are mutable post-boot", what)
			}
		}()
		fn()
	}
	mustPanic("ProvidePort", func() { retained.ProvidePort("widgets.late", widgetImpl{}) })
	// Health is the sharpest post-boot surface: its map is concurrently read by
	// the live health handler.
	mustPanic("Health", func() { retained.Health("late", func(context.Context) error { return nil }) })
	mustPanic("Migrations", func() { retained.Migrations(nil) })
	mustPanic("OpenAPI", func() { retained.OpenAPI([]byte("{}")) })
}

// Closure-review regressions (adversarial closure review 2026-07-17, F-10).

// A module that IGNORES the Port error and returns nil from Register must
// still fail boot: unsatisfied dependencies are a boot contract, not a
// courtesy return value.
func TestBootFailsWhenPortErrorIsIgnored(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			mc.ProvidePort("widgets.clock", widgetImpl{})
			return nil
		}},
		funcModule{name: "zgadgets", reg: func(mc module.Context) error {
			_, _ = mc.Port("widgets.clock") // undeclared dependency; error ignored
			return nil
		}},
	)
	if err == nil {
		t.Fatal("boot succeeded although a module resolved an undeclared dependency and swallowed the error")
	}
}

// A missing provider must fail boot even when the module ignores the error.
func TestBootFailsWhenMissingProviderErrorIsIgnored(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			_, _ = mc.Port("widgets.absent") // never provided; error ignored
			return nil
		}},
	)
	if err == nil {
		t.Fatal("boot succeeded although a module resolved a missing port and swallowed the error")
	}
}

// A typed nil passes impl == nil but panics at first use — boot must reject it.
func TestBootRejectsTypedNilPortImpl(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			mc.ProvidePort("widgets.clock", (*widgetImpl)(nil))
			return nil
		}},
	)
	if err == nil {
		t.Fatal("boot accepted a typed-nil port implementation")
	}
}

// Duplicate collector registrations must fail boot, never silently overwrite.
func TestBootRejectsDuplicateCollectorRegistrations(t *testing.T) {
	err := bootModules(t,
		funcModule{name: "widgets", reg: func(mc module.Context) error {
			mc.OpenAPI([]byte(`{"a":1}`))
			mc.OpenAPI([]byte(`{"b":2}`)) // would previously overwrite {"a":1}
			mc.Health("db", func(context.Context) error { return nil })
			mc.Health("db", func(context.Context) error { return nil })
			return nil
		}},
	)
	if err == nil {
		t.Fatal("boot accepted duplicate OpenAPI/Health registrations (silent overwrite)")
	}
	for _, want := range []string{"duplicate OpenAPI", "duplicate Health"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("boot error %v does not name %q", err, want)
		}
	}
}

// The live readiness handler serves the boot-validated module checks. V2:
// Booted has no exposed Health map to inject into — the former sabotage
// vector is structurally impossible; this asserts the wiring itself.
func TestReadinessHandlerServesBootValidatedChecks(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Health("real", func(context.Context) error { return nil })
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	health := app.Readiness(booted, config.Fingerprint{}, nil)
	rec := httptest.NewRecorder()
	health.Readiness().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if !strings.Contains(rec.Body.String(), "widgets.real") {
		t.Fatalf("readiness payload lost the genuine module check: %s", rec.Body.String())
	}
}
