package config_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
)

func TestEgressExceptionsEnumeratesConfiguredExceptions(t *testing.T) {
	cfg := config.Defaults()
	cfg.Webhook.Outbound.AllowedHosts = []string{"relay.internal.example", "partner.example.com"}
	cfg.Webhook.Outbound.AllowedCIDRs = []string{"10.0.0.0/8", "192.168.1.0/24"}
	cfg.Security.TrustedIssuers = []string{"https://idp.example.com", "https://idp2.example.com"}

	exceptions := cfg.EgressExceptions()
	if len(exceptions) != 3 {
		t.Fatalf("expected 3 exception kinds, got %d", len(exceptions))
	}

	want := map[config.EgressExceptionKind][]string{
		config.EgressWebhookAllowedHosts: {"relay.internal.example", "partner.example.com"},
		config.EgressWebhookAllowedCIDRs: {"10.0.0.0/8", "192.168.1.0/24"},
		config.EgressJWKSTrustedIssuers:  {"https://idp.example.com", "https://idp2.example.com"},
	}
	for _, ex := range exceptions {
		got, ok := want[ex.Kind]
		if !ok {
			t.Fatalf("unexpected exception kind %q", ex.Kind)
		}
		if len(ex.Values) != len(got) {
			t.Fatalf("%s: expected %d values, got %d", ex.Kind, len(got), len(ex.Values))
		}
		for i, v := range ex.Values {
			if v != got[i] {
				t.Fatalf("%s[%d]: expected %q, got %q", ex.Kind, i, got[i], v)
			}
		}
	}
}

func TestEgressExceptionsContainsNoCredentialValues(t *testing.T) {
	cfg := config.Defaults()
	// Intentionally inject values that look like secrets/credentials into the
	// allowlist fields. The report must echo them back as configured values
	// (they are host/CIDR shaped), but must NOT include any DSN/secret values
	// from elsewhere in the config.
	cfg.Webhook.Outbound.AllowedHosts = []string{"token=abc123.example.com"}
	cfg.Webhook.Outbound.AllowedCIDRs = []string{"10.0.0.0/8"}
	cfg.Security.TrustedIssuers = []string{"https://issuer.example.com"}
	cfg.DB.DSN = config.NewSecret("secretref://env/DATABASE_URL", "postgres://user:pass@host/db")

	exceptions := cfg.EgressExceptions()

	for _, ex := range exceptions {
		for _, v := range ex.Values {
			if strings.Contains(v, "DATABASE_URL") || strings.Contains(v, "secretref") {
				t.Fatalf("egress exception %q leaked a secret reference: %q", ex.Kind, v)
			}
		}
	}
}

func TestEgressExceptionsOmitsEmptyAllowlists(t *testing.T) {
	cfg := config.Defaults()
	if len(cfg.EgressExceptions()) != 0 {
		t.Fatalf("default config must produce no egress exceptions, got %v", cfg.EgressExceptions())
	}
}
