// Package envprovider implements secrets.Provider for the "env" scheme:
// references of the form "secretref://env/<VAR>" are resolved from the
// process environment. This is the local/dev provider; cloud providers
// follow the same layout at adapters/secrets/<name>provider.
//
// See docs/blueprint/12-configuration-and-deployment.md §5 and decision
// D-0013 in docs/implementation/decisions.md.
package envprovider

import (
	"context"
	"fmt"
	"os"

	"github.com/qatoolist/wowapi/kernel/secrets"
)

// Provider resolves "secretref://env/<VAR>" from the process environment.
type Provider struct {
	lookup func(string) (string, bool)
}

// New returns a Provider backed by os.LookupEnv.
func New() *Provider {
	return NewWithLookup(os.LookupEnv)
}

// NewWithLookup returns a Provider using an injectable lookup function.
// The function must match os.LookupEnv semantics: returning (value, true)
// when the variable is set (even if empty) and ("", false) when absent.
func NewWithLookup(lookup func(string) (string, bool)) *Provider {
	return &Provider{lookup: lookup}
}

// Resolve returns the environment-variable value named by ref.Path.
//
// ref.Provider must be "env"; any other provider name is rejected so that
// a mis-routed reference fails fast rather than silently resolving nothing.
//
// Missing and empty-string variables are both treated as errors: an absent
// variable is a deployment gap, and an empty secret is a deployment bug
// (a deliberately blank value must be modelled as an absent feature, not an
// empty secret reference).
func (p *Provider) Resolve(_ context.Context, ref secrets.Ref) (string, error) {
	if ref.Provider != "env" {
		return "", fmt.Errorf("envprovider: wrong provider: got %q, want %q", ref.Provider, "env")
	}
	val, ok := p.lookup(ref.Path)
	if !ok {
		return "", fmt.Errorf("envprovider: environment variable %q is not set (ref: %s)", ref.Path, ref)
	}
	if val == "" {
		return "", fmt.Errorf("envprovider: environment variable %q is empty (empty secret is a deploy bug; ref: %s)", ref.Path, ref)
	}
	return val, nil
}
