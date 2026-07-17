package integration

import (
	"context"
	"strings"
	"testing"
)

type nilableProvider struct{}

func (*nilableProvider) Key() string                                   { return "widgets.mail" }
func (*nilableProvider) Kind() string                                  { return "email" }
func (*nilableProvider) HealthCheck(_ context.Context, _ Config) error { return nil }

// Third closure-audit regression (2026-07-17): a nil or typed-nil provider
// must be a collected boot error — Register previously dereferenced p.Key()
// and panicked (nil), or the typed nil surfaced only at first runtime use (an
// interface holding a nil pointer is not itself nil).
func TestRegisterRejectsNilAndTypedNilProviders(t *testing.T) {
	r := NewRegistry()
	r.Register("widgets", nil)
	r.Register("widgets", (*nilableProvider)(nil))
	err := r.Err()
	if err == nil {
		t.Fatal("nil/typed-nil providers passed registration")
	}
	for _, want := range []string{"nil integration provider", "typed-nil integration provider"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %v does not name %q", err, want)
		}
	}
}
