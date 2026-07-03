// Package integration is wowapi's external-provider framework: modules register
// a Provider ADAPTER per provider key (a payment gateway, an SMS gateway, an
// identity source, …); per-tenant/platform rows in integration_providers hold
// the non-secret config plus a credential REFERENCE (never plaintext); and the
// kernel resolves an adapter + its config + its resolved credential on demand
// and aggregates provider health for readiness.
//
// Anti-corruption boundary: the adapter is where a provider's payloads are
// translated to kernel/module types — provider types never cross into services.
// Contract: blueprint 07 §6.
package integration

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Kind classifies a provider; the set is closed so a typo fails registration.
var validKinds = map[string]bool{
	"payment": true, "messaging": true, "identity": true, "storage": true, "device": true,
}

// keyRE constrains provider keys to module.name.
var keyRE = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// Config is a resolved provider instance: identity, non-secret settings, and the
// resolved credential (from the row's credential_ref via the secrets provider).
// Credential is never logged.
type Config struct {
	Key        string
	Kind       string
	Settings   map[string]any
	Credential string // resolved secret value; "" when the row has no credential_ref
	IsPlatform bool   // true when resolved from the platform (tenant_id NULL) row
}

// Provider is the adapter a module registers for one provider key. Concrete
// per-kind behavior lives behind interfaces the module defines; the kernel needs
// only identity and a health probe.
type Provider interface {
	Key() string  // "core.stripe" — module.name
	Kind() string // one of validKinds
	// HealthCheck probes the live provider using the resolved config. A non-nil
	// error marks the provider degraded (surfaced in readiness detail, non-fatal).
	HealthCheck(ctx context.Context, cfg Config) error
}

// Registry collects provider adapters during module registration.
type Registry struct {
	providers map[string]Provider
	errs      []error
}

// NewRegistry returns an empty provider registry.
func NewRegistry() *Registry { return &Registry{providers: map[string]Provider{}} }

// Register adds a provider adapter. A malformed/foreign-module key, an invalid
// kind, or a duplicate is recorded and surfaced by Err().
func (r *Registry) Register(module string, p Provider) {
	key := p.Key()
	if !keyRE.MatchString(key) {
		r.errf("provider key must be module.name: %s", key)
		return
	}
	if prefix := module + "."; len(key) <= len(prefix) || key[:len(prefix)] != prefix {
		r.errf("module %s may not register provider %s", module, key)
		return
	}
	if !validKinds[p.Kind()] {
		r.errf("provider %s has invalid kind %q", key, p.Kind())
		return
	}
	if _, dup := r.providers[key]; dup {
		r.errf("provider registered more than once: %s", key)
		return
	}
	r.providers[key] = p
}

// Get returns the registered adapter for key.
func (r *Registry) Get(key string) (Provider, bool) { p, ok := r.providers[key]; return p, ok }

// Keys returns registered provider keys, sorted.
func (r *Registry) Keys() []string {
	out := make([]string, 0, len(r.providers))
	for k := range r.providers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (r *Registry) errf(format string, args ...any) {
	r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_provider", fmt.Sprintf(format, args...)))
}

// Err returns accumulated registration errors joined, or nil.
func (r *Registry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	msgs := make([]string, len(r.errs))
	for i, e := range r.errs {
		msgs[i] = e.Error()
	}
	joined := msgs[0]
	for i := 1; i < len(msgs); i++ {
		joined += "; " + msgs[i]
	}
	return kerr.E(kerr.KindInternal, "provider_registration_failed", "provider registration failed: "+joined)
}
