package app

import (
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

// Readiness assembles the /readyz aggregator from the booted app: every module's
// registered readiness check (ctx.Health) plus the framework checks the caller
// supplies as `extra` (typically a DB ping and a "migrations current" probe,
// which the composition root wires because it owns the pool). The redacted config
// fingerprint is reported in the response for drift correlation. Mount
// h.Liveness() at /healthz and h.Readiness() at /readyz in the product's api and
// worker mains (blueprint 07 §9).
func Readiness(b *Booted, fingerprint config.Fingerprint, extra map[string]httpx.HealthCheck) *httpx.Health {
	h := httpx.NewHealth(fingerprint.String())
	for name, chk := range b.Health {
		h.Register("module."+name, chk)
	}
	for name, chk := range extra {
		h.Register(name, chk)
	}
	return h
}
