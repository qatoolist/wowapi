// Package secrets defines the secret-reference model and the provider port.
//
// Secrets never appear as raw values in configuration files or environment
// variables: they appear as references ("secretref://<provider>/<path>") that
// are resolved once at boot by a Provider. This package sits at the base of
// the dependency graph and imports only the standard library.
// See docs/blueprint/12-configuration-and-deployment.md §5.
package secrets

import (
	"context"
	"fmt"
	"strings"
)

// Scheme is the URI scheme that marks a string as a secret reference.
const Scheme = "secretref://"

// Ref identifies a secret held by a provider, e.g. "secretref://env/DB_DSN".
type Ref struct {
	Provider string // e.g. "env", "aws", "gcp", "k8s"
	Path     string // provider-specific path, e.g. "DB_DSN" or "prod/db/dsn"
}

// IsRef reports whether s looks like a secret reference.
func IsRef(s string) bool { return strings.HasPrefix(s, Scheme) }

// ParseRef parses "secretref://<provider>/<path>" into a Ref.
func ParseRef(s string) (Ref, error) {
	if !IsRef(s) {
		return Ref{}, fmt.Errorf("secrets: not a secret reference (want %q prefix): %q", Scheme, redactCandidate(s))
	}
	rest := strings.TrimPrefix(s, Scheme)
	provider, path, ok := strings.Cut(rest, "/")
	if !ok || provider == "" || path == "" {
		return Ref{}, fmt.Errorf("secrets: malformed secret reference (want secretref://<provider>/<path>)")
	}
	return Ref{Provider: provider, Path: path}, nil
}

// String renders the reference in canonical form. Safe to log: a Ref carries
// no secret material.
func (r Ref) String() string { return Scheme + r.Provider + "/" + r.Path }

// Provider resolves secret references to values. Implementations live in
// adapters (env, cloud secret managers); resolution happens once at boot in
// the app composition root — never on request or job hot paths.
type Provider interface {
	// Resolve returns the secret value for ref. Implementations must not log
	// the returned value.
	Resolve(ctx context.Context, ref Ref) (string, error)
}

// redactCandidate keeps error messages safe: if a non-ref string was passed
// where a reference was expected, it may itself be a raw secret, so echo
// nothing of it (review finding SEC-2, phase-00).
func redactCandidate(string) string { return "****" }
