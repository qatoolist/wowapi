package httpx

import (
	"context"
	"net/http"
	"sort"
	"time"
)

// health.go — liveness (/healthz) and readiness (/readyz) endpoints (blueprint
// 07 §9). Liveness answers "is the process up"; readiness answers "can it serve"
// by running registered checks (DB ping, migrations current, registries
// validated, config valid, module checks) and reporting the redacted config
// fingerprint. The checks are plain funcs so this package stays free of a
// database/config import — the composition root supplies them as closures.

// HealthCheck reports readiness for one subsystem; a non-nil error = not ready.
type HealthCheck func(context.Context) error

// Health aggregates readiness checks and serves the two health endpoints.
type Health struct {
	fingerprint  string
	checkTimeout time.Duration
	checks       map[string]HealthCheck
}

// NewHealth builds a health aggregator. fingerprint is the redacted config
// fingerprint (a hash — safe to expose) reported by /readyz.
func NewHealth(fingerprint string) *Health {
	return &Health{fingerprint: fingerprint, checkTimeout: 3 * time.Second, checks: map[string]HealthCheck{}}
}

// Register adds a named readiness check (chainable). A nil check is ignored.
func (h *Health) Register(name string, c HealthCheck) *Health {
	if c != nil {
		h.checks[name] = c
	}
	return h
}

// Liveness answers 200 as long as the process is running — it runs NO checks
// (a failing dependency must not make the process get killed by a liveness probe;
// that is readiness' job).
func (h *Health) Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	}
}

// Readiness runs every check (each bounded by checkTimeout) and returns 200 when
// all pass, 503 otherwise, with a per-check status map + the config fingerprint.
func (h *Health) Readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := make(map[string]string, len(h.checks))
		ready := true
		for _, name := range sortedNames(h.checks) {
			ctx, cancel := context.WithTimeout(r.Context(), h.checkTimeout)
			err := h.checks[name](ctx)
			cancel()
			if err != nil {
				ready = false
				results[name] = "error: " + err.Error()
			} else {
				results[name] = "ok"
			}
		}
		status := http.StatusOK
		state := "ready"
		if !ready {
			status = http.StatusServiceUnavailable
			state = "not_ready"
		}
		WriteJSON(w, status, map[string]any{
			"status":             state,
			"config_fingerprint": h.fingerprint,
			"checks":             results,
		})
	}
}

func sortedNames(m map[string]HealthCheck) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
