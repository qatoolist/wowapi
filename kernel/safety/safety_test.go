package safety_test

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/safety"
)

// TestMechanism_String covers every declared Mechanism value plus the
// default/unknown branch, so a regression in the switch (a wrong string, or a
// newly added Mechanism missing a case) fails here.
func TestMechanism_String(t *testing.T) {
	tests := []struct {
		name string
		m    safety.Mechanism
		want string
	}{
		{"None", safety.None, "none"},
		{"InboxEffectLedger", safety.InboxEffectLedger, "inbox_effect_ledger"},
		{"DomainCAS", safety.DomainCAS, "domain_cas"},
		{"ProviderIdempotencyKey", safety.ProviderIdempotencyKey, "provider_idempotency_key"},
		{"unknown/default", safety.Mechanism(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("Mechanism(%d).String() = %q, want %q", tt.m, got, tt.want)
			}
		})
	}
}

// TestMechanism_ZeroValue asserts the zero value of Mechanism is None, so a
// Declarer that forgets to set a mechanism reports the conservative "no
// built-in duplicate-safety" default rather than an unrelated value.
func TestMechanism_ZeroValue(t *testing.T) {
	var m safety.Mechanism
	if m != safety.None {
		t.Fatalf("zero value of Mechanism = %v, want None", m)
	}
	if got := m.String(); got != "none" {
		t.Fatalf("zero value Mechanism.String() = %q, want \"none\"", got)
	}
}

// fakeDeclarer verifies the Declarer contract compiles/behaves as documented:
// an adapter implementing it returns the mechanism it uses.
type fakeDeclarer struct{ mechanism safety.Mechanism }

func (f fakeDeclarer) DuplicateSafety() safety.Mechanism { return f.mechanism }

func TestDeclarer_Contract(t *testing.T) {
	var d safety.Declarer = fakeDeclarer{mechanism: safety.DomainCAS}
	if got := d.DuplicateSafety(); got != safety.DomainCAS {
		t.Fatalf("DuplicateSafety() = %v, want %v", got, safety.DomainCAS)
	}
}
